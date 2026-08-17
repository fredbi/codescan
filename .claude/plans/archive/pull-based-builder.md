> [!NOTE]
> **⏸️ PARKED, TAIL RECORDED** — archived 2026-08-17. Nothing was built and nothing is asking. `SchemaCache` moved to [`internal-document-model.md`](../internal-document-model.md); what remained — the per-package declaration index, `V-builder` proper, the large-main-module measurement gap and the `Names()`/`DefKey()` caveat — is summarized in [`backlog.md`](../backlog.md) §4.
>
> _Original document follows unchanged._

> [!NOTE]
> Last revision: 2026-08-17 — **`SchemaCache` has found its home.** Fred requalified stateless model
> discovery (2026-08-17) as a **maintainability** item rather than a performance one, and it now belongs
> to the `V-imodel` pillar: see [`internal-document-model.md`](../internal-document-model.md) §Stateless
> discovery, which states what is in the code today and what replaces it. This document keeps the rest
> — the per-package declaration index, `V-builder` proper, and the measurement gap — which stay parked.
> Previous revision: 2026-08-07 (split out of `on-demand-scanner.md`).

# Pull-based builder (Stream 8 · `V-builder`)

## Summary

The build side of Stream 8 — everything that consumes the scanner rather than loads it: a schema cache, a
per-package declaration index, and the pull-based spec builder those two were meant to serve.

Deprioritized 2026-08-07, deliberately and not by neglect. `V-builder`'s premise was to *pull packages to
parsing* through the on-demand scanner. That scanner stream closed without shipping lazy materialisation —
its harness measured the remaining prize at ~0.03 s / 11 MB and the loader now decides source-vs-export-data
per dependency at load time, so there is nothing left to pull against. The forcing function is gone.

What survives is the *architectural* argument, which never depended on performance. This document keeps it
alive so it is not rediscovered from scratch in six months.

## Context

Split out of [`on-demand-scanner.md`](on-demand-scanner.md), which retains only its live worklist. Read that
document's Achievements and Appendix A for the measurements referenced here; they are not repeated.

### Why this was deprioritized

| premise | state 2026-08-07 |
|---|---|
| "leverage the on-demand scanner to pull packages to parsing" | ⛔ there is no lazy materialisation to pull against |
| "closes the perf/OOM cluster" | ✅ closed, by the loader branch, by a different route |
| "hard-depends on the declaration contract" | ✅ prerequisite met — unblocked, just not urgent |

The roadmap's ordering note (**V-scanner before V-builder**, `roadmap.md`) is therefore satisfied in an
unexpected way: the prerequisite landed while the thing it was a prerequisite *for* lost its motivation.

### What still argues for it

1. **The fixpoint loop.** Settled 2026-08-05 (Fred): the cache **is** how the fixpoint is reached — recast,
   not removed. That is a statement about the builder's shape, not its speed, and it stands whatever the
   loader does.
2. **A cleaner spec-side index** for Streams 5 and 11 (TUI cross-ref, LSP). Both want to ask "what produced
   this pointer" and "what does this type reach", which is what a pull-based builder maintains anyway.
3. **`SchemaCache` is build-once, reference-many** — consumers stay type-keyed, resolution stays internal.
   Today the same type is rebuilt per reference.

None of these is urgent, and none of them is worth touching before something downstream actually asks.

## Trajectory

1. ➡️ **`SchemaCache`** — **MOVED** to [`internal-document-model.md`](../internal-document-model.md)
   §Stateless discovery (2026-08-17). Build-once, reference-many; resolves per package internally,
   consumers stay type-keyed.
   > It was always "the one item with an argument independent of the retired perf case" — and that
   > argument turned out to be the `V-imodel` pillar's, not this document's. The IR is what makes the
   > cache fall out rather than be built: retiring `ExtraModels` / `postDecls` / the two fixpoint loops
   > is a consequence of builders producing IR nodes instead of mutating the scanner's index.

2. 🔍 **Per-package declaration index** — retires the linear `FindDecl`
   > Carried over as an open *question*, not a plan. Its stated case was performance and that case is gone;
   > whether the linear walk is a maintainability problem in its own right is unmeasured. Route
   > `FindDecl` / `DeclForType` / `FindComments` / `FindEnumValues` / `FileForPos` through it *if* it earns it.

3. ⬜ **`V-builder`** — pull-based spec builder, hard-depends on the above

4. ⚡ **A measurement gap: a large main module**
   > Both benchmark targets have a small main module (19 and 30 packages) against a large closure. A repo whose
   > *main module itself* is thousands of packages is a different shape and is currently unrepresented — and it
   > is the shape where a per-package index and a schema cache would show up, if they ever do.

## Actions

### Decide before implementing

1. 🔍 **Does `SchemaCache` change any output?** The fixpoint recast is the interesting half: a cache that
   reaches the same fixpoint by a different route is a refactor, one that reaches a *different* fixpoint is a
   behaviour change wearing a refactor's clothes. Whole-document A/B over the golden corpus is the only
   acceptable evidence — see `internal/integration/loader_agreement_test.go` for the pattern (compare
   marshalled documents, assert known divergences rather than tolerating them).

2. 🔍 **What triggers this work?** Currently nothing does. Candidate forcing functions, none of them active:
   Stream 5/11 needing a spec-side index; a large-main-module corpus showing the linear `FindDecl`; or a
   builder change that the current rebuild-per-reference shape makes unreasonably awkward.

### Carried forward from the scanner stream

3. ⚠️ **`Names()` / `DefKey()` still read `Comments`** for the `swagger:model` override, so a declaration whose
   syntax was *deferred* rather than absent would silently fall back to its Go name. Harmless today — absence
   is permanent, never deferred — and stated explicitly in its test. It is the one place laziness would bite,
   so it belongs to whoever revives materialisation, if anyone does.

## Achievements

Nothing built. The prerequisite is met (`on-demand-scanner.md` step 2 — the declaration contract), which is
why this is a parked plan rather than a blocked one.
