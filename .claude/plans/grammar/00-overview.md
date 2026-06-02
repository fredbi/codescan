# Codescan annotation grammar — overview

**Status:** draft for iteration · **Layout:** layered, one file per
sub-grammar.

This directory holds the EBNF for codescan's swagger annotation
comments, split into a handful of small sub-grammars rather than one
monolithic file. The top grammar dispatches by *family*; each family
has its own self-contained sub-grammar.

## Files

| File                        | Sub-grammar             | Scope                                             |
|-----------------------------|-------------------------|---------------------------------------------------|
| `10-shared.md`              | `SharedGrammar`         | Lexical primitives, value categories, prose, raw-block shape, extensions block — used by all families |
| `20-schema-grammar.md`      | `SchemaGrammar`         | `swagger:model`, `swagger:parameters`, `swagger:response`, `UnboundBlock` |
| `21-operation-grammar.md`   | `OperationGrammar`      | `swagger:route`, `swagger:operation`              |
| `22-meta-grammar.md`        | `MetaGrammar`           | `swagger:meta`                                    |
| `23-classifier-grammar.md`  | `ClassifierGrammar`     | `swagger:strfmt`, `:alias`, `:name`, `:allOf`, `:enum`, `:ignore`, `:default`, `:type`, `:file` |
| `30-delegated.md`           | (registry, not a grammar) | Opaque sub-language terminals — YAML, route-body, per-raw-block content |
| `40-lexer.md`               | (lexer spec, not a grammar) | Quirks absorption, line classification, multi-line body accumulation, disambiguation rules |
| `50-full.md`                | (synthesis read)        | Full grammar in one file, written against the lexer's terminal vocabulary |
| `60-implementation-roadmap.md` | (handoff)            | Spec-to-code roadmap: pre-flight blockers, phases, sequencing |
| `open-questions.md`         | —                       | Running list of decisions to resolve              |

Each sub-grammar imports the productions it needs from `SharedGrammar`
and possibly cross-references another family's sub-grammar; imports
are listed at the top of each file.

## Top-level dispatch

The top grammar is small enough to live here:

```ebnf
CommentBlock     = AnnotatedBlock | UnboundBlock ;

AnnotatedBlock   = SchemaBlock        (* see 20-schema-grammar.md     *)
                 | OperationBlock     (* see 21-operation-grammar.md  *)
                 | MetaBlock          (* see 22-meta-grammar.md       *)
                 | ClassifierBlock ;  (* see 23-classifier-grammar.md *)

UnboundBlock     = ? see 20-schema-grammar.md, §UnboundBlock ? ;
```

A `CommentBlock` corresponds to one Go comment group. The dispatcher
finds the (at most one) `swagger:<name>` line in the block, looks up
the family for `<name>`, and parses the rest under that family's
sub-grammar. With no `swagger:` line, the block is an `UnboundBlock`
and `SchemaGrammar` applies (struct-field docstring case).

## Annotation → family table

The exhaustive routing table:

| Annotation             | Family       | Sub-grammar          |
|------------------------|--------------|----------------------|
| `swagger:model`        | Schema       | `SchemaGrammar`      |
| `swagger:parameters`   | Schema       | `SchemaGrammar`      |
| `swagger:response`     | Schema       | `SchemaGrammar`      |
| `swagger:route`        | Operation    | `OperationGrammar`   |
| `swagger:operation`    | Operation    | `OperationGrammar`   |
| `swagger:meta`         | Meta         | `MetaGrammar`        |
| `swagger:strfmt`       | Classifier   | `ClassifierGrammar`  |
| `swagger:alias`        | Classifier   | `ClassifierGrammar`  |
| `swagger:name`         | Classifier   | `ClassifierGrammar`  |
| `swagger:allOf`        | Classifier   | `ClassifierGrammar`  |
| `swagger:enum`         | Classifier   | `ClassifierGrammar`  |
| `swagger:ignore`       | Classifier   | `ClassifierGrammar`  |
| `swagger:default`      | Classifier   | `ClassifierGrammar`  |
| `swagger:type`         | Classifier   | `ClassifierGrammar`  |
| `swagger:file`         | Classifier   | `ClassifierGrammar`  |
| *(no annotation)*      | Schema       | `SchemaGrammar` (UnboundBlock) |

## Design principles

1. **Strict EBNF.** The grammar describes the *legal* language.
   Diagnostics and error recovery live in the hand-rolled parser,
   never in the grammar.
2. **Layered, not monolithic.** Each family is its own sub-grammar so
   each one is reviewable in isolation. Shared primitives are imported
   from `SharedGrammar`.
3. **Embedded sub-languages are opaque terminals.** YAML bodies, the
   route-body indented list, and OAS-spec fragments under raw-block
   keywords are named as terminals in the EBNF and listed in
   `30-delegated.md`. The hand-rolled parser establishes block
   boundaries; sub-parsers validate content.
4. **Input is preprocessed lines.** The EBNF describes content *after*
   `Preprocess()` strips `// ` / `/* */` markers and normalizes
   indentation. The Go comment-marker grammar is out of scope.
5. **Notation:** ISO-14977-flavoured EBNF (`=`, `,`, `|`, `[ ]`,
   `{ }`, `( )`, terminals in `"…"`, `(* comments *)`, `?…?` for
   inline-prose terminals).

## Grammar / parser boundary — lexer handles multi-line bodies

Pure EBNF describes context-free languages. The grammar handles
context-sensitive multi-line constructs by **delegating body
accumulation to the lexer**: every multi-line body is emitted as a
single token rather than as a stream of lines. The grammar then
references those tokens as terminals; the EBNF stays clean and
context-free; "spans until the next sibling…" rules never appear in
the grammar.

The lexer-emitted body terminals are:

| Token              | Where it appears                                                | Lexer responsibilities                                                   |
|--------------------|-----------------------------------------------------------------|---------------------------------------------------------------------------|
| `ExtensionYAMLBody` | `Extensions:` / `InfoExtensions:` blocks                       | accumulate body lines until next sibling structural item; elide decorative `---` fences; emit YAML payload as one token |
| `RawBlockBody`     | `Consumes:`, `Produces:`, `Security:`, `SecurityDefinitions:`, `Tos:`, `ExternalDocs:`, … | accumulate body lines until next sibling structural item; emit payload as one token (per-keyword interpretation downstream — see `30-delegated.md`) |
| `RawValueBody`     | `default:`, `example:`, `enum:` keyword bodies (schema)                                  | accumulate body content from after the `:` until next sibling structural item — covers single-line scalars, multi-line YAML/JSON, and `enum:`'s value-list; emit as one token (per-keyword interpretation downstream) |
| `OpaqueYamlBody`   | `--- … ---` fenced body under `swagger:operation` / `swagger:meta` | recognise opening / matching closing fence; emit interior as one token  |
| `OpaqueRouteBody`  | indented `Parameters:` / `Responses:` / `Extensions:` lists under `swagger:route` | accumulate by indentation; emit payload as one token                  |

Other lexer concerns (also context-sensitive, also out of the
grammar): trailing-dot elision (`EOL = [ "." ] , LF`) and
comment-marker preprocessing (`// ` / `/* */` stripping).

What this buys us:

- The grammar has **no per-family terminator-set productions**. There
  is no `MetaRawBlockTerminator`, `OperationRawBlockTerminator`,
  etc. — those concerns are inside the lexer's state machine.
- Every raw-block-style production is the same shape:
  `Head , EOL , BodyToken`. Easy to read, easy to implement.
- The EBNF is the contract for *one* layer (line-structure +
  annotation dispatch + keyword recognition). The lexer is the
  contract for the *other* layer (line-classification, body
  accumulation, fence/indentation handling).

A reader can implement the grammar correctly given the EBNF *plus*
the lexer specification (which defines how each body terminal is
recognised). Neither half subsumes the other.

## Where this lives

For now: `.claude/plans/grammar/` (maintainer-facing). Once stable,
the candidate destinations are:

- `docs/annotation-grammar.md` (public reference, generated from a
  machine-readable EBNF source), or
- both — internal draft + generated public reference.

See `open-questions.md` for the running list of items to resolve.
