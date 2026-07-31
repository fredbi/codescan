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

Live reference docs deliberately kept one level up: `grammar-parser-architecture.md`,
`../grammar/` (language spec), `observed-quirks.md` (quirks catalog),
`deferred-quirks.md` / `quirks-F-series-fix.md` (open tails), `ramblings/`, `workshops/`.
