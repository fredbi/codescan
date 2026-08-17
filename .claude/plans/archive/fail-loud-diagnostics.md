# Fail-loud diagnostics (§8) — scoping plan

Branch: `feat/quick-wins` (current with `fix/backlog-lot1` tip 3c18adf).
Origin: `forthcoming-features.md` §8.1 (#2886, #2804) + §8.2 (#2874).

**STATUS: both phases landed & validated 2026-06-17 (NOT yet merged to lot1 —
awaiting Fred's review).** Phase A §8.2 = `65fd9f0`; Phase B §8.1 = `32c14c3`.
Full suite + lint green on each. New `scan.*` diagnostic family added.

## The problem

Two silent-degradation modes the scanner should make loud:

- **§8.1** — a panic deep in a builder (e.g. the #2886 SIGSEGV in the
  responses builder) aborts `Run` with a raw Go stack trace and **no
  indication of the offending source declaration**.
- **§8.2** — a degraded package load (the #2874 `GOROOT`-unset case:
  type info silently unavailable → `swagger:allOf` stops resolving)
  produces an incomplete spec with **no warning**. `packages.Load`
  only returns the catastrophic error; per-package `pkg.Errors` and
  missing `pkg.Types` / `pkg.TypesInfo` are never consulted.

## Existing infrastructure to reuse (no new plumbing)

- `grammar.Diagnostic{Pos token.Position; Severity; Code; Message}` with
  constructors `Errorf` / `Warnf` / `Hintf`; severities Error / Warning / Hint
  (`internal/parsers/grammar/diagnostic.go`).
- `Options.OnDiagnostic func(grammar.Diagnostic)` → `ScanCtx.OnDiagnostic()`
  accessor; `common.Builder.RecordDiagnostic` (appends + fires callback).
- `ScanCtx.PosOf(token.Pos) token.Position` and `ScanCtx.FileSet()`.
- Per-decl location: `EntityDecl.Ident.NamePos` (→ `PosOf`) and FQN from
  `Obj().Pkg().Path() + "." + Ident.Name`.

## New diagnostic codes (new `scan.` family)

- `CodeInternalPanic = "scan.internal-panic"` — Error severity.
- `CodeDegradedLoad  = "scan.degraded-load"`  — Warning severity.

(One code per concern; detail rides the Message. Could split degraded-load
into package-error vs missing-type-info later if needed.)

---

## §8.2 — fail-loud on degraded package load  *(Phase A — land first)*

**Where:** `scanner.NewScanCtx`, right after `packages.Load` (the
`opts.OnDiagnostic` callback is in scope there).

**Checks — TIERED by what is still recoverable (refined 2026-06-17 after an
empirical probe; see "Refinement" below):**
- **ABORT** (Error diagnostic + returned error wrapping `ErrDegradedLoad`):
  1. `len(pkgs) == 0` — nothing matched the patterns.
  2. a root pkg with a `packages.ListError` — could not be loaded at all
     (missing dir / unresolved import); "code must build" can't even be met.
  3. a root pkg with `Types == nil || TypesInfo == nil` — the #2874 wholesale
     type-check failure; nothing useful to emit and allOf silently breaks.
- **WARN + continue**: a root pkg with only `ParseError`/`TypeError` but
  **usable** type info. `packages.Load` type-checks best-effort, so the
  definitions remain scannable — a single non-building package must not sink a
  whole `./...` scan. The spec emits from what loaded, package flagged.

**Refinement rationale [revises Decision C].** A probe (good pkg + a pkg with
`var _ int = "x"`, loaded `./...`) showed the broken pkg carries `Errors=1`
**but** `Types != nil` and its `Gadget` is fully in scope. So aborting on *any*
`pkg.Errors` would hard-fail a `./...` sweep over one unrelated broken package
— a robustness regression vs. the old silent behaviour, and in tension with
"code must build". The fix: abort only when nothing usable loaded
(`ListError` / nil types / empty), warn-and-continue on error-tolerant partial
loads. A nonexistent path is a `ListError` (empty scope, non-nil placeholder
`Types`) → still aborts.

**Emit channel:** `NewScanCtx` is not a builder — call `opts.OnDiagnostic`
directly (nil-guarded), constructing via `grammar.Warnf`. Package-error
positions come from `pkg.Errors[i].Pos` (a `"file:line:col"` string;
include verbatim in the message rather than re-parsing).

**Test:** load a nonexistent package path (cf. `TestNewScanCtx_InvalidPackage`)
and assert a `CodeDegradedLoad` diagnostic fires via `OnDiagnostic` — no broken
fixture needed. Optional second witness: a fixture in the **separate** `fixtures`
module (`github.com/go-openapi/codescan/fixtures`) with a deliberate undefined
symbol to populate `pkg.Errors` while types partly load; verify it doesn't trip
the fixtures-module CI before adding.

---

## §8.1 — panic recovery with a located diagnostic  *(Phase B)*

**Granularity [Decision A]:**
- *A1 (coarse):* one `recover()` in `api.go Run` around `builder.Build()`.
  ~10 lines; converts a panic into a clean error instead of a raw stack, but
  has no per-decl location and cannot continue.
- *A2 (fine, recommended):* a per-decl `recover()` in `spec.Builder`'s build
  loops. A guard helper

  ```go
  func (s *Builder) guardDecl(d *scanner.EntityDecl, run func() error) error
  ```

  wraps each per-decl unit — `buildDiscoveredSchema(sd)`, per-model,
  per-operation, per-route, per-parameter, per-response (~6 call sites). On
  panic it emits `CodeInternalPanic` with `PosOf(d.Ident.NamePos)` + FQN +
  the recovered value, **skips that decl**, and continues. Yields a
  partial-but-useful spec plus a precise location. Matches the catalog intent
  ("keep the panic non-fatal where a single declaration can be skipped").
- *Recommended:* **A2 for the per-decl loops + a coarse A1 backstop in `Run`**
  for panics outside any decl loop (`reduceDefinitionNames`, the package walk).

  `spec.Builder` holds `ctx` but is not a `common.Builder`; give it a small
  `recordDiagnostic` that calls `ctx.OnDiagnostic()` directly (mirrors
  `common.Builder.RecordDiagnostic`).

**Posture after a recovered panic [Decision B]:** emit an **Error**-severity
diagnostic, skip the decl, continue, and return the partial spec with `nil`
error. The Error-severity diagnostic is the loud signal; no new Options knob.
(Alternative: abort `Run` with an error. Recommend continue — a single bad
decl shouldn't sink the whole spec, and the diagnostic is unmissable.)

**Test [seam]:** a reliable panic can't be provoked from valid source. Use a
test-only injection seam: an unexported `var panicForDecl func(*scanner.EntityDecl)`
in the `spec` package (nil in production, called by `guardDecl`), with an
exported setter in `export_test.go`. The test forces a panic on one named decl
and asserts (a) a `CodeInternalPanic` diagnostic carrying that decl's
`file:line`, and (b) the build continues — sibling decls still emit. No
production API widening (per the repo rule: use `export_test.go`).

---

## Phasing & guardrails

- **Phase A (§8.2):** smaller, isolated, no test-seam complexity → its own commit.
- **Phase B (§8.1):** panic guard + test seam → its own commit.
- Validate each (`go build ./...` + full `go test ./...` + `golangci-lint run
  --new-from-rev master`) before merge. **Do not merge into `fix/backlog-lot1`
  without validation** (per Fred).

## Explicitly out of scope

- Grammar-parser `parse.*` diagnostics reaching `OnDiagnostic`: the parser's
  `WithDiagnosticSink` appears unwired in `common.ParseBlocks` (Block-accumulated
  parse diagnostics may not stream to the callback). Orthogonal to §8 —
  noted, not touched here.
- No new `Options` toggles unless a hard-fail posture is wanted.

## Decisions (settled with Fred 2026-06-17)

- **A — per-decl recovery + top-level backstop.** The per-decl guard is
  required so the diagnostic captures *which* scanned declaration (file:line)
  triggered the panic; the coarse `Run` backstop covers panics outside any
  decl loop (`reduceDefinitionNames`, package walk).
- **B — abort `Run` with an error.** A recovered panic does NOT skip-and-continue.
  The per-decl guard emits the located `scan.internal-panic` diagnostic, then
  returns an error that propagates up and aborts `Run` (Run surfaces a clean,
  located error instead of a raw stack). No partial spec.
- **C — §8.2 is Error/abort.** A degraded load emits an Error-severity
  `scan.degraded-load` diagnostic and `NewScanCtx` returns an error (aborts `Run`).

### Implementation consequence of B/C

Both halves now *fail loud AND hard*: the diagnostic is the located explanation,
the returned error is the abort. Implementation must scope the abort conditions
narrowly enough not to regress healthy scans — the full suite is the empirical
check (validate before merge). In particular §8.2 must distinguish a genuinely
degraded scanned package (nil `Types`/`TypesInfo`, the #2874 signature; zero
packages) from incidental transitive-dep noise; confirm against the existing
fixture corpus that the abort conditions don't trip healthy loads.
