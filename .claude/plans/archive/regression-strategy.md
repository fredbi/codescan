# Regression-testing strategy for the disentanglement refactor

Date: 2026-04-18 (amended 2026-04-19)
Status: DRAFT — pending execution
Branch: `refact/tests`
Baseline: `7e5296640...` (the last pre-refactor commit; current master)

## Problem statement

Since 7e52966 the scanner has been disentangled across ~110 files,
+16.7k / −9.5k lines. Tests pass on the refactor branch, but the suite was
reorganized concurrently with production changes, so "green" does not by
itself prove semantics-preserving behavior. Reading the diff is not enough.

We need an automated way to detect any observable behavior change in the
`*spec.*` objects produced by the scanner, compared against the pre-refactor
baseline.

## Approach (one sentence)

At every site in the existing test suite where a `go-openapi/spec` object is
produced (partial or full), marshal it to stable JSON and compare against a
golden file captured from master; mismatches surface every behavior change
for manual review.

## Harness design

### Capture / compare helper

One helper, two modes controlled by an env var:

```go
// internal/scantest/golden.go  (new)
func CompareOrDumpJSON(t *testing.T, got any, goldenName string) {
    t.Helper()
    path := filepath.Join(FixturesDir(), "..", "fixtures", "integration", "golden", goldenName)
    data, err := json.MarshalIndent(got, "", "  ")
    require.NoError(t, err)

    if os.Getenv("UPDATE_GOLDEN") == "1" {
        require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
        require.NoError(t, os.WriteFile(path, data, 0o644))
        t.Logf("wrote golden %s", goldenName)
        return
    }

    want, err := os.ReadFile(path)
    require.NoError(t, err, "missing golden %s — run with UPDATE_GOLDEN=1 to create", goldenName)
    assert.JSONEqT(t, string(want), string(data))
}
```

Key properties:
- `UPDATE_GOLDEN=1 go test ./...` captures goldens.
- `go test ./...` asserts against goldens.
- `encoding/json` sorts map keys by default → stable file content across runs.
- `assert.JSONEqT` compares JSON semantically, so minor formatting drift in
  the file (e.g. after manual edits) doesn't break the compare.

### Golden file naming — content-based, not test-name-based

The refactor reshuffled tests (moves, renames, splits), so baseline test
names ≠ refactor-branch test names. Naming goldens after `t.Name()` would
leave us with an orphan-rename problem on day one. Instead, name goldens
after **what they capture**, not **who captures them**:

```
fixtures/integration/golden/
  <fixture-bundle>_<object-kind>_<entity>.json
```

Examples:
- `petstore_spec.json`                           — full `*spec.Swagger` from running Run() on the petstore fixtures
- `classification_spec.json`                     — same, classification
- `classification_schema_NoModel.json`           — the `spec.Schema` built for `NoModel`
- `classification_response_ComplexerOne.json`    — the `spec.Response` built for `ComplexerOne`
- `classification_routes_orders.json`            — the `spec.PathItem` built for `/orders`
- `go118_schema_NamedWithType.json`              — the `spec.Schema` built for `NamedWithType` (go118 fixture bundle)
- `transparentalias_params_body.json`            — the body `spec.Parameter` built with `TransparentAliases: true`

Both baseline and refactor tests call `CompareOrDumpJSON(t, got, "<same-name>.json")`
for equivalent content. A test rename on the refactor branch does not
invalidate the golden; the name is what the *object* is, not what the *test*
is called. Caller picks the name.

Rule of thumb:
- One `CompareOrDumpJSON` call per top-level test function (or per iteration
  of a table-driven test where each case is meaningful on its own — then
  `<bundle>_<kind>_<entity>_<case>.json`).
- Subtests roll up into the parent's final captured state.
- What gets captured: whatever the test was already building
  (`*spec.Swagger`, `*spec.Schema`, `map[string]spec.Response`, etc.). The
  helper accepts `any`.

### What to exclude

- Tests that assert on error paths (no spec produced) — no golden.
- Pure regex / string / parser-internals tests — no golden.
- Any test that doesn't construct a `go-openapi/spec` object.

## Workflow

### Phase A — Capture from baseline (long-lived worktree)

1. `git worktree add .worktrees/baseline 7e52966`.
2. In the baseline worktree, add the `CompareOrDumpJSON` helper (adapted to
   the old single-package layout — drop the helper into the test package
   directly, no scantest yet).
3. Walk every `*_test.go` file that produces a `go-openapi/spec` object.
   For each such top-level test:
   - Inject a single `CompareOrDumpJSON(t, <built-object>, "<content-name>.json")`
     at the end of the test body (right before the final `}`), using the
     naming convention above.
   - Leave the existing assertions in place — they still run, and existing
     `assert.*`/`require.*` continue to guard behavior we already knew to
     check. The golden adds a *total-state* check on top.
4. Run `UPDATE_GOLDEN=1 go test -count=1 -p 1 -parallel 1 ./...` (serial, to
   freeze test-order-dependent state like the cached `classificationCtx`).
5. Copy the produced `golden/` tree out of the worktree:
   `cp -r .worktrees/baseline/<path-to>/golden /tmp/baseline-golden`.
6. **Keep the worktree.** It has three ongoing jobs:
   - Investigate any diff surfaced in Phase B (the baseline is the
     authoritative reference for "what did the old code actually do?").
   - Add exploratory test cases on the baseline when investigation suggests
     we're missing coverage, then re-capture the affected goldens.
   - Cherry-pick commits from master that landed *after* 7e52966 but before
     the refactor starts merging back — at the time of writing there is one
     such commit. Each cherry-pick may add tests/fixtures whose goldens we
     then need to capture and carry forward to the refactor branch.

### Phase B — Import goldens + adapt integration tests

1. `mkdir -p fixtures/integration/golden` on the refactor branch.
2. Copy `/tmp/baseline-golden/*` into `fixtures/integration/golden/`.
3. Add the `CompareOrDumpJSON` helper into `internal/scantest` (it can live
   there permanently — it's test infra, depends only on `scantest.FixturesDir`).
4. For each test on the refactor branch that builds an equivalent
   `go-openapi/spec` object, inject the same `scantest.CompareOrDumpJSON`
   call using the **same content-based name** as the baseline.

   On the test-name mismatch question: because we name goldens by *content*
   not by *test*, it doesn't matter that tests moved or were renamed — what
   matters is that for each fixture+object-kind+entity that the baseline
   captured, there is *some* test on the refactor branch that rebuilds it
   and calls `CompareOrDumpJSON` with the same name. If no such test
   exists, either add one, or mark the baseline golden as "not exercised"
   (keep the file, but no new-branch test references it — document in the
   commit).

5. Run `go test -count=1 -p 1 -parallel 1 ./...`. Every diff that surfaces
   is either:
   - a true regression → fix the refactored code.
   - an intentional improvement → update the golden file by hand, review
     the diff in the commit, mention it in the commit message.

   Intentional improvements are expected to be *very rare*. The refactor's
   goal was semantics preservation; at the time of writing Fred knows of
   one or two such changes (e.g. the `inferNames` bug we fixed while
   refactoring). Any diff beyond those should be treated as a regression
   until proven otherwise.

6. Repeat until green. Fred reviews at each round before updating any
   golden file.

### Phase C — Extend with go-swagger corpus

Two well-known golden tests exist upstream at go-swagger — and these are
the **last canary** for the codescan migration out of go-swagger into this
standalone repo. Post-migration, go-swagger kept only a smoke test that
exercises codescan via this corpus; if it passes, the migration is good.

- `/home/fred/src/github.com/go-swagger/go-swagger/cmd/swagger/commands/generate/spec_test.go`
  - fixture: `fixtures/goparsing/spec/api.go`
  - goldens: `api_spec_go111_ref.json`, `api_spec_go111_transparent.json` (for `TransparentAliases: true`).
- `spec_mod_test.go`
  - fixtures: `fixtures/bugs/3125/full`, `fixtures/bugs/3125/minimal`
  - goldens: colocated with fixtures.

Port these as separate integration tests in `internal/integration/` that:
- Copy the fixture files into this repo (don't reference across repos —
  keeps this repo self-contained).
- Call `codescan.Run(...)` with the matching Options.
- Use `CompareOrDumpJSON` with the go-swagger goldens copied into
  `fixtures/integration/golden/` under their existing names
  (e.g. `api_spec_go111_ref.json`, `api_spec_go111_transparent.json`).

If these pass, the refactor is good enough to reintegrate with go-swagger.

## Staging

Execute in slices, commit each:

| Stage | Scope |
|---|---|
| 1 | Phase A+B for petstore + classification fixtures only (widest existing coverage). Harness + ~5 goldens to prove the mechanics. |
| 2 | Phase A+B for all remaining test fixtures (go118, go119, go123 aliased/special, bookings, product, transparentalias, priority, etc.). |
| 3 | Phase C — go-swagger corpus (spec/api.go + bugs/3125). |

Each stage leaves the tree compiling and the golden suite green.

## Key decisions (locked in)

1. **UPDATE_GOLDEN env var pattern** over hand-porting test bodies.
   Why: one helper, minimal edits, same harness for capture & compare.

2. **Content-based golden naming**, not test-name-based.
   Why: refactor renamed/moved tests; naming by what's captured (fixture,
   object kind, entity) makes goldens survive test reshuffling and lets
   baseline↔refactor map by filename.

3. **One golden per top-level test function**, subtests roll up into the
   parent's final captured state.
   Why: simpler mapping, fewer files to reconcile, one dumpable object
   per test.

4. **No auto-acceptance mechanism (no `ACCEPTED.md`).**
   Why (Fred, 2026-04-18): we're not at CI yet. Every diff surfaced during
   review gets manual inspection; intentional changes update the golden
   file directly. Automating acceptance too early would launder bugs into
   "expected behavior." Revisit once the scanner is stable and CI starts.

5. **Serial test execution during capture and compare** (`-p 1 -parallel 1`).
   Why: the cached `scantest.classificationCtx`/`petstoreCtx` globals mean
   output can depend on test order. Serializing makes it deterministic.

6. **Baseline worktree is kept, not thrown away.**
   Why (Fred, 2026-04-19): it's the authoritative reference for "what did
   the old code do?" — used to investigate diffs and to add exploratory
   tests. Also the launchpad for cherry-picking post-baseline master
   commits so their tests/fixtures flow into this strategy too. Nothing is
   pushed from the worktree; it's a local investigation workspace.

## Operational notes

- Content-based naming sidesteps most rename issues. Occasionally you will
  find that a baseline golden has no equivalent on the refactor branch
  (coverage gap) — either add a refactor-branch test that rebuilds the
  same object, or document it as unreferenced.
- When diff review reveals a real regression: fix production code, rerun
  `go test` without UPDATE_GOLDEN; the golden file doesn't change.
- When diff review reveals an accepted behavior change: rerun the single
  test with `UPDATE_GOLDEN=1 go test -run '^TestX$' ./pkg` to refresh just
  that one file, then inspect the file diff in the commit.
- Humans will read these JSON diffs. Keep `MarshalIndent` with 2-space
  indentation and default Go map sorting — don't get clever with custom
  marshalers.

## What this strategy does *not* catch

- Performance regressions.
- Panics on inputs not in the test suite (fuzzing territory).
- Behavior of code paths no test currently exercises — the baseline cannot
  dump what it never built. Mitigated partially by the go-swagger corpus
  in Stage 3, but a gap remains.
- Non-JSON-observable changes (e.g. memory layout, logging). Out of scope —
  only `go-openapi/spec` objects are the contract we care about.

## Acceptance criteria

- Phase A+B completed for all existing tests that produce a spec object.
- `go test -count=1 -p 1 ./...` green on the refactor branch with no
  `UPDATE_GOLDEN`.
- Every accepted behavior change has its commit message explaining the
  *why* (linked to the test and the diff in the golden file).
- Phase C: at least one go-swagger corpus golden (api_spec_go111_ref.json)
  passes against the refactored scanner.

## Out of scope

- Adding new fixtures for coverage gaps — do that as a follow-up once the
  diff-against-baseline surface is clean.
- CI wiring / matrix configuration. This is a local-only investigation.
- Rewriting existing unit tests in terms of goldens. Unit tests stay as-is;
  goldens add a layer on top of them. The golden integration tests remain
  permanently in `internal/integration/` even after this refactor merges —
  they become the long-term regression guard for future changes.
