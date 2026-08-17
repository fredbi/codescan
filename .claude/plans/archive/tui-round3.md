> [!NOTE]
> **✅ ALMOST COMPLETE** — archived 2026-08-17. Merged (PR #108, plus #110). The **relevance pass closed the same day** (Fred): the card stands as shipped and its three candidate refinements were assessed and declined — see §5. The only thing outliving this document is "further round-3 items", never picked; it sits in [`backlog.md`](../backlog.md) §2 beside the light/dark theme. Repro-pack moved out to `backlog.md` §3, now a cross-cutting theme shared with the playground.
>
> _Original document follows unchanged._

> [!NOTE]
> Last revision: 2026-08-17 — ✅ **MERGED** (PR #108, `cac6b454`): `1a182dce` the run-cost card, `463e5aa1` the
> profiled run. Two follow-ups landed straight after: `2b967d42` (PR #110) counts every allocation when checking
> attribution, `72498f8f` keeps two path segments in a profiled function name. **Item 5 closed 2026-08-17**;
> only item 6 (further items, never picked) outlives this document.

# TUI round 3 — small improvements to genspec-tui

## Summary

A new round of small, independent UX improvements to `cmd/genspec-tui`, in the same spirit as rounds 1 and 2: each item is
its own commit, each is shippable on its own, and none of them changes what the scanner produces.

Opening item: **a run-cost overlay** (`m`) reporting what the last scan cost in memory — the gap between before and after
the run — the way the wasi playground reports `jsonRuntime` in its status-bar tooltip. The TUI's status bar is already
packed, so this is a modal shown on demand rather than another field on a line.

The rest of the round is open: further items get appended to Actions as we pick them.

## Context

Base camp: worktree `.worktrees/fix/tui-round3`, branch `tui-round3`, rebased onto master `fac7579` (post PR #100).

The precedent is `cmd/genspec-wasi/envelope.go`: `jsonRuntime` carries `Sys`/`HeapAlloc`/`TotalAlloc`/`NumGC` read straight
after the scan, and `StatusBar.svelte` explains them to the reader in a tooltip. Two things it establishes that we keep:

- **No `runtime.GC()` before reading.** Forcing a collection reports a tidier heap than the one the scan actually ran with,
  which is the opposite of the question being asked.
- **The numbers are the process's, not the scanner's.** They are worth carrying because nobody can recover them afterwards.

What the wasi version cannot do and the TUI can: take a reading *before* the run as well as after. That is the whole ask —
the gap, not the endpoint.

### Two honesty constraints, settled up front

1. **The window is process-wide.** `ReadMemStats` accounts for the whole process, and during a scan the bubbletea render
   loop, the spinner ticking at ~10 Hz, and the fs watcher all allocate on other goroutines. The reading is therefore
   "everything this process allocated across the scan window", not "what codescan allocated". The overlay says so rather
   than implying a precision it does not have.

2. **A rescan holds two specs at once.** `absorbScan` replaces `scan.JSON`/`scan.YAML`, the source index and the spec index
   only *after* the new scan lands, so at the "after" fence the old document and the new one are both live. The retained
   delta on a rescan therefore reads high by roughly one document's worth, and that is correct behaviour rather than a leak.
   The overlay carries this as a one-line footnote on rescans; without it the number invites exactly the wrong conclusion.

### Settled with Fred (2026-08-14)

- **Last run only**, not a history table. A single card, closest to the wasi tooltip.
- **Three fences**, so the card can attribute the cost: before `codescan.Run`, after it, after JSON/YAML rendering.
  Splitting scanning from rendering says whether the cost is the scanner's or ours for serializing the result twice.
- No live-heap sampling / high-water mark for now (the spinner tick could carry it; parked as an appendix idea).

## Trajectory

1. ✅ **Measure the run** — three `MemStats` fences in `scan.Do`, carried on `ScanResult` into `ScanState`
2. ✅ **Show it** — a `runstats` overlay on `m`, registered like the other modals
3. ✅ 🏁 **Tests** — formatter, render, and the open/close/capture path through the model
4. ✅ 📚 **Docs** — help overlay entry, README keymap row, package `doc.go`
5. ✅ **Relevance pass** — figures judged against real runs (Fred, 2026-08-17); the card stands as shipped
6. 📝 **Further round-3 items** — to be picked (see Actions)

## Actions

### 1. Measure the run

1. 📝 `internal/ux/scan/scan.go`: take `runtime.ReadMemStats` at three points inside `Do` — before `codescan.Run`, right
   after it, and after JSON+YAML rendering. No `runtime.GC()` at any of them.
2. 📝 New type in the `scan` package (`RunCost`, or `Cost`) holding what the card needs, derived from the three snapshots:
   allocated (ΔTotalAlloc) split scan/render, retained (ΔHeapAlloc) split scan/render, objects (ΔMallocs−ΔFrees), GC cycles
   (ΔNumGC), live heap before and after, `Sys` at the end. Deltas computed once here, not re-derived at the render site.
3. 📝 Stamp the elapsed split alongside it (scan vs render), since the fences are already there and the header currently
   shows only the total.
4. 📝 Carry it on `scan.ResultMsg`; `absorbScan` stores it on `ScanState` next to `Elapsed`. It is a *reading*, so a failed
   scan carries whatever was measured up to the error rather than nothing.
5. 📝 Nothing in the status bar or header changes.

### 2. Show it

1. 📝 New package `internal/ux/runstats` implementing the `Overlay` interface (`SetSize`/`IsOpen`/`HandleKey`/`View`),
   modelled on `internal/ux/help` — a concrete type the model owns, not a `tea.Model`.
2. 📝 Card layout as agreed:

   ```text
   Last run

     elapsed        1.8s     scanning 1.7s · rendering 0.1s

     allocated      412 MB   during the run
       scanning     356 MB
       rendering     56 MB
     retained       +38 MB   94 MB → 132 MB live
       scanning     +31 MB
       rendering     +7 MB
     objects        +410 k
     GC cycles      7
     from the OS    221 MB

   process-wide, across the scan window
   esc / m: close
   ```

3. 📝 `key.M` binding; `m` opens from the global switch in `handleKey`, `esc`/`m`/`enter` close (help's convention).
   Registered in `Model.overlays()` after `help`.
4. 📝 Empty state before the first result — "no run measured yet" — and the rescan footnote when the run replaced a
   previous document.
5. 📝 Byte formatter alongside the existing `humanDuration`: signed for deltas (`+38 MB`, `−4 MB`), and a short count form
   for objects (`410 k`).

### 3. 🏁 Tests

1. 📝 Formatter table test (bytes, signed deltas, counts; the boundary cases KB/MB/GB).
2. 📝 Render tests in `runstats` on a fabricated cost: full card, empty state, rescan footnote.
3. 📝 Model-level test that `m` opens the overlay, that it captures keys while open (an arrow key must not move a pane),
   and that `esc` closes it — the shape `help_test.go` / `options_test.go` already use.
4. 📝 A scan-level test that `Do` returns a cost with sane invariants (allocated ≥ 0, split parts sum to the total, live
   heap figures non-zero). Deliberately no assertions on magnitudes, which are not reproducible.

### 4. 📚 Docs

1. 📝 `helpSections` "anywhere" entry: `m` → "what the last scan cost (time and memory)".
2. 📝 README keymap table row (it mirrors `helpSections` by hand).
3. 📝 `doc.go` for the new package, carrying the two honesty constraints above.

### 5. ✅ Relevance pass — DONE (Fred, 2026-08-17)

Judged against real runs. **The card stands as shipped**; none of the three candidates below was worth
acting on, so they are kept as the record of what was considered and declined rather than as open work.
Nothing from this section carries into `../backlog.md`.

Candidates, all assessed and left alone:

1. 🔍 **Sub-second durations.** `humanize.Duration` is the header's spelling: everything from 1.0s to 1.4s reads "1s",
   and a 1.7s scan reads "2s". Fine on a status line, coarse on a card whose whole subject is what the run cost. One
   decimal between 1s and 10s would fix it — and would change the header's `ready (…)` too, which is why it is a
   decision rather than a tweak.
2. 🔍 **Is `retained` the useful number?** It is the honest one, but it is measured through two confounds (the rescan's
   double-held document, the process-wide window). If it turns out to be unreadable in practice, the card should lead
   with allocated + churn ratio instead and demote retention.
3. 🔍 **Is the scan/render split earning its space?** Rendering is expected to be a rounding error next to the scan on
   any real tree. If it consistently is, the two sub-lines can collapse into one gloss.

### 6. ✅ Profiled runs — `-profile` (built, awaiting Fred's session test)

> The card's figures are process-wide scalars, and no amount of care makes a scalar attributable. Profile records are
> per-stack: the redraw loop and the watcher stop being a confound and become a named row you can discount.

**Settled with Fred (2026-08-14):** both reports rendered in the TUI (so the pprof `profile` package comes in as a
dependency of the TUI module — probed: self-contained, 2 go.sum lines, no transitive requirements); phases cut as
**scan and render**, matching what the card already splits.

1. 📝 **Launch flags**, not an options toggle. `MemProfileRate` is meant to be set once at startup and held constant,
   and rate 1 changes what it measures — a mid-session toggle would make consecutive readings incomparable.
   - `-profile` — capture a CPU profile and allocation profiles around each scan
   - `-profile-dir` — where the `.pprof` files land (default: a fresh temp dir, reported on the card)
   - `-mem-profile-rate` — `runtime.MemProfileRate`; 0 keeps the runtime default (512 KiB sampling), 1 records every
     allocation exactly, at a cost
2. 📝 **Capture**, in `internal/ux/scan` next to the fences it brackets:
   - a CPU profile per phase (`cpu-scan.pprof`, `cpu-render.pprof`). **Revised 2026-08-14** — the plan said scanning
     only, on the grounds that rendering is too short to sample. Measured: 6 ms for a 54 KB spec, ~17 ms per MB, so a
     large API renders in the 100 ms range and *is* samplable. The bracket is now symmetric, and a phase that caught
     nothing says so — an absent section reads as an oversight, "nothing sampled" is a measurement. What stands is
     not *raising* the rate: `runtime.SetCPUProfileRate` above pprof's 100 Hz makes the runtime print to fd 2, which
     no `log.SetOutput` can silence and which would paint over the alt-screen
   - three heap snapshots at the existing fences (`mem-before` / `mem-after-scan` / `mem-after-render`), so
     `go tool pprof -base` reproduces either phase in the real tool
   - per-phase allocation attribution in-process, by diffing `runtime.MemProfile` records across each fence pair
   - `runtime.GC()` at each memory fence — the opposite of the unprofiled rule, and correct here: a diff without it
     misses the tail of the phase, and accuracy is the whole point of the mode
   - a package mutex with `TryLock`: the CPU profiler is a process-wide singleton, and two scans can be in flight
     (an options change during a scan). A run that cannot get it is not profiled, and says so
   - no profiling failure ever fails a scan; each becomes a note on the card
3. 📝 **Report**, on the same card, which grows scrolling (the help overlay's pattern):
   - top CPU functions for the scan, flat, with their share and the sampled total
   - top allocation sites per phase, sampling-corrected the way pprof's own `scaleHeapSample` does, or exact under
     rate 1
   - the artifact paths and the `go tool pprof` command that opens them
   - a badge on the MemStats block: a profiled run's own figures include the profiler's overhead and its forced
     collections, so the two accounts must not be read as one
4. ✅ 🏁 Tests: the scaling correction, stack attribution, the CPU top-N derivation from a fabricated profile, the
   report's rendering, and that a disabled profiler leaves the scan path untouched.

**Two things the first real profiled run changed**, neither of which the design predicted:

- **Allocation rows named the runtime, not the caller.** `runtime.newarray`, `runtime.convTstring`,
  `runtime.mallocgc` — every allocation goes through the same handful of helpers, so the innermost frame says only
  that memory was allocated. `frameName` now walks out to the first frame outside `runtime.`/`internal/`, and the
  table reads `types.(*Checker).recordTypeAndValue`, `gcimporter.UImportData`, `parser.(*parser).parseIdent`.
- **A failed run reported allocations in a phase that never happened.** The record key hashed a few stack frames
  weakly, so two records collided and differenced against each other — manufacturing allocation, including where a
  fence had been closed twice on the same snapshot and the answer had to be zero. Now FNV-1a over the whole stack,
  with a regression test that differences a 512-record snapshot against itself.

Also added, from reading that first real table: the CPU heading carries the **sample count**, and under 100 samples
(a second of CPU at 100 Hz) the card says outright that this is not a ranking.

**A third thing the two-phase run exposed:** the rendering allocation table was reporting the *profiler* —
`pprof.StartCPUProfile`, `flate.NewWriter`, `emitLocation` — because stopping one phase's profile flushes it through a
compressor and starting the next allocates its buffers, all between two fences. In the short phase that was a visible
share of the table. Records whose stack carries a `runtime/pprof.` frame are now dropped, and the card says the tables
exclude the observer. With that gone, rendering reads as it should: `yaml/v3.yaml_emitter_emit` at 4.8 MB — YAML, not
JSON, is what rendering costs.

### 6.1 ✅ Where the CPU table is charged (2026-08-15)

> Fred, reading a profiled scan of a generated kubernetes server: the two stages come out clearly, but the next-level
> bottleneck does not — "we report only deep runtime dependencies. We get a better sense by capturing the functions
> that are at the level of our direct dependencies, not deeper."

He was right, and the two tables were not even using the same rule. The allocation table already walked out of the
runtime (`frameName`); the CPU table charged `s.Location[0]` verbatim, so it read as `runtime.scanObjectsSmall`,
`runtime.tryDeferToSpanScan`, `maps.ctrlGroup.matchH2` — the machine, not the program.

**The rule now:** a sample is charged to the *deepest frame under `github.com/go-openapi/codescan`* together with
**what that frame called** — the boundary where our code hands work to somebody else's. Every sample is charged
exactly once, so the shares still partition the phase; they stop being flat, and the card says so.

Three cases fall out, each settled against a real profile rather than guessed:

1. **No call of ours above it** — GC, the allocator, the scheduler: one synthetic row, `the runtime itself`. Fred:
   "from my experience this runtime overhead is mostly a consequence of the high memory activity that we can see on
   the memory alloc table" — which is exactly what the row is for, so it sits directly above that table.
2. **A goroutine we did not start** (x/tools fans out through `errgroup`): charged to what the goroutine is *doing*,
   not to the wrapper every one of them roots in. Found by running it: the first version reported
   `errgroup.(*Group).Go.func1` at 41%, a dead end.
3. **Function literals** fold into the function they were written in — `refine.func2.1` and `refine.func3` are one
   piece of work seen at two of its entrances, and two rows the reader has to add up.

Also: the sampler catches the profiler writing its own profile, so those samples are dropped from the table *and* the
total. The card claimed the tables excluded the observer; until now that was only true of the allocation half.

**Before / after**, same scan (go-swagger through the toolchain-free loader, 191 samples):

```
runtime.scanObjectsSmall / tryDeferToSpanScan / spanClass.sizeclass / memclrNoHeapPointers ...
      ↓
  38%    720ms   the runtime itself — collecting, allocating, scheduling
  26%    490ms   packages.(*loadState).loadDir → types.(*Config).Check
  25%    480ms   packages.(*loadState).loadDir → parser.ParseFile
   4%     70ms   packages.(*loadState).loadDir → build.(*Context).ImportDir
   3%     60ms   vfs.(*FS).ReadDir → os.(*unixDirent).Info
```

**The memory table keeps the old rule, deliberately.** Agreed with Fred that the answer there is not obvious and has
to be experimented with: charging allocations to the boundary would fold `types.(*Checker).recordTypeAndValue`
(123 MB — the type-checker's own result set, and the single most informative row on the card) inside
`types.(*Config).Check`. The two tables are now at two altitudes on purpose: *which call of ours costs* above,
*what inside it holds the memory* below. Revisit after a few real sessions.

### 6.2 ✅ What the profile found: the YAML render nobody asked for (2026-08-15)

> Fred: "the rendering shows most of our CPU and memory activity going to yaml/v3 yaml_emit, but this run doesn't
> render as YAML, just json."

The profile was right and the TUI was wrong. `scan.Do` marshalled the spec to JSON and then called `jsonToYAML` —
unmarshal the JSON back into `map[string]any`, run the emitter over it — to fill `ResultMsg.YAML` for the spec pane's
`ctrl+y` toggle, on **every** scan, whether or not the toggle was ever pressed.

Measured on the 46 KB classification golden (×40):

| | time | churn |
|---|---|---|
| `json.MarshalIndent` | 2.0 ms | 773 KB |
| `jsonToYAML` | 4.5 ms | 3.6 MB |

70% of the render phase's time and 82% of its allocations, scaling with document size — on a kubernetes-sized spec,
a large share of every rescan spent on a view nobody had opened. It also explains the negative *retained* the render
phase reported: that much garbage in a short phase is what tips a collection inside the bracket.

**Now:** `scan.RenderYAML` is called when the YAML view is first asked for, as a `tea.Cmd` so the conversion never
blocks the event loop, and the pane says `(rendering YAML…)` meanwhile. The result is cached until the next scan
replaces the document. A conversion overtaken by a rescan is discarded on a generation counter rather than shown
against the document that replaced the one it was made from.

Two consequences worth remembering: the render phase now measures what a scan actually renders (so the card's split
means something different from round 2's readings), and `ResultMsg.YAML` is gone — the model owns that render.

**The general lesson, and the reason the profiling work paid for itself:** the first thing the boundary-charged table
found was not a bottleneck in codescan but a whole computation the TUI did not need to be doing.

### 6.3 ✅ What the profiling stream concluded (2026-08-15)

Fred, reading the boundary-charged tables: *"I don't think we'll be able to go much farther on the scanning side.
96% of the activity is done compiling go code, which means our grammar mechanism has a very low overhead over code
parsing, which is beyond our reach. So perf is near maxed out."*

Agreed, with the evidence: `go/types` + `go/parser` are 51% of a real scan, the allocation table under them is
`recordTypeAndValue` / `parseIdent` / `AddLine` / `updateExprType` top to bottom, and **no package of ours appears in
either table**. Whatever the grammar costs, it is under the sampler's resolution next to compiling the code it reads.

Three things left on the record rather than acted on:

1. ⚡ **The 38% runtime row is the one remaining lever**, and it is a policy rather than an optimisation — filed as
   [`features/gc-tuning-scan.md`](../features/gc-tuning-scan.md) (cross-cutting, `genspec` only: a TUI session must not
   hold several times its live heap between scans). `GOGC=off` is the free upper bound; if it does not move elapsed,
   the question closes.
2. 🔍 **Re-read `scanner.(*ScanCtx).FindDecl` before declaring our side negligible.** Its 4% was measured under the
   old flat-leaf rule, which counts only samples whose leaf was literally inside it; charged at the boundary that row
   absorbs what it calls. It is the only frame of ours that ever showed up at all.
3. 🔍 **The two loaders are not at the same ceiling.** The 25% in `parser.ParseFile` is `-loader own` type-checking
   dependencies from source; the go/packages path takes them from export data instead. `ExportData` is what closes
   that gap and already exists — see `loader-vs-gopackages.md`.

Rendering memory (the document, both indexes, the source index, held across rescans) remains worth reducing as a
**footprint** fix for long sessions, not an elapsed one.

### 7. Further round-3 items

1. 🔍 *(open — to be picked with Fred)*. Carried over from earlier rounds and still open:
   - light/dark theme selection (deferred 2026-08-07 in `tui-ux-enhancements.md`, not dropped).

## Achievements

### 1–4. The run-cost card, bound on `m` ⭐⭐

Landed on `tui-round3` (uncommitted working tree, awaiting Fred's session test):

- ✅ `scan.Cost` + three fences in `scan.Do` (`internal/ux/scan/cost.go`). The godoc states both constraints — the
  window is process-wide, and three fences are not a peak — next to the numbers they qualify.
- ✅ `internal/ux/runstats` — the card, its two disclosure lines, and the empty state before the first scan.
- ✅ `internal/ux/humanize` — `Duration` (moved out of `chrome.go`, so the header and the card spell a duration the
  same way), `Bytes`, `SignedBytes`, `SignedCount`.
- ✅ `key.M`, registered in `overlays()`, help entry, README keymap row.
- ✅ 🏁 Tests: `costOf` arithmetic incl. a collection mid-run and a failed run; card rendering, disclosure lines,
  empty state, key capture, frame stability; model-level open/capture/editor-passthrough plus the rescan flag.
- ✅ 😇 `golangci-lint run --new-from-rev master`: 0 issues. Full suite green, `-race` green.

## Appendix — parked ideas

- ♥️ **Live-heap high-water mark.** The spinner already ticks while a scan runs, so sampling the live heap there would give a
  peak rather than two endpoints, and would catch a scan that balloons and drops back. Parked: `ReadMemStats` is
  stop-the-world, so sampling at 10 Hz perturbs the thing it measures; `runtime/metrics` (`/gc/heap/live:bytes`) would be
  the cheaper read if we ever want it.
- ♥️ **Run history.** A ring of the last N runs would show whether retained memory climbs monotonically across rescans —
  the one reading that would prove or disprove a leak in the TUI's own indexes. Explicitly not this round.
