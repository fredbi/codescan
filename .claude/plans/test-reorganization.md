# Test Reorganization after the internal/ refactor

Date: 2026-04-17
Status: DRAFT — pending review by Fred
Branch: `refact/tests`

## Problem statement

The refactor from a single `codescan` package into `internal/{scanner, parsers,
builders/*, ifaces, logger, scantest}` carved the source files into small
narrow-focused packages, but **the tests were split along the same lines without
re-thinking their scope**. Many tests that were written as unit tests in the
monolith were actually cross-cutting integration tests that reached into
scanner state, used parser regexes, and asserted against the final Swagger
output all in one go. When Go's package boundary forces them apart, they fail
to compile — and any attempt to let one test package reach into another creates
import cycles.

`go vet ./...` currently fails with, among others:

```
internal/builders/operations/operations_go119_test.go:16:15: undefined: NewScanCtx
internal/builders/routes/routes_test.go:16:20: undefined: rxRoute
internal/builders/responses/responses_test.go:370:24: undefined: ScanCtx
internal/builders/parameters/parameters_test.go:27:25: undefined: ScanCtx
internal/builders/schema/schema_go118_test.go:25:15: undefined: Newscanner
```

These symbols are fine where they now live; it's the tests that are wrong about
where to look for them.

## Diagnosis in one picture

```
┌── scanner ─────────────┐   (ScanCtx, NewScanCtx, Application, Options)
│                        │
├── parsers ─────────────┤   (rxRoute, MetaParser, …)
│                        │
├── builders/            │   (schema, operations, routes, parameters, …)
│   ├── schema           │
│   ├── operations       │   each imports scanner, parsers, ifaces
│   └── …                │
│                        │
├── scantest ────────────┤   imports scanner ONLY
│   (load, property,     │   cannot grow deps upward without cycling
│    classification,     │
│    mocks)              │
│                        │
└── codescan (root) ─────┘   imports scanner + builders/spec
    api.go  →  Run()          api_test.go is the only true root-level test today
```

The rule to protect: **`scantest` must stay at the bottom of the production
dependency graph**. Every time we were tempted to put something "convenient"
there that needs `parsers` or `builders/*`, we would reintroduce the cycle.

## Three test layers

Separate tests by what they need to *see*, not by which source file they
verify:

| Layer | Lives in | Package clause | Can import |
|---|---|---|---|
| **unit** | alongside the code | `package X` | only that package |
| **cross-package** | alongside the code | `package X_test` | siblings & higher layers |
| **integration** | `internal/integration/` | `package integration_test` | everything, including root `codescan` |

### Why `package X_test` solves most of it

A file declared `package schema_test` while sitting in `builders/schema/` is
invisible to the production dependency graph — it is effectively a leaf of its
own. It can import `scanner`, `parsers`, any sibling builder, and `scantest`
without creating a cycle, because nothing in production imports `schema_test`.

This is Go's built-in answer to the "tests need to see more than production
code does" problem. We should use it liberally.

### Why a dedicated `internal/integration/` package

Some tests really are full-pipeline — "scan these fixtures, call `Run()`,
verify the whole Swagger doc". They should not live inside any builder: they
don't test *that* builder specifically, they test the whole pipeline through
the public `codescan.Run()` entry point. Moving them to
`internal/integration/` makes two things explicit:

- they're slow/heavy (run selectively in CI)
- they own the "final Swagger spec" assertion vocabulary (don't leak it into
  `scantest`)

### The inviolable rule

> **A test helper lives at the lowest common ancestor of everything that uses
> it — never higher.**

Today `scantest` holds fixture-loading because scanner tests and builder tests
both use it. ✓
Tomorrow, if we add a verifier that needs the full Swagger output, it only has
integration-test consumers → it belongs in `internal/integration/verify/`,
**not** promoted into `scantest` even though that feels tidier. Violating this
rule is how the cycle comes back.

## Target layout

```
/
├── api.go                               public API — unchanged
├── api_test.go                          thin public-API smoke test only
├── internal/
│   ├── scanner/
│   │   ├── *.go
│   │   ├── *_test.go                    package scanner — unit tests
│   │   └── export_test.go               exposes unexported hooks if needed
│   ├── parsers/
│   │   ├── *.go
│   │   └── *_test.go                    package parsers — unit tests
│   ├── builders/
│   │   ├── schema/
│   │   │   ├── schema.go
│   │   │   ├── schema_test.go           package schema — unit tests only
│   │   │   └── schema_external_test.go  package schema_test — cross-pkg tests
│   │   ├── operations/                  (same pattern)
│   │   ├── routes/                      (same pattern)
│   │   └── …
│   ├── scantest/                        leaf helpers: loaders + schema asserts
│   │   ├── load.go
│   │   ├── property.go
│   │   ├── classification/              fixture-specific expected values
│   │   └── mocks/                       interface mocks
│   └── integration/                     NEW — leaf of the dep graph
│       ├── petstore_test.go             full-pipeline runs via codescan.Run
│       ├── classification_test.go
│       ├── go119_test.go
│       └── verify/                      full-spec verifiers
│           ├── petstore.go
│           └── classification.go
└── fixtures/                            unchanged
```

## Migration plan (phased, each phase leaves the tree compiling)

### Phase 0 — Prep

**Decision: prefer `export_test.go` hooks over widening `ScanCtx`'s exported
surface.** (Fred, 2026-04-17.) The iterator-style accessors that already exist
on `ScanCtx` (`Operations()`, `Routes()`, `Responses()`, `Parameters()`,
`Models()`, `Meta()`, `ExtraModels()`, …) are kept where their consumers are
production builders — they're legitimate production API, not test
infrastructure. Anything *beyond* what builders need, tests should reach via
`export_test.go` hooks so the public API doesn't carry test-only weight.

Concretely:

1. **Audit existing accessors on `ScanCtx`.** For each exported method on
   `ScanCtx`, grep for non-test callers. Any method with only test callers is
   a candidate to move into `scanner/export_test.go` as a test-only hook.
   Expected fate:
   - `Operations()`, `Routes()`, `Responses()`, `Parameters()`, `Models()`,
     `Meta()`, `ExtraModels()`, `FindDecl`, `FindModel`, `PkgForPath`,
     `DeclForType`, `PkgForType`, `FindComments`, `FindEnumValues`,
     `MoveExtraToModel`, `NumExtraModels`, `SkipExtensions`, `DescWithRef`,
     `SetXNullableForPointers`, `TransparentAliases`, `RefAliases`, `Debug`
     — all called by builders, keep as-is.
   - Anything that turns out to have only test callers → move to
     `export_test.go`.

2. **Add `export_test.go` hooks for unexported state tests need.** The broken
   tests dereference `sctx.app.Operations`, `sctx.app.Routes`, etc. directly.
   Since the iter-based accessors don't give them indexable slices, they'll
   need either:
   ```go
   // scanner/export_test.go — test-only, not in production API
   func (s *ScanCtx) AppOperationsSlice() []parsers.ParsedPathContent { return s.app.Operations }
   func (s *ScanCtx) AppRoutesSlice()     []parsers.ParsedPathContent { return s.app.Routes }
   ```
   Use these sparingly — prefer rewriting the test to use the iter.Seq
   accessor when a range loop is sufficient.

3. **`export_test.go` scope — important constraint.** `export_test.go` in
   `package scanner` is compiled **only when `go test ./internal/scanner/...`
   runs**. When `go test ./internal/builders/schema/...` builds, it sees
   `scanner` as a production dependency and its `_test.go` files are
   invisible. So `export_test.go` helps:
   - tests in the *same directory* as their target (scanner's own
     `*_test.go` files, including `package scanner_test` external tests
     colocated with scanner), **not**
   - tests in *other* packages that import scanner.

   This means `export_test.go` solves visibility problems for **a package's
   own tests**, not for cross-package tests. For cross-package needs, the
   options narrow to:
   - use an existing exported iter.Seq accessor (preferred);
   - rewrite the test to live in `internal/integration/` and use the public
     `codescan.Run()` path;
   - as a last resort, add a real exported method (widening the API).

   The `export_test.go` pattern is still useful — just in the **builder**
   packages, not in scanner. For example, `builders/schema/export_test.go`
   can expose an unexported builder constructor so
   `builders/schema_test.go` (external package `schema_test`) can construct
   a `SchemaBuilder` without widening `schema`'s real API.

4. **Ensure every existing fixture path still resolves from its new test
   location.** Most tests walk up two levels (`../../fixtures/...`);
   integration tests will walk up one (`../fixtures/...` from
   `internal/integration/`).

### Phase 1 — scanner tests

Scope: `internal/scanner/*_test.go`
- These already compile under `package scanner`. No structural change.
- Audit any test that reaches into other packages (shouldn't exist in scanner).

Acceptance: `go test ./internal/scanner/...` passes.

### Phase 2 — parsers tests

Scope: `internal/parsers/*_test.go`
- `meta_test.go` already imports `scantest/classification` and compiles fine.
- Audit regexes used in tests — they are package-local (like `rxRoute`), so
  tests against them stay `package parsers`.

Acceptance: `go test ./internal/parsers/...` passes.

### Phase 3 — Move full-pipeline tests out of builders

This is the biggest phase.

For each `internal/builders/X/X_test.go`:

1. Classify each test function:
   - **unit** — works on constructed inputs, no `ScanCtx`, no fixtures → keep
     `package X`.
   - **cross-package, not full pipeline** — needs a real `ScanCtx` from
     `scantest` loaders but only exercises this builder → move to
     `package X_test` in a new `X_external_test.go`.
   - **full pipeline** — calls `Run()` or the equivalent end-to-end flow
     through `spec.NewBuilder(…).Build()` → move to `internal/integration/`.

2. External tests import what they need:
   ```go
   package schema_test

   import (
       "testing"
       "github.com/go-openapi/codescan/internal/builders/schema"
       "github.com/go-openapi/codescan/internal/scanner"
       "github.com/go-openapi/codescan/internal/scantest"
   )
   ```

3. Where tests poke unexported fields (`builder.ctx`, `builder.decl`), we have
   three choices in order of preference:
   - Replace with a real constructor if one exists (`schema.NewSchemaBuilder`).
   - Add an `export_test.go` in the builder package with a tiny internal
     constructor for tests.
   - Keep that test as `package X` (unit test) and move the integration-y part
     out.

Acceptance after Phase 3: `go test ./internal/builders/...` passes.

### Phase 4 — Create `internal/integration/`

**Decision: `api_test.go` stays as a thin public-API smoke test only.** (Fred,
2026-04-17.) All fixture-heavy verifications move out.

- Extract from `api_test.go` the fixture-heavy verifications (currently ~500
  lines) and the helpers in `verifyParsedPetStore` etc. Move them to
  `internal/integration/`.
- Put the verifier helpers under `internal/integration/verify/` so multiple
  test files can share them.
- Leave at the root **only** small, fast, public-API-only tests:
  - `TestRun_InvalidWorkDir` — surface that `Run()` errors gracefully on bad
    input.
  - `TestApplication_DebugLogging` — surface that `Debug: true` opt works.
  - A minimal `TestRun_SmokeTest` that points at a tiny fixture and asserts
    only that `Run()` returns a non-nil `*spec.Swagger` with at least one
    path. No detailed field-by-field checks here — those belong in
    `internal/integration/`.
- The root `api_test.go` should not import `scantest` or `classification`
  verifiers. If it does, the smoke test has drifted out of scope.

Acceptance after Phase 4:
- `go test ./...` passes.
- Root smoke tests run in <1s (guideline, not a hard CI gate).
- `go test -count=1 .` at the repo root runs only the root smoke tests.
- `go test ./internal/integration/...` runs the full-fixture suite.

### Phase 5 — CI wiring (optional, separate PR)

Add a split in `go test` invocations so unit tests run fast and integration
tests run in their own matrix slot. Out of scope for this PR but worth noting
as the natural follow-up.

## Open decisions

1. **Fixture loaders caching.** Current `scantest.LoadPetstorePkgsCtx` uses
   package-level globals to cache scan contexts across test functions. This
   works within a single `go test ./package` invocation but each package
   re-runs its own `scantest.init`, so the cache is per-test-binary anyway.
   Should integration tests share the cache? For now, keep the current
   behavior — optimizing compile-time fixture loading is a separate topic.

2. **Naming of external test files.**
   Options: `X_external_test.go`, `X_black_test.go`, `X_pipeline_test.go`.
   Go has no convention here; pick one and be consistent. I suggest
   `X_external_test.go` for its "external test package" literalness.

## Acceptance criteria

- `go vet ./...` clean.
- `go test ./...` passes.
- `golangci-lint run` clean (patch-level, per `--new-from-rev master`).
- No test file imports a sibling `internal/builders/*` package from a
  `package X` test file (only `package X_test` files are allowed that
  privilege).
- `scantest` still imports only `scanner` and `ifaces`.
- `api.go` public API unchanged.

## Risks

- **Hidden coupling.** Some "unit" tests may actually depend on scanner-side
  state we haven't noticed. Mitigated by the phased approach — each phase
  compiles before the next starts.
- **Fixture path drift.** When we move tests between packages, relative
  fixture paths change. Mitigated by grep for `fixtures/` and
  `../../fixtures/` after each phase.
- **Churn for reviewers.** This PR will touch a lot of files. Keep the
  per-phase commits tight and name them clearly (`test(schema): split
  external pipeline tests`) so the reviewer can follow.

## Out of scope

- Performance of fixture loading.
- v2 test strategy (smart detection, LSP parser, etc.) — this plan is
  strictly about unblocking the current suite after the refactor.
- Adding new test cases. Move first, improve later.
