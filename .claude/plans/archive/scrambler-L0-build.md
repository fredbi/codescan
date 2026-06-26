# Scrambler L0 (Prune) — build plan

Date: 2026-06-02
Status: ⬜ ready to build — first milestone of the scrambler/issue-repro tool.
Design rationale: `anonymizer-repro-tool.md` (esp. §2 oracle, §3 contract, §4 L0,
§7 tensions, §12 UX). This doc is the executable task breakdown for **L0 only**.

## Scope

**In:** a self-contained `internal/scrambler` package implementing **Minimize
(prune)** as a files-in → files-out transform, plus the **§2 oracle realized as a
test harness** over the existing fixture corpus.

**Out (later milestones):**
- L2 (redact / gibberish prose + example/default strings).
- The public `codescan.Scramble(...)` wrapper and the **TUI issue-report mode**
  wiring (§12) — L0 ships as an internal, test-driven package first.
- The WASM in-memory loader (L0 uses `packages.Load` natively; the loader is the
  §Phase-2 seam — only the *import-removal* step must already be toolchain-free,
  and it is).
- Crash-repro path (assume the original scans without panicking).

## The shape: files-in → files-out

```
Minimize(opts) :
  load    : packages.Load(Dir, Patterns, Overlay=opts.Files)  → []*packages.Package (with TypesInfo)
  roots   : decls whose doc-comment carries a `swagger:` line          (scan.go)
  closure : transitive type-graph closure from roots (go/types Uses)   (reach.go)
  prune   : drop decls ∉ closure; empty kept func bodies (named-return) (prune.go)
  imports : pure-AST removal of now-unused import specs                 (imports.go)
  emit    : reprint each surviving file; drop emptied files & _test.go  (emit.go)
  → Result{ Files: map[path][]byte, Dropped: []string, Stats }
```

Why **overlay from the start:** in the eventual TUI mode the source is an
in-memory buffer, not disk (§12.2). `packages.Config.Overlay` (`map[string][]byte`)
feeds that buffer to the loader. For L0 tests we pass `Files: nil` → pure disk
load of fixtures. Same code path.

**The scrambler never imports `codescan`** — it's a pure files-in/files-out AST
transform with no reason to. The **§2 oracle** (which *does* call `codescan.Run`,
to scan before/after and compare) is a **test/consumer concern**, not part of the
package: it lives in `package scrambler_test` (free to import
`github.com/go-openapi/codescan`) for T7, and later in the TUI/cmd layer at
export time. The leaf package stays dependency-light; no cycle exists either way.

## Package layout

```
internal/scrambler/
  scrambler.go    # Options, Result, Stats, Minimize() entry + pipeline
  load.go         # packages.Load wrapper (mode = codescan's pkgLoadMode; GOWORK=off; Overlay)
  scan.go         # tiny scanner: annotation roots (swagger: doc-comments) — no dep on internal/scanner
  reach.go        # transitive closure over the type graph (TypesInfo.Uses + the rules below)
  prune.go        # filter decls; named-return body rewrite
  imports.go      # pure-AST unused-import removal (removal-only; no reorder/add)
  emit.go         # reprint surviving files (format.Node); assemble Result
  scrambler_test.go      # unit tests (body rewrite, imports, enum retention, …)
  oracle_test.go         # package scrambler_test: §2 invariant over fixtures (imports codescan)
```

## API (L0)

```go
package scrambler

type Options struct {
    Dir       string            // module / work dir (= codescan Options.WorkDir)
    Patterns  []string          // package patterns (e.g. ["./..."])
    BuildTags string            // passed as -tags
    Files     map[string][]byte // overlay: in-memory source; nil → pure disk
}

type Result struct {
    Files   map[string][]byte // surviving tree, path → reprinted bytes (dropped files absent)
    Dropped []string          // files removed entirely
    Stats   Stats
}

type Stats struct {
    FilesIn, FilesOut         int
    DeclsKept, DeclsDropped   int
    BodiesEmptied             int
}

func Minimize(opts Options) (*Result, error)
```

(No `Minimize bool` toggle yet — this package *is* the minimize pass. Toggles and
the public surface arrive with the wrapper milestone.)

## Tasks

- ✅ **T0 — skeleton.** `scrambler.go`: `Options`/`Result`/`Stats` + `Minimize`
  wiring load → roots → passthrough (emit stubbed). `errors.go`: `ErrScramble`
  sentinel. Compiles; returns input unchanged.
- ✅ **T1 — loader (`load.go`).** `packages.Load` with `pkgLoadMode`
  (`NeedName|NeedFiles|NeedImports|NeedDeps|NeedTypes|NeedSyntax|NeedTypesInfo`),
  `Tests:false`, `Env: GOWORK=off`, `Overlay: opts.Files`, `BuildFlags:["-tags",…]`.
  Collects in-scope `sourceFile`s (path/pkg/ast/src) over the shared `Fset`;
  fails fast on root-package type errors. (Kept local — [[project_gowork_scanning_gotcha]].)
- ✅ **T2 — roots (`scan.go`).** Per in-scope file/decl: a `swagger:` doc-comment
  line (raw scan, not `CommentGroup.Text()` which drops directive-style lines) →
  root `types.Object` via `TypesInfo.Defs`. Handles grouped `GenDecl`s per-spec.
  Tested against `enhancements/named-basic` (Email/Colour/Grade/User). Tests green,
  lint clean.
- ✅ **T3 — closure (`reach.go`).** Worklist from roots over merged
  `TypesInfo.Uses`; rules R1–R6 (signature-only walk for emptied funcs;
  whole-body walk for annotation-bearing bodies; methods R3, typed consts/vars
  R4, const-block grouping for iota/ref safety).
- ✅ **T4 — prune + body rewrite (`prune.go`).** Rebuild `Decls` to the kept set;
  **named-return rewrite** for emptied funcs; **bodies enclosing an annotation
  kept verbatim**; comment list re-derived via `CommentMap.Filter` + re-add of
  annotation-bearing groups.
- ✅ **T5 — import removal (`imports.go`).** Pure-AST, removal-only; accurate via
  `TypesInfo.Uses → *types.PkgName` (no name heuristic); blank/dot imports kept.
- ✅ **T6 — emit (`emit.go`).** `format.Node` per surviving file; files with no
  decls **and** no annotations dropped (so `doc.go` swagger:meta survives);
  `_test.go` dropped; `Result`+`Stats`.
- ✅ **T7 — oracle test (`oracle_test.go`, `package scrambler_test`).** Copies
  fixtures → applies minimized tree → re-scans → asserts **byte-identical** spec.
  Green across petstore, classification, and 5 enhancement fixtures. This caught
  three real bugs: mid-line annotations (`// X swagger:route`), package-doc
  `swagger:meta`, and `swagger:response` types declared **inside func bodies**.
- ✅ **T8 — unit tests (`scrambler_test.go`).** Loader, root detection, Result
  accounting. (Targeted body-rewrite/import/R3/R4 micro-tests folded into the
  oracle corpus, which exercises all of them; can expand if a regression needs a
  narrower repro.)

**Status: L0 complete — tests green, `golangci-lint` clean, module builds.**
The named-return + reachability + body-annotation handling all proven by the
byte-identical oracle over the real fixture corpus.

## Reachability rules (the closure contract)

The kept set must be a **superset** of everything the scanner reads (under-keep is
caught by the byte-identical oracle; over-keep just yields a bigger repro). The
closure follows:

- **R1.** From a root decl: every package-level object it references
  (`TypesInfo.Uses`).
- **R2.** For a kept named type: all types referenced by its underlying type —
  struct field types, embeds, array/slice elems, map key+value, pointer elems,
  chan elems, func param/result types, type-param constraints, alias targets.
- **R3.** For a kept named type: **its methods** (kept, bodies emptied — method
  *presence* drives `IsTextMarshaler`/`MarshalText`) and their param/result types.
- **R4.** For a kept named type: **package-level `const`/`var` of that type** —
  these are its enum members (the type doesn't reference them, so without this
  rule the enum vanishes from the spec). ← the subtle one.
- **R5.** Recognized special/external types (`time.Time`, `strfmt.*`,
  `json.RawMessage`, any non-in-scope package) are **leaves**: keep the reference
  + its import, never drill, never prune their files (not ours to prune).
- **R6.** Multi-package: the closure spans all *in-scope* packages; **keep package
  structure** (don't flatten — import paths can be scanner-relevant). Prune
  unreachable decls within each.

## Named-return body rewrite

For each kept `FuncDecl`/method with a `Body`:

```go
func f(x, y int) (*Type, error) { … }   →   func f(x, y int) (v1 *Type, v2 error) { return }
func g() (out Foo, err error)   { … }   →   func g() (out Foo, err error)         { return }   // names kept
func h(x int)                   { … }   →   func h(x int)                         {}            // no results
```

- Results unnamed (or `_`) → synthesize `v1, v2, …`, avoiding collision with
  receiver + param names.
- Results already named → keep names; body becomes a single bare `return`.
- No results → empty body.
- Bodyless decls (interface methods, asm/external) → untouched.
- Keep type params (generics), receiver, variadics intact.
- **Type-identity invariant:** named vs unnamed results are the *same* function
  type, so interface satisfaction (R3) is preserved.

## Reprint & comment fidelity (the main risk)

L0 mutates the AST and **reprints whole files** (`format.Node`) — reformatting is
fine (gofmt-clean output). The hazard: **annotations live in the doc comments of
kept decls**, and `go/printer` comment *attachment* is position-based, so removing
decls can misplace or drop comments.

- Mitigation: when filtering `file.Decls`, also prune `file.Comments` groups that
  belonged to removed decls; keep each surviving decl's own `Doc` intact.
- Backstop: the **byte-identical oracle (T7) runs over the whole corpus** — any
  comment misattachment shifts an annotation → spec changes → test fails loudly.
  So correctness is *measured*, not assumed.
- If `format.Node` proves fragile here, fall back to **decl-level source surgery**
  (keep original bytes; cut byte-ranges of dropped decls; splice rewritten bodies)
  — heavier, but preserves untouched bytes exactly. Decide empirically from T7.

## Decisions this milestone settles (from design §10)

- ✅ **Reachability rules** — enumerated above (R1–R6).
- ✅ **Verification strictness** — exact JSON equality (spec is deterministic).
- ✅ **No public wrapper needed** — verified empirically that the genspec-tui
  module (`github.com/go-openapi/codescan/cmd/genspec-tui`) **can import
  `internal/scrambler` directly** (the `internal/` rule is path-prefix based, not
  module-bound; `go.work`/`require` handle resolution). So the TUI consumes the
  internal package as-is and the public surface stays untouched
  ([[feedback_test_api_surface]]). A public package is needed *only* for a
  consumer outside the `codescan/` path prefix — not a current goal.
- ✅ **Multi-package** — keep structure (R6).

## Out-of-scope follow-ons (tracked, not built here)

1. **TUI mode** — issue-report mode (§12): `Ctrl+I` in/out, `m`/`s` toggles,
   pin-model buffer, spec remarks, `Ctrl+X` export sheet, export-time oracle
   check. Wires by importing `internal/scrambler` directly (no public wrapper).
2. **L2 — redact** — prose + example/default strings (comment-span classifier;
   source surgery; deterministic gibberish).
3. **WASM loader seam** — swap `packages.Load` for the in-memory loader.

## References

- `internal/scanner/scan_context.go` — `pkgLoadMode`, `GOWORK=off` loader knobs to
  replicate.
- `internal/scantest/load.go` — `FixturesDir()` + fixture loaders for T7/T8.
- `fixtures/enhancements/*` — small annotated one-package scenarios = the starter
  oracle corpus.
- go-swagger `generator/internal/language/format_lite.go` — pure-AST import
  removal reference (removal half only).
- Design: `anonymizer-repro-tool.md` (§2 oracle, §3 contract, §4/§12.3 transforms,
  §7 tensions). Memory: [[project_genspec_tui]], [[project_wasm_playground]].
