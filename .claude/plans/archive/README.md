# Archived planning docs

Build/migration/task-tracking docs for **landed** streams. Kept for provenance,
not active planning. (These are gitignored like all of `.claude/plans/`.)

Rule for what lives here vs. up one level: archive build plans, task trackers,
and merge checklists once their stream is ✅; keep specs, catalogs, and ramblings
as live reference even after the stream lands.

**Quirk / backlog registers (archived 2026-07-30).** The Q-, D- and F-series
registers and the four backlog-triage files moved here in one sweep. They are
**provenance, not status**: their per-entry status lines are demonstrably stale
(items logged in one register were fixed by a different stream, and nothing
updated the original — see `../quirks-open.md` §4 for the verified list). The
single live register is now **`../quirks-open.md`**; anything genuinely open was
carried across or moved into the owning feature doc. Q-numbers are still cited
from code comments, so the files keep their names.

| File | Stream / origin | Why archived |
|------|-----------------|--------------|
| `grammar-parser-tasks.md` | Stream 3 (grammar parser) | task tracker, all tasks done |
| `grammar-ebnf-draft.md` | Stream 3 | early EBNF draft, superseded by `../grammar/` spec |
| `stream-M-grammar2-merge-readiness.md` | Stream 3 (Stream M merge) | merge-readiness checklist, merged 2026-06-03 |
| `legacy-stop-points.md` | Stream 3 | regex→grammar migration stop-points, done |
| `p5-builder-migrations.md` | Stream 1/3 (P5 builders) | per-builder migration tracker, done |
| `p5.1a-items-walkthrough.md` | Stream 1/3 | items-builder migration walkthrough |
| `p5.1b-schema-walkthrough.md` | Stream 1/3 | schema-builder migration walkthrough |
| `dead-code-cleanup.md` | Stream 1 (disentanglement) | post-split dead-code sweep, done |
| `test-reorganization.md` | Stream 2 (tests/goldens) | test-tree reorg, done |
| `coverage-gap-analysis.md` | Stream 2 | one-off coverage audit, folded into golden harness |
| `regression-strategy.md` | Stream 2 | superseded by the golden-comparison harness |
| `scrambler-L0-build.md` | Stream 5 (TUI) | scrambler build plan, landed (`internal/scrambler`) |
| `fix-quirks.md` | Stream 4 (single contract) | PR #20 quirk fixes, landed |
| `observed-quirks.md` | Stream 4 / M6.5 | Q-series register (Q1–Q31); superseded by `../quirks-open.md`, kept for the Q-numbers cited in code |
| `deferred-quirks.md` | Stream 4 | D-series holding area; **all six entries are closed** — statuses inside are stale |
| `quirks-F-series-fix.md` | doc-site quirks fix branch | F1–F9, all ✅ |
| `doc-site-quirks.md` | Stream 10 (doc site) | scanner quirks surfaced while writing the docs; F1–F9, all ✅ |
| `backlog-go-swagger-20260608.md` | Stream 9 | go-swagger backlog triage index — 236/236 triaged, 0 open 🛠 |
| `backlog-triaged-ledger.md` | Stream 9 | ✅ fixed / works-as-designed / N-A rows (212) |
| `backlog-triaged-feature.md` | Stream 9 | 🛑 wont-fix-as-framed, tied to a recorded feature (9) |
| `backlog-triaged-wontfix.md` | Stream 9 | 🛑 pure wont-fix (5) |
| `backlog-triaged-poison.md` | Stream 9 | 🐞 poison queue — empty |
| `prune-unused-models.md` | Stream 9 §12 | `PruneUnusedModels`, merged PR #50 |

## Sweep of 2026-08-17 (pre-v0.36.4 review)

Twenty-one more plans came down in one pass, taking the live set from 38 to 17. Every one of these is
merged or superseded; none had an open action left. Grouped by why.

| File | Stream / origin | Why archived |
|------|-----------------|--------------|
| `on-demand-scanner.md` | Stream 8 `V-scanner` | merged PR #90; worklist closed, named deliverable retired by its own measurement. **Still cited** — `../pull-based-builder.md` points here for the measurements |
| `wasi-build.md` | Stream 6 round 1 | merged PR #79 — the artifact + `internal/packages` |
| `playground-ux.md` | Stream 6 round 1 | merged PR #79 — the front end |
| `wasm-playground.md` | Stream 6 origin | superseded 2026-08-02; several of its guesses did not survive contact, which is why it is worth keeping |
| `genspec-tui-linkage.md` | Stream 5 | design settled, linkage chain complete 2026-07-30 |
| `genspec-tui-linkage-build.md` | Stream 5 | the build tracker for the above |
| `tui-ux-enhancements.md` | Stream 5 | merged PR #92; only the deferred light/dark theme outlived it (tracked on the roadmap row, not here) |
| `tui-ux-round2.md` | Stream 5 | merged PR #100; drove three fixes into `go-openapi/validate` |
| `doc-site-cli.md` | Stream 7 | merged PR #115 — the CLI pages and the rebuilt `usage/` section |
| `doc-site-reference.md` | Stream 7 | the reference migration it planned is done; `usage/reference/` was deliberately deleted |
| `benchmarks-cleanup.md` | Stream 8 | merged PR #118 — corpus in-repo, one document, the loader and history stories |
| `grammar-parser-architecture.md` | Stream 3 | historical since Stream M; the live spec is `../grammar/` |
| `name-identity-cyclic-ref.md` | Stream 9 | shipped v0.35 through Stage 3 — `EmitHierarchicalNames` exists |
| `golden-unit-to-integration.md` | Stream 2 | done: **zero `CompareOrDumpJSON` callers remain under `internal/builders/`**, and the curated builder tests carry property assertions as decided |
| `additional-properties.md` | Stream 9 | phases 1–2c landed 2026-06-17 |
| `fail-loud-diagnostics.md` | Stream 9 | landed; `scan.degraded-load` is what emptied the backlog's poison queue |
| `swagger-omit.md` | Stream 9 | merged PR #67 |
| `typed-extensions.md` | Stream 4 | both rounds landed |
| `tabbed-examples.md` | Stream 7 | feature complete |
| `textmarshaler-toolchain-independence.md` | cross-cutting | merged; `.github/workflows/toolchain-independence.yml` is in master, which is the proof |
| `alias-override-symmetry.md` | Stream 4 / Q32 | the fixes are in master (`1b0e7b2f`, `ed9897cb`, `5ad1df18`) and Q32 is closed in the live register |

**Note on hashes.** Plans archived here cite commits by their **pre-rebase** hashes, which resolve to
nothing — the branches were rebased on merge. `../quirks-open.md` had 28 such citations rewritten to
their master equivalents on 2026-08-17; nobody has done the same for these files, and nobody should
bother. Match by commit subject if you need the real one.

## Sweep of 2026-08-17, part 2 — the "almost complete" six

Six more came down the same day, under a rule that did not exist before it: **a plan that is done except
for a few deferred or parked items is archived stamped *almost complete*, with its tail recapped in
`../backlog.md`.** Each of these opens with a `> [!NOTE]` block saying which it is and where its tail
went; the original document follows unchanged beneath.

| File | State | Tail |
|------|-------|------|
| `genspec-cli.md` | almost complete — merged PR #111 + #114 | `../backlog.md` §1 (5 items; the release `require` bump closed 2026-08-17) |
| `tui-round3.md` | almost complete — merged PR #108 + #110 | `../backlog.md` §2 |
| `pull-based-builder.md` | parked — nothing built, prerequisite met | `../backlog.md` §4; `SchemaCache` went to `../internal-document-model.md` |
| `go127-uuid.md` | almost complete — merged PR #77 | `../backlog.md` §4.2 (a go1.28 GA calendar trigger) |
| `loader-vs-gopackages.md` | almost complete — D1–D9 all resolved | `../backlog.md` §4.1 (the differential harness) |
| `doc-site-backlog-alignment.md` | **complete** — every row ✅, no tail | — |

`doc-site-cli.md` (archived earlier the same day) also contributed a tail — `../backlog.md` §5 and §6 —
including the one item worth knowing about: **the doc lane's changed-paths filter does not watch
`cmd/internal/` or `cmd/genspec-wasi`, so the published playground can silently lag its sources.**

## Still live, one level up

**Ten documents, down from 38 at the start of the day** (the two `validate-*-report.md` moved into
`../validate-fixtures/`, beside the harness they guide — valuable material for testing spec-validation
output from `genspec`/`genspec-tui`, but not active work):

`../roadmap.md` (the index) · `../backlog.md` (tails of finished plans) · `../forthcoming-features.md`
(feature catalog) · `../quirks-open.md` (the single quirk register) · `../doc-site-wishlist.md` (the
W-series) — five registers, each with one job.

Then five active plans: `../release-v0.36.4.md` · `../security.md` · `../internal-document-model.md`
(the v0.37.x pillar) · `../wasi-round2.md` (round-2 receptacle, gate cleared) ·
`../anonymizer-repro-tool.md` (scrambler — L0 built and unmerged, L2 and eight open questions ahead).

Directories: `../grammar/` (the language spec), `../features/` (13 **open** features; the 25 shipped ones
are in `../features/archive/` with their own README), `../ramblings/`, `../validate-fixtures/` (the
validate harness, now also holding the two `validate-*-report.md` written for `go-openapi/validate`).

Archived directories here: `workshops/` — all five workshops (alias handling ×4, enum) are complete.
