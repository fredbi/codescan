# Discriminator subtype discovery — execution plan

**Branch:** `polymorphic-subtypes-discovery` (worktree `.worktrees/feat/polymorphic-subtypes-discovery`)
**Design:** `.claude/plans/features/discriminator-subtype-discovery.md` (stream 9, §15, v0.37,
go-swagger#1913)
**Status:** ✅ P1–P10 done (2026-07-30). Two commits on the branch, nothing pushed, awaiting review:
`5e03328` feature + `3bd0c93` doc-site.

**Gating decision (2026-07-30, Fred):** always on — no new `Options` field. A definition carrying
`discriminator` with none of its subtypes in the document is an incomplete polymorphic family, so
closing the gap is a fix, not a preference. Consequence: default output changes for discriminated
bases only; no `options.md` row on the doc-site.

---

## 1. Grounded starting state

Probed on the new fixture (`fixtures/enhancements/discriminated-subtypes`), pre-change:

| mode | definitions emitted |
|------|---------------------|
| no `-m` | `TeslaCar` **only** — base carries `discriminator: model`, family incomplete |
| `-m` | `Battery PlainBase PlainSub TeslaCar Unrelated modelS modelX modelY` (over-generates) |
| `-m` + `PruneUnusedModels` | `TeslaCar` **only** — the subtypes are *pruned* |

Two distinct holes, both from the same asymmetry: **subtypes `$ref` the base, nothing `$ref`s the
subtypes**, so neither the discovery walk nor the prune reachability walk can ever reach them
top-down.

Relevant machinery:

- `TypeIndex.Models` is populated by classification **regardless of `ScanModels`** — without `-m` the
  subtype decls are indexed, just never built. So the reverse index needs no new scanning pass.
- All definition building funnels through `spec.Builder.buildDiscoveredSchema`; the
  `buildDiscovered` loop already iterates to fixpoint over `s.discovered`. `buildRoutes` /
  `buildOperations` only *read* `s.definitions`, so the loop is the single insertion seam.
- `pruneUnusedModels` (`prune.go`) marks roots then walks `$ref`s via `collectDefRefs`. Definitions
  are deliberately **not** roots.
- Discriminator presence is observable post-build as `s.definitions[key].Discriminator != ""` —
  type-agnostic (works for an interface base with a `discriminator: true` member and for a struct
  base with a `discriminator: true` field alike). Preferred over re-deriving it from source.

## 2. Design

### 2.1 The reverse index (`internal/builders/spec/subtypes.go`)

`subtypeIndex: base type identity → []*scanner.EntityDecl`, built once, lazily, from
`ScanCtx.Models()`:

- a candidate subtype is a `swagger:model` **struct** decl;
- for each anonymous (embedded) field whose doc carries `swagger:allOf` and not `swagger:ignore`,
  resolve the embedded type through `TypesInfo` (unwrapping pointer / alias) to its
  `*types.TypeName`;
- key = `<pkgpath>.<TypeName>` — the **Go type** identity, not the swagger name: it is what both
  ends (subtype embed site and base decl) can compute without knowing the other's annotations.

Entries are sorted by `DefKey()` so pull order (and therefore any emitted diagnostic order) is
deterministic despite the map iteration in `Models()`.

Only explicit `swagger:allOf` counts. `DefaultAllOfForEmbeds` (which promotes a *plain* embed to
allOf composition) is deliberately **not** honoured here: the polymorphic idiom is the explicit
annotation, and letting a global rendering knob change *which types exist* would be a surprising
coupling.

### 2.2 Hook A — discovery (`spec.go`)

In `buildDiscoveredSchema`, right after the schema is built and its post-declarations queued:

```go
s.discovered = append(s.discovered, sb.PostDeclarations()...)
s.discovered = append(s.discovered, s.discriminatedSubtypesOf(decl)...)
```

`discriminatedSubtypesOf` returns the index entry for `decl` **only when the definition just built
carries a discriminator**. Because this feeds `s.discovered`, the existing fixpoint loop handles the
rest: a pulled subtype is built like any discovered decl, its own dependencies (`Battery`) are
discovered from it, and a subtype that is itself a discriminated base cascades.

`decl.DefKey() → base type identity` is recorded in a new `declIdentity` side table (sibling of
`declPos`), so the prune hook can go from a definition key back to the index without re-deriving
decls.

### 2.3 Hook B — prune reachability (`prune.go`)

Inside the transitive-closure loop, a **reachable discriminated base marks its subtypes reachable**:

```go
sch := s.input.Definitions[cur]
collectDefRefs(&sch, mark)
if sch.Discriminator != "" {
    for _, key := range s.subtypeKeysOf(cur) { mark(key) }
}
```

This is the "discriminator-mapping reachability" item `prune-unused-models.md` deferred to this
feature. Note it is *not* redundant with hook A: under `-m` every model is built up front, so the
prune is where the family is preserved.

### 2.4 Diagnostics

New `grammar.CodeDiscoveredSubtype = "scan.discovered-subtype"` (Hint, located at the subtype's Go
decl), one per pull, so "where did this definition come from?" is answerable — mirroring
`scan.pruned-unused` / `scan.renamed-definition`. Only fired for a genuine pull (a subtype already
in `definitions` is silent).

## 3. Phases

| # | Phase | Status |
|---|-------|--------|
| P1 | Fixture `fixtures/enhancements/discriminated-subtypes` (+ `base/`, `sub/`) with both negative controls; probe pre-change behaviour | ✅ done |
| P2 | `subtypes.go`: reverse index + `swagger:allOf` embed classification + `discriminatedSubtypesOf` | ✅ done |
| P3 | Hook A (discovery) + `declIdentity` + `scan.discovered-subtype` Hint | ✅ done |
| P4 | Hook B (prune reachability) | ✅ done |
| P5 | Integration witness (`coverage_discriminated_subtypes_test.go`, 3 modes) + golden | ✅ done |
| P6 | Unit tests for the index (cross-package, negative controls, pointer embed) | ✅ done |
| P7 | Regression sweep (full suite + golden drift review) + `bugs/1913` residual-gap comment refresh | ✅ done |
| P8 | Docs: `internal/scanner/README.md` §subtypes (+ §prune rewrite), CLAUDE.md notable-design note | ✅ done |
| P9 | Nested-hierarchy fixture + witness (3 modes, 2 goldens); interface-embed indexing + `isDiscriminated` gate | ✅ done |
| P10 | Doc-site: tutorial sections "How subtypes are discovered" + "Multi-level hierarchies", two witnesses, type-discovery/prune/ROADMAP consistency edits; verified by a full hugo build | ✅ done |

## 3b. Deviations from §2, found in execution

- **The pull is a no-op under `ScanModels`.** `buildModels` builds every annotated model up front, so
  pulling adds nothing there — and pulling anyway emitted `scan.discovered-subtype` Hints in *model-
  index map-iteration order* (caught by the `-m` witness: 2 Hints, unstable). `discriminatedSubtypesOf`
  now returns early when `scanModels` is set; the `-m` path is served by hook B alone, which is exactly
  where a family can be lost under `-m`.
- **Alias embeds are indexed under two identities** (the alias and the type it names), since which one
  the emitted `allOf` member `$ref`s depends on `RefAliases` / `TransparentAliases`. Identities per
  decl are deduped so a struct reaching one base twice (type + alias) is not listed twice.
- **`internal/builders/spec` has no README**; the long-form docs went to
  `internal/scanner/README.md#subtypes` instead, next to the `§prune` section they interlock with
  (whose "Known limitation: subtypes could be pruned" paragraph this change deletes).
- **The `edges/` fixture package was added** beyond the original fixture plan: pointer embed, alias
  embed, ignored embed, plain embed, and a struct base carrying the discriminator on a property —
  in a family no route references, which also locks that an *unreached* discriminated base pulls
  nothing (the pull is gated on reachability, not on existence).

## 3c. Nested hierarchies (added 2026-07-30, Fred's request)

A second fixture — `fixtures/enhancements/discriminated-subtypes-nested` — covers a subtype that is
itself a discriminated base:

```text
Shape (discriminated)          <- the only referenced type
 |- Circle                     (leaf struct)
 `- Polygon (discriminated)    <- subtype AND base
     |- Square                 (leaf struct)
     `- Triangle               (leaf struct, pulls Coords by ordinary $ref)
```

Probing it found **two genuine gaps** in the P2–P4 implementation, both now fixed:

1. **Interface embeds were not indexed.** `allOfBases` only read `*ast.StructType.Fields`, so
   `Polygon` (an interface embedding `Shape` under `swagger:allOf`) was never a subtype — and the
   interface is exactly how a mid-level type is written in the go-swagger idiom. Now
   `embeddableMembers` reads a struct's fields *or* an interface's method list (an embed is the entry
   with no `Names` in both shapes).
2. **A mid-level base's discriminator is not at the top level.** It emits as
   `allOf: [{$ref Shape}, {own props, discriminator: polygonType}]` — properties (hence the
   discriminator) land in the compound member. The `sch.Discriminator != ""` gate therefore missed it
   and the cascade stopped at level 1. Replaced by `isDiscriminated(sch)` (top level OR any allOf
   member), shared by both hooks. The `$ref` member is deliberately not followed, so a leaf never
   inherits its base's discriminator and never pulls its own siblings — locked by
   `TestIsDiscriminated`.

Cascade result, no `-m`: `Shape Circle Polygon Square Triangle Coords`, with 4
`scan.discovered-subtype` Hints (2 per level, each naming its base) and `Unrelated` excluded. Under
`-m` + prune the same 6 survive — the prune rule has to fire on the mid-level base too.

## 4. Verification contract

Target on the fixture:

| mode | expected |
|------|----------|
| no `-m` | `TeslaCar`, `modelS`, `modelX`, `modelY`, `Battery` — and **not** `Unrelated`, `PlainBase`, `PlainSub` |
| `-m` | unchanged from today (all models) |
| `-m` + prune | same set as no `-m` |

Plus: every subtype keeps the locked `allOf: [{$ref base}, {own props}]` shape
(`TestCoverage_Bug1913` stays green, unchanged), and no other golden in the suite drifts —
any drift means a base somewhere in the corpus is discriminated and its family was incomplete,
which must be reviewed case by case, not blanket-accepted.
