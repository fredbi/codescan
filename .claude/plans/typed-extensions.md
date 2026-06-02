# Typed extension values from the lexer

**Status:** Round-1 ✅ and Round-2 ✅ both landed. Schema builder now
consumes typed extensions via `Walker.Extension`. The architectural
carve-out (grammar2 may import `internal/parsers/yaml`) is recorded in
the yaml package's godoc.

Round-3 (per-entry `*yaml.Node` Position fidelity for LSP) remains
deferred — see "When to pick this up" below.

**Outstanding consumers** (still on grammar v1, will pick up typed
extensions automatically when they migrate to grammar2):

- `internal/builders/parameters/bridge.go:79-83` — YAML-fenced gap;
- `internal/builders/responses/bridge.go:83-84` — headers have no
  extension handling at all today.

---

## Context

Vendor extensions can appear in two shapes:

1. **Flat scalar form**

    ```
    extensions:
      x-tag: foo
      x-priority: 5
    ```

2. **Nested / typed YAML form**

    ```
    extensions:
      x-config:
        enabled: true
        threshold: 0.5
        tags: [a, b, c]
    ```

The lexer (grammar2) currently surfaces extensions as
`Extension{Name, Pos, Value string}` via `block.Extensions()` /
`Walker.Extension`. Its `collectExtensionsFromBody`
(`internal/parsers/grammar2/parser.go`) does a per-line
`strings.Cut(line, ":")` — no YAML, no nesting awareness.

That works for form (1). For form (2), the only consumer that handles
it correctly today is the **schema builder**, which bypasses
`Walker.Extension` entirely (`walker.go:189-195` documents the deliberate
non-wiring) and runs YAML→JSON on the whole `extensions:` raw block via
`applyExtensionsRawBlock`.

## Current state per builder

| Builder | Pipeline | Typed nested? | Notes |
|---|---|---|---|
| `internal/builders/schema` (grammar2) | `Walker.Raw → applyExtensionsRawBlock(body) → yaml.TypedExtensions` | ✓ | Single typed pipeline lives at the parser layer; the schema builder applies the allowed-extension filter and `AddExtension` |
| `internal/builders/parameters` (grammar v1) | `block.Extensions()` flat iterator → `param.AddExtension(name, stringValue)` | ✗ | `bridge.go:79-83` documents the YAML-fenced gap as known-missing |
| `internal/builders/responses` headers (grammar v1) | none | ✗ | `bridge.go:83-84` — "Extensions blocks are not currently supported on the header path" |
| `internal/builders/operations` (mixed) | bespoke per keyword | partial | Each block is its own special case |

## Round-1 (landed)

`internal/parsers/yaml/` gained a `TypedExtensions(body)` service that
encapsulates the YAML→JSON→typed-map pipeline. The schema builder now
calls it from `applyExtensionsRawBlock`; the per-key allowed-extension
filter and `AddExtension` stay at the call site.

This removes the direct `go.yaml.in/yaml/v3` + `swag/yamlutils`
imports from the schema builder and gives the same service to the
parameters and responses bridges for free once they migrate.

## Round-2 (landed)

Replaced the lexer's per-line `collectExtensionsFromBody` with a
YAML-aware body parse that emits typed values on `Extension.Value`.

### What changes in the lexer

1. **`Extension.Value` widens from `string` to `any`**
   (`internal/parsers/grammar2/ast.go:174-179`).
   `Walker.Extension`'s callback signature stays the same — `Value`
   becomes opaque to the callback's type system.

2. **`collectExtensionsFromBody` replaces its per-line `strings.Cut`
   with a single YAML pass** through `internal/parsers/yaml/`
   (importing the wrapper, not yaml.v3 directly — keeps the seam
   stable). Top-level mapping entries with `x-*` names emit Extension
   entries carrying their typed sub-value.

3. **Position fidelity per Extension entry**
   — the genuinely hard part. Today every Extension in a block shares
   the same `t.Pos` (the `extensions:` keyword's). LSP-grade per-entry
   positions ("`x-foo` at line 47 has malformed value") require
   decoding into `*yaml.Node`, walking the top level, and translating
   `node.Line` / `node.Column` (1-indexed relative to body) into
   absolute `token.Position`. Round-1 should keep the coarse
   block-level Pos; round-2 can refine.

4. **Allowed-extension filter** stays at the consumer side, not in
   the lexer. Reason: closed-vocabulary at lex time is opinionated;
   future call sites might want a different policy. Cost is one
   `classify.IsAllowedExtension` line per consumer — already the case.

5. **Error model** — a new `CodeInvalidYAMLExtensions` diagnostic
   replaces the schema builder's silent-drop behaviour. The lexer's
   `Warnf` infrastructure already handles this.

### What disappears downstream once this lands

- `internal/builders/schema/extensions.go` — `applyExtensionsRawBlock`
  deletes; the schema builder's `KwExtensions` arm in
  `schemaRawHandler` deletes; the `walker.go:189-195` "intentionally
  NOT wired" comment retires; `Walker.Extension` becomes the single
  uniform wiring.
- `internal/builders/parameters/bridge.go` — `block.Extensions()`
  loop unchanged in shape but `ext.Value` is now typed; the
  "YAML-fenced unsupported" caveat (`bridge.go:79-83`) deletes.
- `internal/builders/responses/bridge.go` — gains a
  `for ext := range block.Extensions()` loop in `applyBlockToHeader`;
  the "not currently supported" caveat (`bridge.go:83-84`) deletes.
- `internal/parsers/yaml/yaml.go` — `TypedExtensions` stays (the
  underlying YAML-parse engine the lexer now calls), but ceases to be
  consumed directly by the schema builder. The wrapper service
  becomes the lexer's implementation detail.

### Notes on the round-2 implementation

- The body parse via `yaml.TypedExtensions` now dedents the body
  first (strips the common leading-whitespace prefix shared by every
  non-blank line; converts residual leading tabs to two spaces). This
  was necessary because the lexer preserves godoc-level indentation
  per line, but YAML refuses tab indentation and treats leading
  whitespace as structural. Without dedent, the petstore meta block
  failed because every body line carried the source's tab prefix.
- The lexer's `collectRawBlock` uses `Token.Raw` (which preserves
  indentation) instead of the post-trim `Token.Text` for extensions
  bodies. Flat raw blocks (`consumes`, `produces`, `security`, …)
  still use `Text` + `formatKeywordLine` — their bodies don't need
  YAML-aware indentation.
- `Walker.Extension` is wired on both the schema-level walker and the
  `refOverrideCollector` (field-level `$ref` overrides). User-authored
  extensions from `Extensions:` blocks are NOT gated by
  `SkipExtensions` (parity with the v1 raw-block path); only
  scanner-derived `x-go-*` keys are gated.

### Round-3 (deferred) — per-entry positions for LSP

Today every Extension in a block shares the same `t.Pos` (the
`extensions:` keyword's). LSP-grade per-entry positions
("`x-foo` at line 47 has malformed value") require decoding into
`*yaml.Node`, walking the top level, and translating `node.Line` /
`node.Column` (1-indexed relative to body) into absolute
`token.Position`. Adds ~30 lines and a `*yaml.Node` walk to the lexer.

Pick this up when LSP per-entry extension diagnostics become a real
requirement — at that point a diagnostic-validating consumer drives
the granularity needed.

## References

- `internal/parsers/yaml/yaml.go` — `TypedExtensions` service + the
  `normaliseExtensionBody` dedent step (round-1 + round-2).
- `internal/parsers/grammar2/parser.go` — `collectExtensionsFromBody`
  calls the service (round-2).
- `internal/parsers/grammar2/lexer.go` — `collectRawBlock` keeps `Raw`
  for extensions bodies via the `yamlBody` flag (round-2).
- `internal/parsers/grammar2/ast.go` — `Extension.Value` widened to
  `any` (round-2).
- `internal/parsers/grammar2/diagnostic.go` —
  `CodeInvalidYAMLExtensions` (round-2).
- `internal/builders/schema/walker.go` — `schemaExtensionHandler`
  and `refOverrideCollector.onExtension` consume Walker.Extension
  (round-2).
- `internal/builders/parameters/bridge.go:79-83` — outstanding
  consumer; gap closes on grammar2 migration.
- `internal/builders/responses/bridge.go:83-84` — outstanding
  consumer; gap closes on grammar2 migration.
