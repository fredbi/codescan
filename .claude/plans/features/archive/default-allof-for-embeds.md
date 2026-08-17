---
title: DefaultAllOfForEmbeds — opt-in allOf-by-default for embeds
stream: —
origin: ii
status: done
release: v0.36
issues: []
prev: "§3.5"
---

# `DefaultAllOfForEmbeds` — opt-in allOf-by-default for embeds

**Status:** ✅ done on `feat/feature-v0.36` (commit `0d550cf`), awaiting
review/merge. Shipped as `Options.DefaultAllOfForEmbeds` (opt-in, default
false) per the **Shape** below. Implementation reclassifies a plain, unnamed
embed in `scanEmbeddedFields` (allof.go) as an allOf member — `isAllOf` forced
true when `embedNestName(afld, fd) == ""` — so it reuses the existing
`buildAllOf` path (which already does $ref-for-model / inline-otherwise, pointer
peeling, alias resolve, and stdlib specials). Json-named embeds stay nested
(go-swagger#2038); `swagger:allOf` already wins; interface embeds untouched
(out of scope — they allOf-compose regardless). Fixture
`fixtures/enhancements/default-allof-embeds/` + on/off integration test
(`coverage_default_allof_embeds_test.go`); default-off output unchanged. Docs:
schema/README.md §allof + root CLAUDE.md options list.

**Origin.** W3 alias workshop Q-D close-out (2026-06-04). A core-product
enhancement with no dedicated stream; it rides whichever stream next touches the
schema builder's Options surface.

**Scope.** Today the documented rule is "embeds always inline properties, never
`$ref` unless the embed is `swagger:allOf`-tagged"
(`internal/builders/schema/embedded.go`). The Q-D patch enforced this rule for
aliased embeds too (previously aliased embeds were silently promoted to allOf
even without `swagger:allOf` — the bug Q-D fixed).

Some users prefer the opposite default. Generating client code from a spec where
every embed produces allOf composition gives downstream a clean inheritance
hierarchy — each embedded type becomes a base type the generator can reuse. The
inline default loses that structure: every embedding struct emits a flat copy of
the embedded fields, and the generator has no way to recover the "this composes
Y" relationship.

**Shape.** A new option `Options.DefaultAllOfForEmbeds bool` (default `false` so
existing users see no change). When `true`, plain (non-`swagger:allOf`-tagged)
embeds are emitted as allOf members containing `$ref` to the embedded type's
definition — the shape the per-field `swagger:allOf` annotation produces today.
Plain non-embed fields are unaffected; this is purely about how struct embeds
render.

Interactions:

- `swagger:allOf` annotation continues to win (already in allOf shape; the option
  just makes that the default).
- Pointer embeds (`*Base`) take the same path — the pointer is peeled to its
  named target.
- Aliased embeds resolve to their unaliased type (the Q-D contract) and then
  enter the allOf path under this option.
- Embeds of stdlib-special types (`error`, `time.Time`) interact with
  `applyStdlibSpecials` first; the recognizer's canonical shape takes precedence
  over any composition shape.

**When to revisit.** First time a user requests a flatter spec output from a
deeply-nested embed graph, or as part of an Options surface review.
