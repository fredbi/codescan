---
title: Response-level examples (by mime type)
stream: 9
origin: iii
status: done
release: v0.36
issues: [go-swagger#2871]
prev: "§10"
---

# Response-level examples (by mime type)

**Status:** ✅ done (2026-06-23, branch `feat/feature-v0.36`). Second feature of
the v0.36 streak.

**Verified state (as predicted).** The `swagger:operation` YAML path already
emitted response `examples` for free (raw YAML → `spec.Response` unmarshal;
`fixtures/bugs/1713`, `fixtures/bugs/2871`). The struct-based `swagger:response`
path did **not** — no `examples` keyword existed.

**As built.** Added a new grammar keyword **`examples`** (`asRawBlock`,
`CtxResponse`-only), joined to the lexer's YAML-body set so the mime→object
nesting survives. `responses/walker.go::applyBlockToDecl` now finds the
`examples:` property in the response decl block and parses its body via
`yaml.UnmarshalBody` into `Response.examples` (`map[string]any` keyed by mime
type). A malformed block raises a non-fatal `scan` warning and is skipped.

- The singular schema `example:` is untouched (separate decorator; see
  [example-values](example-values.md)). OAS2 reserves the plural `examples` for
  responses.
- Out of scope (unchanged): per-model contextual / dynamic example values.

**Tests.** `fixtures/enhancements/response-examples-by-mime/` +
`internal/integration/coverage_response_examples_test.go` (mime-keyed map with a
JSON object + an XML string, body `$ref` and description preserved) + golden
`enhancements_response_examples_by_mime.json`.

---

_Original grooming notes:_

**Origin.** go-swagger#2871 (backlog verification, 2026-06-13). The reporter's
"dynamic examples" framing (per-endpoint context-dependent values) is **rejected**
— a runtime concern, not a static-spec one. The legitimate nugget: OpenAPI 2.0
lets a Response object carry an `examples` map keyed by mime type
(`examples: {application/json: {…}}`), which codescan does not emit.

**First: verify where we stand.** This grooming assumes codescan does **not**
currently support response examples at all. **Confirm that.** If partial support
already exists, supporting both behaviours is more involved and needs its own
reconciliation.

**Decision (assuming no current support).** Reuse the **`examples:` keyword
(plural)** inside a `swagger:response` block, where the **first-level keys are
mime types** (the OAS2 shape), parsed via the yaml sub-parser into the OAS2
`Response.examples` field.

- **Note the OAS2 keyword split:** in OAS2 `examples` (plural) is the **response**
  keyword; `example` (singular) is for **schemas**. (See the schema-side
  [example-values](example-values.md).)
- Out of scope: per-model contextual / dynamic example values.

**When to revisit.** When response-example demand recurs.
