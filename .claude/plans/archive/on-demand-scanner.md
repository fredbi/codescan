> [!NOTE]
> Last revision: 2026-08-09 (derisking target closed — see the box below)

> [!TIP]
> **✅ CLOSED 2026-08-09 — the derisking target is done.** Base camp
> `.worktrees/fix/derisk-precompiled-dependencies`, commit `fc3d279` on `master` `0273dcf`.
> Unpushed, unreviewed.
>
> | | what | outcome |
> |---|---|---|
> | 1 | **The name-based field bridge** — the prerequisite for the option ever defaulting on | ✅ done, together with the on-demand read-back that turned out to be its other half — see [below](#-done--the-name-based-field-bridge-2026-08-09-fc3d279) |
> | 2 | The two costs that kept it opt-in | one of them is gone: an unannotated dependency's model no longer collapses. **Cold-cache inversion remains, and is enough** — it stays opt-in and experimental |
> | 3 | The A/B scoreboard as the acceptance gate | ✅ **empty**, all three configurations, whole fixture corpus (306 subtests). Six rows → zero |
>
> What is left on this stream is nothing. `CompiledDependencies` now emits the same document as an
> ordinary scan; what it trades is cache behaviour, which is a choice rather than a defect.

> [!IMPORTANT]
> **MERGED 2026-08-08 — PR #90, merge commit `7f6654a`. This stream is closed.**
> Re-squashed from 39 commits into 13 logical ones, rebased straight onto `master` (`1a5a70b`) and
> merged with the history preserved. It is not a descendant of `feat/source-loader` or `bench-grammar`.
>
> **Cleaned up 2026-08-08.** Deleted along with their worktrees: `on-demand-scanner` (merged),
> `on-demand-scanner-preSquash` (`6ee1115`), `feat/source-loader` (`47c5f82`), `bench-grammar`
> (`885f3a2`). **Every commit hash quoted below other than the 13 above is now unreachable** — it
> lived only on the pre-squash branch. The tips are recorded here so `git checkout <hash>` can still
> revive them from the reflog until gc collects them (~90 days, i.e. into November 2026); after that
> they are gone. `bench-grammar` also exists as `upstream/bench-grammar` on the fork, deliberately
> untouched.
>
> | # | commit |
> |---|--------|
> | 1 | `chore(deps): update the fixture corpus dependencies` |
> | 2 | `refact(parsers): read the annotation line instead of matching it` |
> | 3 | `refact(scanner): answer a declaration's name and position from its type` |
> | 4 | `refact(builders): resolve a written type without the expression map` |
> | 5 | `refact(scanner): make a declaration's absent source impossible to ignore` |
> | 6 | `feat(scanner): load and type-check packages without a Go toolchain` |
> | 7 | `feat(scanner): take dependency types from compiled export data` |
> | 8 | `doc(packages): bring the loader's comments to the prose standard` |
> | 9 | `fix(packages): read back a dependency's annotated source` |
> | 10 | `fix(builders): don't fail a scan over a declaration that cannot be read` |
> | 11 | `doc(scanner): state that FS is the whole world a scan can read` |
> | 12 | `test(integration): measure what a scan loads, holds and costs` |
> | 13 | `fix(packages): read paths the way the host writes them` |
>
> Verified: tree byte-identical to the pre-squash tree through commit 12; build, vet, test green;
> `golangci-lint run --new-from-rev master` reports 0 issues.
>
> Commit 13 answers the Windows CI run on PR #90. The class of bug to watch for in `internal/packages`:
> it mixes IMPORT paths (slash, unrooted) with HOST paths (OS separator, drive letter), and every path
> decision has to say which it is. GOPATH split on `":;"` cut `C:\Users\x\go` to `C`, so the module cache
> was unreachable and every cached dependency was synthesized — which is how `strfmt.DateTime` lost its
> format while the scan still reported success.

# On-demand scanner (Stream 8 · `V-scanner`)

## Summary

The scanner fuses three separable jobs into a single eager `packages.Load`: **finding annotations**, **resolving
types**, and **reading declarations**. Because they are fused, all three run at maximum width — we hold the AST and
type-checker output of the entire transitive closure in order to read a few hundred annotation lines.

Measured on two real targets: **95–99% of the loaded package graph is never needed by anything** (dockerctl: 4 of 336;
go-swagger: 19 of 440), and the scan is **94% of total wall clock** with **~690 MB peak RSS** on a 502-file project.

The proposal was to split the three jobs apart: discovery stays exhaustive but becomes cheap (a byte prefilter,
0.05 s, nothing retained); type materialisation becomes lazy for free — `go/types`' export-data importer already
pulls only what is referenced; syntax materialisation becomes lazy by us (~2 ms per package, no second `go list`).

**That is not how it went.** The objective was met from the other end, and the deliverable was retired by the
harness built to judge it. Read the audit below before the Trajectory, which is kept for its reasoning rather
than as a worklist.

**Overtaken in part by `feat/source-loader` (2026-08-05).** That branch owns loading and moved four times in a day.
Its `83972ff` now decides the question **per dependency**, using a byte scan for the literal `swagger:`: a dependency
that carries annotations is loaded from source in the ordinary way, one that does not is taken from export data with
no syntax at all. Petstore 79 ms against 748 ms all-from-source, byte-identical spec.

That is this plan's byte prefilter, landed in the loader, and it settles two of the things this document was written
to propose — the lazy parse (there is no longer an eager per-dependency parse to defer) and exhaustive discovery
(preserved by construction, since anything annotated is fully loaded). What survives:

1. **The declaration contract**, unchanged in importance — see below, it is now what the loader is designing around.
2. **The regexp removal**, which was always maintainability rather than performance.
3. **A new question the fix opens**: the per-dependency choice is forced by five builder sites reading
   `types.Info.Types`, not by `go/types`. Take those off it and an annotated dependency need not pay a full
   type-check either (action 5).

Landed on `feat/source-loader`. The perf/OOM cluster of Stream 9 tickets is closed — by that branch's route, not
this one. `V-builder`'s prerequisite is met and it is parked separately ([`pull-based-builder.md`](pull-based-builder.md)).

## Where this actually stands (audited 2026-08-07)

**Complete for the objective, not for the plan** — and those came apart early. Two steps of seven landed,
plus one salvaged sub-item; one was pre-empted; the rest either died to evidence or never started.

| step | state |
|---|---|
| 1 Harness | 1.1 ✅ · 1.2 ✅ (corpus × config A/B exists, two-tier) · 1.3 ⛔ decided, see below |
| 2 Declaration contract | ✅ done |
| 3 Split discovery from loading | ⛔ pre-empted by `83972ff`, except the regexp removal ✅ |
| 4 Lazy types | 4.1 ✅ (landed on the loader branch, not here) · 4.2 ⛔ mis-specified · 4.3 / 4.4 ⛔ dead with laziness |
| 5 On-demand materialisation | 5.2 ✅ · 5.5 ✅ · 5.6 ⛔ decided opt-in · **and it exists after all** — not as a lazy load, but as the per-declaration read-back of 2026-08-09 |
| 6 `SchemaCache` / 7 `V-builder` | moved to [`pull-based-builder.md`](pull-based-builder.md) |

**The named deliverable was retired by its own control loop, not abandoned.** Once `83972ff` made the
source/export-data choice per dependency, there was no longer an eager per-dependency parse to defer, and the
harness put the residual prize at ~0.03 s / 11 MB on dockerctl and ~nothing on go-swagger. A plan whose centre
is deleted by its own measurement is a good outcome; calling it *complete* would be flattering it.

**Worklist — ✅ CLOSED 2026-08-08. All three were one question: why `CompiledDependencies` cannot default on.**
Nothing on it is a failure; what replaced it is the parked field bridge, which is where the derisking resumes.

| | | |
|---|---|---|
| 5.2 | ✅ **DONE** (`84e0b58`) — the go/packages route reads annotated dependencies back after the load | closed **4** |
| [Q37](../quirks-open.md) | ✅ **DONE** (`46c4da3`) — the strict lookup sites degrade and warn instead of failing | closed **2** |
| 4.2 | ⛔ mis-specified — that route already routes per dependency | — |

`abExpected()` in `loader_agreement_test.go` is the scoreboard: **twelve rows → six**, and **none of the six is a
failure any more**. What is left is two stdlib types rendering thinner (`io.Writer`, `reflect.Type`) and the
family-3 pair. **`CompiledDependencies` now has no blocker left** — see 5.6.

**1.3 — ⛔ settled 2026-08-07 (Fred): no benchmarks run in CI.** The corpus A/B is a manual gate, run once per
minor release (~twice a year) to verify. Cost is measured deliberately, not continuously; a benchmark in CI on
shared runners measures the runner. So `CODESCAN_BENCH=1` and `CODESCAN_AB_CORPUS=1` stay operator tools, and the
release checklist is where they belong.

**The performance objective is met, and largely not by this plan.** The loader branch got there by a
different route — per-package cuts, then a per-dependency source/export-data choice.

**The declaration contract is met** (2026-08-06/07). Step 2 of 7, the one everything else was
supposed to gate on, is closed:

| | |
|---|---|
| ✅ type half always answers | `Name` / `Pos` / `PkgPath` / `IsAlias` / `WrittenRHS` off `Obj()` |
| ✅ call sites migrated | ~81 syntax accesses → 3 hard + 30 nil-tolerant |
| ✅ syntax half unexported | `comments` / `ident` / `spec` / `file` / `pkg`, written as a unit |
| ✅ absence un-ignorable | `HasSource()` gate; `(value, bool)` where a deref would have panicked |

The accessor set, split by whether absence was already a handled outcome:

```go
HasSource() bool                                   // the single gate
Comments() *ast.CommentGroup                       // nil-safe (35 sites unchanged in shape)
File() *ast.File                                   // nil-safe
TypeExpr() (ast.Expr, bool)                        // explicit absence
Imports() ([]*ast.ImportSpec, bool)
PkgImport(path string) (*packages.Package, bool)
EnumSourcePkg() (*packages.Package, bool)
```

No `*ast.Ident` accessor survived — the post-decl dedup is keyed on `Obj()` now, which is injective
because a package cannot declare one name twice (an *invalid* duplicate mints a distinct object, so
two decls never merge onto one key). `Type`/`Alias` stay exported, so a type-only declaration is
constructible from outside the package while a half-built one is not — which is what lets the new
test exist at all.

**The silent-peel trap is now stated rather than latent.** `schema.go` distinguishes WrittenRHS's two
refusals: a declined shape still falls back to `Underlying()`, no source refuses with
`missingSource`. Unreachable today and documented as such; `TestBuildFromDeclWithoutSource` pins it,
and it bites — reverting the guard fails it (verified independently, not only by the author).

**A Go gotcha worth remembering from this change.** `decl.File == nil` still *compiles* after `File`
becomes a method: Go permits comparing a func value to nil, and a method value is never nil, so that
guard would have silently become dead code. It was caught only because `HasSource()` replaced it.
Any future field→method move in this repo wants the same grep.

**Where it landed (2026-08-08).** The three live items were all "why `CompiledDependencies` cannot default on".
5.2 landed, 4.2 was mis-specified, Q37 landed — and then the product decision went **against** defaulting it
on: the option **stays opt-in and experimental**. Two reasons then; **one now** — the unannotated-dependency
collapse closed on 2026-08-09, and what is left is that the cost inverts on a cold cache, which is what CI
runs on. See 5.6.

## Context

### The finding that reframes the stream

The original framing (`ramblings/index-builder-statefulness.md`) was **"syntax now, types later"**. Measurement says
that is not the axis. The axis is **discovery vs. loading**, which today are the same operation.

Three jobs, three natural costs — currently paid at one price:

| job | must be | cost when separated | cost today |
|---|---|---|---|
| find every annotation in the closure | **exhaustive** | 0.05 s, ~0 retained | fused |
| resolve types reachable from annotated decls | **lazy** | free (the importer is already lazy) | fused |
| read a declaration's source (comments, shape) | **on demand** | ~2 ms / package | fused |
| | | | **0.95 s / 616 MB** |

`walkImports` + `pkgLoadMode` (`NeedDeps|NeedTypes|NeedSyntax|NeedTypesInfo`) is what fuses them: it parses and
type-checks every transitive dependency, stdlib included, and retains all of it in `TypeIndex.AllPackages`.

Full measurements in [Appendix A](#appendix-a--evidence).

### What `feat/source-loader` already built, and where this stream joins it

That branch owns package loading now: `internal/packages.Loader`, one seam with two strategies
(`StrategyGoPackages` / `StrategyToolchainFree`) and a shared option vocabulary. This stream builds **on** it, not
beside it — the demand machinery belongs on that Loader, not in a parallel path under `internal/scanner/`.

It has already attacked cost along a **different axis than this plan**, and the two compose rather than overlap:

| axis | what it does | status |
|---|---|---|
| **cheaper per package** | trim `types.Info` to the two maps read; `IgnoreFuncBodies` for deps; drop `ast.Object` | landed `9007329` — petstore 1.25 s → 0.62 s, 655 MB → 346 MB |
| **fewer packages** | dependency types from export data, not from their source | landed `0dc679c` as `CompiledDependencies` |
| **recover what that costs** | parse a dependency's source alongside its export data, joined by name | landed `a193bcf`, **toolchain-free route only, eagerly** |
| **stop paying for what nothing reads** | do that parse *on demand*, and on the go/packages route too | **this stream** |

So what this document called "tier 0" exists already: `compiledDepsMode = loadMode &^ packages.NeedDeps`, which is the
signal upstream reads as "dependency types may come from export data" (`usesExportData` is
`NeedExportFile != 0 || (NeedTypes != 0 && NeedDeps == 0)`). Stop proposing it; depend on it.

**The remaining half is laziness, and it is worth more than the fix that preceded it.** `a193bcf` establishes the
right mechanism — `exportedPackage` takes the types from export data and the comments from source, parse-only with no
second type-check, the two halves joined by name through `bridgeDefs` because every declaration the source names is
already an object in the export-data scope. But it does that **eagerly, for every dependency**, and measured against
this stream's closure numbers that is most of the saving handed straight back:

| parse-for-comments over the closure | wall | retained |
|---|---|---|
| eager — every dependency (335 pkgs, 2341 files) | **0.51 s** | **168 MB** |
| lazy — only what a scan reaches (4 pkgs, 102 files) | **0.02 s** | **6 MB** |

Against `CompiledDependencies`' own 0.36 s / 107 MB, the eager parse more than doubles both. 25× on wall clock and
28× on memory sit between the two rows, and nothing about the output differs — the 331 packages parsed in the first
row are never asked a question.

**And the go/packages route has not had the fix at all.** `a193bcf` touches `exportdata.go`, which serves the
toolchain-free `ExportData` blob; `CompiledDependencies` still raises its hint and still loses `strfmt`'s marks. That
route can have exactly the same treatment, because it keeps the file paths:

```
compiledDepsMode:  graph=336  GoFiles=336  Types=336  Syntax=19        0.36 s   107 MB
   strfmt:         GoFiles=9  Types=non-nil  Syntax=0
```

The dependency whose comments the option gives up is sitting there with nine readable paths and no AST. Parsing them
is 1.7 ms.

### Why discovery must stay exhaustive

Settled 2026-08-05 (Fred). An early draft proposed dropping annotated-but-unreferenced declarations found in
dependencies, on the grounds that `PruneUnusedModels` exists to mop up that noise anyway. **Withdrawn.** You cannot
decide at load time whether a model will end up referenced:

- **Discriminator subtypes are discovered backwards** — a subtype `$ref`s its base and nothing `$ref`s the subtype, so a
  `swagger:model` + `swagger:allOf` type living in a dependency is *never reached by a type reference*. Pure
  demand-outward would silently lose polymorphic families.
- **`$ref` resolution completes late**, in the spec builder, after the discovery loops.

So the *discovered set does not change*. Only what we hold in order to discover it changes. The byte prefilter reads
every file in the closure — including stdlib — precisely so backwards discovery keeps working.

### Why the annotated set is not the working set

Also settled 2026-08-05 (Fred), and quantified. The prefilter yields the **root set**; the working set is the root set
∪ the **type-reference closure** from it. Packages carrying no annotation whatsoever are pulled in purely because a
field is typed that way:

```
dockerctl:   2 annotated roots → closure  4 packages;  added: oklog/ulid, time
go-swagger: 14 annotated roots → closure 19 packages;  added: strfmt/internal/countries,
                                                              scan-repo-boundary/makeplans,
                                                              oklog/ulid/v2, x/text/currency, time
```

Type-following adds **25–36%** on top of the root set. Which is fine, and does not need a reachability analysis of our
own: type-checking the root packages against an export-data importer *discovers the closure as a side effect*, because
the importer only reads export data for packages actually referenced.

### Dependency annotations are real

`github.com/go-openapi/strfmt` carries **155 `swagger:strfmt` annotations in its own source**, read from the
dependency. Any design that reads only the main module is wrong. But those annotations are consumed *only when the type
is reached* — `swagger:strfmt` on `strfmt.Date` matters exactly when a field resolves to `strfmt.Date` — so on-demand
syntax materialisation serves them exactly. This is the load-bearing path, not an edge case.

### Convergence with the WASI stream

`wasi-build.md` is *types now, syntax never*: the standard library arrives as `gcexportdata`, carrying a complete
`*types.Package` and no AST at all. That was framed as a tension with on-demand's *syntax now, types later*.

It isn't a tension — it's the same mechanism with one knob:

| | population | syntax half | materialisation |
|---|---|---|---|
| on-demand | any package not yet read | not yet present | parse `GoFiles` (~2 ms) |
| export data | standard library under WASI | permanently absent | unavailable — answer honestly |

Both need a declaration whose **type half is always available** and whose **syntax half may be absent**. The only
difference is whether asking for it can succeed. That is one contract, not two.

Constraints the WASI side imposes, unchanged from the previous revision of this document:

- **No toolchain, no exec.** A WASI guest cannot run `go list`.
- **Memory is binding.** Type-checking stdlib from source costs 681 MB peak for the petstore under wasmtime; export
  data costs 138 MB, and 1.00 s vs 7.30 s.
- **Stubbing stdlib costs correctness** — no fields, no method sets, so `encoding.TextMarshaler` detection,
  `json.RawMessage`-as-byte-array and interface identity all break (138/143 vs export data's 141/143).
- **Dependency comments are not merely absent, they are unwanted.** stdlib carries no swagger annotations, and reading
  its godoc is what produces [Q38(b)](../quirks-open.md).

### The `EntityDecl` contract — no longer hypothetical

`CompiledDependencies` **produces the broken state today**. A dependency comes back with `Types != nil` and
`Syntax == nil`; `FindDecl` walks `pkg.Syntax`, finds nothing, returns `(nil, false)`, and the type's annotations are
silently unavailable. That is exactly the documented `strfmt` symptom, and it is the contract failing rather than
export data being lossy. There is now a live configuration to write the fixture against.

```go
type EntityDecl struct {
    Comments *ast.CommentGroup   // syntax
    Type     *types.Named        // type
    Alias    *types.Alias        // type
    Ident    *ast.Ident          // syntax
    Spec     *ast.TypeSpec       // syntax
    File     *ast.File           // syntax
    Pkg      *packages.Package   // syntax (used for TypesInfo)
    …
}
```

Five syntax-bound fields, ~81 accesses across `internal/builders/` (production code only):

| field | sites | if nil |
|---|---|---|
| `Comments` | 28 | mostly safe — the parsers tolerate nil |
| `Ident` | 21 | **panics** (`Ident.Pos()`, `Ident.Name`) |
| `Spec` | 12 | **panics** (`Spec.Type`, `Spec.Pos()`, `Spec.Assign`) |
| `Pkg` | 11 | **panics** (`Pkg.TypesInfo.Types[…]`) |
| `File` | 9 | **panics** (`File.Imports`) |

Sorting by *why* each site reaches for syntax is far more encouraging than the raw count:

1. **Positions** — the largest group (`Ident.Pos()`, `Spec.Pos()`). Needs no syntax: `Obj().Pos()` answers, and export
   data populates it.
2. **`TypesInfo.Types[decl.Spec.Type]`** — recovering *the type* from a syntax node, redundant when the type is already
   in hand. Incidental, not essential.
3. **Syntax-shape checks** — `Spec.Type.(*ast.InterfaceType)`, `Spec.Assign.IsValid()`. Both answerable from `go/types`
   (`*types.Interface`, `*types.Alias`).
4. **`File.Imports`** — one site, genuinely syntax.
5. **`Comments`** — genuinely syntax, and meaningless for stdlib.

So the essential syntax dependency is far smaller than 81 sites implies.

**Two precedents already in the file.** `Obj()` and `ObjType()` are already type-only, and panic *loudly* on an
impossible state rather than nil-dereferencing — the discipline the rest of the type wants. And `schema/schema.go:170`
already guards `decl.Spec == nil || decl.Pkg == nil || decl.Pkg.TypesInfo == nil`.

**Two outliers.** `Names()` and `DefKey()` dereference `d.Ident.Name` where `Obj().Name()` serves — the only type-half
methods reaching into the syntax half without needing to.

### The regexp question

Settled 2026-08-05 (Fred): **maintainability, not performance.** The full comment walk over 2354 files costs 0.01 s, so
there is no perf case at all. But `rxCommentPrefix` and friends are painful to maintain for what a string search does,
and the scanner is the last regexp holdout now that the lexer and grammar parser are clean.

Proof of concept, already written for the probes — this reproduces `rxCommentPrefix + "swagger:"` exactly:

```go
func isAnnotationLine(text string) bool {
    i := 0
    for i < len(text) && strings.ContainsRune(" \t/*-", rune(text[i])) {
        i++
    }
    if i < len(text) && text[i] == '|' {          // markdown table pipe
        i++
        for i < len(text) && text[i] == ' ' {
            i++
        }
    }
    return strings.HasPrefix(text[i:], "swagger:")
}
```

The line-start rule matters and is the only subtle part: a naive `strings.Contains` counted **14** annotated packages on
dockerctl where the real rule counts **2** — codescan's own godoc, scanned as a dependency, talks about `swagger:`
constantly. (Note the rule does not save codescan from itself: several of its own doc lines legitimately *begin* with
`swagger:` — the "three such lines exist in this repo's own fixtures" hazard already documented in `regexprs.go`.)

## Coordination — the branch stack

✅ **The stack is dissolved — all four branches are in `master` or deleted (2026-08-08).** Kept for the
rebase protocol below, which is the part worth reusing.

| # | Work | Branch / worktree | Outcome |
|---|---|---|---|
| 1 | Bench harness — the stream's control loop | `bench-grammar` | ✅ absorbed into 3, then 2; branch **deleted** (`885f3a2`) |
| 2 | Declaration contract + byte prefilter + regexp removal | `on-demand-scanner` | ✅ **merged**, PR #90; branch deleted |
| 3 | Loader + loader strategies | `feat/source-loader` | ✅ carried into master by 2; branch **deleted** (`47c5f82`) |
| 4 | WASI build (UX) | `wasi-build` | ✅ **merged**, PR #79 |

Successor base camp: `.worktrees/fix/derisk-precompiled-dependencies` — see the box at the top.

**Rebase protocol.** Fred leads the loader and **signals** which of its commits are worth rebasing onto; branches 1
and 2 rebase on that signal, not on every upstream commit. Rebasing an agent's in-progress branch mechanically is how
work gets lost.

**Harness absorption — ✅ done 2026-08-05.** `bench-grammar` fast-forwarded into `feat/source-loader` (now at
`885f3a2`), and `on-demand-scanner` rebased onto it — five commits replayed clean, full suite green, lint clean,
`fixtures/` still untouched. `bench-grammar` is now redundant and can be retired. Branch 4 rebases when it wants the
harness.

**Measurement note — host RSS is not the WASI memory signal.** A reported 462 → 885 MB "regression" after
`wasi-build` rebased was a false alarm: the figure was the host process's RSS, which includes the wasm runtime and the
20 MB module, not the guest's linear memory. The playground now surfaces guest linear memory directly, which is what
any future drift has to be read from. Two hypotheses died on the way and are worth not re-running — a retained
per-package reason map (~48 KB at 400 packages, four orders off), and over-admission by the loader's substring marker
rule (measured: 1 of 114 vendored packages, and it genuinely carries annotations). A native A/B across the exact
rebase window showed identical `TotalAlloc` (588 MB) and identical output.

**Verification note.** `go.work` is per-worktree with relative `use` paths, so the correct invocation is a plain
`go test ./...` (plus `./cmd/genspec-tui/...`). Running with `GOWORK=off` narrows to the root module *and* fails
`TestResolveImport_WorkspaceSibling` spuriously — the loader's workspace resolution is exactly what that test
exercises. Don't read that failure as a regression.

**Branch 4 feeds back.** WASI reports how much of its known quirk set — lost dependency comments, scan times, peak
memory — is closed by the progress of 1–3, which is the acceptance signal this stream is ultimately judged on.

**Asks to branch 3 are relayed through Fred**, who conveys or pushes back. Open asks:

1. ✅ `bridgeDefs` leaves `types.Info.Types` nil (see action 4) — relayed 2026-08-05.
2. ✅ `hack/genexportdata` lives on `wasi-build`; the loader branch needs it to witness anything export-data shaped —
   relayed 2026-08-05.

## Trajectory

1. **Harness first** [🏁][⚡]
   > None of the cost numbers below were reproducible from the repo. Nothing else in this stream is judgeable
   > without them — and cost is only half of it: laziness has to be proven output-neutral, not just cheaper.

   1. ⚠️ `hack/scanbench` — `phases` / `modes` / `discover` / `closure`, wall clock + retained heap, both the raw
      LoadMode ladder and the `internal/packages` strategies. **It never landed: there is no `hack/scanbench`
      in `master` and no reference to one** (checked 2026-08-08). It was a probe on a now-deleted branch, so
      every figure below sourced to it is unreproducible as written. What *did* land is
      `internal/integration/bench_*_test.go`, gated on `CODESCAN_BENCH=1` with `CODESCAN_BENCH_DIR` /
      `CODESCAN_BENCH_GOSWAGGER_DIR` (and `CODESCAN_BENCH_EXPORTDATA_*` for the blob axes) — read the axes
      from there, not from this bullet
   2. ✅ Spec A/B across loader configurations — `internal/integration/loader_agreement_test.go`, whole-document
      comparison over the corpus × the dependency-type configurations, with a divergence table that fails when
      stale. Two tiers: curated targets by default, `CODESCAN_AB_CORPUS=1` for every fixture bundle
   3. ⛔ ~~Wire the corpus into CI~~ — no benchmarks in CI (Fred, 2026-08-07). Manual gate, once per minor
      release; on shared runners a benchmark measures the runner

2. **Settle the declaration contract**
   > Re-audited 2026-08-05, and its centre of gravity moved. `bridgeDefs` populates `TypesInfo.Defs`, and `FindDecl`
   > needs exactly `pkg.Syntax` + `Defs` — so a dependency served from export data *with source attached* already
   > satisfies it. The contract only bites today in the `ExportOnly` case. Laziness changes that: every dependency
   > then starts unfilled, so the contract stops being "tolerate an absent half" and becomes "**be the trigger that
   > materialises it**". Still first, and now load-bearing rather than hygienic.

   1. ✅ Model decided — **two states, not three**: absence is permanent, never deferred. The third
      (`materialisable`) died with the lazy parse; see ⛔ Retired
   2. ✅ Type half always available: `Obj`, `ObjType`, `Name`, `Pos`, `PkgPath`, `IsAlias`, `WrittenRHS`
   3. ✅ Syntax half unexported and gated on `HasSource()` — enforced by the compiler, not by discipline
   4. ✅ Builder call sites migrated — ~81 syntax accesses down to 3 hard + 30 nil-tolerant

3. **Split discovery from loading** — ⛔ **largely pre-empted by `83972ff`**
   > The byte prefilter now exists in the loader as `scanForAnnotations` / `carriesAnnotations`: same needle, same
   > over-admit-never-under-admit property, memoized per package. It decides *how to load* a dependency rather than
   > *which files discovery walks* — but since any dependency carrying annotations is now loaded from source in
   > full, discovery walking `pkg.Syntax` still sees everything, and decoupling it no longer unlocks what it was
   > meant to unlock. Exhaustive discovery is preserved by construction instead. What survives here is the regexp
   > removal, which was always maintainability rather than performance.

   1. ⛔ ~~`go list` only for the closure~~ — never built; the loader decides per dependency instead
   2. ⛔ ~~Byte prefilter over every file in the closure~~ — exists in the loader as
      `scanForAnnotations` / `carriesAnnotations`, deciding *how to load* rather than *what to walk*
   3. ⛔ ~~Parse + classify the survivors~~ — same reason; discovery still walks `pkg.Syntax` and sees everything
   4. ✅ Regexp-free classification [📚] — `matchers.go` retired in favour of the line scan; the retired
      expressions survive **in the test** as the specification (~294 000 generated lines + a fuzz target).
      `rxRoute`/`rxOperation` kept deliberately — four things in sequence, read back by submatch index

4. **Lazy types**
   1. ✅ Drop `NeedDeps` from the load mode — landed on `feat/source-loader` as `CompiledDependencies`
   2. ⛔ ~~Give the toolchain-free strategy the same narrowing~~ — **mis-specified**, settled 2026-08-08: that
      route already routes per dependency (`loader.go` `importer.Import`), and without a compiler there is no
      export data to narrow *to*. The blob **is** the narrowing. See action 3
   3. ⛔ ~~Confirm the incremental closure equals the measured one~~ — there is no incremental loader
   4. ⛔ ~~Decide what a partial closure means for `detectDegradedLoad`~~ — same reason; the load stayed whole-graph

5. **On-demand syntax materialisation** — ⛔ **the stream's named deliverable, retired unbuilt**
   > `83972ff` left no eager per-dependency parse to defer, and the harness put what remains at ~0.03 s / 11 MB.
   > The decl index (5.3 / 5.4) moved to [`pull-based-builder.md`](pull-based-builder.md) as an open question
   > rather than a plan — its case was this stream's performance case, and that case is gone.

   1. ⛔ ~~Make `exportedPackage`'s parse lazy~~ — nothing eager left to defer
   2. ✅ Give the go/packages `CompiledDependencies` route the same attach, from the `GoFiles` `go list` returns
      — `internal/packages/attach.go`, in `master`. Closed four A/B rows; the two that stayed open produced the
      better finding (action 2)
   3. ✅ Distinguish *not yet parsed* from *no source* — `ExportOnly` already names the second case, per dependency
   4. ⛔ **5.6 — `CompiledDependencies` stays opt-in and experimental.** Decided 2026-08-08, **reaffirmed
      2026-08-09 on narrower grounds**. Q37 and 5.2 removed the failures and the lost annotations; the
      unannotated-dependency collapse closed too (`fc3d279`), so the option now emits the same document as an
      ordinary scan and **no fidelity argument is left**. What keeps it opt-in is cost: it inverts on a cold
      cache (12.7–13.4× slower, 229 MB of build cache — and CI is cold by definition), and it needs the closure
      to **build**, not merely type-check. Documented on the option and in
      `internal/scanner/README.md#compiled-dependencies`

## Actions

### The worklist — ✅ closed 2026-08-08

One question wearing three numbers: **why `CompiledDependencies` cannot default on.** Answered: two of the
three landed, the third was mis-specified, and the product decision went against defaulting it on (5.6). Kept
in full because the reasoning is what the derisking effort resumes from.

1. ✅ **[Q37](../quirks-open.md) — DONE** (`46c4da3`, 2026-08-08). The strict sites now ask
   `Ctx.SourcelessPackage()` before failing: a package the load deliberately did not read means the type is
   complete and only its prose is missing, so it renders from what it is and raises a `scan.sourceless-type`
   **Warning**; a broken graph still errors. `time.Duration` in a response header comes out `integer/int64` —
   byte-identical to the full-source answer, which is the evidence the fallback is the right one. An unreadable
   interface renders as the open schema.

   **Fred's ruling (2026-08-08), recorded because it settles a recurring temptation:** do *not* widen the
   identity recognizers to cover the standard library. A recognizer asserts a wire form for every use of a type,
   and these are precisely the types where none exists to assert. The list may grow opportunistically, but it is
   not the fix. What is lost is a stdlib doc comment the author usually did not want in their API — the
   `opaque-streams` baseline literally publishes `io.Writer`'s method godoc as a description — and
   `swagger:description` is a better answer than any guess.

   **The asymmetry is closed too** (`05b6051`, Fred's call). A lost *whole definition* now warns as well, and
   the rule that avoids the false positives is narrow: only `resolveRefOr`'s **no-fallback arm** asks. That is
   the arm where the difference is a definition existing or not; where `orElse` renders the type structurally
   the definition survives and the miss is the everyday "not a model, inline it" outcome that fires on most
   fields of most scans.

   Measured over twelve fixture bundles: **two new warnings, both real** — `makeplans.Booking`, and
   `reflect.Value`, a stdlib struct whose definition had been disappearing without a word. Zero false
   positives. The benign `time.Time` lookup stays silent because it comes from `classifierTextMarshal`, a
   *speculative* lookup asking "did the author annotate this?" rather than one the rendering depends on — and
   under the marker rule a sourceless package provably carries no annotation to find.

   Original entry: **family 1 is a hard error under export data, not a degradation.**
   `builders/parameters/parameters.go:365` and `builders/responses/responses.go:405` return
   `unable to find package and source file for: <type>`, and `loader_agreement_test.go:206` pins it as a *known*
   divergence. Reproducible today on three targets:

   | target | asks for | via | message |
   |---|---|---|---|
   | `./bugs/2248/...` | `time.Duration` | `builders/responses` | `unable to find package and source file for:` |
   | `./enhancements/opaque-streams/...` | `io.Writer` | `builders/schema` | `can't find source file for type:` |
   | `./goparsing/go123/...` | `reflect.Type` | `builders/schema` | same |

   **The real defect is that the declaration-lookup seam is inconsistently strict**, and it splits by call site
   rather than by package. Soft sites (`resolveRefOr` with an `orElse`) fall back and now raise the deferred
   `scan.sourceless-lookup` notice — `export_only_test.go` pins a `time.Duration` *field* doing exactly that,
   cleanly. Strict sites (`resolveRefOrErr`, `buildNamedField` in parameters/responses) kill the whole scan for
   the same type in a different position.

   That strictness was written when "no declaration" meant "something is broken". It now also means "this
   dependency came from export data", which is a supported configuration — so one unresolvable field takes down
   a 200-definition document. Note where these land: the 2026-08-02 ordering fix covered the three *field* sites
   in parameters/responses and Q38 widened the canonical recognizer set, so this is not a regression — it is the
   parked structural half becoming reachable.

   **Recommended resolution:** soften the strict sites to fall-back-plus-notice, **gated on the declaring package
   being known export-only** (`ScanCtx.exportOnly`, already populated). A genuinely broken load keeps its error;
   a deliberately syntax-less one degrades and says so. The types involved are precisely the ones we have ruled
   must *not* get a recognizer (`time.Duration` → that is what `strfmt.Duration` is for; `io.Writer` → excluded
   from the opaque-stream set on purpose), so reporting is the whole of the answer available.

2. ✅ **5.2 — DONE** (`84e0b58`, 2026-08-07). `internal/packages/attach.go`: after the wholesale export-data load,
   walk the graph and hand source back to every dependency whose files carry the marker — parse, then join to the
   export-data scope by name. Verified by sabotage: disabling the one call fails the unit test, both integration
   tests and exactly the four A/B targets that closed.

   **The finding that matters more than the fix.** `abExpected()` went 12 → 8, not 12 → 6. The two `bookings` /
   `spec` rows did **not** close, and the classification that predicted they would was wrong. `Booking` lives in
   `scan-repo-boundary/makeplans`, which carries **no annotation anywhere in its source** — verified, not assumed —
   so the marker scan correctly passes over it and the definition still collapses to a bare `x-go-name`. That is
   not a routing gap: it is what the export-data trade costs, and **both strategies now pay it identically**, which
   is the alignment. Closing it would mean reading a dependency because a scan *reaches* it — a different rule, and
   the retired materialisation idea wearing a new hat.

   **A second-order effect worth knowing.** The go/packages route now records `ExportOnly` per skipped dependency,
   so `reportSourcelessLookup` can finally speak on this route. Volume measured on four corpora: **1 notice each**
   (5 for `go123`, which is the stdlib-types fixture and errors anyway). One of them — `time.Time` on the petstore —
   is benign: the type is recognized by identity, so the declaration was never needed, and the lookup only happens
   because the recognizers run *after* it at that call site. It costs the spec nothing and is now visible, which is
   Q37's ordering residual surfacing rather than a new fault. Pinned loosely in the test on purpose.

   Original entry: `a193bcf` touched `exportdata.go`, which serves the toolchain-free blob; the go/packages route
   still raised its hint and still lost `strfmt`'s marks (`loader_agreement_test.go:218`).

   **Why the two routes differ is structural, not an oversight.** On the toolchain-free route *we are the
   importer*, so `Import(path)` is a per-dependency callback and the marker scan sits in its condition —
   `strfmt` falls through and is read from source like any other package. On the go/packages route `go list` and
   upstream own resolution, and our entire lever is one `LoadMode` (`compiledDepsMode = loadMode &^ NeedDeps` —
   you don't *ask* for export data, you ask for the graph not to be hydrated). A LoadMode is one value for the
   whole load: no per-package hook, nowhere to say "except strfmt". All-or-nothing by construction.

   So 5.2 is **repair after the fact**, not routing: let the wholesale load happen, then re-attach source to the
   marked dependencies. Three facts make it work — `compiledDepsMode` keeps `NeedFiles` (`strfmt` returns
   `GoFiles=9, Types=non-nil, Syntax=0`, so it is `parser.ParseFile` on known paths, 1.7 ms, no second
   `go list`); the marker scan already exists; and `959c159` made the hybrid legal at all, since a package built
   from export-data types plus parsed syntax has no `types.Info.Types` and cannot get one. Before that commit
   this was not implementable.

   **And it can now be eager.** This was gated on laziness because eager attach over the whole closure costs
   0.51 s / 168 MB. That arithmetic died with `83972ff`: the loader already reads source only where the marker is,
   so "eager" now means the annotated dependencies — 2 of 335 on dockerctl.

   ✅ **Both halves of this are closed.** The attach had been deleted from the tree (`a193bcf` added
   `parseForComments` + `bridgeDefs`; `83972ff` deleted both), so 5.2 was a re-creation rather than a switch —
   it now lives in `internal/packages/attach.go` in `master`. And the deletion's argument in
   `carriesAnnotations`' doc — *"there is no way to put the missing half back afterwards … not a degraded scan
   but a panicking one"* — had been made false by `959c159` and was corrected in place.

   **Why not a second, narrower `packages.Load`** — the intuitive reading, scoping the toolchain to the marked
   dependencies to get a real type-check: it mints a **second type universe**. `declForObj` resolves by
   `(pkg path, name)`, so the lookup would succeed and everything would look fine; but everything downstream
   compares by identity — `types.Identical` is object identity for defined types, the post-decl dedup is keyed on
   `Obj()` (`79c0f1d`), embedded-field matching walks objects. Two universes for one declaration fail those
   comparisons **silently**, producing a subtly wrong document rather than an error. Parse-and-join keeps one
   universe and costs no second `go list`.

   **Residual asymmetry, worth knowing.** After 5.2 the routes agree on output from different internal states:
   toolchain-free type-checks an annotated dependency from source (`types.Info` complete), go/packages serves it
   as export-data types + parsed syntax + bridged `Defs`. Acceptable — the second state is precisely what
   `TestSpecIsBuiltWithoutExpressionTypes` certified — but it argues for keeping `written_rhs.go`'s fallback as
   the only `types.Info.Types` reader, since a new one would work on one route and not the other.

3. ⛔ **4.2 — mis-specified; there is no "same narrowing" to give.** Read the importer
   (`internal/packages/loader.go:530`): the toolchain-free route **already has the per-dependency routing** —
   export data wins for a non-main-module import that does not carry the marker, source otherwise, stub if
   unresolvable. It falls through to source only when `exportFS == nil`, and there is no way out of that: export
   data comes from a compiler, and not having one is the definition of this strategy. So the blob *is* the
   narrowing, and the only remaining lever is per-package cheapness, already landed in `9007329`.

   Which inverts the item: the routing this strategy has is what 5.2 is copying **to** the other one.

### Design record — kept because the reasoning is load-bearing

4. ✅ **`bridgeDefs` left `types.Info.Types` nil — relayed, confirmed, fixed** (`83972ff`, 2026-08-05). The
   prediction below was right and is now witnessed end to end by a fixture
   (`internal/integration/export_dependency_test.go`): the `swagger:model` case **panicked** and turned the whole
   scan into an error; the discriminated-subtype case would have been dropped silently, but nothing reached that far
   because the first fault is fatal. Kept here because the *resolution* opened the next one — see action 5.

   Original finding: `a193bcf` filled only `Defs`. Five production sites read `TypesInfo.Types[…]`, and a nil map
   returns the zero `TypeAndValue`:

   - `builders/spec/subtypes.go:184` — `if !known { continue }`, so a **discriminated subtype declared in an
     export-data dependency drops silently out of the reverse-allOf index**. That is precisely the backwards
     discovery this plan is built around, failing quietly under the new option.
   - `builders/schema/schema.go:227` — no guard: `resolvers.MustBeAType(ti)` **panics** on the zero value.
   - `schema/interface.go:99`, `schema/walker_classifiers.go:296`, `responses/responses.go:271` — guarded, degrade.

   ✅ **Since witnessed.** This was recorded as a prediction — the tests of the day sat at the
   `internal/packages` level and nothing drove the builders over an attached export-data dependency. The fixture
   asked for now exists (`internal/integration/export_dependency_test.go`) and confirmed it: the `swagger:model`
   case panicked, and the discriminated-subtype case would have dropped silently had anything reached it.

5. ✅ **DONE — the five readers were what forced the loader's per-dependency choice, and they are gone**
   (`959c159`, 2026-08-05). All five migrated; `written_rhs.go` is now the only `types.Info.Types` reader left, as a
   fallback. `TestSpecIsBuiltWithoutExpressionTypes` builds six corpora twice, once with `TypesInfo` reduced to
   `Defs` alone, and compares whole documents — byte-identical throughout, with a non-vacuity guard so two empty
   documents cannot pass for agreement. **So a dependency can now be served from export data *with* parsed syntax**,
   which is the input the loader needs to decide whether its per-dependency policy can relax.

   **⚠️ One correction to this plan's own analysis.** Category 2 was described as "recovering *the type* from a
   syntax node, which is redundant when the type is already in hand. Incidental, not essential." That was **wrong for
   three of the five sites**. They want the *written* right-hand side — the `Stamp` in `type StampResp Stamp` — and
   `go/types` genuinely discards it for a defined type: `Underlying()` peels every named layer, and the stdlib
   recognizers key on exactly the layer it peels. The AST is therefore essential; what is incidental is the
   `types.Info` map. `WrittenRHS` resolves the RHS expression against the **package scope**, which is complete
   however the types arrived, and declines generic instantiations rather than guessing. Verified against the
   checker's own record: 16 573 declarations all `types.Identical` (15 declined); `resolvers.Embeds` checked entry
   by entry over 777 declarations / 150 embeds.

   Original reasoning it displaces — `83972ff` reasons
   that a `types.Info.Types` entry cannot be constructed outside `go/types` (the field distinguishing a type from a
   value is unexported), so a package assembled from export data plus parsed syntax cannot be handed to the builders
   *as they are written today*. That is a constraint on **fabricating** the entry, not on the readers. All five sites
   recover *the type* from a syntax node when the type is already in hand — this plan's category 2. Take them off
   `TypesInfo.Types` (`decl.ObjType()`, `types.Struct.Field(i).Type()`, `types.Interface.Method(i).Type()`) and the
   forced choice dissolves: an annotated dependency could be served from export data **plus parsed comments**, rather
   than paying a full type-check because one of its types is used.

   Worth its own measurement before anyone commits to it. On dockerctl the gap is small — `strfmt` is the only
   annotated dependency of 335, and it is 9 files. On a corpus with a large annotated dependency it would not be.
   The risk sits where AST nodes are matched to type-level entities (`subtypes.go:184` walks embedded fields as AST;
   `interface.go:99` the same for interface methods), which is where a faithful-looking migration can quietly
   i.e. diverge without failing a golden.

### ✅ DONE — the name-based field bridge (2026-08-09, `fc3d279`)

**The last output divergence between the dependency-type configurations and an ordinary scan.**
Closed on `fix/derisk-precompiled-dependencies`. The A/B table is now **empty for all three
configurations over the whole fixture corpus** (306 subtests), where it was six rows.

Two halves, and neither works alone — the bridge alone changes nothing, the fetch alone produces
an empty definition:

1. **On demand, not up front** (`ScanCtx.readBackOnDemand`). The marker scan asks what a dependency
   *says*; a model declared in an unannotated dependency is a question about what it *declares*, and
   nothing in that source asks to be read. So the declaration is fetched at the lookup that wants it
   (`FindDecl` / `FileForPos`), one package parsed and bridged to the export-data scope. Threaded
   from the loader (`Loader.ReadBackSource`) rather than a free function, so `Options.FS` is honoured
   — which is what also closed the two `export-data` rows, once `exportedPackage` learned its
   `GoFiles` from the memoized `sourceFiles`.
2. **Matched by name and line** (`resolvers.FindASTFieldFor`). Position first; on a miss, the field's
   own name plus the line, embedded fields via the last identifier of the type expression. Wired at
   all seven field-lookup sites (schema fields/methods/allof, parameters ×2, responses ×2).

**The measurement that decided the shape.** Widening the read-back to every non-stdlib dependency —
the obvious fix, and the one the earlier attempt tried — gives the same agreement and costs
**+37% wall, +34% RSS, +39% alloc** on kubeapi (1.25 s / 415 MB / 619 MB vs 0.91 s / 309 MB /
446 MB). On demand is free: alloc identical to four digits, wall and RSS inside run-to-run spread.

**The positional wall, resolved rather than worked around.** Export data *does* keep filename and
line; only the column is fabricated (always 1, the importer's synthetic line table). That is what
makes a name+line match exact rather than a guess. Recorded because the old note said field-level
correspondence was unreachable — it was reachable, just not by position.

**Witnesses** (all fail without the change, verified by stash-and-run):
`TestDependencyDeclarations_ReadBackOnDemand` (whole-document equality vs a full-source scan, plus
the interface-method and embedded-field shapes), `TestReadBackOnDemand_PaysPerDeclarationWanted`
(the boundedness claim: N lookups → N packages parsed, idempotent), and nine `TestFindASTFieldFor`
subtests that reproduce the two-`token.File` situation rather than describing it.

**Retired with it:** `sourceless_type_test.go` — its three compiled-dependencies subtests pinned
losses that no longer happen. The degrade-and-warn machinery is still live and still covered by
`TestExportOnly_ReportedWhereItCosts`, which drives it from a virtual filesystem where the source
genuinely cannot be reached (`SourcelessFallback` 77.8%).

**So the prerequisite for defaulting the option on is met — and it still should not default on.**
What keeps it opt-in is now cost alone: cost inverts on a cold cache (12.7–13.4× slower, 229 MB of
build cache), and it needs the closure to BUILD, not merely type-check.

### Related quirks — same seam, tracked separately

6. ✅ **[Q38](../quirks-open.md) — FIXED** (`1d5ce79`, merged via PR #81). `recognizeOpaqueStream` joined the
   canonical `ApplyStdlibSpecials` set, so stdlib IO interfaces are no longer drilled into the spec.
   `io.Writer` surviving in the A/B table is **not** this quirk: there the type is deliberately outside the
   opaque-stream set, and what it loses is prose, not a recognizer. Q38's third half — a graph omitting `io`
   entirely — stays parked with `StubStdlib`.

### ⛔ Retired

Kept as one-liners so nobody reopens them from the Trajectory numbering alone.

| was | why |
|---|---|
| The three-prerequisite gate before the lazy parse | there is no lazy parse; two of three landed anyway |
| The declaration handle's *three* states | absence is permanent, never deferred — two states was the right model |
| Does the incremental closure equal the measured one? | no incremental loader was built |
| What does a partial closure do to `detectDegradedLoad`? | same — the load stayed whole-graph |
| Whose `Materialise` is it? | nothing to materialise |
| Base this stream on `feat/source-loader` | ✅ done — 16 commits, rebased twice |
| Fix `Names()` / `DefKey()` off `Ident.Name` | ✅ done (`c5ee6c0`); the `Comments` residual moved to the builder plan |
| Type-check roots with a lazy importer and measure | ✅ answered by `compiledDepsMode`: 0.40 s / 106 MB vs 0.91 s / 513 MB |
| A large main module (unrepresented corpus shape) | moved to [`pull-based-builder.md`](pull-based-builder.md) |
| The 322-vs-336 graph divergence between strategies | belongs to `feat/source-loader`; surfaced here by `scanbench modes` |

## Achievements

### Branch 2 — contract, migration, regexp removal (`on-demand-scanner`, 2026-08-05)

Four commits on `feat/source-loader@83972ff`. Independently verified: `go vet` and the full suite green,
`golangci-lint --new-from-rev` clean, **zero changes under `fixtures/`** — no golden touched or regenerated.

1. ✅ `c5ee6c0` — the type half always answers: `Name()`, `Pos()`, `PkgPath()`, `IsAlias()` built on `Obj()`;
   `Names()`/`DefKey()` off `Ident.Name`. ⭐⭐
2. ✅ `2e76a6b` — ~45 builder sites onto the type half. The duplicate-model resolver compared `Ident` pointers and
   now compares `Obj()` identity; `Spec.Assign.IsValid()` became `IsAlias()`. ⭐⭐
3. ✅ `959c159` — **all five `types.Info.Types` readers migrated**, with the constraint-lifted witness. ⭐⭐⭐
   The headline result of the stream so far: see action 5, including the correction it forced on this plan.
4. ✅ `79ab439` — regexp-free classification. The retired expressions survive *in the test* as the specification,
   compared verdict-and-capture over ~294 000 generated lines plus a fuzz target (13.6 M execs). One deliberate
   divergence, found by that differential: `\S` excludes only ASCII whitespace, so a non-breaking space followed by
   a tab used to start the argument *on* the separator. An argument now begins where separators end. ⭐⭐
   `rxRoute`/`rxOperation` kept deliberately — four things in sequence read back by submatch index.

⛔ Commit 5 (scanner-side byte prefilter) dropped, not started — pre-empted by `83972ff`.

5. ✅ `c9c392d` — the export-data policy's justification, which had expired. `carriesAnnotations` argued the
   whole-package choice from an impossibility that `959c159` had already removed; stated in two places, both
   corrected. ⭐
6. ✅ `84e0b58` — **5.2: the go/packages route keeps a dependency's annotations.** ⭐⭐⭐ The two strategies now
   reach the same policy from opposite ends — per-import during the load where we own resolution, after the load
   where upstream does. Four A/B divergences closed, and the two that didn't produced the better finding (see
   action 2). Sabotage-verified.
7. ✅ `46c4da3` — **Q37: a type whose declaration cannot be read is rendered, not fatal.** ⭐⭐⭐ Closes the last
   blocker on `CompiledDependencies`. The A/B table is twelve rows → six, and none of the six is a failure.
   Sabotage-verified; `TestSourcelessType_DegradesInsteadOfFailing` pins all three outcomes (degrade+warn,
   silence where a recognizer answers, and the fallback landing on the full-source answer).
8. ✅ `05b6051` — a lost *definition* warns too, not only a thinned schema. Narrow rule: only
   `resolveRefOr`'s no-fallback arm. Two new warnings over twelve bundles, both real, zero false positives. ⭐⭐
9. ✅ `c0fad48` — **the compiled-dependencies announcement now comes from the strategy that ran.** Found while
   answering "when does this option fire": `Options.FS` forces the toolchain-free loader, which drops the
   option, but the Hint fired from the request — so a scan was told its dependency types came from export data
   while it read every dependency from source. `Loader.Strategy()` is now exported and the scanner asks it; a
   dropped option raises a Warning instead of nothing. ⭐⭐

**Carried forward** to [`pull-based-builder.md`](pull-based-builder.md): `Names()`/`DefKey()` still read `Comments`
for the `swagger:model` override, so a declaration whose syntax was *deferred* rather than absent would silently fall
back to its Go name. Harmless today — absence is permanent, never deferred — and stated explicitly in its test.

### Branch 1 — the harness (`bench-grammar`, 2026-08-05)

Two commits on `feat/source-loader@83972ff`, in `internal/integration/`. Cost axes gated behind `CODESCAN_BENCH=1`;
the spec A/B runs by default. Lint clean, vet clean, agreement suite green.

1. ✅ `e2223a5` — all eight axes, each rung reporting **cost and shape**. Warm is enforced (a discarded practice run
   per rung); cold gets a private empty `GOCACHE`. Axis E is three-way per package — `fromSource` / `exportServed` /
   `exportOnly`, the last read through the loader's own `WithOnExportOnly`. ⭐⭐⭐
2. ✅ `885f3a2` — the spec A/B across dependency-type configurations, with a divergence table that **fails when
   stale**. ⭐⭐⭐

**Figures corrected against mine** — take the harness's, not this document's earlier ones:

- Cold `CompiledDependencies` is **9.08 s**, not the 4.06 s recorded here; mine was a partially-warm cache. The
  default rung moves only 0.93 → 1.46 s cold, so the option is ~2.3× faster warm and ~6× slower cold.
- Toolchain-free wall clock is **1.12–1.22 s**, not 0.97 s. The retained-heap advantage (277 MB vs 512 MB) stands.
- Axis C measured for the first time: bundled `ExportData` 0.39 s / 87 MB (dockerctl), 0.33 s / 30 MB (go-swagger).

**Axis H is now small, and that is the answer.** The per-dependency policy already captures nearly the whole prize
(0.39 s → 0.04 s on dockerctl). What remains between the loader's substring rule and the strict line-start rule is
~0.03 s / 11 MB on dockerctl and ~nothing on go-swagger. **The next win is not there** — this closes the question
this document opened with.

**Operational hazards found and fixed in the harness:** handing a corpus the wrong export-data blob silently falls
back to source and still looks like a valid measurement (0.70 s/123 MB vs 0.33 s/30 MB on go-swagger), so blobs are
per-corpus env vars. And `graph` is not a closure size under `ExportData` — export-served packages carry an empty
`Imports` map, so the walk stops there; it means "packages materialised", which is what retained heap tracks.

**~~Known-broken~~ — ✅ closed 2026-08-08.** `hack/genexportdata` could not cover the fixtures module because
`fixtures/go.sum` lacked entries for `go-openapi/swag/{jsonutils,stringutils}` (via `runtime v0.29.3`). The
`go mod tidy` that resolved the re-squash conflict on that module fixed it as a side effect: `go list -export
./...` under `fixtures/` now exits 0 with no errors, so the fixtures blob is generatable and the derisking work
can use it as an axis rather than only `std`.

### Design

1. ✅ **The discovery/loading split identified and quantified** (2026-08-05) ⭐⭐⭐
   - Reframed the stream away from "AST vs types". The original framing would have optimised the wrong axis.
   - 95–99% of the loaded graph proven unnecessary, on two independent targets.
   - Established that syntax materialisation needs **no second `go list`** (`GoFiles` + `parser.ParseFile`, 1.7 ms for
     a 9-file package) — which retires the migration risk the rambling called out as its largest
     ("multiple `packages.Load` calls may be slower than one big one"). There is only ever one `go list`.
2. ✅ **Two design questions settled by Fred** (2026-08-05) ⭐⭐
   - Discovery stays exhaustive — backwards subtype discovery and late `$ref` make load-time pruning unsound.
   - The fixpoint loop is recast into the cache, not removed.
3. ✅ **WASI/on-demand framed as one mechanism** (2026-08-05) ⭐⭐ — same handle, one knob (materialisable vs absent),
   rather than two opposed streams needing a compromise.
4. ✅ **Reconciled with `feat/source-loader`** (2026-08-05) ⭐⭐⭐
   - The branch's two perf axes (cheaper per package, fewer packages) are orthogonal to this stream's (recover what
     fewer packages costs). No overlap, no rework — but the baseline moves, so the "before" figures were restated.
   - Found that `CompiledDependencies`' documented fidelity cost is **recoverable in 1.7 ms**: the option keeps
     `GoFiles` populated for the very dependency whose comments it gives up. This turns the stream's deliverable from
     a general performance refactor into a specific, testable one.
   - The `EntityDecl` contract stopped being hypothetical — there is a shipped configuration that produces the
     type-without-syntax state today.

---

## Appendix A′ — the 2026-08-08 snapshot (v0.36.3 → master, one machine)

**This is the baseline. Appendix A below is historical only.** Measured 2026-08-08, go1.26.5, one machine
(Ryzen 7, 31 GB). Method: a throwaway external module calling the **public `codescan.Run`** — the only
surface both versions have — one scan per process (peak RSS is a process high-water mark), a discarded
warm-up, then alternating rounds so cache and thermal drift cannot land on one configuration. v0.36.3 has
exactly one way to load and scan, which is what makes it the baseline.

The workload looks **RAM-bound rather than CPU-bound**, which is why the old i5 box and this Ryzen 7
produced such similar figures: the two differ sharply in CPU and hardly at all in RAM and SSD.

Three corpora, chosen so that size and vendoring vary independently:

| corpus | shape | `.go` files | emitted |
|---|---|---|---|
| `dockerctl` | go-swagger-generated docker client | 1487 | 198 defs / 0 paths |
| `kubeapi` | larger generated server, module, **not** vendored | 2352 | 222 defs / 260 paths |
| `kubeapi2` | the same tree **with vendored dependencies** | 2352 + 646 vendored (14 MB) | 222 defs / 260 paths |

**Warm build cache. Every configuration emits the identical document on its corpus** — the non-vacuity
check: nothing here got faster by scanning less.

| configuration | wall | Δ | peak RSS | Δ | total alloc | Δ |
|---|---|---|---|---|---|---|
| **dockerctl** — v0.36.3 | 1.227 s | — | 672 MB | — | 1103 MB | — |
| master, default | 1.019 s | −17% | 681 MB | +1% | 1103 MB | 0% |
| master + `ToolchainFreeLoader` | 1.124 s | −8% | 397 MB | −41% | 585 MB | −47% |
| master + `CompiledDependencies` | 0.434 s | −65% | 169 MB | −75% | 237 MB | −79% |
| **kubeapi** — v0.36.3 | 1.757 s | — | 747 MB | — | 1214 MB | — |
| master, default | 1.532 s | −13% | 749 MB | 0% | 1214 MB | 0% |
| master + `ToolchainFreeLoader` | 1.362 s | −22% | 424 MB | −43% | 644 MB | −47% |
| master + `CompiledDependencies` | 0.956 s | −46% | 310 MB | −59% | 446 MB | −63% |
| **kubeapi2** (vendored) — v0.36.3 | 1.678 s | — | 789 MB | — | 1214 MB | — |
| master, default | 1.477 s | −12% | 774 MB | −2% | 1214 MB | 0% |
| master + `ToolchainFreeLoader` | 1.308 s | −22% | 404 MB | −49% | 640 MB | −47% |
| master + `CompiledDependencies` | 0.921 s | −45% | 314 MB | −60% | 446 MB | −63% |

### What the matrix says

**1. The default path does not prune, and that is the whole explanation.** Default → default is −12 to −17%
wall on all three corpora with **total allocation identical to four digits** and peak RSS inside the spread.
This is not annotation density defeating the prefilter — it is that the pruning lives in the
*toolchain-free importer* (per-import source-vs-export-data) and in `CompiledDependencies` (export data),
and the default configuration selects **neither**. Ask for a pruning loader and the pruning appears
immediately: −47% allocation, on every corpus, invariably.

**2. So the default's gain is CPU, not memory** — same bytes allocated, less time spent on them. That points
at the scan/build side (classification off regexes, the builders off `types.Info.Types`), not at the loader.
Worth stating because it sets what a default-path user can expect: a modest time win and **no memory win**.
Never quote the −75% against the default path.

**3. Size magnifies the pure-Go loader exactly as predicted.** On dockerctl the toolchain-free route is
*slower* than master's default (1.124 s vs 1.019 s, its known concurrency gap); on the larger kubeapi it is
**faster** (1.362 s vs 1.532 s). Its memory advantage is flat at −43/−49% across every corpus and both
vendoring modes, so what scale changes is only whether the wall-clock deficit survives. It does not.

**4. Vendoring is a much smaller effect than hoped — report it as inconclusive, not as a win.** Comparing
`kubeapi2` against `kubeapi`, vendoring buys everyone ~4% wall clock, and on peak RSS it *helps only the
toolchain-free route* (424 → 404 MB, −4.7%) while making both go/packages rows slightly worse (747 → 789 MB
for v0.36.3). Directionally right, but 14 MB of vendored source against a 2352-file module is too small a
fraction of the closure to magnify anything. A corpus whose dependency closure dominates its own source
would be needed to make this axis speak.

### Cold build cache — the axis that reorders everything

One run per cell, each under a **private empty `GOCACHE`** (`GOMODCACHE` left warm, so this isolates the
build cache and nothing else). "Cold" is the honest default state for a freshly generated and tidied tree
that has never been built — and for CI, always.

| configuration | warm | cold | cold/warm | build cache it produces |
|---|---|---|---|---|
| **kubeapi** — v0.36.3 | 1.757 s | 2.193 s | 1.2× | 7.6 MB |
| master, default | 1.532 s | 1.965 s | 1.3× | 7.6 MB |
| master + `ToolchainFreeLoader` | 1.362 s | **1.294 s** | **1.0×** | **4 KB** |
| master + `CompiledDependencies` | **0.956 s** | 12.161 s | **12.7×** | 229 MB |
| **kubeapi2** — v0.36.3 | 1.678 s | 2.362 s | 1.4× | 7.4 MB |
| master, default | 1.477 s | 2.052 s | 1.4× | 7.4 MB |
| master + `ToolchainFreeLoader` | 1.308 s | **1.380 s** | **1.1×** | **4 KB** |
| master + `CompiledDependencies` | **0.921 s** | 12.362 s | **13.4×** | 229 MB |

**The toolchain-free loader is cache-independent, and that had never been measured.** It produces 4 KB —
nothing — and its cold time equals its warm time to within noise, because it never invokes the go command:
there is no metadata to populate and nothing to compile. Every other configuration is paying, cold, for a
toolchain it has to prime.

**So the ranking inverts on cache state, and no single configuration wins both:**

| | fastest → slowest |
|---|---|
| **warm** | `CompiledDependencies` 0.96 s · toolchain-free 1.36 s · default 1.53 s · v0.36.3 1.76 s |
| **cold** | **toolchain-free 1.29 s** · default 1.97 s · v0.36.3 2.19 s · `CompiledDependencies` **12.16 s** |

Cold, the toolchain-free route is **41% faster than the v0.36.3 baseline and 34% faster than master's
default** — while also holding 43% less memory. It is the only configuration whose cost is *predictable*,
which for a tool that runs in CI is worth more than a warm-cache best case.

**And this sharpens 5.6 rather than merely confirming it.** `CompiledDependencies` is not "2.3× faster warm,
6× slower cold" — on these larger corpora it is **12.7–13.4× slower cold**, and it materialises a **229 MB**
build cache to get there, because `go list -export` must *compile* the closure rather than type-check it. A
first run on a freshly generated tree is exactly the case it loses hardest. Keeping it opt-in was right, and
the derisking should be honest that its win is conditional on a cache the user already paid for.

⚠️ **Read the warm table with this in mind.** Those figures are warm only because the harness warms each
configuration first. On a freshly generated, tidied, never-built tree — which is how both kubeapi corpora
were created — the *first* scan is the cold column.

**Why the old absolutes nearly reproduce.** Cold default 1.48 vs Appendix A's 1.46 s, cold compiled-deps
9.03 vs 9.08 s — across two machines with very different CPUs. Consistent with the RAM-bound reading above;
it is not evidence that the old numbers were sound, only that CPU was never the binding constraint here.

**Not comparable between the two appendices:** "retained heap". Appendix A measured the loaded package graph
while it was held; this harness measures what survives after `Run` returns, when the graph is garbage and
only the spec is live (~2.5 MB in every configuration). Use peak RSS and total allocation for the memory
story.

**Second corpus, with a caveat that disqualifies it as a corpus.** `go-swagger/go-swagger ./...` **fails to
scan on both versions**, identically — `already annotated as parameters, can't also be "model"`. Long-standing,
not a regression, but the run measures load-then-abort and emits zero definitions. Its wall clock (1.049 s →
0.879 s, −16.2%) is consistent with dockerctl only because the load dominates both. It needs a narrower
pattern before it can serve as a second shape.

## Appendix A — evidence (HISTORICAL — superseded by A′)

> [!CAUTION]
> **These figures are not a baseline and must not be compared against a fresh run.** They were taken
> 2026-08-05 on a **different, older and slower machine**, with `hack/scanbench` — a probe that never landed
> (see 1.1). So the absolute numbers carry an unknown hardware factor, and the tool that produced them is
> gone. Every ratio *within* a table below is still meaningful (same machine, same run); nothing across
> tables or across machines is.
>
> **The refresh is a fresh snapshot, not a re-measurement:** baseline **v0.36.3** and current `master`, both
> on one machine, through the **public `codescan.Run` API** — the only surface that exists in both, since the
> bench harness lives inside the module and so does not exist at the tag. Only the default configuration is
> comparable; `CompiledDependencies` / `ToolchainFreeLoader` / `ExportData` are new capability, not speedups
> over a baseline that had them.

Measured 2026-08-05, go1.26.5, warm build cache and warm page cache. Targets:
`go-swagger/dockerctl` (generated docker client, 502 files / 336-package closure) and `go-swagger/go-swagger`
(624 files / 440-package closure). **Not reproducible as written** — `hack/scanbench` does not exist; the
landed equivalent is `CODESCAN_BENCH=1 go test ./internal/integration/ -run Bench` with the corpus env vars.

### Where the time and memory go

Full `codescan` run on dockerctl (`./...`, `ScanModels: true`):

| phase | wall | retained heap |
|---|---|---|
| `NewScanCtx` (load + index) | **1.29 s** | 621 MB |
| `spec.Build` | 0.03 s | +15 MB |
| **peak RSS** | **1.37 s** | **690 MB** |

Scanning is **94%** of wall clock. Within it, `packages.Load` alone is ~0.97 s — the `TypeIndex` walk is ~0.3 s. The
cost is the load, not our walk.

### What each LoadMode costs (dockerctl, 336 packages)

| mode | wall | retained heap | AST files held |
|---|---|---|---|
| **current** (`Deps+Types+Syntax+TypesInfo`) | 0.90 s | **513 MB** | 2354 |
| all types, no syntax (`Deps+Types`) | 0.60 s | 81 MB | 0 |
| **`compiledDepsMode`** (`&^ NeedDeps`, = `CompiledDependencies`) | **0.36 s** | **107 MB** | 502 |
| syntax only, no types | 0.13 s | 39 MB | 502 |
| `go list` only (`Name+Files+Imports[+Deps]`) | 0.08 s | 2 MB | 0 |

And the same question asked of `internal/packages`, which is what the scanner actually runs
(`scanbench modes`, second run):

| loader | wall | retained |
|---|---|---|
| go/packages (default) | 0.92 s | 514 MB |
| go/packages + `CompiledDependencies` | **0.36 s** | **107 MB** |
| toolchain-free | 0.97 s | 279 MB |

The toolchain-free strategy retains ~45% less than go/packages for the same tree — the per-package cuts of `9007329`
— while running slightly slower, which is the concurrency gap its commit message calls out. It also reports a graph
of 322 against go/packages' 336; that divergence belongs to `feat/source-loader`, not here, but it is visible from
this harness and worth someone's attention.

All warm. An earlier cold reading put the `compiledDepsMode` row at 4.06 s, because export data has to be compiled
before it exists — which is the cold-cache caveat `CompiledDependencies` documents, not a steady-state cost.

Critically, `compiledDepsMode` keeps **`GoFiles` populated for all 336 packages** while holding syntax for only the 19
roots. `strfmt` comes back `GoFiles=9, Types=non-nil, Syntax=0` — types complete, source locatable, AST absent. That
is the whole design in one line of output.

### Discovery, separated from loading

| step | what | cost |
|---|---|---|
| 1 | `go list`: closure graph + `GoFiles` (335 pkgs, 2341 files) | 0.09 s, 3 MB |
| 2 | read all 2341 files (22.4 MB of source), byte-scan for `swagger:` | 0.03 s, transient |
| 3 | parse + classify the 127 survivors (**5.4%**) | 0.02 s, transient |
| | **⇒ complete annotation index over the whole closure** | **0.14 s, ≈0 retained** |

The byte prefilter is safe as a *pre*-filter: it can only over-admit (a `swagger:` inside a string literal costs one
wasted parse), never under-admit, since a real annotation contains that literal by definition. Everything surviving it
goes through the real classifier.

### The working set

| target | closure loaded | annotated roots | type-reference closure | never needed |
|---|---|---|---|---|
| dockerctl | 336 | 2 | **4** | 332 (98.8%) |
| go-swagger | 440 | 14 | **19** | 421 (95.7%) |

Packages pulled in purely by type-following, carrying no annotation at all — dockerctl: `oklog/ulid`, `time`;
go-swagger: `strfmt/internal/countries`, `scan-repo-boundary/makeplans`, `oklog/ulid/v2`, `x/text/currency`, `time`.

Caveat: the closure walk follows struct fields, elems, map keys/values, type args, interface methods and signature
results from annotated decls — an approximation of the schema builder, not a replay of it. See action 2.

### Eager vs lazy parse-for-comments (the remaining prize)

Same operation `a193bcf` performs, timed over the whole closure against the closure a scan actually reaches:

| parse-for-comments | wall | retained | packages | files |
|---|---|---|---|---|
| eager — every dependency | 0.51 s | 168 MB | 335 | 2341 |
| lazy — only what is reached | 0.02 s | 6 MB | 4 | 102 |

For scale: `CompiledDependencies` itself is 0.36 s / 107 MB, so the eager parse more than doubles both.

### Syntax materialisation without a second `go list`

The cheap load populates `GoFiles` for **all 336** packages (`pkgsWithGoFiles=335/335` even with no types requested), so
materialising a package's syntax later is plain `parser.ParseFile` on known paths:

```
on-demand syntax for strfmt: 9 files in 1.7 ms (698 comment lines) — no go list
dep type strfmt.Date resolved from export data: struct{wall uint64; ext int64; loc *time.Location}
```

### Dependency annotations

```
github.com/go-openapi/strfmt: 155 `swagger:strfmt` annotations in its own source
```

## Appendix B — the two witnesses (from the WASI corpus A/B)

From the export-data loader A/B against `go/packages` — 141 of 143 patterns byte-identical; **both failures are the
declaration contract, not a type-fidelity gap**:

- `enhancements/in-case-insensitive` — errors with `unable to find package and source file for: io.Reader`. The type is
  fully known; only the declaration is missing.
- `goparsing/go123` — a stdlib named type used as a field emits `{}` where a full graph emits
  `$ref: #/definitions/Duration`. The underlying type is `int64` and is right there in the type information; the builder
  cannot use it because it asked for a declaration first.

The reference behaviour in the second case is itself questionable — it is Q38(b). A fix should probably not aim to
reproduce it exactly.

**Not a fidelity problem.** Every case that *stubbing* the standard library degrades (`text-marshal`,
`raw-message-override`, `wrapper-decl-type-override`, `response-edges`) is byte-identical under export data, because the
types are the compiler's own. What remains is purely the declaration contract.
