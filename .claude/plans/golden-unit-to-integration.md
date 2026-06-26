# Migrate builder-unit golden tests → integration tests

## Problem

Golden-file comparisons (`scantest.CompareOrDumpJSON`) are used in **builder-unit
tests** under `internal/builders/{schema,parameters,responses,routes,operations}`.
These run a single builder directly, **without the spec-level `reduce` stage**, so
they dump the **pre-reduce, fully-qualified build form** of definition keys and
`$ref`s (`#/definitions/github.com/.../order`). After the name-identity work this
became visible as fq "leaks" in 19+ committed goldens.

Diagnosis (verified): the `reduce` rewrite pass is **complete** — every
full-pipeline golden (`*_spec.json`, `enhancements_*`, all `internal/integration`
output) is clean. The fq only appears in builder-unit goldens because reduce never
runs there.

**Decision (Fred):** golden-dumping belongs to **integration (full-pipeline)
tests**, which produce clean, reduced, representative specs. Golden dumps in
unit tests "just don't make sense" — unit tests should assert specific
properties, not snapshot a whole spec. Remove golden dumps from builder-unit
tests; move/ensure the snapshot coverage in `internal/integration`.

## Scope (surveyed)

36 `CompareOrDumpJSON` calls across 7 files, ~32 distinct goldens:

| File | dumps |
|------|-------|
| `schema/schema_test.go` | 13 |
| `parameters/parameters_test.go` | 8 |
| `responses/responses_test.go` | 7 |
| `schema/schema_go118_test.go` | 4 |
| `routes/routes_test.go` | 2 |
| `operations/operations_test.go` | 1 |
| `operations/operations_go119_test.go` | 1 |

Fixtures behind them (petstore, classification/goparsing, bugs/3125, go118/go119,
enhancements pointers, transparent-alias) are **already scanned** by
`internal/integration` (148 bugs/, petstore, classification, goparsing refs) —
so most unit goldens are **redundant** with existing or easily-added integration
goldens. The unit set additionally covers: option matrices (`skipext`,
`descwithref`, `file`, transparent-alias), issue-specific cases (2007/2011),
struct/interface discriminators, pointers±x-nullable.

## Principle: no coverage loss

A unit golden is removed **only** when its fixture+options shape is captured by a
full-pipeline integration golden (or replaced by explicit property assertions).
The integration golden is the *better* artifact (reduced, end-to-end). Verify each
migration by generating the integration golden and confirming it captures the same
shape (modulo fq→reduced names).

## Method (per unit golden)

1. Identify the fixture + builder options behind the dump.
2. **Redundant?** If an integration test already produces an equivalent
   full-pipeline golden → delete the unit golden; keep any *property* assertions
   in the unit test, drop the `CompareOrDumpJSON` line; delete the orphan golden file.
3. **Unique coverage?** (option matrix / issue case not in integration) → add a
   focused integration test (reuse the fixture; add the option variant) that dumps
   a clean reduced golden; then remove the unit dump + delete the orphan.
4. **Golden-only unit test** (no property assertions) → coverage moves entirely to
   integration; delete the unit test.

## Execution

- Phase per builder package: schema → parameters → responses → routes/operations.
- Repetitive once the pattern is set → delegate per-package to subagents.
- Gate each phase: full `go test ./...` + `-race` (spec/integration) + lint, and a
  **new leak-guard test** in `internal/integration` that fails if ANY golden under
  `fixtures/integration/golden/` contains `#/definitions/github.com` or a fq
  definition key — locks the "real output is always reduced" invariant permanently.
- Net result: zero `CompareOrDumpJSON` in `internal/builders/*`; the ~32 fq
  goldens deleted; equivalent clean goldens in integration; unit tests keep only
  property assertions.

## Reassessment (empirical, 2026-06-17)

A full-pipeline scan only snapshots **route-reachable** content. Verified on the
classification fixture:

- **Schema/model goldens DO consolidate**: every classification model appears in
  `classification_spec.json` (ScanModels). The `*_schema_*` unit goldens are
  redundant → delete, keep the unit tests' property assertions.
- **Param/response goldens do NOT consolidate**: the unit tests build *curated
  structs* (`MyFileParams`, `MultipleOrderParams`, `EmbeddedFileParams`, file
  params, …) that are not wired to any route, so the spec never reaches them.

**Revised method by builder:**
- schema / models → consolidate into a full-pipeline `*_spec.json` (option-matrix
  variants where the unit goldens had them); delete the unit goldens.
- parameters / responses (and any curated, non-route-reachable builder input) →
  **drop the golden dump, add targeted property assertions** (Fred's decision).
  The builder is still exercised; we assert the salient shape instead of snapshotting.
- A standalone full-pipeline integration golden is added per fixture where it adds
  real coverage (e.g. `classification_spec.json` ×4 option variants — NEW).

## Decisions (settled)

- **Branch:** follow-on commit(s) on `feat/name-identity-cyclic-ref` (the fq leak
  that motivated it came from that work).
- **Golden-only unit tests:** delete outright (the integration golden fully
  replaces them). Unit tests that ALSO carry property assertions keep the
  assertions and only drop the `CompareOrDumpJSON` line.
