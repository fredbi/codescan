---
title: Auto-discover discriminator subtypes of a referenced base
stream: 9
origin: iii
status: open
release: v0.37
issues: [go-swagger#1913]
prev: "§15"
---

# Auto-discover discriminator subtypes of a referenced base

**Status:** ⬜ open.

**Origin.** go-swagger#1913 (backlog verification, 2026-06-15). Interface-based
discriminator models already work: a `swagger:model` interface with a
`discriminator: true` member emits a base definition carrying `discriminator`, and
struct subtypes that embed it via `swagger:allOf` emit
`allOf: [{$ref base}, {own props}]` (locked by `fixtures/bugs/1913` +
`TestCoverage_Bug1913`). The residual gap is **discovery**: subtypes are emitted
only under `-m` (`ScanModels`), which over-generates. Without `-m`, a route that
references only the base doesn't pull in its subtypes, so the polymorphic family
is incomplete.

**Decision (groomed 2026-06-23).** Add a pass to the spec builder — **part of the
discovery loop, before pruning** — that **reverse-looks-up** subtypes which
`swagger:allOf`-refer to a discriminated parent. When a discriminated base enters
the reachable closure, pull its subtypes in too (and any types they transitively
discover).

- These subtypes **don't surface from top-down exploration**: nothing `$ref`s
  them — *they* `$ref` the base, not vice-versa — so a plain reachability walk
  misses them. Hence the reverse lookup.
- Mechanism: a **reverse `swagger:allOf` index** (subtype → base). Gated on the
  base actually being discriminated (no over-pull). No `-m` required.
- Ordering: because it runs inside discovery (before [prune-unused-models](prune-unused-models.md)),
  the newly-pulled subtypes are part of the reachable set and aren't trimmed.

**When to revisit.** Alongside [prune-unused-models](prune-unused-models.md) (shared
`-m`/discovery machinery), or when discriminator demand recurs.
