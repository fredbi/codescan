---
title: Definition-name auto-disambiguation across packages
stream: 9
origin: iii
status: done
release: v0.35
issues: [go-swagger#1734, go-swagger#2637, go-swagger#2783]
prev: "§14"
---

# Definition-name auto-disambiguation across packages

**Status:** ✅ done (2026-06-17), landed on the name-identity / cyclic-$ref track,
merged into `fix/backlog-lot1`.

Two structs with the same short name in different packages used to both map to
`#/definitions/User` and silently merge into one lossy, scan-order-dependent
definition. Now definitions are keyed by a **fully-qualified identity** during
build; a **reduce stage** projects each to the shortest deterministic name —
bare leaf → minimal-depth PascalCase concat under a readability budget → opt-in
hierarchical fail-safe — and same-package duplicates revert to their Go name with
a diagnostic.

**Origin / coverage.** Resolves go-swagger#2637 + #2783 and the duplicates
#1734 / #2126 / #2398 / #2662; cross-package leaf resolution was also extended to
the type-name keywords (#2251). The prior workaround was an explicit
`swagger:model <name>` to force a distinct key.

Design + execution log:
[name-identity-cyclic-ref.md](../name-identity-cyclic-ref.md); the
test-architecture follow-on is
[golden-unit-to-integration.md](../golden-unit-to-integration.md).

**Deferred tail.** The designed-but-unbuilt advanced enhancements (import-alias
candidate spellings, qualified author short-refs, per-level provenance) are
tracked separately in [name-identity advanced](name-identity-advanced.md).
