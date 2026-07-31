# Cross-ref linkage — build plan

Status: 🔶 in progress. Implements the settled design in
[`genspec-tui-linkage.md`](genspec-tui-linkage.md) (🟢 rev 9). Multi-phase, spans
two modules (codescan library + `cmd/genspec-tui`). Legend: ⬜ todo · 🔶 in
progress · ✅ done. **Phase A ✅, Phase B ✅ (all anchor kinds), Phase C core ✅
(bidirectional `f` follow, all three drivers) — the current series finishes the
chain: C7 tail → C8 → Phase D. See [Series 2](#series-2--finish-the-chain) for
the slice plan — **S7–S13 all ✅ as of 2026-07-30, and the three spun-out
backlog items (B-spec-line-cursor, B-viewer-global-keys, B-rescan-anchor) are
✅ too. The chain is complete and the backlog is empty.**

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

- ✅ **A0 — ~~`GOEXPERIMENT=jsonv2`~~ (decision D1) — SUPERSEDED by `ca0653c`.**
  *Historic:* the first cut decoded with `jsontext`, whose gate is the build tag
  `goexperiment.jsonv2`, so the whole `ux` package became experiment-gated and
  every build/test/CI invocation had to set the flag. **`ca0653c` retired that**:
  both indexes are now built from the `go-openapi/core/json` lexers
  (`default-lexer` + `yaml-lexer`, `WithJSONPointer(true)`), which hand us the
  pointer *and* the line directly. No experiment flag, no toolchain pin, plain
  `go test ./cmd/genspec-tui/...`. Fallout to clean in S7: the offset→line
  `lineTable` (A1) is now dead production code, still carrying an orphaned
  `// itoa …` doc comment.
- ✅ **A1 — offset→line table.** `lineTable` over `\n` offsets; `lineAt` via
  `sort.Search`, 0-based to match the viewport's `strings.Split` addressing.
  *(Dead since `ca0653c` — the lexers report `Line()` themselves. Deleted in S7.)*
- ✅ **A2 — JSON index.** ~~`jsontext.Decoder` token loop~~ → `lexer.NewVerbatim
  WithBytes(b, WithJSONPointer(true))`: per token, `Line()` + `JSONPointer()`;
  first occurrence of a pointer is the member's declaration line. Reads the
  rendered bytes (ordered keys preserved).
- ✅ **A3 — YAML index.** ~~`yaml.Node` walk~~ → `yamllexer.NewWithBytes`, same
  token loop, same RFC 6901 escaping as the JSON side, so the two indexes are
  interchangeable.
- ✅ **A4 — model integration.** `refreshSpec` rebuilds `m.specIndex` per active
  format; `Spec.TopLine()` added; status-line demo shows the pointer of the
  top-of-viewport node while the spec pane is focused.
- ✅ **A5 — tests.** `specindex_test.go`: JSON + YAML pointer/line assertions
  (nested objects, array element, escaped `/pets`→`~1pets`), nearest-preceding
  `PointerAt`, nil/empty, and the line table.

**Status: LX-spec-0 complete — `go test ./cmd/genspec-tui/...` green, lint clean.
No toolchain flag: `ca0653c` removed the experiment dependency (A0).**

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
  honest "no source" for `InputSpec` nodes, "stale" handling while the buffer is
  dirty, format-toggle pointer preservation (→ **S7**), and the driver/follower
  visual split + centering (→ **S8**).
- ⬜ **C8 — tests.** Linker unit tests (nearest-ancestor; reverse nearest-
  enclosing; no-source) — *unit level already covered by `sourceindex_test.go`*.
  Model-level: a synthesized scan → assert the follower target line for a given
  driver line, both directions (→ **S9**).

**Exit:** `f` drives live bidirectional navigation on a real scan.

---

## Phase D — find-references (`LX-refs`, TUI)

- ⬜ **D1 — `$ref` resolution.** From a definition pointer, scan the rendered
  spec for `$ref`s targeting it (handle the OAS2 quirks noted in §3.4); produce an
  ordered candidate list (→ **S10**).
- ⬜ **D2 — `F3`/`Shift-F3` cycling.** Bindings + a candidate cursor; step the
  follower through the multiple locations (§6.4). Plus the inverse,
  go-to-definition on `Enter` (→ **S11**).
- ⬜ **D3 — tests.** A fixture with a type `$ref`'d from ≥2 sites; assert the
  candidate set + cycle order (→ **S12**).

---

## Series 2 — finish the chain

The executable slice plan for the current series (2026-07-30). Scope confirmed
with Fred: **finish the linkage chain** — C7 tail, C8, Phase D — plus two
in-scope additions he green-lit: **go-to-definition on `Enter`** (the inverse of
find-references, near-free once the ref index exists) and **gutter dots**
(§6.5's optional discoverability item) as the closing slice.

Slices are ordered so each is independently reviewable and leaves the tree green
(`go test ./cmd/genspec-tui/...` + `golangci-lint run --new-from-rev master`).

### S7 — C7 tail: honest edges — ✅ DONE

Closed the three §6.4 edge cases the follow loop glossed over, and swept the
`ca0653c` fallout.

- ✅ **Dead-code sweep.** Deleted `lineTable` / `newLineTable` / `lineAt` from
  `index/specindex.go` (+ its test block) and the orphaned `// itoa …` comment
  that had ended up above the type. The lexers report `Line()` directly (A0).
- ✅ **No source is a first-class outcome (§3.8).** The link helpers now name
  which miss occurred instead of collapsing all of them into one message, via a
  shared const block in `model.go`: `noNodeDesc`, `noFileDesc`, `noAnchorDesc`,
  `noProvenanceDesc`, `noSourceSuffix`, `notRenderedSuffix`. The follower holds
  position on every miss rather than mis-jumping.
  - *Bug found and fixed:* `linkSourceToSpec` already computed the honest
    `"… (not in view)"` description — and **both call sites threw it away**,
    `syncFollowIfActive` flattening it to `"(no spec node)"` and `handleEditKey`
    to a hardcoded string. The helper now returns a description that is always
    meaningful, with the bool reporting *only* whether the follower moved.
  - *Also:* `driveDiagToSource`'s `"(no source)"` became `"(diagnostic carries
    no position)"` — a positionless diagnostic is a different failure from an
    unanchored spec node, and it read as the latter.
  - `srcIndex.Len() == 0` now reports `noProvenanceDesc` ("nothing was anchored
    at all") rather than implying this particular node is special.
- ✅ **Stale-while-dirty.** `Model.stale()` (= `FileView.Dirty()`) drives a
  `STALE` chip on the follow badge (`theme.Stale()`, amber). Nav keeps working;
  it just stops pretending to be exact. Cleared by save → watcher → rescan.
- ✅ **Format toggle preserves the pointer, not the line.** New
  `Model.setSpecFormat` captures `PointerAt(TopLine())`, swaps format, rebuilds
  the index, then `LineForPointer` → new `Spec.ScrollTo`. `ctrl+j`/`ctrl+y` route
  through it. **This was a live bug, not just a missing nicety:** the handlers
  called `refreshSpec()` under an unchanged `YOffset`, and since the YAML render
  is roughly half the height of the JSON one, toggling reliably dumped the user
  on an unrelated node. Mutation-verified (the test fails without the fix).

**Tests:** new `model_edges_test.go` — format-toggle preservation (helper +
`ctrl+y` binding + round-trip + same-format no-op + nil-index), a subtest per
miss kind for both `linkSourceToSpec` and `driveSpecToSource` (asserting the
follower *holds*), and the STALE badge across clean → dirty → saved.
Two existing expectations updated for the intentional wording change.
`go test work ./...` green; `-race` green.

Landed as `16621bd`.

**Spun out → [B-rescan-anchor](#b-rescan-anchor--backlog).**

### S8 — C7 tail: nav visuals (§6.5) — ✅ DONE

- ✅ **Driver vs follower are visually distinct.** New `theme.Follower()` (muted
  tint, `colorFollower` 53) beside `theme.Selected()` (the strong bar).
  - *The role bit came for free:* the design's own invariant is that **the driver
    keeps focus** (§6.1), so `focused` already *is* "this pane drives" — no new
    plumbing, no role parameter threaded through the panels. `Spec.xrefStyle()`
    and `panels.navStyle(focused)` each pick from that one bit.
  - `Spec` bakes the style into the viewport content at render time, so
    `Spec.View` re-renders on a focus **transition** (guarded — the spec can be
    thousands of lines and `View` runs on every message). `FileView.viewerBody`
    already rebuilt per call, so it just picks the style.
- ✅ **Centre the follower.** *Deviation from the plan, for a smaller API:*
  rather than adding `CenterLine`/`CenterOn` beside the existing methods, the
  **jump primitives themselves now centre** — `Spec.HighlightLine` and
  `FileView.GotoLine`. Adding parallel methods would have left the old ones
  production-dead (only `HighlightLine` and `GotoLine` were reachable), and S7
  had just finished deleting dead code. The rule that fell out is cleaner than
  the one planned: **a jump centres; incremental cursor movement scrolls
  minimally.** Nav keys (`NavUp`/`NavDown`/`ScrollBy` → `gotoNav`) keep the
  minimal-scroll behaviour, which is what you want when walking line by line;
  search keeps its own `scrollContext = 2` top bias (you want the *next* matches
  visible below). `scrollContext` survives — `Diagnostics.ScrollToLine` uses it.

**Tests:** new `panels/navvisuals_test.go` + `panels/main_test.go`.

⚠️ **The tests needed a forced colour profile to mean anything.** lipgloss
degrades to plain text when stdout is not a TTY, which `go test` never is — so
a driver line and a follower line rendered byte-identically and every assertion
passed against a panel that applied *no style at all*. `TestMain` now sets
`lipgloss.SetColorProfile(termenv.TrueColor)` for the package (promotes
`muesli/termenv` from indirect to a direct require). A first pass at the
`FileView` assertion compared whole views and was **confounded by the border and
title, which also change with focus** — it passed under mutation; it now
compares the nav line's own SGR prefix. Both wirings are mutation-verified
(forcing either panel back to a single style fails the suite).

*Drive-by:* `go mod tidy` also dropped a stale `swag/jsonname` indirect require.

### S9 — C8: model-level linkage tests — ✅ DONE

New `model_join_test.go`, 17 tests. The two indexes already had unit tests; what
was missing was proof that **the model wires them together**. So the fixture
synthesizes a scan the way one actually arrives — a rendered spec body plus the
`[]scanner.Provenance` a scan emits — builds the indexes with the **real**
builders (`refreshSpec` → `BuildJSONIndex`, `BuildSourceIndex`), writes the Go
sources to a temp dir so `loadFileQuietly` reads real files, and then asserts
**where the follower actually lands** rather than only what the status line says.

The provenance set anchors only definitions and properties, mirroring codescan's
anchors-only emission (§3.4). That is deliberate: it leaves `…/email/type` and a
struct's closing brace unanchored, which is what makes nearest-ancestor and
nearest-enclosing resolution *observable* instead of incidental.

Covered: exact-anchor landing both ways · nearest-ancestor (spec→source) ·
nearest-enclosing (source→spec) · file switching when the target moves between
files · the follower tracking a walking driver (table-driven, both directions) ·
above-first-anchor holding position · **follow surviving a rescan** (a definition
inserted *above* the followed node, so a stale line number would land wrong —
asserts re-resolution against the rebuilt index) · exits (search, options, edit,
second `f`, focus change). `TestJoin_SpecIndexMatchesFixture` guards the
hand-counted line constants the rest of the file depends on.

Mutation-verified in both directions: disabling `PositionFor`'s ancestor walk
fails the nearest-ancestor test; disabling `PointerAt`'s nearest-enclosing search
fails two more.

**Two findings:**

1. ⚠️ **The read-only viewer swallows every global key.** `handleViewerKey`
   returns `m, nil` for anything it does not own, so while a file is open `/`
   (search), `o` (options), `r` (rescan), `g` (locate) and `ctrl+j`/`ctrl+y`
   (format toggle) all silently do nothing. The viewer's status line advertises
   only its own keys, so this may well be deliberate — but you cannot toggle
   JSON/YAML or rescan while reading source, which is an odd gap in the
   edit→rescan loop. `TestJoin_ViewerSwallowsGlobalKeys` pins the *current*
   behaviour so a future change is a deliberate one. **Not fixed — UX-polish
   decision, → [B-viewer-global-keys](#b-viewer-global-keys--backlog).**
2. *Test-only:* a `Model` built by hand needs a real `textinput.New()` —
   a zero-value one panics inside `Focus()`. Production always goes through
   `New()`, so this is a fixture requirement, not a bug.


### S10 — D1: the `$ref` site index — ✅ DONE

New `index/refindex.go` + `refindex_test.go`.

- ✅ **One walk, two indexes.** `BuildJSONIndex` / `BuildYAMLIndex` now return
  `(*SpecIndex, *RefIndex)`, sharing an `indexAccum` that folds each token into
  both. *Probed rather than assumed:* the lexer reports the `$ref` **key** and
  its **value** under the same `…/$ref` pointer, and `Value()` hands back the
  target already unquoted (single quotes stripped on the YAML side too) — so a
  non-key scalar token whose pointer ends in `/$ref` is the whole detection rule,
  identical for both lexers.
- ✅ **Shape.** `RefSite{Pointer, Line, Target}` where `Pointer` is the node
  **holding** the ref (the `/$ref` segment trimmed) — that is the node the user
  navigates to, not the `$ref` member itself. `RefsToPointer(ptr)` returns sites
  line-ordered (the order `F3` will step through); `RefAt(line)` backs
  go-to-definition.
- ✅ **Local vs external.** `ParseRefTarget` marks `#/…` fragments local and
  everything else (sibling file, URL, bare filename, plain-name `#Foo`) recorded
  verbatim but non-resolvable. The §3.4 quirks (`$ref` inside `allOf`,
  ref-to-ref chains, ignored siblings) stay **documented, not resolved** — this
  is a render-time site index, not a JSON-Schema resolver.
- ✅ **Percent-decoding, deliberately.** `go-openapi/spec` marshals refs through
  `net/url`, so a definition name containing e.g. a space is emitted as
  `#/definitions/A%20B` — while `SpecIndex` keys on the document's own key text,
  which is *not* escaped. Decoding the fragment is what keeps the two sides
  comparable; a malformed escape falls back to the verbatim fragment rather than
  dropping the ref. (Unreachable with Go-identifier names today, but silent
  mismatch is the worst failure mode for a nav feature.) RFC 6901 `~0`/`~1`
  escaping is *not* decoded — both sides carry it, so it matches verbatim.

**Tests:** 3 local sites (property / array items / response schema) found and
line-ordered, path-key escaping preserved through the join, external ref counted
but never matched as local, `RefAt`, YAML pointer-parity with JSON, a
`ParseRefTarget` table (10 cases incl. hierarchical names and a bad escape), and
nil/empty. `TestRefIndex_TargetsResolveInTheSpecIndex` closes the loop — every
target must be a pointer `SpecIndex` can actually resolve to a line.

*Note:* `Model.refIndex` is wired in `refreshSpec` but not yet read — **S11**
consumes it.

### S11 — D2: `F3` / `Shift-F3` cycling + `Enter` go-to-definition — ✅ DONE

- ✅ **Find-references.** `F3` / `shift+F3` step through the reference sites of
  the node under the spec cursor, wrapping; a backward step into a *fresh* cycle
  enters at the last site. Persistent `REFS` status badge:
  `ref 2/3 of /definitions/User → /definitions/Team/properties/lead`.
  - **Cycle continuity rule:** a cycle continues only while the viewport is
    still where the last jump left it (`refExpectTop`). Scroll away and the next
    `F3` re-anchors on wherever you now are. Without this, "F3 repeatedly" would
    chase the definition of whatever it last landed on instead of walking one
    definition's uses.
  - `RefIndex.RefsNear` walks up to the nearest *referenced* ancestor, so the
    cursor need not sit exactly on the definition line — same segment-trim idiom
    as `SourceIndex.PositionFor`.
  - Reset on every render replacement (rescan, format toggle) plus Esc, entering
    follow mode, search and options.
- ✅ **Go-to-definition.** `Enter` on a `$ref` follows it. Local `#/…` only:
  an external target is *named* (`external ref, not in this spec: …`) rather
  than guessed at, and the viewport holds.
- ✅ **Bindings.** ⚠️ **`shift+F3` is terminal-dependent, and this is a bubbletea
  v1 limitation, not a choice.** `tea.Key` has **no Shift modifier**; the xterm
  family maps shift+F1..F12 onto F13..F24, so shift+F3 arrives as **`KeyF15`**
  (`\x1b[1;2R`). `key.ShiftF3 = "f15"` therefore, with `key.ShiftF3Named =
  "shift+f3"` also accepted so a terminal (or a future bubbletea) reporting the
  modifier directly still works. A terminal emitting nothing distinguishable for
  shift+F3 simply has no prev key — forward cycling still wraps. If that bites,
  the cheap fix is to let `n`/`N` step the ref cycle when one is active and no
  search is running (they already mean next/prev, and search takes precedence).
- ✅ **Dispatch.** New `handleRefNav`, mirroring `handleDiagNav`: gated on the
  spec pane and returning `handled=false` elsewhere, so `Enter` still opens a
  file in the tree. *Not cosmetic* — inlining these three cases pushed
  `handleKey` to cyclomatic complexity 32 (> 30) and tripped `gocyclo`;
  extracting them cleared it.

**Tests:** new `model_refs_test.go` — forward/backward/wrap, fresh-cycle entry
point, cycling from *inside* a definition, re-anchoring after a scroll, the two
distinct misses (unindexed line vs nothing-references-this), both key spellings,
go-to-definition plus its external and no-ref edges, spec-pane-only dispatch,
and reset across all five triggers. Mutation-verified: forcing the cycle to
always continue, and dropping `RefsNear`'s ancestor walk, each fail a test.

**Bug found by the tests:** on a failed re-start the *previous* cycle's status
lingered — the badge kept reading `ref 1/3 of /definitions/User` while the user
was looking at something else. `cycleRefs` now drops the old cycle before
attempting a new one.

⚠️ **Surfaced, not fixed — the spec pane has no line cursor.** "The node under
the cursor" means *the node at the top of the viewport*, the convention `f`
follow and the status line already use. It is consistent, but it is more
noticeable here: to follow a `$ref` you must scroll it to the top rather than
point at it. → [B-spec-line-cursor](#b-spec-line-cursor--backlog).

### S12 — D3: end-to-end tests + module README — ✅ DONE

The planned "fixture with a type `$ref`'d from ≥2 sites" was already covered by
S10/S11 — on **hand-written JSON**. The real remaining gap was that nothing
verified the chain against what codescan *actually emits*. So S12 scans the
petstore fixture for real (`doScan` → `Update(scanResultMsg)` → the finished
model) in a new `model_e2e_test.go`.

- ✅ **7 e2e tests:** every recorded site is really a `$ref` pointing where the
  index claims, ordered by line, with an addressable holder node · response
  refs (`/responses/genericError`) index alongside definition refs · the F3 →
  Enter round trip · one full lap visits every site exactly once then wraps ·
  JSON and YAML find the **same** holder set · ref targets also resolve through
  the provenance index to a `.go` file · spec→source follow opens the file that
  really declares the type (asserts `type Pet struct` is in the buffer).
- ✅ **`cmd/genspec-tui/README.md`** — the module had none. Covers install/flags
  (verified against `-h`), the layout, a complete keymap per pane, how the
  cross-ref navigation works, the internal package map, and an explicit
  **"Honest limits"** section (anchors-only, stale-while-dirty, site-index-not-
  resolver, no line cursor, terminal-dependent `shift+F3`). `markdown_lint` and
  `markdown_links` clean.

**⚠️ Real bug, found only by the e2e test — F3 then Enter was broken.**
`cycleRefs` **centres** its target (S8), so after a jump `TopLine()` is no
longer the line jumped to — and `gotoDefinition` read `TopLine()`. The headline
Phase-D workflow ("find a use, then jump back to the definition") answered
*"no $ref on this line"*. S11's unit test missed it because it scrolled straight
to the ref line instead of arriving there via F3.

*Fix:* `specCursorLine()` — the line the last jump landed on **while the
viewport still sits where that jump left it**, else the viewport top. Scrolling
invalidates it implicitly, with no bookkeeping on the scroll path. This
generalises (and replaces) S11's `refExpectTop`, and both `startRefCycle` and
`gotoDefinition` now read it. `driveSpecToSource` deliberately keeps reading
`TopLine()` — it is driven *by* scrolling, and reading a mark it sets itself
would freeze follow mode. Pinned at both unit and e2e level; mutation-verified.

*Also:* the cursor uses a `specCursorSet bool` rather than a `-1` sentinel, so
the **zero value of `Model` means "no cursor"** — the `-1` form silently gave
every hand-built test fixture a bogus cursor on line 0.

**⚠️ Test-time regression caught and fixed.** One real `packages.Load` per e2e
test took the `ux` package from ~1s to **36s under `-race`**. The scan is now
cached per package behind a `sync.Once` (the same thing the library's `scantest`
does, for the same reason): back to 6.5s. Worth remembering before adding more
e2e tests — cache first, don't scan per test.

### S13 — gutter dots (§6.5) — ✅ DONE

Discoverability: mark which lines actually lead somewhere, now that all three
indexes are live.

- ✅ **Two markers, not one.** `•` (`panels.GutterAnchor`) = this node has a
  source position of its **own**; `→` (`panels.GutterRef`) = a followable
  `$ref`. Both link kinds now exist, and they answer different questions
  ("where did this come from?" vs "where does this go?"), so one glyph would
  have conflated them. A `$ref` line wins the column on collision — the `$ref`
  member is never itself anchored, so in practice they are disjoint.
- ✅ **Deviation from the plan's wording — only EXACT anchors are marked.** The
  plan said "lines whose pointer resolves to a source position". Nearly every
  line resolves to *something* through nearest-ancestor, so that rule would dot
  the entire document and say nothing. A dot therefore means *"following this
  lands exactly here"*, which is the distinction worth drawing.
  `TestGutter_OnlyExactAnchors` pins it, and asserts the unmarked node really
  does still resolve — otherwise the distinction would be vacuous.
- ✅ **External `$ref`s are not marked** — `Enter` cannot follow them, and a
  marker would promise a jump that fails.
- ✅ **Zero cost when there is nothing to say.** A nil/empty gutter renders no
  column at all in either pane, so the layout is unchanged before the first
  scan (and with provenance off). The spec side is driven **from the source
  index** (`AnchoredPointers` → `LineForPointer`) rather than by walking the
  spec, which needed no new `SpecIndex` iteration surface.
- ✅ **Render order matters:** the gutter is prefixed *after* search/xref
  highlighting, so the styles wrap the text the user searched for rather than
  the marker column. Match counting is unaffected (it reads the raw line).
- New API: `SourceIndex.AnchoredPointers` / `AnchorLines`,
  `RefIndex.LocalRefLines`, `Spec.SetGutter`, `FileView.SetAnchors`,
  `theme.Gutter`. Rebuilt from `refreshSpec` (rescan + format toggle) and on
  file load.

**Tests:** `panels/gutter_test.go` (markers land, unmarked lines pad to keep
alignment, no-gutter costs no width, coexistence with search + xref styling,
the viewer's 1-based anchor keying) and `model_gutter_test.go` (anchors vs refs,
external refs unmarked, nil when nothing links, exact-anchors-only, rebuilt on
format toggle, anchor set follows the open file). Plus
`TestE2E_GutterMarksNavigableLines`: against a real petstore scan **every**
marked line must genuinely be navigable — anchors resolve to source, refs
resolve to a rendered node — and the gutter must mark strictly fewer lines than
the document has, since a marker on every line says nothing. Mutation-verified:
marking every resolvable node fails three tests.

README updated with the gutter legend.

### B-spec-line-cursor — ✅ DONE (`2fea048`)

**Gave the spec pane a real line cursor.** `↑↓/jk` move it, `PgUp`/`PgDn` a
page, `Home`/`End` to the ends, and the wheel moves it too — the view follows
the cursor rather than leaving it off screen where `F3`/`Enter` would act on a
node nobody can see. Rendering reuses S8's roles unchanged: strong bar when the
pane is focused (it drives), muted tint when not (it mirrors).

**It simplified more than it added.** `xrefLine` and "the cursor" were two names
for one idea, so `MarkLine` / `HighlightLine` / `ClearHighlight` collapsed into
`SetCursor` (incremental, minimal scroll) and `JumpTo` (centred), and S12's
`specCursorLine` pseudo-cursor — the workaround for not having a real one — is
gone entirely. Search now parks the cursor on the match, so `/` then `F3` or
`Enter` composes. `routePaneKey` was extracted to keep `handleKey` under the
`gocyclo` threshold and to put "which pane gets first refusal" in one place.

*Original problem statement below.*

⬜ ~~Give the spec pane a real line cursor.~~

Surfaced by S11. The spec pane has only a scroll offset, so everything that
needs "the node the user means" reads `TopLine()` — `f` follow, the status-line
node readout, and now `F3` find-references and `Enter` go-to-definition.

- **Consistent, but increasingly awkward.** Scrolling a node to the top to act
  on it is tolerable for follow mode (where you are scanning anyway). It is
  noticeably worse for go-to-definition: you must scroll a `$ref` to the very
  top of the viewport to press Enter on it, rather than pointing at it.
- **The source viewer already solved this** — `FileView` has `navLine` plus
  `NavUp`/`NavDown` and the driver/follower styling from S8. The spec pane wants
  the same shape: a cursor line, `↑↓/jk` to move it, and every "node under the
  cursor" reader switching from `TopLine()` to it.
- **Touches:** `driveSpecToSource`, `startRefCycle`, `gotoDefinition`, the
  status line, and `Spec.MarkLine`/`HighlightLine` (the cursor and the xref mark
  become two different lines, needing two styles — S8 already provides them).
- **Interacts with S13's gutter dots** (same render path) and with
  [B-rescan-anchor](#b-rescan-anchor--backlog) (a cursor is a better anchor to
  preserve across a rescan than the viewport top).

### B-viewer-global-keys — ✅ DONE (`5b291f2`)

**Decided: it passes them through.** `handleViewerKey` now mirrors
`handleDiagNav` — it shadows only the keys it genuinely owns (nav line, follow,
edit, back-to-tree) and reports `handled=false` for the rest, so `/`, `o`, `r`,
`g` and the format toggle work while a file is open. `Tab`, `c` and `ctrl+q`
were duplicated verbatim in both handlers; the global copies now serve the
viewer too. The pinning test was replaced by one asserting both halves: the
globals reach the viewer, and the viewer's own keys still belong to it.

*Original problem statement below.*

⬜ ~~Decide whether the read-only viewer should pass unhandled keys through.~~

Found while writing S9's exit tests. `handleViewerKey` ends in a bare
`return m, nil`, so it consumes every key it does not explicitly own. While a
file is open in the left pane, that silently disables `/`, `o`, `r`, `g`,
`ctrl+j` and `ctrl+y`.

- **Possibly deliberate.** The viewer's status line advertises only its own keys
  (`↑↓/jk`, `f`, `i`, `esc`, `tab`, `c`), and bare-key commands in a text pane
  are exactly the ambiguity the read-only/edit split was introduced to avoid.
- **But the gap is real for the format toggle and rescan.** Reading source is
  precisely when you want to flip JSON↔YAML or re-run the scan; today you must
  Esc back to the tree first. `handleDiagNav` shows the other pattern — it
  returns `handled=false` and lets the global bindings run.
- **The fix, if wanted,** is to mirror `handleDiagNav`: fall through instead of
  swallowing, then decide per key which ones the viewer should shadow.
- `TestJoin_ViewerSwallowsGlobalKeys` pins the current behaviour, so changing it
  is a deliberate act with a failing test to update.

### B-rescan-anchor — ✅ DONE (`2fea048`)

**Preserved.** `refreshSpec` captures the pointer under the cursor before
rebuilding and restores it after, which also subsumes the format-toggle handling
from S7. The open design questions resolved as:

- **Which anchor?** The cursor — which is why this waited on B-spec-line-cursor.
- **Node deleted?** Fall back to its nearest surviving ancestor, so you land in
  the right neighbourhood rather than somewhere arbitrary. If nothing on its
  path survives, the clamped line `SetContent` chose stands.
- **Centre or not?** **Minimal scroll.** This is the hot path — every save fires
  it, and after a typical rescan the node has moved a line or two and is still
  on screen, so yanking the viewport each time would be worse than the drift it
  fixes. An explicit format switch *does* recentre, being a deliberate change of
  view. Both are tested.

Mutation-verified: disabling the restore fails six tests.

*Original problem statement below.*

⬜ ~~Preserve the spec node across a rescan, not just across a format toggle.~~

Found while doing S7; **deliberately not folded into any slice** — Fred's call to
park it.

The *same bug class* S7 fixed for `ctrl+j`/`ctrl+y` also lives on the rescan
path: `scanResultMsg` → `applyScan` → `refreshSpec` → `Spec.SetContent` rebuilds
`m.specIndex` while the viewport keeps its old `YOffset`. If the regenerated spec
gained (or lost) lines *above* where the user is reading — a new definition, a
new path, a widened description — the pane silently slides to a different node.

- **Why it matters more than the toggle did.** This is the hot path. Every save
  fires it, and live-reload-on-save is the tool's whole reason to exist (memory
  `project_genspec_tui`: the live-render loop is "the killer feature"). The
  toggle is occasional; this happens continuously while editing.
- **Why it's parked rather than done.** The mechanical fix is one call to the
  `setSpecFormat` helper's logic (capture `PointerAt(TopLine())` → rebuild →
  `LineForPointer` → `ScrollTo`). But it changes the *feel* of the main loop, and
  there are real judgment calls underneath: should the anchor be the top-of-
  viewport node or the follow/xref node when one is active? What happens when the
  anchored node disappeared from the new spec (a deleted type) — hold the line
  number, jump to the nearest surviving ancestor, or stay put? Should it apply
  when the user hasn't scrolled at all (`YOffset == 0`, where "preserve" and
  "stay at the top" agree, but only by accident)? Those want manual testing on a
  real edit loop, not a plausible-looking patch.
- **Interacts with:** S13's gutter dots (same render path) and the follow-mode
  re-sync already wired into `scanResultMsg`.

### Out of this series

`LX-model` (library-owned derived model), `x-go-origin` materialization,
LSP token-level positions, `LX-prov-2` (precise inside-YAML positions, gated on
the goccy swap), and the roadmap's `Repro-pack` / `Map-vis` items.

---

## Suggested PR sequence

1. **PR-A** = Phase A (`LX-spec-0`). TUI-only, go1.26 bump, self-contained,
   testable; status-line pointer demo. Ships independently.
2. **PR-B** = Phase B (`LX-prov-0`). codescan-only, additive API, callback-tested.
   No TUI change. (Parallelizable with PR-A.)
3. **PR-C** = Phase C (`LX-join`). The visible feature; needs A + B merged.
4. **PR-D** = Phase D (`LX-refs`).

All four land on the single `feat/genspec-tui` branch (A/B/C already committed
there); the split above is the review unit, not four separate branches. Series 2
adds S7–S13 on top, one commit per slice, squashed by logical unit before review.

---

## Decisions to confirm

- ~~**D1 — go1.26 + `GOEXPERIMENT=jsonv2` for the TUI.**~~ **MOOT** since
  `ca0653c`. *Historic:* settled (Fred) as "ship `jsontext` via the experiment
  flag, warn self-builders". The `go-openapi/core/json` lexer swap removed the
  dependency entirely — no experiment flag, no toolchain pin, no release-pipeline
  constraint, and the TUI module keeps a plain `go 1.25.8` directive.
- ✅ **D4 — go-to-definition in scope.** Settled 2026-07-30 (Fred): `Enter` on a
  `$ref` line jumps to its definition (S11), alongside the designed
  `F3`/`Shift-F3` find-references.
- ✅ **D5 — gutter dots in scope.** Settled 2026-07-30 (Fred): §6.5's optional
  discoverability item ships as the closing slice (S13).
- ✅ **D2 — `Provenance` home.** Settled in Phase B (see B0): the type lives in
  `internal/scanner/provenance.go`, beside `Options` — the recommended option.
  The TUI names it as `scanner.Provenance`, which works because a submodule
  under `codescan/` may import the parent's `internal/` tree.
- ✅ **D3 — coarse positions** for routes/ops/meta (grammar block `Pos()`) and
  line-granularity overall — settled in design §3.5.

## References

- Design: [`genspec-tui-linkage.md`](genspec-tui-linkage.md) — §3 (source side),
  §4 (spec side), §5 (join), §6 (UX), §7 (milestones), §9 (recon citations).
- Pattern to mirror: the diagnostics wiring (`b2574fe`) and `OnDiagnostic`
  plumbing (`internal/scanner/options.go`, `scan_context.go`, `common/builder.go`).
- [[project_genspec_tui]] · [[project_lsp_diagnostics_target]].
