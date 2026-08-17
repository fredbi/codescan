# W2 — Enum annotation design

Date: 2026-04-21
Status: **decided** (workshop closed)
Companion: `.claude/plans/grammar-parser-architecture.md` §3.2.1, §3.3
Gate: P4 → P5.1 (schema builder migration consumes these decisions)

---

## 1. Survey — v1 state

Three surface forms exist today. The parser migration must continue to
accept all three (see §2.1).

### 1.1 Inline comma-list on a field

```go
// swagger:model Order
type Order struct {
    // enum: red, green, blue
    Color string
}
```

Regex: `rxEnumFmt = "%s[Ee]num\p{Zs}*:\p{Zs}*(.*)$"`
(`internal/parsers/regexprs.go:56`). Handler: `SetEnum`
(`internal/parsers/enum.go:18`), splits on comma and coerces to the
field's declared Go type via `parseValueFromSchema`.

### 1.2 Inline JSON array on a field

```go
// enum: ["red", "green", "blue"]
```

Same regex as 1.1 — the handler's `ParseEnum`
(`internal/parsers/enum.go:96`) detects the `[` prefix and takes a JSON
path. JSON values may be any JSON-ish type (object, array, null) —
see §2.4.

**Note on an out-of-band "deprecated" comment:** v1's source carries a
remark suggesting comma-list is deprecated in favor of JSON array.
That remark does not reflect project consensus and should not
influence our design — both forms stay first-class.

### 1.3 Linked const via `swagger:enum TypeName`

```go
// swagger:enum Status
type Status string

const (
    Active   Status = "active"
    Disabled Status = "disabled"
)
```

Regex: `rxEnum` (`internal/parsers/regexprs.go:77`). Handler:
`FindEnumValues` in `internal/scanner/scan_context.go:312`. **Strictly
const-only:** a hard `gd.Tok != token.CONST` gate; `var`, `type`, and
composite literals are silently skipped. Value extraction via
`findEnumValue` (`scan_context.go:337`) accepts only `*ast.BasicLit`
(INT, FLOAT, STRING tokens); any expression shape (e.g., a cast, a
function call) is dropped.

### 1.4 Fixture reality check

Every fixture examined uses either plain `string` or `int` enum
values. No fixture exercises:

- Non-scalar enum values (struct, map, slice).
- `var`-declared enum sources.
- Composite-literal or computed enum values.
- The JSON-array inline form with non-scalar content.
- Both `swagger:enum TypeName` **and** a field-level `enum:` on the
  same Go field. No conflict resolution exists in code; the forms are
  orthogonal.

---

## 2. Decisions

### 2.1 Backward-compat surface — keep all three forms

v2 parser migration is **not a breaking release** for the annotation
language. Targeted for a v0.34 / v0.35 minor bump. All three v1
surface forms (comma-list, JSON array, linked-const) continue to
work identically.

The "comma-list is deprecated" remark in v1 code is explicitly
**not** the project position and should be ignored when evaluating
design options.

### 2.2 Value-type contract — introduce an enum sub-parser

Aligned with the §3.3 sibling-sub-parser pattern. The main grammar
parser stays oblivious to enum-value shape: it captures `Property.Value`
as a raw string and (for the linked-const form) emits an `AnnEnumDecl`
Block carrying the referenced type name.

**New subpackage: `internal/parsers/enum/`.** Mirrors
`internal/parsers/yaml/` — thin, focused, imported only by bridge-
taggers (never by `internal/parsers/grammar/`). Exports something like:

```go
// Parse converts a raw enum-value string into a []any, choosing the
// surface dialect automatically based on the input's shape:
//   - JSON-array prefix "[" -> JSON-array path
//   - otherwise            -> comma-list path
//
// fieldType is the Go type of the target field; when non-nil the
// sub-parser coerces values to it (strings to string-kinded types,
// numeric literals to int/float-kinded types, etc.). When nil, the
// sub-parser returns untyped values and the caller decides.
func Parse(raw string, fieldType types.Type) ([]any, []Diagnostic)
```

Future additions (YAML-block values, validator-wrapped scalars —
§4) slot in as additional surface-dialect detections inside this
same sub-parser.

### 2.3 Source extent — const-only, `var` deferred

v1's const-only gate stays for the parser-migration release. `var`,
type aliases, composite literals, and computed values are **not**
supported by v2.0 / v0.34 and no silent scanner-layer rework is
attempted. This is deliberately a consistency choice, not an
oversight.

Expanding the source extent is tracked as a flagged forthcoming
feature (§4.1).

### 2.4 Non-scalar enum values — accept at the sub-parser, emit in spec

The JSON-array path in v1's `ParseEnum` already accepts objects,
arrays, and JSON `null`. No fixture currently demonstrates emission,
but the plumbing is there. v2 preserves this: `internal/parsers/enum/`
returns `[]any`, and the schema builder emits whatever the analyzer
supplies (scalar, object, array, null) into the spec's `enum:`
property per JSONSchema and OpenAPI rules.

Analyzers that want to reject non-scalar enums can do so via
field-type constraints at the bridge-tagger layer, not at the parser.

### 2.6 Override semantics — inline wins over const inference

When both forms target the same field (`swagger:enum TypeName` on
the type + inline `enum: ...` on the field), the **inline
field-level values win**; const inference is suppressed for that
field. No merging, no deduplication — explicit override.

Verified against v1 via integration test
`TestCoverage_EnumOverrides` (commit `4cf0c41`) with five cases:

| Case | Source | v1 result |
|------|--------|-----------|
| A | `swagger:enum` + matching consts | const values |
| B | inline comma-list only | inline values |
| C | inline JSON array only | inline values |
| D | `swagger:enum`, no matching consts | schema emitted without type/enum (silent skip) |
| E | `swagger:enum` + consts + inline on field | **inline wins** |

Golden: `fixtures/integration/golden/enhancements_enum_overrides.json`.

**P5.1 obligations** surfaced by this golden:

1. Inherit the override rule exactly — no merge, no dedup.
2. Strip per-value whitespace in comma-list parsing (v1 quirk:
   case B currently renders `["low", " medium", " high"]` with
   literal leading spaces).
3. Drop the stale `x-go-enum-desc` when the inline form wins (v1
   quirk: case E retains the const-derived description even
   though the inline values supersede it).
4. Emit a diagnostic for case D (`swagger:enum TypeName` resolving
   to zero values) — today silently dropped.

Items 2–4 are explicit v1 divergences we will take during P5.1 and
document in the migration commit's message. Item 1 is parity.

### 2.5 Unified AST shape — `Property` passthrough

At the grammar AST layer:

- **Form 1 & 2** (inline comma-list / JSON array): one `Property` with
  `Keyword.Name == "enum"`, `Value == rawString`, `Typed.Type ==
  ValueNone`. The bridge-tagger invokes `enum.Parse(prop.Value, fieldType)`.
- **Form 3** (`swagger:enum TypeName`): Block with
  `AnnotationKind == AnnEnumDecl` and the referenced `Name` available
  via the block's positional args. The scanner-linked const values
  reach the builder as a separate signal (scanner concern, not
  parser concern — §3.2.1).

No new AST types are introduced by W2.

---

## 3. Implementation outline

To be executed at P5.1 (schema builder migration):

1. **Create `internal/parsers/enum/`** — thin sub-parser package:
   - `enum.go` with `Parse(raw, fieldType) ([]any, []Diagnostic)`.
   - Shape-detection: `[` prefix → JSON array path; else comma-list.
   - Field-type coercion where a `types.Type` is supplied.
   - Returns raw `[]any` (including non-scalars) — emission policy is
     the caller's.
2. **Schema builder bridge-tagger** — on encountering an `enum`
   Property, invoke the sub-parser with the field's Go type.
3. **Scanner-side linked-const path** — unchanged from v1: when the
   bridge-tagger sees an `AnnEnumDecl` Block, it looks up the type
   name in the scanner's existing const-index. No parser involvement.

No grammar-parser changes required. The enum keyword already has a
`ValueCommaList` entry in `keywords.go`; the sub-parser deals with
the actual structure.

---

## 4. Forthcoming features — flagged, not part of v2 parser migration

Mirrored in `.claude/plans/forthcoming-features.md` §1 so they're
tracked alongside other parked items. These are **documented
out-of-scope items** for the current parser migration. They are
real features the project wants eventually; this workshop does not
deliver them but preserves the design space for them.

### 4.1 `swagger:enum` on `var` and richer Go expressions

Current gate: `scan_context.go:312` hard-codes `token.CONST`. Future
work should lift the gate and extend `findEnumValue` to accept:

- `var (...)` declarations — useful for enum values whose type is
  not const-representable (`[]byte`, typed structs, typed maps).
- Composite literals — `Status{Label: "foo", Code: 1}`.
- Call-result initializers — `NewStatus("foo")` (compile-time-known
  via `go/constant` where possible).

No fixture exists today. When implementing, add fixtures first; the
parity harness (`internal/parsers/grammar/grammar_test/`) locks in
the expected output shape.

### 4.2 Inline YAML-block enum values

There was a past (unlanded) proposal to support:

```go
// enum:
//   - complex: {object: value}
//   - another: [array, element]
```

— a YAML-list of enum values, in a block-head body under the `enum:`
keyword. The proposal was debated but never merged; no code and no
fixture exist.

Natural home when added: the `internal/parsers/enum/` sub-parser
grows a third shape-detection branch — "block body is present" →
hand the body to `internal/parsers/yaml/` and fold the resulting
`[]any` into the return value.

### 4.3 Non-scalar emission parity audit

v1's `ParseEnum` JSON path accepts non-scalars but no fixture emits
them. Before any release claims "OpenAPI-complete enum support",
add fixtures with object and array enums to confirm the full
pipeline (parser → sub-parser → builder → spec) round-trips
correctly.

---

## 5. Cross-references

- **Architecture §3.2.1** (enum as cross-cutting case) — the original
  call-out this workshop resolves.
- **Architecture §3.3** (sub-parser pattern) — the same pattern
  `internal/parsers/yaml/` established; `internal/parsers/enum/`
  follows it.
- **W3** (example / examples annotation) — deferred, not held
  pre-P5.1. Same shape as W2 (array of any-type values), not
  supported in v1 beyond a single scalar raw-value. When W3 is
  later held, it reuses W2's sub-parser pattern directly. Tracked
  in `.claude/plans/forthcoming-features.md` §2.
- **P5.1** task template — the schema builder migration is the first
  consumer; it creates `internal/parsers/enum/` and wires the
  bridge-tagger path described in §3.

---

## 6. Change history

| Date | Change |
|------|--------|
| 2026-04-21 | Initial decision doc (W2 workshop closed). |
