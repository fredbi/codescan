> [!NOTE]
> **✅ ALMOST COMPLETE** — archived 2026-08-17. D1–D9 are all resolved and section 3's read-only list has been probed. One item outlived the fixes: a resolution-level differential harness, which is the only thing that could guard these semantics — [`backlog.md`](../backlog.md) §4.1.
>
> _Original document follows unchanged._

# Path resolution & build preparation: our loader vs `go/packages`

Scope: how `internal/packages` resolves patterns and prepares the build, compared with
`golang.org/x/tools/go/packages`. Deliberately excludes what we already know we dropped
(Driver protocol, Overlay/ParseFile/Logf, `fs.FS`, `StubStdlib`).

Upstream read at `/home/fred/src/golang.org/x/tools` (`1f62bc851`); `golist.go`,
`packages.go` and `external.go` are **byte-identical** to the `v0.48.0` we build against,
so findings transfer.

Everything under "Verified" was reproduced with a probe against both strategies. Anything
I only read is marked as such.

---

## 0. The comparison isn't with x/tools, it's with `go list`

Worth stating first because it reframes everything below: **`go/packages` performs almost
no path resolution of its own.** `goListDriver` does three things with patterns:

1. splits off `file=` / `pattern=` queries (`extractQueries`, golist.go:208);
2. passes the rest verbatim to `go list` after `--` (`golistargs`, golist.go:862);
3. reads `Dir` / `ImportPath` / `Root` / `Export` back out of the JSON.

`getPkgPath` + `determineRootDirs` (golist_overlay.go) *look* like a resolver but exist
only to reclaim an import path for a file that isn't on disk yet — i.e. overlay support,
which we don't have.

So the thing we reimplemented is not 1 100 lines of x/tools. It's `cmd/go`'s
`modload` + `search` packages, reached through a subprocess. That is why the divergences
below are all semantic (what a pattern *means*) rather than structural.

The corollary for testing: our loaders can agree on output while disagreeing on *which
packages they were asked about*. The 378/0 corpus A/B never touched any of this, because
every fixture scan is `WorkDir=<single-module dir>` + pattern `.`. It proved the parse and
type-check halves, not the resolution half.

---

## 1. Verified divergences

### D1 — `./...` walks straight through nested modules ⚠️ severe — ✅ FIXED

The go command stops `...` expansion at a directory containing its own `go.mod`. Our
`vfs.walkDirs` skip list (`vfs.go:171`) covers `testdata`, `vendor`, `.` and `_` prefixes
— **not module boundaries**.

Synthetic tree, main module `example.com/main`, nested module `totally.different/thing`
at `sub/`:

```
go/packages    -> [example.com/main/a]
toolchain-free -> [example.com/main/a  example.com/main/sub/b]
```

Two faults, not one:

- the package set is wrong — `sub/b` is a different module and not part of `./...`;
- the *identity* is wrong — `pkgPathFor` (resolve.go:210) derives every path from
  `r.modPath` + relative dir, so `sub/b` is announced as `example.com/main/sub/b` when its
  real import path is `totally.different/thing/b`.

On this repo (`./...` from the root, which has four nested `go.mod` files):

```
go/packages    -> 25 packages
toolchain-free -> 468 packages
```

The fabricated paths happen to look right here only because our nested modules are named
`github.com/go-openapi/codescan/<subdir>`. That is a coincidence of naming, not a
mitigation.

**Bites codescan today**: yes. Point the TUI or `Run` at a repo root with nested modules
and the toolchain-free strategy scans vendored themes, fixture corpora and sibling
modules, then attributes them all to the main module.

### D2 — `go.work` is ignored ⚠️ severe — ✅ FIXED

`findModule` (resolve.go:60) walks up looking for `go.mod` and knows nothing about
`go.work`. At a workspace root there is no `go.mod`, so `modRoot` stays empty and
`pkgPathFor` falls back to `return dir` — the raw directory.

```
"./..." from the workspace root:
  go/packages    -> [./...]                       (one error package, no files)
  toolchain-free -> [/tmp/…/m1/a  /tmp/…/m2/b]    (absolute paths as PkgPath)

"./..." from inside member m1:
  go/packages    -> [example.com/m1/a]
  toolchain-free -> [example.com/m1/a]            (agree)
```

Both are broken at a workspace root, differently: upstream returns nothing usable, we
return packages whose identity is an absolute path. Since codescan keys models,
provenance and `x-go-package` on `PkgPath`, ours is the more damaging failure.

**The bigger half, found later.** Pattern expansion is the visible symptom; the real loss
is *import resolution between workspace modules*. `go list ./...` at a workspace root is
an error for the go command too ("directory prefix . does not contain modules listed in
go.work"), so matching it there costs nothing. But scanning a workspace *member* that
imports a sibling is ordinary, and we get it wrong:

```
/tmp/wsprobe/m2 imports example.com/m1/a, resolved by go.work `use`:
  go/packages    m1/a.T = struct{Field string}   synthesized=[]
  toolchain-free m1/a.T = NOT LOADED             synthesized=[example.com/m1/a]
```

`readRequirements` only knows `require` + `replace` from the member's own `go.mod`, so the
sibling is looked up in the module cache, missed, and **synthesized** — fields and method
set gone, silently. Anything structural (embed promotion, `TextMarshaler` detection, byte
arrays) degrades. That, not the root pattern, is what `go.work` support has to fix.

Note this is the same terrain as the known `GOWORK=off` workaround in `NewScanCtx` — but
the workaround addresses the go/packages side. The toolchain-free strategy has its own,
unaddressed version of the problem.

### D3 — `vendor/` loses to the module cache ⚠️ moderate-severe — ✅ FIXED

The go command uses vendor mode *exclusively* when `vendor/modules.txt` exists (go ≥ 1.14).
Our `resolveImport` (resolve.go:222) tries, in order: stdlib → main module → `modDirs`
(module cache) → vendor. Vendor is the **last** fallback, so whenever a cache entry also
exists, the cache wins.

Probe: main module requires `golang.org/x/mod v0.38.0`, with a `vendor/` copy carrying a
sentinel const:

```
go/packages    golang.org/x/mod/modfile -> vendored sentinel present: true
toolchain-free golang.org/x/mod/modfile -> vendored sentinel present: false
```

We read the pristine cache copy; the go command reads the vendored one. A vendored tree
that has been patched — the usual reason to vendor — is silently ignored. We also never
consult `vendor/modules.txt`, so we cannot tell vendor mode from a stray directory.

### D4 — a bare relative pattern means different things ⚠️ moderate — ✅ FIXED

For the go command a pattern without `./` is an **import path**. `locate` (resolve.go:186)
falls back to treating it as a directory relative to `Dir`.

```
pattern "a", main module example.com/main, package in ./a:
  go/packages    -> [a]  (error: package a is not in std)
  toolchain-free -> [example.com/main/a]
```

Ours is more forgiving, which sounds harmless, but the two return *different identities
for the same input* — so `Options.Packages: ["internal/foo"]` produces a different spec
depending only on the strategy.

### D8 — every directory named `vendor` is skipped, real or not ⚠️ low — ✅ FIXED

Found while writing the D1 regression test, which is the only reason it is here.
`skipDir` (vfs.go:171) excludes `vendor` by name anywhere in the tree. The go command
excludes a *vendor directory* — one the module actually vendors into — not a package that
happens to be called that:

```
main module with a top-level vendor/v.go and NO vendor/modules.txt:
  go/packages    -> [example.com/main/a  example.com/main/vendor]
  toolchain-free -> [example.com/main/a]
```

Fix belongs with D3, since both need us to recognise a real vendor directory
(`vendor/modules.txt` at a module root) rather than pattern-match the name.

### D5 — `...` is only supported as a whole-path suffix ⚠️ low — ✅ FIXED

`resolvePatterns` (resolve.go:149) does `strings.HasSuffix(pat, "...")` and trims. The go
command treats `...` as a wildcard anywhere, including mid-path and mid-segment:

```
"./internal/pack..."      theirs -> [.../internal/packages]     ours -> ErrUnresolvedPattern
"./internal/.../grammar"  theirs -> [.../parsers/grammar]       ours -> ErrUnresolvedPattern
```

Low severity only because it fails loudly rather than silently under-matching.
`std` / `cmd` / `all` are documented as out of scope and do fail cleanly.

### D6 — `GOFLAGS` is ignored ⚠️ moderate — ✅ FIXED

`buildTags` (loader.go:135) reads `-tags` from `cfg.BuildFlags` only. The go command also
honours `GOFLAGS` from the environment.

```
GOFLAGS=-tags=custom, no BuildFlags:
  go/packages    -> [plain.go, tagged.go]
  toolchain-free -> [plain.go]
```

Same class: `-mod=vendor|mod|readonly`, `-overlay`, `-modfile` in either `BuildFlags` or
`GOFLAGS` are all silently dropped — we parse exactly one flag out of `BuildFlags` and
ignore the rest. A `-mod=vendor` that the go command would honour is invisible to us,
compounding D3.

### D7 — cgo files type-check with errors ⚠️ low — ✅ FIXED

We now parse `CgoFiles` (correctly — that was the #1096 fix), but we do not run the cgo
tool, so `import "C"` synthesizes an opaque package and *selectors on it* fail:

```
toolchain-free  errors=1: a.go:10:18: undefined: C.malloc
go/packages     errors=0
```

The declarations codescan reads are unaffected (`type A` resolves in both), but the
package is marked `IllTyped` with a `TypeError`, and `detectDegradedLoad`
(scan_context.go:241) turns any non-fatal `pkg.Errors` into a Warning. Confirmed on the
real fixture:

```
bugs/1096, ToolchainFreeLoader=false -> 0 degraded-load diagnostics
bugs/1096, ToolchainFreeLoader=true  -> 1 warning:
    package ".../bugs/1096" did not fully type-check:
    main.go:33:8: undefined: C.malloc (1 error(s)); its definitions may be incomplete
```

So a cgo-using project gets a spurious warning under this strategy. The spec is correct;
only the diagnostic lies. Cheapest honest fix is to suppress `undefined: C.*` errors for
packages carrying `CgoFiles`, since we know exactly why they are there.

This also corrects something I said when landing the cgo fix: the `"C"` *import* resolves
through synthesis, but uses of it do not.

---

## 2. Checked, and NOT different

Recorded so nobody re-investigates:

- **Build-constraint release tags.** A file behind `//go:build go1.21` in a module
  declaring `go 1.16` is included by both. Both evaluate against the toolchain's release
  tags, not the `go` directive.
- **`-tags` via `BuildFlags`.** Honoured identically, including the `-tags=x` spelling.
- **GOOS/GOARCH.** Now identical by construction, and pinned by
  `TestGoPackagesStrategyHonoursTarget` (this was the bug fixed in `3294669`).
- **`testdata` / `_x` / `.x` exclusion from `...`.** Both skip.
- **GOROOT-internal vendoring.** `net/http` loads with an identical exported scope (1278
  entries) under both, so std's own vendored deps are not a practical problem.
- **A workspace member scanned from inside itself.** Both agree.

---

## 3. The read-only list, now probed

Every item from the original "read but not probed" list, with what the probe actually
showed. Three were real defects, one was a new find, and two turned out to be features.

### Fixed

**Versioned `replace`.** Confirmed exactly as suspected: `replace foo v1.0.0 => ./local`
was applied to a `require foo v1.2.0`, silently substituting a dependency the build never
asked for. Now a pinned replace applies only to the version it names; the unversioned form
still applies to all.

**D9 — an unreadable `go.mod` degraded silently** (new; found because the first probe's
`require example.com/dep v2.0.0` was itself invalid Go — a major ≥ 2 needs a `/v2` path).
`readRequirements` returned on parse error, leaving *no* requirement placed, so every
dependency fell through to synthesis. The caller got a wall of `scan.synthesized-import`
warnings and no mention of the single line at fault. Now fatal, naming the file and line:

```
cannot read go.mod: /tmp/…/go.mod:5: require example.com/dep:
    version "v2.0.0" invalid: should be v0 or v1, not v2
```

A note on the parser: `modfile.ParseLax` looks like the forward-compatible choice and is
the wrong one — it exists for reading a *dependency's* go.mod and **deliberately ignores
`replace`** (rule.go:341). Using it silently dropped every replacement. Strict `Parse` is
what this needs; the cost is that a go.mod naming a directive this `x/mod` has not heard of
fails loudly, which a version bump fixes.

**Absolute paths leaked into the spec.** A tree with no module gave every package the
absolute directory as its import path, and that reaches the output:

```
"x-go-package": "/home/fred/tmp/TestS3_NoModuleIdentity.../a"
```

Now the path relative to the scan directory (`"a"`). The go command refuses a module-less
tree outright, so there is nothing to be faithful to — only something not to leak.

### Deliberate differences, kept

**GOPATH / no module at all.** `go list` refuses (`go.mod file not found in current
directory or any parent directory`); we scan the tree and emit a usable spec. For a
scanner that is the better answer, and with the identity fix above it costs nothing.

**`internal/` visibility.** Not enforced, and this is worth keeping. End to end:

```
importing another module's internal package:
  go/packages    -> codescan FAILS: use of internal package … not allowed
  toolchain-free -> scans fine, definitions emitted
```

Enforcing it would make codescan strictly less able to document code that compiles for its
own author. The divergence is real but it runs in the useful direction.

### Checked, not different

**Symlinks.** Identical: neither follows a symlinked directory during a `...` walk, and both
resolve an explicitly named one.

```
./...    both -> [example.com/main/real]
./link   both -> [example.com/main/link]
```

**`ImportMap`.** Moot since D3: its main job upstream is vendored-import rewriting, and
vendoring now resolves directly to the vendor tree under both strategies. Test variants are
the other user, and `Tests` is forced off.

### Still open, by choice

**Completeness of the requirement list** (filed as "MVS" originally — the wrong name, see
below). A module is placeable when the main `go.mod` names it: a `require`, a `replace`, a
workspace `use`, or the vendor tree. We never walk the module graph.

Probed with a fake module cache, varying one thing at a time (A requires B; the main module
imports A, whose API exposes B's type):

| main `go` | dependency `go` | B in main go.mod | result |
|---|---|---|---|
| 1.25 | 1.25 | yes | resolves |
| 1.25 | 1.25 | **no** | **synthesized** |
| 1.25 | **1.16** | yes | resolves |
| 1.16 | 1.16 | no | synthesized |

Row 3 is the one that matters: a dependency declaring go 1.16 resolves fine. **What any
dependency declares is irrelevant to us** — we never read its go.mod for requirements, so
the pruning rules that make a dependency's own version matter *inside the go command's
graph loader* have no bearing here. The only question is whether the main go.mod lists the
module.

That list is complete when the main module is at go ≥ 1.17 and tidy: the go command
guarantees an explicit require for every module providing a transitively-imported package.
Verified on this repo (`go 1.25.0`): 17 modules needed transitively, every one of them
declared.

So the real limits are a main module at **go < 1.17** (only direct requires listed) or an
**untidy go.mod** at any version — and the second is largely theoretical, since `-mod=readonly`
has been the default since 1.16 and a module that needs go.mod updates does not build. Both
surface as `scan.synthesized-import` warnings rather than silence.

**Correction, recorded because it was wrong twice.** This was first written up as an MVS
problem, with the gap attributed to "a dependency declaring go < 1.17 not restating its own
requirements". Both halves were wrong: we perform no version selection (versions are read
already-selected from go.mod), and a dependency's go directive never enters our resolution
at all. We apply one set of rules — our toolchain's — uniformly, which is the right design;
the honest limitation is list completeness, not selection.

## 3b. Status

D1–D8 are all fixed (`a2d92d3`, and the follow-up covering D3/D4/D5/D7/D8). What each now
does, where it deliberately still differs from `go list`:

| | resolution | deliberate difference kept |
|---|---|---|
| D1 | `...` stops at a nested `go.mod`; import paths come from the containing module | naming a nested module's package explicitly is allowed (go refuses), with that module's own path |
| D2 | `go.work` read, `use` + `replace` honoured, `GOWORK` pinnable | — |
| D3 | vendor authoritative only with `vendor/modules.txt`; `-mod=mod` opts out | the module cache stays a fallback for a package vendoring missed, where go refuses the build |
| D4 | a bare pattern is an import path, never a directory | go reports an error *package*; we fail the load |
| D5 | `...` is a wildcard anywhere, incl. mid-path and partial segment | — |
| D6 | `GOFLAGS` honoured, alongside `GOWORK`/`GOEXPERIMENT` on `GoEnv` | `GOEXPERIMENT` baseline is codescan's own build |
| D7 | `undefined: C.*` errors dropped for packages with cgo files | we still do not run the cgo tool, so those selectors have no types |
| D8 | a wildcard skips *inside* `vendor`; a package named `vendor` matches | — |

Section 3's list has since been probed and resolved: versioned `replace`, the unreadable
go.mod (D9) and the absolute-path leak are fixed; GOPATH-mode tolerance and `internal/`
permissiveness are kept deliberately; symlinks and `ImportMap` were never different. What
remains is the completeness of the requirement list, which is not an MVS problem — see the
correction in section 3.

## 4. Reusing the go command's own tests

The Go tree at `/home/fred/src/github.com/golang/go` (go1.26.5) has two things worth taking.

### Taken: the pattern-matching table

`cmd/internal/pkgpattern` is the wildcard matcher, extracted from `cmd/go/internal/search`
precisely so tools other than the go command could use it, and `pat_test.go` is its
specification written as data.

Run against our hand-written matcher it scored **34/37**, all three failures being deep
vendor cases:

```
matchPattern("./vendor/...")("./vendor/foo/vendor/bar")     ours=true  want=false
matchPattern("mycode/vendor/...")("mycode/vendor/foo/vendor/bar")  ours=true  want=false
matchPattern("x/vendor/y/...")("x/vendor/y/vendor/z")       ours=true  want=false
```

Our rule ("if the pattern mentions vendor, drop the exclusion") is too crude: the exclusion
is per pattern *element*, so a pattern may name one vendor element and still have to exclude
a deeper one. Upstream does it by rewriting non-trailing `vendor` elements to a sentinel
codepoint and defining `...` as "anything but that codepoint" — linear by construction,
where the natural hand-written version goes exponential (their comment says so).

So the matcher is now **copied verbatim** (`internal/packages/pkgpattern.go`, BSD-3-Clause,
`LICENSE-BSD-go`), byte-identical to upstream for all three declarations, excluded from our
linters so a future re-sync is a diff rather than a merge. The table rides along as
`pkgpattern_test.go`.

### Not taken yet: the script corpus

`cmd/go/testdata/script/*.txt` is 882 txtar scripts, **478** in families that touch our
surface (`list_` 74, `mod_` 337, `work_` 55, `vendor` 12). Each is a tree plus commands plus
assertions:

```
env GO111MODULE=off
go list -e -test -deps -f '{{.Error}}' p
stdout '^p[/\\]d_test.go:2:8: cannot find package "d" in any of:'
-- p/d.go --
package d
```

**237** of the 478 use no proxy or network verb, so their trees stand alone.

Running them as written would mean implementing the script DSL — not worth it. The cheap
version is to harvest the txtar trees and ignore the assertions: we have a real go command,
so ground truth comes from running `go list` live, and what the corpus supplies is the
expensive part — hundreds of adversarial trees somebody already thought hard about. Compare
`(PkgPath, GoFiles)` sets between the two strategies over each tree.

Caveats before anyone starts: many scripts set `GO111MODULE=off` (GOPATH mode, which we
deliberately do not support), and some contain deliberately broken code, so the harness has
to filter and report what it skipped rather than quietly pass.

## 5. Suggested order of attack — done

If we fix anything here, severity × likelihood says:

1. **D1** — stop `walkDirs` at a nested `go.mod`, and make `pkgPathFor` resolve against
   the *containing* module rather than always the main one. One change fixes both the set
   and the identity.
2. **D2** — read `go.work` in `findModule`, or at minimum refuse a workspace root with a
   clear error instead of emitting path-shaped identities.
3. **D3 + D6** — detect `vendor/modules.txt` and honour `-mod`/`GOFLAGS`; these interact.
4. **D4** — decide deliberately whether a bare pattern is a directory or an import path,
   and document it either way.

D5 and D7 are defensible as documented limitations.

Whatever we do, the corpus A/B cannot see any of it. A resolution-level differential
harness — same patterns, both strategies, compare `(PkgPath, GoFiles)` sets rather than
the emitted spec — is what would actually guard this.
