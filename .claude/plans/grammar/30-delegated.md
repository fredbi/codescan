# Delegated sub-languages — registry of opaque terminals

This file lists the embedded mini-languages the EBNF treats as opaque.
For each, the grammar establishes block boundaries (head, terminator);
the hand-rolled parser delegates content validation to a sub-parser.

This file is a **registry**, not a grammar — it does not define
productions, only documents what's delegated and where the sub-parser
lives.

## Top-level opaque terminals

| Terminal           | Where it appears                                                | Sub-parser today                   |
|--------------------|-----------------------------------------------------------------|------------------------------------|
| `OpaqueYamlBody`   | `InlineOperationBody` (under `swagger:operation`); `MetaContent` (under `swagger:meta` — new feature) | `internal/parsers/yaml/`           |
| `RawValueBody`     | `default:`, `example:`, `enum:` keyword bodies (`SchemaGrammar`); also the `swagger:enum` annotation argument (multi-line back-port — see Q15) | `internal/parsers/helpers/ParseValueFromSchema` (default/example) and the same helper plus `internal/parsers/helpers/ParseEnum` (enum) — see also `internal/parsers/enum/` (W2 follow-up) |

The body of `swagger:route` is **not** opaque at the EBNF layer:
it is a sequence of inline keyword raw-blocks (`consumes:` /
`produces:` / `security:` / `responses:` / `parameters:` /
`extensions:` / `externalDocs:`). Only the *bodies* of `parameters:`
/ `responses:` / `extensions:` under `swagger:route` are delegated
to opaque sub-parsers, and they are listed below in the per-raw-block
content table — not as a top-level terminal.

### `OpaqueYamlBody`

```
---
parameters:
  - name: limit
    in: query
    …
responses:
  "200":
    …
---
```

A `--- … ---` fenced body containing an OpenAPI-spec fragment. The
grammar captures the bytes between fences; structural validation is
out of scope.

### `RawValueBody`

```
default: 42

example: |
  {
    "name": "Pet",
    "tags": ["dog", "white"]
  }

enum:
  - low
  - medium
  - high

enum: ["low", "medium", "high"]

enum: low, medium, high
```

A unified body terminal for `default:`, `example:`, and `enum:`
under any schema-bearing block. The lexer accumulates content from
just after the `:` until the next sibling structural item (next
keyword, next annotation, EOF), independent of whether the body is
single-line scalar, multi-line YAML/JSON, comma-list, or bracketed
list. Downstream sub-parsers interpret per keyword:

- `default:` / `example:` → `helpers.ParseValueFromSchema` against
  the inferred Go target type.
- `enum:` → re-parsed as an `EnumValueList` (per-item typing via
  `helpers.ParseEnum` / `helpers.ParseValueFromSchema`).

The `swagger:enum` annotation argument shares the same `RawValueBody`
shape (Q15 back-port) — a single-line argument falls out as the
trivial case.

## Per-raw-block content delegations

Each `RawBlock` head carries its own sub-language for body content.
The grammar establishes the block boundary; per-keyword content is
delegated.

| Raw-block head            | Family            | Delegated content shape                            |
|---------------------------|-------------------|----------------------------------------------------|
| `consumes`                | Operation, Meta   | List of MIME-type strings, one per line, optional `-` prefix |
| `produces`                | Operation, Meta   | List of MIME-type strings, one per line, optional `-` prefix |
| `security`                | Operation, Meta   | OAS security-requirement list (YAML-shaped)        |
| `securityDefinitions`     | Meta              | OAS security-scheme map (YAML-shaped)              |
| `responses`               | Operation         | Under `swagger:route`: `<status>: <responseRef>` map (sub-parser: `internal/parsers/routebody/responses.go`). Under `swagger:operation`: YAML-shaped, normally inside the `OpaqueYamlBody`. |
| `parameters`              | Operation         | Under `swagger:route`: `+ name:` indented continuation list (sub-parser: `internal/parsers/routebody/route_params.go`). Under `swagger:operation`: YAML-shaped, normally inside the `OpaqueYamlBody`. |
| `tos`                     | Meta              | Single string (URL or text)                        |
| `externalDocs`            | Schema, Operation, Meta | OAS external-docs object: `description:` / `url:` lines |
| `extensions`, `infoExtensions` | Schema, Operation, Meta | Lexer-emitted `ExtensionYAMLBody` token: multi-line YAML mapping. Top-level keys must match `ExtensionName` (`x-*`); values are arbitrary YAML (booleans, numbers, strings, arrays, objects, null). YAML type inference falls out for free. Under `swagger:route`, the route-side parser (`internal/parsers/routebody/extensions.go`) handles the same YAML-shaped body. |
| `default`, `example`      | Schema            | Lexer-emitted `RawValueBody` token: single-line literal or multi-line block; analyzer interprets via `ParseValueFromSchema` against the target Go type. |
| `enum`                    | Schema            | Lexer-emitted `RawValueBody` token: re-parsed as `EnumValueList` (plain comma list, bracketed list, or YAML-ish multi-line list); per-item typing via the same helper. |

## Lexer-emitted body terminals

All multi-line bodies in this grammar are **lexer-emitted single
tokens**, not line-streams the grammar walks. The lexer's state
machine is responsible for:

- accumulating body content until the next sibling structural item,
- detecting and consuming opening / closing markers (YAML `---`
  fences, indentation thresholds for `OpaqueRouteBody`, etc.),
- eliding decorative fences (e.g. `---` inside an `ExtensionYAMLBody`
  body — see "historical note" below).

The grammar references the result as a terminal. See
`00-overview.md` §"Grammar / parser boundary" for the full token
inventory.

## `ExtensionsBlock` / `InfoExtensionsBlock`

The `extensions:` head is spelled out in `SharedGrammar`; the body
is the lexer-emitted `ExtensionYAMLBody` terminal — a multi-line
YAML mapping fed to a YAML sub-parser. `InfoExtensionsBlock` is
meta-only and is defined in `22-meta-grammar.md`; both reuse
`ExtensionYAMLBody` and `ExtensionName`.

The earlier `---`-fenced wrapping form for these blocks is
**removed** from the grammar. Empirical probe (baseline worktree
fixture `fixtures/goparsing/meta/fenced-ext/` plus
`fenced_ext_probe_test.go`) confirms fenced and unfenced bodies
produce byte-identical output — `yaml.Unmarshal` handles the `---`
separator natively (multi-document YAML), the v1 parser does no
fence-specific work beyond consuming the lines. In the v2 lexer,
fence elision is explicit during `ExtensionYAMLBody` accumulation.
Free-form structured content under `swagger:meta` is expressed via
`OpaqueYamlBody` instead.

## New features

- **`OpaqueYamlBody` under `swagger:meta`**: a `--- … ---` fenced YAML
  body that the analyzer merges into the produced spec. Mirrors the
  existing `swagger:operation` YAML support. Grammar-level cost was
  trivial (one new alternative in `MetaContentItem`); analyzer cost
  is medium — needs deep-merge logic with explicit-keyword precedence.

## Round-1 stance

Round 1 (this draft) leaves all sub-languages in this registry as
opaque. Folding any of them into the EBNF (so the parser can subsume
the corresponding sub-parser) is a follow-up — see `open-questions.md`
Q1.
