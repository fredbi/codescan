# Prune unused models — `Options.PruneUnusedModels`

Date: 2026-06-22
Status: ✅ MERGED to master — PR #50 (merge 3f0612e); 83108cb (provenance) + 9b4fc1f (prune) + cc31875 (doc-site page). Feature worktree now stale (cleanup pending).

Doc-site: prose page `docs/doc-site/shaping-the-output/pruning-unused-models.md`
(weight 16) landed on the FEATURE branch (not stale doc-site-update), per Fred —
example test would need the option in the module, so docs travel with the feature.
doc-site-update branch is stale; needs cleanup (separate task, not done).
Owner: Fred (sponsor) + agent
Branch: `feat/prune-unreferenced-models` (worktree `.worktrees/feat/prune-unreferenced-models`, off `master` @ `b1ce919`)
Origin: forthcoming-features.md §12 · go-swagger#2639
Roadmap: Stream 9 V-flag delivery mode (non-breaking `Options.X bool`, default off)

> Option name is `PruneUnusedModels`; branch name (`…unreferenced…`) is cosmetic.

---

## Problem & motivation

`-m` / `Options.ScanModels` emits **all** `swagger:model` types, reachable or
not. The default (no `-m`) emits only models transitively reachable from
routes/responses/parameters. go-swagger#2639 wants the middle ground: run `-m`
discovery, then prune what no path references.

**Why this matters (user-perspective, the real driver).** Pruning is tightly
tied to **conflict resolution**. Today users work around the missing feature by
running go-swagger's `flatten` command (which knows how to prune) — *precisely
to avoid spurious definition-name conflicts that should never have surfaced*.
When two same-short-named models from different packages are both emitted under
`-m`, name reduction deconflicts them (`pkgA.User`/`pkgB.User` →
`PkgAUser`/`PkgBUser`) even when one of them is unused. Prune the dead one
*first* and the collision evaporates — the survivor keeps its clean short name.
So **prune-before-name-reduction is the feature**, not an implementation nicety.

## Solution shape

New knob `Options.PruneUnusedModels bool` (default false), a **modifier on
`-m`**: run `ScanModels` discovery, then prune any definition not transitively
referenced from a root. With `ScanModels=false` it is a **no-op** (set already
reachable-only) — emit one Hint if set without `ScanModels`. Non-breaking:
flag-off output is byte-identical.

## Settled design decisions (reviewed 2026-06-22)

1. **Roots = paths + shared `responses` + shared `parameters`** (+ overlay).
   A definition survives only if reachable from a root; a model referenced only
   by another *unreferenced* model is pruned. Shared responses/params aren't
   produced by codescan yet but will be, and may arrive via `InputSpec`.
2. **`InputSpec` definitions are pinned** (never pruned) **and seeded as roots**
   (their transitive `$ref`s keep targets alive). Prune what we *discovered*,
   not what the caller handed us.
3. **Discriminator/polymorphism reachability deferred to §15.** A discriminator
   base references subtypes via mapping *strings*, not `$ref`s; a subtype
   reachable only that way could be wrongly pruned. codescan doesn't auto-wire
   discriminator subtypes today (§15). Document the limitation; do **not**
   special-case now.
4. **Name:** `Options.PruneUnusedModels`.
5. **Visibility:** one **Hint** per pruned definition via `OnDiagnostic`, new
   code `scan.pruned-unused`. Diagnostic carries an accurate source
   `token.Position` (the decl site), independent of provenance (§ below).

## Pipeline placement — the critical ordering

```
buildModels → buildParameters → buildResponses → buildDiscovered
  → buildRoutes → buildOperations → buildMeta        (all roots populated)
  → emitParameterAnchors
  → pruneUnusedModels()        ← NEW: fully-qualified namespace, graph consistent
  → reduceDefinitionNames()    ← deconflict over survivors only
  → flushDefinitionOrigins()   ← NEW: provenance, post-reduce, final pointers
```

Prune runs **before** `reduceDefinitionNames` so deconfliction only ever sees
survivors (the motivation above). At that point the graph is still in the
consistent fully-qualified `#/definitions/<pkgpath>/<name>` namespace, so the
reachability walk is unambiguous.

## Prune mechanism

Mirror `reduce.go`'s existing recursive schema descent (`rewriteSchemaRefs`:
`Properties`, `PatternProperties`, `AllOf/AnyOf/OneOf/Not`, `Items`,
`AdditionalProperties/Items`, nested `Definitions`) as a **read-only
reachability collector**:

1. Seed roots: `$ref` targets reachable from every operation across
   `s.input.Paths` (params, responses, bodies), shared `s.input.Responses`,
   shared `s.input.Parameters`, and pinned `InputSpec` definitions.
2. Transitive closure over `s.input.Definitions` with a `visited` set (cycles:
   self-ref / linked-list / tree models).
3. Prune: for each key not in `reachable` and not pinned — delete from
   `s.input.Definitions`, drop its buffered provenance (§), emit
   `scan.pruned-unused` Hint with source pos from `GetModel(...).Ident.Pos()`.

Ref-name extraction reuses the `#/definitions/` prefix-strip (reduce.go
`repointer`). Targets are fully-qualified at this stage. Lean: new
`internal/builders/spec/prune.go`.

## Provenance deferral (pre-existing bug, fixed in this branch)

**The bug (independent of pruning).** `OnProvenance` is a **push stream**:
`RecordOrigin` (scan_context.go:321) fires the callback *immediately* and stores
nothing. Definition anchors are emitted during `buildDiscoveredSchema`
(spec.go:263-278) using the **short** name as base path
(`JSONPointer("definitions", decl.Names())`) and threaded as `WithPath`, so the
schema builder also emits **field-level** sub-anchors under
`/definitions/<short>/properties/...`. But `reduceDefinitionNames` later renames
*colliding* definitions to `PkgAUser`/`PkgBUser` **without touching provenance**
(reduce.go has zero origin handling). Result for cross-package collisions: every
emitted `/definitions/<short>...` pointer is left **dangling** (and the two
colliding decls even collide with each other in the stream). The
`TestCoverage_ProvenanceGeometry` invariant ("every pointer resolves") would
fail — it just lacks a cross-package-collision fixture today, so it's latent.

Pruning would add a second source of dangling pointers (anchors for deleted
defs). Both share one fix.

**Fix: buffer + flush (no mutating interface).** Per Fred: buffer rather than
define an interface that mutates later.

- During build, definition-scoped origins (the def node **and** all its
  field-level sub-anchors) are **buffered**, keyed by the **fully-qualified**
  def key, instead of fired inline. Non-definition anchors (paths, info,
  responses, params) keep firing inline — `reduce` never renames them.
  - Mechanism: a per-definition buffering window on `ScanCtx` (mirrors the
    existing `paramOrigins` deferral). Around `sb.Build()`, set the current
    def-key bucket; `RecordOrigin` routes `/definitions/...` pointers into that
    bucket (full pointer + pos), everything else fires immediately.
  - Base path for the schema builder switches from short name to the
    fully-qualified def key so buffered pointers are unambiguous.
- `pruneUnusedModels` drops the buckets for pruned keys (so no orphan is ever
  flushed).
- `flushDefinitionOrigins(renames)` runs **after** reduce: for each surviving
  bucket, rewrite the `/definitions/<fqn>` segment → final reduced name (reusing
  the `renames` map reduce already computes; leaf name when no collision) and
  fire `OnProvenance`. Every flushed pointer now resolves against the final
  document.

**Invariants (Fred):**
- Every pointer handed to `OnProvenance` is an **actual node** in the final
  document — definition-level **and** field-level (TUI locates nodes + source).
- A **pruned** definition (and, if we add it, a **renamed** definition) emits
  **no orphan `OnProvenance`** record.
- The **diagnostic** for a pruned/renamed def keeps an **accurate source
  location** (decl `token.Position`) regardless — diagnostics ≠ provenance.

**Rename Hint (in scope).** When `reduceDefinitionNames` renames a definition
for collision, emit a `scan.renamed-definition` Hint so the TUI/user learns
`pkgA.User → PkgAUser`. Carries the decl's source `token.Position` (via
`GetModel(fqn).Ident.Pos()`); message names the originating Go type and the
final spec name. Diagnostic only — creates **no** provenance record (the renamed
node's provenance is the normal flushed anchor under its final name).

## Open implementation items

- **O1 — buffering window plumbing.** Cleanest hook for "we're now emitting
  origins for def-key X" around `sb.Build()` in `buildDiscoveredSchema`; verify
  no def-scoped origins are emitted outside that window.
- **O2 — prefix remap correctness.** fqn keys contain `/`; remap by
  bucket-known fqn (`strings.Replace(ptr, "/definitions/"+fqn, …, 1)`), not by
  global longest-match.
- **O3 — non-definition origins.** Confirm responses/routes/operations/meta
  anchors never need post-reduce remap (they shouldn't — reduce only renames
  `/definitions/*`).

## Tasks

- ⬜ T1 — `Options.PruneUnusedModels` + godoc (modifier-on-`-m`, no-op without,
  pinned `InputSpec`).
- ⬜ T2 — diagnostic codes `scan.pruned-unused` **and** `scan.renamed-definition`.
- ✅ T3 — **Provenance deferral** (precursor, own commit): buffered def-scoped
  origins fqn-keyed on `ScanCtx` (`BeginDefOrigins`/`EndDefOrigins`/
  `RecordOrigin` buffering/`DropDefOrigins`/`FlushDefOrigins`); def base path →
  fqn (`DefKey`) in `buildDiscoveredSchema`; `reduceDefinitionNames` returns the
  `renames` map; flush after reduce in `Build()`. Witness: reused
  `name-identity-mixed` + `name-identity-3way` (Item→XItem/YItem, Widget→…) in
  `TestCoverage_ProvenanceGeometry` + focused `TestCoverage_ProvenanceCollisionRename`.
  **Proven**: 8 dangling anchors pre-fix (stash-test), all resolve post-fix; full
  suite + lint green. (`DropDefOrigins` wired by T4.)
- ✅ T1 — `Options.PruneUnusedModels` + godoc; `ScanCtx.PruneUnusedModels()`.
- ✅ T2 — codes `scan.pruned-unused` + `scan.renamed-definition`.
- ✅ T3b — rename Hint in `reduceDefinitionNames` (`diagnoseRenames`, located via
  `declPos`, sorted, skips trivial leaf-lifts).
- ✅ T4 — `internal/builders/spec/prune.go`: `pruneUnusedModels` + `rootRefs` +
  `collectDefRefs`; drops buffered origins for pruned keys; wired into `Build()`
  before `reduceDefinitionNames`; flag via `ScanCtx` accessor; `declPos` map.
- ✅ T5 — Fixture `fixtures/enhancements/prune-unused/` (+ `a`, `b`): root Used,
  chain B→C, recursive Node, dead Unused→OnlyByUnused, collision a.Thing
  (kept) / b.Thing (pruned).
- ✅ T6 — `coverage_prune_unused_test.go`: Off (8 defs, 2 renames) + On (5 defs,
  collision evaporated to bare `Thing`, 0 renames) + goldens
  `enhancements_prune_unused{,_all}.json`.
- ✅ T7 — Diagnostics (3 located Hints, 0 renames), no-op+Hint without
  `ScanModels`, InputSpec-pinned survives, provenance (no orphan, all resolve).
- ✅ T8 — `internal/scanner/README.md#prune`, CLAUDE.md Options bullet, §12 ✅ in
  forthcoming-features, §15 discriminator caveat noted.

## Test matrix (semantic intent)

| Case | ScanModels | Prune | Expect |
|------|-----------|-------|--------|
| Unreferenced `swagger:model` | on | off | emitted |
| Unreferenced `swagger:model` | on | on  | pruned + Hint, no provenance |
| Referenced model | on | on | kept |
| A→B→C, only A referenced | on | on | A,B,C kept |
| Only-by-unused (X refs Y, X unused) | on | on | X,Y pruned |
| Cycle (linked list) referenced | on | on | kept, no infinite loop |
| Collision pair, one side unused | on | on | survivor keeps clean short name |
| Pinned via InputSpec, unreferenced | on | on | kept (pinned) |
| Cross-package collision, both used | on | off | **all provenance pointers resolve** (precursor fix) |
| Flag set | off | on | no-op + Hint |

## Non-goals

- Discriminator-mapping reachability (§15).
- Pruning user-supplied `InputSpec` definitions.
- Any change to default (flag-off) *spec* output. (Provenance stream ordering
  for definitions changes — they flush at the end — but every pointer still
  resolves; this is a bugfix, and provenance is experimental surface.)

## Cross-refs

- Name-identity / reduce machinery: `.claude/plans/name-identity-cyclic-ref.md`
  (the `renames` map + collision resolution this builds on).
- genspec-tui linkage (the provenance consumer):
  `.claude/plans/genspec-tui-linkage*.md`.
```
