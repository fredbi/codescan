# Debug Filter Pipelines with Boundary Integration Tests

**Extracted:** 2026-02-24
**Context:** When investigating why a filter/redaction pipeline fails to prevent bad data from reaching a downstream consumer (e.g., spell-checker seeing redacted words).

## Problem

A length-preserving markdown filter was supposed to redact `@mentions` and bare URLs before hunspell spell-checking. Users reported false positives on GitHub usernames. Testing the filter alone showed it worked correctly. The bug was elsewhere in the pipeline, but the investigation technique is reusable.

## Solution

Don't just test the filter in isolation. Write a **boundary integration test** that:
1. Runs the filter on realistic input
2. Simulates the downstream consumer's tokenization
3. Checks that no "dangerous" tokens survive the boundary

```go
func TestFilter_NoHunspellTokens(t *testing.T) {
    f := New()
    in := `| @fredbi | 148 | https://github.com/example/commits?author=fredbi |`
    out := f.Filter(context.Background(), in)

    // Words that should be fully redacted
    dangerous := map[string]bool{"fredbi": true, "github": true, "author": true}

    // Simulate hunspell tokenization (split on non-letter boundaries)
    words := strings.FieldsFunc(out, func(r rune) bool {
        return !unicode.IsLetter(r) && r != '\''
    })

    for _, w := range words {
        if dangerous[strings.ToLower(w)] {
            t.Errorf("%q survived filtering: %s", w, out)
        }
    }
}
```

## Key Insight

When the boundary test passes but the real system still fails, the bug is NOT in the filter itself but in:
- How the pipeline is wired (filter not applied in some code path)
- Section splitting (text fragments parsed differently)
- A different code path (e.g., preflight vs. main analysis)

This eliminates the filter as a suspect and narrows the investigation.

## When to Use

- Any pipeline where data flows through filter -> analyzer stages
- When users report issues that "should have been filtered"
- When investigating false positives in spell-checkers, linters, or similar tools
- Especially useful with length-preserving redaction filters where offset alignment matters
