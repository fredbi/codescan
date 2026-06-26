# Cross-ref linkage — build plan

Status: 🔶 in progress. Implements the settled design in
[`genspec-tui-linkage.md`](genspec-tui-linkage.md) (🟢 rev 9). Multi-phase, spans
two modules (codescan library + `cmd/genspec-tui`). Legend: ⬜ todo · 🔶 in
progress · ✅ done. **Phase A ✅, Phase B ✅ (all anchor kinds) — Phase C (the
join) is next, the first user-visible payoff.**

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

## Phase B — source-side provenance (`LX-prov-0`, codescan) — ✅ COMPLETE

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
- ✅ **B3 — wire anchor sites** (record from the insertion site where the
  *absolute* pointer is known):
  - ✅ definitions → `spec.go` `buildDiscoveredSchema`: `/definitions/{name}` ←
    `decl.Ident.Pos()`, and `schema.WithPath("/definitions/{name}")` passed in.
  - ✅ properties (all depths) → `schema/fields.go` `applyFieldCarrier`:
    `fieldPath` ← `c.afld.Pos()`, gated on `WithPath` set. Interface methods
    share the same path (they route through `applyFieldCarrier`).
  - ✅ responses → `spec.go` `buildResponses`: `/responses/{name}` ←
    `decl.Ident.Pos()` (top-level `swagger:response`).
  - ✅ paths/operations → `routes.go` + `operations.go`:
    `/paths/{path~esc}/{method}` ← `ParsedPathContent.Pos` (a new coarse field =
    matched annotation comment's `Slash`, §3.5).
  - ✅ parameters → `/paths/{path~esc}/{method}/parameters/{i}`, **deferred pass**
    (`spec.go` `emitParameterAnchors`): parameters build before their op is bound
    to a path and before the array index is final, so the parameters builder only
    captures `(opid → name → pos)` via `ScanCtx.RecordParamOrigin`; the absolute
    pointer is assembled from the finished paths tree.
  - ✅ enum values → `…/enum/{i}` ← const ident position. `FindEnumValues` now
    returns parallel positions; emitted in `walker_classifiers.go`
    `classifierNamedBasic`. **Fires on the common enum-on-field case** (the path
    is *advanced* to the field/items node — see threading note); standalone
    `swagger:enum` definitions are vanishingly rare (the const-enum type always
    inlines onto its referencing field — verified empirically).
  - ✅ `swagger:meta` → `/info` ← meta block `cg.Pos()` (`spec.go` `buildMeta`).
- ✅ **B5 — keyword-line granularity (post-Phase-B extension).** Beyond the
  node-level anchors above, each *scalar keyword* now anchors to its own
  `// keyword: …` comment line (so following e.g. a `maximum`/`default`/`host`
  node lands on the annotation, not the enclosing field/block). Same mechanism
  everywhere — walk the grammar `Property` stream (each carries `.Pos` +
  `ItemsDepth`), map keyword→pointer-segment, emit:
  - schema validations → `schema/walker.go` `recordValidationOrigins` (curated
    set: maximum/minimum/multipleOf/max·minLength/pattern/max·minItems/
    uniqueItems/default/example/enum/readOnly; `base + (/items)×ItemsDepth + seg`).
  - meta keywords → `spec.go` `recordMetaOrigins` (Info.* under `/info`, the rest
    at root — the root-level fields had **no** ancestor anchor before).
  - route-header keywords → `routes/walker.go` `recordRouteKeywordOrigin`
    (schemes/deprecated/consumes/produces under `/paths/{p}/{m}/{seg}`).
  - **Deliberately NOT covered:** `required` (parent array), `$ref`-with-siblings
    overrides (rewritten to allOf), parameter/header validations (different
    builder), and **swagger:operation** keywords (body is one wholesale-
    unmarshaled YAML block — no per-keyword `Property`; would need a `yaml.Node`
    walk; Fred: leave at the coarse `/paths/{p}/{m}` anchor). All resolve to their
    field/block anchor — still correct under anchors-only.
- ✅ **B4 — tests.** `coverage_provenance_test.go`:
  - `TestCoverage_ProvenanceDefinitions` — simplest models-only case.
  - `TestCoverage_ProvenanceGeometry` — **the anchors-only safety invariant**:
    every emitted pointer must resolve to an existing node in the rendered spec
    (pure-stdlib RFC 6901 resolver), run across allOf / embed / interface /
    nested / slice / map / petstore fixtures. Mutation-tested (removing a clear
    surfaces dangling allOf anchors).
  - `TestCoverage_ProvenanceAnchorKinds` — asserts ≥1 anchor of every kind
    (definition, property, enum, response, operation, parameter, info) fires with
    a source line, over petstore + enum-docs.

**Path threading — UPGRADED from "clear" to "advance".** The base path (`s.path`,
set by `schema.WithPath(base)`) now tracks the *exact pointer of the schema node
currently being filled*: `applyFieldCarrier` advances it to the property's
pointer for the value build (`descend`/`repath` helpers, mirroring
`enterEmbed`); the `items` (slice/array/named-array) and `additionalProperties`
(map) arms path-join their segment; allOf members and allOf own-property targets
**clear** it (those subtrees aren't tracked — resolve to the nearest anchored
ancestor). This is what lets enum-on-field, nested inline objects, and slice/map
element fields anchor at the correct pointer instead of dangling. The geometry
invariant test is the safety net for the whole scheme.

**Increment:** `OnProvenance` proven end-to-end across **all anchor kinds**; full
library suite green, lint clean (`--new-from-rev master`), opt-in (nil callback =
zero cost, spec byte-identical — descend/repath are no-ops when `s.path == ""`).

---

## Phase C — join: linker + navigation (`LX-join-0/1`, TUI)

- ✅ **C1 — collect provenance.** `scan.go` sets `cfg.OnProvenance`, collects
  `[]scanner.Provenance` into `scanResultMsg`; the model builds the `SourceIndex`
  on each scan (mirrors the diagnostics wiring).
- ✅ **C2 — source index.** `sourceindex.go`: caller-owned `SourceIndex` —
  forward `map[pointer]token.Position` with `PositionFor` (zero-alloc segment-trim
  nearest-ancestor) and the reverse `byFile []lineAnchor` (sorted) + `PointerAt`
  binary search (nearest-enclosing). Unit-tested (`sourceindex_test.go`:
  exact / nearest-ancestor / nearest-enclosing / per-file isolation / last-wins).
- ✅ **C3 — retire `naiveLinker`.** Done (slice 6). `linkage.go` (the
  name-matching `SourceLinker`/`naiveLinker`/`Selection`/`SpecTarget` scaffold)
  and the dead `Spec.JumpTo` substring scan are deleted; the position-backed nav
  goes through `SourceIndex` directly. `g` (locate a file in the spec) is
  reimplemented exactly via `SourceIndex.FirstAnchor` → `LineForPointer` →
  highlight+focus — unambiguous, consistent with the `f` flows. (No separate
  `positionLinker` type was needed; the model talks to `SourceIndex`/`SpecIndex`.)
- 🔶 **C4 — `f` binding + nav state.** ✅ **Slice 3 (read-only viewer) landed** —
  the keystone. `FileView` is now dual-mode: a navigable, line-numbered read-only
  viewer (highlighted nav line; `↑↓/jk`/wheel move it) + the textarea editor,
  opt-in via `i`/Enter, left via Esc (two-level: editor → viewer → tree). Files
  open read-only. Because the viewer no longer eats `f` for typing, **follow is
  unified on `f`** from both panes (spec pane → source; read-only viewer →
  spec; `ctrl+f` stays as the in-editor shortcut). ✅ **Slice 5 (auto-follow mode)
  landed:** `f` toggles a *persistent* follow mode — the driver pane keeps focus,
  the follower mirrors on every cursor move (`syncFollowIfActive` after each key/
  scroll/rescan); spec-driver marks the node without re-scrolling (no fight),
  source-driver scrolls+highlights the spec; any focus change / edit / search /
  options exits, as do Esc and a second `f`; a `SPEC▸SOURCE`/`SOURCE▸SPEC` status
  badge shows the live target. Model + panel tests cover toggle/drive/exit/no-node.
  **C4 is essentially done.**
- 🔶 **C5 — spec→source flow.** ✅ **Slice 1 landed:** `f` on the spec pane takes
  the pointer at the top of the viewport (`SpecIndex.PointerAt`) → `SourceIndex.
  PositionFor` (nearest-ancestor) → opens the file at that line
  (`FileView.GotoLine`), now landing in the **read-only viewer with the source
  line highlighted** (slice 3), status shows `→ file:line (pointer)`. ⬜
  Spec-side highlight + vertical-centering is C7.
- 🔶 **C6 — source→spec flow.** ✅ **Slice 2 landed:** `ctrl+f` from the file
  editor takes the cursor's source line → `SourceIndex.PointerAt` (nearest-
  enclosing anchor) → `SpecIndex.LineForPointer` → scrolls the spec pane there and
  focuses it (status shows the pointer). `ctrl+f` not `f` because the editor owns
  plain `f`. ⬜ Centering + highlight is C7. The asymmetric keys (`f` spec→source,
  `ctrl+f` source→spec) collapse to a unified `f` once the read-only viewer lands.
- 🔶 **C7 — visuals + edge cases.** ✅ **Slice 4 landed:** both panes now
  highlight the linked node — the read-only viewer's nav line (source side) and
  the spec's xref line (`theme.Selected`, whole-line), set on follow and
  invalidated by search / new content / Esc; source→spec highlights the
  destination, spec→source leaves a breadcrumb on the origin. ⬜ Still ahead:
  a status badge / nav-mode indicator, honest "no source" for `InputSpec` nodes,
  and "stale" handling while the buffer is dirty.
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
