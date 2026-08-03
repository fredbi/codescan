# WASM-based Playgrounds for go-openapi — and the codescan case

Date: 2026-06-01 · **superseded 2026-08-02**
Status: 📚 origin vision — kept for provenance. **The live plan is `wasi-build.md`.**

> [!IMPORTANT]
> Superseded. A prototype now exists and runs in a browser; **the live plan is `wasi-build.md`**.
> Read this for how the thinking got there, and the section below for what it got right and wrong.

## What this document got right, and wrong

Written before any of it was built, so it is a record of reasoning rather than of behaviour.

**Held up.** The real coupling is `packages.Load`, not "WASM can't do I/O" (§3) — everything below it
cross-compiles untouched. Pre-published `gcexportdata` blobs are the answer for dependency types (§5b).
The hard part is the multi-file/import UX, not the type-checking (§6). The TUI shares the core and was
worth building first (§7).

**Did not.**

- **The target is WASI, not `GOOS=js`.** §4 assumed `js/wasm` and an in-memory importer. A WASI guest
  has a real filesystem, so the host mounts one and the "inject sources" problem dissolves.
- **There is no `Loader` seam to introduce** (§4). `go/packages` has no injectable driver — the protocol
  is exec-only — so the loader had to be written outright, as `internal/packages`.
- **`NewFromBytes` (Phase 2) is not the shape.** Nothing is passed as bytes; the guest reads a filesystem
  the host provides.
- **Synthesize-from-usage is a fallback, not a tier** (§5a). It exists, it is diagnosed, and it loses
  structure — fine for degradation, wrong as a default.
- **Export data cannot cover annotated libraries.** §5b hoped `strfmt` blobs would resolve `strfmt.*`.
  They do not: strfmt declares its formats in *comments*, which export data drops. Only its source
  carries them, so a vendored upload is the honest route — closer to §8's note (F) than to §5b.
- **The binary-size guess was low** (§9): about 14 MB, 3.6 MB compressed, before any dependency data.

## TL;DR

- Build interactive, **statically-served** playgrounds for go-openapi libraries:
  a `.wasm` asset compiled once, fed dynamic input text from the browser, with
  no server and no backend (Hugo + GitHub Pages, which we already master).
- Most go-openapi libs are **data-in / data-out** (JSON/YAML → result) and are
  trivial to expose this way: no I/O, no `go/types`, no imports.
- **codescan is the hard case** because it consumes *Go source* and leans on
  `go/types` + package/import resolution. It is the worst-case for the pattern.
- The apparent blocker ("WASM can't do file I/O") is **not** the real problem.
  The real coupling to the host is `golang.org/x/tools/go/packages.Load`, which
  *execs the `go` toolchain*. Everything below it (`go/parser`, `go/types`,
  the scanner index, every builder, `go-openapi/spec`, JSON marshaling) is pure
  computation that already cross-compiles to `js/wasm`.
- The fix is a **`Loader` seam**: native build → `packages.Load`; wasm build →
  `go/parser` + `go/types` over in-memory source. This seam is the *same*
  abstraction the demand-driven v2 scanner wants
  ([[index-builder-statefulness]] "Step 2: DemandLoader").
- **Two senses of "imports" — keep them separate (§5).** *Special-semantics*
  imports (`time`, `encoding/json`, `strfmt`) are recognized by name and need no
  real types. *Drill-down* imports (a field of type `otherpkg.Struct` the scanner
  must walk) need **real type info**. The user's own types are already in the
  editor; external drill-down resolves via **lazily-fetched, pre-published
  `gcexportdata` blobs** (static assets — `fetch()` is allowed in WASM even
  though file I/O is not). Genuinely arbitrary third-party imports cannot be
  drilled statically → degrade with a diagnostic, or paste their source / use the
  TUI.
- The genuinely hard part is **UX for multiple files and imports**, not the
  type-checking. A local **TUI** sidesteps all of it and shares the same core.

---

## 1. Motivation

Two concrete needs:

1. **Internal**: a faster loop for testing and for supporting users — paste a Go
   file, see the generated spec and the diagnostics, immediately.
2. **External**: a public "codescan playground" demo on the doc site, and more
   broadly a **reusable WASM-playground pattern** for the many go-openapi libs
   whose behavior is I/O-free and easy to demonstrate in a browser.

The proposed UI is three zones:

```
┌──────────────────────┬──────────────────────┐
│  Source (left)       │  Generated spec      │
│  editable Go file(s) │  (right) — JSON/YAML  │
├──────────────────────┴──────────────────────┤
│  Diagnostics (bottom)                        │
└──────────────────────────────────────────────┘
```

The bottom zone maps directly onto the existing `Options.OnDiagnostic`
callback (`internal/scanner/options.go`) — no new machinery needed to surface
warnings/errors.

---

## 2. The general pattern (the easy libs)

For libraries that are pure `string → string` transforms, a WASM playground is
almost trivial. There is no `go/types`, no imports, no FS:

| Lib | Demo |
|-----|------|
| `validate` | paste a spec → validation report |
| `spec` / `analysis` | expand / flatten / mixin / analyze a spec |
| `strfmt` | try format parsing/validation on sample values |
| `swag` | name mangling, JSON↔YAML, type utilities |
| `jsonpointer` / `jsonreference` | resolve a pointer against a document |

Shape: `GOOS=js GOARCH=wasm`, expose one `js.Func`
(`globalThis.run = (input) => output`), wire it to a textarea + an output pane.
The whole library runs in the browser. The `loads` lib is a mild exception
(reads files/URLs) but its core parsing is pure once the bytes are in hand.

**Strategic implication**: the *harness* (Hugo page + editor + wasm glue +
layout + theming) is reusable across all of these. Build the harness once on an
easy lib, then drop codescan into the same harness.

---

## 3. Why codescan is the hard case

codescan's input is **Go source code**, and its output depends on **type
information**, not just on the text. The pipeline is:

```
packages.Load ──▶ TypeIndex (scanner) ──▶ builders ──▶ *spec.Swagger
   (toolchain)      (pure)                  (pure)        (pure)
```

Only the first box touches the host environment. `packages.Load` with
`NeedTypes|NeedSyntax|NeedTypesInfo` (`internal/scanner/scan_context.go:24`)
**execs `go list`** and reads GOROOT + the module cache. Exec + that I/O are
what the browser sandbox forbids — *not* reading the user's source, which is
already in memory as the editor's text.

So the problem is precisely: **how do we obtain `*ast.File` + `types.Info`
without the `go` command?**

---

## 4. The reframe and the answer

Drop one layer below `go/packages` to two pure-Go libraries:

- **`go/parser`** — `parser.ParseFile(fset, "playground.go", src, parser.ParseComments)`.
  Pure; takes the editor string directly. No I/O.
- **`go/types`** — `(&types.Config{Importer: imp}).Check(...)`. Pure computation;
  its *only* external dependency is the `Importer` (how `import "x"` resolves).

The importer is the entire ballgame, and the codescan recognizers make it
tractable — see §5.

### The `Loader` seam

```
type Loader interface {
    // minimal contract the scanner needs: parsed + type-checked
    // package(s) for the entities under scan.
}

native  (//go:build !js): packagesLoader  → packages.Load   (full fidelity)
wasm    (//go:build js):  inMemoryLoader   → go/parser + go/types + fake importer
```

Two payoffs:

1. It makes the WASM build possible while keeping the **same scanner + builder
   core** below the seam — so the playground is a *faithful* repro of the real
   tool, which is the whole point for support/testing.
2. It is the **same abstraction the v2 demand-driven scanner wants**
   ([[index-builder-statefulness]] "Step 2: DemandLoader"). The playground is a
   forcing function for carving `packages.Load` out from behind an interface —
   work that pays into the architecture roadmap regardless. Build-tagging
   `golang.org/x/tools/go/packages` *out* of the wasm build also keeps the
   binary small (it drags in `os/exec`).

---

## 5. Imports — two distinct problems (do not conflate)

There are **two** unrelated senses of "imports" in this design. Conflating them
is a trap (an earlier draft did).

### 5a. Special-semantics imports (recognized, never drilled)

`time`, `encoding/json`, `unsafe`, `strfmt` … — imports codescan *recognizes by
name* and maps to a fixed schema. codescan keys these off
`(package path, type name)`, **not** the type's fields
(`internal/builders/resolvers/assertions.go:96,104`):

```go
func IsStdTime(o *types.TypeName) bool { return o.Pkg().Name() == "time" && o.Name() == "Time" }
```

For *these*, a **synthesize-from-usage** importer suffices (fabricate a package
whose scope holds the referenced names as opaque types with the right
path/name). No real export data needed. The one nuance is **behavior-based**
recognizers — `IsTextMarshaler` (`assertions.go:77`, via `types.Implements`)
needs a real method set, which an opaque type lacks (affects `strfmt`); covered
by 5b's real export data.

This part is *not* the hard problem.

### 5b. Drill-down imports (the real constraint Fred raised)

The scanner's **core job** is to walk a field of type `otherpkg.SomeStruct` into
`SomeStruct`'s real fields and emit a faithful definition. That requires the
imported package's **actual type information** (fields, nested types,
transitively). Synthesize-from-usage produces opaque types → good for a `$ref`
*name*, useless for *drilling*. So it cannot answer this case.

Fred's conclusion stands: for **arbitrary, unpredictable** third-party imports
there is no way to drill in a static context — we simply don't have their types.
But two refinements make it less bleak than "stdlib only, embedded":

**(1) The types you most need to drill into are the user's own — already in the
editor.** In a single-file / N-tabs-one-package playground, the user's own types
are present in the source, so `go/types` drills them with zero external data.
The drill-down gap is *only* external imports.

**(2) WASM forbids file I/O and exec — but NOT `fetch()`.** So embed nothing;
instead **pre-publish per-package `gcexportdata` blobs as static assets** and
fetch them lazily:

- Build step generates `/exportdata/<import/path>.gob` for a chosen set.
- The wasm importer does `fetch("/exportdata/...")` on demand, caches it, feeds
  `gcexportdata`. Export data carries full *exported* declarations (exactly what
  codescan emits — it only renders exported fields) and `gcexportdata` resolves
  **transitive** deps through the same importer, so
  `pkgA.Foo{ B pkgB.Bar }` drills correctly when `pkgB`'s blob is also fetchable.
- This is the **same `Loader`/DemandLoader seam** with an HTTP backend instead of
  `packages.Load`. Stdlib is the obvious published set (versioned to the
  toolchain); **go-openapi's own libs** can be published too (so `strfmt.*` and
  our types drill faithfully — also resolves 5b's method-set need).

### The honest gradient (centered on drilling)

| Type being drilled | In static WASM | Fidelity |
|---|---|---|
| User's own types (editor tabs, 1 pkg) | type-checked from source | **Full** |
| Stdlib / any package we pre-publish export data for | fetch `.gob` on demand | **Full** (exported fields, transitive) |
| Recognized special types (`time.Time` …) | name match (5a); no drilling | Full (by design) |
| Genuinely arbitrary third-party | **impossible** statically | degrade: opaque `$ref` + diagnostic |

Escape hatch for the last row: **(a)** user pastes the dependency's source as
extra tabs (the multi-package UX — faithful but heavy), or **(b)** use the TUI
(real toolchain). Otherwise the diagnostics panel honestly says "third-party
type `x.Y` not resolvable in the playground" — degradation as a teaching signal,
never a crash.

So the refined position is **not** "stdlib only embedded in the binary" but
"**any package we pre-publish export data for, fetched lazily as static
assets**," with stdlib (+ our own libs) the natural published set.

---

## 6. The hard part: multi-file / multi-package UX

Type-checking is *not* the hard part. Scan **scope** and its UX are. The
difficulty grows along one axis — how many packages are in scope:

| Scope | Type-checking | UX | External imports (§5b) |
|---|---|---|---|
| 1 file = 1 pkg, 1 editor | trivial | trivial | fetched export data or degrade |
| N files = 1 pkg, N tabs | easy (`Check` takes a fileset) | moderate (tabs) | fetched export data or degrade |
| N packages | hard (must parse+check all in-project pkgs) | heavy (file tree / memfs) | intra-project pkgs must be type-checked from pasted source |

The sweet spot is **N files / 1 package** (tabs): it covers most realistic
annotation examples (models + handlers + responses in one package), keeps
type-checking simple. The user's own types drill from source; external imports
drill from fetched export data (§5b) or degrade with a diagnostic.

True **multi-package** scanning is where it gets genuinely hard: to walk into an
in-project package's types (codescan emits their definitions), those packages
need *real* type bodies — i.e. a virtual FS tree + a project-aware loader. That
is a large UX and loader investment.

NOTE: js framework pick: likely svelte.

### Two modes (a pragmatic split)

- **Scratchpad mode**: free-form single-/multi-file editing in one package. Own
  types drill from source; external imports drill from fetched export data (§5b)
  or degrade gracefully. *(Ships first; proves the pipeline.)*
- **Guided-example mode**: a few curated example modules (e.g. petstore) shipped
  with **pre-generated export data for their full dependency set**. The user
  edits *within* a known module whose imports are all pre-resolved → full
  fidelity, no arbitrary-import problem. *(Higher fidelity, fixed scope.)*

These two modes bracket the fidelity/effort trade-off cleanly and let us ship
value before solving the general multi-package case.

---

## 7. The TUI escape hatch (local, easy, shared core)

For local/internal use a **TUI** is dramatically easier: it runs natively, uses
the real `packages.Load` (full fidelity, real multi-package, real imports), and
needs none of the importer/UX gymnastics. Same three zones in a terminal.

Crucially it **shares the same core**: TUI = native loader, WASM = in-memory
loader, identical scanner/builder pipeline + `OnDiagnostic` below. The only real
investment is the `Loader` seam; the two front-ends are thin. Building the TUI
first also exercises the seam and the three-zone interaction model on the easy
side before the wasm constraints bite.

→ The concrete, agreed phase-by-phase plan is **§8**.

---

## 8. Sequencing — TUI → WASM → IDE (Fred's plan, annotated)

Agreed shape: build the **TUI first** (full fidelity, native, easy), let it
drive the diagnostic contract and the interaction model, then reuse the *core +
model* (not the views) for the WASM web app, then IDE integrations.

### Phase 1 — Local TUI spec generator

Independent module `cmd/genspec-tui/go.mod` (codescan becomes a **monorepo**:
`go.work` + multiple `go.mod` + adapted CI) so TUI deps (bubbletea, …) don't
pollute the lean library. Model app reference: `fredbi/git-janitor`
(multi-panel, async task orchestration, clipboard integration).

**Primary mode = watch + view, not edit.** The user often edits in their own IDE
in another terminal; the TUI watches the tree and re-renders. In-TUI editing is a
secondary convenience. **Disk is the source of truth** — edits (TUI's own or an
external editor's) land on disk; the watcher drives regen. No overlay, no
in-memory caching. (Rationale: incremental spec change is ~impossible today;
full-scope regen is the honest model.)

UX: left = source tree browser (navigate + optionally edit current file);
right = generated spec (Ctrl-J/Ctrl-Y toggles JSON/YAML); bottom = scrollable
diagnostics. Files with diagnostics get an orange marker in the tree (scope =
file, **no position yet**). Spec / source / diagnostics copy to clipboard.
Regenerates on **any** file change (fs watcher + debounce; async render so
`packages.Load` latency never blocks the UI).

Out of scope: rich editing/highlighting/search (a later VIM, then VS Code,
integration supersedes this), token-accurate diagnostics.

Demonstrates: full toolchain works instantly; diagnostics work to file level; a
useful, comfortable UX we can iterate on. Precursor to IDE work.

**Decisions (settled 2026-06-01):**

- ✅ **(C) Generation unit = ONE spec for the whole scanned scope**, regenerated
  on any change. Not per-file (a model in file A may be referenced by a route in
  file B; per-file specs would be a new scanner capability we don't have).
- ✅ **(B/D) No overlay, no in-mem caching.** Disk is the source of truth; the
  watcher reacts to on-disk edits. (Drops the earlier `Options.Overlay` idea.)
  Keeps "no scanner changes" intact for the watch/render core.
- ✅ **(E) Branch off `master` in `.worktrees/master`** (a new branch, yet to be
  created) — **NOT** `feat/new-parsing-layer`. The TUI work itself advances our
  (so-far theoretical) diagnostic capabilities. Once `feat/new-parsing-layer`
  merges (approaching readiness), the TUI branch **rebases onto it** and inherits
  grammar2 + real diagnostic emission.
- ⚠️ **(A) Public `codescan.Diagnostic` contract — defined by the TUI branch.**
  master-now has **no `OnDiagnostic` channel** (that's feat-branch), so the TUI
  defines the public diagnostic shape on its own terms — `{File, Severity,
  Message}` now, `{Line, Col}` later. Post-rebase, `grammar2.Diagnostic`
  (`internal/parsers/grammar2/diagnostic.go:86`) maps *into* this public contract.
  This avoids ever leaking the internal type and lets the consumer drive the
  shape that P2/P3/LSP also use.
- 📉 **Diagnostics ladder.** Pre-rebase the bottom panel shows only `Run`'s single
  returned error; post-rebase it fills with rich per-file diagnostics. Build the
  panel against the public contract from day one regardless.
- ℹ️ `go.work` is dev-only (`go install …@latest` ignores it) → TUI go.mod needs
  a real `require codescan vX`; release order = tag codescan first; CI gains a
  per-module matrix.

#### Phase 1 status (2026-06-01) — milestone reached

**No longer a POC — a useful tool that could ship.** It renders a spec live
from source locally, with a reasonably rich UX. Pushed to `origin` + the
`fredbi` fork (no PR yet). Monorepo CI adopted (go-test-monorepo, etc.).

Shipped on `feat/genspec-tui`: scaffold → logger-mute → fsnotify live-reload →
header/spinner/stats + mouse(click-focus/wheel) → in-spec search(`/` `n`/`N`) →
scanner-options popup(`o`, descriptions) → tree↔spec linkage **seam** →
scan-duration in header → file viewer → **basic editor** (`bubbles/textarea`,
line numbers, `Ctrl-S` save → watcher → rescan; `●` dirty marker) → monorepo CI.

**What this milestone buys us (beyond the tool itself):**

1. A **model for an ergonomic UX** — readily adaptable to a JS front-end (the
   exact JS stack is TBD — an open question for Phase 2).
2. A proven **interaction model with the scanning library** (load → scan →
   render JSON/YAML → diagnostics → live reload → options → linkage).
3. The WASM version can **reuse most of this logic, if not the code**, so Phase 2
   narrows to three concerns: (a) the wasm port, (b) user-file integration,
   (c) the web UX (Hugo + JS asset).

**Now effectively paused, pending the grammar2 rebase.** Next step once rebased:
**wire diagnostics + implement the cross-ref feature with visual linkage.** The
two big remaining features need data this branch doesn't have yet:

- **Real diagnostics panel** — bottom pane + orange tree markers fill in once
  builders emit `grammar2.Diagnostic` through the public `codescan.Diagnostic`
  contract (decision A).
- **Position-backed cross-ref linker** — replace `naiveLinker` (name match) with
  the source↔spec position map; make it bidirectional (file line ↔ spec node),
  landing the cursor on a line in the editor. The line-numbered editor is ready.

#### Phase 1 backlog — TUI chrome (polish, not rebase-gated)

- ✅ **JSON/YAML syntax highlighting** in the spec pane (3fa4aaf). chroma was
  NOT used: the lexer that builds the line↔pointer index already classifies
  every token and we were throwing that away, so highlighting became a third
  product of the same walk — zero new dependencies. The composition problem the
  chroma sketch would have hit (truncating an already-coloured string cuts
  through escapes) dissolves with `(line, col, kind)` spans: truncate the RAW
  text at rune boundaries, style last.
- ✅ **Go syntax highlighting** in the read-only source viewer (fb945b7), via
  `go/scanner` onto the same spans/renderer/palette. Easier than this entry
  assumed, because the file pane is no longer only a textarea — the viewer has
  its own per-line render loop, so it is the same job as the spec pane.
  Annotation comments get the spec-key class rather than the dimmed comment
  class: in a spec generator the `swagger:` comment is the payload.
  **Edit mode is still out of scope** — `bubbles/textarea` emits the buffer
  verbatim, so highlighting there means replacing the widget, i.e. the
  VIM/VS-Code editor that supersedes the hand-rolled one.
- ✅ **Grammar-keyword highlighting** (342b3be) — `required:` / `min:` / `enum:`
  inside an annotated comment read as keywords rather than dimmed prose, via
  `grammar.Lookup` so aliases and case come from the parser's own table. Scoped
  per FILE, not per comment group: a field's constraints sit in the field's doc
  comment while the `swagger:model` that gives them meaning is on the type.
- ✅ **Diagnostics at the site** (47b4c13) — each finding underlined in its
  severity colour on the run it points at, re-derived on every rescan. Started
  as "colour deprecated keywords"; there is no deprecated-keyword table, and
  `validate.deprecated` reports at the DECLARATION, so a deprecation-only visual
  would have read as "this type is deprecated". Driving from the diagnostic
  stream keeps one source of truth.
  - **Open upstream nit:** `parse.invalid-enum-option` reports the column of the
    space BEFORE the value (`// collection format: pipe` → points at `" pipe"`),
    so the mark starts one column early. Cosmetic; fix belongs in codescan.
- **Save / reload file (when buffered-in).** Save exists (`Ctrl-S`); add an
  explicit **reload** (e.g. `Ctrl-R` / `F5`) to re-pull the open file from disk
  into the buffer — with an unsaved-edits guard (confirm before discarding). The
  auto-reload was removed because it clobbered edits; a manual, guarded reload is
  the right replacement (useful when the file changed on disk externally).
- ✅ `?` help overlay (3f872dd) — plus an `h: help` chip in the header, since a
  help key nobody can see is not a help key.
- Smaller, still open: adjustable split sizes; light/dark theme.

### Phase 2 — WASM web spec generator

Reuses Phase 1's **core + interaction model** (not its terminal views): the wasm
exposes a pure `Generate(sources) → (spec, diagnostics)` JS function; the page is
plain HTML/JS. All input is client-side — no real upload, no server.

This bypasses `packages.Load` via a new constructor (**`NewFromBytes`**): build
an in-memory FS from the input → `go/parser` all `.go` → `go/types.Check` with
an importer resolving (a) in-tree packages from source, (b) stdlib from an
embedded asset, (c) else → diagnostic → hand-assemble `[]*packages.Package` →
**same `NewTypeIndex`**. Confirmed cheap: the index/builders read only ~8 fields
of `*packages.Package` (`ID, PkgPath, Name, Syntax, Types, TypesInfo, Fset,
Imports`). `packages.Load` gets `//go:build !js` so its `os/exec` deps never
enter the wasm binary. First step: verify all imports resolve; report
unresolved ones as diagnostics.

### Input ladder (lead with the cheap tier — archive is tier 3 only)

1. **Single editable file, stdlib-only** — the DEFAULT doc-site experience. Zero
   friction: paste annotations, see spec. No upload.
2. **Multiple in-memory editor tabs, one package** — still no upload; covers most
   multi-file examples by typing.
3. **Project import** — the "bring your real, multi-package project" escape hatch.
   Higher friction; fewer users. Preferred intake = **`<input webkitdirectory>`
   directory pick** (no archiving step; Chromium-solid, patchier in FF/Safari);
   **archive (tar/zip/tgz) = portable fallback**. Same `NewFromBytes` behind both.
   See: https://developer.mozilla.org/en-US/docs/Web/API/HTMLInputElement/webkitdirectory

### Mechanics & assessment (workable: yes; default: no)

- **All pure-Go unpack.** `archive/tar`, `archive/zip`, `compress/gzip` compile to
  wasm — the JS "scriptlet" is trivial (read file → `js.CopyBytesToGo` → invoke);
  the real work stays in Go. **Pass the whole buffer in one shot — do NOT stream**
  (chunked JS↔wasm `io.Reader` is awkward; oversized = "too big for the
  playground" signal).
- **Run the wasm in a Web Worker.** Unpacking + type-checking a multi-package tree
  on the main thread freezes the UI. Post bytes to a worker → get back
  `{spec, diagnostics}`. (Single-file is fine on main thread; worker is cleaner
  for both.)
- **Parse `go.mod` via `golang.org/x/mod/modfile`** (pure Go) to get the module
  path + locate `vendor/`. Without it, intra-project import resolution is guesswork.
- 💡 **(F) Vendored project = full fidelity, zero export data.** A `go mod
  vendor`'d tree carries *all* deps as source under `vendor/`, so in-browser
  `go/types` resolves everything from source — no embedded/fetched export data.
  So: tier 1 → embedded stdlib export data (§5b asset); tier 3 → recommend a
  vendored tree for full multi-package drilling; non-vendored → external imports
  degrade with diagnostics.
- ⚠️ **Scale ceiling.** A real vendored tree (tens of MB / hundreds of packages)
  is seconds-to-tens-of-seconds + memory-heavy in wasm. Sweet spot = small/medium
  projects and curated examples. Add a size guard + honest diagnostic ("too large
  for the in-browser scanner; use the TUI"). The **TUI stays the tool for serious
  / large scans.**
- ✅ **Privacy is a selling point.** Pseudo-upload never leaves the browser →
  truthfully advertise "your code never leaves your machine." A real differentiator
  vs server-based playgrounds; lowers the bar for pasting proprietary code.

### Phase 3 — IDE integrations (later)

Same pattern as the TUI (local), but diagnostics become **token-level**. The new
challenge is position-aware diagnostics: e.g. mapping a position *inside an
embedded YAML operation block* back to file coordinates, plus other lexer
position fidelity ([[project_lsp_diagnostics_target]]). Out of scope until the
playgrounds prove the UX and Stream M lands.

### Optional pre-step

- ⬜ **P0 — Harness on an easy lib.** Stand up the reusable Hugo+WASM+editor
  harness on `validate`/`spec` (pure `string→string`, no `go/types`) to de-risk
  static hosting + wasm glue + layout *before* the codescan-specific WASM work.

---

## 9. Open questions

- **Logger injection / silent-by-default scanner.** *(Surfaced by the TUI,
  2026-06-01.)* codescan writes through the **process-global standard `log`
  package** (default sink: stderr), which paints over bubbletea's alt-screen.
  Immediate fix is consumer-side (`genspec-tui` `main` does
  `log.SetOutput(io.Discard)`). The reflection: a library should be **silent by
  default** and accept an **injected sink**. Two distinct output classes today:
  (a) *debug tracing* — `logger.DebugLogf`, gated on `Options.Debug`; (b)
  *semantic warnings* — `logger.UnsupportedTypeKind` + ~10 scattered
  `log.Printf("WARNING: …")` sites (`scan_context.go` DeclForType/PkgForType,
  `enum.go`, `schema.go` ×8) that are **NOT gated** and always emit. Proposed
  split: (a) → an injected `*slog.Logger`/`io.Writer` on `Options` (default
  discard; go-swagger CLI sets stderr); (b) → these are **diagnostics about the
  scanned code**, so route them through the existing `OnDiagnostic` channel
  (with positions) and the public `codescan.Diagnostic` contract (decision A) —
  unifying with the diagnostics work ([[project_lsp_diagnostics_target]]). The
  schema.go WARNING sites already hold a builder context to attach positions.
- **Loader contract.** What is the *minimal* interface the builders need from "a
  loaded package"? Today the surface is wide (`FindDecl`, `GetModel`,
  `DeclForType`, `PkgForType`, `FindComments`, `FindEnumValues`, the iterators,
  + `ExtraModels` mutation). The seam should expose the read-only subset; the
  mutation/discovery half is the [[index-builder-statefulness]] SchemaCache
  concern. Designing this is the spine shared by playground + v2.
- **Binary size.** `go/types` + spec libs in a Go wasm blob ≈ a few MB gzipped.
  Acceptable for a playground; measure early. TinyGo almost certainly won't work
  (`go/types` leans on reflection). Plan on the standard toolchain.
- **Export-data versioning.** `gcexportdata` blobs are tied to the Go release;
  regenerate in CI. One-line step; don't let it rot.
- **Synthesize-from-usage type vs value positions.** For §5a, the pre-walk must
  synthesize a *type* for `pkg.X` in type position vs a var/func in value
  position. For the schema use case, a targeted walk of type expressions
  suffices; value positions matter less. Confirm against real fixtures.
- **Export-data total size & fetch granularity.** Per-package blobs fetched
  lazily keep the working set small, but the *published* stdlib set is tens of MB
  on disk. Acceptable for static hosting? Per-package vs bundled-by-area? Measure.
- **Transitive fetch correctness.** `gcexportdata` resolves a blob's deps through
  the same importer — verify the fetch-importer satisfies that recursively and
  caches to avoid refetch storms.
- **Faithfulness ceiling.** Single-file + synthesized imports cannot reproduce a
  real multi-package scan. Be explicit in the UI: "playground green ≠ real scan
  green" when imports fall outside the bundle.
- **Which lib first for the harness?** `validate` (most demo-able) vs `spec`
  (most central). Either proves the harness; codescan reuses it.

---

## 10. Evidence index (so this stays a faithful reference)

- `internal/scanner/scan_context.go:24` — `pkgLoadMode` = the toolchain coupling.
- `internal/scanner/options.go` — `OnDiagnostic` → diagnostics panel.
- `internal/builders/resolvers/assertions.go:96,104` — name-based recognizers
  (`IsStdTime`, `IsStdJSONRawMessage`) → synthesize-from-usage works.
- `internal/builders/resolvers/assertions.go:77` — `IsTextMarshaler`
  (`types.Implements`) → method-set fidelity needs real export data (strfmt).
- `internal/builders/schema/special_types.go` — recognizer dispatch order.
- Related design: [[index-builder-statefulness]] (DemandLoader seam),
  [[vision]] (v2 directions).

## 11. Other ideas (future/maybe)

* Code subset selection: allow to exclude selected folders (packages) -> faster scan on selected areas
* Easy Issue reporting tool
- ease the process of reporting issues on public github:
  so far, it has been super painful to help users reporting issues as they rarely want to share
  their code. As a result, many issues remain pending, bereft of sufficient information to reproduce.
- The UX(es) (TUI+WASM) could generate some "anonymize/desensitivized" version of the source code,
  copy / paste with user annotations so pasting this as an archive attached to a github issue allows
  maintainers to reproduce and test without the users exposing their original code.
- ideas: strip all unused code (func bodies), rename types, gibberish description and title replacement...
- ideas: interactive gibberish replacement/renaming

