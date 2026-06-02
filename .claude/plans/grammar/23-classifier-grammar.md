# ClassifierGrammar — complete EBNF

**Status:** complete (round 1) for the classifier annotation family.
Standalone modulo the explicitly listed imports from `SharedGrammar`.

## Scope

`ClassifierGrammar` defines, in full, the legal language of the
**eight** classifier annotations (`swagger:name` left this family
during P7/S3 — see note below):

| Annotation header   | Block kind        |
|---------------------|-------------------|
| `swagger:strfmt`    | `ClassifierBlock` |
| `swagger:default`   | `ClassifierBlock` |
| `swagger:type`      | `ClassifierBlock` |
| `swagger:allOf`     | `ClassifierBlock` |
| `swagger:ignore`    | `ClassifierBlock` |
| `swagger:alias`     | `ClassifierBlock` |
| `swagger:file`      | `ClassifierBlock` |
| `swagger:enum`      | `ClassifierBlock` |

> **swagger:name moved to the schema family.** Although its primary
> role (member rename) reads classifier-shaped, its body accepts the
> full `SchemaAnnotationBody` production (validation keywords, raw
> blocks, etc.) — so it dispatches under `SchemaBlock` in the
> grammar (see `50-full.md` §3 / `20-schema-grammar.md`).
> The `NameAnnotation` production below is preserved here for
> historical context; the live definition is in §3 of `50-full.md`.

A classifier annotation **tags** the enclosing Go declaration with
metadata. It does not host a structured body — the comment block's
body is prose only.

`swagger:enum` fits in this family: every form keeps the annotation
as a single-line header followed by a prose-only body. The richer
argument grammar is contained inside `EnumArgs`; multi-line bracketed
lists are not allowed (the `EOL` after `EnumArgs` enforces single-line).

## Imports from `SharedGrammar`

This grammar uses the following productions defined in
`10-shared.md`. They are not redefined here.

| Imported            | Role                                                    |
|---------------------|---------------------------------------------------------|
| `LF`                | end-of-line — `"\n" \| "\r\n"`                           |
| `EOL`               | `[ "." ] , LF` — annotation/keyword line terminator with optional trailing dot elided |
| `Whitespace`        | one or more space / tab characters (Unicode `\p{Zs}`)   |
| `Letter`            | Unicode letter (`\p{L}`)                                |
| `Digit`             | Unicode digit (`\p{N}`)                                 |
| `NameStartChar`     | `Letter`                                                |
| `NameChar`          | `Letter \| Digit \| "_" \| "-" \| "."`                   |
| `IdentifierName`    | `NameStartChar , { NameChar }`                          |
| `PrefixTag`         | the literal `"swagger:"`                                |
| `Prelude`           | `{ ProseLine \| BlankLine }`                            |
| `ProseLine`         | a non-blank, non-`swagger:`, non-keyword-shaped line + `LF` |
| `BlankLine`         | `{ Whitespace } , LF`                                   |
| `RawValue`          | `StringValue` — any non-LF text, verbatim               |
| `StringValue`       | any non-LF text, verbatim (no escapes interpreted)      |
| `JsonValue`         | full JSON literal — `JsonScalar \| JsonArray \| JsonObject` |
| `JsonScalar`        | `JsonString \| JsonNumber \| "true" \| "false" \| "null"` |
| `JsonString`        | JSON string literal (double-quoted, standard escapes)   |
| `JsonNumber`        | JSON number literal (RFC 8259 number)                   |
| `JsonArray`         | `"[" , [ JsonValue , { "," , JsonValue } ] , "]"`        |
| `JsonObject`        | `"{" , [ JsonPair , { "," , JsonPair } ] , "}"`          |
| `JsonPair`          | `JsonString , ":" , JsonValue`                          |

## Top-level productions

```ebnf
ClassifierBlock      = Prelude , ClassifierAnnotation , ClassifierBody ;

ClassifierBody       = { ProseLine | BlankLine } ;
                      (* Prose only. No keywords, no raw blocks, no
                         YAML, no extensions. Title is the first
                         paragraph; description is what follows. *)
```

A classifier comment block is exactly:

1. zero or more `Prelude` lines (prose / blank — becomes Title +
   Description),
2. one `ClassifierAnnotation` line,
3. zero or more `ClassifierBody` lines (more prose / blank — appended
   to the description).

## `ClassifierAnnotation` dispatch

```ebnf
ClassifierAnnotation = StrfmtAnnotation
                     | NameOverrideAnnotation
                     | DefaultAnnotation
                     | TypeAnnotation
                     | AllOfAnnotation
                     | IgnoreAnnotation
                     | AliasAnnotation
                     | FileAnnotation
                     | EnumAnnotation ;
```

Each branch is a single-line annotation header.

## Annotation headers

```ebnf
StrfmtAnnotation       = PrefixTag , "strfmt"  , Whitespace , StrfmtName             , EOL ;
NameOverrideAnnotation = PrefixTag , "name"    , Whitespace , MemberNameOverride     , EOL ;
DefaultAnnotation      = PrefixTag , "default" , Whitespace , DefaultValue           , EOL ;
TypeAnnotation         = PrefixTag , "type"    , Whitespace , TypeRef                , EOL ;
AllOfAnnotation        = PrefixTag , "allOf"   , [ Whitespace , PolymorphicClassName ] , EOL ;
IgnoreAnnotation       = PrefixTag , "ignore"  , EOL ;
AliasAnnotation        = PrefixTag , "alias"   , EOL ;
FileAnnotation         = PrefixTag , "file"    , EOL ;
EnumAnnotation         = PrefixTag , "enum"    , [ Whitespace , EnumArgs ]           , EOL ;
```

`swagger:file` is a syntactic synonym of `swagger:type file`. Both
produce equivalent semantic content (a Go decl tagged as the `file`
type); they parse to distinct `Annotation` nodes but the analyzer
treats them as one.

## Argument productions

### Name productions (single-token arguments)

```ebnf
StrfmtName             = IdentifierName ;
                        (* def: Strfmt namespace. *)

MemberNameOverride     = IdentifierName ;
                        (* def: MemberName namespace, scoped per
                           enclosing container (Go struct or
                           interface). Overrides the JSON name of
                           the member (struct field or interface
                           method) the comment is attached to. *)

PolymorphicClassName   = IdentifierName ;
                        (* def: PolymorphicClass namespace
                           (accumulating — multiple subtypes sharing
                           a class name register as siblings; no
                           uniqueness constraint within). When
                           present, the subtype is also tagged with
                           an `x-class` extension carrying the class
                           name. *)

EnumTypeName           = IdentifierName ;
                        (* def: Enum namespace. The schema (type)
                           name of the enum, NOT a value name.
                           When omitted, inferred from the enclosing
                           Go type name. *)
```

### Closed-vocabulary references

```ebnf
TypeRef                = "string"  | "integer" | "number"  | "boolean"
                       | "array"   | "object"  | "file"    | "null" ;
                        (* ref: closed vocabulary of JSON Schema /
                           OpenAPI primitive type tokens. `null` is
                           a distinct type in every JSON Schema
                           version. `file` is an OpenAPI 2.0 token;
                           when targeting OAS 3.x, the analyzer
                           diagnoses its use. *)
```

### Value argument

```ebnf
DefaultValue           = JsonValue | RawValue ;
                        (* The argument is a *value*, not a name.
                           `swagger:default` does not introduce a
                           name into any namespace.

                           Disambiguation: try `JsonValue` first.
                           If the trim-stripped argument starts with
                           `"`, `[`, `{`, a digit, `+`, `-`, or
                           equals one of `true`/`false`/`null`,
                           parse as `JsonValue`. Otherwise fall back
                           to `RawValue`. *)
```

### Enum argument grammar

`swagger:enum` accepts an optional schema name and an optional value
list. **At least one of the two must be present.** The value list
may appear inline on the annotation line, or as a multi-line body
following the annotation header — see "Multi-line value-list body"
below.

```ebnf
EnumArgs               = EnumWithName | EnumValuesOnly ;

EnumWithName           = EnumTypeName , [ Whitespace , EnumValueList ]
                       | EnumTypeName , LF , RawValueBody ;
                        (* Name alone → values come from code
                           discovery (matching consts in the package).
                           Name + inline values → inline values fully
                           override code discovery.
                           Name + multi-line `RawValueBody` (see Q15)
                           → body re-parsed as `EnumValueList`. *)

EnumValuesOnly         = EnumValueList
                       | RawValueBody ;
                        (* Schema name inferred from the enclosing
                           Go type name. Inline `EnumValueList` (single-
                           line, comma or bracketed) and multi-line
                           `RawValueBody` (re-parsed as `EnumValueList`)
                           are both legal. *)

EnumValueList          = EnumPlainList | EnumBracketedList ;

EnumPlainList          = EnumPlainItem , { "," , EnumPlainItem } ;
                        (* ≥1 items. Single item is allowed —
                           equivalent to a JSON Schema `const`. *)

EnumPlainItem          = ? a StringValue containing no top-level ","
                            and no leading "[" or trailing "]", with
                            leading and trailing whitespace stripped;
                            empty items dropped ? ;

EnumBracketedList      = "[" , [ EnumListItem , { "," , EnumListItem } ] , "]" ;
                        (* Single-line: the closing "]" must appear
                           on the same line as the opening "[". *)

EnumListItem           = JsonValue | EnumPlainItem ;
                        (* Strict JSON value (string, number, boolean,
                           null, nested array, object) OR a bare
                           trim-stripped token (treated as a string
                           by the analyzer). The two are tried in
                           order: JsonValue first, EnumPlainItem on
                           fallback. *)
```

#### Disambiguation rule for `EnumArgs`

Applied to the trim-stripped argument string after `swagger:enum `:

| Step | Condition                                                                  | Branch                                                |
|------|----------------------------------------------------------------------------|-------------------------------------------------------|
| 1    | Starts with `[`                                                            | `EnumValuesOnly = EnumBracketedList`                  |
| 2    | Starts with an `IdentifierName`-shaped token followed by whitespace + non-empty rest | `EnumWithName` — name is the leading token; rest is `EnumValueList` (sub-dispatch: starts with `[` → `EnumBracketedList`, else → `EnumPlainList`) |
| 3    | Is exactly one `IdentifierName`-shaped token (no trailing content)         | `EnumWithName` with no `EnumValueList` (name only)    |
| 4    | Otherwise (e.g. digit-led, comma-led, non-identifier token first)          | `EnumValuesOnly = EnumPlainList`                      |

A bare single identifier is always a name (rule 3), preserving v1
back-compat. Single-value lists with non-identifier-shaped values
fall under rule 4 (`swagger:enum 1` → plain list with item `"1"`); a
single string value that happens to be identifier-shaped requires
the bracketed form (`swagger:enum [foo]`).

#### Multi-line value-list body

A `swagger:enum` annotation whose annotation line carries only the
schema name (or no argument at all) opens a multi-line
`RawValueBody` for the value list. The lexer accumulates content
until the next sibling structural item; the captured bytes are
re-parsed as an `EnumValueList`.

```
// swagger:enum Priority
//   - low
//   - medium
//   - high
type Priority string

// swagger:enum
//   ["low", "medium", "high"]
type Priority string
```

This deliberately backports the same body shape used by the `enum:`
keyword (per Q15) so the annotation argument and the keyword body
share one vocabulary. The single-line forms above remain legal —
multi-line is an opt-in that triggers when no inline value list
appears on the annotation line.

## Value typing — grammar vs. analyzer

The grammar extracts items but does **not** type them:

- `EnumPlainItem` items are always strings at the grammar level
  (the verbatim text, trim-stripped). When the enclosing Go type is
  numeric or boolean, the **analyzer** coerces each item to the
  target type. Example:

  ```
  // swagger:enum 1,2,3
  type Foo int
  ```

  The grammar produces items `["1","2","3"]`; the analyzer coerces
  to integers `[1,2,3]` because the underlying Go type is `int`.

- `JsonValue` items inside `EnumBracketedList` carry their JSON
  type natively (number, boolean, null, array, object) and are not
  re-coerced.

## Annotation → name-binding table

| Annotation         | Argument production                | Binding role + namespace                          |
|--------------------|------------------------------------|---------------------------------------------------|
| `swagger:strfmt`   | `StrfmtName` (required)            | def: Strfmt                                       |
| `swagger:name`     | `MemberNameOverride` (required)    | def: MemberName (per-container scope)             |
| `swagger:default`  | `DefaultValue` (required)          | value, not a name; analyzer-typed                 |
| `swagger:type`     | `TypeRef` (required)               | ref: closed type vocabulary                       |
| `swagger:allOf`    | `PolymorphicClassName` (optional)  | def (accumulating): PolymorphicClass — produces an `x-class` extension on the subtype |
| `swagger:ignore`   | (no args)                          | —                                                 |
| `swagger:alias`    | (no args)                          | —                                                 |
| `swagger:file`     | (no args)                          | — (synonym of `swagger:type file`)                |
| `swagger:enum`     | `EnumArgs` (required: name and/or value list) | def: Enum (when name token given); inline value list fully overrides code discovery |

## Examples

Each example shows the comment block (preprocessed: `// ` prefixes
stripped) that the grammar parses.

### `swagger:strfmt`

```
A MAC address.

swagger:strfmt mac
```

With trailing dot (godot-style) — equivalent to the form above:

```
swagger:strfmt uuid.
```

### `swagger:name`

```
Owner is the user that owns the resource.

swagger:name ownerID
```

### `swagger:default`

```
Identifier of the priority preset.

swagger:default high
```

```
Default request payload.

swagger:default {"limit": 10, "offset": 0}
```

### `swagger:type`

```
Custom timestamp encoded as a JSON string.

swagger:type string
```

### `swagger:allOf`

Plain `allOf` (no class):

```
swagger:allOf
```

Polymorphic discriminator hint (sibling subtypes share `Animal`):

```
swagger:allOf Animal
```

### `swagger:ignore`

```
Internal-only field; never emitted to the spec.

swagger:ignore
```

### `swagger:alias`

```
swagger:alias
```

### `swagger:file`

```
Multipart upload.

swagger:file
```

Equivalent to:

```
swagger:type file
```

### `swagger:enum`

Name only — values from code discovery:

```
swagger:enum Priority
```

Plain comma list, no name (schema name inferred from Go type):

```
swagger:enum low, medium, high
```

Plain list with numeric values:

```
swagger:enum 1, 2, 3
```

Single-item plain list (JSON Schema `const` equivalent):

```
swagger:enum 1
```

Bracketed list, no name:

```
swagger:enum [low, medium, high]
```

Single-item bracketed list — required when the value is identifier-shaped:

```
swagger:enum [foo]
```

Name + plain list:

```
swagger:enum my_enum a, b, c
```

Name + bracketed list:

```
swagger:enum my_enum [a, b, c]
```

Hybrid items (mix of bare strings and JSON values):

```
swagger:enum my_enum [a, {"x":1, "y":[1,2,3]}, c, [1,2,3], ["u","v"]]
```

The grammar identifies five enum values:

```json
{
  "my_enum": {
    "enum": [
      "a",
      {"x":1, "y":[1,2,3]},
      "c",
      [1,2,3],
      ["u","v"]
    ]
  }
}
```

(`a` and `c` are bare `EnumPlainItem`s — the analyzer treats them
as strings; the others are explicit `JsonValue`s.)

## Productions glossary

For self-check completeness, every production used in this grammar:

**Defined locally (this file):**

`ClassifierBlock`, `ClassifierBody`, `ClassifierAnnotation`,
`StrfmtAnnotation`, `NameOverrideAnnotation`, `DefaultAnnotation`,
`TypeAnnotation`, `AllOfAnnotation`, `IgnoreAnnotation`,
`AliasAnnotation`, `FileAnnotation`, `EnumAnnotation`,
`StrfmtName`, `MemberNameOverride`, `PolymorphicClassName`,
`EnumTypeName`, `TypeRef`, `DefaultValue`,
`EnumArgs`, `EnumWithName`, `EnumValuesOnly`,
`EnumValueList`, `EnumPlainList`, `EnumPlainItem`,
`EnumBracketedList`, `EnumListItem`.

**Imported from `SharedGrammar` (10-shared.md):**

`LF`, `EOL`, `Whitespace`, `Letter`, `Digit`, `NameStartChar`,
`NameChar`, `IdentifierName`, `PrefixTag`, `Prelude`, `ProseLine`,
`BlankLine`, `RawValue`, `StringValue`, `JsonValue`, `JsonScalar`,
`JsonString`, `JsonNumber`, `JsonArray`, `JsonObject`, `JsonPair`.

No undefined references. The grammar is closed.

## Notes

- All defining/referencing name tokens share `IdentifierName`. The
  binding role and namespace are carried by the wrapping production
  names, not by character-class divergence.
- Annotation header lines (and keyword body lines, defined elsewhere)
  use `EOL` so that a trailing `.` (godot-linter style) is elided.
  Prose lines use `LF` (the `.` is part of the sentence).
- The `EnumBracketedList` is constrained to a single line by design —
  `EOL` after `EnumArgs` enforces this. Multi-line bracketed lists
  would break the "annotation header is one line" property that
  defines this family.
- `swagger:default` and `swagger:enum` (in the bracketed form) are
  the only classifier annotations that reach into the JSON
  sub-grammar from `SharedGrammar`.
- Diagnostics, error recovery, and per-namespace
  uniqueness/resolution checks are *not* described by this grammar —
  they are the parser's and analyzer's responsibility. The grammar
  defines only the legal language.
