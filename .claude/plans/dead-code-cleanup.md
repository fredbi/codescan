# Bucket D — dead / unreachable code cleanup plan

Captured from the R2 + R3 coverage sweep (baseline @ `7e5296640`, final
coverage **84.0%**). Every item below is either fully unreached
post-sweep (0% coverage with full test matrix), or structurally
unreachable given the call graph and go/types guarantees (reached only
when preceding dispatchers already exhaust the case).

This is intended as the shopping list for a **cleanup commit on the
refactor branch**, after porting the R2+R3 tests. Each item names the
baseline file:line; the refactor-branch equivalent may live in a
different package — check before deleting.

**Process rule:** delete one symbol per commit, run `go vet ./... && go build ./... && go test ./...`
between each. If anything breaks, the symbol wasn't dead; investigate
before continuing.

## D.1 — Fully dead symbols (0% coverage after full sweep)

These have no reachable call site. Safe to remove.

| Symbol | File | Line | Notes |
|---|---|---|---|
| `testError.Error` | `assertions.go` | 13 | Constant `errInternal` is only ever `panic`'d, not compared. Keep the constant, drop the `Error()` method or drop both. |
| `scanCtx.PkgForPath` | `application.go` | 350 | **Refactor already relocated this to `export_test.go`.** Baseline can delete it outright in the cleanup commit. |
| `setupRefParamTaggers` | `taggers.go` | 65 | **Structurally unreachable.** Fires only when `ps.Ref != ""` inside `processParamField`, but nothing on the code path leading there sets `ps.Ref` (it only sets `ps.Schema().Ref`). Confirmed 0% after adding R2b non-transparent-alias fixtures that *would* have exercised it if reachable. |
| `paramTypable.Level` | `parameters.go` | 21 | Return `0`. Never called; the sole consumer reads `.Level()` on the `schemaTypable` wrapper instead. |
| `paramTypable.SetRef` | `parameters.go` | 27 | Body params go through `schemaTypable.SetRef`; non-body params never have a Ref. |
| `paramTypable.AddExtension` | `parameters.go` | 55 | `addExtension()` free function is used directly; this helper has no callers. |
| `itemsTypable.In` | `parameters.go` | 80 | — |
| `itemsTypable.Level` | `parameters.go` | 82 | — |
| `itemsTypable.Schema` | `parameters.go` | 92 | — |
| `itemsTypable.AddExtension` | `parameters.go` | 104 | — |
| `itemsTypable.WithEnum` | `parameters.go` | 108 | — |
| `itemsTypable.WithEnumDescription` | `parameters.go` | 112 | — |
| `responseTypable.In` | `responses.go` | 21 | — |
| `responseTypable.SetSchema` | `responses.go` | 73 | `.Schema` is set directly via field access in callers. |
| `responseTypable.CollectionOf` | `responses.go` | 77 | `CollectionOf` helper not used; callers build schemas directly. |
| `responseTypable.AddExtension` | `responses.go` | 81 | — |
| `responseTypable.WithEnum` | `responses.go` | 85 | — |
| `responseTypable.WithEnumDescription` | `responses.go` | 89 | — |

Total: **16 symbols** + 1 already-relocated (`PkgForPath`).

**Action — consider shrinking the `swaggerTypable` interface:**
several interface methods (`Level`, `In`, `AddExtension`, `WithEnum`,
`WithEnumDescription`) have zero-coverage implementations on at least
two of the three concrete types. Evaluate whether the interface itself
can be trimmed, or whether the unused implementations should be left as
no-ops to keep the interface usable by callers we've missed.

## D.2 — Dead branches inside live functions

These are reached by exhaustive `switch` statements, but the preceding
dispatcher already catches the input so the branch is unreachable. The
fixtures we *added* to try to hit them (`top-level-kinds`,
`named-struct-tags`, `ref-alias-chain`) still produce correct output
because the live path handles the case; the branches are pure cruft.

### `buildFromDecl` non-Named/non-Alias cases
- **File:** `schema.go:208-225`
- **Dead cases:** `*types.Basic`, `*types.Struct`, `*types.Interface`, `*types.Array`, `*types.Slice`, `*types.Map`
- **Why dead:** `entityDecl.ObjType()` always returns a `*types.Named` for declared types and `*types.Alias` for alias decls — never the underlying type directly. The author already noted this with a `TODO(fredbi): we may safely remove all the cases here that are not Named or Alias` at line 209.
- **Action:** collapse the switch down to `Named` + `Alias` + `default: (error)`. Keeps the diagnostic fallback for future go/types changes.

### `buildNamedStruct` typeName branch
- **File:** `schema.go:562-565`
- **Why dead:** `buildNamedType` at line 432-436 already short-circuits on `typeName(cmt)` before dispatching to `buildNamedStruct`. By the time we reach line 562 the `typeName` tag has already been consumed.
- **Action:** delete the branch. Verified by R3a named-struct-tags fixture: the `Code LegacyCode` field (with `swagger:type string` on `LegacyCode`) is emitted as `{type: string}` via `buildNamedType`, never reaching `buildNamedStruct`.

### `buildNamedStruct` isStdTime branch
- **File:** `schema.go:552-555`
- **Why dead:** `buildNamedType` handles `isStdTime(tio)` at line 404 before dispatch. `time.Time` field references never reach `buildNamedStruct`.
- **Action:** delete the branch.

### `buildDeclAlias` isAny / isStdError / isStdTime branches on `o`
- **File:** `schema.go:617-630`
- **Why dead:** `o = tpe.Obj()` is the left-hand side of the alias declaration — i.e. the TypeName being defined in our own package (`Wildcard`, `Timestamp`, etc.). Checking `isAny(o)` / `isStdError(o)` / `isStdTime(o)` against that TypeName can never match, since you cannot redeclare `any` / `error` / `time.Time` in a user package. The reachable case is `isXxx(ro)` in the rhs switch at line 669 / 674, which is already covered.
- **Action:** delete lines 617-630. Verified by R2e `ref-alias-chain` fixture.

### `specBuilder.Build` error-propagation plumbing
- **File:** `spec.go:55-85`
- **Why dead:** every `if err := s.buildXxx(); err != nil { return nil, err }` line is reached only when an inner builder returns an error. Inner errors are all bucket-B malformed-input cases; without a way to produce an error mid-build the wrappers can't fire.
- **Action:** leave alone. These are harmless ~1-line guards; deleting them would remove meaningful error plumbing. Flag as "acceptable uncovered."

## D.3 — Language-guarded exhaustive cases (structurally unreachable)

Every `switch tpe.(type)` against `types.Type` across the schema/builder
layers carries trailing cases for `*types.Union`, `*types.TypeParam`,
`*types.Chan`, `*types.Signature`, and a `default` warning. These
cannot fire because:

- `*types.Union` — only exists inside generic type-set constraints, not in declared field types.
- `*types.TypeParam` — method signatures that take type parameters are rejected by `ExplicitMethods()` / `EmbeddedTypes()` iteration, or would have been unified away by `types.Unalias`/`.Underlying()` before reaching the switch.
- `*types.Chan`, `*types.Signature` — fields cannot directly be `chan T` / `func(...)` in a `swagger:model` struct (we'd skip them earlier) and embedded types of these kinds are not allowed by Go.

**Affected functions:**
- `schema.go:1395-1413` `buildEmbedded`: Union, TypeParam, Chan, Signature, default
- `schema.go:1453-1464` `buildNamedEmbedded`: Union, TypeParam, Chan, Signature
- `schema.go:1366-1383` `buildNamedAllOf`: TypeParam, Chan, Signature, default
- `schema.go:234-245` `buildFromDecl`: TypeParam, Chan, Signature, default
- `schema.go:472-488` `buildNamedType`: TypeParam, Chan, Signature, default
- `schema.go:1313-1324` `buildAllOf`: TypeParam, Chan, Signature (Pointer recursion IS reachable)

**Action — collapse, but route the default through a shared helper:**

The goal is not branch-exhaustiveness; it's *no silent fallthroughs*. `go/types` evolves across Go minor versions (the Go 1.26 `reflect.Type` iterator-methods incident is a live reminder), and the scanner runs on arbitrary user code in uncontrolled environments. A silent `default: return nil` hides both kinds of surprise.

Rules for each collapsed switch:

1. Factor the warning-message production into a single helper in `internal/logger` — e.g. `logger.UnsupportedTypeKind(where string, tpe types.Type)` — so every switch emits a uniform, greppable diagnostic.
2. Per-switch, decide explicitly: **warn-and-skip** or **panic**?
   - **Warn-and-skip** is the default for production paths. The tool must degrade gracefully when user code presents a shape we don't yet handle.
   - **Panic** only when the case is genuinely a programmer error on our side — e.g. reached after a `must*` precondition that already ruled the shape out. Reserve sparingly.
3. Collapsed form: `default: logger.UnsupportedTypeKind("buildNamedType", tpe); return nil` (or `panic(...)` per rule 2).

This still removes the ~60 lines of per-kind cases. The type-kind names live in `go/types` godoc; they don't need to be re-listed in our switch statements. But the single default must carry the diagnostic so future `go/types` additions surface one clear log line instead of disappearing.

## D.4 — Post-must assertion error returns

Every code path following a `mustNotBeABuiltinType(o)` or
`mustHaveRightHandSide(tpe)` call has a redundant `if err != nil` or
`FindModel not found` guard: the `must*` panic already fires on the
invalid precondition, so the subsequent guard can never trigger.

**Examples:**
- `schema.go:788-789` `buildAlias` — `FindModel` miss after `mustNotBeABuiltinType`.
- `schema.go:658-660`, `682-684` `buildDeclAlias` refAliases branches — same shape.
- `parameters.go:256-257`, `269-270`, `278-281` `buildAlias`, `buildFieldAlias` — same.
- `responses.go:336-337`, etc. — same.

**Action:** leave alone. These guards cost one line each and would be
needed if `must*` were ever downgraded to a `return err`. The
"defensive consistency" value outweighs the ~20 lines saved. Flag as
"acceptable uncovered."

## D.5 — `setXxx.Parse` empty-line and numeric-parse guards

Across `parser.go` and `meta.go`, every tag-parser `Parse(lines []string)`
begins with:

```go
if len(lines) == 0 || (len(lines) == 1 && len(lines[0]) == 0) {
    return nil
}
```

followed by a regex match, a `strconv.ParseX`, and `if err != nil { return err }`.

**R4 update — most of these are actually dead, not bucket B:**

- **Empty-line guard**: the tagger's `Matches(line)` gates entry into
  `Parse`. Every `rxXxxFmt` regex requires `name:\s*(.+)` — at least one
  character after the colon. So `Parse` is never invoked with empty
  lines. The empty-line guard is dead code.
- **`strconv.ParseFloat` / `strconv.ParseBool` error**: the same regexes
  gate the captured value. Numeric regexes require
  `\p{N}+` so only digits get through; boolean regexes require
  `(true|false)` verbatim. `ParseFloat`/`ParseBool` cannot fail on input
  the regex accepted. Dead code.

**Reachable exceptions** (moved to bucket B and covered in R4):
- `setDefault.Parse` / `setExample.Parse` — these use `rxDefaultFmt` /
  `rxExampleFmt` which are `name:\s*(.*)$` (permissive). Their call
  into `parseValueFromSchema` then runs `strconv.Atoi` / `ParseFloat`
  /etc., which **can** fail on non-numeric input. Covered by R4 fixtures
  `malformed/default-int/` and `malformed/example-int/`.
- `parseValueFromSchema` object nilerr fallback (925-927) — reachable
  with a non-JSON `default:` on a map-typed field, but doesn't actually
  raise an error (swallows with `//nolint:nilerr`). Not worth testing.

**Action:** reclassify the empty-line and ParseX error guards as **bucket
A** (defensive, regex-guarded). Can be dropped in the D.2 cleanup pass
alongside the other dead branches — ~40 more lines removable.

## Summary

| Category | Lines | Action |
|---|---|---|
| D.1 Fully-dead symbols | ~120 | Delete (16 symbols, one commit each) |
| D.2 Dead branches in live functions | ~50 | Delete; keep diagnostic default |
| D.3 Language-guarded exhaustive cases | ~60 | Collapse to single default per switch |
| D.4 Post-must error returns | ~20 | Leave (acceptable) |
| D.5 Empty-line / numeric-parse guards | ~90 | **Reclassified to A after R4** — also deletable |

Net removable: **~270 lines** across 16 symbols + ~10 switch
collapses + D.5 empty/parse guards. Expected coverage after cleanup:
**~92-94%** (the acceptable-uncovered remainder is D.4 + occasional true
bucket-B cases we haven't covered).

## Sequencing

1. **Land R2+R3 tests first** on the refactor branch (they're the safety net).
2. **Verify the tests that exercise D.1 / D.2 paths** — they must still be green after deletions. Specifically: `TestCoverage_AliasExpand`, `_AliasRef`, `_AliasResponseRef`, `_InterfaceMethods`, `_RefAliasChain`, `_TopLevelKinds`, `_NamedStructTags` each pin observable behaviour that the dead code doesn't affect.
3. **Execute D.1 deletions** in isolation, one commit per symbol.
4. **Execute D.2 branch deletions** next; each requires grep-verification that no other caller paths exist.
5. **Execute D.3 switch collapses** last; biggest blast radius, but purely cosmetic on the live path.
6. Leave D.4 and D.5 untouched.

**Do not delete if:** a symbol or branch is referenced by anything
outside the codescan package (public API). Run `go vet ./...` and `gopls find-references` on each before deletion.

---

## Post-mortem (pass #1, 2026-04-19)

Landed on branch `chore/remove-dead-code` in two commits: `chore: dead-code removal (pass #1)` and the D.3 helper/collapse commit. Final scope was **much smaller** than the plan predicted. Recording the gap so future cleanup passes don't re-litigate the same analysis.

### Outcome by bucket

| Bucket | Plan estimate | Actual | Status |
|---|---|---|---|
| D.1 | ~120 lines across 16 symbols | ~20 lines across 3 symbols | partial — see below |
| D.2 | ~50 lines, 3 dead-branch groups | ~50 lines across 3 groups | done as planned |
| D.3 | ~60 lines across 6 switches | ~60 lines + `logger.UnsupportedTypeKind` helper, semantics refined | done, with behavior changes |
| D.4 | leave | left | kept — contract enforcers |
| D.5 | ~90 lines "also deletable" | 0 lines | **skipped** — claim overstated, see below |

Net delivered: roughly **~130 lines removed**, versus the plan's projected ~270. The remainder is held up by architectural dependencies (interface contracts) or is not actually dead.

### D.1 — the interface-mandated block

**11 of the 16 "fully dead" symbols cannot be removed without a non-trivial refactor.** The plan's coverage analysis flagged methods as 0%-covered, but didn't check compile-time requirements. All `paramTypable` / `itemsTypable` / `responseTypable` methods listed (`Level`, `SetRef`, `AddExtension`, `In`, `WithEnum`, `WithEnumDescription`, plus the items/response equivalents) are mandated by `ifaces.SwaggerTypable` because those types are assigned to `SwaggerTypable`-typed variables at 4+ call sites. Removing any one method causes:

```
cannot use paramTypable{…} as ifaces.SwaggerTypable value: paramTypable does not implement ifaces.SwaggerTypable (missing method Level)
```

Each interface method also has live polymorphic call sites (e.g. `.AddExtension()` at `schema.go:205`, `.SetRef()` at `schema.go:1392`, `.Level()` at `parameters.go:194`), so the interface itself can't be shrunk either — any impl that dispatches through it must provide the full method set.

Removing these 11 items would require **interface segregation**: split `SwaggerTypable` into smaller interfaces so each concrete type implements only what it's actually used for. That's a real design change, not a cleanup commit.

**Decision: defer indefinitely.** The v2 grammar-parser rewrite (see `.claude/plans/ramblings/vision.md`) will replace the `Typable` / taggers dispatch entirely with a grammar → AST → emitter visitor pipeline. Investment in interface segregation now would be thrown away. Captured in memory `project_taggers_to_sunset.md`.

Two more items already didn't need work on the refactor:
- `scanCtx.PkgForPath` — already relocated to `export_test.go` during the package split.
- `testError.Error` — renamed to `internalError.Error`, now load-bearing: `ErrInternal` is wrapped via `%w` into panic errors, so the `Error()` method satisfies the `error` interface.

### D.2 — held up as predicted, but required guiding comments

All three dead-branch groups were removed cleanly (`buildFromDecl` 6 unreachable cases, `buildNamedStruct` isStdTime + TypeName, `buildDeclAlias` LHS short-circuits). Each survived the test suite unchanged.

Added guiding comments at each site per maintainer request. The common thread across all three deletions was not "Go forbids the shape" — it was "the sole caller already handled the case upstream". That invariant is easy to lose when someone later refactors the dispatcher; the comments pin it.

Takeaway for future passes: when removing a dead branch whose deadness depends on an upstream gate, leave a comment naming the gating caller and the precondition it enforces. Otherwise the invariant becomes tribal knowledge.

### D.3 — plan direction correct, two behavior refinements

The "collapse to single default per switch" direction was right, but with two refinements driven by the maintainer rules:

1. **No silent fallthroughs.** The single `default:` must route through a shared helper (`logger.UnsupportedTypeKind(where, tpe)`) so a future `go/types` evolution surfaces one uniform log line instead of disappearing. Memory: `feedback_go_types_defensive_guards.md`.
2. **Panic → warn-and-skip for two existing defaults.** `buildFromType` previously panicked on the default case; `buildAllOf` / `buildNamedAllOf` wrapped an error. Both were inconsistent with the explicit `TypeParam/Chan/Signature` cases right above them (which returned nil). Unified to warn-and-skip because the scanner runs on uncontrolled user code and a partial spec beats a crash.

Additional win: three `case *types.Union:` empty cases (three sites) that silently swallowed union constraints now log via the helper. Was hidden before.

### D.4 — kept deliberately

The `if err != nil` / `if !found` guards that sit right after `MustNotBeABuiltinType` / `MustHaveRightHandSide` are technically unreachable (the `must*` panics already fire on invalid preconditions), but they're kept as **contract enforcers**. Two reasons:

1. If `must*` is ever downgraded to `return err` (not unreasonable — panicking on user-input shape is the kind of thing a future robustness pass might revisit), the guards already exist and do the right thing.
2. The guard pattern is the repository's idiomatic "I'm checking the postcondition of the precondition-asserting call" and removing it would break reader pattern-matching.

Cost is ~1 line per site (~20 lines total). Value is documentation of intent + future-robustness insurance.

### D.5 — post-mortem of the "~90 lines also deletable" claim

**Verdict: skipped; the plan's claim did not hold.**

Empty-line guards (`if len(lines) == 0 || (len(lines) == 1 && len(lines[0]) == 0) { return nil }`):
- **Plan's theory:** dead because `SectionedParser` gates entry via `Matches`, which enforces a regex requiring `name: value`.
- **Reality:** pinned by unit tests in `validations_test.go` (lines 86, 341, 376, 455) that directly call `Parse(nil)`, `Parse([]string{})`, `Parse([]string{""})`. Bypassing `Matches` reaches the guard. Removing the guard converts those calls into `index out of range` panics at `lines[0]`.
- **Lesson:** coverage tools only see what tests actually exercise. A 0%-covered guard may still be pinned by tests that exercise the *tolerance* (input never triggers the guard, but the test proves absence of panic). Grep for `Parse(nil)` / `Parse([]string{` before deleting input guards.

Numeric `ParseFloat` / `ParseInt` error guards:
- **Plan's theory:** dead because regex captures only `\p{N}+`, so parse can't fail.
- **Reality:** `\p{N}+` is a character class, not a range constraint. A 20+ digit `MaxItems:` matches the regex but overflows `int64` → `ParseInt` returns `ErrRange`. Same for `Maximum:` with 310+ digit float overflow. Real user-input edge case.
- **Lesson:** "regex-gated input" proves *character class*, not *numeric magnitude*. Overflow is a separate failure mode that regex can't prevent.

`ParseBool` error guards:
- Genuinely dead (regex captures exact `(true|false)`, so `ParseBool` cannot fail).
- ~8 lines across 4 parsers. Left as API-stability insurance in case the regex is loosened later. Not worth the churn.

### Meta-lessons

1. **Coverage ≠ reachability.** Go's type system compiles in what the interface requires; 0% test coverage doesn't mean 0% compile-requirement.
2. **Tests pin contracts.** Dead-code analysis must grep tests for direct calls to the "dead" symbol before deleting. The empty-line guards looked dead but were contract-under-test.
3. **Regex guards are partial.** They constrain character classes, not semantics (numeric range, length limits, encoding validity). Error handling downstream may still be load-bearing.
4. **"One symbol per commit" is directional.** In practice, related deletions in the same file (e.g. all `schema.go` D.2 branches) were bundled logically. Bisectability survives as long as each commit is internally test-green.
5. **Guiding comments survive refactors better than invariants.** When deletion depends on an upstream gate, writing "the sole caller enforces X" in place of the deleted code pays dividends when someone later touches the gate.
6. **Defer what v2 will obsolete.** ~70 lines of potentially-removable interface cruft are staying because the grammar-parser rewrite will drop the whole dispatch. Don't refactor code you're about to delete.
