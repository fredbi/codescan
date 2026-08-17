---
title: Name-identity — deferred advanced features
stream: 9
origin: iv
status: open
release: unscheduled
issues: []
prev: "§14.1"
---

# Name-identity — deferred advanced features

**Status:** ⬜ open · **low priority (deferred)** — already designed; revisit when
alias-driven naming / qualified author refs are demanded.

**Origin.** (iv) deferred refinement on the landed name-identity / cyclic-`$ref`
engine — see [name-identity-disambiguation](archive/name-identity-disambiguation.md) as
the shipped parent. Designed-but-unbuilt enhancements left by that engine.
Cross-reference: `.claude/plans/name-identity-cyclic-ref.md` §9.3 (import
aliases), §10/W3 (many-to-one), §13/ST3 (qualified refs), §9.5+K5 (per-level
provenance).

**Scope.** Three independent refinements:

- **Import-alias candidate spellings (plan K3 + W3).** Enrich the concat candidate
  pool with the *author's import alias* for the colliding type's package, taken in
  the referencing file's context (`mongo-go-driver/mongo` imported as `mongodb` →
  `MongodbBook`, friendlier than `MongoBook`). Rides the `$ref` reachability graph:
  annotate each ref edge with the alias used at its source file
  (`ast.File.Imports`), aggregate per target identity. **Many-to-one rule (W3):**
  offer the alias only when every ref edge agrees (or there is a single site); on
  disagreement, fall back to the canonical leaf — aliases *enrich*, never *force*
  an arbitrary pick. Mechanics are fiddly (AST alias→type mapping); a
  transient-extension fallback is sketched in the plan §9.3. Tech refs:
  `go-openapi/testify/codegen/internal/scanner/import.go` (an import-alias index
  built this exact way) and the `ast-types-bridging` local skill.
- **Qualified author short-refs (plan ST3).** Author `$ref`s by short name
  currently resolve PURE-LEAF only (unique→promote, ambiguous→diagnose+drop).
  Support a dotted/qualified author spelling (`pkg.Type`) so an author can pinpoint
  one specific colliding model instead of being forced to a `swagger:model`
  override. Overlaps the alias work above.
- **Per-level `x-go-package` on hierarchical containers (plan K5).** The opt-in
  hierarchical fail-safe stamps `x-go-package` on the *innermost* container only;
  deeper container levels carry `additionalProperties:true` but no provenance. Add
  per-level provenance if a genuine multi-level hierarchical case warrants it.

**When to revisit.** When alias-driven naming or qualified author refs are
demanded, or alongside [prune-unused-models](archive/prune-unused-models.md) (pruning dead
colliders *before* reduce yields better names — a dead `b.Test` no longer taxes a
live `c.Test` into a qualified name).
