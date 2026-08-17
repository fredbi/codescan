---
title: Prune unused models under -m
stream: 9
origin: iii
status: done
release: v0.35
issues: [go-swagger#2639]
prev: "§12"
---

# Prune unused models under `-m`

**Status:** ✅ done (2026-06-22) · merged PR #50 (precursor `83108cb`, feature `9b4fc1f`, doc-site `cc31875`).

`Options.PruneUnusedModels`: with `ScanModels` (`-m`), prune discovered
definitions not transitively reachable from a path / response / parameter /
overlay root — a middle ground between "all models" (`-m`) and "route-reachable
only" (no `-m`).

**Key property.** Pruning runs **before** `reduceDefinitionNames`, so an unused
model can't force a spurious cross-package collision rename on a used one (the
survivor keeps its bare leaf name). `InputSpec` defs are pinned and seeded as
roots. Each drop raises a `scan.pruned-unused` Hint; each collision rename a
`scan.renamed-definition` Hint.

**Precursor bug fixed in the same branch.** Definition provenance is now buffered
and re-pointed through name reduction — it was dangling on any cross-package
collision (`OnProvenance` is a push stream; def anchors emitted under short names
went stale after a rename).

**Origin.** go-swagger#2639 — a large shared library scanned with `-m` emits all
`swagger:model` types; the reporter wants only the `$ref`-referenced ones.
Classified works-as-designed + need-doc; this is the optional enhancement.

**Fixture:** `fixtures/enhancements/prune-unused`. **Discriminator-mapping
reachability** deferred → see [discriminator-subtype-discovery](discriminator-subtype-discovery.md).
