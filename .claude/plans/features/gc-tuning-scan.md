---
title: Reclaim the GC share of a scan (genspec)
stream: —
origin: ii
status: open
release: null
issues: []
prev: null
---

# Reclaim the GC share of a scan (`genspec`)

**Status:** ⬜ open · ⚡ perf **experiment**, not a feature — the deliverable is a measurement and a
decision, and "measured, it does not pay" is a complete outcome.

**Origin.** (ii) — fell out of the genspec-tui profiling work (2026-08-15, `tui-round3.md` §6.1). Once the
CPU table was charged at the boundary between our code and its dependencies, the largest single row on a
real scan stopped being a function at all.

## What the profile says

`go-swagger` scanned through the toolchain-free loader, 191 samples:

```
  38%    720ms   the runtime itself — collecting, allocating, scheduling
  26%    490ms   internal/packages.(*loadState).loadDir → go/types.(*Config).Check
  25%    480ms   internal/packages.(*loadState).loadDir → go/parser.ParseFile
```

Alongside: **589 MB churned, 368 MB still live at the end, 14 GC cycles** in about a second of wall clock.

Fred's reading, which the tables support: `go/types` and `go/parser` are 51% and they are not ours to
make faster — the grammar's own overhead never even reaches the top eight. So the scanning side is close
to its ceiling **as code**. The 38% is the exception: it is not a bottleneck to optimise but a policy to
choose.

## The hypothesis

A scan is a batch job that allocates ~590 MB to *keep* ~370 MB. At the default `GOGC=100` the collector
runs every time the heap doubles, marking a live set that is mostly still live — work spent on memory the
process is about to hand back wholesale. Raising `GOGC` (or pinning `GOMEMLIMIT` and turning `GOGC` off)
trades resident memory for collections.

`GOGC=off` is the free upper bound on the whole idea: **if it does not move elapsed, nothing in this
direction will**, and the question closes for good.

## Protocol

One tree, one loader, `-profile`, five runs each, median — and a matching unprofiled set for elapsed,
since the sampler perturbs what it measures.

| knob | what to record |
|---|---|
| default (`GOGC=100`) | baseline |
| `GOGC=400` | elapsed · the runtime row's share · allocated · retained · from the OS |
| `GOGC=off` | the ceiling: the most this direction can ever give |
| `GOMEMLIMIT=<n>` + `GOGC=off` | the shape a batch job actually wants |

Run it on two shapes, because they load differently: a big tree through `-loader own` (parses and
type-checks from source) and the same tree through `-loader go` with compiled dependencies (export data,
much less parsing). The GC share is not the same in both.

## Decision rule, fixed in advance

- **Adopt** if the median elapsed improves by ≥10% at a high-water cost the CLI can pay (≈2× is fine — the
  process exits).
- Otherwise **close it**, and record the numbers here so nobody re-opens the question in six months.

## Where a knob would live

`cmd/genspec` — `debug.SetGCPercent` before the scan. A CLI process exits when the document is written, so
the memory it does not give back costs nobody anything; that is what makes the trade free *here* and only
here.

Not automatic in `cmd/genspec-tui`: a session that holds several times its live heap between scans is worse
than one that scans slightly slower, and the run-cost card would report a high-water mark that says more
about the knob than about the scan. If the TUI wants any of this it wants a `GOMEMLIMIT` ceiling, not a
percentage.

Note that `GOGC` and `GOMEMLIMIT` are already the user's to set through the environment. So the increment
is only ever **which default we pick and whether we document it** — which is also why the honest outcome
may be a paragraph in the genspec README rather than a line of code.

## What this is not

Not an attempt to make `go/types` faster, and not the *other* lever: how much Go we compile at all
(`ExportData`, `SkipCompiledDependencies`, tighter package patterns) is a separate question and a larger
one. See `loader-vs-gopackages.md`.
