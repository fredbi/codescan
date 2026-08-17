> [!NOTE]
> Last revision: 2026-08-16

# Documenting the CLI tools on the doc site

## Summary

The repo now ships three commands (`genspec`, `genspec-tui`, `genspec-wasi`) over one shared flag surface
(`cmd/internal/cliopts`) and one configuration-file contract (`cmd/internal/cliconf`). The doc site knows about
none of it except the TUI: there is no page for the headless CLI, and the options reference sits in
**Maintainers**, described only as Go struct fields — which is now one of *three* ways to set the same knob.

This pass opens a base camp to fix both: a new **Getting started → Usage as a headless CLI** page, and a revived
top-level **Usage** section holding the options reference, rewritten so every knob is shown in all three of its
spellings (library field, flag, config key) with the precedence between them stated once.

Base camp: worktree `.worktrees/doc/cli`, branch `doc-cli`, off master `52d7de00`.

## Context

**What exists in the code but not on the site.** `cmd/genspec` is a standalone CLI equivalent to go-swagger's
`swagger generate spec`, released independently. `cmd/genspec-wasi` is the same scan with no dependency beyond
the library, runnable under a WASI runtime, and speaks a machine-readable `-format=json` envelope. Both READMEs
are good and current (`cmd/genspec/README.md`, `cmd/genspec-wasi/README.md`) — they are the source material,
not something to re-derive.

**The three surfaces.** A knob can be set as a Go field (`Options.NameFromTags`), as a flag
(`-name-from-tags`), or as a config-file key (`emit.name-from-tags`). The mapping is total and mechanical:

- a flag is the kebab-case of the field it writes, *without exception* (`cliopts/doc.go`);
- a config key **is** the flag, spelled exactly as on the command line, under a section;
- sections are the four questions a scan answers, in order (`cliopts/sections.go`):
  `scan` (which code), `go` (built how), `load` (read how), `emit` (rendered how) — plus each command's own:
  `document` / `diagnostics` for `genspec`, `profile` for `genspec-tui`;
- precedence: **anything typed wins**, including a flag typed with the value it already had
  (`-scan-models=false` beats a file saying true). Everything else lands through `flag.FlagSet.Set`, the same
  path the command line takes.

Not every option is a flag: `FS`, `ExportData`, `InputSpec` and the callbacks are the command's business
(the command opens the path, loads the document, wires the sink). The reference has to say which is which.

**Site decisions taken with Fred (2026-08-16).**

1. The options reference moves out of Maintainers into a **revived top-level `usage/` section** — the #37
   scaffold that has sat stale at weight 2 since the site was built.
2. The new CLI page is **one page**: `genspec` front and centre, closing with a `genspec-wasi` section for
   no-toolchain / sandboxed runs, linking the Playground for the browser build.
3. The stale `usage/` tree: `usage/reference/` is deleted (its material migrated to Maintainers long ago),
   `usage/examples/basic-scan.md` is **kept** — moved into Tutorials rather than lost.

## Trajectory

1. ✅ Open the base camp and settle the target structure
2. ✅ Getting started → **Usage as a headless CLI** (the new page)
3. ✅ Revive **Usage** as the options home
   * ✅ Move `maintainers/options.md` → `usage/`, rewritten around the three surfaces
   * ✅ A page for the configuration file: discovery, sections, precedence
4. ✅ Clean up the #37 scaffold
   * ✅ Delete `usage/reference/`
   * ✅ Move `usage/examples/basic-scan.md` into Tutorials
5. ✅ Sweep the site for what the move breaks or now understates
6. ✅ Verify and land — built, linted, reviewed by Fred, amended; branch complete and unpushed

## Actions

### 1. Base camp and structure

1. ✅ Worktree `.worktrees/doc/cli`, branch `doc-cli` off master `52d7de00`.
2. ✅ Copy the gitignored Relearn theme into the worktree (build env: [[project_doc_site_build_env]] —
   `/home/fred/bin/hugo`, **not** `/bin/hugo`).

Target tree:

```
Usage                        (w2, revived)
├── _index.md                how codescan is driven: the three surfaces, at a glance
├── setting-options.md       precedence + the .codescan.yaml contract
└── options-reference.md     moved from maintainers/, three-spelling tables
Getting started              (w10)
├── Usage as a library       (w1)
├── Usage as a terminal UI   (w2)
└── Usage as a headless CLI  (w3, new)
Tutorials                    (w20)
└── Scan a package           (w5, moved from usage/examples/)
```

### 2. Usage as a headless CLI

3. ✅ `getting-started/usage-as-a-headless-cli.md`, weight 3. Sourced from `cmd/genspec/README.md`:
   install and first run; **output** (`-output`, `-format` incl. `auto` reading the extension, `-compact`,
   `-input` overlay, and the key-order cost of deriving YAML from JSON); **diagnostics** (`-quiet`,
   `-verbose`, `-color`, `-validate`, `-fail-on`, and why `-fail-on` covers validation too);
   **exit status** (0/1/2/3/4, and that a document is written whenever one could be produced);
   a pointer to the configuration file rather than a second copy of it.
4. ✅ Closing section **"Without a Go toolchain"** on `genspec-wasi`: what it trades, the `-format=json`
   envelope, running under wasmtime/wazero and what has to be mounted, export data and `-stub-stdlib`.
   Links out to the Playground page for the same engine in a browser.
5. ✅ Update `getting-started/_index.md`: the install block currently offers library + TUI only, and the
   closing line says "**Both** drive the same scanner" — now three.

### 3. Usage: how an option is set

6. ✅ `usage/_index.md` rewritten: what the section is (driving codescan, whichever way), the
   one-knob-three-spellings table as the section's opening idea, children cards.
7. ✅ `usage/setting-options.md`: the derivation rule (kebab-case, no exception), the section vocabulary
   (`scan`/`go`/`load`/`emit` + per-command `document`/`diagnostics`/`profile`), file discovery (searching
   upwards, `.codescan.yaml`/`.yml`/`.json`), `-config`/`-c` and `--no-config` (and that asking for both is
   an error), one file serving the whole family (unknown section skipped, unknown key inside a known section
   refused), and the precedence rule with the `-scan-models=false` witness.
8. ✅ `usage/options-reference.md`: `git mv` of `maintainers/options.md`, keeping its six groups, with each
   row gaining **Flag** and **Config key** columns. Options that are *not* flags (`FS`, `ExportData`,
   `InputSpec`, `OnDiagnostic`, `OnProvenance`) say so in the flag column with the reason — that is the
   part a reader cannot derive.
9. ✅ Reweight the moved page and check the Maintainers landing: `maintainers/_index.md` lists Options as one
   of its five reference documents, in prose *and* through a `children` shortcode.

### 4. Scaffold cleanup

10. ✅ Delete `usage/reference/_index.md` (links out to GitHub for annotations/keywords/grammar/sub-languages,
    all first-class Maintainers pages since June).
11. ✅ `git mv usage/examples/basic-scan.md tutorials/scan-a-package.md`, weight 5 (before model-definitions
    at 10) — it is the smallest end-to-end story and reads as the first tutorial. Fix its two relative links
    (`../../getting-started/usage-as-a-library/`) to relrefs; delete `usage/examples/_index.md`.

### 5. Sweep

12. ✅ Inbound references to the moved options page: `maintainers/_index.md`, `getting-started/usage-as-a-library.md`
    ("Options worth knowing" table + "See the godoc"), and any `relref "options"` elsewhere. Basename stays
    unique site-wide, so filename-only relrefs survive the move — the ones to fix are the prose claims, not
    the links.
13. ✅ `getting-started/usage-as-a-library.md`: its "Options worth knowing" table is a stale subset (no
    `PruneUnusedModels`, no `CleanGoDoc`, no loader knobs). Cut it down to the handful a first call needs and
    point at the reference, rather than maintaining a second list.
14. ✅ `usage-as-a-tui.md` and the TUI README already describe `profile.*`; make sure the section vocabulary
    matches what `setting-options.md` says, in both directions.

### 6. Verify and land

15. ✅ Hugo build clean (0 warn/err), no unresolved `{{<`/`{{%` in the new pages, page count accounted for.
16. ✅ Markdown lint + link check over the changed pages.
17. ✅ One commit per idea: the CLI page; the Usage revival + options move; the scaffold cleanup.

## Appendix: open questions

**A. The `usage/` section's weight.** ✅ SETTLED by Fred's amendment — he reweighted the whole top-level nav:
About 5, Getting started 10, Playground 15, Advanced Usage 20, Project 30, Tutorials 40, Shaping 50,
Annotation index 60, Maintainers 100. Reads in the order a visitor meets them. Within Getting started:
library 1, headless CLI 3, terminal UI 10.

**B. The reference's width.** Six tables gaining two columns each will be wide. If they read badly in the
theme, the fallback is a per-family split under `usage/options/` rather than dropping a column: the flag and
the config key are the point of the move.

## Achievements

Three commits on `doc-cli`, each building clean on its own (108 Hugo pages, 0 warn/err; markdown lint,
link check and spellcheck clean).

### 1. `686dfa39` — the options reference moves beside the flags ⭐⭐

- `maintainers/options.md` → `usage/options-reference.md`. All six grouped tables gained **Flag** and
  **Section** columns, filled from `cliopts`' own tables rather than by hand, so all 40 rows are the
  code's answer and not a transcription. The five options that are nobody's library flag say what
  reaches them instead (`-input` is `genspec`'s, `-export-data` is `genspec-wasi`'s, `FS` and the
  callbacks are Go values).
- New `usage/setting-options.md`: the kebab-case derivation rule and its two non-derivable shapes
  (positional packages, `-loader`'s third answer), the seven sections, upward file discovery,
  `-config`/`-c`/`--no-config`, the "anything typed wins" rule with its `-scan-models=false` witness,
  and the one deliberate library-vs-command divergence (`-scan-models` defaults `true` on the commands).
- `usage/_index.md` rewritten around one knob / three spellings.
- Scaffold retired: `usage/reference/` deleted, `usage/examples/basic-scan.md` → `tutorials/scan-a-package.md`
  at weight 5, where it reads as the first tutorial.
- Swept the four inbound `relref "options"` (the basename changed) and trimmed the stale
  "Options worth knowing" subset on the library page down to what a first call needs.

### 2. `bd835ca6` — Usage as a headless CLI ⭐⭐

`getting-started/usage-as-a-headless-cli.md`, weight 3. The two streams and why a pipeline is safe;
output selection with the `-format auto` rule; the diagnostics flags and why `-fail-on` covers
validation too; the five exit statuses; a pointer to the configuration file rather than a second copy.
Closes on `genspec-wasi` — the JSON envelope, the two things a WASI guest cannot work out for itself,
and the host-exposure table — linking the README for the depth rather than restating it.

### 3. `25827abd` — the terminal UI page catches up ⭐

Not planned, found while writing: the TUI page predates the CLI alignment and had drifted.
`-toolchain-free-loader` no longer exists (it is `-loader`, three answers); "everything else is a live
toggle **rather than** a flag" is now false in both directions; `-packages` is the deprecated spelling
of positional patterns; and the page told the reader the TUI has no configuration file, which it now
has. Also documents the run-cost card on `m` and the `-profile` flags, which landed in TUI round 3 with
no page to land on — and which `setting-options.md` now names a `profile` section for.

### 4. Fred's amendment `3d623328`, and the two passes over it

Fred restyled the whole site on top of the three commits: line reflow throughout, the top-level reweighting
(Appendix A), a rewritten home page (hand-written `{{< cards >}}` → `{{< children >}}`, plus a go-swagger
comparison), a tab set on `setting-options`, and six TODO markers.

**`20ba660c` — review fixes.** Three real defects, one of them structural: the tab set had swallowed
*Which spelling wins*, *One file several commands* and *What has no flag*, which are about all three
spellings rather than about configuration keys — and the cross-link to `#what-has-no-flag` from the first
tab pointed into a `display:none` panel (verified in the built HTML: link at panel 1, target at panel 3).
The tab set now closes after the configuration-file section. Also `-exportdata` → `-export-data`, and
`genspec-wasi` restored to the list of what a flag is available on (it registers `cliopts` too; only the
*configuration-file* half legitimately excludes it, and that now says why).

**`18403c65` — five of the six TODOs.** The `-validate` section written with captured output (running the
built binary also confirmed exit 4, which `go run` had been masking as 1); a `version` shortcode for the
`params.codescan.*` values CI has been generating for nobody; the two "maintainers details" blocks promoted
into a new `maintainers/commands.md`; the Project landing reformulated.

**`7fa16416` — the run-cost card, plain and profiled.** Fred's two captures (`stats.png`, `profile.png`)
wired in, with the prose the pictures now let the page carry: allocated-vs-retained as two different
problems, CPU charged to our call rather than the leaf frame, and the profiled scalars carrying overhead
the tables below exclude. **No TODO markers left on the site.**

**`338280fd` — room to work on the playground** ⭐⭐ (Fred: "it works fine for me"). Expand / Full screen
controls on the playground box. Notable for the method rather than the CSS: the geometry was verified in
headless Chromium against the real theme, which corrected two assumptions — a `100vw` breakout would have
scrolled *inside* `#R-body-inner` (its `overflow-y:auto` makes `overflow-x` compute to `auto`), and the
theme's width cap does not bind at 1920 at all, so expanding gains height only below ~2100px of content
width. Measured: 1600/1920 → no width gain, 2560 → +560px; height 608 → 702 at a 900px viewport;
collapse restores exactly. That is why the control is called "Expand" and not "Wide".

## Follow-ups

1. 📝 **`maintainers/commands.md` is the seam for further command-internals material** — Fred: "further
   additions will come in the same vein". Anything about how the tools are built rather than what they do
   lands there, not on the visitor-facing pages.
2. 🔍 **Table width.** Six columns render fine in the theme's scroll container, but the Effect column
   does most of the work. If it reads badly on a narrow screen, Appendix B is the fallback.
3. 📝 **The doc lane's changed-paths filter** does not watch `cmd/internal/` or `cmd/genspec-wasi`, so
   the published playground can lag its sources. Noticed during the update-doc fix, still open.
