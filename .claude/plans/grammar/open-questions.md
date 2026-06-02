# Open questions — running list

Items to resolve as we iterate. Each question references the
sub-grammar(s) it affects.

## Q10 — Value-format heterogeneity (v1 legacy)

**Affects:** `10-shared.md`, every family file.

The grammar inherits ~9 distinct value formats from v1 (NumberValue,
IntegerValue, BooleanValue, StringValue, CommaListValue,
StringEnumValue, RawValue, JsonValue, plus YAML in extension bodies
and raw blocks). Same semantic content is expressed differently
depending on the keyword — see the inventory table in `10-shared.md`.

This round describes v1 faithfully. A v2 cleanup may want to
consider:

- **Unify "value of any type" on JsonValue.** `swagger:default`,
  `default:` keyword body, `example:` keyword body, and `Extensions:`
  body values would all be JSON. Loses YAML's multi-line ergonomics
  but standardises shape.
- **Unify on YAML bodies for anything multi-line.** Drop ad-hoc
  raw-block content shapes (consumes/produces lists, security
  requirements) in favour of a single YAML body convention.
- **Keep heterogeneity, document it.** Round-1 stance.

Defer to v2 design discussion; the grammar inventory is the
prerequisite.

## Q11 — EBNF expressiveness for context-sensitive boundaries — **SETTLED**

**Affects:** `00-overview.md`, `10-shared.md`, all family files.

The grammar delegates multi-line body accumulation to the **lexer**.
Every opaque multi-line body (`ExtensionYAMLBody`, `RawBlockBody`,
`OpaqueYamlBody`, `OpaqueRouteBody`) is a single lexer-emitted
terminal — the lexer's state machine handles "until next sibling
structural item", indentation tracking, and decorative-fence elision.

The grammar then references those tokens as terminals and stays
context-free: every raw-block-style production is the same shape,
`Head , EOL , BodyToken`. There is no per-family terminator-set
production in the grammar — that complexity lives in the lexer.

Trade-off: the grammar is no longer a complete parser-generator
input on its own. A reader needs the EBNF *plus* the lexer
specification (which defines how each body terminal is recognised).
Neither half subsumes the other. This is acceptable: the EBNF stays
clean and reads as a contract.

## Q1 — `OpaqueRouteBody` long-term plan — **RESTATED (round 1.1)**

**Affects:** `21-operation-grammar.md`, `30-delegated.md`.

Earlier rounds modeled the entire `swagger:route` body as a single
opaque `OpaqueRouteBody` terminal. That conflated two independent
asymmetries between `swagger:route` and `swagger:operation`:

1. **Top-level body shape.** `swagger:route` accepts inline keyword
   raw-blocks at the top level (`consumes:` / `produces:` /
   `security:` / `responses:` / `parameters:` / `schemes:` /
   `extensions:`); `swagger:operation` accepts those *plus* an
   optional `--- … ---` `OpaqueYamlBody`. The YAML fence is illegal
   under `swagger:route`.
2. **Per-keyword body sub-language.** The `parameters:` / `responses:`
   / `extensions:` *bodies* under `swagger:route` use distinctive
   sub-languages (`+ name:` continuation list, `<status>: <ref>` map,
   YAML-shaped extensions) handled by `internal/parsers/routebody/`.
   Under `swagger:operation`, those keyword bodies are YAML-shaped
   (and normally live inside the `OpaqueYamlBody`).

The grammar now models these as **two distinct block productions**
(`RouteBlock` and `InlineOperationBlock`, both alternatives of
`OperationFamilyBlock`) with their respective body shapes, and the
per-keyword sub-language is delegated **per-raw-block** in
`30-delegated.md`. `OpaqueRouteBody` as a top-level terminal is
removed.

A later iteration may fold the `internal/parsers/routebody/`
sub-grammars into the EBNF — see Q19.

## Q19 — Folding the routebody sub-grammars into the EBNF — **OPEN**

**Affects:** `21-operation-grammar.md`, `30-delegated.md`.

The `parameters:` / `responses:` / `extensions:` bodies under
`swagger:route` currently delegate to hand-written sub-parsers in
`internal/parsers/routebody/`. Replacing those with EBNF sub-grammars
is plausible but non-trivial — the `+ name:` continuation list and
the `<status>: <ref>` map have their own indentation conventions.

Round-1 stance: keep delegated. Revisit when the route bridge is
otherwise stable.

## Q2 — `swagger:operation` non-YAML keywords — **SETTLED**

**Affects:** `21-operation-grammar.md`.

Confirmed: `swagger:operation` accepts plain keyword properties
outside the `--- … ---` YAML body (e.g. `deprecated: true` on a
separate line). The current EBNF already permits this — `OperationBody`
is `{ … }` and `OpaqueYamlBody` is just one optional element. This
gives the user some leeway without forcing the heavy hammer of
inline YAML.

## Q3 — Annotation arg shapes — **SETTLED**

**Affects:** `10-shared.md`, all family files.

Final grouping codified as the **argument shape categories** table
in `10-shared.md`. The categories are:

- `NoArgs` — `swagger:alias` (confirmed), `swagger:file`,
  `swagger:meta`, `swagger:ignore`.
- `NameSpec` (def) — single defining name token (required or
  optional). Examples: `swagger:strfmt`, `swagger:model`,
  `swagger:response`, `swagger:name`, `swagger:enum` (name-override
  form), `swagger:allOf`.
- `NameSpec` (ref) — single referencing name. Example:
  `swagger:type`.
- `NameRefList` — one-or-more referencing names. Example:
  `swagger:parameters`.
- `RouteSpec` — `<METHOD> <path> [<tags>] <OperationID>`. Realised
  by `OperationArgs` in `21-operation-grammar.md`. Used by
  `swagger:route` and `swagger:operation`.
- `ValueSpec` — single-line value of any type. Realised by
  `swagger:default`'s argument *and* by every value-bearing
  validation/decorator keyword body (`maximum: <NumberValue>`,
  `default: <RawValue>`, `example: <RawValue>`, etc.). Unifies
  annotation arguments and keyword bodies under one notion.
- `EnumSpec` — `swagger:enum`'s tri-form (name override / plain list
  / bracketed list).

## Q5 — Name binding semantics — **SETTLED**

**Affects:** `10-shared.md`, `20-schema-grammar.md`,
`21-operation-grammar.md`, `23-classifier-grammar.md`.

All name tokens in the grammar share one lexical class
(`IdentifierName`), but the EBNF tracks **binding role** (defining
vs. referencing) and **namespace** in the wrapping production names.
This makes the grammar productions the contract for downstream
checks:

- per-namespace uniqueness on defining occurrences
  (no two `swagger:model` introducing `funky_model`, no two
  `swagger:operation` introducing `listPets`, etc.);
- per-namespace resolution on referencing occurrences (every
  `OperationIDRef` from `swagger:parameters` must resolve to an
  `OperationID` defined by some `swagger:route` / `swagger:operation`).

Initial taxonomy (full table in `10-shared.md` §"Namespace registry"):

| Namespace      | Defining          | Referencing        | Defined by                                    | Referenced by         |
|----------------|-------------------|--------------------|-----------------------------------------------|-----------------------|
| Model          | `ModelName`       | —                  | `swagger:model`                               | (Go-type tracking)    |
| Response       | `ResponseName`    | —                  | `swagger:response`                            | (Go-type tracking)    |
| OperationID    | `OperationID`     | `OperationIDRef`   | `swagger:route`, `swagger:operation`          | `swagger:parameters`  |
| Strfmt         | `StrfmtName`      | —                  | `swagger:strfmt`                              | (Go-type tracking)    |
| Enum           | `EnumTypeName`    | —                  | `swagger:enum` (when a leading name is given) | (Go-type tracking)    |
| Type token     | (closed vocab)    | `TypeRef`          | (predefined JSON Schema / OpenAPI vocabulary) | `swagger:type`        |

All classifier annotations now have final productions; see Q5b for
the settled record.

## Q5b — Classifier annotation binding semantics — **SETTLED**

**Affects:** `23-classifier-grammar.md`, `10-shared.md`.

- `swagger:name <MemberNameOverride>` — defining occurrence in the
  `MemberName` namespace, scoped per enclosing container (Go struct
  or interface). Overrides the JSON name of the member (struct field
  or interface method) the comment is attached to. The grammar does
  not enforce placement; the analyzer checks per-container uniqueness
  post-override.
- `swagger:default <DefaultValue>` — value annotation, not a name
  annotation. Argument is `JsonValue | RawValue`. The analyzer
  validates that the supplied value fits the target Go type. Not
  registered in any namespace.
- `swagger:allOf [<PolymorphicClassName>]` — without an argument, the
  annotation tags a struct member (typically an embedded type or
  inlined anonymous struct) for plain JSONSchema `allOf` composition.
  With the optional class name, the subtype is also tagged with an
  `x-class` extension carrying the class name (the OAS 2 polymorphic
  hint, allOf + discriminator). `PolymorphicClass` is an
  **accumulating** namespace: multiple subtypes sharing a class name
  register as siblings; no uniqueness constraint within. Placement
  (only on inlined struct members, never on a type itself) is a Go-
  level convention not enforced by the grammar.
- `swagger:ignore` — takes **no parameter**. Classifies the
  associated Go item (field, type, method, …) as ignored during spec
  production. The grammar produces a `ClassifierBlock` like any
  other; how dependent models are refit is the analyzer's job (the
  builder packages).

## Q5c — `swagger:type` token vocabulary — **SETTLED**

**Affects:** `23-classifier-grammar.md`.

`TypeRef` (renamed from `OASTypeRef` so the grammar layer carries no
explicit reference to OpenAPI) keeps the full closed vocabulary
unconditionally: `string` | `integer` | `number` | `boolean` |
`array` | `object` | `file` | `null`. `null` is a distinct type in
every JSON Schema version. `file` is an OpenAPI 2.0 token but stays
in the grammar — diagnosing its use when the user wants an OAS 3.x
output spec is a downstream-validation concern, not a grammar
concern.

## Q6 — `swagger:enum` argument shapes — **SETTLED**

**Affects:** `23-classifier-grammar.md`, `10-shared.md`.

`swagger:enum` accepts an **optional schema name** and an **optional
value list**, with the constraint that at least one of the two must
be present:

- `EnumWithName` = `EnumTypeName , [ Whitespace , EnumValueList ]`
  - Name alone → values come from code discovery.
  - Name + values → inline values fully override code discovery.
- `EnumValuesOnly` = `EnumValueList`
  - Schema name inferred from the enclosing Go type.

The value list itself is one of:

- `EnumPlainList` — `EnumPlainItem , { "," , EnumPlainItem }` (≥1
  items; single-item form is JSON Schema `const` equivalent).
- `EnumBracketedList` — `"[" … "]"` (single-line; ≥0 items). Items
  may be strict `JsonValue`s or bare trim-stripped tokens
  (`EnumListItem = JsonValue | EnumPlainItem`).

Disambiguation rule (on the trim-stripped argument):

1. Starts with `[` → `EnumValuesOnly = EnumBracketedList`.
2. Starts with an `IdentifierName`-shaped token + whitespace +
   non-empty rest → `EnumWithName` (name + value list per
   sub-dispatch).
3. Is exactly one `IdentifierName`-shaped token, no trailing content
   → `EnumWithName` with no value list (name only — v1 back-compat).
4. Otherwise → `EnumValuesOnly = EnumPlainList`.

Value typing: `EnumPlainItem`s are strings at the grammar level; the
analyzer coerces to numeric/boolean per the underlying Go type.
`JsonValue` items in bracketed lists carry their JSON type natively.

## Q7 — *(skipped — old numbering)*

## Q8 — `SchemaBlock` body vs. `UnboundBlock` body — **SETTLED**

**Affects:** `20-schema-grammar.md`.

The motivation is **expressiveness, not strictness**: the grammar
should *recognise* schema-bearing content in two distinct contexts so
the analyzer can apply context-specific rules — but the grammar
itself parses optimistically (recognise as much as possible, let a
later validator work out issues).

Definition of `UnboundBlock` (clarified): a comment block on a
member-level Go declaration — struct field, interface method, or
individual `const` / `var` declaration — i.e. anything **not** the
godoc headline of a top-level type/function. UnboundBlocks may
appear under any schema-bearing parent annotation (model, response,
parameters, *and* enum: const docstrings under a `swagger:enum`-tagged
type are UnboundBlocks too). The grammar production is the same
regardless of the parent annotation; the analyzer determines parent
context from the Go AST.

Resolution: split into `SchemaAnnotationBody` (under a
`swagger:model`/`parameters`/`response` header) and `UnboundBlockBody`
(member-level, no annotation). Both productions have the **same RHS**
(`SchemaBodyItem`) — they differ only in label, so the AST carries
context info downstream.

## Q9 — Where this lives long-term — **SETTLED**

**Affects:** all files.

Stay in `.claude/plans/grammar/` for now (private/maintainer view).
Once we have a workable grammar, the next step will be separate
testing and comparison against the current minimal grammar parser —
not yet at that stage.

## Q12 — Schema body keyword vocabulary — **SETTLED**

**Affects:** `20-schema-grammar.md`.

The round-1 schema body keyword set is exactly what v1 supports (see
`internal/parsers/grammar/keywords_table.go`). The following OAS /
JSON Schema fields **are not** body keywords in this grammar; they are
either covered by classifier annotations or simply unsupported:

| Field          | Round-1 mechanism                                  |
|----------------|-----------------------------------------------------|
| `type`         | `swagger:type` classifier annotation only           |
| `format`       | `swagger:strfmt` classifier annotation only         |
| `title`        | Godoc prose — first paragraph of declaration doc    |
| `description`  | Godoc prose — body paragraphs                       |
| `allOf`        | `swagger:allOf` classifier annotation only          |
| `nullable`     | `extensions:` block (writes `x-nullable`) / auto on pointer fields |

A v2 rev should consider promoting these to body keywords for
JSON-Schema-aligned consistency. Logged in `20-schema-grammar.md`
§"Out of scope — v2 enhancements".

## Q13 — `required:` and `discriminator:` polymorphism — **SETTLED**

**Affects:** `20-schema-grammar.md`.

Both keywords are **field-level boolean markers** in v1 — never the
schema-level forms (string list for `required`, property-name string
for `discriminator`) that JSON Schema / OAS define. The grammar models
the boolean form only:

```ebnf
RequiredLine      = "required"      , ":" , [ Whitespace ] , BooleanValue , LF ;
DiscriminatorLine = "discriminator" , ":" , [ Whitespace ] , BooleanValue , LF ;
```

The analyzer keys on the property name and writes into the enclosing
schema's `Required` slice / `Discriminator` field (see
`internal/builders/schema/bridge.go` `setRequired` / `setDiscriminator`).

Schema-level string-list `required:` and string `discriminator:` are
v2 enhancements — see `20-schema-grammar.md` §"Out of scope — v2
enhancements".

## Q14 — `enum:` keyword body alignment with `swagger:enum` — **SETTLED**

**Affects:** `20-schema-grammar.md`, `23-classifier-grammar.md`.

The `enum:` keyword body and the `swagger:enum` annotation argument
share the **same value-list shape** (`EnumPlainList | EnumBracketedList`).
The keyword body has no name component — the keyword targets the
enclosing field/property, which produces an anonymous schema in the
final spec.

```ebnf
EnumValidation = [ ItemsPrefix ] , "enum" ,
                 ":" , [ Whitespace ] , EnumValueList , LF ;
                (* EnumValueList imported from ClassifierGrammar *)
```

Disambiguation rule (re-uses the bracketed-vs-plain detection from
`swagger:enum`): starts with `[` → `EnumBracketedList`; otherwise
`EnumPlainList`.

Bridges the v1 quirk where the keyword accepted CommaListValue
(strings only, no JSON typing) and the annotation accepted typed
JSON values — both now share one vocabulary, with per-value typing
applied via `helpers.ParseValueFromSchema` against the target's
inferred type (see `internal/parsers/helpers/enum.go` `ParseEnum`).

## Q15 — `default:` / `example:` / `enum:` multi-line bodies — **SETTLED**

**Affects:** `00-overview.md`, `20-schema-grammar.md`,
`23-classifier-grammar.md`.

The three value-bearing keyword bodies are unified under a single
**lexer-emitted `RawValueBody` terminal**:

```ebnf
DefaultLine    = [ ItemsPrefix ] , "default" , ":" , [ Whitespace ] , RawValueBody ;
ExampleLine    = [ ItemsPrefix ] , "example" , ":" , [ Whitespace ] , RawValueBody ;
EnumValidation = [ ItemsPrefix ] , "enum"    , ":" , [ Whitespace ] , RawValueBody ;
```

The lexer accumulates body content from after the `:` until the next
sibling structural item (next keyword, next annotation, EOF). The
single-line case (`default: 3`) is the trivial path; multi-line
object/array JSON, multi-line YAML, multi-line bracketed enum lists
all fall under the same token. Downstream sub-parsers then interpret
the captured bytes per keyword: `default:` / `example:` go through
`internal/parsers/helpers/ParseValueFromSchema` against the target
Go type; `enum:` is parsed as `EnumValueList` (per Q14).

This deliberately backports multi-line support to the `swagger:enum`
annotation argument too — its argument should accept the same
`RawValueBody` shape as the `enum:` keyword body, so the two stay
aligned (per Q14). The exact annotation-argument syntax for the
multi-line case is left to the lexer-spec follow-up; the EBNF stays
clean by referencing the terminal.

`in:` is **not** a schema-grammar keyword — it is recognised in the
parameters dispatch path only. Removed from `SchemaDecorator`.

## Q16 — `ItemsPrefix` semantics — **SETTLED**

**Affects:** `10-shared.md`, `20-schema-grammar.md`.

`ItemsPrefix` is the dotted prefix that injects a validation into a
nested array level (`items.maxLength`, `items.items.maxLength`, …).
Each repeated `items.` segment descends one array level in the target
Go type. The grammar checks **only the surface syntax** (one or more
`items.` segments before a validation keyword); the analyzer (see
`internal/builders/schema/bridge.go` `collectItemsLevels`) walks the
declared type and validates that the depth fits.

```ebnf
ItemsPrefix      = ItemsSegment , { ItemsSegment } ;
ItemsSegment     = "items" , "." ;
```

The prefix is therefore unbounded in the grammar — a 5-deep
`items.items.items.items.items.maxLength: 10` parses cleanly; the
analyzer rejects it if the field's Go type doesn't have five array
levels.

## Q17 — Annotation crossover within a single comment block — **SETTLED**

**Affects:** `00-overview.md`, `20-schema-grammar.md`,
`23-classifier-grammar.md`.

A Go comment block may carry **multiple** `swagger:` annotations
stacked top-to-bottom. The dispatcher splits the block on every
recognised annotation header: each new `swagger:<name>` line **closes
the preceding annotation block** and starts a fresh one of the new
annotation's family. Lines between two annotations belong to the
earlier annotation; lines after the last annotation belong to it.

Practical consequence inside a struct: a member-level comment opening
with a classifier annotation produces a `ClassifierBlock`, and its
trailing prose (or further annotations) attach there — *not* to a
synthesised `UnboundBlock`.

## Q18 — `required:` / `readOnly:` are validations, not decorators — **SETTLED**

**Affects:** `20-schema-grammar.md`.

The grammar keeps the validation/decorator split for documented
intent. `required:` and `readOnly:` constrain *whether* / *how* a
value may be present — both are validations. `default:`, `example:`,
`discriminator:` are decorators. `in:` is removed from the schema
grammar entirely (parameters-only — handled outside `SchemaBodyItem`).

## Q20 — `swagger:route` and `swagger:operation` are identical writes, differ only in body capture — **OPEN (v2 realignment)**

**Affects:** `21-operation-grammar.md`,
`internal/builders/routes/`, `internal/builders/operations/`.

The two operation-bearing annotations target the **same write
surface**: both produce a `*oaispec.Operation` and store it via
`operations.SetPathOperation` into the same `Paths[<path>].Get` /
`Post` / … slot. From the produced Swagger spec's perspective they
are indistinguishable — there is no way to tell which annotation
authored a given operation by inspecting the output.

What differs is **how the comment body is captured**:

- `swagger:route` walks `block.Properties()` and dispatches inline
  keyword raw-blocks (`schemes`, `consumes`, `produces`, `security`,
  `parameters`, `responses`, `extensions`, `deprecated`) to
  per-keyword body parsers in `internal/parsers/helpers/` and
  `internal/parsers/routebody/`.
- `swagger:operation` walks `block.YAMLBlocks()` and feeds the first
  `--- … ---` body straight to `op.UnmarshalJSON` via
  `yaml.Unmarshal → fmts.YAMLToJSON`. Per-keyword dispatch does not
  exist on this path; the parser does recognise top-level inline
  raw-blocks under `swagger:operation` (per
  `internal/parsers/grammar/parser.go:275-278`'s allowed contexts)
  but the operations builder **silently drops** them.

In practice, anything one form can express the other can express,
with one OAS-feature exception: response-level constructs (e.g.
status-keyed response headers, examples) are only reachable through
`swagger:operation`'s YAML body because there is no inline-keyword
representation for them.

**v2 realignment:** make the two annotations true synonyms — same
write semantics, with both body conventions accepted under either
name. The dispatcher decides at parse time whether a block uses
inline keywords, a YAML fence, or both, and the unified builder
handles all three uniformly. The current split into two builder
packages plus two grammar block productions is then a pure
historical artifact and can collapse into one.

Round-1 stance: keep the two productions distinct (the grammar must
faithfully describe v1). The realignment is a v2 redesign; logged
here so it isn't forgotten.

## Round-1 known omissions

These are not "open questions" but recorded simplifications:

- **No quoted strings** in `StringValue` / `CommaListValue`.
  v1 doesn't have them. Workshop W2 may revisit.
- **No JSON/YAML inline values** in `RawValue`. `default:` /
  `example:` are single-line text; the analyzer parses against the
  target Go type. Workshop W3 may revisit.
- **`example` block-body absorption** for object/array bodies is
  not modelled — a known v1 quirk.
- **Diagnostics are not part of the grammar.** Severity, codes, and
  recovery strategy live in `internal/parsers/grammar/diagnostic.go`.
