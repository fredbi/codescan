> [!NOTE]
> Last revision: 2026-08-16

# Benchmarks cleanup — one corpus, one document, one story

## Summary

Turn the accumulated benchmark apparatus into something reproducible by anyone who clones the repo, and
replace the pile of exploration prose with a single document that reports results.

Three moves:

1. The corpus ships **in the repo** (`internal/benchmarks/testdata/corpus.tgz`) and is unpacked on demand —
   no more `CODESCAN_BENCH_*_DIR` pointing at Fred's home directory, no more "go-swagger" corpus that was
   never an API.
2. The measurement matrix is stated once: two corpora × the loader configurations that matter (toolchain
   loader cold, toolchain loader warm + compiled export data, pure-Go loader).
3. Two stories get told with numbers: **the loader** (which one, and what it costs) and **the history**
   (v0.33.3 → v0.35.1 → master, six months of work seen from outside the module).

## Context

Branch: `cleanup-bench` (worktree `.worktrees/chore/cleanup-bench`), on top of a `temp` commit that already
moved `internal/integration/bench_*` → `internal/benchmarks/` and `hack/loader-benchmark` →
`internal/benchmarks/loader-benchmark`, and added the corpus archive. That commit leaves the package
**not compiling** (the env constants it dropped are still referenced) — first thing to fix.

What was wrong with the literature:

* The second corpus was **go-swagger itself** — a hand-written CLI tool, not an API, and nobody could
  reproduce where its figures came from.
* Three markdown files (`bench_scanner_axes.md`, `bench_dockerctl_report.md`,
  `bench_dockerctl_design_notes.md`) plus `loader-benchmark/README.md` document an exploration path
  (Axes A–H, estimates, "where the next win is") rather than results.
* Corpora were external checkouts named through the environment: absent for everyone but Fred.

Settled with Fred (2026-08-16):

* Keep the phase split, the profiling wrapper and the loader ladders; **drop Axes F/G/H**
  (`bench_discovery_test.go`, `bench_workingset_test.go`) — instrumentation for a closed stream.
* **Vendor kubeapi** into the archive so the corpus is offline-reproducible (+1.5 MB, 6.9 MB total).
* One document at `internal/benchmarks/README.md`; scanner README and the doc-site options page link to it.
* Leave `.claude/plans/on-demand-scanner.md` and `loader-vs-gopackages.md` alone — they are design records,
  not benchmark literature.

Corpus shapes (measured):

| corpus | what | `.go` files (own) | vendored | emits |
|---|---|---|---|---|
| `dockerctl` | go-swagger-generated **client** for the docker engine API | 501 | 1460 files | 198 defs / 0 paths |
| `kubeapi` | go-swagger-generated **server** for the kubernetes API | 2352 | 646 files | 222 defs / 260 paths |

## Trajectory

1. ✅ Corpus plumbing [🏁]
   * ✅ Re-roll `corpus.tgz` with `kubeapi/vendor`
   * ✅ A package that unpacks it on demand, idempotently, shared by the Go harness and `run.sh`
2. ✅ Fix and de-personalize the Go harness [🏁]
   * ✅ Make the package compile again; corpora resolve through the archive, never through the environment
   * ✅ Drop Axes F/G/H; rename the dockerctl-specific tests to run over both corpora
3. ✅ The cross-version harness [🛠️] [⚡]
   * ✅ `run.sh` unpacks the corpus itself, and builds one probe per historic version, not just one baseline
4. ✅ Measure [⚡]
   * ✅ The loader matrix (warm/cold × 2 corpora × 3 configurations)
   * ✅ The history matrix (v0.33.3, v0.35.1, current) on both corpora
5. ✅ One document [📚]
   * ✅ `internal/benchmarks/README.md`: corpus, loader story, history story, then how to run it
   * ✅ Delete the four exploration documents; re-point every reference
6. ✅ Doc-site page [📚]
   * ✅ `maintainers/performance.md` (w60): the version story + the loader options, four mermaid
     `xychart-beta` diagrams over kubeapi only. Verified by rendering in headless chromium —
     which caught a clipped axis label and a warm/cold chart whose outlier flattened everything
     else (split into two charts at different scales).
7. ✅ Review & land — **MERGED** via PR #118 (`27cc6295`), 2026-08-17
   * `9d12b4f2` harness · `ed742971` doc-site page · `717f109c` + `a84b6b1e` the two CI fixes

## Actions

### 1. Corpus plumbing

1. ⏳ **Re-roll the archive** — `go mod vendor` in kubeapi, repack sorted/normalized. Verified: the vendored
   tree emits the identical 222 defs / 260 paths under both loaders.
2. 📝 **`internal/benchmarks/corpus`** — `Ensure()` unpacks `testdata/corpus.tgz` into
   `testdata/corpus/{dockerctl,kubeapi}` when a content stamp says it is stale. Pure stdlib
   (`archive/tar` + `compress/gzip`), so it works on every platform the tests run on.
3. 📝 **A thin `unpack` command** over the same package, so `run.sh` uses one implementation rather than
   its own `tar` call.
4. 📝 **`.gitignore`** — the unpacked tree; and fix the two stale `hack/loader-benchmark/` entries.

### 2. Go harness

1. 📝 **Compile again** — the `temp` commit dropped `envBenchEnabled` / `envDockerctlDir` /
   `envGoSwaggerDir` while they were still referenced.
2. 📝 **Corpora from the archive** — `knownCorpora()` returns the two unpacked trees; absence is no longer a
   normal case, so the skip apparatus goes with it. `CODESCAN_BENCH=1` still gates the long reports.
3. 📝 **Delete** `bench_discovery_test.go`, `bench_workingset_test.go`.
4. 📝 **Rename** `bench_dockerctl_test.go` → `bench_scan_test.go`, `bench_dockerctl_profile_test.go` →
   `bench_profile_test.go`; both drive either corpus.
5. 📝 **Purge** every "go-swagger corpus" reference in the surviving prose.

### 3. Cross-version harness

1. 📝 **`run.sh` unpacks the corpus** and drops `CODESCAN_BENCH_*_DIR` (kept only as an escape hatch for
   measuring somebody else's tree).
2. 📝 **Fix the build path** — it still says `./hack/loader-benchmark`.
3. 📝 **History mode** — `CODESCAN_BENCH_HISTORY="v0.33.3 v0.35.1"` builds one throwaway module per
   version in `.work/<version>` with `GOWORK=off`, each measured in the default configuration.
   De-risked: both versions build against the current probe source and scan both corpora, same shape.
4. 📝 **Master's default moved** — v0.36.4 flipped compiled dependencies on. The history table therefore
   carries master twice (source deps for the like-for-like line, default for what a user gets).

### 4. Measurement

1. 📝 Warm matrix, n=3 alternating, on a **quiet machine** — wall clock moved 3× under load during
   de-risking, while allocation and peak RSS did not budge.
2. 📝 Cold pass (private empty `GOCACHE`, `GOMODCACHE` untouched), n=1.
3. 📝 History pass over both corpora.

### 5. Documentation

1. 📝 Write `internal/benchmarks/README.md` — results first, method last.
2. 📝 Delete `bench_scanner_axes.md`, `bench_dockerctl_report.md`, `bench_dockerctl_design_notes.md`,
   `loader-benchmark/README.md`.
3. 📝 Re-point `internal/scanner/README.md` §loader and `docs/doc-site/usage/options-reference.md`.
4. 📝 Add `internal/benchmarks` to the package layout table in the project instructions.

## Achievements

Everything is staged on `cleanup-bench`, uncommitted, awaiting review. `go test work ./...` and
`golangci-lint run` are clean; 1183 lines added, 1913 deleted.

### 1–3. Plumbing and harness ⭐⭐

* `internal/benchmarks/corpus` — `Ensure()` unpacks the archive under `testdata/corpus/`, stamped
  with the archive's SHA-256 so a re-rolled archive is never measured stale. Extraction refuses a
  traversing entry rather than clamping it, and is covered against a synthetic archive so CI does
  not write 56 MB per job. `corpus/unpack` is the same code as a command, which is how `run.sh`
  avoids a second copy of the layout.
* **The `go.work` gotcha bit again, in a new place.** The corpus now lives *inside* the repo, so the
  root workspace claims it and every load failed with `directory prefix . does not contain modules
  listed in go.work`. Handled at all three entry points (`Options.GOWORK`, the raw loads' `Env`, and
  a `-gowork` flag on the probe that goes through the environment, since `Options.GOWORK` is younger
  than v0.33.3).
* Every report is now behind `CODESCAN_BENCH=1`, including the phase split — CI runs
  `go test work ./...` with `-race` on six job combinations, and an 800 MB scan has no business
  there.

### 4. Measurements ⭐⭐⭐

Taken on an idle machine (Ryzen 7 5800X, 31 GB, go1.26.5), n=3 warm / n=1 cold.

* **Six months, end to end, on kubeapi: 7.087 s → 0.970 s and 4555 MB → 447 MB**, for the identical
  222 definitions and 260 paths.
* **The grammar migration is witnessed, not inferred.** Two sub-patterns of the same corpus settle
  it: `./models/...` (nothing to emit) allocates 764.6 MB under v0.33.3 and 764.1 MB under current —
  the loading half is unchanged, as it must be. `./restapi/...` (the whole document) allocates
  4475.6 MB against 1215.8 MB. The whole 3.3 GB lives in the phase that reads annotations, and only
  where route bodies do.
* **No configuration wins both cache states.** Compiled dependencies: −55% warm, 6× slower cold and
  231 MB of build cache. Pure-Go loader: 4 KB of cache, cold time equal to warm, −45% memory flat.
* Wall clock needed the idle machine — under load it moved 3× while allocation and peak RSS
  reproduced to four digits. Said so in the document.

### 5. The document ⭐⭐

`internal/benchmarks/README.md`, results first and method last, replacing 923 lines across four
files that documented an exploration path. `internal/scanner/README.md#loader` and the doc-site
options page now cite it.

### 6. What CI caught after review ⭐⭐

Both were in the extractor, and both were the kind only a second platform or a static analyser
finds:

* **A guard with a platform-dependent verdict.** `filepath.IsAbs("/x")` is false on Windows (it
  wants a volume), so the rooted entry Linux refused was accepted there — landing *inside* the
  destination, so nothing escaped, but a security check that answers differently per host is wrong
  whatever it lets through. Entry names are now judged in the tar slash form, with `\` and `:`
  refused everywhere.
* **The measurement gate had a hole.** `Find` resolved the archive before validating the name, so
  the ungated test covering an unknown name unpacked 55 MB on every CI job of every platform.
  Validating first fixes the gate and the behaviour.
* CodeQL's zip-slip alert took the containment check restated on the joined path — its own
  documented sanitizer shape — kept as redundant belt-and-braces.

## Appendix — risks and open points

* **`corpus.tgz` in git.** 6.9 MB after vendoring. Re-rolling it later (a newer dockerctl, a newer
  kubeapi) rewrites a large blob; worth doing rarely and deliberately.
* **kubeapi's `go.mod` says `go 1.26.5`** — a patch-level directive, which is against house style but is
  generated corpus, not our code. It does impose a ≥1.26.5 toolchain on anyone running the benchmarks.
* **`go.work` interference.** Every throwaway module must be built with `GOWORK=off`, or the workspace
  substitutes the working tree for the released library and the history table silently measures master
  three times.
* **Cold pass cost.** Compiled dependencies on a cold cache took 12–16 s per cell during de-risking and
  writes ~230 MB of build cache. That is the finding, not an accident.
