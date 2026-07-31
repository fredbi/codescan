# TextMarshaler toolchain-independence — fix, maintainability swap, and CI guard

**Origin:** external PR `fix/textmarshaler-toolchain-independent`
(worktree `.worktrees/contributions/fix/textmarshaler-toolchain-independent`, commit `e4307a9`).
**Status:**
- ✅ **PR merge** — contributor's fix merged as-is (PR #60, commit `e4307a9`).
- ✅ **Maintainability swap** — `mustIfaceFromSource` + nil-guard + `TestMustIfaceFromSource`
  landed on `chore/toolchain-independence-enhancement` (commit `651294b`).
- ✅ **CI guard** — three-job pipeline (`.github/workflows/toolchain-independence.yml`) +
  `TestToolchainIndependence_FullScan` landed (commit `d420f3f`). **CI green** — job C's self-check
  passed on a real runner (build toolchain genuinely absent, scan still produced strings), so the
  guard discriminates for real, not a rubber stamp.
- ⬜ **Merge to master** — awaiting Fred's review before merge (per the wait-for-review rule).

---

## 1. The bug (confirmed & reproduced)

`resolvers.IsTextMarshaler` is the **only** function in the codebase that resolves a stdlib type
out-of-band. Pre-fix it did:

```go
encoding, err := importer.Default().Import("encoding")   // ← independent of the loaded graph
...
return types.Implements(tpe, asInterface)
```

`importer.Default()` is a **source importer** that reads `$GOROOT/src/encoding`, where GOROOT is
`envOr("GOROOT", runtime.GOROOT())` — the value **baked at build time**. When a codescan-embedding
binary (e.g. go-swagger) is built against one GOROOT and run where that path is absent or a
different toolchain is active (`GOTOOLCHAIN` switch, "built on CI, run elsewhere"), the import
fails, the error is **silently swallowed**, `IsTextMarshaler` returns false, and every
TextMarshaler-implementing type renders as `{type: object}` instead of `{type: string}`.

Reproduced: a compiled binary run with `GOROOT=/nonexistent` prints
`can't find import: "encoding": cannot find package "encoding" in .../nonexistent/src/encoding`.

### Why this does NOT generalise to `time.Time` / `error` / `json.RawMessage` / `any`

Two **different** resolution mechanisms — this is the key correction to the PR's framing:

| Recognizer | Mechanism | Toolchain-dependent? |
|---|---|---|
| `time.Time`, `error`, `json.RawMessage`, `any` | **Identity** on the `*types.TypeName` the main `packages.Load` **already resolved** (`IsStdTime`: `o.Pkg().Name()=="time" && o.Name()=="Time"`) | **No** — built by the *runtime* environment's toolchain, whatever it is |
| `TextMarshaler` | **Structural** (`types.Implements`) against an interface fetched via a **fresh `go/importer`** | **Yes** — the lone outlier |

The main load mode is `NeedTypes | NeedDeps | NeedImports | NeedSyntax | NeedTypesInfo`
(`scan_context.go:29`), so the full stdlib type graph of the *scanned* code is always built by the
runtime toolchain. Identity recognizers ride on that graph and are immune. Only `IsTextMarshaler`
re-imported the stdlib independently. Blast radius: one function.

---

## 2. The fix, and the maintainability swap

The contributor's fix builds `encoding.TextMarshaler` **structurally** with `go/types`
(`NewInterfaceType`/`NewFunc`/`NewSignatureType`/`NewVar`) — correct and robust (`types.Implements`
is purely structural; `[]byte` and `error` are Universe types), but write-only to read/maintain.

**Swap (follow-up commit):** resolve the interface by **type-checking a one-line source snippet
with a nil importer** instead. Same zero-runtime-resolution guarantee (the snippet imports nothing,
so the type-checker never consults an importer), but reads like Go and extends to
`json.Marshaler` / `fmt.Stringer` / `sql.Scanner` by editing a string:

```go
var textMarshalerIface = mustIfaceFromSource( //nolint:gochecknoglobals // immutable, built once
	`package p; type T interface{ MarshalText() (text []byte, err error) }`)

func mustIfaceFromSource(src string) *types.Interface {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "iface.go", src, 0)
	if err != nil { panic(fmt.Errorf("parsing synthetic interface source %q: %w: %w", src, err, ErrInternal)) }
	pkg, err := new(types.Config).Check("p", fset, []*ast.File{f}, nil) // nil importer: snippet imports nothing
	if err != nil { panic(fmt.Errorf("type-checking synthetic interface source %q: %w: %w", src, err, ErrInternal)) }
	iface, ok := pkg.Scope().Lookup("T").Type().Underlying().(*types.Interface)
	if !ok { panic(fmt.Errorf("synthetic interface source %q did not declare an interface T: %w", src, ErrInternal)) }
	return iface
}
```

`panic`-on-malformed is correct: `src` is a compile-time constant, so failure is a programming
error, never a runtime condition. Panics wrap `ErrInternal`, consistent with the sibling `Must*`
assertions in the same file. Verified: build ✅, `go vet` ✅, `golangci-lint` 0 issues ✅, the
contributor's `TestIsTextMarshaler` (all subtests, incl. value→pointer method promotion) passes
**unchanged** against the new implementation ✅.

**Follow-up should also add:** a unit test for `mustIfaceFromSource` itself (malformed `src`
panics wrapping `ErrInternal`; non-interface `T` panics), to pin the helper's contract *before* it
gets generalised to other stdlib interfaces.

### Why NOT the "inject a synthetic package" alternative

Considered and rejected. Injecting a package that imports `encoding`/`time` and resolving from the
loaded graph:
1. **Module-visibility wall** — an anchor package lives in *codescan's* module; when codescan runs
   embedded as a compiled binary scanning a third-party module, `packages.Load` runs in the
   *scanned* module's context, our internal package isn't in their `go.mod`, and codescan's source
   isn't on disk at runtime.
2. **Deeper:** *any* approach that resolves the **real** `encoding` at runtime — `go/importer`, a
   fresh `packages.Load`, or an injected anchor — re-runs stdlib resolution and inherits the exact
   GOROOT/toolchain fragility we're escaping. The synthetic approach is robust *because it resolves
   nothing.*

---

## 3. CI guard — three jobs

Bespoke pipeline; **not** part of the shared/reusable golden-fixture workflow. Codescan makes **no
library-wide go-less pledge** — the go-less property is **local to job (B)** only.

### Empirical ground truth (real `go` on PATH)

| Env condition | `go/importer` (old code) | `packages.Load` (full workflow) |
|---|---|---|
| `GOROOT=/nonexistent` | FAIL | **FAIL** — `go: cannot find GOROOT directory` |
| `GOROOT` **unset**, baked path present | OK | OK |
| `GOROOT` **unset**, baked build-path **absent** | **FAIL** (falls back to absent `runtime.GOROOT()`) | OK (`go` self-locates from its own binary) |

The bug and the workflow diverge **only** in the last row: `GOROOT` unset **and** build-time GOROOT
absent at runtime. Two traps fall out of this:
- **Never set `GOROOT` to a bogus path** — it kills `go` itself, so the workflow fails for the
  wrong reason (false red). Job (C) must **unset** GOROOT.
- **`setup-go` exports `GOROOT`** to a valid path → old importer finds `encoding` there → **bug
  masked** (false green). Job (C) must strip it (`env -u GOROOT`) right before running the binary.

### Job A — build

`setup-go` (stable), normal env. Produce the pre-built test binaries as artifacts; run no tests here.
- `go test -c -o resolvers.test ./internal/builders/resolvers/`      → for job (B)
- `go test -c -o scan_toolchain.test ./internal/integration/...`     → for job (C) (or a dedicated pkg)

### Job B — pinpoint, hermetic (**needs zero Go**)

Claim locked: *the TextMarshaler interface resolution depends on nothing at runtime.*
Do **not** run `setup-go` — its absence is the test. Run the prebuilt binary under a scrubbed env:

```
env -i HOME=$HOME ./resolvers.test -test.run TestIsTextMarshaler -test.v
```

A `go test -c` binary is self-contained; the contributor's test builds its subject types in-process
(no `packages.Load`), and the old bug's trigger is a pure env/baked-path condition, not a
"is `go` on PATH" one. **Proven discriminator:** this binary *passes* post-fix and *fails* pre-fix
(spliced old `go/importer` impl) in the identical `env -i GOROOT=/nonexistent PATH=/nonexistent`
environment — so it locks the invariant, it is not a tautology.

### Job C — wide, full-workflow (**needs a different, older Go at runtime**)

Claim: *a complete scan survives build≠run toolchain divergence and still emits correct output.*

```
setup-go: oldstable          # major-different from job A's stable; working `go` for packages.Load
download scan_toolchain.test; chmod +x
env -u GOROOT GOTOOLCHAIN=local ./scan_toolchain.test -test.run TestTextMarshaler_FullScan -test.v
```

- `GOTOOLCHAIN=local` — stops `go list` (inside `packages.Load`) auto-downloading a toolchain to
  match the fixture's `go` directive (needs network; could heal the env). Pair with a fixture whose
  `go.mod` targets a low version (≤ oldstable) so no switch is attempted.
- The stable/oldstable delta is a genuine build≠run smoke test (catches *other* latent build-time
  toolchain assumptions too), and makes the baked build-GOROOT (a version-specific toolcache dir)
  likely absent on the run runner.

**Two hardening touches** (so it can never silently pass on buggy code):
- **Self-check the divergence.** The test prints the binary's baked `runtime.GOROOT()` and asserts
  that path does **not** exist at runtime (and differs from `go env GOROOT`). If they coincide,
  `t.Skip` loudly — a skip beats a false pass. For absolute determinism regardless of runner-image
  caching, run job (C) in a container holding only oldstable, or build job (A) against a Go unpacked
  at a bespoke prefix the run runner provably lacks.
- **Assert the symptom, not the boolean.** Scan a real fixture with a `MarshalText() ([]byte, error)`
  type and golden-compare that the field is `type: string`, exercising
  `codescan.Run` → `packages.Load` → `schema.go:350` (`IsTextMarshaler`) end to end.

### Fixture to add

`fixtures/bugs/textmarshaler-toolchain/` — a type with a value-receiver `MarshalText() ([]byte, error)`
used as a schema field (and, for `IsJSONMapKey` coverage, as a map key), `go.mod` pinned low, golden
asserting `type: string`.

---

## 4. Merge / rebase order

1. ⬜ Merge the contributor's PR (`e4307a9`) as-is.
2. ⬜ Rebase the `mustIfaceFromSource` swap on top as a separate commit
   (`refactor: resolve TextMarshaler via source-checked interface`) + the `mustIfaceFromSource` unit test.
3. ⬜ Add the CI pipeline (jobs A/B/C) + fixture, in its own commit.

Open question for step 3: settle whether job (C)'s determinism rides on the cross-major version
delta alone (lighter, small masking risk) or a container/bespoke-prefix build (heavier, fully
deterministic). Leaning container-or-bespoke-prefix for the self-check to be a hard assert rather
than a skip.
