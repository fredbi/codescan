> [!NOTE]
> Last revision: 2026-08-17 — **scrubbed**. Round 1 shipped, and so did most of what round 2 was holding.
> What is left is nine items, checked against `master` `5e4cf04d` rather than inherited. The closed list
> is at the bottom, with what closed each one.

# WASI & playground — round 2

## Summary

Round 1 (PR #79) put a working playground on the doc site. This file was opened as a receptacle for what
round 1 left open — and a scrub on 2026-08-17 found **more than half of it had since been settled**, most
of it by other streams rather than here. What survives is a short list with one real decision in it.

**The one decision: do we go single-loader?** Everything else is bounded work.

## Open

### The loader

1. 🔍 **L1 — Make `internal/packages` *the* loader, or keep both?** The standing cost is two
   implementations of the same job selected per call (`ToolchainFreeLoader` / `FS`). **The blocker that
   held this is now discharged:** it was "equivalence is proven on fixtures, not on real-world code",
   and `internal/benchmarks` now measures two real go-swagger-generated projects under both loaders —
   dockerctl (501 files) and kubeapi (2352 files, 222 defs / 260 paths identical either way). So this is
   a decision awaiting a decider, not a decision awaiting evidence.

   Two items ride on it and should be decided in the same breath, not separately:
   - **Retire `x/tools/go/packages` from the runtime path.** Kept for its type vocabulary, which is
     aliased precisely so this stays a decision rather than a migration.
   - **Retire the toolchain-version CI guards** (`archive/textmarshaler-toolchain-independence.md`),
     which exist because the toolchain does the type-checking. If we own it, they go.

2. 📝 🏁 **L2 — Resolver tests against real trees on disk.** `replace` directives, vendor, the module
   cache: `fstest.MapFS` cannot reach any of them. ⚠️ Sharper than when this was filed —
   `internal/packages/list/` contains **exactly one test file** (`pkgpattern_test.go`, which covers the
   verbatim-copied matcher). `resolve.go` and `workspace.go`, which carry all of the D1–D9 `go list`
   semantics we deliberately reproduce, have **no test file of their own**. That is the gap, and it is
   also the prerequisite that would make L1 safe to answer "yes".

### Fidelity when the stdlib is not real

3. 🔍 **F1 — `encoding.TextMarshaler` is unreachable under `-stub-stdlib`.** A synthesized type has no
   method set, so nothing is seen to implement it. Half of this pair closed (Q38 gave `io.Reader` and
   friends identity recognizers); this half needs a real method set. **The open part is policy, not
   code: should a stubbed scan *fail* on a type whose meaning it cannot see, or degrade with a
   diagnostic?** Today it degrades silently enough that the user may not notice.

4. 🔍 **F2 — Truncation policy for rendered output.** The mechanics are measured; what is undecided is
   fidelity — `title`/`description`/`format` on a stubbed `time.Time` and friends — and whether
   truncation depth becomes a knob.

5. 🔍 **F3 — Two depth-1 divergences never investigated** (`classification`, `go123`). `go123` failed
   with `unsupported type "invalid type"`, which smells like a type-check failure cascading rather than
   a truncation semantic. **Nobody should trust the depth-1 number until this is looked at.**

### The artifact

6. 📝 🛠️ **A1 — A WASI runtime in CI**, so `TestWASIArtifact*` stops skipping there. Verified still
   open: `internal/integration/wasi_test.go` skips when neither runtime is on `PATH`, and no workflow
   installs one. The runtime *choice* is settled — the test prefers **wazero**, falls back to wasmtime;
   hermeticity won the argument. What is left is installing it in the lane.

### The playground

7. ⬜ **P1 — `W-report`, the "report issue" rail.** Pack annotated source + diagnostics + spec output +
   browser metadata into a pre-filled GitHub issue. One of the two conditions under which the playground
   stops being a PR stunt. ⚠️ **Blocked only by `feat/scrambler` being an unmerged one-commit branch.**
   ➡️ **Tracked as one theme with the TUI's Repro-pack in `backlog.md` §3** (Fred, 2026-08-17): one
   capability, two front ends, and the house method is spike in the TUI then port here. This entry is
   the playground half; the ordering lives there.

8. ⬜ **P2 — `W-privacy`.** No upload, no telemetry, enforced by static-only deploy. True today by
   construction; **unstated and untested**, and it is the kind of claim that wants writing down and
   pinning rather than assuming.

9. 📝 🏁 **P3 — Test the thing the helpers are assembled into.** Verified: 10 `.test.ts` files, all under
   `src/lib/` — pure helpers, well covered. There is **no component or store test**: `Playground` is
   never instantiated and driven. One of the eight bugs Fred caught by hand was four lines of store
   test. Do store tests before anyone prices a Playwright job.

10. ⚡ **P4 — The scan retains nearly everything it touches** — 372 MB obtained, 352 MB still live, two
    collections. Retention, not churn; the type graph is the suspect. The 64 MB upload ceiling depends on
    this number.

## Closed by the 2026-08-17 scrub

Kept as a list so nobody re-opens them from the round-1 documents.

| Was | Closed by |
|---|---|
| Adopt `swag/fs` once promoted | ⛔ **Dropped (Fred).** `swag/fsutils/FS` is real and may be useful to callers, but it will not replace what `internal/packages` does. Not our dependency to wait on |
| Promote the loader seam | `scanner.LoadPackages` is gone; selection is on `Options` |
| Types without declarations (export-data corpus gaps) | The declaration contract (PR #90) + per-declaration read-back (PR #99); the A/B emptied for all three configurations over the whole corpus |
| Release wiring — wasm artifact, export-data artefact | **Moved**, not closed: it is `release-v0.36.4.md` §3, the goreleaser thread |
| Runtime choice for CI — wasmtime vs wazero | Settled in code: wazero preferred, wasmtime fallback |
| How the stdlib reaches the browser | Settled by what shipped: the worker runs `'embedded'` (the `exportdata` build tag). `'mounted'` and `'stub'` stay selectable in `flags.ts`; the 8.4 MB first-load cost was accepted |
| Front-end polish beyond round 1 | Landed (Fred) |
| Stop npm dev-dependency majors auto-merging | Landed, and **better than this plan proposed**: `versioning-strategy: lockfile-only` puts majors out of reach entirely — the manifest's declared ranges are the ceiling — plus a `typescript` group carved out of the auto-merge substring |
| Do not take TypeScript 7 yet | Same mechanism; the carve-out keeps a human in the loop for the day svelte-check widens its peer range |
| Turn npm version-update noise down | `lockfile-only` + weekly + `open-pull-requests-limit: 3` |
| Share link | ⛔ won't do — a fragment holds an edited example; an opened module is three orders of magnitude past it |

## Companion repos — decided in principle, gated in practice

**Direction (Fred, 2026-08-08):** codescan stays a pure Go library. The npm/TypeScript half moves to
`codescan-js`, editor clients to per-language repos. Later, not now.

The gate is the release wiring, and it is one item rather than two: `npm run wasm` reaches three
directories up into the Go module, so a separate repo must consume a *published* `.wasm` release asset or
the coupling just moves into a sibling checkout in CI. **The split cannot precede the release wiring.**
The trigger to act is the second TS consumer — a VS Code extension.

Knock-on worth remembering: once the artifact is published, `update-doc.yml` should *fetch* the pack
rather than build it, which drops Go from the documentation build.

## Notes for whoever picks this up

- **Rebase first.** This branch was cut at `5c36b34` and master is 60-odd commits past it, through a
  `fixtures/` → `testdata/` rename (PR #91) and a command-line refactor (PR #111 / #114). Every fixture
  path cited in the round-1 documents is wrong.
- **The doc-site pack is a build artifact, not a checkout.** `npm run dist` builds the wasm artifact
  *and* the app and refuses a half pack on purpose. CI does the same in `update-doc.yml`, and needs
  `GOWORK=off` plus `cmd/genspec-wasi/` and `internal/` in the sparse checkout.
- **Styling inside somebody else's page:** the `@layer` contract means our styles win on specificity only
  for properties we actually state, so every colour has to be spelled out. Both round-1 styling bugs were
  this.
- **Paths are the recurring bug class.** `absolutePath` shipped testing absoluteness with
  `strings.HasPrefix(p, "/")` and Windows CI caught it. Anything here that touches a path gets read for
  host-vs-guest confusion before it is called verified.
- **Round-1 measurements worth not re-taking:** filesystem syscalls are ~1.8 % of a WASI scan; stdlib
  from source 7.3 s / 681 MB → export data 1.0 s / 138 MB → stubbed 0.98 s / 147 MB on the petstore; the
  artifact imports 21 WASI functions and every one reads, pinned by `TestWASIArtifactImportsOnlyReads`.

Predecessors: `archive/wasi-build.md` (the artifact + `internal/packages`), `archive/playground-ux.md`
(the front end), `archive/wasm-playground.md` (origin vision, several guesses superseded).
