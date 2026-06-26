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

**Already done — v1 enum quirks (§1.2a).** The concrete v1 enum bugs
(leading-whitespace in comma lists, silent drop of `swagger:enum TypeName` with
no matching consts, stale `x-go-enum-desc` when an inline override wins) were
**migration-commit obligations fixed in the P5.1 schema-builder migration**, not
pending items. Listed here only as a crossref; nothing remains to do on them.

**When to revisit.** A scanner feature, logically independent of the
parser-migration — pick up when non-const enum values are demanded by real
users. The non-scalar audit blocks any "spec-complete enum" marketing claim, but
not a release inheriting v1's scalar-only corpus.
