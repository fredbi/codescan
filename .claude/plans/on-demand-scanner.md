> [!NOTE]
> Last revision: 2026-08-02

# On-demand scanner (Stream 8 · `V-scanner`)

## Summary

Make the scanner parse types **on demand** instead of loading and type-checking everything up front: syntax first,
type information materialised only for constructs the API surface actually reaches. Benchmarks in the design
rambling predict ~3× on wall clock and a large drop in peak memory, and it closes the perf/OOM cluster of Stream 9
tickets. Target: released in September.

No spec has been written yet. The vision is `ramblings/index-builder-statefulness.md`; the roadmap row is
`V-scanner`, with `V-builder` (pull-based spec builder) hard-depending on it.

**This document opens the stream with one thing only: a contract requirement discovered from another stream.** The
`EntityDecl` type silently assumes every declaration has both a type half and a syntax half. On-demand breaks that
assumption, and so does the WASI work, from the opposite direction. Whoever picks this up should settle the
contract before the mechanics.

## Context

Two streams push on the same axis, in opposite directions:

- **On-demand scanning** wants *syntax now, types later* — hold the AST, materialise `go/types` information only
  when something reaches the construct.
- **The WASI/export-data work** (`wasi-build.md`) is *types now, syntax never* — the standard library is read from
  precomputed `gcexportdata`, which carries a complete `*types.Package` and no AST at all.

Both violate the same implicit invariant: that an `EntityDecl` handed to a builder has a populated `Type` **and** a
populated `Spec`/`Ident`/`File`/`Pkg`/`Comments`. Today nothing states that invariant and nothing enforces it —
the fields are exported and dereferenced directly across the builders.

That is why the contract belongs here rather than in the branch that found it: a fix owed to on-demand regardless,
which the WASI stream would otherwise have to route around by letting a builder bypass the scanner's index.
Decided 2026-08-02 (Fred): builders must not do that — resolving a type to a declaration is the scanner's job.

### Why the WASI stream is types-first — rationale and constraints

Not a preference. The target environment removes every alternative:

- **No toolchain, no exec.** A WASI guest cannot run `go list`, so nothing can be asked of the go command at scan
  time. Whatever a dependency's types are, they must already be in the guest or on a mounted filesystem.
- **Memory is the binding constraint, not time.** Type-checking the standard library from source costs **681 MB
  peak** for a fixture as small as the petstore (wasmtime). That is not something a browser tab can host. Reading
  the same packages from export data costs **138 MB** — and 1.00 s against 7.30 s.
- **The cost is compute, not I/O.** Measured: filesystem syscalls are **1.8%** of a full WASI scan. So the saving
  does not come from avoiding reads; it comes from not parsing and type-checking 190 packages and 1195 files, on
  top of wasm's ~5–6× compute tax. Shipping the *source* more cleverly cannot help — only not doing the work can.
- **Withholding the standard library entirely was the alternative, and it costs correctness.** `StubStdlib`
  synthesises stdlib types from usage: same speed and footprint, but no fields and no method sets, so
  `encoding.TextMarshaler` detection, `json.RawMessage` as a byte array and interface identity all break
  (138/143 against export data's 141/143). Export data is what buys the footprint *without* the fidelity loss.
- **Version-locked by nature.** Export data is tied to the Go release that produced it, which is why it ships as a
  separate, regenerable artifact (9.3 MB, 4.2 MB compressed) rather than being baked into the library.

### What this stream must accommodate, and what it need not

The bound is narrower than "declarations may lack syntax":

- **Only dependencies lose syntax, never the code under scan.** Export data is gated on standard-library import
  paths; the main module and third-party dependencies are still parsed from source, exactly as today.
- **Comments for a dependency are not merely absent — they are unwanted.** The standard library carries no swagger
  annotations, and reading its godoc is what produces Q38(b).
- **`TypesInfo` for a dependency is not needed.** Its uses recover a type from a syntax node, and the type is
  already in hand.
- **Positions do not degrade.** Export data populates the `FileSet`, so diagnostics keep pointing somewhere real.

Which suggests the two directions are complementary rather than opposed, and that the balanced contract turns on
*which half is optional, for which population, and with what failure mode*:

| | population | missing half | nature | what the caller should get |
|---|---|---|---|---|
| on-demand | the module under scan | types (transiently) | resolvable — materialise it | a blocking accessor that forces the work |
| export data | standard library | syntax (permanently) | structural — nothing to materialise | an accessor that can honestly answer "no source" |

If that reading holds, `Obj()`/`ObjType()` may legitimately *block* while `Syntax()` may legitimately *not exist* —
two different API shapes for two different absences, rather than one nullable struct standing for both.

### What the export-data path actually loses

Nothing about types. `gcexportdata` returns a complete `*types.Package`: scope, objects, fields, method sets,
underlying types, interface identity. Positions survive too — `time.Duration` reports
`$GOROOT/src/time/time.go:915:1` with `Pos().IsValid()` true.

What it has no notion of is **syntax**: no `*ast.File`, no `*ast.TypeSpec`, no `*ast.CommentGroup`. Since
`FindDecl` walks `pkg.Syntax` (`internal/scanner/scan_context.go`, `index.go:178`), such a package can never enter
the index and `DeclForType` can never hit for it. Not "sometimes incomplete" — structurally impossible.

## The `EntityDecl` contract

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

1. **Positions** — the largest group (`Ident.Pos()`, `Spec.Pos()`). Needs no syntax: `Obj().Pos()` answers, and
   export data populates it.
2. **`TypesInfo.Types[decl.Spec.Type]`** — recovering *the type* from a syntax node, which is redundant when the
   type is already in hand. Incidental, not essential.
3. **Syntax-shape checks** — `Spec.Type.(*ast.InterfaceType)`, `Spec.Assign.IsValid()`. Both answerable from
   `go/types` (`*types.Interface`, `*types.Alias`).
4. **`File.Imports`** — one site, genuinely syntax.
5. **`Comments`** — genuinely syntax, and *meaningless* for a dependency: the standard library carries no swagger
   annotations. Reading them there is what produces [Q38(b)](quirks-open.md) — stdlib godoc leaking into the
   user's spec as `#/definitions/Duration`, `#/definitions/Kind`.

So the essential syntax dependency is much smaller than 81 sites implies.

**Two precedents already in the file.** `Obj()` and `ObjType()` are already type-only, and panic *loudly* on an
impossible state rather than nil-dereferencing — the discipline the rest of the type wants. And
`schema/schema.go:170` already guards `decl.Spec == nil || decl.Pkg == nil || decl.Pkg.TypesInfo == nil`, so the
nil-ness has been felt before, once, locally.

**Two outliers.** `Names()` and `DefKey()` dereference `d.Ident.Name` when `Obj().Name()` would do — the only
type-half methods that reach into the syntax half without needing to.

## Trajectory

1. **Settle the declaration contract** (this document's contribution — do it first)
   > Both the on-demand and export-data shapes need a declaration that can honestly say which half it has.

   1. 🔍 Decide the model: type-only `EntityDecl`, or a second scanner entry point (see Actions)
   2. 📝 Make the type half always-available: `Obj`, `ObjType`, `Name`, `Pos`
   3. 📝 Make the syntax half explicitly optional, enforced by the compiler rather than by discipline
   4. 📝 Migrate the builder call sites — mostly mechanical, most collapse to a position accessor

2. **The on-demand mechanics** — see `ramblings/index-builder-statefulness.md`, not re-specified here
   1. ⬜ `SchemaCache` — build-once, reference-many (the rambling notes this can land in v1.x)
   2. ⬜ `DemandLoader` — lazy package loading; shares its seam with the WASI loader (roadmap: `W-loader-seam ⇄
      V-scanner DemandLoader`, already built as `internal/packages` on `wasi-build`)
   3. ⬜ `AnnotationScan` — lightweight root discovery
   4. ⬜ Retire the fixpoint loop

3. **Downstream**
   1. ⬜ `V-builder` — pull-based spec builder, hard-depends on the above

## Actions

### Decide before implementing

1. 🔍 **Type-only declaration, or a separate entry point?** Note the two absences are not the same shape (see
   Context): the on-demand one is transient and resolvable, the export-data one is structural and permanent. A
   contract that models them as one nullable struct will serve neither well. Two honest models:
   - `FindDecl` may return a declaration whose syntax half is absent. Builders keep their call shape; the contract
     has to make absence impossible to ignore.
   - `FindDecl` keeps returning `(nil, false)` when there is no source, and the scanner answers "what shape is this
     type" through a *different* API. `EntityDecl` keeps honestly meaning "there is source behind this", at the
     cost of a second path through every builder that resolves a type.

   The first keeps builders unchanged; the second keeps the type honest. Worth settling explicitly, because the
   on-demand work will hit it on every lazily-materialised construct.

2. 🔍 **Does the planned rework already reshape `EntityDecl`?** If so, fold this in as a requirement on that design
   rather than landing a separate refactor underneath it.

3. 📝 **Fix `Names()` / `DefKey()` regardless.** They dereference `Ident.Name` where `Obj().Name()` serves. Small,
   independent, correct either way.

### Related quirks — same seam, tracked separately

4. 🔍 **[Q37](quirks-open.md)** — identity recognizers run *after* the declaration lookup that can fail without
   them. Latent with a full graph; the ordering fix is on `wasi-build` pending handover to the quirk stream.
5. 🔍 **[Q38](quirks-open.md)** — stdlib IO interfaces have no recognizer, so they are drilled and leak into the
   spec. Not latent: `io.Reader` already emits an untyped parameter and publishes `#/definitions/Reader` on a full
   graph today.

## Achievements

_(empty — the stream has not started)_

## Appendix — evidence

All figures measured 2026-08-02 on branch `wasi-build`; reproduce with `cmd/genspec` and `hack/genexportdata`.

**Positions survive export data.** `gcexportdata.Read` populates the `FileSet`:

```text
obj=Duration  pos.IsValid=true  position="$GOROOT/src/time/time.go:915:1"
underlying=int64  methods=11
```

**The two witnesses**, from the corpus A/B of the export-data loader against `go/packages` (141 of 143 patterns
byte-identical; both failures are this contract, not a type-fidelity gap):

- `enhancements/in-case-insensitive` — errors with `unable to find package and source file for: io.Reader`. The
  type is fully known; only the declaration is missing.
- `goparsing/go123` — a standard-library named type used as a field emits `{}` where a full graph emits
  `$ref: #/definitions/Duration`. The underlying type is `int64` and is right there in the type information; the
  builder cannot use it because it asked for a declaration first.

Note the reference behaviour in the second case is itself questionable — it is Q38(b). A fix should probably not
aim to reproduce it exactly.

**Not a fidelity problem.** Every case that *stubbing* the standard library degrades (`text-marshal`,
`raw-message-override`, `wrapper-decl-type-override`, `response-edges`) is byte-identical under export data,
because the types are the compiler's own. What remains is purely the declaration contract.
