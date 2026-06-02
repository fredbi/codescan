# MetaGrammar — complete EBNF

**Status:** complete (round 1) for the meta annotation family.
Standalone modulo the explicitly listed imports from `SharedGrammar`.

## Scope

`MetaGrammar` defines, in full, the legal language of the single meta
annotation:

| Annotation header   | Block kind     |
|---------------------|----------------|
| `swagger:meta`      | `MetaBlock`    |

`swagger:meta` is conventionally placed **at the end** of the package
godoc comment (per Go convention: title, description, structured
keywords, then `// swagger:meta` on the last line). The grammar
therefore allows the annotation to appear **anywhere** within the
comment block — flanked on either side by content items. Title /
Description extraction is a parser concern (the prose lines preceding
the first structured item or annotation are taken as Title +
Description, in source order); the grammar itself does not
distinguish them.

**No JSON Schema validations are legal in `MetaGrammar`.** Keywords
here are top-level API metadata.

## Imports from `SharedGrammar`

This grammar uses the following productions defined in `10-shared.md`.
They are not redefined here.

| Imported              | Role                                                    |
|-----------------------|---------------------------------------------------------|
| `LF`                  | end-of-line — `"\n" \| "\r\n"`                           |
| `EOL`                 | `[ "." ] , LF` — annotation/keyword line terminator with optional trailing dot elided |
| `Whitespace`          | one or more space / tab characters (Unicode `\p{Zs}`)   |
| `Letter`              | Unicode letter (`\p{L}`)                                |
| `Digit`               | Unicode digit (`\p{N}`)                                 |
| `PrefixTag`           | the literal `"swagger:"`                                |
| `ProseLine`           | a non-blank, non-`swagger:`, non-keyword-shaped line + `LF` |
| `BlankLine`           | `{ Whitespace } , LF`                                   |
| `StringValue`         | any non-LF text, verbatim (no escapes interpreted)      |
| `CommaListValue`      | comma-separated values, whitespace around commas trimmed |
| `ExtensionsBlock`     | `extensions:` block; body is lexer-emitted YAML token    |
| `ExtensionYAMLBody`   | lexer-emitted token: multi-line YAML mapping payload     |
| `RawBlockBody`        | lexer-emitted token: generic raw-block body payload      |
| `ExtensionName`       | `( "x-" \| "X-" ) , Letter , { Letter \| Digit \| "-" \| "_" }` — constraint on top-level keys of the parsed YAML mapping |
| `OpaqueYamlBody`      | lexer-emitted token: `--- … ---` fenced YAML payload     |

## Top-level productions

```ebnf
MetaBlock              = MetaContent , MetaAnnotation , MetaContent ;
                        (* The MetaAnnotation appears exactly once in
                           the block, flanked by content on either or
                           both sides. Conventionally the annotation
                           is at the tail; the grammar accepts it
                           anywhere. *)

MetaContent            = { MetaContentItem } ;

MetaContentItem        = MetaKeyword
                       | MetaRawBlock
                       | ExtensionsBlock         (* shared *)
                       | InfoExtensionsBlock     (* meta-only, defined below *)
                       | OpaqueYamlBody          (* free-form spec merge — see §"YAML body" *)
                       | ProseLine
                       | BlankLine ;
```

## Annotation header

```ebnf
MetaAnnotation         = PrefixTag , "meta" , EOL ;
                        (* No arguments. *)
```

## Single-line keyword productions

Each meta single-line keyword has its own production so the value
type is explicit per keyword.

```ebnf
MetaKeyword            = VersionLine
                       | HostLine
                       | BasePathLine
                       | LicenseLine
                       | ContactLine
                       | SchemesLine ;

VersionLine            = "version"  , ":" , [ Whitespace ] , StringValue    , EOL ;
HostLine               = "host"     , ":" , [ Whitespace ] , StringValue    , EOL ;
BasePathLine           = BasePathKw , ":" , [ Whitespace ] , StringValue    , EOL ;
LicenseLine            = "license"  , ":" , [ Whitespace ] , StringValue    , EOL ;
ContactLine            = ContactKw  , ":" , [ Whitespace ] , StringValue    , EOL ;
SchemesLine            = "schemes"  , ":" , [ Whitespace ] , CommaListValue , EOL ;

BasePathKw             = "basePath" | "base path" | "base-path" ;
ContactKw              = "contact"  | "contact info" | "contact-info" ;
```

The `license` and `contact` values are single-line strings that the
analyzer further parses — typically `<name> [<URL>]` for license and
`<name> [<email>] [<URL>]` for contact. The grammar captures one
verbatim line; the substructure is delegated.

`schemes` is also a legal keyword in `OperationGrammar` (it's the one
keyword that crosses meta and operation contexts). The shape is the
same in both families.

## Raw-block keyword productions

Multi-line keywords with a lexer-emitted body terminal. The grammar
declares the head; the lexer accumulates body content until the next
sibling structural item and emits a single `RawBlockBody` token —
see `00-overview.md` §"Grammar / parser boundary" and the
`RawBlockBody` definition in `10-shared.md`. Per-keyword content
interpretation is downstream (`30-delegated.md`).

```ebnf
MetaRawBlock           = ConsumesBlock
                       | ProducesBlock
                       | SecurityBlock
                       | SecurityDefsBlock
                       | TosBlock
                       | ExternalDocsBlock ;

ConsumesBlock          = "consumes"      , ":" , EOL , RawBlockBody ;
ProducesBlock          = "produces"      , ":" , EOL , RawBlockBody ;
SecurityBlock          = "security"      , ":" , EOL , RawBlockBody ;
SecurityDefsBlock      = SecurityDefsKw  , ":" , EOL , RawBlockBody ;
TosBlock               = TosKw           , ":" , EOL , RawBlockBody ;
ExternalDocsBlock      = ExternalDocsKw  , ":" , EOL , RawBlockBody ;

SecurityDefsKw         = "securityDefinitions" | "security definitions" | "security-definitions" ;
TosKw                  = "tos" | "terms of service" | "terms-of-service" | "termsOfService" ;
ExternalDocsKw         = "externalDocs" | "external docs" | "external-docs" ;
```

The lexer tags the emitted `RawBlockBody` with the opening keyword so
the analyzer can dispatch to the correct sub-parser. The grammar
itself doesn't need a per-family terminator-set production —
boundary detection lives in the lexer.

Per-keyword body content shapes (delegated):

| Block               | Body content shape                                    |
|---------------------|-------------------------------------------------------|
| `ConsumesBlock`     | List of MIME-type strings, one per line, optional `-` prefix |
| `ProducesBlock`     | List of MIME-type strings, one per line, optional `-` prefix |
| `SecurityBlock`     | OAS security-requirement list (YAML-shaped)           |
| `SecurityDefsBlock` | OAS security-scheme map (YAML-shaped)                 |
| `TosBlock`          | Single string (URL or text)                           |
| `ExternalDocsBlock` | OAS external-docs object: `description:` / `url:` lines |

## Info-extensions block (meta-only)

`InfoExtensionsBlock` declares `x-*` vendor extensions on the OpenAPI
`info` object specifically. It mirrors `ExtensionsBlock` (top-level
extensions) shape-wise but targets a different output location.

```ebnf
InfoExtensionsBlock    = InfoExtensionsHead , EOL , ExtensionYAMLBody ;

InfoExtensionsHead     = ( "infoExtensions" | "info extensions" | "info-extensions" ) , ":" ;
```

`ExtensionYAMLBody` and `ExtensionName` are imported from
`SharedGrammar` — `InfoExtensionsBlock` and `ExtensionsBlock` share
the body terminal; only the head differs.

## YAML body (free-form spec merge)

`swagger:meta` accepts an optional `--- … ---` fenced YAML body that
the analyzer merges into the produced spec. This complements the
explicit single-line and raw-block keywords above and lets users
express anything that isn't covered by a dedicated keyword (or that
they prefer to write as inline YAML).

```ebnf
OpaqueYamlBody         = "---" , LF , { ? any line ? } , "---" , LF ;
                        (* Captured as a single opaque terminal; the
                           grammar establishes block boundaries
                           (between the two "---" fences). YAML
                           parsing and merge-into-spec are the
                           analyzer's job. *)
```

Merge semantics (analyzer concern, not grammar):

- Explicit keywords (`Version:`, `Host:`, etc.) take precedence over
  values supplied by the YAML body.
- The YAML body fills in fields the explicit keywords do not cover.
- The YAML body's structure mirrors the OpenAPI top-level spec
  object (so users familiar with OpenAPI can write the YAML directly).

A `MetaBlock` may carry zero or more `OpaqueYamlBody` blocks; the
analyzer merges them in source order.

## Annotation → keyword table

| Keyword                | Production                | Value shape    | Notes                                         |
|------------------------|---------------------------|----------------|-----------------------------------------------|
| `version`              | `VersionLine`             | `StringValue`  | API version string.                           |
| `host`                 | `HostLine`                | `StringValue`  | Host (and optional port).                     |
| `basePath`             | `BasePathLine`            | `StringValue`  | URL prefix for all API paths.                 |
| `license`              | `LicenseLine`             | `StringValue`  | License info — name, optional URL.            |
| `contact`              | `ContactLine`             | `StringValue`  | Contact info — name, email, URL.              |
| `schemes`              | `SchemesLine`             | `CommaListValue` | API schemes (http, https, ws, wss). Shared with OperationGrammar. |
| `consumes`             | `ConsumesBlock`           | raw block      | Default MIME types the API consumes.          |
| `produces`             | `ProducesBlock`           | raw block      | Default MIME types the API produces.          |
| `security`             | `SecurityBlock`           | raw block      | Default security requirements.                |
| `securityDefinitions`  | `SecurityDefsBlock`       | raw block      | Declared security schemes (apiKey, basic, oauth2). |
| `tos`                  | `TosBlock`                | raw block      | Terms-of-service URL or text.                 |
| `externalDocs`         | `ExternalDocsBlock`       | raw block      | External documentation reference.             |
| `extensions`           | `ExtensionsBlock` (shared) | raw block      | Custom `x-*` vendor extensions at the spec level. |
| `infoExtensions`       | `InfoExtensionsBlock` (local) | raw block | Custom `x-*` extensions on the info block.    |

## Examples

Each example shows the comment block (preprocessed: `// ` prefixes
stripped) that the grammar parses.

### Trailing annotation (canonical godoc form)

```
Package booking API.

the purpose of this application is to provide an application
that is using plain go code to define an API

Schemes: https
Host: localhost
Version: 0.0.1

Consumes:
- application/json

Produces:
- application/json

swagger:meta
```

The annotation `swagger:meta` is at the tail. Title = "Package
booking API.", Description = the second paragraph. The structured
keywords precede the annotation; the grammar accepts them as the
`MetaContent` before `MetaAnnotation`.

### Leading annotation

```
swagger:meta

Schemes: http, https
Host: api.example.com
BasePath: /v2
Version: 1.4.0
License: MIT https://opensource.org/licenses/MIT
Contact: API Team team@example.com https://example.com/support
```

Annotation first, structured keywords after. Equally valid.

### With trailing dot (godot-linter style)

```
swagger:meta.
```

```
Version: 0.0.1.
Host: localhost.
```

The `EOL` production elides the trailing `.` on annotation and
keyword lines.

### Raw blocks with `-` list markers

```
swagger:meta

Consumes:
  - application/json
  - application/xml

Produces:
  - application/json

Security:
  - apiKey: []
  - oauth2:
    - read
    - write
```

### `securityDefinitions` and `externalDocs`

```
swagger:meta

SecurityDefinitions:
  apiKey:
    type: apiKey
    name: X-API-Key
    in: header

ExternalDocs:
  description: Find more info here
  url: https://example.com/docs
```

### Vendor extensions

```
swagger:meta

Extensions:
  x-tagGroups:
    - name: User Management
      tags: [users, accounts]

InfoExtensions:
  x-logo: https://example.com/logo.png
```

### Free-form YAML body (new feature)

```
swagger:meta

---
info:
  title: My API
  termsOfService: https://example.com/tos
tags:
  - name: users
    description: User management endpoints
servers:
  - url: https://api.example.com/v1
---
```

The YAML body merges into the produced spec; explicit keywords (when
also present) take precedence over YAML-supplied values. Multiple
YAML bodies in one block are merged in source order.

### Mixing explicit keywords and a YAML body

```
Booking API.

Version: 1.0.0
Host: api.example.com

---
tags:
  - name: bookings
  - name: users
externalDocs:
  description: Onboarding guide
  url: https://example.com/onboarding
---

swagger:meta
```

The explicit `Version:` and `Host:` set the canonical fields; the
YAML body adds tags and `externalDocs`. Annotation appears at the
tail per the godoc convention.

## Productions glossary

For self-check completeness, every production used in this grammar:

**Defined locally (this file):**

`MetaBlock`, `MetaContent`, `MetaContentItem`, `MetaAnnotation`,
`MetaKeyword`, `VersionLine`, `HostLine`, `BasePathLine`,
`LicenseLine`, `ContactLine`, `SchemesLine`, `BasePathKw`,
`ContactKw`, `MetaRawBlock`, `ConsumesBlock`, `ProducesBlock`,
`SecurityBlock`, `SecurityDefsBlock`, `TosBlock`,
`ExternalDocsBlock`, `SecurityDefsKw`, `TosKw`, `ExternalDocsKw`,
`InfoExtensionsBlock`, `InfoExtensionsHead`.

**Imported from `SharedGrammar` (10-shared.md):**

`LF`, `EOL`, `Whitespace`, `Letter`, `Digit`, `PrefixTag`,
`ProseLine`, `BlankLine`, `StringValue`, `CommaListValue`,
`ExtensionsBlock`, `ExtensionYAMLBody`, `ExtensionName`,
`RawBlockBody`, `OpaqueYamlBody`.

No undefined references. The grammar is closed.

## Notes

- The annotation may appear anywhere in the block; conventionally it
  sits at the tail (after Title + Description + structured keywords)
  per Go-godoc convention.
- `schemes` is the one keyword that crosses families: it's defined
  here and also in `OperationGrammar`. The lexical shape is identical.
- `ExtensionsBlock` is imported from `SharedGrammar` since it's used
  by Schema, Operation, and Meta. `InfoExtensionsBlock` is defined
  here because it's meta-only — it shares `ExtensionYAMLBody` and
  `ExtensionName` with `ExtensionsBlock`; only the head differs.
- All multi-line bodies (`ExtensionYAMLBody`, `RawBlockBody`,
  `OpaqueYamlBody`) are **lexer-emitted single tokens**, not
  decomposed line-by-line in the grammar. The lexer accumulates body
  content until the next sibling structural item, elides decorative
  fences (e.g. `---` inside `Extensions:` bodies), and emits one
  token per body. See `00-overview.md` §"Grammar / parser boundary".
  Empirical verification of YAML-typed extension bodies:
  `.worktrees/baseline/fenced_ext_probe_test.go`.
- All raw-block bodies are opaque to the grammar; per-keyword content
  shape is delegated to sub-parsers — see `30-delegated.md`.
- `OpaqueYamlBody` is a **new feature** in this round: a `--- … ---`
  fenced YAML body that the analyzer merges into the produced spec.
  Mirrors how `swagger:operation` already supports a YAML body.
  Grammar-level cost: trivial (one new alternative in
  `MetaContentItem`). Analyzer cost: medium — needs deep-merge logic
  with explicit-keyword precedence.
- `EOL` is used as the line terminator on every annotation and
  keyword line so that a trailing `.` (godot-linter style) is
  elided. Prose lines and blank lines use `LF`.
- Diagnostics, error recovery, and per-keyword body validation are
  *not* described by this grammar — they are the parser's and
  analyzer's responsibility. The grammar defines only the legal
  language.
