---
title: Infer required from field shape (opt-in)
stream: 9
origin: iii
status: open
release: v0.37
issues: [go-swagger#3275]
prev: "§16"
---

# Infer `required` from field shape (opt-in)

**Status:** ⬜ open · **approach undecided (TODO)** — the exact heuristic isn't
pinned (candidate: non-pointer && no `,omitempty`), and whether
`SetXNullableForPointers` gates it is undecided.

**Origin.** go-swagger#3275 (backlog verification, 2026-06-15). Today a property is
`required` only when its field carries an explicit `// required: true`. The
reporter has hundreds of structs and wants `required` inferred from field shape
rather than annotating each one.

**Scope.** An opt-in option that derives `required` from a heuristic — e.g. a
non-pointer field without `,omitempty` in its json tag is required; a pointer or
`omitempty` field is optional (optionally keyed off `SetXNullableForPointers`'s
nullability signal). Must stay opt-in: it would flip the required-set of every
existing spec, and the json-omitempty↔required mapping is a convention, not a
guarantee. Explicit `// required:` always wins.

**When to revisit.** When required-inference demand recurs; pairs with the
pointer-nullability handling (`SetXNullableForPointers`).
