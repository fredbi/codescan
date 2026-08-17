> [!NOTE]
> Last revision: 2026-08-06

# Browser playground — prototype to experimental feature

## Summary

Take `hack/doc-site/genspec-wasi` from a 710-line scaffold to something we are willing to call an experimental
feature and put in front of a doc-site visitor. Two halves, and the first unblocks the second:

- **Go side** — a structured result contract. The guest speaks argv → stdout/stderr, so the front-end gets a JSON
  blob and an unstructured text stream. Diagnostics with positions and source↔spec provenance never reach it. A
  `-format=json` envelope carries all three, and both callbacks already exist in the library.
- **Front-end** — design tokens and a real app shell, a CodeMirror editor with Go highlighting, a file tree, a
  pointer-addressable spec view, gutter diagnostics, source↔spec cross-highlight, a lazy Swagger-UI tab, and an
  entry experience aimed at somebody who arrived from a tutorial rather than from a checkout.

Source-loading quirks are **out of scope** here — that work has its own branch (`feat/source-loader`).

**Standing assumptions (Fred, 2026-08-06).**

1. The standard library reaches the guest as **baked-in export data**, and the remaining fidelity gaps are the
   loader stream's problem, to be closed by lazy source-fetching in the loader and scanner.
2. **A real project arrives vendored.** `go mod vendor` puts every third-party dependency in the tree as source,
   which is the only thing that resolves a library whose meaning lives in comments — `strfmt` declaring its formats
   with `swagger:strfmt` — since export data holds types and not comments.

Together those close the fidelity question, so this plan carries no advisory, no degraded-mode banner and no
"avoid `strfmt` in the examples" constraint. If a scan comes back thin, that is a bug filed against the loader.

What the second assumption *does* buy is two obligations on the front-end, both now actions below: carry
`vendor/modules.txt` (the switch that makes a vendored tree authoritative), and make vendoring the guided path
rather than folklore a visitor is expected to already know.

## Context

The app today is a toolbar, a `<select>` plus a `<textarea>`, a `<pre>`, and six checkboxes. It works end to end and
is deliberately bare. The reference for what "finished" looks like is `cmd/genspec-tui`, which has had far more
polish and whose headline feature — ask a spec node which Go code produced it, and vice versa, *by position rather
than by name* — is the thing the playground cannot currently do at all.

Prior art and constraints:

- `wasi-build.md` — the live plan for the artifact. §"The result contract" is the open decision this plan settles;
  item 5 of its **Open** list is "🎨 Polish the front-end", which is this document.
- `tabbed-examples.md` §"Tomorrow — the Playground (W11)" — the doc site's contract with us: the playground
  **replaces** the whole source/spec box rather than becoming a third tab, and **injects Swagger-UI itself**. That
  is why Swagger UI is our concern and not the shortcode's.
- `doc-site-wishlist.md` W11 (live playground) and W10 (annotation → spec cross-highlight, explicitly "unlocked by
  W11" and "rides on it").
- `genspec-tui-linkage.md` — the cross-ref design whose source half is `scanner.Provenance`. The spec-side pointer
  and the source-side pointer are contracted to be byte-identical, so the join is free.

Settled with Fred before writing this (2026-08-06): JSON envelope now · design tokens + headless bits, no visual
framework · Swagger UI in scope, lazy-loaded · doc-site visitor is the default audience.

## Trajectory

1. ✅ **Result contract** — the guest tells the host everything it knows
   > One decision, one Go commit, and every interactive feature below depends on it.

   1. ✅ `-format=json` on `genspec-wasi`: `{spec, diagnostics[], provenance[]}` on stdout
   2. ✅ Positions relative to the module root, not the guest's `/src`
   3. ✅ 🏁 Round-trip test: envelope parses, positions land on real lines

2. ✅ **Foundations** — the app stops being a scaffold [🎨]
   > Nothing here is visible as a feature; everything visible depends on it.

   1. ✅ Design tokens (`--cs-*`), light/dark, and an embed mode that defers to host variables
   2. ✅ App shell: resizable split, tabbed right pane, collapsible diagnostics drawer, responsive collapse
   3. ✅ Store from module singleton to a per-mount instance passed down through context
   4. ⛔ Headless primitives — **deferred, not adopted**: what Phase 2 needed was a tablist and a
      splitter, both small and completely specified. A library earns its bytes at focus trapping and
      inert backgrounds, which arrive with Phase 8's dialogs. Revisit there.

3. ✅ **The source side** — an editor rather than a textarea
   1. ✅ CodeMirror 6 with `@codemirror/lang-go`
   2. ✅ File tree replacing the `<select>`, collapsing single-child directory chains — what makes
      1,490 files navigable. Add / rename / delete not done and not yet wanted
   3. ✅ Debounced auto-scan on edit
   4. ✅ Open-module path kept, and made a second-class entry point

4. ✅ **The spec side** — addressable, not a `<pre>`
   1. ✅ Pointer → line-span map, from our own renderer rather than `JSON.stringify`
   2. ✅ Search (`/`, `n`/`N`, wrapping), copy, download — folding not done, and not missed so far
   3. ✅ YAML toggle, sharing the renderer so tracking survives the switch

5. ✅ **Diagnostics** — positioned, not a text dump
   1. ✅ Gutter marks in the editor, by severity
   2. ✅ A listbox: arrows move, the current row is marked, moving follows into the source
   3. ✅ Severity filter — the counts became switches, honoured by the gutters and the tree too

6. ✅ **Cross-highlight (W10)** — the TUI's headline feature, in a browser
   1. ✅ Source line → spec node, spec node → source line, both by position
   2. ✅ Nearest-anchored-ancestor fallback for nodes that carry no provenance of their own

7. ✅ **Swagger UI** [🎨]
   1. ✅ `Spec | Swagger UI` tabs, bundle fetched on first activation, in its own chunks
   2. ✅ Degrades honestly with no `paths` — dockerctl's shape exactly — and after a failed scan
   3. ✅ Pulled towards a dark theme by a colour transform, reversible via "show as published"

8. ✅ **Entry experience** — for somebody who arrived from a tutorial
   1. ✅ Lands on an example and scans immediately, so the first frame already shows output
   2. ✅ Examples gallery — models, routes, operation, enums, polymorphism — each run through the
      artifact and its output read before being embedded
   3. ✅ Real empty, loading, error and refusal states, including three named phases on a cold visit
   4. ✅ **Guide the vendoring**, plus a `?` guide covering module → vendor → open, with the reasons
   5. ⛔ Share link — dropped, see "Won't do"

9. ✅ **Prototype → experimental** [📚][🏁]
   1. ✅ README rewrite: the playground's, `cmd/genspec-wasi`'s, and the root's browser section
   2. ✅ 🔍 Decided: verified by hand, stated in the README, automation deferred
   3. ✅ Weight budget stated: 157 KB of app, 419 KB of Swagger UI behind its tab, 8.4 MB of artifact

## Actions

### Phase 1 — Result contract (Go) ✅ landed

1. ✅ **`-format=json`** on `cmd/genspec-wasi`
   - Default stays `spec` (bare document), so the WASI tests, `hack/browser` probes and any CLI use are untouched
   - `json` emits one object on stdout; diagnostics stop going to stderr in that mode
   - Shape follows the Go types we already have — `grammar.Diagnostic{Pos, Severity, Code, Message}` and
     `scanner.Provenance{Pointer, Pos}` — so this is serialization, not design:

     ```
     {"spec": {...},
      "diagnostics": [{"severity","code","message","file","line","col"}],
      "provenance":  [{"pointer","file","line","col"}]}
     ```
   - `Options.OnProvenance` is marked experimental; the envelope inherits that and says so

2. ✅ **Paths are module-relative.** The guest scans `/src`; the front-end knows `models/pet.go`. Strip the workdir
   prefix at the emission site rather than making every consumer do it.

3. ✅ 🏁 **Test.** Native tests that run the command with `-format=json` and check a written module's
   diagnostics and provenance land on the lines they claim.

### Phase 2 — Foundations (front-end)

4. ✅ 🎨 **Design tokens.** One `tokens.css`: colour, space, radius, type scale, elevation. Light and dark via
   `prefers-color-scheme` with an explicit override. Every token falls back to a doc-site variable when one is
   present, which is what makes the embed inherit the host theme instead of fighting it.

5. ✅ 🎨 **App shell.** Header, resizable source/spec split, tabbed right pane, collapsible diagnostics drawer,
   status line. Panes stack below a breakpoint.

6. ✅ **A playground per mount.** Module-level singleton today, which is correct for one instance and wrong the moment
   a page has two. Cheap now, invasive later.

### Phase 3–6 — The interactive core

7. ✅ **CodeMirror 6** with Go highlighting, replacing the textarea. Chosen because it is also the substrate for
   the gutter marks (Phase 5) and the range decorations (Phase 6) — one dependency, three features.

8. ✅ **File tree** replacing the `<select>`, collapsing single-child chains. Add / rename / delete not
   built and not yet wanted; the refusal is kept and now measured (64 MB, see 14b).

9. ✅ **Debounced auto-scan.** The worker keeps the compiled module, so a rescan is instantiation plus scan. Must be
   cancelable: a scan in flight when the next edit lands is discarded, not awaited.

10. ✅ **Pointer → line-span map.** Walk the pretty-printed spec once, record a span per JSON pointer. Gives folding,
    node selection and both directions of cross-highlight from one structure.

11. ✅ **Gutter diagnostics + drawer**, in both margins, arrow-navigable and severity-filtered.

12. ✅ **Cross-highlight.** Source line → provenance entries on that line → spec spans; and the reverse, resolving a
    node with no provenance of its own to its nearest anchored ancestor.

### Phase 7–9 — Payoff and packaging

13. ✅ 🎨 **Swagger UI tab**, lazy — 419 KB gzipped against an 8.4 MB artifact, and its network defaults
    turned off. Never fetched unless the tab is activated.

14. ✅ **Examples gallery + real states + the vendoring prompt.**

14b. ✅ **Raise the upload ceiling.** Done — 64 MB with a warning band at 16, both measured. See the
     achievement below.

15. ✅ 📚 **README rewrite** — the root README gains a browser section; `cmd/genspec-wasi`'s documents the
    runtime figures and the sentinel errors. — playground, `cmd/genspec-wasi`, and the root README's mention.

16. ✅ 😇 **Lint pass.** Nineteen to none. Three were real: unbounded `uint64`→`int` on lengths read from a
    module, subprocesses without a context, and errors a caller could only match on by prose. One documented
    `nolint` remains, where the guards prove the range and the analyser will not carry it across a return.

17. ✅ 🔍 **Testing decision — settled 2026-08-07: manual, and said so.** The playground is offered for
    demonstration. Automation is **deferred**, not rejected: CI already carries the golden suite and this is not
    the work to add a chromium download to it. If the browser tier keeps growing it may move to a repository of
    its own, dedicated to exactly this — UX and WASI integration — which is the right place to pay for a browser
    in CI. The analysis behind the decision is kept in the appendix below rather than thrown away.

18. 📝 🏁 **When automation is revived**, in this order: the store first (cheap, no dependency, and it would have
    caught a bug that shipped), then a browser smoke test, then CI for it.

## Achievements

### Phase 0 — making the standing assumption true (2026-08-06) ⭐

The playground did not in fact bake in the export data: `npm run wasm` built without the `exportdata` tag and the
worker passed `-stub-stdlib=true`, so every browser scan ran in the degraded mode. Now true rather than assumed:

- ✅ `npm run wasm` regenerates `internal/exportdata/exportdata.zip` and builds with `-tags exportdata`
- ✅ `argvFor` takes `'embedded' | 'mounted' | 'stub'`; embedded names no flag at all, the command picking up its
  own archive. The old two-way `'stub' | 'export-data'` could not express "carried inside the binary".
- ✅ The worker asks for `embedded`

Witness: a fixture with `time.Time` / `json.RawMessage` / `time.Duration`, run under wasmtime with only the module
mounted, raises **no** `scan.synthesized-import` — where the stub build synthesizes the whole standard library.

### Phase 8 (part) — arriving somewhere (2026-08-07) ⭐⭐

- ✅ Scans on arrival: the artifact is fetched either way, so waiting for permission cost the first
  frame and bought nothing
- ✅ Five examples in the chrome, each a whole module, each **run through the artifact and its output
  read** before being embedded. Models keeps one hint deliberately, so the diagnostics pane is found
  doing something real rather than saying "none"
- ✅ A `?` guide: init → vendor → open, each with its reason, since the second is the one nobody
  guesses. Native `<dialog>` — **which settles the Phase 2 deferral**: focus trap, Escape, backdrop
  and inertness are the platform's, so the headless library is not needed at all
- ✅ The vendoring mistake is caught from `go.mod` before a scan rather than described by its
  consequences afterwards
- ✅ `humanBytes` existed twice with different rounding; consolidated into `format.ts`

**A near miss worth recording.** I edited `store.svelte.ts` by slicing between two index anchors, the
end anchor sat before the start, and `replace("", …)` inserts between every character — a 5.87 MB
file. svelte-check reported 160,140 errors, so it was caught immediately and restored from HEAD, but
index-slice edits on source are not worth the risk again.

- ✅ The first visit names its three phases (fetch / compile / scan) rather than calling a download a
  scan. Progress is counted through a transform so `compileStreaming` still compiles as it arrives,
  and the bar is determinate only when `Content-Length` describes what the stream will deliver — a
  compressed body overshoots it, and a bar that overshoots teaches distrust of the whole page

Phase 8 complete. The share link is dropped — see below.

### Phase 7 — Swagger UI (2026-08-07) ⭐⭐

- ✅ Lazy: 1.4 MB JS + 177 KB CSS in their own chunks, never fetched unless the tab is opened. Main
  bundle unchanged at 454 KB / 157 KB
- ✅ **Both of Swagger UI's network defaults disabled.** `validatorUrl` POSTs the whole document to
  validator.swagger.io — verified present in the shipped bundle, so this is load-bearing — and "Try
  it out" fires real requests at whatever host the document names. Either would contradict the
  privacy line in the status bar
- ✅ No-paths gets an explanation rather than a blank frame, which Swagger UI draws identically for
  "no operations" and "broken"
- ✅ Dark theme via `invert(0.92) hue-rotate(180deg)`, the pair chosen so method badges keep the hues
  that carry their meaning; reversible through "show as published"

Tested by Fred against several fixtures including the petstore.

**One bug worth remembering.** Pre-darkening the container before inverting it is the intuitive move
and is exactly backwards: `filter` applies to the element's own background, so a near-black surface
came back near-white and the inverted (light) text sat on it washed out. The background must stay
white and be inverted with everything else.

### The diagnostics pane pulls its weight (2026-08-07) ⭐⭐

- ✅ Severity filter: the counts in the drawer bar became switches. In the store rather than the
  drawer, because the source gutter, the spec gutter and the file tree all read it — a filter that
  only thinned the list would be half a filter
- ✅ The pane resizes. It was the only thing on screen whose size was somebody else's decision. Height
  in pixels, not a percentage: you size it to hold rows, and rows do not scale with the window
- ✅ A diagnostic now points **both** panes. The other two directions start from a cursor, so the pane
  holding it already shows where you are and only the far side needs marking; a diagnostic sits in
  neither, so both have to be told. Spec→source had a smaller version of the same gap and now marks
  the node that answered

### The tree, and the spec's own gutter (2026-08-07) ⭐⭐

Reviewed by Fred: gutter, tree and diagnostic→spec tracking all confirmed working.

- ✅ File tree, collapsing single-child directory chains — `vendor/github.com/go-openapi/swag/mangling`
  is one row rather than five. The pane shows the tree **or** the file, as genspec-tui's left pane
  does: at that width, side by side leaves neither enough room to read
- ✅ Files carrying a diagnostic show a severity dot, so a clean package looks clean
- ✅ The spec gutter now carries diagnostics, reached through provenance (nearest anchor to the source
  line, then where its pointer is written). Inexact, and better than a column that never renders
- ✅ The trail is on both sides, each naming what the other would light up

Still open in this area: a **severity filter** for the drawer (193 diagnostics on dockerctl), and
folding in the spec pane — not missed so far.

### Phases 3–6 — the panes are joined (2026-08-07) ⭐⭐⭐

Both panes are CodeMirror. That was the decision the rest followed from: a `<pre>` on the result side
would have meant two highlight mechanisms, two scroll-into-view paths, and a join exact in one
direction and approximated in the other.

- ✅ The document is **rendered by us**, one walk emitting lines and recording the lines each RFC 6901
  pointer occupies. `JSON.stringify` gives the text and nothing to find it by, and provenance is keyed
  by pointer — so that map is the only route from a spec node back to Go source
- ✅ The same walk emits YAML, which is what stops the format toggle dropping cross-references.
  Round-tripped through `js-yaml` in tests, including strings that look like numbers and one holding
  `: `
- ✅ Three tracking directions, two exact and one honest about not being. `forLine`'s tie-breaks are
  stated rather than left to sort order: downwards first (Go docs sit above what they document), then
  the shorter pointer (a line carrying a field and one of its keywords cannot separate them)
- ✅ Search ported from the TUI's semantics — `/` from empty, line-based, case-insensitive, `n`/`N`
  wrapping — including the property its own comment calls out: the **cursor** goes on the match, so a
  search navigates rather than only finds
- ✅ Edit-and-rescan, debounced, with a scan that finishes stale re-running rather than showing output
  for source that has moved
- ✅ The worker keeps the encoded tree; only changed files are sent

**Three bugs worth remembering**, all found without a browser or by a test:

- `EditorView.editable.of(false)` is the obvious reading of "read-only" and drops the caret with it —
  no arrows, no Home/End, no PgUp/PgDn. The recipe is `readOnly` **with** `editable` left on
- the active line was themed but `highlightActiveLine()` was never added, so nothing applied the class
- reporting the cursor on `docChanged` as well as selection meant every rescan fired spec→source and
  dragged the source pane to line 1
- `humanDuration` rounded after comparing, so 999.6 ms printed as `1000ms`; the Go original truncates
  and never had the bug

Bundle 60 KB → **445 KB / 154 KB gzipped**, CodeMirror and three language modes. 1.8% of the artifact.

### Rebase onto `on-demand-scanner`, and the vendored-module ceiling (2026-08-06) ⭐⭐

Onboarded the parallel loader/scanner work; branch is six commits on `8306eb7`.

- ✅ One conflict, add/add on `hack/genexportdata` — the parallel stream had built its own, ours
  evolved (`-dir`, `-with-annotated`, `GOWORK=off`, the `WithExportData` rename). Took the base's
  wholesale; our commit no longer adds those files
- ✅ The flag guard caught the new `CompiledDependencies` option and it got a real flag rather than an
  excuse, with the trade stated: a dependency's comments are not read, so `strfmt`'s `swagger:strfmt`
  marks are lost. Not exposed in the playground, which always runs `-loader=own` where it is inert
- ✅ Full retest green: both Go suites, all three WASI tests, the loader/export-data suites, every
  commit building standalone, the guest on the browser's exact argv, the node conformance probe.
  The large probe got **faster** — 3.1 s against 3.6 s — from `perf(packages): stop type-checking
  what nothing reads`
- ⚠️ Lint delta went 4 → 18 findings, all in four of our files. Not a regression on our side: the base
  did a sweep we did not inherit, and `noctx` / `gosec` on `exec.Command` in the WASI test helpers are
  new categories. Still the Phase 9 sweep, now bigger

**The ceiling.** 8 MB refused dockerctl outright. 64 MB now, warning band at 16, both from measurement:
25 MB of source scans in the guest in under four seconds holding 462 MB, ~18× amplification, so the
ceiling sits near what a desktop bears and past what a phone does. Weighing moved ahead of reading (the
old tally refused halfway, having already paid to read that far), a refusal names the heaviest directory,
and reading batches 64 at a time with a count.

Confirmed by Fred on the real tree: opens, scans in a few seconds, **instant on re-scan** — which is the
compile-once design paying off, the module being kept between runs.

### Phase 2 — foundations (2026-08-06) ⭐⭐

The app has a shell, a theme, and a state that belongs to a mount rather than to the module.

- ✅ `styles/tokens.css` — colour, spacing, radius, type, elevation, all named by role. Declared
  inside `@layer` and scoped to `.cs-root`: unlayered declarations beat layered ones regardless of
  specificity, so an embedding page overrides `--cs-accent` by setting it, with no `!important` and
  no guessing what hugo-relearn calls its own variables
- ✅ No `prefers-color-scheme` block. The theme store resolves "follow the system" in JS and always
  stamps `data-theme`, so each theme is stated once instead of the dark one twice — the two copies
  being exactly what drifts
- ✅ Shell: toolbar, resizable split, tabbed result pane, collapsible diagnostics drawer, status line.
  The split is keyboard-adjustable, which a drag-only splitter is not, and the ratio decides how much
  of either pane can be read
- ✅ Stacking measured on **the shell**, not the viewport: embedded in an article the app may be half
  the window, and it is our own box that decides whether two panes fit
- ✅ `Playground` is constructed per mount and reaches components through context
- ✅ The worker consumes Phase 1's envelope, so diagnostics are structured and the drawer, the
  severity pills and the status counts are real rather than placeholder
- ✅ Counts derived from the document rather than carried — the same reason `stats` was dropped
- ✅ A non-zero exit surfaces the guest's own line from stderr, and a truncated document (a scan that
  ran out of memory mid-write) reports that instead of a JSON parse error, which would blame the
  format for a resource failure
- ✅ Vendored files sit behind a toggle in the file picker: they are scanned either way, but in a
  vendored tree they outnumber the user's own by two orders of magnitude

⚠️ **Not visually verified.** No browser was reachable from this session, so the evidence is
svelte-check clean over 178 files, 25 unit tests, and a build. What the shell actually looks like is
unreviewed — which is itself the open question in Phase 9's testing item.

### Phase 1 — the result contract (2026-08-06) ⭐⭐

`-format=json` wraps the document with everything the scan observed. `spec` stays the default, so the WASI tests,
the browser probes and any pipeline are untouched.

- ✅ `{spec, diagnostics[{severity,code,message,file,line,col}], provenance[{pointer,file,line,col}]}`
- ✅ `stats` **dropped** from the shape as planned-but-wrong: counts come from the document and elapsed time is the
  caller's own clock, so it would have been a second source of truth for nothing
- ✅ Positions relative to `-workdir`, so a caller who handed over `models/pet.go` is told `models/pet.go` and not
  the mount point the guest happened to see. Out-of-tree positions (GOROOT, module cache) stay absolute rather
  than becoming a chain of `..` — which is also how a consumer tells "yours" from "not yours"
- ✅ Anchors sorted by pointer (a lookup table, so emission order says nothing, and sorting makes a scan
  reproducible); diagnostics keep report order; both arrays empty rather than null
- ✅ Under `-format=json` stderr stays clean, which is what makes the envelope safe to read from a pipe
- ✅ Six tests, including the compatibility guard on `-format=spec` and the deliberate absolute-path exception

Verified through the guest, not only natively: under wasmtime with the module mounted at an arbitrary path, the
anchors come back as `models/m.go` with the right lines.

One thing the tests pinned that was not obvious: a dropped-description hint marks **the prose that was dropped**,
a line above the field it belonged to. A gutter mark drawn from it is meant to sit there.

### Phase 0b — the vendoring premise had a gate, and the picker closed it (2026-08-06) ⭐⭐

`FilePicker.wanted()` kept `go.mod` and non-test `.go` files. `vendor/modules.txt` is neither, so it was dropped —
and that file is precisely what `Resolver.vendorMode` reads to decide a vendor directory is authoritative. A user
doing exactly the documented thing got their vendored tree ignored, every dependency synthesized, and a wall of
warnings pointing nowhere near the mistake. Silent, and on the one path the whole fidelity story rests on.

- ✅ `wanted` moved from the component to `lib/tree.ts` and now filters on the **path**, not the basename — a
  `modules.txt` is only kept when it is the vendor directory's
- ✅ Unit tests for the filter, including the two ways `vendor/modules.txt` arrives (rooted at the module, or under
  the picked directory's name) and the `docs/modules.txt` that must not be mistaken for it

Confirmed against a real `go mod vendor` of codescan: `modules.txt` is the only non-Go file that matters, and a
vendored tree carries **no** nested `go.mod`, so re-rooting on the outermost one stays correct.

## Appendix — what the testing analysis found

Kept because the decision was to defer, not to dismiss, and whoever revives it should not have to
redo this.

**What the suite reaches.** 104 tests over 10 pure modules. Seventeen components rendered by nothing,
and the store — where most of the behaviour actually lives — with no tests at all, despite being plain
TypeScript.

**Measured against the eight things Fred caught that every check passed:**

| what was caught | a store test | a headless browser |
|---|---|---|
| the shell collapsed to content height | no | yes |
| the artifact predated the `-format` flag | no | yes |
| arrow keys dead in the spec pane | no | yes |
| no active-line highlight anywhere | no | yes |
| Swagger UI stayed white under the dark filter | no | yes |
| a diagnostic lit one pane, not both | **yes** | yes |
| the spec gutter carried nothing | no — a missing feature | no |
| the gallery button did not look like a control | no — a judgement | no |

**The uncomfortable half.** I assumed the diagnostic bug needed a browser. It did not. `Playground`
instantiates and drives in vitest — `$derived` class fields included, proven — and that bug is four
lines:

    p.reveal('models/pet.go', 6);
    expect(p.trackedSource).not.toBe(null);
    expect(p.trackedPointer).toBe('/definitions/pet');   // was null

So the gap is not only "no browser". It is that the pure helpers were tested thoroughly and the thing
they are assembled into was not tested at all. That is the cheaper half and the one to do first.

**Costs, if revived.** Store tests: no new dependency, runs in the existing vitest. Browser smoke:
Playwright plus a ~150 MB chromium, and a job that must build the 20 MB artifact before it can run.

## Doc-site integration (2026-08-08)

Beyond the plan's nine phases: the playground now runs inside a documentation page.

- ✅ The app mounts into any `[data-codescan-playground]` element and takes its artifact URL from a
  data attribute. It was hardwired to `#app`, which made it a page rather than a component
- ✅ `npm run dist` builds the artifact, builds the app, and copies both under
  `hugo/themes/codescan-static/playground/`. It refuses a half pack: building the app alone leaves it
  pointing at a scanner that is not there, and that failure surfaces in a browser
- ✅ The entry pair keeps fixed names so the shortcode needs no build manifest; everything else stays
  content-hashed. `base: './'` keeps the pack location-independent
- ✅ Pack gitignored, `README.txt` committed beside it — the Relearn and Mermaid pattern already in
  `themes/`
- ✅ `{{< playground >}}` shortcode, and a page at `docs/doc-site/playground.md` (weight 3)
- ✅ Verified by building the site and serving it under the real base path: page, entry pair, worker
  chunk and the 19.5 MB artifact all resolve

**Two bugs worth remembering, both about styling somebody else's page.**

The help dialog carried `class="cs-root"`, which re-declares the *light* tokens on the dialog itself
and beats the dark ones inherited from the shell. A `<dialog>` in the top layer still inherits through
the DOM, so dropping the class was the whole fix.

Its key caps were unreadable because the colour was never stated: Relearn sets
`kbd { color: var(--INTERNAL-TEXT-color) }`, light under a dark variant, on the light box we drew. The
`@layer` contract means our styles win on specificity **only for properties we actually state** —
so inside a host page, every colour has to be spelled out even where it looks redundant.

- ✅ CI builds the pack in `update-doc.yml`, after the theme fetch and before Hugo. Two things the
  runner needed that a local build does not: the sparse checkout had to grow `cmd/genspec-wasi/` and
  `internal/`, and `GOWORK=off` — `go.work` names modules a sparse checkout does not fetch, so every
  go command in the workspace fails on them. Verified by reproducing the sparse tree locally: without
  it, `cannot load module ../../../cmd/genspec-tui listed in go.work file`.

## Won't do

- ⛔ **Share link** (2026-08-07). The idea was a URL carrying its own content, as some popular
  playgrounds do, and it belonged to a wider effort that is parked. It does not carry its weight here
  on its own: a fragment holds a few kilobytes, which covers an edited example and nothing else — an
  opened module is three orders of magnitude past it — so the feature would exist for the one case a
  visitor is least likely to want to send anyone. May be revived with that wider effort.

## Appendix — risks and things that will bite

- **Provenance coverage is anchors only.** Type declarations, fields, values, route/meta blocks. A visitor clicking
  a `minimum` keyword gets the field, not the keyword. That is the design, and the UI has to make the granularity
  legible rather than feel broken.
- ✅ **The upload ceiling stands, and the alarm was mine to withdraw.** I raised it from process RSS
  (885 MB), which counts the WASI runtime and the compiled module — neither of which a browser tab
  pays for. The figure that matters is the guest's linear memory: **394 MB** for dockerctl's 25 MB of
  source, about 15×, in line with the 18× the 64 MB ceiling was sized on. Now measured on every scan
  rather than by hand.
- 🔍 **The scan holds nearly everything it touches.** 372 MB obtained, 352 MB still live, two
  collections. Not churn — retention, and the type graph is the obvious suspect. The on-demand-scanner
  stream is where that would change; noted here because the ceiling depends on it.

- **The artifact more than doubles.** Embedding the standard library's export data takes it from 3.7 MB to
  **8.4 MB gzipped** (15 MB → 20 MB raw). Phase 8 lands on an example and scans immediately, which now means
  8.4 MB has to arrive before the first frame can show output. The loading design is a feature, not a spinner.
  The alternative — fetching `exportdata.zip` into the guest filesystem — keeps the artifact at 3.7 MB and makes
  the 4.2 MB independently cacheable; `flags.ts` and `.gitignore` both already carry that variant. Open in
  `wasi-build.md` as 🔍 "How the standard-library types get there"; see the note below on why we are not deciding
  it here.
- **Swagger UI needs a whole document.** `tabbed-examples.md` records this as a blocker for the tutorials' fragment
  goldens. It is not a blocker for us — we always produce a whole spec — but a scan that produced no `paths` still
  renders as an empty Swagger UI, which looks like a failure. Detect and say so.
- **Auto-scan on a large tree.** Seconds of solid CPU in a worker. Debounce is not enough on its own; the
  in-flight scan has to be discardable and the UI has to say a scan is stale rather than silently showing old output.
- **Editor weight.** CodeMirror plus a Go mode is real bytes on top of a 24 KB app. Still two orders of magnitude
  under the artifact, but the weight budget should be stated rather than discovered.
