# SharedGrammar — primitives shared across families

Imported by every family sub-grammar. Defines the lexical foundation,
value categories, prose/blank handling, and the generic `RawBlock`
shape used by `OperationGrammar` and `MetaGrammar`.

This file does **not** define any annotation-specific productions —
those live in `20-schema-grammar.md`, `21-operation-grammar.md`,
`22-meta-grammar.md`, `23-classifier-grammar.md`.

## Lexical primitives

```ebnf
LF               = "\n" | "\r\n" ;
EOL              = [ "." ] , LF ;
                  (* Annotation and keyword lines: a trailing period
                     is non-significant and elided.

                     Why: many Go projects run the `godot` linter,
                     which adds a "." at the end of every comment
                     sentence — including annotation and keyword
                     lines. The grammar absorbs this so
                     `swagger:strfmt uuid.` and
                     `swagger:strfmt uuid` are equivalent.

                     Annotation headers and keyword body lines
                     terminate with EOL; prose lines (whose trailing
                     "." is part of the sentence) terminate with LF. *)

Whitespace       = ? one or more space / tab characters
                     (Unicode \p{Zs} accepted) ? ;
Letter           = ? Unicode letter (\p{L}) ? ;
Digit            = ? Unicode digit (\p{N}) ? ;
Digits           = Digit , { Digit } ;
```

The grammar is line-oriented; productions never span lines except
where explicitly noted (`RawBlock`, `ExtensionsBlock`, `OpaqueYamlBody`,
`OpaqueRouteBody`).

## Annotation prefix

```ebnf
PrefixTag        = "swagger:" ;
                  (* The single literal that introduces every
                     annotation. Defined once here so every family's
                     annotation header production references this
                     production rather than repeating the literal. *)
```

## Names — binding roles and namespaces

Every name token in the grammar shares one lexical class
(`IdentifierName` below). What matters for downstream tooling is **not**
how the token is spelled, but **where it is introduced and what
namespace it belongs to**:

- A **defining** occurrence introduces a name into a namespace. The
  namespace must be unique-within: e.g. two `swagger:model`
  declarations introducing `funky_model` are a name conflict.
- A **referencing** occurrence points to a name introduced elsewhere.
  The analyzer verifies that every reference resolves to a definition
  in the matching namespace.

The grammar tracks role + namespace in the *production names* that
wrap `IdentifierName`. Each family sub-grammar defines its own
namespace-specific productions (e.g. `ModelName`, `OperationID`,
`OperationIDRef`) so that:

- a `Map<Namespace, List<DefiningOccurrence>>` falls naturally out of
  AST traversal (one map entry per defining production);
- per-namespace uniqueness checks are scoped to the corresponding
  defining productions;
- reference checks become "every `*Ref` production resolves to a
  defining production in its namespace."

```ebnf
IdentifierName       = NameStartChar , { NameChar } ;
                      (* Shared lexical class for every defining and
                         referencing name token in the grammar. The
                         binding role is carried by the wrapping
                         production. *)

NameStartChar        = Letter ;
NameChar             = Letter | Digit | "_" | "-" | "." ;
                      (* Settled: one shared character class for all
                         annotation name tokens. *)
```

### Namespace registry

The full taxonomy across all family sub-grammars. Productions are
defined locally in each family file; this table is the index.

| Namespace          | Defining production(s)   | Referencing production(s) | Defined by                                    | Referenced by                    |
|--------------------|--------------------------|---------------------------|-----------------------------------------------|----------------------------------|
| Model              | `ModelName`              | —                         | `swagger:model`                               | (resolved by Go-type tracking)   |
| Response           | `ResponseName`           | —                         | `swagger:response`                            | (resolved by Go-type tracking)   |
| OperationID        | `OperationID`            | `OperationIDRef`          | `swagger:route`, `swagger:operation`          | `swagger:parameters`             |
| Strfmt             | `StrfmtName`             | —                         | `swagger:strfmt`                              | (resolved by Go-type tracking)   |
| Enum               | `EnumTypeName`           | —                         | `swagger:enum` (when a leading name is given) | (resolved by Go-type tracking)   |
| MemberName         | `MemberNameOverride`     | —                         | `swagger:name`                                | (per-container scope; analyzer)  |
| PolymorphicClass   | `PolymorphicClassName`   | —                         | `swagger:allOf` (with optional class name)    | (analyzer aggregates subtypes)   |
| Type token         | (closed vocabulary)      | `TypeRef`                 | (predefined JSON Schema / OpenAPI vocabulary) | `swagger:type`                   |

Notes:

- `PolymorphicClass` is an **accumulating** namespace: multiple
  subtypes declaring the same class name register as siblings; there
  is no uniqueness constraint within. The analyzer aggregates
  per-class subtype lists for `x-class` extension output.
- `swagger:default` is intentionally **not** in the registry — its
  argument is a *value*, not a name. The `DefaultValue` production
  is defined locally in `23-classifier-grammar.md`.
- `swagger:ignore`, `swagger:alias`, `swagger:file` take no
  arguments — they classify the enclosing Go decl without
  introducing or referencing any name.

### `GoIdentifier` (special case)

```ebnf
GoIdentifier         = Letter , { Letter | Digit | "_" | "-" } ;
                      (* Distinct from IdentifierName: NO "." allowed.
                         Used only by the swagger:route godoc-prefix
                         exception in OperationGrammar — the leading
                         token is a real Go function name, not an
                         introduced spec/operation name. *)
```

## Annotation argument shape categories

A handful of recurring "argument shapes" appear across annotation
headers and keyword bodies. They are not standalone productions —
each is realised by a more specific production at the use site (e.g.
`OperationArgs` in `21-operation-grammar.md` realises `RouteSpec`).
Naming the categories explicitly helps when reasoning across
sub-grammars and keeps the per-family productions consistent.

| Category         | Meaning                                                                                | Realised at                                                |
|------------------|----------------------------------------------------------------------------------------|------------------------------------------------------------|
| `NoArgs`         | No argument.                                                                           | `swagger:alias`, `swagger:file`, `swagger:meta`, `swagger:ignore` |
| `NameSpec` (def) | A single name token introduced into a namespace (required or optional per annotation). | `swagger:strfmt <StrfmtName>`, `swagger:model [ModelName]`, `swagger:response [ResponseName]`, `swagger:name <MemberNameOverride>`, `swagger:enum <EnumTypeName>` (when a leading name is given), `swagger:allOf [PolymorphicClassName]` |
| `NameSpec` (ref) | A single referencing name (closed vocabulary or namespace lookup).                     | `swagger:type <TypeRef>`                                   |
| `NameRefList`    | One or more referencing names.                                                         | `swagger:parameters <OperationIDRef>+`                     |
| `RouteSpec`      | `<METHOD> <path> [<tags>] <OperationID>` — introduces an `OperationID`.                | `swagger:route`, `swagger:operation`                       |
| `ValueSpec`      | A typed value. Most use sites are single-line: the type is one of the value categories above (NumberValue, IntegerValue, BooleanValue, StringValue, CommaListValue, StringEnumValue, RawValue, JsonValue). The `default:`, `example:`, and `enum:` schema keyword bodies are the exception — they take a multi-line-capable `RawValueBody` lexer terminal (see Q15). | `swagger:default <DefaultValue>`; every validation / schema-decorator keyword body — `maximum: <NumberValue>`, `unique: <BooleanValue>`, `default: <RawValueBody>`, `example: <RawValueBody>`, `enum: <RawValueBody>`, etc. |
| `EnumSpec`       | Optional schema name + optional value list (plain comma list or bracketed list); at least one of the two must be present. | `swagger:enum`                                             |

`ValueSpec` is the unifying category for **all** value-bearing
constructs in the grammar — annotation arguments that take a value
(`swagger:default`) and keyword body lines that take a value
(every validation keyword and most schema decorators). The exact
lexical shape per use site is determined by the keyword/annotation
in question and lives in its declaring sub-grammar.

### Keyword aliases (informational)

Keyword name comparison (in family sub-grammars) is **case-insensitive**
and matches both the canonical name and any registered alias. Many
keywords accept hyphen / underscore / space-separated alternative
spellings (e.g. `max length`, `max-length`, `maxLen` for `maxLength`);
the alias set is in `internal/parsers/grammar/keywords_table.go`.

## Value categories

```ebnf
NumberValue      = [ CmpOperator ] , [ Sign ] , Digits , [ "." , Digits ] ;
CmpOperator      = "<=" | ">=" | "<" | ">" | "=" ;     (* `maximum: <5` *)
Sign             = "+" | "-" ;

IntegerValue     = Digits ;

BooleanValue     = "true" | "false" ;                   (* case-insensitive *)

StringValue      = ? any non-LF text, verbatim (no escapes interpreted) ? ;

CommaListValue   = StringValue , { "," , StringValue } ;
                  (* whitespace around commas is trimmed; empty items dropped *)

StringEnumValue  = ? exactly one of the keyword's declared values,
                     case-insensitive; canonical form returned ? ;

RawValue         = StringValue ;
                  (* analyzer-side typed: numeric for int fields,
                     boolean for bool fields, JSON-ish for objects. *)
```

## Value-format inventory (legacy heterogeneity)

The grammar inherits a non-trivial number of distinct value formats
from v1. Same semantic content (a number, a list, a structured
object) is expressed differently depending on the keyword. This is
**known v1 legacy** — the grammar describes it faithfully without
trying to fix it in this round. A future v2 cleanup may unify
toward fewer formats.

| Format            | Where used                                                                            | Lexical shape                                                |
|-------------------|---------------------------------------------------------------------------------------|--------------------------------------------------------------|
| `NumberValue`     | `maximum`, `minimum`, `multipleOf`                                                    | `[Op] [Sign] Digits [.Digits]`                               |
| `IntegerValue`    | `maxLength`, `minLength`, `maxItems`, `minItems`                                      | `Digits`                                                     |
| `BooleanValue`    | `required`, `readOnly`, `discriminator`, `deprecated`, `unique`                       | `true \| false`                                              |
| `StringValue`     | `pattern`, `version`, `host`, `basePath`, `license`, `contact`, …                      | any non-LF text, verbatim                                    |
| `CommaListValue`  | `enum` keyword body, `schemes`                                                        | comma-split text, items trim-stripped                        |
| `StringEnumValue` | `in`, `collectionFormat`                                                              | one of a closed token set                                    |
| `RawValue`        | `swagger:default` argument fallback (single-line) — see `JsonValue` for the primary form | any non-LF text — analyzer types it per target Go field     |
| `JsonValue`       | `swagger:default` argument; `EnumBracketedList` items (single-line, in brackets)      | RFC 8259 JSON literal                                        |
| `ExtensionYAMLBody` | `Extensions:`, `InfoExtensions:` block bodies                                       | lexer-emitted token; multi-line YAML mapping; YAML type inference |
| `RawBlockBody`    | `Consumes:`, `Produces:`, `Security:`, `SecurityDefinitions:`, `Tos:`, `ExternalDocs:` | lexer-emitted token; multi-line opaque body, content shape per-keyword (see `30-delegated.md`) |
| `RawValueBody`    | `default:`, `example:`, `enum:` keyword bodies; `swagger:enum` annotation argument (multi-line back-port — see Q15) | lexer-emitted token; single-line scalar or multi-line block; analyzer interprets per keyword (see `30-delegated.md`) |

Notable inconsistencies (factual; not fixed in this round):

- **Same logical "value of any type" in three formats.**
  `swagger:default` takes `JsonValue` (or `RawValue`); the `default:`
  keyword body takes a `RawValueBody` lexer terminal that the
  analyzer types via `ParseValueFromSchema`; `Extensions:` body
  values are YAML-typed. All three express "an arbitrary value" but
  do so via three different syntactic shapes. (The `default:` /
  `example:` / `enum:` keyword bodies were unified onto the single
  `RawValueBody` terminal under Q15 — earlier drafts of this file
  listed `RawValue` for these slots; the lexer terminal supersedes
  it.)
- **Single-line vs multi-line lists.** `enum:` (keyword body) is
  now a `RawValueBody` token that may contain a single-line comma
  list, a single-line bracketed list, or a multi-line block (per
  Q15). `consumes:` body is one-item-per-line with optional `-`
  prefix. `Extensions:` body is YAML-list-shaped.
- **Quoting / escaping varies.** `JsonValue` requires JSON-strict
  quoting (`"foo"`). `EnumPlainItem` does not (bare tokens).
  `StringValue` is verbatim (no escapes interpreted). YAML bodies
  follow YAML quoting.

These are documented here so future iterations can decide whether to
preserve, deprecate, or unify any of them. The current grammar
captures v1 behaviour as-is.

## JSON literal sub-grammar

Several productions accept a JSON literal — an arbitrary JSON value,
including objects, arrays, scalars, and `null`. The shape mirrors
RFC 8259 (JSON). Whitespace handling and escape rules are standard
JSON.

```ebnf
JsonValue        = JsonScalar | JsonArray | JsonObject ;

JsonScalar       = JsonString
                 | JsonNumber
                 | "true" | "false" | "null" ;

JsonString       = ? a JSON string literal — double-quoted, with
                     standard JSON escapes (\", \\, \n, \uXXXX, …) ? ;

JsonNumber       = ? a JSON number literal — int or float, optional
                     exponent, no leading zero on integers ? ;

JsonArray        = "[" , [ JsonValue , { "," , JsonValue } ] , "]" ;

JsonObject       = "{" , [ JsonPair , { "," , JsonPair } ] , "}" ;
JsonPair         = JsonString , ":" , JsonValue ;
```

## Prose, blanks, prelude

```ebnf
ProseLine        = ? any non-blank line that does not start with
                     `swagger:` and does not match a keyword line for
                     the enclosing family ? , LF ;

BlankLine        = { Whitespace } , LF ;

Prelude          = { ProseLine | BlankLine } ;
                  (* In source order, becomes Title (1st paragraph)
                     and Description (remaining). The parser also
                     accepts a godoc-style placement where the
                     annotation line follows the prose. *)
```

## Items prefix (validation re-targeting on nested arrays)

```ebnf
ItemsPrefix      = ItemsToken , { ItemsToken } ;
ItemsToken       = "items." | ( "items" , Whitespace ) ;
                  (* `items.maximum: 10` validates each array item.
                     `items.items.maximum: 10` validates the inner-array's
                     items, etc. Depth = number of repetitions; "maxItems"
                     is *not* an items-prefixed "Items" because recognition
                     requires a separator after the literal "items". *)
```

`ItemsPrefix` is used only by `SchemaGrammar` (validations + a subset
of decorators). It does not appear in operation/meta keywords.

## Generic raw-block shape — **superseded**

The earlier parameterised `RawBlock(Keyword, Terminator)` form (where
each family defined its own keyword set + terminator set) is
superseded in this round by the lexer-emitted `RawBlockBody`
terminal defined above (see §"Generic raw-block body terminal").

Family productions now read `Head , EOL , RawBlockBody` directly;
no terminator-set production is needed because the lexer's state
machine handles boundary detection.

The `RawBlock(Keyword, Terminator)` form remains referenced by
`21-operation-grammar.md` until that file is refactored to the new
shape; for new family files, prefer the lexer-terminal form.

## Extensions block

Used by Schema, Operation, and Meta families to declare `x-*` vendor
extensions. The block is a head + a lexer-emitted body terminal:

```ebnf
ExtensionsBlock      = ExtensionsHead , EOL , ExtensionYAMLBody ;

ExtensionsHead       = "extensions" , ":" ;

ExtensionYAMLBody    = ? lexer-emitted token: a multi-line YAML
                         mapping payload accumulated by the lexer.
                         The lexer collects body lines until the next
                         sibling structural item (per the enclosing
                         family's terminator set, in lexer state),
                         elides any decorative `---` fences, and
                         emits the payload as a single token. Each
                         top-level key in the parsed mapping must
                         match `ExtensionName`. Values are arbitrary
                         YAML — booleans, numbers, strings, arrays,
                         objects, null. See 30-delegated.md and
                         00-overview.md §"Grammar / parser boundary". ? ;

ExtensionName        = ( "x-" | "X-" ) , Letter , { Letter | Digit | "-" | "_" } ;
                      (* Constraint on each top-level key of the
                         YAML mapping produced from ExtensionYAMLBody. *)
```

`InfoExtensionsBlock` is **not** defined here — it is meta-only and
lives in `22-meta-grammar.md` (it reuses `ExtensionYAMLBody` and
`ExtensionName`).

(Historical note: an earlier `---`-fenced wrapping form of
`ExtensionsBlock` existed in v1 but is **purely decorative** — the
v1 parser feeds the entire body, with or without fences, to
`yaml.Unmarshal`. A probe fixture in the baseline worktree
(`fixtures/goparsing/meta/fenced-ext/`) confirms that fenced and
unfenced bodies produce byte-identical output. YAML's multi-document
`---` separator semantics make the fences free no-ops at parse time.
In the v2 lexer, fence elision is explicit — the lexer drops them
when accumulating `ExtensionYAMLBody`. Free-form structured content
under `swagger:meta` is expressed via `OpaqueYamlBody` instead —
see `22-meta-grammar.md`.)

## Generic raw-block body terminal

Other multi-line raw-block keywords (`Consumes:`, `Produces:`,
`Security:`, `SecurityDefinitions:`, `Tos:`, `ExternalDocs:`, …)
share one lexer-emitted body terminal:

```ebnf
RawBlockBody         = ? lexer-emitted token: opaque multi-line body
                         payload. The lexer accumulates lines until
                         the next sibling structural item (per the
                         enclosing family's terminator set, in lexer
                         state) and emits the payload as a single
                         token. Per-keyword content shape — MIME list,
                         OAS security-requirement list, OAS
                         external-docs object, etc. — is downstream;
                         see 30-delegated.md. ? ;
```

Family sub-grammars use `RawBlockBody` directly in their raw-block
productions: `ConsumesBlock = "consumes" , ":" , EOL , RawBlockBody ;`
and similar. The lexer tags the emitted token with the opening
keyword so the analyzer can dispatch downstream.

## External docs block

```ebnf
ExternalDocsBlock    = ExternalDocsHead , EOL , RawBlockBody ;
ExternalDocsHead     = ( "externalDocs" | "external docs"
                       | "external-docs" ) , ":" ;
```

`externalDocs` body is captured as a lexer-emitted `RawBlockBody`
token; its OAS shape (`description:` / `url:` lines) is parsed
downstream — see `30-delegated.md`.
