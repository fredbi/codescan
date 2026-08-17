---
title: Example value coercion (verification)
stream: —
origin: i
status: closed
release: v0.37
issues: [go-swagger#1268, go-swagger#2246]
prev: "§2"
---

# Example value coercion (verification)

**Status:** ✅ **CLOSED 2026-08-02** — verification done, and it was not a no-op: both issues were
still open at **declaration** sites. Fixed on `fix/strfmt-dispatch-symmetry` (`94845da`, `47710cf`)
while closing quirk Q36.

**Origin.** W3 deferred 2026-04-21 — richer `example:` / `examples:` support. Groomed 2026-06-23 to a
verification task: confirm #1268 / #2246 are "allegedly already fixed", then retire.

## What the verification found

Both were fixed at **field** sites and still broken at **declaration** sites. `fixtures/bugs/1268`
was an **empty directory** — the verification had never actually been run, and the "allegedly"
in the previous revision was doing real work.

| issue | shape | field site | decl site (before) | now |
|---|---|---|---|---|
| #1268 | `example: ["a","b"]` on `[]string` | ✅ array | ❌ the string `"[\"a\",\"b\"]"` | ✅ array |
| #2246 | `example: {"x":1}` on a struct | ✅ object | ❌ an escaped string | ✅ object |

Root cause was shared and is written up as **Q36** in `quirks-open.md`: a declaration's comment block
is dispatched before its Go type is resolved onto the schema, so `default:` / `example:` / `enum:`
coerced against an empty type and fell back to the author's raw text. `handlers.RecoerceDeclValues`
re-types them at the same seam where `RecheckSchemaShape` already re-gates validations.

`TestIssue2540` had been asserting #2246's defect all along — a decl-level object example as an
escaped string, two lines above a field-level one asserted as real JSON, in the same expectation.

## Deliberately NOT changed

`example: a, b` (bare comma list) on a `[]string` stays the string `"a, b"`. `CoerceValue` absorbs
JSON parse failures on `object` / `array` targets by design — `validations/README.md`
§coercion-dispatch: *"the assumption is that an author who wrote `default: notjson` against an object
target intended a textual placeholder"*. Only the JSON-array form is a coercion request. The comma
list is an `enum:`-only input shape (§enum-shapes).

That README also says numeric and boolean parse failures *are* surfaced "so the consumer can decide
whether to emit a diagnostic" — which no consumer did until `47710cf`. Uncoercible scalars are now
dropped with a warning at both decl and field sites, and an uncoercible `enum:` member is dropped by
name. The object/array absorption is untouched; the README's "a future strict-mode option could turn
this into a diagnostic" remains the open door.

## Scope that was already settled

- **`examples:` (plural)** — in OAS2 specific to responses, covered by
  [response-examples-by-mime](response-examples-by-mime.md). In OAS3 (name-keyed map) deferred to
  OAI v3 (Stream 10). Nothing to decide.

## Witness

`fixtures/enhancements/default-example-typing` + `TestDefaultExampleTyping` — every declaration cell
paired with a field-site control carrying the identical literal, asserted to agree rather than
against a hard-coded coercion result. `default:` and `example:` are paired in every cell, since they
share `ParseDefault` and any divergence between them is itself a defect.

The empty `fixtures/bugs/1268` directory can be removed, or left as a pointer; the behaviour it was
meant to hold is covered by the fixture above.
