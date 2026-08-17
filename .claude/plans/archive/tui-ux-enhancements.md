> [!NOTE]
> Last revision: 2026-08-07 (rev 3 — ALL FEATURES LANDED; 11 commits on `tui-ux-enhancements`, awaiting review)

# genspec-tui — UX chrome enhancements

## Summary

Land the remaining UX chrome on `cmd/genspec-tui`, closing out most of the roadmap's Stream 5 `UX-polish` item and
adding four new ones Fred raised. **Five items in three streams**: the **diagnostics pane** grows severity colouring,
spec-side tracking and a validation tab; **layout & chrome** gains guarded reload and adjustable splits; the
**grammar** gains annotation doc strings surfaced as in-viewer help.

Light/dark theme is **deferred** — nice-to-have, and it makes noise in the middle of the rest.

Base camp: `.worktrees/feat/tui-ux-enhancements`, branch `tui-ux-enhancements`, forked from `c195e87` (v0.36.3).

## Context

Two of the five are already registered as open in `roadmap.md` §5 (`UX-polish`) and detailed in
`wasm-playground.md` §"Phase 1 backlog — TUI chrome": guarded reload and adjustable splits. (The third registered item,
light/dark theme, is deferred — see D5.) The other three are new asks.

Two of the new asks turn out to be partly built already, which changes their shape:

- **Severity colouring** exists for the severity *label only* (`diagnostics_render.go:105`). The `Render` godoc already
  claims rows are "colored by severity", which overstates what the code does — so this is as much a correctness fix on
  the comment as a feature.
- **Diagnostic tracking** drives the source pane (`followDiag` → `driveDiagToSource`). The spec half is missing, but
  every part it needs exists: `resolveSourceToSpec` is exactly what `linkSourceToSpec` already calls.

### Settled decisions

| # | Question | Ruling |
|---|----------|--------|
| D1 | Where do validation findings land? | An extra **"validation" tab** in the diagnostics pane — not merged with scan findings, not a modal. The tab **appears only while validating** |
| D2 | Direction of truth for doc strings? | **Hand-written in Go.** No generator either way; the doc site keeps its own prose |
| D3 | Tooltip scope? | **Annotations only** |
| D4 | Do the two tabs share tracking? | **No — independent mechanisms.** The scan tab tracks source *and* spec; the validation tab can only ever track the spec |
| D5 | Light/dark theme? | **Deferred.** Genuinely nice-to-have, and it would add noise across every styled surface while the rest is in flight |
| D6 | Self-document keywords? | **No** — there are too many. Annotations get documented, and each annotation's doc carries a brief on the keywords available under it |

**On D4** — an earlier draft of this plan claimed A2 was a hard prerequisite for A3, on the reasoning that validation
findings carry a JSON pointer and so need the spec-tracking A2 builds. That was wrong. The two tabs are separate
mechanisms: the validation tab resolves a pointer straight through `specIndex.LineForPointer`, with no source
resolution and no `srcIndex` involved. **A2 and A3 are independent and can land in either order.**

## Trajectory

1. Stream A — the diagnostics pane
   > Three items over `diagnostics_render.go` and the follow-mode state machine. A1 first (it fixes a live bug the
   > others would inherit); A2 and A3 are independent of each other.

   1. ✅ A1 — colour diagnostic rows by severity [🎨]
   2. ✅ A2 — the scan tab tracks the spec pane as well as the source
   3. ✅ A3 — `v` validates the spec; findings land in their own tab, tracking the spec only

2. Stream B — layout & chrome
   > Independent of Stream A and of each other.

   1. ✅ B1 — guarded reload of the open file
   2. ✅ B2 — adjustable split sizes [🎨]

3. Stream C — grammar help
   > Independent.

   1. ✅ C1 — `Doc()` on `AnnotationKind` + an in-viewer help key

4. ✅ Close-out
   1. ✅ README + roadmap status updates [📚]
   2. ✅ Quality gate: `go test work ./...` + `-race` on both modules, lint against master, coverage on the patch

## Actions

### Stream A — the diagnostics pane

1. ✅ **A1 — colour rows by severity** [🎨] — done, see Achievements
   - Today `formatDiagnostic` colours only `d.Severity.String()`; loc, message and code render plain.
   - ❌ **Hazard to fix while here.** `Render` wraps the already-styled row: `theme.Selected().Render(row)` where `row`
     contains ANSI escapes. The inner resets terminate the outer background early, so a selected row loses its
     highlight from its first coloured span onward. This is precisely the "truncating an already-coloured string cuts
     through escapes" class that the spec pane's span mechanism was built to dodge.
   - Fix follows the doctrine already stated in the README — *precedence is cursor, then search match, then syntax*:
     when the row is selected, render the **raw** text under `Selected()`/`Follower()`; otherwise render with severity
     colours. No nesting, no spans needed here.
   - Correct the `Render` godoc, which currently claims more than the code does.
   - Do this first: A2 and A3 both add rows to this renderer and would otherwise inherit the bug.

2. ✅ **A2 — the scan tab tracks spec as well as source** — done, see Achievements
   - `followDiag` currently mirrors one follower. Add the spec: `d.Pos` → `resolveSourceToSpec(srcIndex, specIndex,
     file, line)`, the same call `linkSourceToSpec` makes.
   - Model change: `followTarget` is a single string and the badge reads `DIAG ▸ SOURCE`. A diag driver now feeds
     **two** followers, so the field and the badge both need to carry a pair (`DIAG ▸ SOURCE + SPEC`).
   - Honesty case to preserve: a diagnostic whose source line produced no spec node (a parse error on a malformed
     annotation is the common one) must leave the spec follower where it is and say so, matching how
     `driveSpecToSource` handles its misses.
   - `Enter` (the one-shot `jumpDiagToSource`) stays source-only — it exists to *go and work on* the finding.

3. ✅ **A3 — `v` validates the spec, into its own tab** — done, see Achievements
   - Add `github.com/go-openapi/validate` to the **TUI module only**. Same argument that keeps bubbletea out of the
     lean library covers this; the library gains no dependency.
   - Tab strip in the diagnostics pane header, shown **only once a validation has run** — no empty tab sitting there
     advertising a mode the user has not asked for. `v` runs the validator against the last good `*spec.Swagger`,
     async like the scan, then reveals the tab and switches to it.
   - **Its own tracking mechanism** (D4), not a reuse of A2's: findings carry a JSON pointer, so the tab's `f`/`Enter`
     resolve through `specIndex.LineForPointer` and drive the spec pane. No `srcIndex`, no source resolution, and a
     distinct follow mode rather than a widened `followDiag`.
   - ✅ **Dependency approved** (2026-08-07): `go-openapi/validate` goes into the **TUI module only**, tail included.
   - 🔍 To settle in build: exact validate entry point (`validate.Spec` vs `NewSpecValidator`) and how its error/warning
     split maps onto the three-severity display.
   - 🔍 Second-order: what the tab does when a rescan lands, since its findings are now stale against a changed spec.
     The "appears only while validating" framing points at the cheapest honest answer — retire the tab on rescan and
     let the user press `v` again — but confirm that is what Fred means before building it.

### Stream B — layout & chrome

1. ✅ **B1 — guarded reload** — done, see Achievements
   - Bind `F5` (and/or `ctrl+r`; `r` is taken by rescan, `ctrl+r` is free) to re-read `currentFile` from disk.
   - The guard has its signal already: `fileView.Dirty()`, which is what `m.stale()` feeds the STALE badge from.
   - Dirty buffer ⇒ a confirm step. A small modal fits the existing `overlays()` precedence machinery rather than
     inventing a second interaction mode.
   - This is the deliberate replacement for the auto-reload that was removed for clobbering edits — the point is that
     the discard is *asked for*.

2. ✅ **B2 — adjustable splits** [🎨] — done, see Achievements
   - `recalcLayout` (`model.go:258`) hardcodes exactly two fractions: `m.height/4` for the diagnostics strip,
     `m.width/3` for the left pane. Making those two numbers Model state is the whole change.
   - Clamp against the minimums already encoded there (`max(…, 5)`, `max(…, 3)`, `max(…, 1)`).
   - Keys: `ctrl+←/→` for the vertical split, `ctrl+↑/↓` for the horizontal. Mouse-drag on the border is feasible —
     regions are already stored for hit-testing — but it is strictly more work; keyboard first.
   - Persistence is per-session only: it is Model state, so it survives rescans but not a restart. A config file is out
     of scope; note it in the README rather than half-building it.

### Stream C — grammar help

1. ✅ **C1 — `Doc()` on `AnnotationKind` + an in-viewer help key** — done, see Achievements
   - Implementation is a **`Doc()` method on `AnnotationKind`** — a switch, matching how `annotations.go` is already
     written (an enum driven by switches, with no data table). No new lookup table, no new option plumbing. This is
     simpler than the table an earlier draft proposed and fits the file's existing shape.
   - Content per annotation: the one-liner, **plus a brief on the keywords available under that annotation** (D6).
     Seed the one-liners by hand from the doc-site frontmatter, which already has them written and reviewed — e.g.
     `swagger-name.md` carries `description: "Overrides the emitted property name of a struct field or interface
     method."` Per D2 this is a **copy, not a coupling**: no generator, no build-time dependency on `docs/`.
   - ⚠️ **Consequence for the surface.** Because the doc now carries a keyword brief as well as the one-liner, it runs
     to several lines — so the status line an earlier draft proposed is too small. Use a **compact popup** near the
     line instead, reusing the existing `overlays()` machinery.
   - Trigger: a TUI has no hover, so help needs a key. The source viewer already knows which runs are annotations (the
     spec-key syntax class), so with the nav line on such a line, `K` — after the LSP hover convention — opens the doc.

### Close-out — ✅ done 2026-08-07

1. ✅ **README** — `cmd/genspec-tui/README.md` key tables gain `v`, `F5`, `K`, the split keys; the Limitations section
   gains split persistence and whatever A3's rescan ruling turns out to be. [📚]
2. ✅ **Roadmap** — `UX-polish` kept 🔶 with light/dark theme as its only remaining item; two new ✅ rows (UX-validate,
   UX-annref); `wasm-playground.md` chrome backlog reconciled, with the deferred theme's findings parked there. [📚]
3. ✅ **Quality gate** — `go test work ./...` green, `-race` clean on both modules, lint clean on both, coverage
   86.6% `internal/ux` · 93.6% `ux/validation` · 100% `ux/confirm` and `ux/reference`. [😇]
4. ✅ One commit per item, DCO signed, Fred authors and signs. 11 commits. [😇]

**Awaiting review** — nothing merged, per [[feedback_wait_for_review_before_merge]].

## Achievements

### Stream A — the diagnostics pane

> A1 (`c4a031a`) and A2 (`8b890fe`) both manually tested and accepted by Fred, 2026-08-07.

1. ✅ **A1 — colour rows by severity** (2026-08-07) ⭐⭐
   - Severity now reaches the **message**, not just the label: label bold in the severity hue, message the same hue at
     normal weight, `[code]` dimmed via a new `theme.Dim()`. The pane can be scanned for red.
   - ✅ **Fixed the live nesting bug as a side effect.** Confirmed empirically before claiming it — a standalone lipgloss
     probe showed the inner reset closing the outer background, then an A/B with production reverted showed the witness
     failing at 2 resets on a selected row. A selected row's highlight really did stop at the severity label.
   - The fix is the doctrine the README already stated (*cursor, then search match, then syntax*): the highlight goes
     over **raw** text via a new `plainDiagnostic`, so nothing is ever styled twice.
   - 🏁 Three tests, two of them true regression witnesses (both verified failing pre-change): reset-count on the
     selected row, severity reaching the message, and a text-identity guard against the styled/plain pair drifting.
   - 🏁 Added the package's missing `TestMain` colour profile — its absence is *why* this hid: with lipgloss degraded to
     plain text under `go test`, the nesting left no trace to assert on.
   - Coverage 89.6% on the package; lint clean; `go test work ./...` green.

2. ✅ **A2 — the scan tab tracks spec as well as source** (2026-08-07) ⭐⭐
   - `followDiag` now drives BOTH followers: source to the position the diagnostic carries, spec to the node that
     position produced — through the same `resolveSourceToSpec` nearest-anchor walk the source pane's own `f` uses, so
     "what did this line turn into" has one answer regardless of which pane asked.
   - The two halves resolve **independently** and the badge reports both (`DIAG ▸ SOURCE + SPEC`). A finding on a line
     that produced no spec node is the ordinary case, so a spec-half miss must not hide that the source half resolved.
   - The widened contract landed in the type's own godoc: most modes drive one follower because the driver is itself one
     end of the link; the diagnostics pane is a third place, so it drives both.
   - 🏁 Two new tests on the existing `joinFixture` (real files, real provenance, real spec index) plus a tightened badge
     assertion — all three verified failing pre-change.
   - ❌ **Found a weak assertion.** `TestFollowBadge` used `Contains`, so the old `"DIAG ▸ SOURCE"` still matched the new
     `"DIAG ▸ SOURCE + SPEC"` and the label change slipped through green. Tightened to the full label.
   - 📚 README corrected: it claimed follow works "diagnostic → source", which this makes false.
   - Coverage 86.8% on `internal/ux`; lint clean; `go test work ./...` green.

### Stream B — layout & chrome

1. ✅ **B1 — guarded reload** (2026-08-07) ⭐⭐
   - `F5` re-reads the open file from disk. Asks only when there is something to lose — a clean buffer just reloads,
     because a prompt with no stakes trains people to dismiss prompts.
   - New `ux/confirm` overlay: generic yes/no, first in the `overlays()` precedence (a question buried under another
     modal is a question you cannot answer). Follows the established contract — the overlay records the answer, the
     model decides what it means, mirroring `applyOptions`. `TakeAnswer` **consumes**, so one yes cannot re-fire.
   - Safe default: `y` accepts; `n`/`Esc`/**`Enter`** decline. The destructive answer has to be typed on purpose.
   - **F5 only, not also `ctrl+r`** — the editor is a live textarea and `ctrl+r` is redo or reverse-search in enough
     editors that binding it to a discarding action would be a trap.
   - Two behaviours found by testing rather than assumed, both A/B-verified:
     - F5 **is** swallowed by the textarea without an explicit editor binding — so it is bound in `handleEditKey` too,
       which is the state the guard exists for in the first place.
     - `SetFile` resets the nav line to the top (right when opening, wrong when re-reading), so reload restores the
       line. By NUMBER — the honest approximation, since nothing anchors Go source to anything stable across an
       external edit.
   - Reload always lands in the read-only viewer: the buffer now holds something the user did not type.
   - 🏁 7 model-level tests + 6 on the overlay (100% on `ux/confirm`).
   - Coverage 86.9% on `internal/ux`; lint clean; `go test work ./...` green.

2. ✅ **B2 — adjustable splits** (2026-08-07) ⭐⭐
   - `ctrl+←/→` and `ctrl+↑/↓` move the two dividers, each in its own arrow's direction — the only mapping that stays
     right whichever pane you happen to be thinking about. 5% a step.
   - Held as **proportions, not cell counts**, so a terminal resize keeps the chosen layout instead of handing the
     difference to whichever pane was measured in cells. Defaults (33% / 25%) reproduce the historic geometry, pinned
     by a test.
   - Travel is bounded (`15–85%` / `10–60%`) over the pre-existing absolute floors. The reason is recoverability, not
     neatness: a pane driven to nothing cannot be dragged back, because the keys that would restore it are advertised
     in a status line it no longer has room to show.
   - Reachable from the editor too (shared `handleSplitKey`), A/B-verified: the textarea otherwise swallows the key.
   - ❌ **Two test weaknesses surfaced, both fixed.**
     - `help`'s `TestPaging_EveryNavigablePaneSupportsIt` located sections by bare substring, so a new *entry* ending in
       the word "diagnostics" made it inspect the wrong block. Anchored to a line of its own — and re-verified by
       mutation that it still fails when a section genuinely loses its paging keys.
     - B1's "confirmation popup" section became the LAST one, and `TestHelp_ShortTerminalStillReachesTheEnd` uses the
       last entry's action as its needle — which was the bare word "no", passing trivially. Reworded to something
       distinctive.
   - 🏁 6 tests: direction of travel per key, panes still tiling, both bounds, proportion surviving a resize, defaults
     matching the historic layout, and reachability from the editor.
   - Coverage 87.3% on `internal/ux`; lint clean; `go test work ./...` green.

3. ✅ **B3 (unplanned) — stop the scrolling modals resizing themselves** (2026-08-07) ⭐⭐
   - Reported by Fred against the help overlay after B2. `theme.Modal()` sizes to its content, so a *scrolling* overlay
     was framed to whichever lines happened to be visible: 8 distinct widths between 58 and 95 columns while scrolling
     the keymap, measured by the A/B.
   - **The options overlay had it too** — same root cause, same fix. It windows its rows around the cursor, so its frame
     changed on cursor moves (84 ↔ 95).
   - Both now measure over EVERYTHING they can show, before the window is taken, and pin the frame to it.
   - ❌ **lipgloss counts padding INSIDE `Width`.** Probed rather than assumed: `Width(10)` on a 10-wide body with
     `Padding(1,3)` leaves 4 columns and wraps the rest. Pinning to the measured text width therefore wrapped exactly
     the longest lines — the ones the measurement came from. New `theme.ModalAt(textW)` adds the padding back, keeping
     that knowledge beside the `Padding` that causes it.
   - ⛔ **Deliberately NOT capped to the terminal.** A cap makes lipgloss wrap the overflow, splitting a toggle's label
     from its explanation; an over-wide modal is simply clipped, which is what it already did. Narrow-terminal behaviour
     is unchanged — widening that is a separate job needing truncation, not wrapping.
   - 🏁 3 tests, all A/B-verified: width stable while scrolling (both overlays) + the widest line surviving the frame
     intact, which is the guard against the padding trap above.

4. ✅ **B4 (unplanned) — stop the header running off the right of the screen** (2026-08-07) ⭐⭐
   - Reported by Fred with a screenshot: on a deep worktree path the stats and the ready/scanning indicator were cut off.
   - Cause: the work dir got `max(m.width-54, 12)` — a hand-tuned constant standing in for every other field. All of
     them grow (stats with the spec, tail with the scan duration, match counter only while searching), so past the guess
     the line overflowed. A/B measured **~40 columns of overflow at every width tested**.
   - Now measured, not guessed: the inelastic fields are built first and the work dir gets the remainder. It keeps no
     floor of its own — anything reporting scan STATE is worth more columns than the directory it ran in.
   - Same `MaxWidth` clip applied to the status line, whose variants embed unbounded JSON pointers / follow targets /
     paths (`statusLine` → `statusContent` + clip). Verified 87 columns of content clipping to 40.
   - 🏁 2 tests over a width matrix with every field at its longest, both A/B-verified.
   - ✅ **Checked and NOT a bug:** the diagnostics pane in the same screenshot appears to overrun its border. Probed
     directly — it clips at the pane edge with the border intact. Nothing to fix.

### Stream C — grammar help

1. ✅ **C1 — annotation reference popup** (2026-08-07) ⭐⭐
   - `AnnotationKind.Doc()` in the grammar package: `{Usage, Summary, Keywords}` for all 20 annotations, a switch
     matching how `annotations.go` is already written. Summaries carried by hand from the doc-site frontmatter — a copy
     on purpose, since coupling the parser to the docs tree to save twenty short strings is a bad trade.
   - `Keywords` is prose, not a generated list: the keyword table knows which CONTEXTS a keyword is legal in, not which
     annotations open those contexts, so generating it would mean inventing that mapping and getting it subtly wrong.
   - `K` in the source viewer opens `ux/reference`. Checked on the RAW spelling and before the nav keys — `MsgBinding`
     lowercases, so `K` would otherwise read as `k` and scroll. Same treatment `N` already gets for search.
   - Reads the BUFFER, so it works on an annotation still being typed — which is when it is wanted.
   - The lookup is scoped to what follows a `//`, so annotation text inside a string literal (of which this repo's own
     fixtures are full) is not offered as a live annotation.
   - This overlay WRAPS its prose at a fixed 72 columns — the opposite case from the scrolling ones, and why
     `theme.ModalChromeW` came back. Both directions pinned by tests.
   - ❌ **An A/B caught a bug in my own test.** The string-literal guard pointed at line 10 (a plain comment) instead of
     line 11, so it passed for the wrong reason — removing the `//` scoping did not fail it. Fixed, then re-A/B'd until
     it bit.
   - 🏁 5 grammar tests (completeness over the enum range, usage-names-its-own-annotation, summary shape) + 8 model
     tests + 7 on the overlay (100% on `ux/reference`).
   - ⛔ Keyword-level docs stay out (D6): each annotation's entry says which family its body accepts instead.
   - Coverage 87.1% `internal/ux`, 100% `ux/reference`; lint clean; `go test work ./...` green.

2. ✅ **C2 (unplanned) — click a directive to open its reference** (2026-08-07) ⭐⭐
   - Fred asked what mouse HOVER would cost. Assessed: the hover is trivial, the tooltip is not — `WithMouseAllMotion`
     kills native terminal text selection for everyone, overlays are centred rather than positioned (lipgloss v1 has no
     compositing, so a tooltip means hand-rolled ANSI splicing), and overlays are modal where a tooltip must be passive.
     Click gets ~80% of the value from the mouse events already enabled. Fred took click.
   - Requires the pointer on the directive's own TOKEN, not its line — focusing a pane is the commonest click there is,
     and it must not throw a popup up. The click leaves the nav line alone: it asks "what is that", not "go there".
   - New `FileView.LineColAt`, sharing the prefix arithmetic with the renderer rather than restating it.
   - ❌ **The geometry pin caught a real bug.** The first version of the round-trip test inverted `LineColAt` with
     `LineColAt` — circular, and it passed. Rewritten to check the mapping against the RENDERED frame, it failed
     immediately: a line drawn truncated ends in `…`, and the mapping happily resolved that cell to a buffer character
     the user cannot see. Also added the missing right-border bound. Both fixed.
   - 🏁 8 tests: token-vs-line targeting, focus-without-popup, nav line untouched, rune columns under non-ASCII prose,
     adjacency of the name to the prefix, and the frame round-trip.

3. ✅ **A3 — the validation tab** (2026-08-07) ⭐⭐
   - `v` runs the rendered JSON through `go-openapi/validate`; `V` switches tabs. The tab exists only once `v` has been
     pressed (D1), and a **rescan retires it** — findings judged the document that scan replaced, and every row of such
     a list invites navigating to a node that may have moved. A/B-verified that the retirement bites.
   - The two tabs track different ends, which is why they are tabs and not one list: a scan diagnostic knows a source
     position, a validation finding knows only a JSON pointer. So `f`/`Enter` here drive the **spec** and nothing else
     (`followValidation`, `VALIDATION ▸ SPEC`). D4 confirmed correct in the build — no code is shared with A2's path.
   - Validates the rendered BYTES, not the `*spec.Swagger`, so what is judged is what is on screen.
   - ❌ **I documented the wrong limitation, and Fred caught it.** The first README note led with a dot-in-a-path-template
     ambiguity — real in theory, never observed in validator output — while missing the two that fire on a trivial spec.
     Measured properly: 4 of 7 findings resolve EXACTLY (including deep paths like
     `/definitions/User/properties/email/type`); 3 land on the parent, because
     (a) **the validator omits array indices**, so anything inside `parameters` resolves to the array — and parameter
     lists are always arrays, so this is the common case; and (b) a **`required`-but-missing** finding names a node that
     by definition is not there, so its parent is the only honest landing.
   - The headline is therefore that navigation **works**, not that it is unreliable. An imprecise landing is always an
     ANCESTOR of what was reported, never a sibling. `TestValidation_PointerResolutionAccuracy` measures and pins all
     three counts, so a regression in either direction is visible.
   - **Lesson:** the original test only asserted a finding resolved to *something*, which the walk-up fallback
     guarantees — so it certified the feature while measuring nothing. Cf. [[feedback_test_input_space_vs_contract]].
   - Findings carry their location two ways: `*errors.Validation` has it as a field; everything else has it quoted at
     the front of the message. Both read; anything else simply has no location and says so.
   - 🏁 10 model tests + 13 in `ux/validation`, including the same selected-row nesting guard A1 needed, and that
     package's own `TestMain` colour profile.
   - Coverage 86.6% `internal/ux`, 93.6% `ux/validation`; both modules lint clean; `go test work ./...` green.
   - ⛔ True hover tooltips parked: they are really "build a positioned-overlay layer", which would also pay off for
     keyword hover and LSP-style diagnostic popups. Not this branch.

---

## Deferred and won't-do

- ⛔ **B3 — light/dark theme** (D5). Deferred, not dropped: it stays an open `UX-polish` sub-item on the roadmap.
  When it comes back, two findings from this pass still hold:
  - `theme.go:14` states the current design deliberately — *"Constants rather than variables: a theme nothing can
    reassign at runtime is one less thing a rendering bug can be."* A switch has to respect that, e.g. a `Palette`
    value fixed once at startup rather than a mutable global, and no live toggle key.
  - ⚠️ A prerequisite refactor is waiting: `Match()` (16/226), `Selected()` (231), `Follower()` (252), `Chip()` (16)
    and `Stale()` (16) hardcode colour literals outside the palette, so a second palette lands half-applied until they
    move in. `termenv` is already a direct dependency, so auto-detection is free when the time comes. The real cost is
    **choosing** the light palette — the current 256-colour values are picked for dark backgrounds.
- ⛔ **C2 — keyword doc strings** (D6). There are 39 canonical keywords and self-documenting each is too much; the
  keyword brief inside each annotation's `Doc()` covers the need instead.
- ⛔ **Not in scope:** persisting split sizes across sessions (needs a config file); highlighting in *edit* mode
  (`bubbles/textarea` owns its rendering — the standing answer is the VIM/VS-Code integration).

## Appendix — issues, risks and loose ends

- ❌ **The nested-style bug in A1 is live today**, independent of the feature. A selected diagnostic row already loses
  its highlight partway across, because the severity label is styled before the row is wrapped. A1 fixes it as a
  side effect; worth calling out in the commit message so it reads as the fix it is.
- ⚠️ **A2 widens the follow-mode contract** from one driver + one follower to one driver + two. The state machine in
  `syncFollowIfActive` is currently written around a single `followTarget`; the tests in `crossref_test.go` assert
  against that shape and will need extending, not just passing.
- 🔍 **A3's dependency direction is worth a sanity check.** `validate` depends on `spec`, `loads`, `strfmt` and
  `analysis`. All are go-openapi and none reach the lean library — but it is a visible growth of the TUI module's
  tree, so confirm it is wanted before committing to it.
- 📝 **Parked, belongs upstream not here:** `parse.invalid-enum-option` reports the column of the space *before* the
  value (`// collection format: pipe` points at `" pipe"`), so the diagnostic underline starts one column early. It is
  registered in `wasm-playground.md`; the fix is in codescan proper, not the TUI.
