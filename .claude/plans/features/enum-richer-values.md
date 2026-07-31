---
title: Enum on var and richer Go expressions
stream: —
origin: i
status: open
release: v0.37
issues: []
prev: "§1"
---

# Enum on `var` and richer Go expressions

**Status:** ⬜ open · **approach undecided (TODO)** — shaping and priority not yet
settled.

**Origin.** W2 §4.1 (workshop 2026-04-21); raised by Fred as a real use case —
enum values that aren't const-representable (`[]byte`, typed structs, typed
maps). This is a core-product enhancement with no dedicated stream; it rides
whichever stream next touches the scanner / enum seam.

A coherent "richer enum values" feature spanning three sub-concerns:

- **`var` declarations and richer value expressions.** Lift the hard
  `token.CONST` gate at `internal/scanner/scan_context.go` and extend
  `findEnumValue` to accept value expressions beyond `*ast.BasicLit`:
  - `var (...)` declarations (the immediate need).
  - Composite literals (`Status{Label: "foo", Code: 1}`).
  - Call-result initializers — via `go/constant` where compile-time-known.
- **Inline YAML-block enum values.** A surface where the `enum:` keyword has a
  block body that carries non-scalar values:

  ```go
  // enum:
  //   - complex: {object: value}
  //   - another: [array, element]
  ```

  When `enum:` has a block body (captured as `Property.Body`), hand it to
  `internal/parsers/yaml/` and fold the result into the returned `[]any`.
- **Non-scalar emission parity audit.** v1's JSON path accepts objects, arrays,
  and null, but no fixture demonstrates emission. Before any release claims
  "OpenAPI-complete enum support", add object-valued and array-valued enum
  fixtures and confirm the full pipeline (parser → sub-parser → schema builder
  → spec) round-trips them.

**Already done — v1 enum quirks (§1.2a).** Leading-whitespace in comma lists and
the silent drop of `swagger:enum TypeName` with no matching consts were
**migration-commit obligations fixed in the P5.1 schema-builder migration**. The
`swagger:enum TypeName`-without-consts case now raises
`parse.invalid-enum-option` ("no matching const values found; enum semantics
dropped"), verified 2026-07-30.

**§1.2b — residual: an inline override discards the type's per-value docs,
silently.** The original §1.2a entry claimed the `x-go-enum-desc` quirk was
closed with "nothing remains to do". Half of it was: the *stale* desc no longer
survives to contradict a narrowed enum — it is stripped. What remains is that the
strip is **silent and lossy**. Probed 2026-07-30 against
`fixtures/enhancements/enum-overrides` (case E):

```go
// swagger:enum PriorityE     → consts low / medium / high, each with a doc comment
type PriorityE string

type NotificationE struct {
    // enum: urgent, normal   ← the field narrows to values the type never declared
    Priority PriorityE `json:"priority"`
}
```

emits `{"type":"string","enum":["urgent","normal"]}` — no `x-go-enum-desc`, and
**no diagnostic** (only case D, the no-consts one, warns). So the per-value docs
`PriorityE` contributed vanish without a trace.

Whether that is a bug depends on intent, which is why it sits with this feature
rather than in the quirk register: if the field's override is a genuine narrowing
to *different* values, dropping the type's docs is correct and only the silence is
wrong (→ emit a Hint). If overrides are meant to *subset* the type's values, the
matching docs should be carried through (→ filter `x-go-enum-desc` rather than
strip it). Decide when this feature is picked up; it is the smallest useful slice
of it.

**When to revisit.** A scanner feature, logically independent of the
parser-migration — pick up when non-const enum values are demanded by real
users. The non-scalar audit blocks any "spec-complete enum" marketing claim, but
not a release inheriting v1's scalar-only corpus.
