---
title: Per-operation field views (projections)
stream: 9
origin: iii
status: open
release: v0.37
issues: [go-swagger#1992]
prev: "§19"
---

# Per-operation field views (projections)

**Status:** ⬜ open · **low priority (deferred)**.

**Origin.** go-swagger#1992 (backlog verification, 2026-06-15). The reporter wants
to hide server-assigned fields (e.g. `Id`, `Created`) from create (POST) requests
while keeping them in responses — different field subsets of one model per
operation.

**Where we stand (groomed 2026-06-23).** #1992 is **allegedly already fixed** —
`readOnly: true` (`// read only: true`) marks a field server-owned (present in
responses, ignored by clients on write), covering the common Id/Created case
without a second Go type (witnessed by `fixtures/bugs/1992` +
`TestCoverage_Bug1992`). `writeOnly` is the dual. The original issue makes **no
mention of `readOnly`**.

**Action before any further design.** Re-test the #1992 fixture and check whether
it works with an additional `readOnly` keyword on top of the composed inner
model. Re-read the original issue in detail and explore with the TUI to determine
what — if anything — is still missing. The earlier "named field-views → derived
definitions" sketch is **parked**: not clearly warranted until that investigation
shows a concrete gap.

**When to revisit.** Low priority; revisit after the #1992 re-test, or when
per-operation projection demand recurs.
