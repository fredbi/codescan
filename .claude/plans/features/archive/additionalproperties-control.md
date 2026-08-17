---
title: Explicit additionalProperties control on a model
stream: 9
origin: iii
status: done
release: v0.35
issues: [go-swagger#2539, go-swagger#3005]
prev: "§17"
---

# Explicit `additionalProperties` control on a model

**Status:** ✅ done (2026-06-17, `feat/additional-properties` Phase 2). Plan:
[additional-properties.md](../../archive/additional-properties.md).

Added the `swagger:additionalProperties <spec>` type/model marker and the
`additionalProperties: <spec>` field keyword (`true | false | TypeSpec`), plus
the typed `swagger:patternProperties "<re>": <spec>, …` marker. A model can now
set `additionalProperties: false` to forbid extra keys, or point it at a
type/$ref, without relying on a `map[string]X` field shape. Resolves #3005 (named
props + free-form values) too.

**Origin.** go-swagger#2539 (backlog verification, 2026-06-15). A `map[string]X`
field already emitted `additionalProperties: X`, but there was no annotation to
set `additionalProperties` explicitly on a struct model; the reporter's
`// additional properties: false` comment was unrecognized and silently absorbed
into the model description.
