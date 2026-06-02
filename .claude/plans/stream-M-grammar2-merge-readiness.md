# Stream M — grammar2 merge-readiness

**Branch:** `feat/new-parsing-layer`
**Status (2026-06-02, end-of-M6.7):**
M0 ✅ · M1 ✅ · M2 ✅ · M2.5 ✅ · M3 ✅ · M4 ✅ · M5 ✅ · M5.5 ✅ ·
M6.1 ✅ · M6.2 ✅ · M6.3 ✅ · M6.4 ✅ · M6.5 ✅ · M6.6 ✅ · M6.7 ✅ ·
**M7 queued** (history cleanup + merge).

The Stream-M code refactor + doc work is **complete**. Every builder
runs on the single grammar; common.Builder embeds across every
per-decl builder; routebody hosts the `+ name:` and `code:` sub-
languages; handlers/ owns the SimpleSchema + full-Schema dispatch
seam; meta extensions align with routes' diagnose+drop posture;
every M-stream-touched godoc carries current-state prose without
multi-version archaeology; eight package READMEs + four `./docs/`
files publish the contract end-to-end.

**Goal:** complete the grammar migration across every builder, then
merge to `master`. Code + doc done; merge prep is what's left.

This plan is the merge-readiness counterpart to the historical P-stream
(`grammar/60-implementation-roadmap.md`, `grammar/p7-schema-builder-redesign.md`).
P-stream took grammar2 + schema from greenfield to plateau; M-stream
takes the rest of the fleet across and clears the way to ship.

---

## 0. Recap — what is already in place

- **grammar2 lexer + parser** (`internal/parsers/grammar2/`) — final
  terminal vocabulary, Walker callbacks, typed
  `Extension.Value any` (post round-2), `OnDiagnostic` plumbing,
  `CodeInvalidYAMLExtensions` diagnostic. Grammar v1
  (`internal/parsers/grammar/`) is still resident because the
  non-schema builders import it.
- **Schema builder fully on grammar2** — P7/S1–S6 migration done; P7/S7
  polish (allOf $ref overrides, DescWithRef revival, vendor-extension
  lifting) done; post-summary closures done (A.2 SkipExtensions parity
  for `recognizeError`, A.3 empty schema for `json.RawMessage`, Gap C
  field-level `swagger:type`, Gap B' wrapper-decl `swagger:type`).
  Quirks split into ✅ resolved vs 🟡 deferred / 🟦 documented behaviour
  in `internal/builders/schema/README.md`.
- **YAML sub-language** (`internal/parsers/yaml/`) — `TypedExtensions`
  service + `normaliseExtensionBody` dedent. grammar2 carve-out
  granted: it may import this package.
- **Golden harness extended** — option-matrix goldens (SkipExtensions,
  DescWithRef), isolation fixtures
  (`raw-message-override`, `wrapper-decl-type-override`).

Outstanding bridges still on v1: `parameters`, `responses` (decl + headers),
`operations`, `routes`, `items`. All consume
`internal/parsers/grammar` + `internal/parsers/helpers`; `routes`
additionally consumes `internal/parsers/routebody`.

---

## 1. Settled design calls

### SimpleSchema vs Schema split — Mode B adopted

The schema builder gains a third Build mode parallel to `WithDefinitions`
/ `WithType`:

```go
sb.Build(schema.WithSimpleSchema(tpe, simpleTarget, in))
```

The `in` parameter carries the caller's resolved location
(`"query"` / `"path"` / `"header"` / `"formData"`, or empty for a
response header). The schema builder needs it for the SimpleSchema
inner rules — `file` type requires `in == "formData"`,
`allowEmptyValue` requires `in ∈ {"query", "formData"}`, and the
exit-validator's allowed-type set widens to include `file` only
under `in == "formData"`. Passing it through the option keeps the
contract explicit instead of routing it via a side channel.

Response headers pass `in = ""` (or a dedicated sentinel) so the
header-specific exclusions (`file`, `allowEmptyValue`) apply
automatically.

- Parameters and response-header callers own `in:` resolution end-to-end
  upstream of the schema call (already the case at
  `parameters/parameters.go:393` and `responses/responses.go:355` —
  they read `parsers.ParamLocation` and pick the typable). With Mode B
  they additionally select the build mode and surface `in` to the
  builder via the option, so the schema builder no longer infers
  simple-vs-full from the target's shape.
- SimpleSchema mode owns nested items drill-down internally. The
  `internal/builders/items` package becomes redundant by construction.

#### Allowed keywords under SimpleSchema (OAS v2)

Reference: <https://swagger.io/specification/v2/#parameter-object> and
<https://swagger.io/specification/v2/#header-object>.

A SimpleSchema is a parameter with `in: {query|path|header|formData}`
(anything except `in: body`) or a response header. The notion
disappears in OAS v3.

**Common allowed surface (params + headers):**

| Keyword | Notes |
|---|---|
| `type` | restricted to {string, number, integer, boolean, array, file} for params; headers exclude `file` |
| `format` | same vocabulary as full Schema |
| `items` | required when `type: array`; nested SimpleSchema (recursive) |
| `collectionFormat` | {csv, ssv, tsv, pipes, multi} default csv — SimpleSchema-only keyword |
| `default` | |
| `maximum`, `exclusiveMaximum`, `minimum`, `exclusiveMinimum`, `multipleOf` | |
| `maxLength`, `minLength`, `pattern` | |
| `maxItems`, `minItems`, `uniqueItems` | |
| `enum` | |
| vendor extensions | `x-*` allowed per the generic extensibility rule |

**Parameter-only extras:**

| Keyword | Notes |
|---|---|
| `allowEmptyValue` | params only **and** only when `in: {query, formData}` — forbidden on `in: path` and `in: header`; absent on response headers |
| `type: file` | only allowed when `in: formData` |

**Forbidden under SimpleSchema** (would emit `CodeUnsupportedInSimpleSchema`,
`SeverityWarning`, and be dropped from the target):

- `$ref`, `allOf`, `oneOf`, `anyOf`, `not`
- `properties`, `additionalProperties`, `maxProperties`, `minProperties`
- object-level `required` (array of required field names)
- `discriminator`, `readOnly`, `xml`, `externalDocs`
- a Go-level object / interface that resolves to a non-primitive
  schema (see §catch-at-exit below)

#### Type-resolution contract under SimpleSchema — "catch at exit"

The schema builder does **not** pre-filter based on the incoming Go
type. `*types.Struct` and `*types.Interface` are allowed to enter the
resolution pipeline because they can legitimately resolve to a
primitive — `time.Time` becomes `{string, date-time}`, a
`TextMarshaler` implementation becomes `{string, format}`,
`json.RawMessage` becomes `{}`, user overrides (`swagger:strfmt`,
`swagger:type`) win the cascade as they always do.

What the SimpleSchema mode does is **inspect the resolved target on
exit**:

- If the resolved target's `Type` is in the allowed SimpleSchema set
  (`{string, number, integer, boolean, array}`, plus `file` when the
  caller's `in == "formData"`) — accept.
- Empty `{}` (recognizer chose "any") — accept.
- Anything else (`object`, `$ref`, `allOf`, `properties` populated,
  …) — emit `CodeUnsupportedInSimpleSchema` `SeverityWarning` and
  **reset** the target.

The reset wipes the target back to empty `{}` rather than degrading
to `{type: string}` — empty is honest ("we couldn't represent this
in OAS v2 SimpleSchema") and avoids silently mistyping a complex
shape as a string.

Why Mode B (vs variadic flag, vs target-interface marker): mirrors the
existing two-mode dispatch (`Build` already panics on misuse), keeps
the `ifaces` surface unchanged (which `project_taggers_to_sunset` flags
as v2-sunset territory), and signals intent at call time.

### Routes / operations unification — out of M-scope

Routes and operations are functionally the same annotation with subtle
quirky differences. Making them perfect synonyms is a v2 objective
(per `project_v2_vision`), not part of M-stream. M5 ports routes onto
grammar2 in place; the unification ships after M7.

---

## 2. Phases

### M0 — Schema next-wave refactor ✅

**Landed in commits 8a526ba, 26cbdea, 98e4275, 34a888a, 0ef5698 and follow-ups.**
Schema package fully on the recipe (user-classifier first + applyStdlibSpecials uniformly + maintainer prose lifted to README); README disclaimer removed. All four next-wave files (`allof.go`, `struct.go`, `embedded.go`, `fields.go`) on the pattern; quirks split into ✅ resolved vs 🟡 / 🟦 open in `internal/builders/schema/README.md`. SPDX gaps closed. classifierAliasOwnDocStrfmt deleted (dead after the buildNamedAllOf unification).


**Scope.** Apply the `special_types.go` recipe to the remaining schema
files: `allof.go → struct.go → embedded.go → fields.go`. Retire
`scratch.go`.

**Recipe.**
- User-classifier-first precedence at every dispatch point.
- `applyStdlibSpecials` uniformly with `skipExt` plumbing.
- Maintainer prose lifted to `internal/builders/schema/README.md`
  §anchors with `# Details` godoc headings in code referring back.
- Goldens refreshed where behaviour changes; option-matrix harness
  extended where needed.

**Exit criteria.**
- README scope disclaimer at `schema/README.md` head removed (no
  longer says "applies to `schema.go` and `special_types.go` for now").
- No file in `internal/builders/schema/` references `scratch.go`.
- Full integration suite green; golden delta explained per fixture.

**References.** `MEMORY.md → project_schema_refactor_next_wave.md`.

### M1 — Parameters → grammar2 + SimpleSchema ✅

**Landed in commits 573e530, 09504ee, 8f35428.**
- 573e530 — parameters bridge rewritten on grammar2.Block + Walker; per-shape handler pattern; items-level dispatch inline via ItemsDepth (drops `items.ApplyBlock`); Walker.Extension wired; `schema.WithSimpleSchema(tpe, target, in)` Mode-B option added; exit validator + `SimpleSchemaProbe` interface + `CodeUnsupportedInSimpleSchema` diagnostic; isolation fixture + test.
- 09504ee — `isAliasParam` replaced with `s.simpleSchema` flag; closes the `in: header` omission; isolation fixture + test; README §quirks-resolved entry.
- 8f35428 — inline maintainer prose lifted to README §simple-schema-mode + §classifier-walkers; plan-doc references replaced with inline content; all P-stream / Q-numbered references stripped from source.


**Scope.**
- Rewrite `internal/builders/parameters/bridge.go` to consume `grammar2.Block`
  + `Walker` instead of `grammar.NewParser`. Pattern model is the schema
  builder's walker handlers + diagnostic plumbing.
- Add `schema.WithSimpleSchemaTarget(...)` Build mode (see §1).
- Switch the `paramTypable` write path for non-body parameters to drive
  the schema builder in SimpleSchema mode for any nested array/items
  shape. Drop the local `collectParamItemsLevels` recursion + the call
  to `items.ApplyBlock` — schema.Builder now owns items drill-down.
- Wire typed `Walker.Extension` (replaces the v1 flat-string iteration
  + `classify.IsAllowedExtension` filter at the bridge layer).
- Re-resolve `in:` against grammar2 (the current `parsers.ParamLocation`
  call uses v1 regex; grammar2's lexer already classifies `in:` as a
  TokenKeywordValue, so the parameter builder reads it off the block
  directly).

**Exit criteria.**
- `internal/builders/parameters/` imports `grammar2`, not `grammar`.
- `parameters/typable.go` still wraps `*oaispec.Parameter`; the per-field
  call into the schema builder uses `WithSimpleSchemaTarget` for
  non-body and `WithType` for body.
- Integration suite green; existing parameter goldens preserved.
- New fixture coverage for a SimpleSchema-illegal validation (e.g.
  `allOf` under `in: query`) producing the new diagnostic.

### M2 — Responses → grammar2 ✅

**Landed in commits 1ea5c7e, ff44b31.**
- 1ea5c7e — responses bridge rewritten on grammar2.Block + Walker (same per-shape handlers shape as parameters); `responseTypable` implements `SimpleSchemaProbe`; header path gets Walker.Extension wiring (closes v1 gap); `internal/parsers` import retired from the package; `KwExtensions` keyword context extended to `CtxParam` + `CtxHeader` so `Extensions:` blocks now parse in those contexts; header-extensions isolation fixture + test.
- ff44b31 — `bridge.go` / `bridge_test.go` renamed to `walker.go` / `walker_test.go` for parameters and responses, matching the schema package's file-naming convention. "bridge" phrasing inside renamed files updated to "dispatcher" / "builder".


**Scope.**
- Rewrite `internal/builders/responses/bridge.go` (decl + headers) to
  consume `grammar2.Block` + `Walker`.
- Headers reuse the SimpleSchema mode added in M1 for nested items.
- Headers gain `Walker.Extension` wiring — extension support today is
  zero on the v1 path (`bridge.go:83-84` documents the gap).

**Exit criteria.**
- `internal/builders/responses/` imports `grammar2`, not `grammar`.
- Existing response/header goldens preserved.
- New fixture covering header extensions (closes the v1 gap).

### M2.5 — `common.Builder` consolidation for parameters + responses ✅

**Landed in commits 9acaf9c, 7b9ff12, 9326c8b, 0ddb3fe, a010b89.**
- 9acaf9c — ParameterBuilder + ResponseBuilder embed `*common.Builder`; per-builder ctx/decl/postDecls fields dropped; `recordDiagnostic` forwarders deleted; field-comment parsing routes through the shared `ParseBlocks` cache; struct-literal tests rewritten to `NewBuilder(sctx, td)`; minor responses cleanups (panic("yay") remnant, IsAny/IsStdError TODO, `buildFromFieldInterface` signature tightened). Side-by-side review surfaced six quirks (Q1–Q6) logged in the commit message for follow-up.
- 7b9ff12 — new `internal/builders/handlers` package extracts the SimpleSchema Walker callbacks (`Number`, `Integer`, `UniqueBool`, `PatternString`, `CollectionFormatString`, `ComposeBool`, `ComposeString`, `Raw` with parameterised `errSink`). parameters/responses level-0 + items-level dispatchers consume from it. parameters/walker.go: 263 → 117 lines; responses/walker.go: 304 → 169 lines.
- 9326c8b — handlers.IsSimpleSchemaKeyword predicate + `simpleSchemaAllowed` map as the single source of truth for the OAS v2 SimpleSchema vocabulary. schema's Bool handler gates on `s.simpleSchema && !IsSimpleSchemaKeyword(name)` → emits `CodeUnsupportedInSimpleSchema` for `readOnly` and `discriminator`; silently skips `required:` (semantics belong to `paramRequiredBool`). Unit test pins the allowed set; integration test exercises the gate end-to-end via an anonymous-struct query parameter.
- 0ddb3fe — collapsed three `parsers/helpers` facades (`ParseValueFromSchema`, `ParseEnum`, `GetEnumDesc` / `EnumDescExtension`); hoisted `ExtEnumDesc` + `GetEnumDesc` to `internal/builders/resolvers/enum_desc.go` as cross-builder single source. schema's private `extEnumDesc`/`getEnumDesc` deleted; handlers/parameters/responses now reach `validations.CoerceValue` and `validations.ParseEnumValues` directly. Zero `parsers/helpers` imports remain in the consolidated builders.
- a010b89 — `paramExtensionHandler` + `headerExtensionHandler` + `schemaExtensionHandler` folded into `handlers.Extension(target ExtensionTarget)`. `ExtensionTarget` interface satisfied by every VendorExtensible-embedding type. `refOverrideCollector.onExtension` (schema/walker.go) intentionally left in place — it carries an extra `c.markExtension()` side effect that doesn't belong on the shared handler.

**Quirks log from M2.5.5 side-by-side review (queued for follow-up).**
- Q1 ✅ — `in:` default asymmetry. Fixed in **e9a9119**: explicit `inHeader` default on response side; `scanInLocation` returns invalid-candidate info; processResponseField emits CodeInvalidAnnotation warning for non-vocab `in:` values; fixture + golden capture four variants (empty / all-headers / mixed / invalid-`in:`).
- Q2 ✅ — `responseTypable.SetRef` blindly wrote `response.Schema.Ref` even when `in != body`, corrupting the body schema with a header-field reference. Fixed by adding a `refAttempted *bool` flag: SetRef no-ops + flips the flag under non-body mode; `HasRef` reads it; the SimpleSchema exit validator emits `CodeUnsupportedInSimpleSchema` + resets. Fixture `enhancements/response-header-ref-leak/` pins the two-variant emission (plain + with strfmt override).
- Q3 ✅ — `swagger:file` on a response field fired unconditionally, rewriting `resp.Schema` to `{file, ""}` even when the field was a header — silently corrupting the body schema from a misplaced annotation. Per OAS v2 the allowed header types are `{string, number, integer, boolean, array}` — `file` is forbidden on headers. Fixed by gating the file branch on `in == inBody`; misuse emits `CodeUnsupportedInSimpleSchema` + falls through to the normal field build. Fixture `enhancements/response-file-types/` pins three variants: legitimate file body, file on explicit-header, file on implicit-default-header.
- Q4 ✅ — `responses.buildFieldAlias` was missing the "non-body OR no-RefAliases → expand inline" gate that parameters has. Headers typed as an alias of a named primitive emitted as empty `{}` under default and RefAliases modes; alias chains on body fields polluted `definitions` with intermediate alias-name entries. Fixed by adding the parameter-side gate: `if typable.In() != inBody || !r.Ctx.RefAliases()` then expand via `types.Unalias(tpe)`. Header-aliased-basic now surfaces correctly as `{string, ""}`; body alias chains under default mode dissolve to the canonical decl (no chain pollution). Fixture `enhancements/alias-response-shapes/` pins the three-mode emission (default / RefAliases / Transparent). Top-level response-as-alias under default mode remains broken (anonymous-struct rejection) — documented in the fixture's package godoc, deferred to v2 with Q5.
- Q5 ✅ — stale `TODO(fred): this is wrong without checking for aliases?` removed. `buildFromType`'s dispatch already routes `*types.Alias` to `buildAlias`, so `buildNamedType` only ever sees `*types.Named`; `o.Type().Underlying()` returns the structural form without any aliased intermediate to skip past. Confirmed by tracing the alias paths via Q4's `alias-response-shapes` fixture.
- Q6 ✅ — responses' `processResponseField` now logs `"skipping field %s because it's not exported"` for unexported fields, matching parameters' `processParamField` shape exactly.


**Scope.** Embed `*common.Builder` in `ParameterBuilder` +
`ResponseBuilder`. This pulls forward the cross-builder factorisation
that was scheduled for M6, before M3 / M4 / M5 widen the diff
further. The motivation surfaced when looking at the level-of-detail
of M1+M2: each builder reimplemented a thin `recordDiagnostic` forwarder
and reparsed the same field comment twice. Unifying onto
`common.Builder` makes the shape of parameters + responses easier to
read side-by-side, which in turn makes residual quirks easier to spot
before routes / operations land their own migration noise.

`common.Builder` already ships everything required (see
`internal/builders/common/builder.go`):

- `Ctx *scanner.ScanCtx` and `Decl *scanner.EntityDecl` fields.
- `PostDeclarations()` + `AppendPostDecl(decl)` with per-Builder
  Ident dedup.
- `Diagnostics()` + `RecordDiagnostic(d)` with `OnDiagnostic`
  callback dispatch.
- `ParseBlocks(cg)` cache (memoises `ParseAll(cg)` per
  `*ast.CommentGroup`) + `ParseBlock(cg)` convenience for callers
  that want the single primary block.
- `Warn` / `Debug` slog helpers.

The schema builder already embeds it (`schema.Builder` has
`*common.Builder`). M2.5 replicates that pattern for parameters +
responses.

**Sub-steps.**
1. **Embed.** `ParameterBuilder` and `ResponseBuilder` gain a
   `*common.Builder` field. `NewBuilder` calls `common.New(ctx, decl)`.
2. **Drop duplicate state.** Per-builder `ctx`, `decl`, `postDecls`
   fields and `PostDeclarations()` methods delete. Callers reach
   them via the embedded base. `recordDiagnostic` methods delete —
   the embedded `RecordDiagnostic` takes over.
3. **Route through the parse cache.** `scanFieldDocSignals` takes a
   pre-parsed `[]grammar2.Block` instead of parsing internally;
   `applyBlockToField` / `applyBlockToHeader` read `s.ParseBlock(cg)`
   instead of calling `grammar2.NewParser(...).Parse(cg)` directly.
   Per-field parsing collapses to one cache miss + N cache hits.
4. **Catch quirks.** Reading parameters + responses side-by-side
   after the cleanup is the deliverable's other half — surface and
   log any inconsistency / quirk that surfaces (likely in
   `.claude/plans/stream-M-grammar2-merge-readiness.md` as a sub-list
   for follow-up before M3 / M4 / M5).

**Exit criteria.**
- `ParameterBuilder` and `ResponseBuilder` embed `*common.Builder`.
- Zero per-builder reimplementation of `recordDiagnostic` /
  `PostDeclarations` / context-and-decl fields.
- `scanFieldDocSignals` and the walker dispatchers share the parse
  cache (no double-parse of `afld.Doc`).
- Routes / operations remain on their current shape — they will
  embed `common.Builder` at the same time as their grammar2
  migration (M4 / M5).
- Any quirks surfaced during the side-by-side comparison are
  captured (resolved in a follow-up commit, or logged for M-stream
  scope).

### M3 — Items disposition ✅

**Outcome.** Dissolved `internal/builders/items/` into
`resolvers/items_adapters.go`. The package's premise — "SimpleSchema
in schema.Builder will absorb the items chain, leaving no consumer" —
proved optimistic: `*oaispec.Items` is a distinct shape from
`*oaispec.Schema` so parameters' and responses' items recursion can
not share the schema builder's `SchemaOrArray` walk. Two small
ifaces-adapters survived (`Typable` for the recursive `Items()`
return value, `Validations` for the items-level Walker target).

Rather than keep a 2-file dedicated package, the adapters relocated
to `resolvers/` (the existing home of small ifaces-glue) as
`ItemsTypable` / `ItemsValidations`. The legacy `bridge.go` /
`bridge_test.go` (grammar v1 dispatcher) deleted outright — schema /
parameters / responses all dispatch grammar2 directly via
`*Walker.FilterDepth` since M1/M2.

**Landed.**
- `internal/builders/resolvers/items_adapters.go` — `ItemsTypable`,
  `ItemsValidations`, `NewItemsTypable`, `NewItemsValidations`.
- 4 call sites updated: `parameters/typable.go`, `parameters/walker.go`,
  `responses/typable.go`, `responses/walker.go`.
- `internal/builders/items/` directory removed.
- Stale `schema/walker.go` comment referencing
  `items.ApplyBlock(grammar.v1)` cleaned up.

**Exit criteria.**
- ✅ `internal/builders/items/` removed.
- ✅ No production code or test file imports it.
- ✅ `go build ./...` + `go test ./...` green.

**Follow-up.** Update the top-level CLAUDE.md package-layout table to
drop the `items` row — done in the same commit as the M3 changes.

### M4 — Operations → grammar2 ✅

**Outcome.** The operations bridge collapsed from 87 lines to 22.
Grammar2's lexer already classifies prose into TokenTitle /
TokenDesc and isolates `---` fenced bodies into TokenOpaqueYaml,
so `block.Title()` / `block.Description()` map straight to
`op.Summary` / `op.Description` and the first body from
`block.YAMLBlocks()` feeds the unmarshal pipeline. No more
`CollectScannerTitleDescription` / `JoinDropLast` dance — the
lexer did it all at lex time.

The local `unmarshalOpYAML` (dedent → yaml.Unmarshal map[any]any
→ YAMLToJSON → MarshalJSON → unmarshal callback) moved to
`internal/parsers/yaml/operation.go` as `UnmarshalOperationBody`.
`helpers.RemoveIndent` moved alongside it as `yaml.RemoveIndent`
(the dedent algorithm is yaml-specific). The redundant
`MarshalJSON` step disappeared — `yamlutils.YAMLToJSON` already
returns `json.RawMessage` = `[]byte`, so the unmarshal callback
receives it directly.

**Landed.**
- `internal/parsers/yaml/dedent.go` — `RemoveIndent` (moved from
  `helpers/indent.go`).
- `internal/parsers/yaml/operation.go` — `UnmarshalOperationBody`.
- `internal/parsers/yaml/operation_test.go` — round-trip, invalid
  YAML, empty body, tab-indent scenarios.
- `internal/builders/operations/bridge.go` — rewritten on
  grammar2 + `yaml.UnmarshalOperationBody`.
- Deleted: `internal/parsers/helpers/indent.go`,
  `internal/builders/operations/bridge_test.go` (tests live in
  the yaml package now).

**Exit criteria.**
- ✅ `internal/builders/operations/` imports `grammar2`, not `grammar`.
- ✅ No direct import of `go.yaml.in/yaml/v3` or `fmts.YAMLToJSON`
  in `operations/`.
- ✅ All operation + integration goldens preserved (no `UPDATE_GOLDEN`
  needed).

### M5 — Routes → grammar2 ✅

**Outcome.** Routes joined the grammar2 + common.Builder cluster:
- `bridge.go → walker.go` rewritten on grammar2 (88-line dispatcher,
  matches operations' shape).
- `routes.Builder` embeds `*common.Builder` with nil Decl (same
  pattern as M4 follow-up for operations).
- `internal/parsers/routebody/` package deleted — the body parsers
  relocated wholesale into `routes/` as `body_params.go`,
  `body_responses.go`, `body_extensions.go`. The dead matcher
  surface (`Matches()` methods + `rxParameters` / `rxResponses` /
  `rxExtensions`) dropped since grammar2 routes by keyword name.
  Body parsers' internal regex logic survives unchanged.
- One quirk surfaced and resolved during the migration: route docs
  are usually `/* ... */` block comments, which retain godoc-style
  per-line tab/whitespace indent that grammar2's lexer preserves
  verbatim. `walker.go:trimCommentPrefix` strips the same set v1's
  `rxUncommentHeaders` did, applied per-line on `block.Title()`
  and `block.Description()`. Pinned in `routes/README.md`
  §quirks-block-comment-prefix.

**Landed.**
- `internal/builders/routes/walker.go` — new grammar2 dispatcher.
- `internal/builders/routes/walker_test.go` — grammar2 ParseAs +
  relaxed body-absorption assertions.
- `internal/builders/routes/routes.go` — common.Builder embed.
- `internal/builders/routes/body_{params,responses,extensions}.go`
  — relocated from `internal/parsers/routebody/`, ErrParser → ErrRoutes,
  dead Matches() + rx regex globals removed.
- `internal/builders/routes/README.md` — package overview,
  per-keyword dispatch table, body-parser disposition, quirks log
  (resolved + open).
- Deleted: `internal/parsers/routebody/` directory.

**Exit criteria.**
- ✅ `internal/builders/routes/` imports `grammar2`, not `grammar`
  (still imports `helpers` for `SchemesList` / `YAMLListBody` /
  `SecurityRequirements` — M6 cleanup target).
- ✅ `internal/parsers/routebody/` removed.
- ✅ Quirks log updated; new routes README §quirks section in
  place.
- ✅ All route + integration goldens preserved (no
  `UPDATE_GOLDEN=1` needed).

### M5.5 — `builders/spec` (meta) → grammar2 ✅

**Outcome.** Meta bridge rewritten on grammar2 in one pass —
fixtures use `//` line comments throughout (no block-comment quirk
to absorb), and grammar2's prose classifier handles the trailing
`swagger:meta` annotation position natively. Every meta keyword
had a grammar2 constant ready (`KwTOS`, `KwConsumes`, `KwProduces`,
`KwSchemes`, `KwSecurity`, `KwVersion`, `KwHost`, `KwBasePath`,
`KwLicense`, `KwContact`, `KwSecurityDefinitions`,
`KwInfoExtensions`, `KwExtensions`).

The local YAML pipeline (`unmarshalYAMLBody`, `removeYAMLIndent`,
`leadingWhitespaceLen`) collapsed into one call to
`yaml.UnmarshalBody` (the M4 helper, renamed from
`UnmarshalOperationBody` to reflect its broader applicability).
Both `extensions:` and `infoExtensions:` raw blocks now route
through `yaml.TypedExtensions` directly (JSON-typed values out of
the box). The mandatory `x-*` validation lives one level up so a
typo still surfaces as an error.

Bonus dependency drop: `github.com/go-openapi/loads` had only one
caller left in the tree — `fmts.YAMLToJSON` in the old
`unmarshalYAMLBody`. With the pipeline gone, `go mod tidy` removed
loads + 2 sum entries.

`spec.Builder` deliberately does NOT embed `*common.Builder` —
it's the orchestrator over every other builder, not a per-comment
target. Embedding would have added empty `Decl` / unused
`ParseBlocks` cache surface to a type that doesn't parse anything
itself (it only delegates to child builders).

**Landed.**
- `internal/parsers/yaml/operation.go` — renamed
  `UnmarshalOperationBody` → `UnmarshalBody`, godoc updated to
  reflect dual operations/meta use.
- `internal/builders/operations/walker.go` +
  `internal/parsers/yaml/operation_test.go` — call sites updated.
- `internal/builders/spec/meta_bridge.go` — rewritten on grammar2;
  ~50 lines shorter (local YAML pipeline + extension validation
  helpers gone; bodyLines helper added).
- `internal/builders/spec/spec.go` — `buildMeta` switched to
  `grammar2.NewParser`.
- `go.mod` / `go.sum` — `go-openapi/loads` removed.

**Exit criteria.**
- ✅ `internal/builders/spec/` imports `grammar2`, not `grammar`.
- ✅ No direct import of `go.yaml.in/yaml/v3` or `fmts.YAMLToJSON`
  in `builders/spec/`.
- ✅ All meta + petstore + integration goldens preserved (no
  `UPDATE_GOLDEN=1` needed).
- ✅ Bonus: `go-openapi/loads` no longer a transitive dependency.

### M6 — Cleanups (M6.1 → M6.7) ✅

M6 was originally a single step but the scope outgrew that framing.
Splitting into focused sub-steps lets each land independently with
a small reviewable diff and lets us reassess after each one. The
order is linear (each step assumes the previous closed); M6.4 in
particular is risky enough that it benefits from a stable helpers
landscape underneath.

#### M6.1 — Delete grammar v1 ✅

**Outcome.** Grammar v1 vanishes outright. After M5.5 the package
had zero production importers; the only remaining consumers were
internal to the package (parser self-tests, the parity harness
under `grammar_test/`, the doc generator under `gen/`). None
encode coverage grammar2 doesn't already carry — the parity
harness was specifically built to compare v1 ⟷ v2 output during
migration, and the 7 golden snapshots under
`grammar_test/testdata/golden/` capture "what v1 produced" which
is the wrong target now.

The doc generator (`gen/main.go`) generated
`docs/annotation-keywords.md` from `keywords_table.go`. Both die
in this commit. The doc gets a stale-notice header pointing at
M6.7 — its body remains as a reference until M6.7 ports the
generator to grammar2's `keywords.go` and rewrites the prose.

**Landed.**
- Deleted: `internal/parsers/grammar/` (entire directory — parser,
  lexer, keywords, preprocess, AST, diagnostics, style, gen,
  grammar_test, all `*_test.go`, all `testdata/golden/`).
- Updated: `docs/annotation-keywords.md` — stale header replaces
  the AUTO-GENERATED banner; prose intro tweaked to drop the
  obsolete "v2 grammar parser" framing.

**Exit criteria.**
- ✅ `internal/parsers/grammar/` gone.
- ✅ `go build ./...` + `go test ./...` green (17 packages tested,
  down from 20; the missing 3 are the deleted dirs).
- ✅ Grammar2 coverage unchanged — nothing needed porting.

#### M6.2 — Helpers consolidation ✅

**Outcome.** `internal/parsers/helpers/` dissolved outright. The
audit found one helper per remaining use case — no patterns worth
preserving as a package.

Per-helper landing:
- **Dead, deleted** — `Setter`, `CleanupScannerLines`,
  `CollectScannerTitleDescription` (+ its three regexes
  `rxUncommentHeaders` / `rxPunctuationEnd` / `rxTitleStart`).
  Every consumer now reads `block.Title()` / `block.Description()`
  from grammar; the v1 heuristics (blank-line split, ATX heading
  promotion, punctuation-ends → title) live in grammar's lexer as
  the `classifyProseRun` heuristics.
- **Relocated to `internal/scanner/enum_value.go`** —
  `GetEnumBasicLitValue` became unexported `enumBasicLitValue`. AST
  → runtime-value coercion is scanner concern, not parsers
  concern.
- **Relocated to `internal/builders/common/routemeta.go`** —
  `SchemesList`, `SecurityRequirements`. Two callers each (routes
  + spec); their shared home matches the routes-↔-meta commonality
  layer the plan called out. `SchemesList`'s placement is
  pending the M6.4 challenge — likely replaceable by a grammar
  `SplitCommaList` export.
- **Inlined** — `DropEmpty` + `JoinDropLast`. Had one combined
  caller in `spec/walker.go` for the Terms-Of-Service body.
  Equivalent to "join non-blank lines with `\n`" — landed as a
  small local `joinNonBlank` helper in `spec/walker.go`.

**Landed (commit f907790).**
- Deleted: `internal/parsers/helpers/` (6 files, 360 lines).
- New: `internal/scanner/enum_value.go`,
  `internal/builders/common/routemeta.go`,
  `spec/walker.go:joinNonBlank`.
- CLAUDE.md package-layout refreshed.

**Exit criteria.**
- ✅ `internal/parsers/helpers/` deleted.
- ✅ Every former helper has a single canonical home that matches
  its layer.
- 🟡 Routes ↔ meta common layer landed in `common/routemeta.go`
  with the understanding that M6.4 will likely shrink it further
  by exporting a grammar comma-list splitter.

#### M6.3 — Regex classifier ports ✅

**Outcome.** Scope shifted from the original plan as the audit
landed.

The originally-listed targets (`ParamLocation`, `StrfmtName`,
`FileParam`) were already deleted in M6.2.0 (`be1cf07`) as part of
the matcher-cleanup — they had zero production callers post the
grammar migration. The scanner classification surface
(`ExtractAnnotation`, `ModelOverride`, `ResponseOverride`,
`ParametersOverride`, `ParsedPathContent`) was kept on regex
intentionally: it's a per-comment-group / per-line classification
hot path; restructuring the scanner to parse with grammar at that
layer is deferred to a separate "ast bits of interest" rework
outside this branch.

Three small regex matchers in the build path (not the scanner
classification path) were ported to byte-loop helpers:

- `parsers/yaml/list.go:rxLineLeader` (`^[\p{Zs}\t/\*]*\|?`) →
  `stripListLeader` ASCII byte loop.
- `spec/walker.go:rxStripTitleComments`
  (`^[^\p{L}]*[Pp]ackage\p{Zs}+[^\p{Zs}]+\p{Zs}*`) →
  `stripPackagePrefix` (skip non-letters, match Package/package,
  eat whitespace, eat identifier, eat trailing whitespace).
- `spec/walker.go:httpFTPScheme` (`(?:(?:ht|f)tp|ws)s?://`) → a
  six-entry `urlSchemes` lookup table with leftmost-match scan.

**Landed (commit 2072770).** Two files, +97 / −29.

**Remaining `regexp` imports under `internal/` — kept by design:**
- `parsers/{regexprs,matchers,parsed_path_content}.go` —
  scanner-layer classification, deferred per above.
- `scanner/index.go` — user-facing package include/exclude
  patterns take regex by API contract.
- `schema/walker.go:432` — RE2 compile-check on
  user-supplied `pattern:` values; correctness requires the
  actual engine.

**Exit criteria.**
- ✅ Builder-path regex eliminated outside the three legitimate
  use cases above.
- 🟡 Original "no `regexp` in `internal/parsers/`" exit criterion
  retired — the scanner classification rework that would close it
  is a separate workstream.

#### M6.4 — Cross-builder commonality sweep ✅

(small-three landed in-stream; the route-bodies item flagged as
`M6.4-W` was completed under M6.5-C when `routebody` replaced
`builders/routes/body_params.go` + `body_responses.go` entirely —
both sub-grammars now live in `internal/parsers/routebody/` and
dispatch through the handlers seam.)

**Outcome.** The plan's original candidate list was audited live;
four small extractions landed in this branch (C1, C3, C4, plus the
parameters PostDecl bug fix). The route/parameters and
route/responses body-grammar work surfaced as a much bigger lift —
elevating typed productions into grammar — and is queued for a
dedicated tomorrow session (M6.4-W, not a workshop sub-stream;
trial/error/revert permitted).

**Landed commits.**
- `c7db9e9` — C1: `buildOption` → `schema.OptionFor`. Verbatim
  helper hoisted to the schema package; parameters + responses
  drop their local copies.
- `5dca35b` — C3: `grammar.SplitCommaList` exported (was
  unexported `splitCommaList`); `common.SchemesList` retired.
- `454ab5d` + `db0b8db` — C4a/b: a fixture that witnesses a real
  parameters.buildFromFieldMap PostDecl bug (missing model
  registration when only reachable via a map field), then the
  one-loop fix. The golden delta on the fix commit shows
  LocalItem materialising in definitions; this pair locks the
  regression risk into git history.
- `df9f646` — E1: typed Security accessor on Block. New
  `internal/parsers/security/` sibling sub-parser
  (yaml-precedent), grammar dispatches at lex time, builders read
  off `block.SecurityRequirements()`. `common/routemeta.go`
  retires entirely.
- `8411413` — E2: typed Contact and License accessors on Block.
  Parse logic moved into `grammar/meta_info.go`. spec/walker.go
  sheds parseContactInfo, parseLicense, splitURL, urlSchemes, and
  the `net/mail` import.

**Surveyed and decided against extraction:**
- `buildFromFieldStruct` / `buildFromFieldInterface` — verbatim
  across params + responses, hoist tempting. Kept local: avoids
  growing `common.Builder` into a mini-DSL. The divergences across
  buildFromField family (params' IsAny early return,
  buildFieldAlias's body-vs-non-body logic) are useful
  quirk-hunting signal anyway.
- `buildFromField` switch dispatcher — generics-based hoist
  feasible but net cost ≈ net win; left alone.
- `processParamField` / `processResponseField` outer loops — the
  shared pre-checks are 5 lines of filter; helper would duplicate
  what's already a tight filter.
- `makeRef` — already hoisted in M2.5 (`common.Builder.MakeRef`).
  Confirmed no local copies remain.
- Items-chain adapters — already shared via
  `resolvers.ItemsTypable` / `ItemsValidations` from M3.

**Architectural insight surfaced (queued):**
The route body parsers (`body_params.go`, `body_responses.go`)
carry ~500 lines of regex-era parsing for the `+ name:`
parameter sub-grammar and the `200: body Foo description` response
mini-tag-language. Both should ultimately move into grammar as
typed productions — sibling sub-parsers under
`internal/parsers/`, dispatched from grammar's emitRawBlock at lex
time (the yaml/security/extensions precedent). The shape:

- `parsers/routebody/Parameters([]string) []ParamDecl`
- `parsers/routebody/Responses([]string) []ResponseRef`
- new `block.RouteParameters()` / `block.RouteResponses()`
  accessors

The `+ name:` multi-line sub-grammar is the hard part — the
existing lexer is line-oriented; encoding "+ at column 0 starts a
new param, indented lines belong to it" needs design before code.

**Deferred to M6.4-W (workshop session, this branch):** the two
route-body accessors. Trial/error/revert encouraged.

**Exit criteria.**
- ✅ Each candidate either landed under a shared owner or
  explicitly logged as "kept duplicate; reason X."
- ✅ No new shared helpers without a one-line godoc on the host
  package explaining the call-site invariant.
- 🟡 Route-body typed accessors deferred to M6.4-W.

#### M6.5 — Code refactor wave (PRE → A → B → C → D → E + Q21 audit) ✅

**Scope evolved.** The originally-planned M6.5 ("README + quirks
refresh") expanded into the full code-refactor wave when handler
hoisting + routebody migration surfaced as the highest-value
pre-merge structural moves. The README + quirks refresh work merged
into M6.6 (READMEs as the comment-sweep deliverable) and the
post-M6.5-E quirks audit (which became `observed-quirks.md` Q1–Q25
status refresh + the Q21 fix commit).

**What landed in M6.5:**

- **M6.5-PRE** (`d6e9d42`) — swagger:route fixture suite (28
  fixtures + goldens). The witness safety net every subsequent
  M6.5 commit's diff is judged against.
- **M6.5-A** (`4b6a92e`) — SimpleSchema dispatcher hoist into
  `handlers/dispatch_simple.go`. parameters / responses become
  AST-side orchestrators.
- **M6.5-B** (`ca99185`) — Schema dispatcher hoist into
  `handlers/dispatch_schema.go`. `SchemaOptions` + `SchemaValidations`
  + checkShape + applyPattern + setRequired + setDiscriminator +
  schemaTypeOf all migrate; refOverrideCollector stays in schema.
- **M6.5-C** (`9f2665c`) — routebody sub-parser + handlers dispatch.
  Retires `body_params.go` + `body_responses.go` (~479 LOC); 8
  quirks resolved (Q14, Q17-Q20, Q22 + Q15/Q16 via the Schema
  hoist's checkShape). Bonus: `Q23` cleanup landed in the same
  area (`76f481b` — retire `trimCommentPrefix`).
- **M6.5-D** (`5449e20`) — unified list parsing via `Property.AsList`.
  `KwSchemes` widens to `asRawBlock`; lexer's `collectRawBlock` gains
  inline-value capture. Retires `yaml.ListBody`, `SplitCommaList`,
  `routes.bodyLines`.
- **M6.5-E** (`06db4a5`) — meta extensions align with routes via
  `block.Extensions()`. `Extension.Source` field added.
  `validateExtensionNames` + `ErrBadExtensionName` retired.
  Bonus: `stripPackagePrefix` simplified.
- **Q21 audit fix** (`cdc4be7`) — caught during the post-M6.5-E
  observed-quirks audit. The original M6.5-C commit claimed Q21
  (leading-space artifact on response descriptions) was resolved
  but the routebody description accumulator still produced
  `" not found"` etc. — the witness goldens locked the still-buggy
  state. Fixed the routebody join logic, regenerated 27 affected
  goldens, updated Q21 entry in `observed-quirks.md`.

**Exit criteria** — all met:
- Every builder runs on the shared handlers seam.
- Routebody owns the `+ name:` and `code:` sub-languages cleanly.
- 11 Q-entries resolved (Q14–Q22 plus the three documented
  behaviour shifts Q23–Q25).
- All goldens green; all lint clean; observed-quirks ledger
  refreshed.
- PR-description summary still pending — folded into M7's merge
  prep (where the squash story is documented and the PR body
  drafted).

#### M6.6 — Comment summarization sweep ✅

**Shipped end-of-stream.** Per-package commit per the recipe (four
rules: strip historical narration; one grammar only; lift plan refs
to README; M0 `# Details` pattern). The "only one grammar in
published docs" rule extended the four-rule recipe — every
`grammar2` / `v1` / `pre-Stream-M` / `regex-era` / `P7 plan` /
`round-N` reference scrubbed from production godoc.

**Commits (in order shipped):**

- `5fcc19c docs(common)` — warm-up: README + builder.go cleanup.
- `bea341f docs(scanner)` — README + Options strip.
- `2527936 docs(handlers)` — README + v1 archaeology strip.
- `2816a00 docs(operations)` — README + walker.go strip.
- `1a69d5f docs(resolvers)` — stale cross-file pointer fix
  (`schema/extensions.go:clearStaleEnumDesc` → moved to
  `handlers/dispatch_schema.go` in M6.5-B).
- `f33a939 docs(validations)` — README (7 anchored sections) lifting
  P7 plan references.
- `357e1aa docs(routes)` — Q-number archaeology + "legacy parser"
  refs scrubbed; routes/README.md refactored for current shape.
- `cc79671 docs(schema)` — post-M0 stragglers (5 files, ~9 marker
  hits each).
- `13997fb docs(yaml)` — README + typed-extensions plan refs lifted.
- `a263fce docs(routebody)` — README (full sub-language grammar
  spec lifted from doc.go) + M6.5-C archaeology stripped.
- `9983dba docs(grammar)` — README (~1000 LOC, 21 anchored sections).
  The heavy one: 27 plan refs + 27 marker hits resolved.

**Deliverables:**
- 8 new package READMEs (~2100 LOC of lifted maintainer prose).
- Existing `builders/schema/README.md` and `builders/routes/README.md`
  refreshed in place.
- Zero `.claude/plans/...` pointers remain in production code.
- Zero "v1" / "grammar2" / "Q14-Q25" / "M6.5-C" / "P7 plan" /
  "round-N" references remain in shipped godoc.
- The grammar `# Details` pattern uniformly applied across every
  M-stream-touched file.

**Follow-up filed:** godoc link-compliance pass — the
`[§anchor](./README.md#anchor)` shape used uniformly across the
sweep doesn't render as clickable on pkg.go.dev. The proper godoc
reference-link format with absolute GitHub URLs (template established
in `builders/common/builder.go`) is filed in §3 backlog as a
mechanical post-merge sweep.

**Premise.** Multiple waves of refactoring left inline godoc
carrying *change narration* (what v1 did, what pre-Stream-M did,
what Qn fixed, what we used to do before grammar2). At merge time
the reader doesn't know — or care — what the code used to look
like; they need to know **what it does now and why**. The M0
schema lift pattern already proved out across schema /
parameters / responses / routes READMEs: lift dense prose into a
README §section, replace the inline comment with a one-line
intent + `# Details\n//\n// See [§anchor](./README.md#anchor)`
pointer.

**Scope.** Walk every package that hasn't been through the M0
lift yet — at minimum `operations`, `spec`, `validations`,
`handlers`, `resolvers`, `common`, `scanner`, the grammar2 internals
that surface to builders, and `parsers/yaml`. For each file:

1. **Strip historical narration.** Any sentence that talks about
   what changed, what v1 did, what was previously the case,
   what Qn / Mn fixed, "pre-Q3 the file branch fired", "v1 quirk
   preserved", "matches v1 behaviour", "carry-over from a more
   permissive era", "post-grammar-migration" — gone. Replace with
   a current-state statement: what this code does and why it
   does it that way *today*.

   Tone check: a reader landing on the file via `go doc` should
   learn the contract, not the archaeology. If a historical
   detail is genuinely load-bearing (a semantic that survives
   only because changing it would break a public contract), keep
   it — but tag it as a present-tense invariant, not a past-tense
   story.

2. **Expand internal-plan references.** Every `.claude/plans/...`
   reference, every `see plan §X.Y`, every `per the P7 plan`
   needs to leave the code. `.claude/plans/` is gitignored, so a
   reader downstream of the merge will never find what the
   comment points at. Two outcomes per reference:

   - **Lift the rationale into the package README** (markdown,
     under a new §section if needed) and replace the inline
     pointer with a `See [§anchor](./README.md#anchor)` line.
   - **Inline the explanation** when it's short enough to live
     in three lines of godoc without overwhelming the function
     comment.

   Examples of what to lift:
   - `validations/coerce.go`: `per the P7 plan §5.5` →
     materialise the §5.5 reasoning into
     `validations/README.md` (new).
   - `parsers/yaml/yaml.go`: `see .claude/plans/typed-extensions.md`
     → fold the round-1 / round-2 plumbing rationale into
     `parsers/yaml/README.md` (new) or into the package godoc
     directly.
   - `parsers/grammar2/doc.go`: the
     `.claude/plans/grammar/00-overview.md` / `40-lexer.md` /
     `50-full.md` references are the most extensive — lift the
     dispatch-table sketch + lexer contract into the grammar2
     package godoc itself.

3. **Apply the M0 `# Details` pattern** to every dense block left
   standing. The shape:

   ```go
   // FuncName does the one-sentence thing it does.
   //
   // # Details
   //
   // See [§anchor](./README.md#anchor) — keyword landing rules,
   // edge cases, the rationale for choosing this shape over the
   // obvious alternative.
   func FuncName(...) {
   ```

   The inline godoc stays short and intent-focused; the README
   carries the dense prose with examples, tables, and
   cross-links.

4. **No new emojis, no new TODO(claude) markers.** If a sweep
   surfaces real follow-up work that can't land in M6.6, file it
   as a fixture-backed `quirks-open` entry in the affected README
   — not as a TODO in the code.

**Per-package pass list (initial scope; M6.5 README refresh may
expand it).**

| Package | Pass needed | Notes |
|---------|-------------|-------|
| `builders/operations` | yes | M4-era walker has lingering "compared to v1" narration. |
| `builders/spec` | yes | Post-M5.5; meta_bridge comments still describe SectionedParser. |
| `builders/validations` | yes | Carries multiple `per the P7 plan` references. |
| `builders/handlers` | yes | Small package; quick pass. |
| `builders/resolvers` | yes | Items-adapter file (M3) has migration narration. |
| `builders/common` | yes | TODO(fred) on `WithDiagnosticSink`; verify. |
| `builders/routes` | partial | M5 already README-anchored most blocks; do the rest. |
| `parsers/yaml` | yes | Typed-extensions plan refs to lift. |
| `parsers/grammar2` | yes | Largest plan-ref surface (`grammar/00..70`). |
| `scanner` | review | Less change-narration but plan refs may exist. |
| `internal/integration` | no | Tests; out of scope. |

**Exit criteria.**
- `grep -rn '\.claude/plans/' --include='*.go' internal/`
  returns no matches.
- `grep -rnE '\b(pre-Q[0-9]|pre-M[0-9]|pre-Stream|v1 quirk|v1
  parity|legacy SectionedParser|matches v1|formerly|previously)\b'
  --include='*.go' internal/` returns no matches outside test
  files where the assertion is testing v1-compat shape.
- Every dense godoc block (>15 lines) either justified
  in-place or lifted to a README §anchor.
- Every package with a README has its §-anchors used from at
  least one godoc; orphan README sections flagged.
- New README files added where the lift needed a home
  (`validations/README.md`, `parsers/yaml/README.md`,
  `parsers/grammar2/README.md`, `builders/common/README.md`,
  `builders/handlers/README.md`, `builders/resolvers/README.md`
  most likely).

#### M6.7 — `docs/` rewrite ✅

**Shipped as `aae1738`.** Replaces the stale auto-generated
`docs/annotation-keywords.md` (330 LOC, carried a STALE banner
pointing at a long-deleted generator) with four hand-authored
topic-split docs under `./docs/`. Minimal Hugo frontmatter
(title + weight) for future Hugo migration; author-leaning content
shows first (lower weights).

**Files:**

| File | Weight | LOC | Audience entry |
|------|--------|-----|----------------|
| `annotations.md` | 10 | 876 | Annotation authors — every `swagger:*` with examples + an annotation × keyword matrix |
| `keywords.md` | 20 | 645 | Keyword reference card grouped by family |
| `sub-languages.md` | 30 | 491 | Prose classification, flex-list, Parameters / Responses grammars, YAML extensions, security requirements, contact/license |
| `grammar.md` | 40 | 559 | Formal EBNF lifted from `.claude/plans/grammar/50-full.md`, actualised against current code |

**v0 scope** (intentional per the M6.7 brief):
- Comprehensive but not exhaustive — some users will benefit from
  expanded keyword samples and more worked end-to-end examples in a
  future pass.
- Hybrid sample format: inline snippet (so each doc reads
  end-to-end) plus a one-line fixture-path reference for the full
  executable example.
- Cross-links use relative paths (`./grammar.md#anchor`); future
  Hugo migration handles URL rewriting.
- No doc generator at this stage — punt on auto-generation; a
  testify-codegen-style approach is filed in §3 backlog as a
  long-term option.

**Stale generator removal:** `docs/annotation-keywords.md` deleted.
No code/Makefile/generate-directive cleanup needed — the actual
`internal/parsers/grammar/gen/` directory was already removed when
grammar v1 was deleted in `433247a`.

**Premise.** The repo's user-facing documentation lives in
`docs/`. Today the only file there — `annotation-keywords.md` — is
auto-generated by `internal/parsers/grammar/gen` from grammar v1's
keyword table. Once M6.1 deletes grammar v1, both the generator
and its source vanish. The doc itself opens with "the 34
`keyword: value` forms recognized by the v2 grammar parser" —
ambiguous wording that conflates **OAS v2** (the spec being
emitted) with **grammar v2** (the obsolete in-tree parser
generation). The non-auto-generated prose below the keyword table
similarly mixes still-valid spec semantics with stale parser
mechanics. The whole document needs a rewrite from grammar2's
point of view, plus a survey for any other docs additions worth
landing.

**Scope.**
- **Generator migration.**
  - Port `internal/parsers/grammar/gen/main.go` to read from
    `internal/parsers/grammar2/keywords.go` (the new source of
    truth). Likely land it as `internal/parsers/grammar2/gen/`
    so the generator dies with grammar v1 in M6.1, not before.
  - Update the `//go:generate` directive that currently lives on
    `grammar/keywords_table.go`.
- **`docs/annotation-keywords.md` rewrite.**
  - Drop the "v2 grammar parser" framing. The doc is about the
    OAS v2 (Swagger 2.0) annotation surface this scanner
    recognises; mention grammar2 only where the lexer-level
    contract actually leaks into the user model (e.g. comma-list
    vs raw-block keywords).
  - Regenerate the keyword table from grammar2 — note any
    additions, removals, or context changes vs the master version
    of the doc.
  - Rewrite the §Details section (or whatever follows the table)
    to match grammar2 reality: the four annotation families
    (schema / parameters / responses / operation / route / meta /
    classifier), the context legality matrix, the YAML body /
    comma-list / inline-value shapes, what happens to keywords
    used in the wrong context.
  - Add a §Quirks section pointing at the per-builder README
    quirk logs for the cases users actually hit (block-comment
    Title/Desc, `swagger:file` body-only, etc.).
- **`docs/` audit beyond annotation-keywords.**
  - Scan the doc-site repo refs (`go-openapi.github.io`
    maintainer docs are linked from CLAUDE.md) and identify
    anything else worth landing as `docs/` content — getting
    started, annotations primer, troubleshooting common-mistakes.
    Land at most one new doc; this isn't a docs-site
    replacement.

**Exit criteria.**
- `docs/annotation-keywords.md` regenerates cleanly from
  grammar2's keyword table; the auto-gen header points at the
  grammar2 source.
- No prose in `docs/` references grammar v1, the SectionedParser,
  or `.claude/plans/`.
- `go generate ./...` re-emits the doc unchanged on a clean tree
  (golden semantic for the doc itself).

### M7 — History cleanup + merge 🟡 queued

**Scope.**
- Squash the 83-commit branch history into ~10–20 reviewable units.
  Chunk shape provided by fred at this point — likely along the lines
  of:
  - `feat(grammar2): lexer + parser + diagnostics`
  - `feat(parsers): YAML sub-language + typed extensions`
  - `refactor(schema): grammar2 migration + redesign`
  - `feat(schema): close override gaps + recognizer plumbing`
  - `refactor(parameters,responses,items): grammar2 migration + SimpleSchema`
  - `refactor(operations,routes): grammar2 migration`
  - `chore: retire legacy parsers + regex consumers`
- DCO-signed, real-email co-author on each squash commit per
  `.claude/rules/contributions.md`.
- Final `golangci-lint run --new-from-rev master` + full
  `go test ./...` clean.
- Merge to `master`.

**Exit criteria.**
- Branch merged.
- Memory entries refreshed (`project_grammar_parser_migration`,
  `project_v2_vision`, `project_schema_refactor_next_wave`) — note
  what landed and which v2 items remain.

---

## 3. Post-merge hand-off

The post-merge roadmap splits into two streams already living in
plan-docs:

- **Feature work** — `forthcoming-features.md` (workshop W2 enum, W3
  example, W4 private-comment; sub-parser typing improvements; column
  precision under `/* */` continuation; etc.).
- **v2 redesign** — `project_v2_vision` memory: OAI 3.x, smart
  detection, LSP-grade per-entry diagnostic positions
  (`.claude/plans/typed-extensions.md` round-3 hooks into this), the
  routes / operations synonym unification (this M-stream defers it
  explicitly), allOf-composition for stale-enum-desc
  (`schema/README.md` §quirks-open), and the cross-package
  definition-name collision design call (same).

### Post-merge backlog items added during the M-stream

These accumulated during M6.5 / M6.6 / M6.7 sweeps as "real work but
not blocking the merge." Captured here so they're not lost.

**Quirks-refresh follow-ups** (decided post-M6.5; sweep session of
its own, before release):

- **Q26** — `Terms Of Service:` raw-block absorbs an adjacent
  `Schemes:` keyword in real-world meta blocks (observed in the
  genspec-tui demo against go-swagger/examples/generated petstore).
  Likely triggered by M6.5-D's KwSchemes shape change
  (asCommaList → asRawBlock); both keywords are now asRawBlock and
  the terminator rule in
  `internal/parsers/grammar/lexer.go:isSiblingTerminatorFor` may not
  pair them. No witness fixture yet; minimised repro needed.
- **Q7 / Q11 / Q12 / Q13** — alias-handling theme. Cross-cutting
  design call needed; defer to v2 redesign.
- **Q8 / Q9** — embed-as-allOf vs inline asymmetry; interface
  property naming. Document-tractable; either accept as intentional
  or open the design call.

**Godoc link-compliance pass** (added after M6.6 sweep):

Every M6.6-touched godoc uses the `[§anchor](./README.md#anchor)`
markdown-link form. That renders cleanly for in-repo readers but
NOT on pkg.go.dev — pkg.go.dev treats relative URLs as plain text.
Convert to godoc's reference-link format with absolute GitHub URLs:

```go
// See [§blockcache] for cache scope.
//
// [§blockcache]: https://github.com/go-openapi/codescan/blob/master/internal/builders/common/README.md#blockcache
```

Plus `Symbol.Method` → `[Symbol.Method]` for cross-package symbol
references (so pkg.go.dev auto-links them).

The canonical reference shape is already in
`internal/builders/common/builder.go` (the only M6.6 file converted
to the compliant form). Mechanical sweep — find/replace per package
+ footer injection. Defer to **after merge** because the diff
churn is huge (~100+ link references across 8 READMEs and matching
godoc) and a script is the right tool, not hand-editing.

Lands LAST in the post-merge backlog — after the quirks-refresh
session and any v2-related work that would shift link targets
again.

**Doc generator (long-term)** — referenced from M6.7 plan: a doc
generator like the one built for `go-openapi/testify`
(`github.com/go-openapi/testify/codegen/internal/generator`,
~5000 LOC, produces <https://go-openapi.github.io/testify>) could
eventually replace `./docs/` hand-maintenance. Out of scope for the
M-stream and the immediate post-merge backlog — a project of its
own. Filed here so the option doesn't get forgotten.

---

## 4. References

- `grammar/60-implementation-roadmap.md` — the source-of-truth Phase
  1–5 roadmap that M-stream operationalises for the post-schema
  builders.
- `grammar/p7-schema-builder-redesign.md` §9.1 — post-P7
  cross-builder factorisation list (informs M6 helper relocations).
- `forthcoming-features.md` — post-merge feature pipeline.
- `deferred-quirks.md` — v2 deferrals; routes/operations quirks
  surfaced during M5 land here.
- `internal/builders/schema/README.md` — recipe template the M0
  refactor wave replicates; quirk-classification pattern M5 replicates
  for routes.
- `typed-extensions.md` — Round-3 deferred; informs LSP timing.

---

## 5. Closing reflection — Stream M wrap

Captured from the post-M6.5-E debrief, end of the code-refactor phase
of the M-stream, before the M6.6 comment sweep and M6.7 docs rewrite.
A bit self-congratulating — but the journey is worth noting on its
own terms, and the patterns that worked here deserve to be locked
into the institutional memory before they fade into "of course we
did it that way."

### The longer arc (Fred's framing)

The starting point was 2015: a soup of regexes that *apparently
worked*. Compact, dense, unique across the Go ecosystem, with a
nice-looking doc site. The design had every appearance of pure
genius.

In practice:

- Even relinting was painful.
- The tight coupling + fancy regexes made the code indistinguishable
  from magic.
- The style itself was cryptic — infinitely nested ifs and fors,
  functions unrolling weird AST + regex logic over pages of code,
  no comments worth speaking of.
- Every contribution added its own "special case." No rules, no
  order. Rapidly degraded the tool with hundreds of unfixable
  issues; every bug fix added new ones.

Fred stayed clear of this package for years and focused on the other
half of go-swagger (codegen). Then, in 2025, the dam broke:

- Adding type-alias support (go1.22).
- A large relinting pass, just to see more clearly.
- Carve-out to a separate repo (more maintainable, releasable on
  its own).
- The package layout refactor — failed twice solo, failed twice with
  AI, eventually accomplished together.

Once those structural moves landed, the dependencies and structural
issues became visible. The next enablers were:

- The goldens test harness — crucial. The lever that made big
  structural changes safe.
- A 2026 vision:
  - Grammar-based, no regexp, fewer quirks — **we are here at the
    end of Stream M**.
  - On-demand code scanning with cache (next).
  - Improved spec builder (no hidden state for discovery).
  - Planned: decouple output from `go-openapi/spec` (internal model).
  - Planned: evolve toward LSP capabilities + IDE integration.
  - OpenAPI v3 ambition becoming palatable in 2026.

Parallel UX work (other sessions, other branches): interactive
genspec-tui to render spec live; planned Hugo + WASM static
"spec gen playground"; WIP TUI/WASM widget for issue reporting with
scrambled/anonymised context.

### What worked, from inside the stream

Three patterns made M6.5 land cleanly across seven commits, where
prior refactor attempts collapsed:

**The witness-then-fix pattern is the real superpower.** PRE → A →
B → C → D → E was only sane because every commit had its golden
diff as the proof. Without the fixtures, the refactor would have
been "trust me." With them, behaviour changes were visible,
discussed, and chosen. The 12 quirks (Q14–Q25) surfaced and
resolved over the stream were almost all NOT in the original
`observed-quirks.md` list — the harness *found* them. Pre-refactor
quirk discovery was guesswork; with the fixture suite it became
arithmetic.

**Unification was a probe, not just an aesthetic move.** Each
"let's make this consistent across builders" step exposed
something that had been broken or asymmetric for years:

- `Consumes: application/json` inline silently lost in the lexer
  (latent since the grammar's day-1 design — only
  `collectRawValue` had a single-line path, `collectRawBlock`
  didn't).
- `Schemes:\n  - http` multi-line silently lost (KwSchemes was
  `asCommaList()`, never expanded into body).
- `parsePathAnnotation`'s synthetic per-line ast.Comments tripping
  the lexer's default branch (trimCommentPrefix workaround).
- `validateExtensionNames` being the lone hold-out for
  error-on-typo when every other annotation silently dropped.

The principle: when consolidating helpers across builders surfaces
weird asymmetries, those asymmetries are usually bugs hiding in
sheep's clothing. Challenge ad-hoc helpers; expect to find latent
bugs.

**The branch never broke.** Seven commits, every one green, every
one a defensible standalone refactor. Achieved by the "pause and
reassess" rhythm — debate the design as documents before code,
chunk by concerns rather than by code-paths, capture witness
fixtures before refactoring the behaviour they witness. The
alternative — big-bang merge that goes red for a week — is how
most refactors die.

### What's structurally ready for the next chapter

- `handlers/` seam is solid; OAS3 keywords plug in by adding
  entries, not reshaping dispatch.
- `routebody` proved that sub-parsers in `parsers/` coexist with
  the grammar lexer cleanly. Repeatable pattern for future
  sub-languages.
- `block.Extensions()` with `Source` carries the right metadata
  surface for future targets (Server objects in v3, etc.).
- Fixture suite + golden capture is now muscle memory — new
  fixtures cost minutes.
- `observed-quirks.md` is the institutional memory: Q1–Q25 with
  fix locations, witnessed by fixtures. A future contributor reads
  that and *knows* what's intentional.

What's still untouched but worth flagging: the scanner-level
regexes in `parsers/` (rxRoute, rxOperation, etc.) and the
package-loading state in `scan_context.go`. The grammar work didn't
reach them because they're at the discovery/classification
altitude, not the parsing altitude. Both are "next chapter"
(post-merge scanner rework) rather than "still on this chapter."

### Genspec-tui live demo (end-of-stream snapshot)

Fred shared a screenshot of `genspec-tui` running against the
`go-swagger/examples/generated` petstore source — a round-trip
test: source that was code-generated FROM a spec, scanned back to
produce a spec. Status line: **`JSON · 12 paths · 5 defs · ready
(811ms)`**.

What the TUI demonstrates structurally:

- Three-panel layout — tree | live-spec | diagnostics — that maps
  directly to "what was scanned, what came out, what got flagged."
  Diagnostics is a first-class panel; the diagnostic stream we
  carefully built in M6.5 lands in the user's eye-line, not buried
  in a log.
- 811ms for the full loop (fs walk, package load, go/types,
  grammar, builders, render) on a non-trivial example. Inside the
  ~300ms perceived-budget for interactive TUI work after on-demand
  scanning + cache lands.
- The TUI branch lags M6.5 fixes; it rebases onto master once
  Stream M merges. The diagnostics panel showing "(no diagnostics)"
  in the screenshot is therefore the *pre-fixes* state; the
  post-merge demo will surface the new diagnostic stream.

### Open observation — TOS absorbing Schemes

One thing in the screenshot that warrants follow-up
(post-merge, not now):

```
"termsOfService": "http://helloreverb.com/terms/\nSchemes:\nhttp"
```

`Schemes:\nhttp` was absorbed into the TOS body. `KwTOS` is
`asRawBlock()` and should terminate at the next sibling keyword
(`KwSchemes`). The diagnostics panel was empty, so the lexer
didn't even notice. Likely one of:

- The petstore example's source formatting (indentation that makes
  the Schemes line look like body continuation to
  `isSiblingTerminatorFor`).
- A genuine gap in the TOS terminator rule.
- M6.5-D's KwSchemes shape change (asCommaList → asRawBlock)
  shifted the terminator semantics for adjacent raw-block keywords.

Invisible to the existing fixture suite (no meta fixture currently
puts TOS adjacent to Schemes). A minimised repro fits in the M6.6
sweep or the post-merge scanner rework. Logged here so it doesn't
fall through the cracks.

### Notes for future-Fred (and future-me)

- The patterns that worked here are repeatable; the next big
  refactor (scanner rework, OAI 3 support, LSP plumbing) should
  open with a fresh witness-fixture suite, the same way M6.5-PRE
  did. Don't skip the harness step — that's where the discovery
  budget actually lives.
- "Challenge ad-hoc helpers" is a reliable heuristic. When a
  function exists only in one builder and has no equivalent
  elsewhere, the question is rarely "should we keep it" but "what
  asymmetry is it papering over."
- The `observed-quirks.md` Q-numbering is a low-overhead
  institutional memory. Every behaviour change that ships should
  earn a Q-number with a witness fixture. The cost is ~10 minutes;
  the payoff is that the next maintainer doesn't have to
  reverse-engineer the rationale.
