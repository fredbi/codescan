---
title: Restrict emitted formats to the official OAS set
stream: —
origin: ii
status: open
release: null
issues: []
prev: null
---

# Restrict emitted formats to the official OAS set

**Status:** ⬜ open · 📝 planned · **approach mostly settled, priority not**.

**Origin.** (ii) — surfaced by Fred while reviewing the go-swagger#3412 enum work
(2026-08-01): "we had some knob to disable extended formats such as `int16` and
stick to only official formats". Verified there is **no such knob in codescan**:
`Options` has no format-related field, and `resolvers.SwaggerSchemaForType`
(`internal/builders/resolvers/resolvers.go`) maps Go types to formats with no gate.

## What we emit today

Swagger 2.0 §4.3 lists `int32`, `int64`, `float`, `double`, `byte`, `binary`,
`date`, `date-time`, `password` as the *defined* formats, and explicitly allows
any other string as an open value. go-openapi / go-swagger use that latitude for
the widths Go actually has:

| emitted | official? | Go source |
|---|---|---|
| `int32`, `int64`, `float`, `double` | ✅ | `int32`, `int64`, `float32`, `float64` |
| `int8`, `int16`, `uint8`, `uint16`, `uint32`, `uint64` | ❌ extended | the same-named Go types |
| the `strfmt` registry (`uuid`, `email`, `date`, …) | mixed | `swagger:strfmt` / strfmt types |

The extended widths are load-bearing for round-tripping (go-swagger reads them
back to pick the Go type), which is why they are the default and why the knob
would be opt-**in** to strictness, not opt-out of a mistake.

## Shape

`Options.StandardFormatsOnly bool` (default false). When set, an emitted format
outside the OAS 2.0 defined set is **widened, not dropped**: `int8` / `int16` →
`int32`, `uint32` / `uint64` → `int64`, and each widening raises a Hint naming the
Go type it came from. Widening rather than dropping keeps the schema *checkable*
— `{type: integer}` with no format validates strictly less than `{integer,
int32}` — and keeps the value domain honest (never narrower than the Go type).

Open sub-questions for whoever picks this up:

1. **Does it cover `strfmt` formats too?** `uuid` and `email` are not in the
   defined set either, but they are the whole point of the strfmt package and no
   sane consumer wants them widened to a bare `string`. Almost certainly: no,
   the knob covers **numeric width formats only**, and the name should say so
   (`StandardNumericFormats`?).
2. **Where does it apply?** One place — `resolvers.SwaggerSchemaForType` — covers
   definitions, SimpleSchema targets, items and enums alike, since they all route
   through it (verified while fixing #3412). A post-pass over the built spec
   would be the alternative, and is worse: it cannot tell an inferred format from
   an author-written `format:` annotation.
3. **Does an author-written `format:` override survive it?** It must: the knob is
   about what *we* infer from a Go type, never about overriding an explicit
   annotation.

## When to revisit

Not blocking anything. Pick it up when a consumer that validates against the OAS
2.0 schema strictly complains, or alongside any other `resolvers` work — the
change is confined to one function plus an Option and its documentation.
