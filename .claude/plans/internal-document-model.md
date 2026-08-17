> [!NOTE]
> Last revision: 2026-08-17 (opened — **shaping, not building**)

# Internal document model (Stream 8 · `V-imodel`) — the v0.37.x pillar

## Summary

Put an internal representation between the builders and the output, so codescan stops writing
`go-openapi/spec` structs directly. This is the **largest internal change the project has attempted**,
and it is the single hard gate for OAI v3 (Stream 10): without it, codescan cannot render into a v3
shape at all.

Two things it buys, in order of how sure we are of them:

1. **The `spec` dependency leaves the core.** Only a renderer imports it. The scanner and builders stop
   caring which spec library, or which version of it, exists downstream.
2. **One scan, many outputs.** OAI 2.0 and 3.x from the same build; and, if we ever want it, protobuf /
   JSON Schema / markdown as renderers rather than as forks.

This document is **shaping**. It has no task list and should not grow one until §Decisions is settled.
Fred's standing preference applies with full force here: clear domain modelling before implementation,
type design iterated across rounds, and a real debate before committing to the architecture.

## Context

### Where the material already is

- `ramblings/builder-renderer-separation.md` (2026-03-25) — the **primary source**. Contains the
  build-vs-render argument, a full IR sketch (`Model` / `Field` / `TypeRef` / `Validations` /
  `Operation` / `Parameter` / `Response` / `SourcePos`), the renderer interface, an IR→format mapping
  table for OAI 2.0 / 3.1 / protobuf, five stated risks with mitigations, and — importantly — a later
  *revision* section that repositions the IR against the go-openapi **JSON Document** core.
- `ramblings/index-builder-statefulness.md` (2026-03-25) — the discovery machinery this work also
  retires. See §Stateless discovery.
- `archive/pull-based-builder.md` — `SchemaCache` and the per-package declaration index, parked. The cache is
  not a separate idea from the IR; it is where the IR graph gets built once instead of per reference.
- `ramblings/vision.md` — the original v2 framing, superseded for *packaging* only.

### What has changed since the ramblings were written

Five months of work moved the ground under them, all of it favourably:

- The **package split** (Stream 1) and the **golden suite** (Stream 2) exist. The rambling's "testing
  simplifies dramatically" argument is now measurable against something.
- The **grammar parser** (Stream 3) replaced the regex engine, so the parse side is no longer entangled
  with the build side.
- The **declaration contract** landed (PR #90). Builders ask the scanner for a declaration through one
  contract instead of reaching into an index.
- **Provenance** (`OnProvenance`, `SourcePos` in the sketch) already exists and is consumed by the TUI.
  The rambling proposes it as an IR design principle; half of it is built.
- The **`spec` dependency is no longer the only downstream concern** — `cmd/genspec`'s `-validate`
  pulls `go-openapi/validate`, and the playground consumes a JSON envelope.

### The honest scheduling problem

The published roadmap promises **v0.37.0 in September 2026**: "decouple from `Spec`, go1.26+, internal
document model". That is roughly a month, for the biggest internal refactor in the project's history,
following a release week. Naming this now rather than at the September retrospective:

- The published date is a **projection on a maintainers' page**, not a contract with users. It can move.
- The pillar is genuinely severable — §Shape below sketches how — so *something* real can ship in
  v0.37.0 without the whole pillar landing.
- What must not happen is a half-migration living on a branch for a quarter, with the golden suite
  straddling two representations. That is the failure mode this project has explicitly learned to avoid
  (`ramblings/cascading-refactor-antipattern.md`, and the chunking-risk feedback).

## Decisions — settle these before any code

> These are the debate, in dependency order. Nothing below §D4 can be answered before §D1.

1. 🔍 **D1 — Is the go-openapi JSON Document core real, and on what timeline?**
   The rambling's revised architecture has renderers emit a `*jsondoc.Document`, not JSON bytes, with
   the Document library owning immutability, indexing, JSONPointer/JSONPath and serialization. **If
   that core does not exist yet, this plan's target moves**: the OAI 2.0 renderer would emit
   `*spec.Swagger` (keeping today's public API exactly) and the Document becomes a later renderer.
   That is arguably the better order anyway — it makes the first increment non-breaking by
   construction — but it has to be a decision, not a discovery. **This is the one genuinely external
   dependency, and the roadmap already flags upstream v3 packages as "gating but outside our control".**

2. 🔍 **D2 — Does the IR carry OAI 2.0's shape or the richer 3.x shape?**
   The rambling says: always produce the richer form (`RequestBody`, cookie params), and let the 2.0
   renderer degrade it. That is right in principle and expensive in practice — it means the 2.0
   renderer, which must reproduce today's output *byte for byte*, is doing a downgrade rather than a
   transcription. The cheaper first cut is an IR shaped like what we already emit, widened later.
   Trade-off: the cheap cut risks baking 2.0 assumptions into the IR, which is the whole thing we are
   trying to avoid. **Recommend debating this one first among the modelling questions** — it decides
   how much of the rambling's sketch survives.

3. 🔍 **D3 — Does the IR replace `ifaces.SwaggerTypable` and the taggers, or coexist with them?**
   The rambling says `SwaggerTypable` "disappears entirely" and `ValidationBuilder` collapses into a
   plain `Validations` struct. Memory already records the taggers machinery as v2-sunset. If that is
   the plan, then **this pillar is where they die**, and the interface segregation work deliberately
   skipped earlier was correctly skipped. Confirm, because it changes the size estimate a lot: it is
   the difference between "add a layer" and "remove two".

4. 🔍 **D4 — What is the first shippable increment?** See §Shape.

5. 🔍 **D5 — What is the evidence standard?** The project's own rule
   ([[feedback_golden_over_assertions]], [[feedback_schema_discovery_verify_with_witness]]): whole-spec
   goldens, A/B'd, plus a witness fixture — "no golden diff" alone is not proof. For a change this
   size, the bar should be **byte-identical output across the entire golden corpus** for the OAI 2.0
   renderer, with any divergence enumerated as a known-and-argued list rather than tolerated. The
   pattern exists already: `internal/integration/loader_agreement_test.go` compares marshalled
   documents and asserts known divergences. That is the harness this work should reuse rather than
   invent.

6. 🔍 **D6 — `go1.26+`?** The published roadmap couples it to this release. Confirm whether it is
   actually a prerequisite of the model work or an unrelated bump that got bundled into the same row.

## Shape — how this could ship without a big-bang branch

> Sketch, not a commitment. Its purpose is to make D4 answerable.

The severability argument: the IR can be introduced **behind** the existing public API, one production
site at a time, because `Run` keeps returning `*spec.Swagger` throughout. Nothing outside the module
observes the change until a second renderer appears.

1. **`ir` package, models only.** The rambling's own `Schema`/`API` layering (Open Questions, last
   bullet) says models are separable from operations. Models are also the largest, most-tested and
   most-quirk-laden surface — so doing them first is the honest test of the design rather than the easy
   half. The schema builder produces `ir.Model`; a 2.0 renderer turns it back into `spec.Schema`.
   Ships as a pure refactor, provable by D5's byte-identical bar.
2. **Operations, parameters, responses.** Same move, three more builders.
3. **Retire the fixpoint machinery** — see §Stateless discovery, which becomes natural here rather than
   being a separate project.
4. **A second renderer** — the first increment a *user* can see, and the point where the pillar starts
   paying for itself. Whether that is OAI 3.x or a JSON Document depends entirely on D1.

Only step 4 is release-visible. Steps 1–3 are internal, individually shippable, and individually
revertible — which is what makes the September date survivable even if the pillar is not finished:
**v0.37.0 can honestly ship "the internal model, models layer" without shipping a v3 document.**

## Stateless discovery — folded in here, requalified

**Requalified 2026-08-17 (Fred): this is a maintainability item, not a performance one.** It arrived as
part of the parsing-on-demand stream, whose performance case was closed by a different route — the
loader's per-package cuts and per-dependency source/export-data choice — leaving the residual prize
measured at ~0.03 s / 11 MB. The architectural argument never depended on the performance one, and it
belongs to this pillar because the IR is what makes it fall out rather than be built.

**What is actually there today** (verified against `master` `27cc6295`, unchanged since the rambling):

- `TypeIndex.ExtraModels` (`internal/scanner/index.go:93`) — discovered models, written **back into the
  scanner's index by the builders**, during Phase 2.
- `common.Builder.postDecls` (`internal/builders/common/builder.go:36`) — a *second*, parallel queue
  doing the same job for parameters and responses.
- `spec.Builder.joinExtraModels` / `buildDiscovered` (`internal/builders/spec/spec.go:311,1093`) — two
  fixpoint loops draining the two queues.
- `ScanCtx.AddDiscoveredModel` carries a comment about a "`Models`↔`ExtraModels` bouncing loop" it has
  to avoid. That comment is the design smell in one line.

**Why it is a maintainability problem.** The build phase mutates the scan phase's data structure, so
there is no clean phase boundary to reason about; two queues that do the same thing must be kept in
step by hand; and the fixpoint is reached by iterating to exhaustion rather than by construction.
Every schema-discovery change has to be argued against all three at once, which is precisely why those
changes carry a witness-fixture rule.

**What replaces it.** The `SchemaCache` from `archive/pull-based-builder.md` §1: build-once, reference-many,
keyed by (package path, type name), with a `pending` set for cycles. Encountering a field of type
`Owner` becomes `cache.Get(pkg, "Owner")` — built now, recursively, memoized. No queues, no write-back,
no fixpoint loop. Fred settled the framing on 2026-08-05: **the cache *is* how the fixpoint is reached
— recast, not removed.**

⚠️ **The one real hazard, already identified:** a cache that reaches the *same* fixpoint by a different
route is a refactor; one that reaches a *different* fixpoint is a behaviour change wearing a refactor's
clothes. Only a whole-document A/B over the golden corpus distinguishes them (D5).

📌 Also carried from the scanner stream: `Names()` / `DefKey()` still read `Comments` for the
`swagger:model` override, so a declaration whose syntax was *deferred* rather than absent would fall
back to its Go name. Harmless today — absence is permanent, never deferred — but it is the one place
laziness would bite, and this pillar is where that stops being hypothetical.

## Trajectory

> Deliberately short. This plan earns a task list after §Decisions, not before.

1. 🔍 **Settle D1** — the external dependency; everything else is downstream of the answer
2. 🔍 **Settle D2/D3** — the modelling debate proper, expected to take more than one round
3. 🔍 **Settle D4** — the first increment, informed by §Shape
4. 📝 **Write the type design down** as a reviewable document (a `features/` or workshop file), A/B'd
   against a handful of real fixtures before any production code
5. 📝 **Then, and only then**, a build plan

## Actions

### Before the next session on this

1. 🔍 **Answer D1 from outside this repo** — check the actual state of the go-openapi JSON Document
   core. It is the only question here whose answer is not in this repository, and it gates the target.
2. 🔍 **Re-read the two ramblings with five months of hindsight** and mark what did not survive. The
   loader-seam guess in `wasm-playground.md` is the precedent: several of its predictions were wrong in
   instructive ways, and saying so up front is what kept round 2 honest.
3. 📝 **Pick the witness corpus now, not later.** D5's bar is byte-identical output over the golden
   suite; knowing *which* fixtures exercise the quirkiest schema paths (allOf edges, aliases,
   discriminated families, cyclic `$ref`, name collisions) is what makes the first increment provable.

## Achievements

Nothing built — plan opened 2026-08-17 as the shaping document the roadmap said was missing.

## Appendix — risks and open points

- ⚠️ **The schedule is the biggest risk, not the design.** A month, post-release, for the project's
  largest internal change. §Shape exists specifically so that v0.37.0 can ship something true rather
  than something whole. Deciding *that* early is worth more than any amount of design velocity.
- **A half-migrated builder set is the specific failure mode.** Two representations straddling the
  golden suite, on a long-lived branch, is the cascading-refactor antipattern this project has already
  written up. Mitigation is in the shape: one builder at a time, each provable, each mergeable.
- **The IR is a superset and will be argued down.** Every "the IR carries X even though 2.0 can't
  express it" is a place someone will later ask why. The rambling's answer (superset, not intersection;
  renderer decides what to drop, with a diagnostic citing `Source`) is right and should be written into
  the package doc, not just this plan.
- **`InputSpec` does not go through the IR.** The rambling is explicit: the base spec is loaded as a
  document and merged at document level, never converted into IR. That is a real constraint on the
  renderer's interface and is easy to forget until it breaks the overlay feature.
- **This pillar deletes abstractions, which is where the surprises live.** `SwaggerTypable`, the
  taggers, `ValidationBuilder`, `postDecls`, `ExtraModels` — each removal is a place where some
  behaviour was encoded implicitly. Expect the golden corpus to catch things nobody documented.
