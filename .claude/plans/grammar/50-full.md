# Full grammar — synthesis read

A single-document reading of the codescan annotation grammar after
the layered family files (`10-shared.md`–`23-classifier-grammar.md`)
have been settled and the lexer (`40-lexer.md`) has absorbed all
quirks, multi-line bodies, and **lexical disambiguation**. The EBNF
here consumes lexer-emitted terminal tokens, not raw text.

This file is the contract a reader can hold in their head. Per-keyword
lexical detail and quirk handling are documented in their respective
files; this synthesis stays at the level of "what token sequence is
a legal program".

The grammar is rigorous ISO-14977 EBNF. Required vs optional
arguments and value typing are **grammar-visible** — every legality
constraint that can be expressed by token sequencing is expressed
that way, not bundled into a richer terminal.

## 1. Terminal vocabulary

The lexer emits a fixed alphabet of terminal tokens. Each terminal
is one specific (token-kind, name) pair. Per-terminal the lexer
performs lexical typing only — recognising what a value *looks*
like (number / string / boolean / JSON literal / identifier).
Semantic coercion against the Go target type happens in the
analyzer, never in the lexer.

### 1.1 Annotation name terminals

`TokenAnnotation`. Each terminal recognises the annotation **name**
only — positional arguments are emitted as separate terminals (§1.2).

| Terminal           | Annotation             |
|--------------------|------------------------|
| `ANN_MODEL`        | `swagger:model`        |
| `ANN_RESPONSE`     | `swagger:response`     |
| `ANN_PARAMETERS`   | `swagger:parameters`   |
| `ANN_ROUTE`        | `swagger:route`        |
| `ANN_OPERATION`    | `swagger:operation`    |
| `ANN_META`         | `swagger:meta`         |
| `ANN_STRFMT`       | `swagger:strfmt`       |
| `ANN_ALIAS`        | `swagger:alias`        |
| `ANN_NAME`         | `swagger:name`         |
| `ANN_ALLOF`        | `swagger:allOf`        |
| `ANN_ENUM`         | `swagger:enum`         |
| `ANN_IGNORE`       | `swagger:ignore`       |
| `ANN_DEFAULT`      | `swagger:default`      |
| `ANN_TYPE`         | `swagger:type`         |
| `ANN_FILE`         | `swagger:file`         |

### 1.2 Argument terminals

Annotation arguments — positional values that follow the annotation
name on the same line.

| Terminal       | Lexer kind         | Recognises                                                                                       |
|----------------|--------------------|--------------------------------------------------------------------------------------------------|
| `IDENT_NAME`   | `TokenIdentName`   | Identifier-shaped token (per `IdentifierName` in `10-shared.md`); used for every named argument and reference |
| `JSON_VALUE`   | `TokenJsonValue`   | RFC-8259 JSON literal (string/number/boolean/null/array/object). Tried first for value-shaped args |
| `RAW_VALUE`    | `TokenRawValue`    | Verbatim non-LF text — fallback when `JSON_VALUE` recognition fails                              |
| `TYPE_REF`     | `TokenTypeRef`     | Closed vocabulary: `string` / `integer` / `number` / `boolean` / `array` / `object` / `file` / `null` |
| `HTTP_METHOD`  | `TokenHttpMethod`  | `GET` / `POST` / `PUT` / `PATCH` / `HEAD` / `DELETE` / `OPTIONS` / `TRACE` (case-insensitive)    |
| `URL_PATH`     | `TokenUrlPath`     | An RFC-3986 URL path token (used as `OperationArgs`'s second positional arg)                     |

### 1.3 Inline-value keyword terminals

`TokenKeyword`. Each terminal recognises the keyword **name** only
— the value following the `:` is emitted as a separate terminal
from §1.4.

| Terminal               | Keyword              |
|------------------------|----------------------|
| `KW_MAXIMUM`           | `maximum`            |
| `KW_MINIMUM`           | `minimum`            |
| `KW_MULTIPLE_OF`       | `multipleOf`         |
| `KW_PATTERN`           | `pattern`            |
| `KW_MAX_LENGTH`        | `maxLength`          |
| `KW_MIN_LENGTH`        | `minLength`          |
| `KW_MAX_ITEMS`         | `maxItems`           |
| `KW_MIN_ITEMS`         | `minItems`           |
| `KW_UNIQUE`            | `unique`             |
| `KW_COLLECTION_FORMAT` | `collectionFormat`   |
| `KW_REQUIRED`          | `required`           |
| `KW_READ_ONLY`         | `readOnly`           |
| `KW_DISCRIMINATOR`     | `discriminator`      |
| `KW_DEPRECATED`        | `deprecated`         |
| `KW_SCHEMES`           | `schemes`            |
| `KW_VERSION`           | `version`            |
| `KW_HOST`              | `host`               |
| `KW_BASE_PATH`         | `basePath`           |
| `KW_LICENSE`           | `license`            |
| `KW_CONTACT`           | `contact`            |

### 1.4 Inline-value terminals

The lexer types the value following a keyword's `:` per its lexical
shape. The grammar pairs each keyword with its expected value
terminal; lexical mismatches surface as grammar parse errors.
Semantic coercion against the Go target type is the analyzer's
job — see §9.

| Terminal             | Lexer kind             | Recognises                                       |
|----------------------|------------------------|--------------------------------------------------|
| `NUMBER_VALUE`       | `TokenNumberValue`     | Signed decimal literal (integer or fractional)   |
| `INT_VALUE`          | `TokenIntValue`        | Unsigned decimal integer                         |
| `BOOL_VALUE`         | `TokenBoolValue`       | `true` / `false` (case-insensitive)              |
| `STRING_VALUE`       | `TokenStringValue`     | Verbatim non-LF text                             |
| `COMMA_LIST_VALUE`   | `TokenCommaListValue`  | Comma-separated list of strings (trim-stripped)  |
| `ENUM_OPTION_VALUE`  | `TokenEnumOptionValue` | One of a closed token set declared per keyword (e.g. `query`/`path`/… for `in:`, `csv`/`ssv`/… for `collectionFormat`) |

### 1.5 Multi-line body terminals

Single tokens spanning multiple source lines (lexer absorbs the
head and the body lines).

| Terminal                        | Parent keyword          |
|---------------------------------|-------------------------|
| `RAW_BLOCK_CONSUMES`            | `consumes`              |
| `RAW_BLOCK_PRODUCES`            | `produces`              |
| `RAW_BLOCK_SECURITY`            | `security`              |
| `RAW_BLOCK_RESPONSES`           | `responses`             |
| `RAW_BLOCK_PARAMETERS`          | `parameters`            |
| `RAW_BLOCK_EXTENSIONS`          | `extensions`            |
| `RAW_BLOCK_INFO_EXTENSIONS`     | `infoExtensions`        |
| `RAW_BLOCK_EXTERNAL_DOCS`       | `externalDocs`          |
| `RAW_BLOCK_SECURITY_DEFINITIONS`| `securityDefinitions`   |
| `RAW_BLOCK_TOS`                 | `tos`                   |
| `RAW_VALUE_DEFAULT`             | `default`               |
| `RAW_VALUE_EXAMPLE`             | `example`               |
| `RAW_VALUE_ENUM`                | `enum`                  |
| `OPAQUE_YAML`                   | (`--- … ---` fenced YAML body) |

`RAW_BLOCK_*` and `RAW_VALUE_*` are emitted whether the body is
single-line or multi-line — the single-line case is the trivial
path of the body accumulator (per Q15 / `40-lexer.md` A6).
Per-keyword body content shape is documented in `30-delegated.md`.

### 1.6 Prose & structural terminals

| Terminal | Lexer kind        | Carries                              |
|----------|-------------------|--------------------------------------|
| `TITLE`  | `TokenTitleLine`  | One prose line classified as title   |
| `DESC`   | `TokenDescLine`   | One prose line classified as description |
| `BLANK`  | `TokenBlank`      | Empty `//` line (never terminates a body) |
| `EOF`    | `TokenEOF`        | End of comment group                 |

### 1.7 Side-channel attributes

A small set of attributes ride on terminals as side-channel data
the grammar does not enumerate:

- **`ItemsDepth`** — on every `KW_*`, `RAW_BLOCK_*`, `RAW_VALUE_*`
  token: the number of leading `items.` segments stripped by the
  lexer (per Q16 / `40-lexer.md` §5). Walked by the analyzer.
- **`Pos`** — source position on every terminal. Diagnostic-only.

These are the only side-channel pieces. Argument shape, value type,
and required-vs-optional are all grammar-explicit.

### 1.8 Convenience non-terminals

```ebnf
Title       = TITLE , { TITLE } ;
Description = DESC  , { DESC  } ;
```

**Prose precedence by family.** Title and Description tokens may
appear before AND after the annotation in the comment. Two API
accessors expose them:

- `Block.ProseLines()` — every prose line, in source order, blanks
  preserved as `""`.
- `Block.PreambleLines()` — only lines that appear **before** the
  annotation. For UnboundBlock (no annotation) it equals ProseLines().

Family conventions:

- **Schema family** (`swagger:model`, `swagger:response`,
  `swagger:parameters`, `swagger:name`, fields without annotation):
  consume `PreambleLines()`. Post-annotation prose is not
  description — it classifies as body content (validation lines,
  raw-block bodies, etc.). Matches v1's SectionedParser semantics.
- **Operation family** (`swagger:route`, `swagger:operation`):
  consume `ProseLines()`. Post-annotation prose between the
  annotation header and the body fences is operation summary /
  description.
- **Meta family** (`swagger:meta`): consume `ProseLines()`.
- **Classifier family**: consume `ProseLines()`.

These are accessor conventions, not productions — the grammar treats
TITLE/DESC tokens uniformly; the family decides which subset to
read.

## 2. Top-level dispatch

```ebnf
CommentBlock     = AnnotatedBlock | UnboundBlock ;

AnnotatedBlock   = SchemaBlock
                 | OperationFamilyBlock
                 | MetaBlock
                 | ClassifierBlock ;

UnboundBlock     = [ Description ] , UnboundBlockBody ;
                  (* No annotation terminal opens the input.
                     Member-level Go declarations (struct fields,
                     interface methods, individual const / var).
                     No title — the lexer's prose classifier emits
                     only DESC tokens when no annotation is present
                     (see §8 of 40-lexer.md). *)
```

A `CommentBlock` corresponds to one Go `*ast.CommentGroup`. The
dispatcher reads the first `ANN_*` terminal: its identity selects
the family. If no annotation appears, the input is an
`UnboundBlock`.

Within a single comment block, **a second annotation closes the
preceding block and opens a fresh one** of the new annotation's
family — see Q17.

## 3. Schema family

Bodies of `swagger:model`, `swagger:parameters`, `swagger:response`,
plus the bodies struct fields produce.

```ebnf
SchemaBlock          = SchemaAnnotation
                     , [ Title ]
                     , [ Description ]
                     , SchemaAnnotationBody ;

SchemaAnnotation     = ModelAnnotation
                     | ResponseAnnotation
                     | ParametersAnnotation
                     | NameAnnotation ;

ModelAnnotation      = ANN_MODEL ,      [ IDENT_NAME ] ;             (* def, optional *)
ResponseAnnotation   = ANN_RESPONSE ,   [ IDENT_NAME ] ;             (* def, optional *)
ParametersAnnotation = ANN_PARAMETERS , IDENT_NAME , { IDENT_NAME } ; (* ref, ≥1 *)
NameAnnotation       = ANN_NAME ,       IDENT_NAME ;                 (* def: MemberName — required *)
                      (* swagger:name is a field-level rename that
                         accepts the same SchemaAnnotationBody as a
                         schema field (min length, pattern, required,
                         etc.). Dispatches under the schema family
                         even though its primary role is renaming
                         a Go member's JSON key — see fixtures/.../
                         classification/models for the canonical
                         interface-method shape. *)

SchemaAnnotationBody = { SchemaBodyItem } ;
UnboundBlockBody     = { SchemaBodyItem } ;

SchemaBodyItem       = Validation
                     | SchemaDecorator
                     | ExtensionsBlock
                     | ExternalDocsBlock
                     | BLANK ;

Validation           = NumericValidation
                     | StringValidation
                     | ArrayValidation
                     | EnumValidation
                     | RequiredLine
                     | ReadOnlyLine ;

NumericValidation    = NumericKw , NUMBER_VALUE ;
NumericKw            = KW_MAXIMUM | KW_MINIMUM | KW_MULTIPLE_OF ;

StringValidation     = KW_PATTERN , STRING_VALUE
                     | StringLengthKw , INT_VALUE ;
StringLengthKw       = KW_MAX_LENGTH | KW_MIN_LENGTH ;

ArrayValidation      = ArrayCountKw , INT_VALUE
                     | KW_UNIQUE , BOOL_VALUE
                     | KW_COLLECTION_FORMAT , ENUM_OPTION_VALUE ;
ArrayCountKw         = KW_MAX_ITEMS | KW_MIN_ITEMS ;

EnumValidation       = RAW_VALUE_ENUM ;
RequiredLine         = KW_REQUIRED , BOOL_VALUE ;       (* field-level marker — Q13 *)
ReadOnlyLine         = KW_READ_ONLY , BOOL_VALUE ;

SchemaDecorator      = RAW_VALUE_DEFAULT
                     | RAW_VALUE_EXAMPLE
                     | DiscriminatorLine
                     | DeprecatedLine ;

DiscriminatorLine    = KW_DISCRIMINATOR , BOOL_VALUE ;  (* field-level marker — Q13 *)
DeprecatedLine       = KW_DEPRECATED , BOOL_VALUE ;
                      (* Also legal under OperationFamilyBlock — see §4. *)
```

## 4. Operation family

`swagger:route` and `swagger:operation` are kept as separate block
productions reflecting the v1 corpus asymmetry — see Q1 (restated)
and Q20.

```ebnf
OperationFamilyBlock = RouteBlock | InlineOperationBlock ;

RouteBlock           = ANN_ROUTE , OperationArgs
                     , [ Title ]
                     , [ Description ]
                     , RouteBody ;

InlineOperationBlock = ANN_OPERATION , OperationArgs
                     , [ Title ]
                     , [ Description ]
                     , InlineOperationBody ;

OperationArgs        = HTTP_METHOD , URL_PATH , { IDENT_NAME } , IDENT_NAME ;
                      (* The trailing IDENT_NAME is the OperationID
                         (def: OperationID namespace). The leading
                         { IDENT_NAME } are tags. The lexer emits
                         all as IDENT_NAME; the analyzer marks the
                         trailing one as the OpID. *)

RouteBody            = { CommonOperationBodyItem | BLANK } ;

InlineOperationBody  = { CommonOperationBodyItem
                       | OPAQUE_YAML
                       | BLANK } ;
                      (* OPAQUE_YAML is legal under swagger:operation only. *)

CommonOperationBodyItem = OperationKeyword
                        | OperationDecorator
                        | OperationRawBlock
                        | ExtensionsBlock
                        | ExternalDocsBlock ;

OperationKeyword     = KW_SCHEMES , COMMA_LIST_VALUE ;

OperationDecorator   = DeprecatedLine ;
                      (* Single-element today; the production exists
                         so future operation-level decorators have a
                         home without re-touching the body item rule. *)

OperationRawBlock    = RAW_BLOCK_CONSUMES
                     | RAW_BLOCK_PRODUCES
                     | RAW_BLOCK_SECURITY
                     | RAW_BLOCK_RESPONSES
                     | RAW_BLOCK_PARAMETERS ;
```

The `<GoIdent> swagger:route ...` godoc-prefix exception is absorbed
by the lexer (`matchGodocRoutePrefix` in `lexer.go`); the EBNF sees
a plain `ANN_ROUTE` regardless.

## 5. Meta family

`swagger:meta` defines the top-of-spec metadata: version, host,
basePath, license, contact, schemes, security, externalDocs,
extensions, info-extensions, terms-of-service, plus an optional
free-form OAS YAML overlay.

```ebnf
MetaBlock         = ANN_META
                  , [ Title ]
                  , [ Description ]
                  , MetaBody ;

MetaBody          = { MetaBodyItem | BLANK } ;

MetaBodyItem      = MetaKeyword
                  | MetaRawBlock
                  | ExtensionsBlock
                  | InfoExtensionsBlock
                  | ExternalDocsBlock
                  | OPAQUE_YAML ;
                  (* OPAQUE_YAML under swagger:meta — new-feature: deep-merged
                     into the produced spec by the analyzer. *)

MetaKeyword       = MetaStringKw , STRING_VALUE
                  | KW_SCHEMES , COMMA_LIST_VALUE ;
MetaStringKw      = KW_VERSION | KW_HOST | KW_BASE_PATH | KW_LICENSE | KW_CONTACT ;

MetaRawBlock      = RAW_BLOCK_CONSUMES
                  | RAW_BLOCK_PRODUCES
                  | RAW_BLOCK_SECURITY
                  | RAW_BLOCK_SECURITY_DEFINITIONS
                  | RAW_BLOCK_TOS ;

InfoExtensionsBlock = RAW_BLOCK_INFO_EXTENSIONS ;
                  (* Same body shape as ExtensionsBlock; meta-only. *)
```

A `MetaBlock` may carry the annotation either at the top of the
comment group or after the godoc-style prose — both placements
parse the same; the lexer handles both.

## 6. Classifier family

Single-line, argument-only annotations whose role is to **classify**
a Go declaration (or comment block).

```ebnf
ClassifierBlock      = ClassifierAnnotation
                     , [ Title ]
                     , [ Description ] ;

ClassifierAnnotation = StrfmtAnnotation
                     | AliasAnnotation
                     | AllOfAnnotation
                     | IgnoreAnnotation
                     | DefaultAnnotation
                     | TypeAnnotation
                     | FileAnnotation ;
                     (* swagger:enum has its own block production —
                        see EnumDeclBlock below.

                        swagger:name is **not** a classifier — it
                        moved to SchemaAnnotation in §3 because its
                        body accepts the full SchemaAnnotationBody
                        production. *)

StrfmtAnnotation     = ANN_STRFMT , IDENT_NAME ;             (* def: Strfmt — required *)
AliasAnnotation      = ANN_ALIAS ;                           (* no args                 *)
AllOfAnnotation      = ANN_ALLOF , [ IDENT_NAME ] ;          (* PolymorphicClass (acc, optional) *)
IgnoreAnnotation     = ANN_IGNORE ;                          (* no args                 *)
DefaultAnnotation    = ANN_DEFAULT , ( JSON_VALUE | RAW_VALUE ) ;  (* value — required  *)
TypeAnnotation       = ANN_TYPE , TYPE_REF ;                 (* ref: TypeRef — required *)
FileAnnotation       = ANN_FILE ;                            (* no args; synonym of TypeAnnotation with TypeRef = "file" *)

EnumDeclBlock        = ANN_ENUM , EnumArgs
                     , [ Title ]
                     , [ Description ]
                     , [ RAW_VALUE_ENUM ] ;
                      (* The optional trailing RAW_VALUE_ENUM is the
                         multi-line value-list back-port — Q15. *)

EnumArgs             = IDENT_NAME , [ EnumValueList ]        (* name first; values optional *)
                     | EnumValueList ;                       (* values only — name inferred from Go type *)
                      (* The lexer pre-disambiguates per the four-step
                         rule (23-classifier-grammar.md). The grammar
                         sees one of the two alternatives. *)

EnumValueList        = ? lexer-emitted typed list — see 23-classifier-grammar.md ? ;
```

`DefaultAnnotation`'s `JSON_VALUE | RAW_VALUE` alternation is the
canonical site of the JSON-then-fallback dispatch from §6 of
`40-lexer.md`; the lexer tries `JSON_VALUE` first and emits
`RAW_VALUE` only on failure.

`TypeAnnotation`'s `TYPE_REF` is a closed-vocabulary terminal — if
the lexer sees a token outside the vocabulary it emits no
`TYPE_REF`, the production fails, and the grammar reports the
mismatch.

## 7. Cross-cutting productions

```ebnf
ExtensionsBlock     = RAW_BLOCK_EXTENSIONS ;
                      (* Body content interpreted as YAML mapping;
                         keys must match ExtensionName (`x-…`).
                         Decorative `---` fences sometimes seen
                         around v1 extension bodies are absorbed by
                         the lexer (§5 of 40-lexer.md). *)

ExternalDocsBlock   = RAW_BLOCK_EXTERNAL_DOCS ;
                      (* Body: `description:` / `url:` lines. *)
```

`Title` and `Description` are defined in §1.8.

## 8. Notation and conventions

- **EBNF flavour.** ISO-14977-flavoured: `=` is rule definition, `,`
  is concatenation, `|` is alternation, `[ ]` is optional, `{ }` is
  zero-or-more, `( )` groups, `"…"` is a literal terminal, `(* … *)`
  is a comment, `?…?` is an inline-prose terminal.
- **Terminals are lexer tokens.** Every ALL-CAPS identifier
  (`ANN_*`, `KW_*`, value terminals from §1.4, `RAW_BLOCK_*`,
  `RAW_VALUE_*`, `OPAQUE_YAML`, argument terminals from §1.2,
  `TITLE`, `DESC`, `BLANK`, `EOF`) is a fixed terminal defined in
  §1.
- **Required vs optional is grammar-visible.** Argument presence,
  cardinality, and pairing of keywords with value terminals are
  expressed in the productions (`[ X ]`, `{ X }`, `X , Y`). The
  lexer recognises *individual* tokens; it never enforces "this
  annotation requires an argument" or "this keyword requires a
  number" — the grammar does.
- **Lexer responsibilities.** The lexer absorbs quirks (CR/CRLF,
  trailing dot, case insensitivity on first character, godoc
  prefix on `swagger:route`, decorative fences, `items.` runs),
  classifies prose (title vs description), emits typed argument
  and value terminals based on **lexical shape only**, and
  accumulates multi-line bodies as single tokens. Everything else
  is the grammar's or the analyzer's job.

## 9. What this grammar does *not* describe

- **Semantic value coercion.** Lexical typing (`maximum: 10` →
  `KW_MAXIMUM , NUMBER_VALUE`) is grammar-visible. Semantic typing
  against the Go target field (`default: 3` → `int(3)` if the
  field is `int`, otherwise the raw string) lives in the builder
  layer via `internal/builders/validations.CoerceValue` /
  `CoerceEnum`. The lexer emits the most specific lexical shape it
  can recognise; pairings the grammar rejects produce parse errors
  at the right layer.
- **Namespace resolution.** Defining vs referencing names per Q5
  are tracked by the analyzer using the binding-role annotations
  documented in the family files. The grammar productions carry
  `(def)` / `(ref)` comments but do not enforce uniqueness.
- **Cross-block validation.** "Every reference to an OperationID
  resolves to a defined OperationID" is an analyzer pass, not a
  grammar constraint.
- **Diagnostics and recovery.** Severities, codes, and recovery
  strategies live in `internal/parsers/grammar/diagnostic.go`.

## 10. Per-file index

| Topic                                               | File                          |
|-----------------------------------------------------|-------------------------------|
| Layered overview, dispatch                          | `00-overview.md`              |
| Lexical primitives, value categories                | `10-shared.md`                |
| Schema-bearing annotations                          | `20-schema-grammar.md`        |
| Route + Operation annotations                       | `21-operation-grammar.md`     |
| Meta annotation                                     | `22-meta-grammar.md`          |
| Classifier annotations                              | `23-classifier-grammar.md`    |
| Delegated sub-languages (per-raw-block content)     | `30-delegated.md`             |
| Lexer specification (quirks, body accumulation)     | `40-lexer.md`                 |
| **This synthesis read**                             | `50-full.md`                  |
| Settled / open questions                            | `open-questions.md`           |
