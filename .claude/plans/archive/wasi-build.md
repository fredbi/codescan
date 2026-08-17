> [!NOTE]
> Last revision: 2026-08-01

# WASI build & the package-loading seam

## Summary

Make codescan run with **no Go toolchain and no `exec`** — first as a releasable WASI artifact, ultimately as a
browser playground on the doc site.

The spike (branch `wasi-build`, worktree `.worktrees/experimental/wasi-build`) settled the feasibility question
empirically: codescan already cross-compiles to `wasip1/wasm`, the filesystem already works there, and the *only*
runtime blocker is `packages.Load` exec'ing `go list`. Replacing it with **`internal/packages`**, our own loader,
reproduces `packages.Load` **byte-for-byte across all 143 fixture patterns** — natively, through a virtualized
`fs.FS`, and inside a WASI guest.

Beyond the playground, this makes the library immune to the "which Go version is installed" class of environment bugs
we keep fighting (see `textmarshaler-toolchain-independence.md`).

Parent vision: `wasm-playground.md` (Stream 6). This document supersedes its §4 `Loader`-seam and Phase-2
`NewFromBytes` sketches.

## Context

Five measurements reframed the problem. They were taken with a throwaway probe harness that has since been dropped;
what survives of it is `internal/packages` itself and the tests around it.

1. **There is no build problem.** `GOOS=wasip1` and `GOOS=js` builds of the root module both succeed on master today.
   `os/exec` compiles for wasm; it only fails at runtime. The one real blocker was two missing `/go.mod` hashes in
   `go.sum` (`swag/loading`, `swag/stringutils`), which break *any* cross-compile.

2. **WASI gives us the filesystem for free.** `os.ReadDir` / `os.ReadFile` / `go/parser` behave identically to native
   under `wasmtime` preopens. The "`go/packages` reads through `os`, not an injected `fs.FS`" concern does not apply:
   the FS is injected at the *host* level. This is the decisive advantage of WASI over `GOOS=js` and validates the
   pivot.

3. **Exec is the only runtime blocker.** Stock `codescan.Run` under WASI fails with exactly one error —
   `pipe: Not implemented on wasip1` — from the `go list` subprocess. Nothing else fails.

4. **No export data is in play, and from-source is fast.** Under codescan's load mode
   (`NeedDeps|NeedSyntax|NeedTypesInfo`), `packages.Load` type-checks **209 of 210** packages from source (190 of them
   stdlib, 1195 `.go` files) in **553 ms**, with `ExportFile` empty for every package. So our loader needs **no
   `gcexportdata` path at all** — correcting an earlier assumption in `wasm-playground.md` §5b.

5. **Our `go/packages` surface is tiny.** Production code uses only `Load`, `Config`, `Package`, `Error`, `ListError`
   and 7 `Need*` constants, across 4 files, touching 8 `Package` fields and just two `types.Info` maps (`Defs`,
   `Types`).

### Why not the two obvious alternatives

- ⛔ **Fork `x/tools/go/packages`.** Assessed on `.worktrees/experimental/packages-load` (a full vendored copy builds
  as an independent module). Too many sub-packages and transitive deps to maintain; stripping it down is real work and
  the result still carries permanent doubt about tracking future Go releases. Rejected.
- ⛔ **In-process `GOPACKAGESDRIVER` hook.** The driver seam is exec-only: `type driver func(...)` at
  `external.go:83` is unexported and `findExternalDriver` is the sole entry point. Confirmed unchanged at x/tools tip
  (`1f62bc851`). There is no public func-based hook and none in prospect.

The `DriverRequest`/`DriverResponse` protocol remains the right **contract** — it is documented, stable and
version-independent. What the spike shows is that the expensive half of `go/packages` is the half that *produces* a
response (`go list` invocation and parsing, overlay marshalling, chunking, cgo, test synthesis). The half that
*consumes* one is small, and it is the only half we need.

Note: `Overlay` is explicitly **not** the answer. It is documented as slow, requires an in-memory map of every source
file, and works against the demand-driven scanner (`ramblings/index-builder-statefulness.md`).

## Trajectory

1. **Own the loading seam** — `internal/packages`, one loader, every platform
   > `packages.Load` is all-or-nothing by construction, so it can never do demand-driven parsing. Owning the loader
   > is the same seam Stream 8's `DemandLoader` needs. The golden corpus is the equivalence proof.
   >
   > Shape settled: entry point mirrors upstream (`NewLoader(...Option).Load(cfg, patterns...)`), vocabulary aliased
   > to `x/tools/go/packages` so a caller can switch back by changing the call and nothing else. No driver protocol,
   > no JSON, no `Overlay` — an `fs.FS` one level lower instead.

   1. ✅ 🧪 Loader seam in the scanner, replacing the direct `packages.Load` call
   2. ✅ 🧪 `internal/packages`: patterns → dirs → build-constraint selection → parse → type-check
   3. ✅ 🧪 `WithFS` option virtualizing the source tree
   4. ✅ 🏁 Independent test suite for the package (build tags, targets, patterns, intra-module imports)
   5. ✅ `WithTarget(goos, goarch)` + `Options.GOOS` / `Options.GOARCH`, honoured by both loader paths
   6. ✅ `Options.FS` bubbles the `fs.FS` seam to the public API; a non-nil FS selects our loader
   7. ✅ `StubStdlib` — leave the standard library out of the graph, synthesize its types from usage
   8. ✅ `StdlibExportData` — read the standard library from precomputed export data: truncation's speed and
      footprint, full type fidelity
   9. 🔍 Truncation: needs its own investigation (how stdlib types render title/description when stubbed)

2. **Ship a WASI artifact**
   > First milestone: a releasable `.wasm` demonstrating codescan no longer needs a Go compiler installed.

   1. ✅ `cmd/genspec` — headless spec generator, zero deps beyond the library, cross-compiles to wasip1
   2. ✅ 🏁 Integration test running the artifact under a WASI runtime, zero exec
   3. ✅ Self-contained build — `stdlibdata` tag embeds the export data; no GOROOT, no toolchain, nothing
      mounted but the project
   4. ⛔ `wasip1/wasm` artifact in the release pipeline — **parked** (Fred, 2026-08-02): releasing artifacts
      means configuring goreleaser, and that stream is led elsewhere
   5. ⛔ Install a WASI runtime in CI so the test stops skipping there — **parked**, rides with the release
      wiring above
   6. ✅ 📚 `cmd/genspec/README.md` — WASI build, runtime invocation, what to mount, self-contained build

3. ✅ **Browser execution — reached 2026-08-02.** See the *Milestone 2* section below
   1. ✅ Runtime chosen and proven: `@bjorn3/browser_wasi_shim`, `hack/browser` (`634404c`)
   2. ✅ Sources materialised into the guest filesystem; scan runs off the main thread — the
      `hack/doc-site/genspec-wasi` scaffold (`36afec0`)
   3. ⏳ 🔍 **How the standard library reaches the guest.** The app currently runs `-stub-stdlib`, which is
      fast and needs no asset but degrades — and degrades third-party imports too, which matters where the
      examples use `strfmt`. Deciding this is the next blocker; it is one argument in `src/lib/flags.ts` plus a
      published asset.
   4. ✅ Upload of whole directories, options overlay, panes, re-scan — **MVP works** (Fred, 2026-08-02:
      "minimal UI, no chrome at all, but it works"). Two bugs found on first use and fixed: genspec
      exposed only 2 of 19 boolean options (`e0f3cbf`), and opening a module merged into the tree
      instead of replacing it (`1df88c7`).
   5. 🔍 Decide the result contract: keep the CLI shape, or return structured output with diagnostics
   6. 🔍 Dependency story for uploaded projects — the last remaining gap now that the stdlib is solved
   7. 📝 📚 Hugo shortcode + example pre-fill (doc-site stream owns it; follow `{{< example >}}`)

4. **Doc-site playground** [📚]
   1. 📝 Serve the WASI + JS assets statically (Stream 7 / `D-playground`)

5. 😇 **Environment immunity** (falls out of 1)
   1. 🔍 Retire toolchain-version CI guards once the loader owns type-checking

## Actions

### To leave prototype status

1. 🔍 **Decide how annotated dependencies reach a browser.** A vendored tree uploaded with the
   project is the only mechanism that carries comments. Our resolver already reads `vendor/`, so the
   work is UX and packaging, not scanning.
2. 🔍 **Own the loader deliberately, or do not.** `internal/packages` is currently a second
   implementation living beside `packages.Load`, chosen per call by `Options.FS`. Either it becomes
   the one loader — which wants real-project scans first — or the duplication is a standing cost.
3. 🔍 **Settle the declaration contract** (`on-demand-scanner.md`). Two corpus cases fail without it,
   and the on-demand work needs it regardless.
4. 📝 🛠️ **Publish the export-data artefact**, regenerated when the toolchain moves. Parked with the
   release wiring.
5. 📝 🎨 **Polish the front-end.** It is a scaffold: no highlighting, no cross-references, no chrome.
   A component library would get it closer to the TUI for less code than hand-rolling.

### Consolidated

1. ⏳ **Adopt `swag/fs` once it lands as `go-openapi/swag/fs`** — not our item to schedule. codescan is a new use
   case for that package, still experimental in `fredbi/core/swag/fs` and due for promotion. What our loader needs
   from it, as input to that design:

   - `go/build.Context`'s hooks are the real client, and they want more than `fs.FS` offers: `OpenFile`
     (`io.ReadCloser`), `ReadDir` returning **`[]fs.FileInfo`** rather than `fs.DirEntry`, `IsDir`, plus the path
     algebra `JoinPath` / `IsAbsPath` / `HasSubdir`.
   - The impedance mismatch is paths, not I/O: `io/fs` is slash-separated, relative and unrooted, while `go/build`
     and `go/packages` deal in absolute OS paths. Our `vfs.clean` normalises one onto the other.
   - `NewReadOnlyOsFS` would collapse the loader's `fsys == nil ? os : fs.FS` dual path into a single
     always-`fs.FS` path — the simplification worth having.
   - `OverlayFS` is directly interesting for the playground tier (a user's edits over a base tree).

2. 📝 🏁 **Extend the package's tests to module layouts.** Done: build-tag matrix, target matrix and precedence,
   recursive patterns, intra-module imports, the function-body regression. Still open: `replace` directives, vendor,
   and module-cache resolution — those need a tree on disk, not an `fstest.MapFS`.
3. ⏳ 🧪 **Promote the loader seam.** `scanner.LoadPackages` is a package-level `var` (spike shortcut). Decide its real
   shape: `Options` field vs. build-tagged constructor. Must not widen the public API — see `feedback_test_api_surface`.
4. 📝 **Land the `go.sum` fix** — two missing `/go.mod` hashes. Standalone, independently useful, unblocks all
   cross-compilation.
5. ⏳ **Recognizer ordering — merged, handover requested.** `fix/strfmt-dispatch-symmetry` merged cleanly
   (`89f6064`): it reorders `buildFieldAlias` in the parameters and responses builders, this branch reorders
   `buildNamedField` in the same two files — the same idea from opposite ends, neither undoing the other.
   Pass 2 should move the `buildNamedField` hoist into the quirk stream and drop this branch's copy, so both halves
   of the symmetry are audited together. See the handover request below.
6. ♥️ **More identity recognizers for common stdlib shapes** — the cheap way to move 133/138 up. `io.Reader` and
   friends are the obvious candidates, since an interface has no identity recognizer today.
7. 🔍 **Two stdlib shapes still unreachable when stubbed.** `io.Reader` (and stdlib interfaces generally) have no
   identity recognizer, so a stubbed scan errors rather than degrading; `encoding.TextMarshaler` detection needs a
   real method set. Options: give interfaces an identity recognizer, or degrade to a diagnostic instead of failing
   the scan. The schema builder's `resolveRefOrErr` has no stdlib-specials pre-check, which is the same ordering
   shape the parameters and responses builders just lost — worth auditing, but it changes native behaviour so it
   needs its own corpus pass.
8. 🔍 **Truncation investigation.** Largely answered above; what remains is policy. The open question is fidelity of *rendered output* for
   stubbed stdlib types — title/description/format on `time.Time` and friends — not just whether the scan completes.

### Next — the releasable artifact

6. 📝 🏁 **Runtime choice.** Spike used `wasmtime 41`. Needs a decision for CI: wasmtime (CLI, proven here) vs.
   `wazero` (pure Go, embeddable in `go test` with no external binary — likely better for a hermetic integration test).
7. 📝 🛠️ **Release wiring** — `wasip1/wasm` artifact alongside the binaries already planned for release.

### Later

8. 🔍 **Retire `x/tools/go/packages` from the runtime path.** Currently kept for its type vocabulary (aliased) and as
   the default `LoadPackages`. Switching the default over is a decision, not a task — it wants real-project scans
   first (Appendix 1).
9. 🔍 **Truncation depth as a knob.** See Achievements — depth is a payload/fidelity dial, and the corpus measures it.

## Achievements

### 1. `internal/packages` — parity reached ⭐⭐⭐

- ✅ **138/138 fixture patterns byte-identical to `packages.Load`**, with no `go list` and no exec — through the
  real filesystem *and* through `WithFS(os.DirFS("/"))`, same score both ways.
- ✅ **Identical spec running inside a WASI guest** (`wasmtime`, petstore). 7.3 s vs 839 ms native — the gap is
  per-file I/O through the WASI shim, which is exactly what truncation would attack.
- ✅ **Build constraints resolve in pure Go** via `go/build`: `//go:build` expressions (negation, `&&`),
  GOOS/GOARCH filename suffixes, and `-tags`. No reimplementation needed — this was the biggest unknown in the
  first draft of this plan and it turned out to be free.
- ✅ **`build.Context` *is* the `fs.FS` seam.** It already exposes `OpenFile`, `ReadDir`, `IsDir`, `JoinPath`,
  `IsAbsPath`, `HasSubdir` as injectable func fields, so `WithFS` wires straight into constraint matching too.
- ✅ **Hermetic test suite** — 6 tests / 12 subtests over an `fstest.MapFS`, so the package is exercised with no
  toolchain, no disk and no network, which also makes every one of them a `WithFS` test.
- ✅ **End-to-end virtual scan** — `internal/integration/virtual_fs_test.go` runs `codescan.Run` against a module
  that exists only in memory: models discovered, title/required/validations intact, and build tags honoured inside
  the virtual tree. This is the property the WASI build and the browser playground both rest on.
- ✅ **`Options.FS` selects the loader.** Not a separate flag: `packages.Load` reaches the filesystem through
  `go list`, so asking for a virtual FS is already a statement about which loader must run.
- ✅ Resolver covers main module, `replace` directives, vendor, module cache (Go 1.17+ go.mod lists all relevant
  deps, so no MVS walk needed) and GOROOT.
- ❌ **`IgnoreFuncBodies` is NOT safe** — corpus caught it. `classification/operations/responses.go` declares a
  `swagger:response` *inside a function body*; skipping bodies leaves no `TypesInfo.Defs` entry and the response
  vanishes. Now explicitly false, with the reason recorded at the call site. This was also the true cause of the
  `classification` divergence in the graph-replay run below.

### 2. Export data — the standard library, precomputed ⭐⭐⭐

The premise that host FS I/O was the bottleneck did **not** survive measurement: syscall accounting puts it at
**1.8%** of a full WASI scan (0.149 s of 8.14 s; native 0.071 s of 1.32 s). The cost is parsing and type-checking
190 stdlib packages, amplified ~5–6× by wasm. So embedding the stdlib *source* would have bought back ~1%.

What the compiler already computed is another matter. `Options.StdlibExportData` + `hack/genexportdata`:

| petstore, wasmtime | time | peak RSS | corpus |
|---|---|---|---|
| stdlib from source | 7.30 s | 681 MB | 143/143 |
| **export data** | **1.00 s** | **138 MB** | **141/143** |
| stubbed | 0.98 s | 147 MB | 138/143 |

- ✅ **Level with truncation on both time and memory, without its fidelity loss.** Every case `StubStdlib` degrades
  — `text-marshal`, `raw-message-override`, `wrapper-decl-type-override`, `response-edges` — is byte-identical
  here, because the types are the compiler's: real fields, real method sets, real interface identity.
- ✅ **GOROOT never mounted.** 9.3 MB for the whole stdlib, **4.2 MB compressed** — embeddable.
- ⏭️ **The two remaining gaps are a new class: types without declarations — DEFERRED to the on-demand-scanner
  stream** (`on-demand-scanner.md`), decided 2026-08-02. A builder must not bypass the scanner's index to work
  around it: resolving a type to a declaration is the scanner's job, and the same contract is owed to the
  on-demand work regardless. Export data carries no syntax and no
  comments, so a builder reaching for the declaring source finds nothing. `in-case-insensitive` errors on
  `io.Reader` (Q37/Q38 again); `go123` renders a stdlib named type used as a field as `{}` where a full graph
  emits a `$ref` to a definition built from the stdlib's own godoc. Worth noting the reference behaviour there is
  itself questionable — it is Q38(b), stdlib types leaking into the user's spec.

### 3. Stdlib truncation — measured, and it decides the browser tier ⭐⭐⭐

Petstore, wall clock and peak RSS, full graph vs `StubStdlib`:

| runtime | full stdlib | stubbed | speedup | peak RSS full | peak RSS stubbed |
|---|---|---|---|---|---|
| native | 1.32 s | **0.18 s** | 7.3× | 346 MB | **53 MB** |
| wasmtime | 8.14 s | **0.95 s** | 8.6× | 848 MB | **143 MB** |
| wazero | 24.01 s | **6.13 s** | 3.9× | 2.5 GB | **347 MB** |

Larger fixture (classification) under wasmtime: 8.76 s / 1.08 GB → **1.07 s / 152 MB**.

- ⚡ **Memory is the deciding number, not time.** ~850 MB peak under wasmtime for a *fixture* is not something a
  browser tab can host; 143 MB is. Time merely follows.
- ✅ **Mounting tiers** (petstore, wasmtime) — the real choice is how much of the host to expose to the guest:

  | mounted | mode | time | peak RSS | result |
  |---|---|---|---|---|
  | GOROOT + module cache | full graph | 8.1 s | 848 MB | byte-identical |
  | module cache only | `-stub-stdlib` | 1.0 s | 143 MB | byte-identical |
  | project tree only | `-stub-stdlib` | 0.1 s | 123 MB | one `date-time` format lost |

  The last row matters: an unresolvable import is synthesized rather than fatal, so a scan completes with
  **nothing but the project tree** — no Go installation, no module cache download.

- ✅ **The loss is no longer quiet.** Every synthesized import path raises one `scan.synthesized-import`
  diagnostic on the import that caused it — Hint when `StubStdlib` withheld it (the caller asked), Warning when the
  import is merely absent (usually a mounting or module-cache mistake). Before this, a synthesized type used in a
  field position produced *nothing at all*; the only signal was the downstream wreckage of a value-position use,
  reading as `cannot convert err to type log.Fatal` — an error in the scanned code rather than a missing dependency.
- 🐛 Fixed alongside: a recursive pattern over a virtual filesystem handed back io/fs's own rooted paths unmapped,
  so file names — and every `token.Position` derived from them — lost their leading separator.
- ✅ **Fidelity cost, corpus-wide: 133/138 byte-identical.** Three differ structurally (`time.Duration` no longer an
  integer, `json.RawMessage` no longer a byte array, one type override); two fail outright — stdlib *interfaces*
  (`io.Reader`) which, unlike `time.Time`, have no identity recognizer to fall back on.
- ✅ Full-stdlib mode remains **138/138**, so the option costs nothing when unused.
- ✅ The recognizer-ordering change is no longer untestable: truncation is what tests it. Hoisting the identity
  recognizers above the declaration lookup in the parameters and responses builders converted four hard failures
  into byte-identical scans, and leaves the corpus unchanged with a full graph.

### 4. WASI artifact — milestone reached, and now self-contained ⭐⭐⭐

- ✅ **`stdlibdata` build tag embeds the export data**, as a zip (`archive/zip`'s reader is already an `fs.FS`, so
  one file in the binary still reads one package at a time). Opt-in: 4.7 MB most builds have no use for.
- ✅ **Nothing mounted but the project tree, no GOROOT named anywhere** — 1.1 s / 150 MB on the petstore under
  wasmtime, byte-identical to a `go/packages` scan. Guarded by `TestWASIArtifactIsSelfContained`.
- ℹ️ Artifact 20 MB raw, **8.5 MB gzipped** (against 15 MB / 3.7 MB without). Embedding costs ~0.05 s and ~10 MB
  over reading the same data from a mounted directory — close to free.
- ⚠️ Measurement note: the very first run of a freshly written 20 MB module read 2.37 s / 585 MB. It does not
  reproduce (four consecutive runs: 1.06–1.11 s / ~150 MB) — cold cache plus wasmtime compiling the module.
  Discard first-run figures for large artifacts.

- ✅ **`cmd/genspec` runs under wazero and produces a byte-identical spec** to an in-process scan of the same
  fixtures. The guest has no process model, so nothing it did could have reached `go list`.
- ✅ **`-loader=auto`** picks the toolchain-free loader wherever the build cannot start a subprocess (every wasm
  target), so the guest never attempts the go command it could not run.
- ✅ **Natively toolchain-free too**: `-loader=own` with `PATH=/nonexistent` produces the identical spec, while
  `-loader=go` fails with `go command required, not found`. The control and the claim in one pair of runs.
- ✅ Test skips cleanly with no runtime on PATH and under `-short`; drives wazero or wasmtime, whichever is found.
- ℹ️ **Artifact**: 15 MB raw, **3.7 MB gzipped**.
- ⚡ **Runtime cost**: wazero ~21 s, wasmtime ~7 s for the petstore — wazero's compile cache (`-cachedir`) does not
  close the gap, so it is execution speed, not compilation. Both dwarf the ~0.8 s native scan; the cost is per-file
  I/O through the WASI layer against ~1200 stdlib files.
- ⚠️ **The guest still needs GOROOT and the module cache by path.** Nothing in a WASI environment can ask the go
  command where they are, so they must be passed in and mounted. This is precisely the bill that truncation
  (Trajectory 1.7) would pay: no stdlib in the graph means no GOROOT to mount.

### 5. Graph replay — spike complete ⭐⭐⭐

- ✅ **Byte-identical spec from a recorded graph, natively.** Full graph (210 pkgs / 1195 files) → our loader →
  `diff` clean against `codescan.Run` on the petstore.
- ✅ **Byte-identical spec under WASI, zero exec** (the gate). Same fixture, `wasmtime`, 794 ms, while stock
  `codescan.Run` fails on `pipe: Not implemented` in the same binary. ⭐⭐⭐
- ✅ **Truncation works and is a large win.** Dropping stdlib from the graph and synthesizing absent imports
  *from usage* (opaque named types carrying the right package/name) still produced an identical spec:

  | graph | packages | files | metadata | scan time |
  |---|---|---|---|---|
  | full | 210 | 1195 | 163 KB | 727 ms (native `packages.Load`) |
  | depth-1 stdlib | 55 | 398 | 49 KB | — |
  | stdlib dropped | 19 | 79 | 17 KB | **145 ms** |

- ✅ **Corpus-wide A/B** over 138 fixture patterns (all of `goparsing/` + every `enhancements/` tree), measured with
  the throwaway probe harness (since dropped):
  - stdlib fully dropped → **131/138 identical**
  - stdlib depth-1 kept → **136/138 identical**
- ✅ **Divergences are one coherent class**, not noise: code that drills *structurally* into stdlib types —
  method sets (`types.Implements` for `IsTextMarshaler`), `json.RawMessage` as `[]byte`, `time.Duration` as `int64`,
  `io.Reader` identity. Name-keyed recognition is unaffected, exactly as predicted.
- ✅ Full suite green (19 packages, 0 failures) with the seam and the hoist in place.

### Corrections made during the spike

- ❌ I asserted from-source stdlib type-checking was unviable and export data mandatory. **Wrong** — that was a
  pathological hand-rolled importer in an early probe. `packages.Load` does exactly that, from source, in 553 ms.
  The plan above reflects the corrected finding.

## Milestone 2 — what the browser needs

Planning only; nothing built. The point of this section is to name the pieces and the decisions before writing
any JavaScript.

### What the artifact actually demands of a runtime

Read straight out of the compiled module's import section, so this is a specification rather than a guess:

- **21 unique `wasi_snapshot_preview1` functions**, and nothing else — no `env` module, no Go-specific glue.
  This is the sharpest difference from `GOOS=js`, which needs `wasm_exec.js` and a bespoke `go` import module.
  Any conforming preview-1 shim can host the artifact.

  ```text
  args_get  args_sizes_get  clock_time_get  environ_get  environ_sizes_get
  fd_close  fd_fdstat_get  fd_fdstat_set_flags  fd_filestat_get
  fd_prestat_dir_name  fd_prestat_get  fd_read  fd_readdir  fd_write
  path_filestat_get  path_open  path_readlink
  poll_oneoff  proc_exit  random_get  sched_yield
  ```

  **Every one of them reads.** It was 24 until the `go list` loader was build-tagged out of WebAssembly builds
  (`e7b2f8b`): `path_create_directory`, `path_remove_directory` and `path_unlink_file` came in through the
  temporary files `go list` needs, not through anything a scan does. Now pinned by
  `TestWASIArtifactImportsOnlyReads` (`a2a7ea2`), which also rejects a second import module and any non-function
  import — it needs a toolchain but no WASI runtime, so it runs anywhere.

  ⛔ **Going further is not worth it, and partly not possible.** `os/exec` is still linked, but on wasip1 it is a
  stub that adds no WASI import, so it costs bytes and nothing else. It arrives through `go/build` *and* through
  `gcexportdata` itself (`gcexportdata` → `os/exec`, and → `gcimporter` → `go/build`), so it cannot leave while
  we read export data — the decoders have no stdlib equivalent, `go/internal/gcimporter` being internal.
  Dropping `go/build` from our own code would save a measured 0.55 MB but gcexportdata drags it back. The
  remaining `go/packages` residue is the type aliases, whose whole point is switch-back symmetry.

- **No imported memory** — the module owns its own. So no `SharedArrayBuffer`, and therefore **no COOP/COEP
  headers**, which is what makes plain static hosting (GitHub Pages) viable at all.
- **No `sock_*`** — nothing to stub for networking.
- ✅ **`poll_oneoff` — checked, and it is not a problem** (`634404c`, `hack/browser`). Go reaches it only through
  the netpoller, and only once the scheduler runs out of work, which a single-goroutine scan does not do: **zero
  calls** across workloads from 1 to 200 packages. When it does fire (longer runs under wazero show 9–14), the
  only shape Go asks for is the single clock subscription `netpollinit` registers, which `browser_wasi_shim`
  handles. It implements it by *spinning* rather than yielding, which is a reason to keep the scan in a worker,
  not a reason to reject the shim.

### The pieces to assemble

1. ✅ **A preview-1 shim — `@bjorn3/browser_wasi_shim`**, 114 KB and zero dependencies, verified running the real
   artifact. `@tegmentum/wasi-polyfill` is out: 6.3 MB, v0.1.1, and its OPFS/IndexedDB backends sit on the
   **wasip2** (Component Model) path, which a Go `wasip1` core module cannot use.
2. **A guest filesystem**, populated from JavaScript before the run. Note this is a *different layer* from
   `swag/fs`: the shim's filesystem is what the guest's `os` package sees; `Options.FS` composes trees inside Go.
   Either can carry the user's sources; they should not both try to.
3. **Source intake**, cheapest tier first: an editor (needs a synthesized `go.mod`, since the resolver maps import
   paths through one), then a directory picker (`webkitdirectory`), then an archive.
4. **An execution shape**: a Web Worker, since even a one-second scan must not freeze the page.
5. **A result contract** — see the decisions below.

### Decisions to take before building

- ✅ **Which shim** — settled, see above.
- ⛔ **IndexedDB cannot back the guest filesystem.** `fd_read` is a synchronous host call and IndexedDB has no
  synchronous API; bridging needs `Atomics.wait` + `SharedArrayBuffer`, hence COOP/COEP, hence no plain static
  hosting. If persistence is wanted the mechanism is **OPFS sync access handles** (worker-only, genuinely
  synchronous, no special headers) — `browser_wasi_shim` already ships `fs_opfs.js`. In-memory is enough for the
  MVP and is the same shape OPFS slots behind later.
- 🔍 **How the standard-library types get there.** Two now-real options, roughly equal in bytes:
  *embedded* (`stdlibdata` tag — one 8.5 MB compressed fetch, simplest) versus *fetched into the guest
  filesystem* (3.7 MB artifact + 4.2 MB data, independently cacheable, and lets a "no stdlib" quick mode ship
  without it). The embed exists only because a WASI guest had no other way in; a browser does.
- 🔍 **The result contract.** `genspec` already speaks argv → stdout/stderr, which is exactly what a shim offers,
  so the first cut needs no new Go API. But diagnostics currently arrive as text on stderr, and a playground
  wants them structured and positioned. Decide whether to keep the CLI shape and parse, or add a JSON result
  mode carrying diagnostics alongside the document.
- 🔍 **Third-party dependencies** — the only remaining gap now the standard library is solved. A vendored tree
  works; anything else degrades to synthesis plus `scan.synthesized-import`. Whether to accept that, or require
  vendoring for the "real project" tier.
- 🔍 **How this is tested.** The Go-side tests stop at the artifact. Browser execution needs either a headless
  browser in CI or an explicit decision that it is verified by hand.

### Unknowns worth measuring early

- **Compile cost in a browser engine.** Under Node/V8 the 14.2 MB module compiles in 22 ms and instantiates in
  4–5 ms, flat in workload size — but that is V8 with a warm file, not a browser fetching over a network, and
  says nothing about Firefox or Safari.
- **The memory ceiling.** 138–150 MB for the petstore is comfortable; a real project is not measured, and wasm32
  caps at 4 GB with browsers often lower. A size guard and an honest diagnostic will be needed.

## Handover request to the quirk stream

Raised from this stream, to be applied on the quirk-fixing branch rather than here.

1. ⚠️ **Hoist the identity recognizers in `buildNamedField`**, mirroring what `ClassifierAliasStrfmt` did for the
   alias path. `parameters.go` and `responses.go` both ask `Ctx.DeclForType` before `resolvers.IsStdTime`, which
   reads only (package, type name) and never touches the declaration — so a graph that omits the declaring package
   fails on `time.Time` where it need not. The fix is the same principle the strfmt work already states: classify
   before the path that would discard it. It currently lives in this branch out of necessity (stdlib truncation is
   what made it testable, turning four hard failures into byte-identical scans); it belongs with the alias half.

2. 🔍 **`schema.resolveRefOrErr` has no stdlib-specials pre-check** (`ref.go`, untouched by the strfmt work). It
   goes `GetModel` → `missingSource`, whereas `buildDeclNamed` / `buildDeclAlias` both try `applyStdlibSpecials`
   first. That asymmetry sits inside the dispatch family the strfmt work is normalising, so it is worth an audit on
   its own merits; it is also the remaining `text-marshal` failure under a truncated graph.

3. ⚠️ **No identity recognizer for stdlib IO interfaces** — registered as **Q38** in `quirks-open.md`, and it is
   not merely a truncation concern: on a full graph `io.Reader` already emits an untyped parameter, and as a model
   field it publishes `#/definitions/Reader` and `#/definitions/ReadCloser` carrying io's godoc, the latter with a
   fabricated `close: {type: string}` property. Fixing it also closes the last truncated-graph failure.

## Appendix — open questions and risks

1. ⚠️ **Equivalence is proven on fixtures, not in general.** 138 patterns is good coverage but the fixture corpus is
   not real-world code. Before this replaces `packages.Load` natively, it needs a scan of a few large real projects.
2. ⚠️ **Build constraints.** The spike loader parses whatever `CompiledGoFiles` lists, which is correct *because the
   recorder resolved constraints natively*. A toolchain-free producer (action 8) must resolve `//go:build` itself —
   this is the single biggest unknown in the whole plan.
3. ✅ **GOOS/GOARCH are explicit.** Resolved in three tiers, weakest first: the running platform, then `Config.Env`
   (parity with `packages.Load`), then `WithTarget(goos, goarch)`. Surfaced to users as `Options.GOOS` /
   `Options.GOARCH`, honoured by *both* loader paths — the stock `packages.Load` path pushes them into `Config.Env`,
   since `go list` reads the target from the environment.
4. 🔍 **Dependency story for uploaded projects.** Unresolved, and the main UX risk for the browser tier. A
   `go mod vendor`'d tree needs no export data and no network — probably the honest answer, with an explicit
   diagnostic otherwise.
5. 🔍 **Payload for the browser.** Guest FS must carry the source of everything in the graph. Truncation is what makes
   this tractable (79 files vs. 1195 for the petstore); the depth knob is the lever.
6. ℹ️ **Artifact size.** 14 MB raw, **3.5 MB gzipped**. Fixed startup ~120 ms; throughput at rough parity with native.
   TinyGo remains a non-starter (`go/types` reflection).
7. ⚠️ **Two remaining depth-1 divergences** (`classification`, `go123`) are uninvestigated. `go123` fails with
   `unsupported type "invalid type"`, which suggests a type-check failure cascading rather than a truncation
   semantic — worth a look before trusting the depth-1 number.
