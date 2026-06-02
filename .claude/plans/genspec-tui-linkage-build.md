# Cross-ref linkage — build plan

Status: ⬜ ready to build. Implements the settled design in
[`genspec-tui-linkage.md`](genspec-tui-linkage.md) (🟢 rev 9). Multi-phase, spans
two modules (codescan library + `cmd/genspec-tui`). Legend: ⬜ todo · 🔶 in
progress · ✅ done.

## Scope

Replace the name-matching `naiveLinker` with a position-backed, bidirectional
spec ↔ source linker. Four phases, two of them independent and parallelizable:

- **Phase A — spec-side index (`LX-spec-0`)** — TUI only, not rebase-gated.
  Rendered-line ↔ JSON-pointer index. *Start here.*
- **Phase B — source-side provenance (`LX-prov-0`)** — codescan only. The
  `OnProvenance` callback + `RecordOrigin` at anchor sites. Independent of A.
- **Phase C — join (`LX-join-0/1`)** — TUI. `positionLinker` + the `f` link-nav
  loop, both directions. Needs A + B.
- **Phase D — find-references (`LX-refs`)** — TUI. Render-time `$ref` resolution
  + `F3`/`Shift-F3` cycling. Needs C.

Out of this plan: `LX-model` (library-owned derived model — rides the
internal-model stream), `x-go-origin` materialization, LSP token-level positions.

## Dependency graph

```
A (spec index) ─┐
                ├─► C (join + nav) ─► D (find-refs)
B (provenance) ─┘
```

A and B have no ordering constraint — build in either order or in parallel.

---

## Phase A — spec-side index (`LX-spec-0`, TUI)

New file `cmd/genspec-tui/internal/ux/specindex.go` + `specindex_test.go`.

- ✅ **A0 — `GOEXPERIMENT=jsonv2` (decision D1).** *Finding: no `go.mod`/`go.work`
  bump needed* — the `jsontext` gate is the build tag `goexperiment.jsonv2`, not
  the go version, so it compiles under the current go1.25 directives once the flag
  is set (avoids the workspace-wide ripple; library CI untouched). Cost: the whole
  `ux` package is now experiment-gated, so **all TUI builds/tests/CI must set
  `GOEXPERIMENT=jsonv2`** (⚠️ CI workflows need updating — see note below).
- ✅ **A1 — offset→line table.** `lineTable` over `\n` offsets; `lineAt` via
  `sort.Search`, 0-based to match the viewport's `strings.Split` addressing.
- ✅ **A2 — JSON index.** `jsontext.Decoder` token loop; first occurrence of each
  `StackPointer()` + `InputOffset()`→line is the member's line (semantics verified
  by a throwaway probe). Decodes the rendered bytes (ordered keys preserved).
- ✅ **A3 — YAML index.** `yaml.Node` walk (`.Line`), same RFC 6901 escaping as
  the JSON side, so the two indexes are interchangeable.
- ✅ **A4 — model integration.** `refreshSpec` rebuilds `m.specIndex` per active
  format; `Spec.TopLine()` added; status-line demo shows the pointer of the
  top-of-viewport node while the spec pane is focused.
- ✅ **A5 — tests.** `specindex_test.go`: JSON + YAML pointer/line assertions
  (nested objects, array element, escaped `/pets`→`~1pets`), nearest-preceding
  `PointerAt`, nil/empty, and the line table.

**Status: LX-spec-0 complete — `GOEXPERIMENT=jsonv2 go test ./cmd/genspec-tui/...`
green, lint clean. ⚠️ CI must add `GOEXPERIMENT=jsonv2` for the TUI module or its
build breaks (the experiment is required to compile `ux` now).**

---

## Phase B — source-side provenance (`LX-prov-0`, codescan)

- ✅ **B0 — `Provenance` type + option (decision D2).** `internal/scanner/
  provenance.go`: `Provenance{Pointer, Pos}` + `JSONPointer(...)` helper;
  `Options.OnProvenance func(Provenance)` on `Options` (auto-public via
  `codescan.Options`).
- ✅ **B1 — plumbing.** *Deviation from the plan:* `RecordOrigin` + `OnProvenance`
  + `OriginEnabled` live on **`ScanCtx`**, not `common.Builder` — because the
  emit is callback-only (no accumulation, unlike diagnostics) *and* `spec.Builder`
  isn't a `common.Builder`. A `ScanCtx` method is callable uniformly from every
  builder (`s.ctx`/`s.Ctx`). Better choice; recorded here.
- ✅ **B2 — pointer helper.** `scanner.JSONPointer(segments...)` — RFC 6901
  escaping matching the spec-side jsontext output (verified by the matching enum
  fixture pointer in the LX-spec-0 tests).
- 🔶 **B3 — wire anchor sites** (record from the insertion site where the
  *absolute* pointer is known):
  - ✅ definitions → `spec.go` `buildDiscoveredSchema`: `/definitions/{name}` ←
    `decl.Ident.Pos()`, and `schema.WithPath("/definitions/{name}")` passed in.
  - ✅ properties (first-level) → `schema/fields.go` `applyFieldCarrier`:
    `s.path + /properties/{json}` ← `c.afld.Pos()`, gated on `WithPath` set.
  - ⬜ responses → `responses/responses.go`; `/responses/{name}` (unambiguous).
  - ⬜ paths/operations → `routes.go`/`operations.go`; `/paths/{path~esc}/{method}` (grammar block `Pos()`, coarse §3.5).
  - ⬜ parameters → `/paths/{path~esc}/{method}/parameters/{i}`.
  - ⬜ enum values → `…/enum/{i}` ← const position.
  - ⬜ `swagger:meta` → `/info` (children resolve upward), meta block `Pos()`.
- 🔶 **B4 — tests.** ✅ `coverage_provenance_test.go`: `/definitions/User`
  pointer+position, and opt-in (off → nothing, spec unchanged). More assertions
  land with each anchor.

**Path threading — decided (Fred): `schema.WithPath(base)`.** General, not
definition-specific (not all pointers live under `/definitions`, and
`RecordOrigin` isn't schema-only). Top builders *initiate* the base
(`/definitions/User`, `/paths/.../responses/200/schema`, …); sub-builders
**path-join** their segment (`/properties/x`, `/items`, …). Empty base = record
nothing. Done for definitions + first-level properties; **deeper recursion
(nested objects, `items`, `allOf` members) still threads the same `path` field
into the recursive build calls — staged next.**

**Increment so far:** `OnProvenance` proven end-to-end; **definitions +
first-level properties** anchored via `WithPath`; full library suite green, lint
clean, opt-in (nil callback = zero cost, spec byte-identical).

---

## Phase C — join: linker + navigation (`LX-join-0/1`, TUI)

- ⬜ **C1 — collect provenance.** In `scan.go`, set `cfg.OnProvenance`; collect
  `[]scanner.Provenance` into `scanResultMsg`; carry to the model (as the
  diagnostics wiring does).
- ⬜ **C2 — source index.** Build the caller-owned structure: forward
  `map[pointer]token.Position` (+ zero-alloc segment-trim nearest-ancestor walk)
  and the reverse `(file,line)`-sorted slice + binary search (§3.7).
- ⬜ **C3 — `positionLinker`.** New `linker.go` implementing `SourceLinker`
  (and the reverse direction), replacing `naiveLinker`. Delete the name-match impl.
- ⬜ **C4 — `f` binding + nav state.** Add `F` to the `key` enum + dispatch; a
  `navMode` + driver-pane bit on the model; enter from the read-only viewer / spec
  viewport, `ESC` exits (§6.6).
- ⬜ **C5 — spec→source flow.** Cursor move in spec nav → `line2ptr` →
  nearest-ancestor `pos` → open file, vertical-center + highlight (§6.2).
- ⬜ **C6 — source→spec flow.** Cursor move in source nav → reverse
  nearest-enclosing anchor → `ptr2line` → center + highlight spec node (§6.3).
- ⬜ **C7 — visuals + edge cases.** Distinct driver/follower highlight styles;
  status badge (`SPEC-NAV`/`CODE-NAV` + resolved target / `no source`); debounced
  auto-follow; honest "no source" for `InputSpec` nodes; "stale" handling while
  the buffer is dirty (§6.4–6.5).
- ⬜ **C8 — tests.** Linker unit tests (nearest-ancestor; reverse nearest-
  enclosing; no-source). Model-level: a synthesized scan → assert the follower
  target line for a given driver line, both directions.

**Exit:** `f` drives live bidirectional navigation on a real scan.

---

## Phase D — find-references (`LX-refs`, TUI)

- ⬜ **D1 — `$ref` resolution.** From a definition pointer, scan the rendered
  spec for `$ref`s targeting it (handle the OAS2 quirks noted in §3.4); produce an
  ordered candidate list.
- ⬜ **D2 — `F3`/`Shift-F3` cycling.** Bindings + a candidate cursor; step the
  follower through the multiple locations (§6.4).
- ⬜ **D3 — tests.** A fixture with a type `$ref`'d from ≥2 sites; assert the
  candidate set + cycle order.

---

## Suggested PR sequence

1. **PR-A** = Phase A (`LX-spec-0`). TUI-only, go1.26 bump, self-contained,
   testable; status-line pointer demo. Ships independently.
2. **PR-B** = Phase B (`LX-prov-0`). codescan-only, additive API, callback-tested.
   No TUI change. (Parallelizable with PR-A.)
3. **PR-C** = Phase C (`LX-join`). The visible feature; needs A + B merged.
4. **PR-D** = Phase D (`LX-refs`).

---

## Decisions to confirm

- ✅ **D1 — go1.26 + `GOEXPERIMENT=jsonv2` for the TUI.** Settled (Fred): ship
  `jsontext` via the experiment flag; official builds from go1.26 with the flag
  (controlled release pipeline), users building their own are warned. TUI module
  drops go1.25; library keeps its 2-version window. (jsontext is still
  experiment-gated even in go1.26.3 — verified.)
- ⬜ **D2 — `Provenance` home.** `internal/scanner` (recommended, beside
  `Options`) vs the `codescan` root. Either way it's public via `codescan.Options`
  and importable by the TUI through the internal path.
- ✅ **D3 — coarse positions** for routes/ops/meta (grammar block `Pos()`) and
  line-granularity overall — settled in design §3.5.

## References

- Design: [`genspec-tui-linkage.md`](genspec-tui-linkage.md) — §3 (source side),
  §4 (spec side), §5 (join), §6 (UX), §7 (milestones), §9 (recon citations).
- Pattern to mirror: the diagnostics wiring (`b2574fe`) and `OnDiagnostic`
  plumbing (`internal/scanner/options.go`, `scan_context.go`, `common/builder.go`).
- [[project_genspec_tui]] · [[project_lsp_diagnostics_target]].
