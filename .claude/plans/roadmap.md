# codescan — roadmap

Date: 2026-06-11
Status: living overview
Owner: Fred (sponsor) + agent collaborators
Audience: maintainers (private — lives under `.claude/plans/`)

This document tracks the parallel streams of work the project has
opened, so the multiple axes of progress can be reviewed at a glance
without re-discovering them from individual planning documents. It is
deliberately **light on detail** — each stream cross-links to the
deeper plans, workshop notes and design ramblings where the real
content lives.

When a stream lands, mark it ✅ and add the merge commit / PR.
When a stream changes shape, update its row; do not start a fresh
roadmap.

Status legend: ✅ done · 🔶 in progress · ⬜ not started · 🟡 design /
shaping · ⏸ paused.

---

## Overview — streams & features

Hierarchical map: each **stream** → its **features** (each links to its
`features/<slug>.md` detail file). Stream detail sections follow below.

- **Stream status:** ✅ done · 🔶 in progress · 🟡 shaping · ⬜ not started.
- **Feature/release markers:** ✅ shipped (with release) · 🔷 targeted release · ⬜ unscheduled.
- **Origins:** (i) vision · (ii) mid-build · (iii) triaged issue · (iv) deferred refinement.
- The flat feature catalog (alternate lens) lives in `forthcoming-features.md`.

**1. Disentanglement** ✅ (v0.33) — sub-package split (`.claude/CLAUDE.md`).

**2. Integration tests + goldens** ✅ — `internal/integration/`, `scantest.CompareOrDumpJSON`.
- ⬜ [Property-based Block generator](features/property-based-block-generator.md) · (i)

**3. Grammar-based parser** ✅ (v0.34) — lexer+grammar replaces regexp; build trackers in `archive/`.

**4. Single contract across builders** ✅ — alias-handling, `archive/observed-quirks.md`.

**5. genspec TUI** 🔶 — v0.36 `genspec-tui-linkage.md`, `project_genspec_tui` (memory).

**6. genspec Web UI (WASM)** 🟡 — v0.37 `wasm-playground.md`, `project_wasm_playground` (memory).

**7. Doc site (Hugo, GH Pages)** ✅ v1 (v0.35) / 🔶 next wave — next-wave items
(the W-series, with parked entries) are kept in `doc-site-wishlist.md`, **not
expanded here**. On the *published* roadmap we'll likely surface only the WASM
"playground UI" (Stream 6), nothing more granular.

**8. Core engine refactors (non-breaking)** ⬜ — v0.37 `ramblings/vision.md`. Pillars `V-scanner / V-builder / V-imodel` (sub-table below).
- ✅ v0.36 [After-declaration annotation comments](features/comment-source-filtering.md) · (i)
- ✅ v0.36 [Godoc-syntax filtering & idiom recomposition](features/godoc-filter.md) · (i)

**9. Wring out go-swagger backlog** 🔶 — `archive/backlog-go-swagger-20260608.md`.
- ✅ v0.35 [Prune unused models](features/prune-unused-models.md) · (iii) · #2639 (PR #50)
- ✅ v0.35 [Definition-name auto-disambiguation](features/name-identity-disambiguation.md) · (iii) · #1734
- ✅ v0.35 [Explicit additionalProperties control](features/additionalproperties-control.md) · (iii) · #2539/#3005
- ✅ v0.35 [Map additionalProperties for non-string keys](features/map-additionalproperties-keys.md) · (iii) · #2251
- ✅ v0.35 [Scanner robustness / fail-loud](features/fail-loud-diagnostics.md) · (iii) · #2886/#2874
- ✅ v0.35 [ExternalDocs on non-meta objects](features/externaldocs-non-meta.md) · (iii) · #2872/#2655
- ✅ v0.35 [Single-line comment as description](features/single-line-description.md) · (iii) · #2626
- ✅ v0.35 [Emit x-go-type](features/emit-x-go-type.md) · (iii) · #2924
- ✅ v0.36 [Naming from struct tags (form:, schema:)](features/naming-tags.md) · (iii) · #2912/#1391
- ✅ v0.36 [Response-level examples by mime](features/response-examples-by-mime.md) · (iii) · #2871
- ✅ v0.36 [Shared swagger:parameters / swagger:response](features/shared-parameters.md) · (iii) · #2632
- ✅ v0.36 [Inner markdown — `swagger:description \|` block scalar](features/inner-markdown.md) · (iii) · go-swagger#3211
- ✅ v0.36 [Discriminator subtype discovery](features/discriminator-subtype-discovery.md) · (iii) · #1913
- ⬜ v0.37 [Infer required from field shape](features/infer-required-from-shape.md) · (iii) · #3275 · TODO
- ⬜ v0.37 [Per-operation field views](features/per-operation-projections.md) · (iii) · #1992 · low
- ⬜ [Name-identity advanced](features/name-identity-advanced.md) · (iv) · low
- ⬜ [withPatternProperties auto-inference](features/pattern-properties-inference.md) · (iv) · low

**10. OAI v3 support** ⬜ — `vision.md`; gated on V-imodel.
- 🔷 v0.39 [Godoc-identifier prefix on swagger:operation](features/godoc-identifier-prefix.md) · (i)

**11. LSP & IDE** ⬜ — `vision.md`; PR-stunt sibling to Stream 6.
- 🔷 v0.38 [Token-level YAML positions](features/yaml-token-positions.md) · (i)
- 🔷 v0.38 [Column precision beyond ASCII](features/column-precision-unicode.md) · (i)

**Cross-cutting features** — no dedicated stream (origins i/ii); ride whichever stream next touches their seam.
- ✅ v0.36 [Skip-jsonify-interfaces opt-out](features/skip-jsonify-interfaces.md) · (ii) · `Options.SkipJSONifyInterfaceMethods`
- ✅ v0.36 [swagger:description / swagger:title overrides](features/swagger-description-override.md) · (ii)
- ✅ v0.36 [DefaultAllOfForEmbeds](features/default-allof-for-embeds.md) · (ii) · `Options.DefaultAllOfForEmbeds`
- ⬜ v0.37 [Example value coercion (verification)](features/example-values.md) · (i) · #1268/#2246 · verify
- ⬜ v0.37 [Enum richer values](features/enum-richer-values.md) · (i) · TODO
- ⬜ v0.37 [DiscoverAliasesAsTypes](features/discover-aliases-as-types.md) · (ii)
- ⬜ [Bullet-list dash preservation](features/bullet-dash-preservation.md) · (iv)

---

## Two levels of increment

Two levels, plus a calendar projection:

- **Stream** — a long-lived initiative (the 11 above). Spans many releases;
  done only when all its features ship. Tracked here, stream-first.
- **Feature** — a discrete shippable increment (a knob / V-flag / annotation),
  one file under `features/`. Maps to one release. Flat catalog in
  `forthcoming-features.md`.
- **Release** — the calendar projection (published `maintainers/ROADMAP.md`
  timeline). A release is a time-slice of features across streams.

**Feature origins** (a feature's detail level follows its origin — short ≠ gap):

- **(i)** initial design vision · **(ii)** idea surfaced mid-build ·
  **(iii)** triaged go-swagger issue filed as a forthcoming feature
  (the only origin that maps cleanly to a stream → Stream 9) ·
  **(iv)** deferred refinement left on a shipped fix.

Origins (i)/(ii) features (enum, example, the knobs) have no dedicated stream —
they are core-product enhancements grouped under "Cross-cutting features" above.

---

## Principle — the non-breaking lens

Every change shipped from any stream is reviewed through this lens
before merge:

- **Surface-breaking changes** — public API shape changes, removed
  `Options` fields, dropped annotation keywords, grammar rules that
  reject previously-accepted input. **Avoid.** Always provide a
  deprecation path. If a true forced break ever surfaces, this
  principle will surface it as a real decision — not as a
  pre-committed plan.
- **Behaviour-changing bug fixes** — output-shape changes for cases
  that were demonstrably wrong. The golden diffs from PR #20
  (`fix/quirks`) and PR #32 (`fix/aliased-types`) are the canonical
  examples. **Fine, even encouraged.** The old output was broken in
  some other way; replacing one breakage with correctness is a win,
  not a regression. Document the change in release notes; don't
  pre-flag it as breaking.
- **The lexer-pattern is proven.** Better syntax / better semantics
  can coexist with older usage indefinitely. Users on the old shape
  are never forced to migrate. The grammar stream (Stream 3) +
  contract-unification (Stream 4) + alias-handling (Stream 4
  follow-on) all shipped this way.

**Implication: there is no "v2" forcing function on the horizon.**
The former "Road to v2" framing (from `ramblings/vision.md`) is
superseded for *packaging* — the deliverables it described (OAI v3,
LSP, on-demand scanner, internal IR) are still real targets but they
ship piecewise as v1 minors. Stream 8 is now "Core engine refactors
(non-breaking)". Stream 10 is "OAI v3 support", gated only on
V-imodel.

If we ever discover a forced surface break (something a flag, a
deprecation path, or the lexer-pattern can't absorb), v2 becomes a
real decision *at that point*. Until then we do not pre-commit to it.

**Scheduling corollary — no rendez-vous, no long chains.** A bundled
"v2" release would have created an artificial rendez-vous point that
all pillars converged into before any of them shipped. Dropping the
bundle dissolves that constraint. What's left is the *actual*
dependency graph: a handful of short chains, no long ones, no
central convergence. Anything not on a chain can ship in parallel
the moment it's ready.

The real dependencies (full list):

1. **V-scanner → V-builder** — pull-based builder needs the scanner's API.
2. **V-imodel → O-runv3 → O-keywords (per family)** — three stages,
   but each keyword family after `O-runv3` is its own independent ship.
3. **TUI LX-* → B-batch-verify-only and L-goto / L-refs** — the
   spec ↔ source linkage chain.
4. **W-loader-seam ⇄ V-scanner DemandLoader** — *shared abstraction*,
   not a chain. Designed once, used by both.
5. **D-site-bootstrap → D-playground (W-deploy)** — Stream 6's
   "earns its keep" condition.

Everything else — `D-site-bootstrap`, `B-triage-drop`, individual
V-flag rollouts, `V-imodel` itself, `V-scanner` itself, paste-mode
WASM, diagnostics-only LSP — has no upstream dependency and can
march in parallel.

---

## 1. Disentanglement — ✅ landed

**Goal.** Break the original monolithic `codescan` package into focused
internal sub-packages so that scanner / parsers / builders can evolve
independently and each have a documented contract.

**Outcome.** Layout described in `.claude/CLAUDE.md`. Three layers
(`internal/scanner/`, `internal/parsers/`, `internal/builders/`) with a
thin `internal/ifaces/` glue, classification helpers in
`internal/parsers/classify/`, and test-only helpers in
`internal/scantest/` that never re-enter production code.

**Why it mattered.** The package split + a real golden suite produced
a "step-change in session efficiency" (memory `project_refactor_golden_payoff`)
— follow-on work moved much faster because each sub-package has a
clear blast radius.

**Pattern to replicate for v2.** Same split discipline: pull every
v2 boundary into its own sub-package before writing the code that
crosses it. Memory entry tags this as the v2 multiplier.

---

## 2. Integration tests + golden output — ✅ landed

**Goal.** Black-box regression coverage for every fixture tree we care
about, comparable across refactors via `UPDATE_GOLDEN=1`.

**Outcome.** `internal/integration/` runs against
`fixtures/integration/golden/*.json`. `internal/scantest/golden.go`
exposes `CompareOrDumpJSON` so any test in the repo can opt into the
same protocol. Fixture corpora live under `fixtures/{goparsing,
enhancements, bugs, integration}/`.

**Why it mattered.** Every quirk fix and grammar migration cycle below
relied on this. No silent semantic drift can land without flagging a
golden diff.

**Carry-over rule.** Schema-discovery refactors must land with a
dedicated witness fixture (memory `feedback_schema_discovery_verify_with_witness`).
"No golden diff" alone is not proof of safety.

---

## 3. Grammar-based parser — ✅ landed (Stream M merged 2026-06-03)

**Goal.** Drop regex-based comment-block decoding in favour of a
hand-rolled lexer + grammar that produces typed `Block` /
`Property()` iterators, with:

- predictable parsing behaviour,
- a diagnostics surface (file:line:col) usable later by LSP,
- explicit sub-language seams for YAML / parameters / meta / enum,
- generated docs straight from the grammar.

**Outcome.** Plan landed via Stream M (memory
`project_stream_m_grammar2_merge`) — 21 squashed commits on master.
Benchmarks: roughly 2× faster, many times less memory than the regex
engine. The whole P5/P6/P7 plan (architecture +
tasks + workshops) is now historical; references kept under
`grammar-parser-architecture.md`, `grammar-parser-tasks.md`,
`grammar/00–60/*` and `stream-M-grammar2-merge-readiness.md`.

**LSP-diagnostics constraint** (memory
`project_lsp_diagnostics_target`): the grammar2 lexer must continue
to preserve per-line `file:line:col`. Never replace `Preprocess` with
`CommentGroup.Text()`. Diagnostic positions are the contract that
downstream LSP work depends on.

---

## 4. Single contract across builders — ✅ landed

**Goal.** Reconcile the many dissonant behaviours across schema /
parameters / responses (annotation overrides, special stdlib types,
alias dispatch, embed semantics) under one documented contract per
package.

**Outcome.** Five waves:

1. `feat/quirks` (PR #20) — initial regularisation: enum coercion,
   `Q26` TOS/Schemes terminator, `Q28` securityDefinitions
   indent + dedup, `Q29` case-insensitive `in:` comparison, `Q9`
   `swagger:name` verbatim contract.
2. `fix/aliased-types` (PR #32, merged 2026-06-11) — alias-handling
   stream: annotation gates first-class alias identity at use sites in
   schema, parameters and responses, with full contract sections in
   the builder READMEs (`§alias-handling`). Closed Q3, Q7, Q8, Q12,
   Q13 and the stdlib-recognizer asymmetry. Phase D scrub workflow
   (memory `feedback_phase_d_scrub_workflow`) — promote rules to
   README, then scrub internal vocabulary, then squash.
3. Audit: every entry in `archive/observed-quirks.md` is resolved, reframed,
   or closed-no-action; every phase in `fix-quirks.md` is ✅.

**Surviving deferrals.** *(2026-07-30: all six D-entries verified closed; the
register moved to `archive/`. Live quirks now in `quirks-open.md`.)*
`archive/deferred-quirks.md` — D1 (alias-expand
parameters body/query semantic), D2 (embed-as-allOf vs embed-as-
inline asymmetry on named interfaces), D3 (`swagger:strfmt` +
`swagger:model` named-strfmt inconsistency). All routed into v2
design or into the `forthcoming-features.md` flags catalogue.

---

## 5. genspec TUI — 🔶 in progress (`feat/genspec-tui`)

**Goal.** A shippable bubbletea TUI (`cmd/genspec-tui`) that
live-renders the spec from source on every save, with diagnostics,
spec ↔ code navigation, and a packaging mode for repro reports.
Audience: maintainers and contributors.

**Status.** Phase 1 of the wider WASM-playground vision (memory
`project_genspec_tui`). **Linkage chain COMPLETE 2026-07-30** — Phases A–D plus
the S7–S13 series and its three spun-out backlog items, all on
`feat/genspec-tui` (22 commits, unpushed, awaiting review). The UX-polish
**chrome** backlog is mostly done on top of it — help overlay + header chip,
diagnostics-pane consistency, spec-pane and Go syntax highlighting (no new
dependencies: both reuse a lexer pass we already ran), grammar-keyword
highlighting inside annotated comments, and the scanner's diagnostics drawn on
the token they name. Remaining stream-5 work: guarded reload, adjustable
splits, light/dark theme, then Repro-pack and Map-vis.

### Open sub-items

| Tag | Item | Status | Notes |
|-----|------|--------|-------|
| LX-spec | Phase A — spec-side offset index, JSON / YAML pointer surfaces | ✅ | `genspec-tui-linkage-build.md`; built on the `go-openapi/core/json` lexers (no `GOEXPERIMENT`) |
| LX-prov | Phase B — codescan-side provenance callback + anchor wiring | ✅ | `Provenance` + `OnProvenance` in scanner Options; all anchor kinds + keyword-line granularity |
| LX-join | Phase C — caller-owned join + `f` nav state in TUI | ✅ | bidirectional `f` (3 drivers); C7 tail = S7/S8, C8 tests = S9 |
| LX-refs | Phase D — `$ref` resolution + F3 / Shift-F3 jump nav | ✅ | S10–S12; `Enter` go-to-definition added to scope |
| UX-polish | UX polish pass over status line, scroll behaviour, panes | 🔶 | nav/visual half done (S8 roles, S13 gutter, spec line cursor, viewer key pass-through, rescan anchor); **chrome** mostly done — help overlay + header chip, spec-pane and Go syntax highlighting; still open: guarded reload, adjustable splits, light/dark theme. See `wasm-playground.md` §"Phase 1 backlog — TUI chrome" |
| Repro-pack | Reporting mode: pack repro artefacts for GitHub issues | ⬜ | uses `internal/scrambler/` (memory `feedback_verify_module_assumptions` — internal/ reachable from cmd module) |
| Map-vis | Visualise code ↔ spec node mapping | ⬜ | new sub-feature; depends on LX-join |
| Diag-nav | Diagnostic ↔ code navigation | ✅ | landed as the `followDiag` driver (diag pane `f`) |

**Decisions — both settled** (carried from
`genspec-tui-linkage-build.md`):

- ~~**D1** — go1.26 for the TUI module so `jsontext` is stdlib.~~ **MOOT**
  (`ca0653c`): the `go-openapi/core/json` lexer swap removed the
  `GOEXPERIMENT=jsonv2` dependency entirely. No toolchain pin, no
  release-pipeline constraint; the TUI module keeps `go 1.25.8`.
- ✅ **D2** — `Provenance` lives in `internal/scanner` beside `Options`,
  public via `codescan.Options`, TUI-importable like `grammar.Diagnostic`.

**Suggested PR order.** PR-A and PR-B can land in either order
(parallel). PR-C joins them. PR-D last.

---

## 6. genspec Web UI (WASM playground) — 🟡 shaping

**Goal.** WASM build of the scanner core + a JS runner (e.g. Vite) for
the same three-zone UX as the TUI, published as a static asset.
Privacy invariant: no code or data leaves the user's browser.
Audience: end users and bug reporters.

**Honest framing.** Primarily a **PR stunt** — a visible, shareable
artefact that says "this project is alive and you can try it without
installing anything." That alone is real value for adoption. It
**earns its keep beyond PR** if it gets deployed as an actual
playground people use to:

1. evaluate codescan before adopting it ("does it parse my style?"),
2. file cleaner inbound issues — the "report issue" rail lets a user
   reproduce a problem in the playground, then submit a packaged
   repro (annotated source + diagnostics + spec output) straight to
   GitHub. Each cleaner report saves several round-trips on the
   backlog (Stream 9).

The second point is small-but-real — not a major lever, but worth
designing for from the start because it costs little extra and
materially improves the quality of new tickets entering the queue.

**Status.** Design rambling in `wasm-playground.md`; memory
`project_wasm_playground` summarises. The hard case for the
broader go-openapi WASM-playground pattern, because codescan needs
`go/types` and package/import resolution.

### Key design points (from the rambling)

- **Real coupling is `packages.Load`, not "WASM can't do I/O".**
  Everything below it (`go/parser`, `go/types`, scanner index,
  builders, `go-openapi/spec`, JSON marshaling) already cross-compiles
  to `js/wasm`.
- **`Loader` seam** — native build → `packages.Load`; wasm build →
  `go/parser` + `go/types` over in-memory source. Same abstraction as
  v2's `DemandLoader` (Stream 8).
- **Two senses of "imports":**
  - *Special-semantics* imports (`time`, `encoding/json`, `strfmt`)
    are recognised by name → no real types needed.
  - *Drill-down* imports → real type info needed. User's own types are
    in the editor; external drill-down via lazily-fetched, pre-published
    `gcexportdata` blobs (static `fetch()` is allowed in WASM).
- **The hard part is UX for multi-file + imports**, not the
  type-checking.

### Open sub-items

| Tag | Item | Status | Notes |
|-----|------|--------|-------|
| W-loader-seam | Introduce `Loader` seam in scanner (same shape as v2 `DemandLoader`) | ⬜ | becomes free once Stream 8 lands |
| W-wasm-build | `GOOS=js GOARCH=wasm` build of the scanner + runner | ⬜ | depends on W-loader-seam |
| W-ux | Three-zone web UX (source / spec / diagnostics) | ⬜ | mirrors TUI; reuse layout decisions |
| W-import-fetch | Lazy `gcexportdata` blob fetch for known stdlib + go-openapi imports | ⬜ | publishes blobs on the doc-site (Stream 7) |
| W-degrade | Diagnostic-driven graceful degrade for arbitrary imports | ⬜ | falls back to "paste source / use TUI" |
| W-privacy | Privacy invariant — no upload, no telemetry | ⬜ | enforced by static-only deploy |
| W-report | "Report issue" rail — pack annotated source + diagnostics + spec output + browser metadata into a pre-filled GitHub issue link | ⬜ | mirrors Stream 5 / Repro-pack; feeds cleaner reports into Stream 9's intake |
| W-deploy | Deploy as a real playground page on the doc site (Stream 7) | ⬜ | this is the "earns its keep" condition; without deploy it's only a PR stunt |

---

## 7. Doc site (Hugo, GitHub Pages) — ✅ v1 built · 🔶 next wave

**Goal.** A dedicated codescan doc site that supersedes the "generate
spec" section of the go-swagger doc site. Hosts reference docs for
the comment grammar, the rules / knobs that govern code-construct
interpretation (aliased types, special types, `$ref` behaviour, etc.),
and serves the Web UI as a static playground.

**Status.** The v1 site shipped with **v0.35 (June 2026)** — Hugo +
hugo-relearn under `docs/doc-site/`, GH Pages workflow live, content
mounted from test-covered examples in `docs/examples/`. Populated
sections: `getting-started`, `tutorials` (×11), `shaping-the-output`
(×14 knob pages), `maintainers` (grammar / annotations / keywords /
sub-languages), `usage`, `annotation-index`. What remains is the
**next wave**, not the bootstrap.

### Built (v1)

| Tag | Item | Status | Where |
|-----|------|--------|-------|
| D-site-bootstrap | Hugo scaffold + GH Pages workflow | ✅ | `docs/doc-site/`, `hack/doc-site/hugo/` |
| D-grammar-ref | Grammar reference + EBNF surface | ✅ | `maintainers/grammar.md` (hand-maintained; visual render = W15) |
| D-annot-ref | Annotation / keyword reference | ✅ | `annotation-index/`, `maintainers/{annotations,keywords}.md` (hand-maintained; generate-from-grammar = W9) |
| D-contracts | Builder contract pages (alias-handling, embeds, specials) | ✅ | `shaping-the-output/*` |
| D-knobs | `Options` knobs reference | ✅ | `shaping-the-output/*` (one page per knob) |
| D-examples | Worked examples for supported constructs | ✅ | `tutorials/*` + `docs/examples/` (test-covered) |
| D-playground | Embed the WASM Web UI as a static playground | ⬜ | = W11; depends on Stream 6 |

### Next wave

The W-series (priority markers, sub-tiers, first-wave sequencing) lives in
**`doc-site-wishlist.md`** — deliberately **not surfaced in the Overview**; the
doc site is tracked here only as a stream, not feature-by-feature. Parked during
the 2026-06-23 groom: **W9** (generate-reference-tables-from-grammar — large
digression; W14 covers the cheap part) and the **W15 railroad-diagram tier**
(needs a railroad generator; the prettified-EBNF tier W15a stays). Marquee bets
**W11 → W10** ride the genspec-tui/WASM stream (Stream 6). Satellites:
`doc-site-reference.md` (build notes), `archive/doc-site-quirks.md` (scanner bugs the docs
revealed), `doc-site-backlog-alignment.md`.

[maintainers-doc-site]: https://go-openapi.github.io/doc-site/maintainers/index.html

---

## 8. Core engine refactors (non-breaking) — ⬜ not started

**Goal.** A small set of high-impact internal refactors that don't
change the public API or annotation grammar but materially improve
performance, internal architecture, and extensibility. Each pillar
ships as a v1 minor release. Originally framed as "Road to v2";
that bundling is superseded by the non-breaking-lens principle —
see the top-of-roadmap section.

**Features** (see Overview, origin i, 🔷 v0.36 — the clean-godoc cluster):
[After-declaration annotation comments](features/comment-source-filtering.md) ·
[Godoc-syntax filtering & idiom recomposition](features/godoc-filter.md). The `V-*`
pillars (scanner, builder, imodel) are pillar-level, tracked in the sub-table
below — not feature files.

**Note on risk.** Each pillar is *internally* high risk — they touch
core systems (scanner, builder dispatch, IR). The non-breaking
guarantee is about user-facing surface, not internal churn. Stream 2
(goldens) + Stream 4 (contract docs) are the safety net that makes
this tractable. Schema-discovery refactors must land with dedicated
witness fixtures (memory `feedback_schema_discovery_verify_with_witness`).

### Pillars

| Tag | Pillar | Description / Unlock |
|-----|--------|----------------------|
| V-imodel | Internal IR / model | Parks the dependency on `go-openapi/spec`. The single hard gate for Stream 10 — shipping just this opens OAI v3 incremental work. API stays non-breaking via conversion at the public boundary. Detailed design: `ramblings/builder-renderer-separation.md` |
| V-scanner | On-demand scanner | Fully parses types on demand; AST otherwise; cache of parsed reachable constructs. Benchmarks predict ~3× perf, drastic peak-memory drop. Doesn't unlock any mainline feature, but **closes the perf/OOM cluster of Stream 9 tickets** (currently drowning in the 236-noise). Detailed design: `ramblings/index-builder-statefulness.md` |
| V-builder | Pull-based spec builder | Leverages the on-demand scanner to pull packages to parsing. Cleaner spec-side index for Streams 5/11. Hard depends on V-scanner |
| V-cli | Standalone CLI | Sibling to `cmd/genspec-tui`. Small, opportunistic — ship when someone needs it |

### Things that used to live here, moved out

- **V-flags** — decoupled. Each parked flag in
  `forthcoming-features.md` §3.x is a *delivery mode* driven by
  Stream 9 batch triage: cluster of "users want X by default"
  tickets → ship the flag (defaulting off) → close the cluster.
  Listed under Stream 9's sub-items. The proof point: every flag in
  §3.x can ship as `Options.X bool` defaulting false → never breaks.
- **V-lsp** — always was Stream 11.
- **Parked items mapped to other streams.** §3.1 / §4.1 → Stream 11
  LSP prerequisites. §3.x feature flags → Stream 9 delivery mode.
  §5.x test-infra items → live in their respective streams when
  consumed. The big "fold in at v2" table is gone — there is no v2 to
  fold into.

### Sequencing constraints

- **V-imodel can ship first.** It's the highest-leverage pillar (gates
  Stream 10 entirely) and is decoupled from V-scanner / V-builder.
- **V-scanner before V-builder.** The pull-based builder needs the
  on-demand scanner's API.
- **V-loader seam** (Stream 6) and **V-scanner DemandLoader** are the
  *same abstraction* — land it once, used by both.
- All four pillars can land independently as v1 minors. No bundling
  required.

---

## 9. Wring out go-swagger backlog — 🔶 in progress

**Goal.** Burn down the 236 open issues on go-swagger's "generate
spec" use case. Many are likely already fixed (4 streams above have
closed dozens of quirks); some need fresh fixtures.

**Features** (origin iii — full list with links in the Overview). 8 landed at
v0.35 (prune, name-identity, additionalProperties ×2, fail-loud, externalDocs,
single-line, x-go-type); open: naming tags, response examples, shared params,
discriminator discovery, infer-required, projections + two deferred (iv) tails
(name-identity advanced, withPatternProperties).

**Delivery shape.** Stream 9 is a single stream on the roadmap but
ships in **smaller chunks** — each batch is its own delivery
(release notes, closed issues, screenshots attached) rather than
one monolithic burn-down. We accumulate maintainer / community
signal pass after pass, not at the end.

### Method

- **Triage table.** Landed at `.claude/plans/backlog-go-swagger-20260608.md`
  — 236 issues, one row each, with a summary column, an **Example?**
  marker, and a per-issue **Status** column (legend in the table
  intro). Issues whose body embeds an example spec or Go snippet
  have the snippet reproduced in the in-document **Appendix —
  embedded examples** so a repro can be lifted into a TUI buffer
  without leaving the file. Source JSON: `issues-spec-20260608.json`.
- **Triage drop pass first.** Likely a large fraction of the 236 are
  undetailed, unargumented, or stale — close-as-unactionable in one
  visible sweep before partitioning. The actionable remainder is
  what gets batched.
- **Batch triage** — *not done yet*. Partition the actionable
  remainder by **topic-cluster** (not by severity). Each batch then
  becomes a forcing function for the right code path:
  - "Already-fixed by recent waves" cluster → verify-only batch,
    cheapest first delivery, harvests credit from Streams 3 + 4.
  - "Perf / OOM" cluster → forcing function for V-scanner (Stream 8).
  - "Alias / stdlib-noise" cluster → forcing function for a specific
    V-flag (see V-flag rollouts below).
  - "Docs / UX confusion" cluster → forcing function for Stream 7.
  - "OAI v3 need" cluster → feeds into Stream 10.
- **Fast lane.** TUI verification per issue: lift the appendix
  snippet (or paste from the original ticket if not appendixed),
  open under `genspec-tui` (Stream 5), screenshot the now-correct
  spec, attach to the issue as proof of fix, close.
- **No mandatory fixture.** Don't generate a fresh `fixtures/bugs/`
  entry per closed issue. Only when:
  - The issue exposes a behaviour not already covered by an existing
    enhancement / quirk fixture, OR
  - The fix needed a code change in this stream.
- **Per-batch cadence.** Each batch concludes with a release note
  enumerating closed issues + screenshots + any fixture additions.
  Between batches: merge / housekeeping interval, reassess what the
  batch revealed, re-prioritise the remaining queue if needed.

### V-flag rollouts as a delivery mode

The parked flags in `forthcoming-features.md` §3.x are now driven by
Stream 9 batch triage. Each flag is `Options.X bool` defaulting
false — non-breaking by construction. When triage surfaces a
cluster of "users want X by default" tickets, ship the corresponding
flag as a v1 minor, then close the cluster.

| Flag (parked) | Cluster it closes | Origin |
|---------------|-------------------|--------|
| ✅ `Options.SkipJSONifyInterfaceMethods` (shipped v0.36, `54cf1fd`) | Interface-method name-mangling complaints | `forthcoming-features.md` §3.3 (Q9 fallout) |
| `Options.DescriptionOverrides` map (per-decl `swagger:description`) | Stdlib godoc noise on `time.Time` / `error` / etc. | §3.4 (Q30) |
| ✅ `Options.DefaultAllOfForEmbeds` (shipped v0.36, `0d550cf`) | Client-generator users wanting allOf composition shape from embeds | §3.5 (Q-D close-out) |
| `Options.DiscoverAliasesAsTypes` | Users preferring pre-R6 alias discovery (every alias surfaces as a definition) | §3.7 (Q-E close-out) |

Each is XS, the design work is already documented in
`forthcoming-features.md`, the implementation is "add a knob, gate
the existing-vs-alternative code path on it".

### Open sub-items

| Tag | Item | Status | Notes |
|-----|------|--------|-------|
| B-table | Triage table | ✅ | `archive/backlog-go-swagger-20260608.md` — 236 issues + embedded-example appendix |
| B-status-col | Add per-issue status column (open / verify-only / repro-needed / fixed / wont-fix / dup) | ✅ | landed 2026-06-11 with the legend in the table intro |
| B-triage-drop | Sweep-close issues lacking sufficient detail / stale; first visible delivery, no code | ⬜ | cheapest ship; shrinks actionable remainder |
| B-triage-batches | Partition actionable remainder into topic-clusters | ⬜ | prerequisite for every B-batch-N; uses the Status column; output is the cluster list checked into the same backlog file |
| B-batch-verify-only | Verify-only cluster — issues already fixed by Streams 3 + 4 + alias-handling, screenshot-and-close | ⬜ | gated on B-triage-batches and Stream 5 LX-*; harvests credit from recent waves |
| B-batch-N (per cluster) | One batch per topic-cluster, one delivery each | ⬜ | each is its own ship; cadence informed by what the previous batch revealed; some batches trigger a V-flag rollout |
| B-flag-rollouts | Ship V-flags as cluster batches close (one minor per flag) | ⬜ | see "V-flag rollouts" table above |
| B-release-notes | Release-note generator from closed-issue list | ⬜ | nice-to-have; per-batch summary autogenerated from the table |

**Dependency.** Most batches benefit from the TUI's repro-pack mode
(Stream 5 / Repro-pack) so users + maintainers exchange clean
artefacts.

---

## 10. OAI v3 support — ⬜ not started

**Goal.** Full OpenAPI v3 surface, shipped keyword family at a time
as **v1 minor releases**. Originally listed in `ramblings/vision.md`
as the headline v2 objective; now decoupled from v2 per the
non-breaking-lens principle.

**Features** (see Overview). [Godoc-identifier prefix on `swagger:operation`](features/godoc-identifier-prefix.md)
(C9; step toward `swagger:operation` ≈ `swagger:route`) · origin (i), 🔷 v0.39. The
OAI 3 `examples:` map shape is deferred here too (the schema-side `example:` lives
in the cross-cutting [example-values](features/example-values.md) verification).

### Dependencies

- **V-imodel (Stream 8) is the only hard prerequisite.** Without an
  internal IR, codescan can't render into a v3 shape. Everything
  else is sequencing, not gating.
- **Upstream go-openapi v3 spec packages** — gating but outside our
  control. Track release cadence; light a fire if needed.

### Delivery shape (non-breaking)

- **One keyword family per minor release** — `links`, `callbacks`,
  `requestBody` (multi-media), expanded examples, OAuth flows, etc.
  Each family is a self-contained ship.
- **Public API stays additive.** `codescan.RunV3(*Options) (*v3spec.Document, error)`
  (or similar) lives alongside the existing `Run`. No break.
- **Annotation surface is additive.** Either new `swagger:*` keywords
  or a new `openapi:` prefix family — decided per-keyword. Old
  annotations keep working unchanged.

### Scope notes

- OAI v3 doubles the annotation surface — significant grammar evolution
  ahead. Each keyword family wants its own workshop before shipping.
- The `openapi:` prefix discussion (`forthcoming-features.md` §3.6)
  becomes relevant; coexistence with `swagger:` is the default
  expectation.
- A single-source dual-emit mode (one annotated codebase → both v2
  and v3 specs) is a real UX question that needs early answering —
  but it can be flag-gated and answered before any keyword family
  ships.

### Open sub-items

| Tag | Item | Status | Notes |
|-----|------|--------|-------|
| O-upstream | Track go-openapi v3 spec packages | ⬜ | gating; outside our control |
| O-runv3 | Additive `codescan.RunV3` public entry point | ⬜ | requires V-imodel |
| O-prefix | `openapi:` prefix style — decide co-existence with `swagger:` | ⬜ | additive surface decision |
| O-dual-emit | Single-source dual-emit (one annotated codebase → v2 + v3) | ⬜ | flag-gated; design call |
| O-keywords | New OAI 3 keyword surface, one family per minor release (`links`, `callbacks`, `requestBody`, `examples`, OAuth flows, …) | ⬜ | each family is its own ship; needs grammar update + workshops |
| O-doc | OAI v3 reference doc on the doc site | ⬜ | depends on Stream 7 |

---

## 11. Language Server Protocol & IDE integrations — ⬜ not started

**Features** (see Overview). Prerequisites (origin i, 🔷 v0.38):
[Token-level YAML positions](features/yaml-token-positions.md) ·
[Column precision beyond ASCII](features/column-precision-unicode.md) — both
needed before per-token IDE highlighting is precise.

**Goal.** Ship a codescan Language Server speaking LSP so that
developers can, inside their IDE:

- see live diagnostics on `swagger:*` annotations (parse errors,
  context-invalid keywords, unknown references, duplicate keys,
  shape mismatches) — the same diagnostic surface the grammar parser
  and the TUI already produce;
- get keyword completion contextual to the annotation block they are
  in (`swagger:route` vs `swagger:operation` vs schema property
  vs YAML sub-block);
- get hover docs on every keyword (lifted straight from the
  grammar-derived reference produced for Stream 7's doc site);
- jump from an annotation in source to the corresponding spec node
  (and back), reusing the spec-side index from Stream 5 / TUI
  linkage work (`LX-spec`, `LX-prov`, `LX-join`, `LX-refs`);
- run a code-action set: "promote alias to `swagger:model`",
  "convert const list to typed enum", "extract response", etc.

Then publish thin client extensions for the popular IDEs — VS Code,
Neovim (native LSP), JetBrains (LSP4IJ), Helix, Zed — so the LSP
server is reachable from each developer's familiar editor without
re-implementation.

**Status.** Tracked as a v2 objective in `ramblings/vision.md`. No
code yet, but several Stream 3 / 5 / 7 design choices were
deliberately made with LSP in mind — they are listed under
"prerequisites already met" below.

### Prerequisites already met

These earlier choices were locked specifically so this stream
becomes cheap to start:

- **Grammar positions (Stream 3).** Every grammar token, block and
  diagnostic carries `file:line:col`. The lexer must continue to
  preserve per-line positions (memory `project_lsp_diagnostics_target`):
  never replace `Preprocess` with `CommentGroup.Text()`. The
  grammar's own diagnostic shape becomes the LSP diagnostic with a
  trivial mapping.
- **Diagnostic callback (Stream 5).** `Options.OnDiagnostic` is
  already the conduit the TUI uses for its diagnostics pane
  (b2574fe). The LSP server consumes the same callback.
- **Provenance callback (Stream 5 / `LX-prov`).** Same callback the
  TUI uses for spec ↔ source linkage; LSP go-to-definition and
  find-references reuse it.
- **Spec-side offset index (Stream 5 / `LX-spec`).** Pointer-under-
  cursor / line-under-pointer mapping; LSP reuses the same data
  structures.
- **Grammar-derived keyword tables (Stream 3).** Single source of
  truth for completion + hover content.

### Dependencies (outside this stream)

- **Stream 5 LX-* sub-items must land first.** The TUI is the test
  bed for spec ↔ source linkage; the same data structures and
  callbacks then back the LSP server. Don't ship LSP go-to-definition
  before the TUI has it.
- **Stream 7 doc site** for hover content. The hover surface is the
  same reference the doc site renders — generated once, consumed
  by both. LSP can ship without it (fall back to keyword name + EBNF
  rule) but is much more useful with full reference text.
- **`forthcoming-features.md` §4.1 — column precision under multi-byte
  runes.** LSP positions are zero-based line + UTF-16 character
  offset; the grammar's current byte-based column math overstates
  columns past the first multi-byte rune. Fix lives in `stripLine`
  (`internal/parsers/grammar/...`). Must land before the LSP
  becomes correct for non-ASCII content.
- **`forthcoming-features.md` §3.1 — `goccy/go-yaml` POC for
  token-level YAML positions.** Today the YAML sub-parser
  (`internal/parsers/yaml/`) gives coarse positions. LSP wants to
  highlight the specific key inside a 20-line embedded YAML block;
  that needs token-level positions. Q28's duplicate-mapping-key
  surface (memory `project_grammar_parser_migration`) is one
  immediate consumer.

### Open sub-items

| Tag | Item | Status | Notes |
|-----|------|--------|-------|
| L-server | `cmd/codescan-lsp` (or `go-openapi/codescan-lsp` repo) — LSP server skeleton (stdio + tcp); document sync; workspace folders | ⬜ | reuses scanner + grammar as a library |
| L-diag | Push diagnostics on document save / on change (debounced); reuse `OnDiagnostic`; map `grammar.Diagnostic` → `lsp.Diagnostic` | ⬜ | needs §4.1 column fix for correctness on non-ASCII |
| L-pos | Position translation layer (grammar `Pos` ↔ LSP `Position` UTF-16) | ⬜ | depends on §4.1 |
| L-yaml | YAML sub-block diagnostics with token positions inside the embedded block | ⬜ | depends on §3.1 (`goccy/go-yaml` POC) |
| L-goto | Go-to-definition: annotation site ↔ spec node ↔ Go type decl | ⬜ | depends on Stream 5 `LX-join` |
| L-refs | Find references: `$ref` resolution + reverse map | ⬜ | depends on Stream 5 `LX-refs` |
| L-complete | Completion provider: keyword catalog from grammar tables, contextual on block kind | ⬜ | needs context-aware grammar surface query |
| L-hover | Hover provider: keyword reference (rule + doc) | ⬜ | depends on Stream 7 reference generation; degrade to rule-only |
| L-outline | Document outline: tree of annotations in the file | ⬜ | trivial once grammar block iterator is exposed |
| L-actions | Code actions: promote alias to `swagger:model`, const list → typed enum, extract response, etc. | ⬜ | one per refactor; ship the framework first, populate incrementally |
| L-ide-vscode | VS Code extension (TypeScript thin client around the LSP server) | ⬜ | published to VS Code Marketplace |
| L-ide-nvim | Neovim LSP config snippet + lazy.nvim plugin | ⬜ | native LSP — config-only |
| L-ide-jetbrains | JetBrains plugin (LSP4IJ) | ⬜ | nice-to-have; LSP4IJ accepts generic LSP servers |
| L-ide-other | Helix / Zed config docs | ⬜ | docs-only |
| L-release | Release artefacts (binaries per OS/arch) on the codescan-lsp repo | ⬜ | GH releases mirroring codescan cadence |

### Sequencing constraints

- **L-pos before L-diag.** Otherwise LSP diagnostics are positionally
  wrong on any file with non-ASCII characters.
- **L-server before any of L-*.** The skeleton has to exist first.
- **Stream 5 `LX-*` before L-goto / L-refs.** Both depend on the
  same provenance data + spec-side index.
- **Stream 7 reference doc before L-hover (preferred).** L-hover can
  ship in a degraded "rule only" mode first.
- **VS Code extension first among IDE clients.** Largest audience;
  forces the LSP surface to face a real client early.

---


## Cross-stream dependencies

The dependency graph is between *sub-topics*, not between streams.
With the v2 rendez-vous dissolved, what remains is a handful of
short chains; everything not on a chain ships in parallel as soon as
it's ready.

### The five chains

```
1. Scanner pipeline
   V-scanner ──► V-builder

2. OAI v3 pipeline
   V-imodel ──► O-runv3 ──► O-keywords (links)
                       ├──► O-keywords (callbacks)
                       ├──► O-keywords (requestBody)
                       └──► O-keywords (OAuth flows, …)
              [each family is its own independent ship]

3. Linkage pipeline
   LX-spec + LX-prov ──► LX-join ──► LX-refs
                            │
                            ├──► B-batch-verify-only  (Stream 9)
                            ├──► L-goto                (Stream 11)
                            └──► L-refs                (Stream 11)

4. Loader seam (shared abstraction, not a chain)
   W-loader-seam  ≡  V-scanner DemandLoader
   [designed once; used by Streams 6 and 8]

5. Doc-site → playground
   D-site-bootstrap ──► D-playground (W-deploy)
                   └──► L-hover content (degradable)
```

### Parallel-ready (no upstream blockers, ship the moment capacity allows)

- `D-site-bootstrap` (Stream 7) — leverage multiplier for everything
- `B-triage-drop` (Stream 9) — first visible delivery, no code
- `V-imodel` (Stream 8) — opens the OAI v3 valve
- `V-scanner` (Stream 8) — closes the perf/OOM backlog cluster
- Each individual V-flag rollout (Stream 9 delivery mode)
- WASM paste-mode (`W-wasm-build` + `W-ux`) — PR-stunt MVP
- `W-report` (Stream 6) — designable today against Stream 5 Repro-pack
- Diagnostics-only LSP (`L-server` + `L-diag` + §4.1) — if maintenance
  bandwidth allows; otherwise hold

### Minor / opportunistic edges

- **6 → 9** the WASM playground's `W-report` rail feeds cleaner
  inbound tickets into Stream 9's intake. Small lever, but worth
  designing in from the start because it's nearly free.
- **3 → everything downstream** the grammar parser's
  `file:line:col` positions are load-bearing for LSP correctness;
  never regress them (memory `project_lsp_diagnostics_target`).

---

## Process / hygiene

- This roadmap stays **private** under `.claude/plans/`. Per
  `.claude/.gitignore`, `plans/` is excluded — only `.claude/CLAUDE.md`
  and `.claude/rules/` are tracked.
- **Status updates.** When a stream lands, mark the row ✅ and add
  the merge commit / PR. When sub-items move, update the sub-table.
- **Don't expand items inline.** Each row should stay one line —
  the depth lives in the linked plan / rambling / workshop.
- **Phase D scrub workflow** (memory `feedback_phase_d_scrub_workflow`)
  applies to every multi-cycle feature stream. Promote rules to
  README contracts before scrubbing internal vocabulary; squash by
  logical unit before PR.
- **Schema-discovery refactors** (memory
  `feedback_schema_discovery_verify_with_witness`) — always land
  with a dedicated witness fixture, not just "no golden diff".

---

## Change history

| Date       | Change |
|------------|--------|
| 2026-06-22 | **Consolidation pass.** Archived ~14 done/migration plans to `archive/` (`archive/README.md` indexes them). Refreshed stale statuses: Stream 7 ⬜→**✅ v1 built (v0.35) / 🔶 next wave** (was "not started" while the site has 40+ pages); Stream 9 ⬜→**🔶** (V-flag features shipping). Introduced the two-level-increment framing + feature origins. |
| 2026-06-23 | **Per-feature split + projections.** Every feature now has its own file under `features/<slug>.md` (pure slugs; `prev:` frontmatter keeps the old `§N` traceable). This file's overview became a **hierarchical streams→features outline** (replaced the one-glance table) with release markers (✅ v0.35 / 🔷 vX / ⬜). `forthcoming-features.md` slimmed to a "where we stand" recap + origins legend + one flat catalog. Dropped the "cohort" concept and the interim Feature-index table. |
| 2026-06-23 | **Feature grooming.** Wrote groomed decisions into all 13 under-/needs-decision features; split **godoc-filter** out of comment-source-filtering (29 files now). Deferred low: pattern-properties-inference, name-identity-advanced, per-operation-projections; TODO: infer-required, enum-richer; verify-and-maybe-retire: example-values. |
| 2026-06-23 | **Doc-features triage.** Parked as unrealistic in `doc-site-wishlist.md`: **W9** (generate-reference-tables-from-grammar) and the **W15 railroad-diagram tier**; **versioned docs** out-of-scope. Briefly lifted the W-series into Stream 7's Overview, then **pulled it back out** (Fred): the doc site is tracked as a stream only, not feature-by-feature; the wishlist stays the sole detail home. The published roadmap will surface at most the WASM "playground UI". |
| 2026-06-23 | **v0.36 streak opened.** Base camp `feat/feature-v0.36`. First feature landed: ✅ **`NameFromTags`** (naming from struct tags — #2912/#1391; backlog #1391 closed in `archive/backlog-triaged-feature.md`). |
| 2026-06-23 | **v0.36 feature #2:** ✅ **Response-level `examples` by mime** on struct `swagger:response` (#2871) — new `examples` grammar keyword (CtxResponse) → `Response.examples`. Operation-YAML path already worked; struct path was the gap. Backlog #2871 ledger note updated. |
| 2026-06-11 | Initial roadmap drafted post-PR #32 (alias-handling close-out). Streams 1–4 marked ✅; Stream 5 🔶 with LX-spec/prov/join/refs sub-items; Stream 6 🟡; Streams 7–10 ⬜. |
| 2026-06-11 | Added Stream 11 — LSP & IDE integrations. Pulled together prerequisites already met (grammar positions, diagnostic / provenance callbacks, spec-side index, keyword tables), explicit dependencies on `forthcoming-features.md` §3.1 (token YAML positions) and §4.1 (multi-byte column precision), and L-* sub-items covering server skeleton, diagnostics, position translation, completion, hover, go-to / find-refs, code actions, and IDE clients (VS Code, Neovim, JetBrains, Helix, Zed). |
| 2026-06-11 | Backlog table (`archive/backlog-go-swagger-20260608.md`) gained a Status column with legend (⬜ open · 👀 verify-only · 🐞 repro-needed · 🛠 fix-needed · ✅ fixed · 🛑 wont-fix · ♻️ duplicate). Closed Stream 9 sub-item `B-table`; opened `B-status-col` (done as part of this same change). |
| 2026-06-11 | Stream 9 broken down into smaller chunks: each batch is its own delivery, not part of a monolithic burn-down. Added prerequisite sub-item `B-triage-batches` (partition into priority-ordered batches of ~50 — *not done yet*). `B-pass-1` / `B-pass-2..5` renamed to `B-batch-1` / `B-batch-2..N` and gated on `B-triage-batches`. |
| 2026-06-11 | Stream 6 reframed honestly: primarily a PR stunt with two genuine "earns its keep" conditions — (1) deploy as a real doc-site playground (`W-deploy`), (2) "report issue" rail for cleaner inbound tickets (`W-report`, feeds Stream 9). Added cross-stream note 6 → 9. |
| 2026-06-11 | **The big reframe.** Added "Principle — the non-breaking lens" as a top-level section. Stream 8 retitled "Road to v2" → "Core engine refactors (non-breaking)"; the V-flags pillar and the "parked items to fold in at v2" table are gone — V-flags moved to Stream 9 as a delivery mode driven by batch triage; the rest of the parked items live in their natural home streams. Stream 10 retitled "v2 + OAI v3 support" → "OAI v3 support"; gated only on V-imodel, ships keyword family per minor release with additive `RunV3` entry point. Stream 11 demoted to "PR-stunt sibling to Stream 6 — defer until maintenance bandwidth allows". `ramblings/vision.md` v2 framing flagged as superseded for packaging (deliverables remain). |
| 2026-06-11 | **Scheduling corollary captured.** Dropping the v2 rendez-vous dissolves the artificial dependency cluster; the real graph is sub-topic-level, not stream-level — five short chains, no long ones, no central convergence. Cross-stream dependencies section rewritten around (1) Scanner pipeline, (2) OAI v3 pipeline, (3) Linkage pipeline, (4) Loader seam shared abstraction, (5) Doc-site → playground. Everything not on a chain (`D-site-bootstrap`, `B-triage-drop`, `V-imodel`, `V-scanner`, V-flag rollouts, WASM paste-mode, diagnostics-only LSP) ships in parallel. |
