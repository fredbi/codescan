# Coverage Analysis Report

Date: 2026-03-23
Coverage: 79.1% of statements (up from ~78.5% baseline)

## Per-file coverage

| File | Statements | Covered | % |
|------|-----------|---------|---|
| enum.go | 9 | 9 | 100.0 |
| routes.go | 36 | 35 | 97.2 |
| route_params.go | 96 | 93 | 96.9 |
| parser_helpers.go | 23 | 22 | 95.7 |
| operations.go | 55 | 50 | 90.9 |
| parser.go | 763 | 676 | 88.6 |
| meta.go | 108 | 92 | 85.2 |
| taggers.go | 33 | 27 | 81.8 |
| application.go | 345 | 276 | 80.0 |
| schema.go | 875 | 640 | 73.1 |
| responses.go | 245 | 172 | 70.2 |
| spec.go | 109 | 75 | 68.8 |
| parameters.go | 294 | 198 | 67.4 |
| assertions.go | 7 | 4 | 57.1 |

## What we fixed / covered in this session

1. **Bug fix**: `setMultipleOf.Parse` (parser.go) had `len(matches) > 2` instead of `> 1`.
   The regex for `multiple of:` only captures 1 group, so the condition was always false.
   All `multiple of:` annotations were **silently ignored**.

2. **Newly fully covered functions** (were 0%):
   - `debugLogf` — via `Options{Debug: true}` test
   - `schemaValidations.SetMultipleOf` — via existing fixture annotation (now that the bug is fixed)
   - `paramValidations.SetMultipleOf` — same
   - `headerValidations.SetMultipleOf` — same
   - `itemsValidations.SetMaximum` — via new `num_slice` fixture field
   - `itemsValidations.SetMinimum` — same
   - `itemsValidations.SetMultipleOf` — same
   - `itemsValidations.SetUnique` — same
   - `itemsValidations.SetCollectionFormat` — same

3. **Newly partially covered** (were 0%):
   - `schemaBuilder.buildNamedArray` — via `NamedArray GoArray` field in SpecialTypes fixture

4. **Error path coverage**:
   - `Run()` error path — via `TestRun_InvalidWorkDir`

## Classification of the remaining ~21% uncovered code

### 1. Interface-required dead methods (~15 functions, ~2% of total)

These exist solely to satisfy the `swaggerTypable` interface but are never dispatched to
on these concrete types:

- `paramTypable.{In, Level, SetRef, AddExtension}`
- `itemsTypable.{In, Level, Schema, AddExtension, WithEnum}`
- `responseTypable.{In, SetSchema, CollectionOf, AddExtension, WithEnum}`
- `schemaTypable.In`

**Verdict**: Structurally unreachable. Not worth testing. A future refactoring could split
the interface into smaller role-specific interfaces to eliminate these dead methods.

### 2. Dead feature paths (~3 functions)

- `setupRefParamTaggers` — requires `paramTypable.SetRef` to fire (also dead), which would
  mean a non-body parameter resolves to a `$ref`. This can't happen with valid Swagger 2.0
  parameter types. The ref tagger exists but is never exercised.
- `parameterBuilder.makeRef` — same reason, unreachable for parameters.
- `responseBuilder.makeRef` — similar: response-level `$ref` resolution doesn't occur through
  the current response building flow.

**Verdict**: Likely vestigial code from an earlier design or intended for future use.
Could be removed or explicitly marked as unused.

### 3. Error/defensive paths (~50+ locations, ~8% of total)

Scattered `return err` / `return nil` paths in builder methods. Examples:
- `specBuilder.Build` (spec.go:55-85) — 8 error returns from sub-builders
- `typeIndex.build/walkImports` error paths
- `schemaBuilder.buildFromDecl` uncovered switch cases for exotic Go types
- `parseTags` error paths for malformed struct tags
- Various `buildFrom*` methods' early error returns

These are hard to trigger because they require:
- Internal builder failures (not input-controllable)
- Exotic Go type constructs not present in fixtures
- Malformed Go source that still parses as valid AST

**Verdict**: Low value for unit testing. Some could be covered with carefully crafted
fixture files that produce unusual AST constructs. The error paths in `specBuilder.Build`
could be covered by mocking individual builders, but the current design doesn't support
dependency injection — all builders are created internally.

### 4. Untested feature paths (~10% of total)

These are real code paths exercised by specific Go type/annotation combinations:

**Alias handling in parameters and responses** (~60 uncovered lines each):
- `parameterBuilder.buildAlias` / `buildFieldAlias`
- `responseBuilder.buildAlias` / `buildFieldAlias`
- These handle type aliases in parameter/response struct fields with `RefAliases: true`
- The `TransparentAliases` tests cover some alias paths but not the ref-producing ones
- Would need new fixture files with aliased types in `swagger:parameters`/`swagger:response` structs

**Schema builder edge cases** (~100 uncovered lines):
- `buildNamedType` exotic type switches (channels, functions, unsafe pointers as named types)
- `buildNamedEmbedded` / `buildEmbedded` — embedded alias/interface handling
- `processAnonInterfaceMethod` / `processInterfaceMethod` — interface method return types
- `buildNamedAllOf` — allOf with aliases
- `buildFromTextMarshal` — TextMarshaler edge cases (cross-package, strfmt)

**Application scanning** (~40 uncovered lines):
- `scanCtx.findEnumValue` — enum value extraction from const blocks
- `typeIndex.detectNodes` — swagger annotation detection for various node types
- `typeIndex.walkImports` — transitive dependency walking error paths

## Recommendations for future test improvement

### Short term: more fixtures

Add fixture files for:
1. **Aliased types in parameters** — a `swagger:parameters` struct with fields typed as
   aliases, scanned with `RefAliases: true`
2. **Aliased types in responses** — same for `swagger:response`
3. **Interface methods with complex return types** — interfaces returning slices, maps, pointers
4. **Enum constants** — const blocks with iota patterns for enum extraction

### Medium term: design refactoring for testability

The main barrier to higher coverage is that builders are tightly coupled:
- `specBuilder` creates `schemaBuilder`, `parameterBuilder`, `responseBuilder` internally
- No way to inject failures or mocks at builder boundaries
- Error paths in `specBuilder.Build` require actual builder failures

Refactoring options:
1. **Extract builder interfaces** — allow `specBuilder` to accept builder factories,
   enabling test doubles that return errors
2. **Split `swaggerTypable`** — the monolithic interface forces dead method implementations.
   Splitting into `Typable`, `Refable`, `Extensible` etc. would eliminate dead code
3. **Make `scanCtx` methods return errors consistently** — some methods silently skip
   (e.g. `FindModel` returns bool), making error paths untestable

### Long term: integration test harness

A comprehensive integration test that runs `Run()` against a large real-world Go project
(or a synthetic one covering all annotation types) would exercise many of the partially
covered paths without needing individual unit tests for each builder method.
