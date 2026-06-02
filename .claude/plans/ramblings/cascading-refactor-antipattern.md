# Cascading Refactor Antipattern: Lessons from the Phase 3 Failure

Date: 2026-03-23

## What happened

During the refactoring of `codescan` into internal subpackages, Phase 3 (extracting
`internal/parsers`) devolved into a "cascading mess" — a long loop of incremental sed
fixes that kept revealing new problems, eventually requiring a full revert after ~30
minutes of fruitless iteration.

The failure mode: each fix exposed 5-10 new issues of the same class, because the
underlying problem (unexported struct fields accessed via positional literals across
a package boundary) was systemic, not isolated. Fixing one instance didn't reduce
the remaining work — it just revealed the next batch.

## The root cause

Moving types to a separate Go package changes the visibility rules:

1. **Unexported fields become inaccessible** — `&setMaximum{builder, rx}` works within
   the same package but fails across packages because `builder` and `rx` are lowercase.

2. **Positional struct literals break** — even if fields are exported, Go requires
   keyed syntax (`Field: value`) for struct literals from external packages. The
   codebase had ~50+ positional literals across 6 consumer files.

3. **Type aliases don't help with methods** — `type X = pkg.Y` means you can't define
   new methods on X in the importing package. This blocked the scanner extraction
   until we moved the methods too.

These three constraints interact multiplicatively: each type move requires (a) exporting
fields, (b) converting all literals to named syntax, AND (c) moving all methods. Missing
any one of these for any one type causes compilation failures that cascade through every
consumer.

## How we could have detected it earlier

### Signal 1: Count the cross-package struct literal sites BEFORE starting

The exploration agent identified ~50+ struct literal construction sites across 6 files.
This should have been a **hard stop** — the rule should be:

> If moving types to a new package requires converting >10 struct literals from positional
> to named syntax, do that conversion as a separate committed step BEFORE any file moves.

We partially recognized this (the "preparation step" idea) but tried to do it inline
with the move rather than as a committed, verified intermediate state.

### Signal 2: The first sed that needed a second sed was the warning

When the first `sed` to convert positional literals missed cases (multi-line literals,
already-partially-converted ones, different field counts), that was the signal that
sed is the wrong tool for this job. The pattern is:

> If a mechanical transformation requires pattern matching on Go syntax (struct literals,
> method receivers, field names), sed will fail on edge cases. Use `gopls rename`,
> `gorename`, or write a small Go AST tool instead.

We should have switched tools after the first round of misses, not after the fifth.

### Signal 3: "Fix 5 errors, get 10 new ones" means the approach is wrong

The cascading pattern — fix compilation errors, get new ones of the same class — is
a clear signal that the change is not incremental. Each fix doesn't reduce the problem
space; it just reveals more of it. The rule:

> If after 2 rounds of "fix and rebuild" the error count is not strictly decreasing,
> STOP. Revert to green and redesign the approach.

We violated this by continuing for ~8 rounds, each time discovering new categories of
the same underlying issue (unexported fields, positional literals, missing bridges,
double-keyed literals from re-running sed on already-converted code).

### Signal 4: The bridge file grew too large

The `parsers_bridge.go` file ended up with ~130 lines of aliases and bridges. When a
bridge/adapter layer approaches the size of the code it's wrapping, it's a sign that
the extraction is premature — the consumer code hasn't been prepared for the separation.

> If the bridge file for a package extraction exceeds ~30 declarations, the consumers
> aren't ready. Do more preparation in-place first.

## What the correct approach looks like

### Step 1: Export fields (committed, verified, GREEN)

Convert all struct fields that will cross package boundaries from unexported to exported.
This is a pure refactoring within the single package — no package moves, no imports.

```go
// Before
type setMaximum struct {
    builder validationBuilder
    rx      *regexp.Regexp
}

// After
type setMaximum struct {
    Builder validationBuilder
    Rx      *regexp.Regexp
}
```

Commit this. Run tests. Verify green.

### Step 2: Convert positional literals to named syntax (committed, verified, GREEN)

Convert ALL struct literal construction sites from positional to named field syntax.
This is also a pure refactoring — no behavioral change.

```go
// Before
&setMaximum{schemaValidations{ps}, rxf(rxMaximumFmt, "")}

// After
&setMaximum{Builder: schemaValidations{ps}, Rx: rxf(rxMaximumFmt, "")}
```

Use `gopls` or a Go AST rewriter, NOT sed. Commit. Test. Green.

### Step 3: Move files (the actual extraction)

NOW the types can move to `internal/parsers/` with minimal friction:
- Fields are already exported ✓
- Literals already use named syntax ✓
- The only changes needed are: package declaration, import paths, type aliases

This step should be nearly mechanical and produce <5 compilation errors.

### Step 4: Clean up bridge file

Thin the bridge file by having consumers import from `internal/parsers` directly
(in later phases when they move to their own packages).

## The meta-lesson

**Large refactorings must be decomposed into steps where each step is independently
committable and verifiable.** The moment you're making changes that depend on other
uncommitted changes to compile, you've lost the ability to detect failures early.

Phase 2 (scanner extraction) succeeded because we did this naturally:
1. Export fields (committed) → 2. Move types (committed) → 3. Wire aliases (committed)

Phase 3 failed because we tried to do all three simultaneously:
- Exporting fields in the parsers package copy
- Moving files
- Converting struct literals
- Wiring bridges

...all in one uncommitted change. When it broke, there was no safe intermediate state
to fall back to.

## Checklist for future package extractions

Before moving ANY type to a new package:

- [ ] Count cross-package struct literal sites. If >10, do preparation step first.
- [ ] All fields accessed from outside are already exported (committed, green).
- [ ] All struct literals use named field syntax (committed, green).
- [ ] All methods that will move are self-contained (no references to symbols
      that will stay in the source package).
- [ ] The bridge file will have <30 declarations.
- [ ] Each step produces a compilable, testable intermediate state.

If any of these are not met, do preparation steps until they are. Don't start
the move.
