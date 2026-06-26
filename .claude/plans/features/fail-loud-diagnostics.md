---
title: Scanner robustness & fail-loud diagnostics
stream: 9
origin: iii
status: done
release: v0.35
issues: [go-swagger#2886, go-swagger#2874]
prev: "§8"
---

# Scanner robustness & fail-loud diagnostics

**Status:** ✅ done (2026-06-17, merged into `fix/backlog-lot1`) · `bc568c7` (8.1
panic-guard), `d2479a4` (8.2 degraded-load). Design + settled decisions:
[fail-loud-diagnostics.md](../fail-loud-diagnostics.md).

Two halves of the same fail-loud theme, both landed. A new `scan.*` diagnostic
family (`scan.internal-panic`, `scan.degraded-load`) surfaces conditions that
previously failed silently or with only a raw Go stack trace.

**8.1 — panic recovery with a located diagnostic.** A per-decl `guard` wraps the
six spec build loops, plus a coarse `Run` backstop; a recovered panic emits a
located `scan.internal-panic` diagnostic (file:line + label) and aborts via
`ErrInternalPanic`. Witness: `TestBuilder_guard_*`.

**8.2 — tiered fail-loud on degraded package load.** `detectDegradedLoad` runs
after `packages.Load` and inspects the matched root packages. **Tiered:** ABORT
(`scan.degraded-load` Error + `ErrDegradedLoad`) on an empty set /
`packages.ListError` / nil `Types`/`TypesInfo`; WARN + continue on parse/type
errors that still leave usable types (so a `./...` sweep is not sunk by one
broken sibling). Witnesses: `TestNewScanCtx_InvalidPackage` (abort),
`TestNewScanCtx_PartialLoad_WarnsAndContinues` (warn).

**Origin.** go-swagger#2886 (a SIGSEGV in the responses builder aborted with only
a Go stack trace, no offending source) and go-swagger#2874 (with `GOROOT` unset,
`swagger:allOf` silently stopped resolving, emitting a degraded spec with no
warning). Both triaged from backlog verification (2026-06-13).
