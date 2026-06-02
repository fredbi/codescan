# Idempotent Auto-Fixer Testing Pattern

**Extracted:** 2026-02-20
**Context:** Building auto-fix tools that modify source code (doc comments, linting, formatting)

## Problem

Auto-fixers that apply text replacements at byte positions can introduce new
issues on re-run if the fix output isn't properly handled by the same analysis
pipeline. Common failure modes:

- Fix introduces syntax that the analyzer re-matches (e.g. `[fmt]` not redacted
  on second pass, causing re-wrapping to `[[fmt]]`)
- Fix changes byte positions, causing overlapping or misaligned subsequent fixes
- Fix creates a pattern that a different analyzer in the pipeline picks up

## Solution

Always verify idempotency with a two-pass test:

1. **Pass 1**: Run fixer, apply all changes, record count
2. **Pass 2**: Run fixer again on the modified output
3. **Assert**: Pass 2 produces exactly 0 fixes

When idempotency fails, trace the issue by examining what the pipeline sees on
the second pass. The root cause is almost always that the first pass's output
isn't properly recognized/redacted by the input filters.

## Example

```go
// Pass 1: apply fixes
result1 := fixGodoc(pkg, rules, dryRun=false)
assert(result1.TotalFixes > 0)

// Pass 2: verify idempotency
result2 := fixGodoc(pkg, rules, dryRun=true)
assert(result2.TotalFixes == 0, "non-idempotent: %d fixes on second pass", result2.TotalFixes)
```

Common idempotency fixes:
- Extend input filters to recognize the output format (e.g. `[fmt]` bracket links)
- Use context-aware redaction (check symbol index before redacting ambiguous patterns)
- Ensure deduplication handles overlapping byte ranges from multiple analyzers

## When to Use

- Building any tool that modifies source files based on analysis
- Adding new fix rules to an existing auto-fixer pipeline
- After fixing bugs in fix application logic
