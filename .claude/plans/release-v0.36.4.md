> [!NOTE]
> Last revision: 2026-08-17 (rev 3 — compiled-deps default DECIDED: revert, Windows confirmed as its cause;
> the checklist becomes a private-repo dress rehearsal; `hack/browser` retirement decided)

# Release v0.36.4 — closing the tail

## Summary

Close v0.36.4 this week. The code is essentially done — 99 commits and ten PRs since `v0.36.3`
(2026-08-05) — so what is left is a tail of three threads plus a checklist that had scattered itself
across four other plans and would have been discovered at tag time:

1. 😇 **Vulns** — down to one decision, the `.codescan.yaml` write redirect. The terminal-escape half
   merged as PR #119.
2. ⚡ **The Windows perf problem**, which is almost certainly the compiled-dependencies default
   meeting a permanently cold cache. Likely outcome: revert the default to opt-in.
3. 🛠️ **goreleaser** for the binaries — the one thread with no prior write-up anywhere.

Plus a release-shape question worth deciding before the tag rather than after: **v0.36.4 currently
carries a new command and a default-behaviour flip**, which is not what a patch number promises.

## Context

`v0.36.3` was tagged at `c195e87c` on 2026-08-05. Everything below landed after it: the WASI build and
playground (#79), the on-demand scanner (#90), auto-detection (#94), compiled dependencies by default
(#99), TUI rounds 2 and 3 (#100, #108), the standalone CLI and the shared option surface (#111, #114),
the doc-site CLI pages (#115), the `fixtures/`→`testdata/` rename (#91) and the benchmark corpus (#118).

Where the three threads live today:

| thread | branch / worktree | state | plan before today |
|---|---|---|---|
| terminal escape sequences | `tui-esc-vuln` | ✅ **MERGED** PR #119 (`ef5c149b`) | `security.md` filed it as ♥️ someday, three days before it shipped |
| goreleaser | `goreleaser-cli-tools` (`.worktrees/ci/goreleaser-cli-tools`) | **empty — no commits** | none; requirements stranded in `wasi-round2.md` + `archive/genspec-cli.md` |
| Windows perf | `test/windows-cache` | 1 commit, a `ci-workflows` pin bump | none |

The Windows commit bumps all eleven workflow pins from `81c95ab8` to `69ea5ee4` — both labelled v0.6.0,
so the second is an unreleased SHA of the shared repo, colocalizing the build cache on the project's
drive. That framing (cache placement) and the benchmark numbers point the same way, which is why the
default flip and the Windows problem are treated as one item below rather than two.

## Trajectory

1. ✅ ⚡ **Compiled-dependencies default — DECIDED: revert.** Windows diagnosed and confirmed; the
   implementation is running in parallel
2. ⚠️ 😇 **Close the security items** — ✅ escapes merged (PR #119), ✅`hack/browser` retirement merged;
   the `.codescan.yaml` call is the last one
3. 📝 🛠️ **goreleaser** — what ships, built how, and what it must not foreclose
4. 📝 🛠️ **Dress-rehearse the release in a private repo** — the checklist becomes a rehearsal

## Actions

### 1. ⚡ The compiled-dependencies default, and Windows

1. ✅ **DECIDED 2026-08-17 (Fred): revert `SkipCompiledDependencies` to opt-in for v0.36.4.** Work
   started, in parallel with the rest of this plan. The three arguments that supported it, none of them
   about Windows alone:
   - `internal/benchmarks/README.md` prices the default at **−55 % warm, 6× slower cold, ~231 MB of
     build cache**. A CI runner is cold by construction, and a first-run user is too.
   - It is the *only* behaviour-affecting default change in the release. Reverting makes v0.36.4 an
     honest patch; keeping it makes the release a minor wearing a patch's number (see §4.1).
   - Nothing is lost: PR #99 established the option **costs no fidelity** — the same document either
     way. So this is purely a cost default, and cost defaults should favour the cold case.
2. ✅ **Windows diagnosis CONFIRMED (Fred, 2026-08-17).** No longer a hypothesis: the CI perf problem
   **is** this flag. So the revert is both the correct default and the fix, and the two threads that
   opened this week as separate problems turn out to be one.
3. 📝 **The revert has a documentation tail — four places, all of which currently state the wrong
   default.** Easy to ship the code change and leave these:
   - `internal/scanner/README.md#compiled-dependencies`
   - the doc-site options reference (`docs/doc-site/usage/options-reference.md`)
   - `.claude/CLAUDE.md`, which says "Unset is the default since v0.36.4 and costs no meaning"
   - the `Options.SkipCompiledDependencies` godoc in `internal/scanner/options.go:148`

   Keep the *measurement* wherever it appears — the −55 % warm / 6× cold / ~231 MB figures are the
   reason for the default, and they stay true. What changes is which way round the default sits.
4. 🔍 **What now happens to `test/windows-cache`?** It was the other candidate explanation, and the
   diagnosis did not need it. Two things stay true independently of the cause: colocating the build
   cache on the project drive is probably still worth having on its own merits, and the branch **pins
   an unreleased `ci-workflows` SHA** (`69ea5ee4`, labelled v0.6.0 but not the v0.6.0 tag) — fine for a
   test branch, not fine to merge. If it is kept, it needs a real v0.6.x tag first; if the revert makes
   it pointless, delete the branch rather than leaving it to look like pending work.

### 2. 😇 Security

1. ⚠️ **`.codescan.yaml` can redirect where `genspec` writes — Fred's call, still undecided.** Full
   analysis in `security.md` §Outstanding 1. The recommendation there stands: make `output` (and
   probably `input`) settable only from a config file the user *named* with `-config`, never from a
   discovered one — two entries in the existing `notConfigurable` table. It is small and it is inside
   the stated trust boundary. **This is the last item that can genuinely hold the release**, because
   shipping the CLI as a release binary is exactly what makes the finding reachable by strangers.
2. ✅ **`tui-esc-vuln` — MERGED**, PR #119 (`ef5c149b`). Shipped as one sanitizer with a two-mode API
   (`safetext.Sanitize` / `SanitizeBlock`) called from 14 sites across 7 files, plus five test suites.
   The one-choke-point question raised while it was uncommitted resolves in its favour — the rule was
   one implementation at the render boundary, and that is what landed.
3. 📝 😇 **Retire `hack/browser`** — decided 2026-08-17 (Fred): delete it, a remnant of earlier prototypes
   superseded by the playground. Not release-blocking, but it is repo surface a release publishes, and it
   drags two inbound references with it (`hack/doc-site/genspec-wasi/README.md`, `.github/dependabot.yaml`).
   Detail in `security.md` §Outstanding 3.
4. 📝 😇 📚 **Write `SECURITY.md`'s threat model** (`security.md` §Outstanding 2). It is 3–5 lines and it
   is the durable deliverable of the whole Aikido exercise; a release that ships binaries is the right
   moment for it to exist.
5. ⚠️ **Build the release binaries with go1.26.6 or later — HALF DONE, and the half that is done is not
   the half that ships.** `govulncheck` reported four findings — `GO-2026-6218` (net/url),
   `GO-2026-6090` (crypto/tls), `GO-2026-5972` (encoding/asn1), `GO-2026-5026` (x/net/idna) — **all
   standard library, all fixed in go1.26.6**, nothing in codescan's own code.

   `a6e00d67` (in PR #119) added `toolchain go1.26.6` to the **root `go.mod`**. But the artifacts
   goreleaser builds are the *command* modules, and as of `5e4cf04d`:

   | module | `go` | `toolchain` |
   |---|---|---|
   | `.` (library) | 1.25.0 | **go1.26.6** ✅ |
   | `cmd/genspec` | 1.25.0 | — |
   | `cmd/genspec-tui` | 1.25.0 | — |
   | `cmd/genspec-wasi` | — | — |
   | `go.work` | 1.25.0 | — |

   Built outside the workspace — which is what a release pipeline does — those three accept any
   toolchain ≥1.25.0 and will happily produce binaries carrying all four advisories.

   Two ways to close it, and they are not equivalent:
   - **Pin in the pipeline** (`setup-go` with an explicit version). Contained, no effect on CI's
     module matrix, and it is where §3.4 already put this. **Recommended** — and now cheap to verify,
     since §4.4 makes it a rehearsal step rather than a belief.
   - **Add `toolchain go1.26.6` to the three command modules.** Belt-and-braces, but 🔍 check the
     oldstable CI job first: `go.work` carries an explicit comment that the `go` directive is held at
     1.25.0 precisely so the floor does not "take the oldstable CI job out". A `toolchain` line is a
     different directive from `go` and should be fine — the root module already carries one — but that
     is worth confirming rather than assuming.

### 3. 🛠️ goreleaser

> The empty worktree. Everything here is new material — the only prior art is `wasi-round2.md` C10–C12,
> which parked release wiring on the grounds that "that stream is led elsewhere". This is elsewhere.

1. 🔍 **Decide the artifact set first, because it decides the rest.** The obvious three are `genspec`,
   `genspec-tui` and `genspec-wasi`. Each is **its own Go module**, which is the first real constraint:
   goreleaser is normally single-module, so this is either a monorepo configuration or one config per
   command.
2. ✅ **The `require` bump is handled** (Fred, 2026-08-17): `cmd/genspec` and `cmd/genspec-tui` both
   carry `replace github.com/go-openapi/codescan => ../..`, and the CI release pipeline
   (`bump-release-monorepo.yml`) does the `go mod` update. Not a manual step.
   ⚠️ **But it is the rehearsal's first real test**, because the two halves pull opposite ways:
   `go help install` states that `go install pkg@version` requires the target module's go.mod to hold
   **no `replace` or `exclude` directives**. The directive that makes the command build in-tree is
   precisely the one `@latest` refuses — so the pipeline has to strip or rewrite it, and nothing has
   proved it does. See §4.5.
3. 📝 **Do not foreclose the WASI artifact.** `wasi-round2.md` C10/C11/C13 ride this work: a
   `wasip1/wasm` artifact in the pipeline, and **publishing the export-data artefact** regenerated when
   the toolchain moves. They do not have to ship in v0.36.4, but the pipeline should have a place for
   them — retrofitting a second artifact class later is the expensive version. Note the knock-on
   recorded there: once the wasm artifact is published, `update-doc.yml` should *fetch* the pack rather
   than build it, which drops Go from the documentation build, and it is the precondition for ever
   splitting the front-end into `codescan-js`.
4. 📝 **Pin the build toolchain in the pipeline** (≥go1.26.6, per §2.4) rather than inheriting the
   runner's. A released binary's stdlib version is a supply-chain fact about the artifact.
5. 🔍 **Decide what "released" means for `genspec-wasi`** — a GitHub release asset, or nothing yet. The
   doc site currently builds the pack in CI, so nothing is broken if it stays unpublished; but C13 (how
   the stdlib reaches the browser: 8.4 MB embedded vs 3.7 MB + a cacheable 4.2 MB fetch) becomes much
   easier to answer once there is a published artifact to fetch.

### 4. 🛠️ The dress rehearsal

> **Approach settled 2026-08-17 (Fred).** Rather than run the checklist for real and discover its gaps
> at tag time, copy the whole codebase into a **private repo** and tag-and-release it there, repeatedly,
> until we are convinced **the new shared release workflow can support goreleaser**. The checklist below
> becomes the rehearsal's script rather than a list to tick once.

This is the right call for a reason worth writing down: **a release is the one workflow you cannot
practise in place.** Tags are cheap to create and expensive to retract, `go install` resolution is
sticky, and half of what this release changes — a new command module, a shared package pair, a
multi-module tag set — only misbehaves *at* the tag. A throwaway repo makes the whole thing rehearsable
and every failure free.

**What the rehearsal has to prove**, in rough dependency order:

1. ⚠️ **The multi-module tag set works at all.** `cmd/genspec-tui`, `docs/examples` and `testdata` each
   carry their own tags in the v0.36.3 series, `cmd/genspec` is new, and **`testdata` is a rename of
   `fixtures` as of PR #91** — a module changing name across a tag boundary is exactly the case nobody
   has run.
2. ⚠️ **The `replace` → `require` rewrite.** The pipeline claims this; nobody has watched it happen
   against a real tag. Two ways it can go wrong and both are silent: the `replace` survives into the
   published module (and `go install @latest` refuses it outright — see §3.2), or the `require` lands
   pointing at the previous tag. Item 5 is what catches either.
3. 📝 **goreleaser against three separate Go modules** (§3.1) — one config or three, and whether the
   shared workflow's shape permits either.
4. 📝 **The toolchain floor holds** (§2.4): build in the rehearsal repo *outside* any workspace and
   `govulncheck` the resulting binaries. This is the check that catches the command modules' missing
   `toolchain` line for real rather than by reading go.mod.
5. 📝 **`go install …/cmd/genspec@latest` resolves** from the rehearsal repo. The end-to-end proof, and
   the thing that is currently impossible.

**Two cautions for setting the rehearsal up**, both of which would quietly invalidate it:

- ⚠️ **`GOWORK=off`, everywhere.** The workspace substitutes the working tree for the released library,
  so a rehearsal run inside it would resolve the very `require` bump it is supposed to be testing. This
  repo has been bitten by `go.work` three times now (the corpus, the history benchmarks, `./...`
  scanning) — assume it bites here too.
- 🔍 **A private repo has a different module path**, so every `require`, every import in the command
  modules and `go.work` itself point at `github.com/go-openapi/codescan`. Decide up front whether the
  rehearsal rewrites those (testing the real mechanics under a fake name) or carries `replace`
  directives (easier, but then it is not testing resolution — which is item 5, the main event).

**Version number — settled by §1.** With the default reverted, v0.36.4 carries no behaviour change and a
patch is honest. Keeping the flip would have made it a minor wearing a patch's number and pushed it to
v0.37.0, which is already promised to the internal document model.

### 5. ⛔ Release notes — not this plan's work

**Dropped (Fred, 2026-08-17).** Release notes are generated by **git-cliff** from the commit history.
Changing that process is wanted eventually but is an **org-wide decision**, not codescan's to make
unilaterally — so nothing here proposes hand-written notes.

The practical consequence, worth stating once: **the commit titles are the release notes.** That is
already the house standard, and it means the one thing this release needs from the notes — that any
surviving default change is legible to a reader — has to be carried by the commit that makes the
change, not added afterwards. If §1.1 reverts the compiled-dependencies default, that revert's title
and body are where a user meets it.

## Achievements

Nothing yet — plan opened 2026-08-17.

## Appendix — risks and open points

- **The Windows diagnosis is a hypothesis.** Two plausible causes (cold-cache compiled dependencies;
  cross-drive cache I/O) and a branch that addresses only the second. §1.2 is the cheap discriminator
  and should run before either change is called the fix.
- **goreleaser has no prior art here and three modules to serve.** It is the least-specified thread of
  the three and the only one that can quietly ship a wrong artifact — an unbuildable `cmd/genspec` (if
  §3.2 is missed) or a binary carrying four known stdlib vulns (if §2.4 is missed). Both are ordering
  mistakes rather than hard problems, which is exactly the kind that gets made at tag time.
- **`wasi-round2` is still empty while its gate has cleared.** Only `feat/scrambler` remains unmerged
  of everything it was waiting for. If the goreleaser work lands without a slot for the wasm artifact,
  that stream gets harder rather than merely later.
- **The escape-sequence work is uncommitted in a worktree.** Uncommitted work is the state most easily
  lost. Committing it — even as WIP on its branch — costs nothing and should not wait for the design
  question in §2.2.
