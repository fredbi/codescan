# OperationGrammar — route and operation annotations

**Imports from `SharedGrammar`:**
`PrefixTag`, `Prelude`, `ProseLine`, `BlankLine`, `Whitespace`, `EOL`,
`LF`, `Letter`, `Digit`, `IdentifierName`, `GoIdentifier`,
`ExtensionsBlock`, `ExternalDocsBlock`, `RawBlockBody`,
all value categories.

**Imports from `SchemaGrammar`:** `DeprecatedLine` (the `deprecated:`
keyword is legal here too — wired into `OperationDecorator` below).

**Delegated terminals:** `OpaqueYamlBody` — see `30-delegated.md`.
The `parameters:` / `responses:` / `extensions:` body content under
`swagger:route` is delegated per-keyword (also in `30-delegated.md`),
not as a single opaque body terminal.

## Scope

`OperationGrammar` covers the two operation-bearing annotations as
**two distinct block productions**, reflecting the v1 corpus
asymmetry:

| Block kind            | Annotation header   | Body shape                                                                 |
|-----------------------|---------------------|----------------------------------------------------------------------------|
| `RouteBlock`          | `swagger:route`     | Prose + inline keyword raw-blocks. **No `--- … ---` YAML fence.**          |
| `InlineOperationBlock`| `swagger:operation` | Prose + plain keywords + (optional) `OpaqueYamlBody` for OAS-fragment YAML.|

Both bodies share the same set of inline keyword raw-blocks
(`consumes:` / `produces:` / `security:` / `responses:` / `parameters:` /
`schemes:` / `extensions:` / `externalDocs:` / `deprecated:`); only
the YAML body is exclusive to `swagger:operation`. The per-keyword
*body shape* of `parameters:` / `responses:` / `extensions:` differs
between the two: under `swagger:route` they take the v1 routebody
sub-languages (`+ name:` lists, `<status>: <ref>` maps); under
`swagger:operation` they conventionally live inside the
`OpaqueYamlBody` and follow YAML conventions there. See
`30-delegated.md`.

**No JSON Schema validations are legal in `OperationGrammar`.** The
keywords here are operation/route metadata only.

## Block productions

```ebnf
OperationFamilyBlock = RouteBlock | InlineOperationBlock ;

RouteBlock           = Prelude , RouteAnnotation , RouteBody ;
InlineOperationBlock = Prelude , InlineOperationAnnotation , InlineOperationBody ;
```

## Annotation headers

Each annotation has its own header production, sharing the
`OperationArgs` realisation of the `RouteSpec` argument-shape category.

```ebnf
RouteAnnotation           = PrefixTag , "route"     , Whitespace , OperationArgs , EOL ;
InlineOperationAnnotation = PrefixTag , "operation" , Whitespace , OperationArgs , EOL ;

OperationArgs    = HttpMethod , Whitespace , UrlPath
                 , [ Whitespace , FreeTextTags ]
                 , Whitespace , OperationID ;
                  (* Realisation of the `RouteSpec` argument shape
                     category — see `10-shared.md` §"Annotation
                     argument shape categories". *)

OperationID      = IdentifierName ;          (* def: OperationID namespace *)
                  (* The trailing token of swagger:route /
                     swagger:operation introduces a fresh operation
                     id. Referenced from `swagger:parameters`
                     annotations as `OperationIDRef`. *)

HttpMethod       = Letter , { Letter } ;            (* GET, POST, … *)
UrlPath          = "/" , { UrlChar | "{" PathParamName "}" | "/" } ;
PathParamName    = Letter , { Letter | Digit | "_" | "-" } ;
UrlChar          = ? URL-safe character (RFC 3986 unreserved + sub-delims) ? ;
FreeTextTags     = WordRun , { Whitespace , WordRun } ;
WordRun          = ? non-whitespace run, not the trailing OperationID ? ;
```

### Godoc-prefix exception (route only)

`swagger:route` is the **only** annotation allowed to follow a single
godoc-style identifier on the same line. The leading identifier is a
Go function name and is discarded.

```ebnf
RouteWithGodocPrefix = GoIdentifier , Whitespace , RouteAnnotation ;
```

The lexer recognises this shape during line classification (see
`internal/parsers/grammar/lexer.go` `matchGodocRoutePrefix`): when it
sees a comment line of the form `<GoIdent> swagger:route <args>`, it
strips the leading identifier and emits the rest as a plain
`RouteAnnotation` token. The grammar therefore sees the same
`RouteAnnotation` shape regardless of whether the godoc prefix was
present in the source.

## Bodies

```ebnf
RouteBody              = { CommonOperationBodyItem | ProseLine | BlankLine } ;

InlineOperationBody    = { CommonOperationBodyItem
                         | OpaqueYamlBody
                         | ProseLine | BlankLine } ;

CommonOperationBodyItem = OperationKeyword
                        | OperationDecorator
                        | OperationRawBlock
                        | ExtensionsBlock
                        | ExternalDocsBlock ;
```

`OpaqueYamlBody` — the `--- … ---` fenced OAS-fragment body — is
**legal under `swagger:operation` only**. Recognising it under
`swagger:route` is a grammar error.

## Operation keywords (single-line)

```ebnf
OperationKeyword = SchemesLine ;

SchemesLine      = "schemes" , ":" , [ Whitespace ] , CommaListValue , EOL ;
```

## Operation decorators (single-line)

```ebnf
OperationDecorator = DeprecatedLine ;
                    (* DeprecatedLine is imported from SchemaGrammar
                       (`KW_DEPRECATED` in the synthesis read). The
                       production exists so future operation-level
                       decorators have a home without re-touching
                       `CommonOperationBodyItem`. *)
```

## Operation raw-blocks (multi-line)

Each raw-block production has the same shape: `Head , EOL , RawBlockBody`.
The lexer accumulates body content until the next sibling structural
item; per-keyword interpretation lives downstream (see `30-delegated.md`).

```ebnf
OperationRawBlock = ConsumesBlock
                  | ProducesBlock
                  | SecurityBlock
                  | ResponsesBlock
                  | ParametersBlock ;

ConsumesBlock     = "consumes"   , ":" , EOL , RawBlockBody ;
ProducesBlock     = "produces"   , ":" , EOL , RawBlockBody ;
SecurityBlock     = "security"   , ":" , EOL , RawBlockBody ;
ResponsesBlock    = "responses"  , ":" , EOL , RawBlockBody ;
ParametersBlock   = "parameters" , ":" , EOL , RawBlockBody ;
```

The body shape *inside* each `RawBlockBody` depends on which
annotation opened the enclosing block:

| Keyword       | Under `swagger:route` (delegated to `internal/parsers/routebody/`) | Under `swagger:operation`           |
|---------------|---------------------------------------------------------------------|--------------------------------------|
| `parameters:` | `+ name:` continuation list (`route_params.go`)                     | YAML-shaped (rare at top level; usually inside `OpaqueYamlBody`) |
| `responses:`  | `<status>: <responseRef>` map (`responses.go`)                      | YAML-shaped (rare at top level; usually inside `OpaqueYamlBody`) |
| `extensions:` | YAML-shaped extensions block (`extensions.go`)                      | Same shared `ExtensionYAMLBody`       |
| `consumes:`, `produces:`, `security:` | one-item-per-line / OAS security YAML — same shape under both annotations | Same                                 |

`externalDocs:` is **not** an operation-specific raw-block — it is
the shared `ExternalDocsBlock` defined in `SharedGrammar` and is
recognised at the body level alongside `ExtensionsBlock`.

## Examples

### `swagger:route` with godoc prefix and inline raw-blocks

```go
// GetPets swagger:route GET /pets pets listPets
//
// Lists the pets known to the store.
//
// Consumes:
// - application/json
// - application/x-protobuf
//
// Produces:
// - application/json
//
// Schemes: http, https
//
// Security:
//   api_key:
//   oauth: read, write
//
// Parameters:
// + name:        request
//   description: The request model.
//   in:          body
//   type:        petModel
// + name:        id
//   description: The pet id
//   in:          path
//   required:    true
//
// Responses:
//   default: body:genericError
//   200: body:someResponse
//   422: body:validationError
//
// Extensions:
//   x-some-flag: true
func GetPets(w http.ResponseWriter, r *http.Request) {}
```

- Lexer strips the leading `GetPets` per the godoc-prefix exception.
- `Consumes:` / `Produces:` / `Schemes:` / `Security:` open
  `RawBlockBody` lexer tokens; their content is per-keyword YAML-ish.
- `Parameters:` opens a `RawBlockBody` whose interior is the
  routebody `+ name:` continuation list — re-parsed downstream by
  `internal/parsers/routebody/route_params.go`.
- `Responses:` opens a `RawBlockBody` whose interior is the
  `<status>: <ref>` routebody map — re-parsed by `responses.go`.
- `Extensions:` opens a `RawBlockBody` re-parsed as YAML by
  `extensions.go` (under route).
- **No `--- … ---` fence appears.** Trying to add one would be a
  grammar error under `swagger:route`.

### `swagger:operation` with `OpaqueYamlBody`

```go
// SomeFunc do something
//
// swagger:operation POST /api/v1/somefunc someFunc
//
// Do something
//
// Deprecated: false
//
// ---
// parameters:
//   - name: body
//     in: body
//     required: true
//     schema:
//       $ref: "#/definitions/pet"
// responses:
//   "200":
//     description: pet found
//     schema:
//       $ref: "#/definitions/pet"
//   default:
//     description: error
//     schema:
//       $ref: "#/definitions/genericError"
func SomeFunc(rw http.ResponseWriter, req *http.Request) {}
```

- `Deprecated:` is an `OperationDecorator` (the imported
  `DeprecatedLine` from `SchemaGrammar`).
- The `--- … ---` fence opens an `OpaqueYamlBody`; the lexer
  captures the interior as a single token, fed to the YAML
  sub-parser (`internal/parsers/yaml/`).
- All `parameters:` / `responses:` structure lives inside the YAML
  body — they are not parsed by `OperationGrammar` itself.

### Top-level inline raw-blocks under `swagger:operation`

```go
// swagger:operation POST /pets pets createPet
//
// Creates a new pet.
//
// consumes:
//   - application/json
// produces:
//   - application/json
// schemes: http, https
//
// ---
// parameters:
//   - name: body
//     in: body
//     required: true
//     schema:
//       $ref: "#/definitions/pet"
// responses:
//   "201":
//     description: pet created
func CreatePet(w http.ResponseWriter, r *http.Request) {}
```

- `consumes:` / `produces:` are inline raw-blocks at the top level
  of the operation body — legal but rare in v1 corpora (per
  `internal/parsers/grammar/context_test.go:76`).
- The `OpaqueYamlBody` follows. The two coexist; per-keyword
  precedence is an analyzer concern (top-level `consumes:` lines
  union with any `consumes:` inside the YAML body).

## Productions glossary

Local productions defined in this file:

`OperationFamilyBlock`, `RouteBlock`, `InlineOperationBlock`,
`RouteAnnotation`, `InlineOperationAnnotation`, `OperationArgs`,
`OperationID`, `HttpMethod`, `UrlPath`, `PathParamName`, `UrlChar`,
`FreeTextTags`, `WordRun`, `RouteWithGodocPrefix`, `RouteBody`,
`InlineOperationBody`, `CommonOperationBodyItem`, `OperationKeyword`,
`SchemesLine`, `OperationRawBlock`, `ConsumesBlock`, `ProducesBlock`,
`SecurityBlock`, `ResponsesBlock`, `ParametersBlock`.

Imported (used but defined elsewhere): `PrefixTag`, `Prelude`,
`ProseLine`, `BlankLine`, `Whitespace`, `EOL`, `LF`, `Letter`, `Digit`,
`IdentifierName`, `GoIdentifier`, `ExtensionsBlock`,
`ExternalDocsBlock`, `RawBlockBody`, `CommaListValue`, `DeprecatedLine`,
`OpaqueYamlBody`.
