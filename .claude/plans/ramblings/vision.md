# Vision

Date: 2026-04-19
Author: Fred

Vision of how codescan should evolve.

## Primer

We want to drastically transform the codescan package to make it:

* more powerful, yet easier to reason about
* can be extended to support OpenAPI v3 with new tags
* less error prone and better tested
* easier to maintain and debug
* faster and more memory-efficient
* easier to document
* infer more information from coe

After the restructuration of the repo into smaller packages, this should be possible to reach in progressive steps.

At some point, breaking changes will inevitably have to be introduced. This will be the start of a v2 release.

The objective of v2 are:

* to fully support OpenAPI v3.
* to provide LSP support
* to provide a standalone CLI to generate specs

## Step-by-step approach

We assume that our starting point is the refactored package.

1. Replace regexp by internal grammar-based parser
2. Replace global AST index by on-demand parser with cache
3. Introduce a flexible rendering layer

### Grammar-based parser

The objective is to entirely drop regexp usage from our parser.

* Performance is an important issue, but first and foremost, regexp are brittle and hard to maintain.
  We had one problem and we introduced 62 regexps. Now we have 62 problems.
* Another issue is the inability to convert the parser logic into relevant, up-to-date documentation.
  Eventually, all "parser-like" objects in the parsers package are unduly complex, and hard to maintain. 
* We want a solution that is less prone to break on user input in comments, and can be used to produce
  clean, accurate documentation.

See [the preliminary study](./grammar-vs-regexp-for-v2.md).
See [the analysis on grammar building](./hand-rolled-grammar-design.md)

Design goal:

* we design a grammar
* we implement a tokenizer (lexer) that analyzes comments
  * comments may be standard comments (for types, functions) or could be in the future inline comments (private).
* we implement a parser that understands the annotation structure
* we implement an emitter that implements the visitor pattern 

Design decisions:

* we won't use external libraries unless they genuinely earn their keep (e.g. super-fast tokenizer, etc)
  * we don't keep a strict zero-dependency policy, but libs must bring genuine value - otherwise we just copy the
    interesting bits
* we won't use PEG or "tag-driven grammar"
* we'll codegen our lexer and parser based on a grammar spec
* the grammar spec is used to generate markdown documentation
* we would use EBNF grammar - its expression power should be sufficent for us
  * challenge: context - however at the parsing level all expressions have the same form
  * all expressions are however not valid in all contexts (e.g. header, parameters, etc)
* interesting sources
  * golang.org/x/exp/ebnf could come in handy (and ebnflint too) - or worth taking inspiration from
  * this one may also be inspiring github.com/alecthomas/participle/v2/ebnf
  * also inspiring: parser combinator lib: https://github.com/tombenke/parc
  *  we may roll our own

Challenges:

* YAML blocks
* Produce better documentation than this: https://goswagger.io/go-swagger/generate-spec/

### On-demand AST parsing & cache

The current architecture scans the full code, maintains an index of all types, then walk these
in a stateful way to disover models.

This makes the "spec" builder particularly hard to follow, with its stafeful exploration loop.

See [the preliminary study](index-builder-statefulness.md).

Design:
* we want a lightweight scan (just syntax, not types) that first scans for annotations
* further usage by builder would ask for full parsing on demand
* we should not have to run this loop any longer and keep the state or maintain an index of ALL
  full resolved source code any more than necessary


### Flexible rendering layer

This lib internals are deeply entangled with go-openapi/spec types. This comes with the associated limitations
(OpenAPI 2.0 only, poor ordering of keys, perhaps other stuff).

Objectives: we want to decouple the parsing & analysis part from the rendering part.
Ideally we want to be able to generate _anything_ from our "OpenAPI-enabled" code scanner.

Design:
* we need to introduce a pivot model that captures all things detected by the emitter
  then eventually render it as go-openapi/spec object (or something else that marshals to JSON)
  then output as JSON or YAML.

See [the preliminary study](./builder-renderer-separation.md).

