> [!NOTE]
> Last revision: 2026-08-17 (rev 2 — Fred's annotations: the `require` bump closed, wasi→cliconf scheduled
> for v0.37.0, Repro-pack promoted to a theme of its own)

# Backlog — the tails of finished work

## Summary

Where a deferred or parked item goes when its plan is otherwise done. Six plans reached "everything
shipped except a handful of small things"; keeping a whole document alive for three follow-ups made the
live plan set look busier than the work actually is. Those plans moved to `archive/` stamped
**almost complete**, and their tails are recapped here, one section per theme.

**This file is a recap, not a replacement.** Every item names the archived plan that holds the real
reasoning — the probe that found it, the alternative that was rejected, the reason it was deferred.
Read that before acting on anything here.

## Context

**What belongs here:** a leftover from a plan that is otherwise complete. Small, self-contained, with no
stream still pushing it.

**What does not:**

- Anything in an **active** plan — `security.md`, `release-v0.36.4.md`, `internal-document-model.md`,
  `wasi-round2.md`, `anonymizer-repro-tool.md` own their own open items. **One deliberate exception:**
  §3 (Repro-pack) gathers a live theme whose two surfaces were otherwise orphaned lines in two different
  documents. Where a theme spans plans and neither owns it, it lands here.
- **Scanner quirks** — `quirks-open.md` is the single register and stays that way.
- **Features** — `forthcoming-features.md` + `features/<slug>.md` are the catalog.
- **Doc-site wishes** — `doc-site-wishlist.md` holds the W-series. Only doc items *left over from a
  finished plan* land here (§6).
- **Stream-level work** — that is `roadmap.md`.

**Promotion rule.** If an item here grows a design question or a sequence of steps, it stops being a
backlog line and gets its own plan. Going the other way is what created this file.

Priority markers follow the house convention: ⚠️ needs attention · 📝 planned · 🔍 needs investigation ·
♥️ enhancement · ⛔ won't do. Within each section, higher priority first.

---

## 1. Command line & configuration

> From `archive/genspec-cli.md` (merged PR #111 + #114). The shared surface is
> `cmd/internal/{cliopts,cliconf}`; the commands are `genspec`, `genspec-tui`, `genspec-wasi`.

1. 🔷 **v0.37.0 — Wire `genspec-wasi` to `cliconf`.** Gate cleared 2026-08-16 when the TUI landed. wasi
   is the easy consumer — it already registers the whole shared surface and would just call
   `cliconf.Parse`, no koanf. Needs sections for its own flags (`format`, `output`, `indent`, `quiet`,
   `export-data`) and the same addressability guard.

2. 🔍 **`-color=always` cannot colour a run whose stdout is redirected.** `resolveColor` answers
   correctly, but `slogcolor` renders through `fatih/color`, whose package-level `NoColor` is decided
   once at init from whether **stdout** is a terminal — and diagnostics go to *stderr*. So
   `genspec ./... > spec.json` prints uncoloured diagnostics to a terminal that could show them. Fix is
   one line (`color.NoColor = !colorize` in `diagnostics.logger`), but it writes a process-global shared
   with anything else linking `fatih/color`, which is why Fred deferred it rather than taking it as a
   drive-by. Probed directly; `logger_test.go` covers the refusal half and says why.
3. ♥️ **Environment variables (`GENSPEC_*`)** — koanf's env provider onto the same seam, ~10 lines.
   Nobody has asked.
4. ♥️ **TOML config** — koanf's toml provider onto the same seam, ~10 lines. Nobody has asked.
   `cliconf` is parser-agnostic by construction (`cliconf.YAML` satisfies koanf's interface
   *structurally*), so a second format is genuinely additive.
5. ♥️ **Further short aliases (`-o`, `-q`)** — `-c` earned one by being typed often; `flag` lists an
   alias as its own help entry, so the rest are not obviously worth the noise.

✅ **Closed 2026-08-17 (Fred): the release `require` bump.** Both `cmd/genspec` and `cmd/genspec-tui`
carry `replace github.com/go-openapi/codescan => ../..`, and the CI release pipeline
(`bump-release-monorepo.yml`) does the `go mod` update — so this is the pipeline's job, not a manual
checklist step.
⚠️ **One thing for the dress rehearsal to prove, not a reopening:** `go help install` states that
`go install pkg@version` requires the target module's go.mod to contain **no `replace` or `exclude`
directives**. So the very directive that makes the command build in-tree is the one `@latest` refuses —
the pipeline must strip or rewrite it, and `release-v0.36.4.md` §4.5 is where that gets verified.

## 2. Terminal UI

> From `archive/tui-round3.md` (merged PR #108) and `archive/tui-ux-enhancements.md` (merged PR #92).

1. 📝 **Light/dark theme.** The single UX-chrome item deliberately left open when the rest of that round
   landed on 2026-08-07 — deferred, not dropped. It is also the last thing standing between Stream 5 and ✅ on the roadmap.
2. ♥️ **Migrate to bubbletea/v2.**
3. ♥️ **Further round-3 items** — the round was left open for more to be picked, and none were.

✅ **Closed 2026-08-17 (Fred):** the *relevance pass* on the run-cost card. The figures were judged
against real runs and **the card stands as shipped** — which also settles the three refinements it was
holding, including the sub-second duration reading ("1s" for everything from 1.0 s to 1.4 s). All three
are recorded as considered-and-declined in `archive/tui-round3.md` §5, not carried here.

## 3. Repro-pack — a theme of its own

> Cross-cutting (Fred, 2026-08-17): **one capability, two front ends.** Formerly split between the TUI's
> `Repro-pack` and the playground's `W-report`, which made them look like two items sharing a blocker.
> They are one item with two surfaces. Engine: `internal/scrambler`. Design:
> `anonymizer-repro-tool.md` §12, still a live plan.

**How this project builds a UX feature — the house method, worth stating once:** spike or prototype it
in the **TUI**; if it works, port it to the **Playground** (as Svelte/TS). The TUI is cheap to iterate in
and the port is the design review. This is why the two surfaces are one theme rather than two efforts,
and it decides the order below.

1. ⚠️ **Merge `feat/scrambler`.** One commit on a branch (`8f0b21a3`, L0 — the source-tree minimizer),
   and it is the sole blocker on everything else here. Highest ratio of unblocked-work to
   remaining-work anywhere in the plans.
2. 📝 **Repro-pack in the TUI** — pack repro artefacts (annotated source + diagnostics + spec output)
   for a GitHub issue. The spike, per the method above.
3. 📝 **`W-report` in the playground** — the same payload plus browser metadata, into a pre-filled
   GitHub issue link. The port. Tracked in `wasi-round2.md` §P1, which points here.
   It is one of the two conditions under which the playground stops being a PR stunt: cleaner inbound
   tickets feed Stream 9's intake.

## 4. Scanner, builders & loader

> From `archive/pull-based-builder.md` (parked, nothing built), `archive/go127-uuid.md` (merged PR #77)
> and `archive/loader-vs-gopackages.md` (D1–D9 all resolved).

1. 🔍 **A resolution-level differential harness for the two loaders.** The corpus A/B compares *emitted
   specs*, so it cannot see a resolution divergence at all — same patterns, both strategies, comparing
   `(PkgPath, GoFiles)` sets is what would actually guard the `go list` semantics we reproduce. Named in
   the loader comparison as "whatever we do, the corpus A/B cannot see any of it", and it is the one
   item there that outlived the fixes. Worth doing before anyone answers `wasi-round2.md` A1 (make
   `internal/packages` *the* loader), because that decision needs evidence this harness would produce.
2. 📝 **Retire the `//go:build go1.27` tags at go1.28 GA.** Repo policy is the two latest stable minors;
   when the supported set becomes 1.27 + 1.28 the stdlib-uuid fixture folds into the plain `go12x` style.
   A calendar trigger, not a decision.
3. 🔍 **Per-package declaration index**, retiring the linear `FindDecl`. Its stated case was performance
   and that case is gone; whether the linear walk is a *maintainability* problem in its own right is
   unmeasured. Route `FindDecl` / `DeclForType` / `FindComments` / `FindEnumValues` / `FileForPos`
   through it only if it earns it.
4. 🔍 **`V-builder`, the pull-based spec builder.** Parked 2026-08-07: unblocked (the declaration
   contract landed) but nothing downstream is asking. Its remaining architectural argument — a cleaner
   spec-side index for the TUI and LSP — is real but not urgent. ⚠️ Note the overlap: `SchemaCache`, the
   piece with the strongest independent argument, **moved to `internal-document-model.md`**, so what is
   left here is genuinely the residue.
5. ⚡ **A measurement gap: a large main module.** Both benchmark corpora have a small main module (19 and
   30 packages) against a large closure. A repo whose *main module* is thousands of packages is
   unrepresented — and it is the shape where items 3 and 4 would show up, if they ever do.
6. 📌 **Carried caveat, not a task.** `Names()` / `DefKey()` still read `Comments` for the
   `swagger:model` override, so a declaration whose syntax was *deferred* rather than absent would fall
   back to its Go name. Harmless today (absence is permanent, never deferred) and stated in its test.
   Belongs to whoever revives lazy materialisation — see `internal-document-model.md`.

## 5. CI

> From `archive/doc-site-cli.md` (merged PR #115), noticed while fixing something else.

1. ⚠️ **The doc lane's changed-paths filter does not watch `cmd/internal/` or `cmd/genspec-wasi`**, so
   **the published playground can silently lag its own sources**. Found during the `update-doc` fix and
   never addressed. The most consequential item in this file: it is a correctness gap in what gets
   published, not a polish item, and it fails silently by construction.

   Related and already tracked elsewhere: `wasi-round2.md` §21 (npm dev-dependency *majors* auto-merge
   because the shared workflow matches the group name by substring with no update-type condition —
   observed live on a TypeScript 7 bump) and §23 (turning npm version-update noise down).

## 6. Documentation site

> One leftover from `archive/doc-site-cli.md`. Everything else doc-site lives in `doc-site-wishlist.md`
> — including the `go:generate` page (W8c) and the `maintainers/commands.md` seam (folded into W20),
> both of which used to be restated here.

1. 🔍 **Options-reference table width.** Six columns render fine in the theme's scroll container, but the
   Effect column does most of the work. If it reads badly on a narrow screen, the fallback is a
   per-family split under `usage/options/` — **not** dropping a column, since the flag and the config key
   are the whole point of the move.

---

## Appendix — where each section came from

| Section | Source | Its state |
|---|---|---|
| 1 | `archive/genspec-cli.md` | almost complete — merged PR #111 + #114 |
| 2 | `archive/tui-round3.md`, `archive/tui-ux-enhancements.md` | almost complete — merged PR #108, #92 |
| 3 | `anonymizer-repro-tool.md` (live), `wasi-round2.md` §P1 | **not a tail** — a live theme, gathered here 2026-08-17 because both its surfaces were sitting as orphaned lines |
| 4 | `archive/pull-based-builder.md` | parked — nothing built, prerequisite met |
| 4 | `archive/go127-uuid.md` | almost complete — merged PR #77 |
| 4 | `archive/loader-vs-gopackages.md` | almost complete — D1–D9 resolved |
| 5, 6 | `archive/doc-site-cli.md` | almost complete — merged PR #115 |

**Audited and found clean** (archived with no tail, 2026-08-17): `doc-site-backlog-alignment.md` (every
row ✅, boundary advanced), `tui-ux-round2.md`, `typed-extensions.md`, `tabbed-examples.md`,
`swagger-omit.md`, `additional-properties.md`, `fail-loud-diagnostics.md`, `name-identity-cyclic-ref.md`,
`golden-unit-to-integration.md`, `textmarshaler-toolchain-independence.md`, `alias-override-symmetry.md`.
Four archived plans carry open-looking markers that are **stale rather than pending** — checked
individually: `wasi-build.md` and `playground-ux.md` (carried into `wasi-round2.md`),
`on-demand-scanner.md` (`hack/scanbench`, superseded by `internal/benchmarks`) and
`genspec-tui-linkage-build.md` (the viewer key-swallowing bug, since fixed as viewer key pass-through).
