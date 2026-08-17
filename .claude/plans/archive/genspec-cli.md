> [!NOTE]
> **✅ ALMOST COMPLETE** — archived 2026-08-17. Merged (PR #111 + #114). Six follow-ups outlived it and now live in [`backlog.md`](../backlog.md) §1 — wiring `genspec-wasi` to `cliconf` (scheduled v0.37.0), the `-color=always` redirect defect, and three enhancements. The release `require` bump is **closed**: both command modules carry a `replace` and the CI release pipeline does the `go mod` update. Nothing here is being worked.
>
> _Original document follows unchanged._

> [!NOTE]
> Last revision: 2026-08-16 — ✅ **MERGED** (PR #111). What is left is listed under Open, which gained the
> `-color=always` quirk found while covering the internal packages.

# genspec — standalone CLI

## Summary

Ship `cmd/genspec`: the ordinary, native command-line way to turn annotated Go source into a Swagger 2.0 document.
Minimal by construction — stdlib `flag`, no cobra/viper — but free to take the few dependencies that make a CLI pleasant:
a colored `slog` handler for diagnostics (as go-swagger does), YAML output, an `-input` spec to merge into, and
optional validation of what it produced.

It lives in its own module (like `cmd/genspec-tui`) so the library's go.mod stays lean and `cmd/genspec-wasi` stays
dependency-free and wasm-capable. The flag→`codescan.Options` mapping — today written out three times, guarded twice,
covering the option surface fully in neither place — moves to a shared `internal/cliopts` package.

A config file supplementing the flags (koanf, not viper) is round 2, deliberately out of round 1.

## Context

The trigger for a plan rather than a dash: **`cmd/genspec-wasi` is already most of this CLI.** It has stdlib flags, a
drift-guarded table of every boolean `Options` knob, `-workdir` / `-output` / `-format` / `-indent` / `-quiet`, and
diagnostics on stderr. What it does *not* have is precisely the dependency-bearing half — colored logging, YAML, spec
merge, validation — and it cannot have it: it must cross-compile to `wasip1/wasm` with no dependency beyond the library,
because the doc-site playground ships it.

So this is not a green field. Decisions settled with Fred, 2026-08-14:

1. **Sibling module, shared flag table.** New `cmd/genspec` with its own go.mod; `genspec-wasi` stays lean and keeps its
   role; the option surface is declared once in a root-module `internal/cliopts` and consumed by both.
2. **Binary name `genspec`** — the family reads as one set: `genspec`, `genspec-tui`, `genspec-wasi`.
3. **Round 1 scope** beyond full option coverage + colored diagnostics: YAML output, `-input` overlay, spec validation.
   The `-format=json` provenance envelope stays a `genspec-wasi` speciality (it exists to feed the playground).

Two facts verified before planning, both load-bearing:

- A submodule under `codescan/` **can** import the root module's `internal/` — the internal rule is path-prefixed, and
  `cmd/genspec-tui` already imports `codescan/internal/parsers/grammar`. So `internal/cliopts` serves all three commands
  with no public API widening. (Also recorded in memory `feedback_verify_module_assumptions`.)
- `github.com/SladkyCitron/slogcolor v1.9.0` — what go-swagger uses in `swagger generate spec` — resolves from the local
  module cache, so round 1 needs no network.

### Why the shared table is the first step, not a refactor to do later

The three existing mappings each cover a different slice of `codescan.Options`, and each is guarded against a different
notion of "complete":

| Site | Covers | Guard |
|------|--------|-------|
| `cmd/genspec-wasi/options.go` | every `bool` field | `TestFlagsCoverEveryBoolOption` (bools only) |
| `cmd/genspec-tui/main.go` | a hand-picked subset incl. strings/slices | `TestFlags_CoverEveryValueTypedOption` (its own subset) |
| go-swagger `cmd/swagger/.../spec.go` | most of everything, via go-flags struct tags | none |

Neither guard would notice a new `[]string` option landing unreachable from `genspec-wasi`, and `genspec-wasi` today
genuinely cannot set `-include` / `-exclude` / `-include-tags` / `NameFromTags` / `NameConcatBudget`. Writing a fourth
mapping by hand would make that four partial surfaces and three partial guards. One table, one guard, covering every
value-typed field of `Options`, is the thing that makes "the CLI exposes everything" true and keeps it true.

## Trajectory

1. ✅ **`internal/cliopts` — the option surface, declared once** (root module, no new deps)
   1. ✅ One table over every value-typed `Options` field: bool, string, `[]string`, float64
   2. ⛔ Group selection, so a command registers a subset — **dropped**, see Appendix 5
   3. ✅ One drift guard covering the whole of `Options`, with an excuse table carrying a reason per exclusion
   4. 🏁 ✅ Migrate `genspec-wasi` onto it — surface **widened** rather than held identical (Appendix 5)
2. ✅ **`cmd/genspec` — the command** (new module)
   1. ✅ Module skeleton, `go.work` entry, `run(argv, stdout, stderr) error` shape so the whole command is testable
   2. ✅ Full option coverage through `cliopts`, positional package patterns, `-workdir`
   3. ✅ Output: `-output` (`-` = stdout), `-format json|yaml|auto` inferring from the output extension, `-compact`
   4. ✅ `-version` from `runtime/debug.ReadBuildInfo`, and a usage block naming what is not a flag
3. ✅ **Colored diagnostics** — the reason this module exists apart from `genspec-wasi`
   1. ✅ `slog` + `slogcolor` handler on stderr; `-color auto|always|never`, `auto` honouring `NO_COLOR` / `TERM=dumb`
   2. ✅ Severity→level mapping, hints muted unless `-verbose`; positions rendered relative to `-workdir`
   3. ✅ End-of-run summary (counts by severity, muted ones included) and a `-fail-on` policy deciding the exit status
4. ✅ **`-input` overlay** — merge discoveries onto an existing document, via `go-openapi/loads`
5. ✅ **`-validate`** — `go-openapi/validate` on the produced document, findings through the same reporter, in the
   exit status
6. 📚 🏁 ⏳ **Docs & tests**
   1. 🏁 ✅ Command-level tests driving the real entry point: document, formats, overlay, validation, every exit status
   2. 📚 ✅ `cmd/genspec/README.md` and a root README section — 📝 doc-site page still to write
   3. 😇 ✅ Lint clean, 84.9% coverage, CI covered by the workspace (`go test work ./...`)
7. ✅ **Round 2 — config file** (koanf)
   1. ✅ `internal/cliconf`: discovery (upwards from cwd), precedence, applying onto the flag set, unknown-key
      rejection — **dependency-free**, so `genspec-wasi` and the TUI can be wired to it later
   2. ✅ Sections declared per flag in `cliopts`, guarded like the flag names are
   3. ✅ `cmd/genspec` reads it through koanf; `-config <path>` / `-c <path>` / `--no-config`; reported under `-verbose`
   4. 📝 Wire `genspec-wasi` and `genspec-tui` to it (the reason it is shared; not done in this round)

## Actions

### Open

1. 📝 **Wire `genspec-wasi` to `cliconf`** — **gate CLEARED 2026-08-16**: the TUI's factorization merged as
   PR #114 (`52d7de00`), so the shape wasi was told to wait for now exists and this is ordinary work.
   Original framing (Fred, 2026-08-14): wasi is the
   easy consumer: it registers the whole shared surface already and would just call `cliconf.Parse`, no koanf. The
   TUI is the one that tests whether these packages generalize, so it goes first and wasi follows whatever shape
   comes out of it. Doing wasi now risks doing it twice.
   When the TUI lands, wasi needs sections for its own flags (`format`, `output`, `indent`, `quiet`, `export-data`)
   and the same addressability guard
2. 🔍 **Bump the `require` on the library** once codescan releases `internal/cliopts` + `internal/cliconf`, so
   `go install .../cmd/genspec@latest` resolves. Until then the module builds through the workspace. **Release
   checklist item** — it is the one thing standing between this and an installable command
3. ✅ 📚 **Doc-site page** for `genspec` — **DONE** 2026-08-16 (PR #115, `bf1d4f37`): landed as
   `getting-started/usage-as-a-headless-cli.md` plus the whole `usage/` section rebuilt around
   one-knob-three-spellings. Plan: `doc-site-cli.md`
4. 📝 **`-color=always` cannot colour a run whose stdout is redirected** — super-minor, deferred by Fred
   (2026-08-16), recorded so it is not re-discovered. `resolveColor` answers correctly and `logger` sets
   `opts.NoColor = false`, but `slogcolor` renders through `fatih/color`, whose package-level `NoColor` is decided
   **once, at init, from whether `os.Stdout` is a terminal** — and diagnostics go to *stderr*. So the ordinary
   `genspec ./... > spec.json` prints uncoloured diagnostics to a terminal that could show them, whatever `-color`
   says. Probed directly: forcing the global off makes the same call emit escape codes.
   **Fix:** `color.NoColor = !colorize` in `diagnostics.logger` — one line, but a write to a process-global shared
   with anything else linking `fatih/color`, which is why it wants a deliberate decision rather than a drive-by.
   Note this also softens the tty-detection note under Round 1 §P2–P5: "escape codes in a file" is not currently
   reachable, since the colour library refuses a non-terminal stdout before our own check is consulted.
   Covered as far as it can be: `logger_test.go` asserts only the refusal half and says why.
5. ♥️ Environment variables (`GENSPEC_*`) — koanf's env provider onto the same seam, ~10 lines, nobody has asked
6. ♥️ Further short aliases (`-o`, `-q`) — `-c` earned one by being typed often; the rest are not obviously worth
   it, since `flag` lists an alias as its own help entry

### Handled elsewhere

- **`cmd/genspec-tui` onto the shared options model** — Fred reported 2026-08-14 that the TUI is aligning to it on
  its own. Not this plan's work: do not start it here, and expect its hand-written field→flag map in `main_test.go`
  to be retired by that stream rather than by this one.

  **It is also the design review this track has not had yet.** `genspec` and `genspec-wasi` both register the whole
  surface, so nothing has pushed back on the shape. The TUI will, and these are the signals worth watching:

  - it exposes a **subset** by choice — if honouring that needs anything beyond "register everything", the group
    selection dropped in P1 (Appendix 5) comes back, and it should come back *once*, shaped by a real consumer
  - it carries flags that are **not scan options at all** (`-profile`, `-profile-dir`, `-mem-profile-rate`, and
    `-packages`, which it spells as a flag where the others take positional patterns) — a second command with a
    private vocabulary is what tells us whether `Schema` merging is enough
  - its **defaults differ** in places (`-toolchain-free-loader` as a bool rather than the tri-state `-loader`)
  - a `.codescan.yaml` shared between the TUI and `genspec` is the first real test of "unknown section is skipped,
    unknown key in a known section is an error"

  If any of those needs `cliopts`/`cliconf` re-cut, that is the moment — cheaper with two consumers than three

### Settled while building (decisions worth knowing)

- **Exit statuses**: `0` success · `1` scan failed · `2` usage · `3` `-fail-on` tripped · `4` `-validate` found the
  document invalid. Verified end to end; `-validate` outranks `-fail-on` as the more specific answer
- **`-fail-on` covers validation findings too**, not just scan diagnostics: they reach the reader as one stream, in
  one set of colours, so a threshold seeing half of it is a trap rather than a policy
- **`-fail-on` defaults to `never`** — warnings are the ordinary case, and failing a build over one teaches people
  to stop reading them
- **`-compact`, not `-indent`** — follows `swagger generate spec`, whose users are the ones migrating
- **Validation runs after the document is written**, so an invalid one is still there to look at
- **Config keys are `<section>.<flag-name>`**, leaf spelled exactly as the flag — sections group, they do not
  rename. `output.output` is why the command's sections are `document` / `diagnostics` rather than
  `output` / `diagnostics`
- **Discovery searches upwards from the working directory**, not from `-workdir`: the file may set `-workdir`, and
  a file found through the value it sets would be reasoning in a circle
- **One filename for the whole family** (`.codescan.yaml`), not one per command — unknown *sections* are skipped
  and reported, unknown *keys in a known section* are errors
- **koanf gets the bytes, not the path**: its file provider carries a filesystem watcher for a file read once
- **`--no-config` replaced the `-config off` value** (Fred's call to add the switch; dropping `off` keeps one way to
  say each thing). `-c`/`-config` are stored apart so naming two different files is refused, not resolved by parse
  order; naming a file *and* `--no-config` is a usage error rather than a coin toss

### Round 1 — done

1. ✅ **P1 — `internal/cliopts`** (`af2d59b`)
   - Table entries carry: flag name, default, help, and a setter (`func(*codescan.Options) *T`), the shape
     `genspec-wasi` already proved — coverage is decided by *which field the setter writes*, never by mangling a field
     name into a flag name (`SkipJSONifyInterfaceMethods` is not mechanically derivable)
   - Flag names stay kebab-case of the field name, no exceptions, so a caller never has to guess a shorter spelling
   - Preserve the three-way `NameFromTags` contract: unset ⇒ nil (`["json"]`), `-name-from-tags=` ⇒ empty non-nil
     (Go field name), otherwise the list. `flag.Visit` is what distinguishes "unset" from "set to empty"
   - Excuse table entries needed for: `FS`, `ExportData`, `InputSpec`, `OnDiagnostic`, `OnProvenance` (not
     flag-shaped), `Debug` + `DescWithRef` (deprecated), `ToolchainFreeLoader` (reached through the tri-state `-loader`)
   - The tri-state `-loader go|own|auto` moves into `cliopts` too — `genspec` wants the same surface, and it is where
     the `ToolchainFreeLoader` excuse is discharged
   - ⛔ "`genspec-wasi`'s registered set must not change" — abandoned deliberately. Honouring it needed a per-flag
     exclusion list, which defeats the shared table. Its surface went 29 → 38 flags: 9 gained, **none renamed, none
     dropped**, default invocation byte-identical (Appendix 5)
2. ✅ **P2–P5 — the command** (`b97c29a`)
   - `run(argv, stdout, stderr) error` + `flag.ContinueOnError`, so tests drive the real entry point and `flag.ErrHelp`
     exits 0 (the `genspec-wasi` shape, which is right)
   - tty detection is `os.File.Stat` + `ModeCharDevice`, not `mattn/go-isatty` — the stdlib already answers it, and the
     cost of being wrong is escape codes in a file, which is what `-color=never` is for
   - Diagnostics follow go-swagger's handler shape (severity → `slog` level, `file`/`line`/`column` attrs, hints
     behind a switch) rather than inventing a second dialect. Timestamp and caller-file are off: a scan takes a second,
     and the position that matters travels as an attribute
   - `-input` rejects a directory as a usage error; `-validate` reads findings off the **first** validator result,
     which holds both lists — the second is a warnings-only view carrying them as its *errors*, and reading
     `.Warnings` there reports a document with warnings as clean
3. 📚 🏁 ✅ **P6 — docs & tests** (in `b97c29a`)
4. ⛔ **Widen `genspec-wasi` separately** — folded into P1 instead, for the reason above

## Achievements

### Round 1 — complete, awaiting review ⭐⭐

Two commits on `feat/genspec-cli`, nothing pushed:

1. ✅ `af2d59b` **`internal/cliopts`** — the option surface declared once ⭐⭐
   - One table over every value-typed `Options` field + one coverage guard, keyed by the setter
   - **The guard was proved to bite**: dropping the `emit-x-go-type` entry fails
     `TestFlagsCoverEveryValueTypedOption` by name, with the instruction to add an entry or an excuse
   - `genspec-wasi` migrated: 29 → 38 flags, none renamed or dropped; petstore output **byte-identical** to master
     on an unchanged invocation; `wasip1/wasm` and `js/wasm` still cross-compile
2. ✅ `b97c29a` **`cmd/genspec`** — the command ⭐⭐
   - Its own module; YAML / `-input` / `-validate` / colored diagnostics / `-fail-on` / exit statuses
   - Every exit status verified against a built binary, including `4` on a real invalid document
   - 84.9% coverage, lint clean, whole workspace green

### Merged

✅ **PR #111**, 2026-08-14 — 7 commits, rebased onto master as `6231aadc..2198496b`. Everything below is in master.

### Round 2 — config file, complete ⭐⭐

3. ✅ `e68ae489` **`internal/cliconf`** — the configuration-file contract, dependency-free ⭐⭐
   - Discovery upwards from cwd, precedence via `flag.Visit`, values applied through `flag.Set` so the file is
     parsed by the same code as the command line
   - `cliconf.YAML` satisfies koanf's parser interface **structurally** — the trick that lets the shared package
     stay importable from a wasm binary while `genspec` uses koanf
   - Sections declared per flag in `cliopts`, with a guard: a flag addressable from no key is one nobody can
     configure. 98.1% / 99.0% coverage
4. ✅ `5932aa54` **`cmd/genspec` reads it** ⭐⭐
   - `document` / `diagnostics` sections for the command's own flags; `-config <path>` / `-config off`
   - Verified by hand end to end: a `.codescan.yaml` two directories up, 6 keys applied, a `tui:` section skipped
     and reported, YAML out, validation run
   - Existing command tests now pass `-config=off`, so a stray file above the checkout cannot change what they mean

### Base camp

- ✅ Branch `feat/genspec-cli`, worktree `.worktrees/feat/genspec-cli` (2026-08-14)
- ✅ Design settled with Fred before any code: module layout, binary name, round-1 scope (2026-08-14)

## Appendix — risks and open points

1. **Three commands, one table, different audiences.** `cliopts` has to serve a dep-free wasm binary, a TUI with a
   deliberately narrow surface, and a full-fat CLI. If group selection starts sprouting per-command special cases, the
   shared table has stopped paying for itself — that is the signal to stop and re-cut it.
2. **`genspec` vs `swagger generate spec`.** go-swagger already exposes this exact functionality over the same library.
   `genspec` earns its place by being installable alone and by exposing options go-swagger has not surfaced (loader,
   GOOS/GOARCH/GOWORK, export data). Worth keeping the two flag vocabularies from drifting apart pointlessly — where a
   go-swagger flag name is already established and good, reuse it.
3. **Exit codes and `-fail-on` are policy**, and policy is what CI pipelines end up depending on. Settle before P3
   ships, not after.
4. **YAML fidelity.** go-swagger round-trips JSON → `any` → YAML. Key order is lost that way, which is exactly the
   ordered-map problem parked project-wide (memory `project_ordering_wontfix_boundary`). Round 1 does the same and says
   so, in `cmd/genspec/README.md` and at the function; it is not a regression, it is the known boundary.
5. **Group selection dropped, and `genspec-wasi` widened instead.** The plan wanted each command to register a subset,
   with `genspec-wasi`'s surface held identical so the migration was provably pure. Neither survived contact: the group
   boundaries do not line up with what `genspec-wasi` registers (it has `exclude-deps` but not `include`/`exclude`,
   `goos`/`goarch` but not `gowork`), so a subset would have needed a per-flag exclusion list — the exact thing the
   shared table exists to abolish. So the table registers everything, and `genspec-wasi` gained the 9 options it could
   not reach. The A/B stands in for the purity argument: no flag renamed, none dropped, output byte-identical.
6. **The module cannot be `go install`ed until codescan releases `internal/cliopts`.** Normal monorepo release
   ordering — `cmd/genspec-tui` lives with the same arrangement — and CI tests through the workspace, so nothing is
   blocked. It is a release-checklist item, not a design problem, recorded under Open.
