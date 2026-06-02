# Test Comments to Explicit Subtests

**Extracted:** 2026-01-29
**Context:** Go tests with inline comments describing assertions (e.g., "// Should contain...", "// Should NOT have...")

## Problem

Inline comments like `// Should contain the function call` or `// Should NOT have outer braces` describe test intent but are invisible in `go test -v` output. When a test fails, the developer sees only the error message, not the intent. Comments are also easy to drift out of sync with the actual assertion.

## Solution

Replace `// Should ...` comments with `t.Run("should ...", ...)` subtests. This makes the intent visible in test output, individually addressable with `-run`, and self-documenting.

## Example

Before:
```go
func TestParse_Render(t *testing.T) {
    rendered := getRendered(t)

    // Should contain the function call.
    if !strings.Contains(rendered, "Greet") {
        t.Errorf("expected rendered code to contain 'Greet', got:\n%s", rendered)
    }

    // Should NOT contain "// Output:" lines.
    if strings.Contains(rendered, "// Output:") {
        t.Errorf("expected rendered code to NOT contain '// Output:', got:\n%s", rendered)
    }

    // Should NOT have outer braces.
    trimmed := strings.TrimSpace(rendered)
    if strings.HasPrefix(trimmed, "{") {
        t.Errorf("expected rendered code without outer braces, got:\n%s", rendered)
    }
}
```

After:
```go
func TestParse_Render(t *testing.T) {
    rendered := getRendered(t)

    t.Run("should contain the function call", func(t *testing.T) {
        if !strings.Contains(rendered, "Greet") {
            t.Errorf("expected rendered code to contain 'Greet', got:\n%s", rendered)
        }
    })

    t.Run("should not contain output comments", func(t *testing.T) {
        if strings.Contains(rendered, "// Output:") {
            t.Errorf("got:\n%s", rendered)
        }
    })

    t.Run("should not have outer braces", func(t *testing.T) {
        trimmed := strings.TrimSpace(rendered)
        if strings.HasPrefix(trimmed, "{") {
            t.Errorf("got:\n%s", rendered)
        }
    })
}
```

Benefits:
- `go test -v` output shows each assertion as a named subtest
- Individual assertions can be targeted with `-run TestParse_Render/should_not`
- The subtest name IS the documentation, so error messages can be shorter
- Eliminates empty blocks that trigger `revive` linter warnings

## When to Use

- A test function has 2+ independent assertions preceded by `// Should ...` comments
- The assertions check distinct properties of the same result
- Especially when an assertion block would otherwise be empty (triggering `revive`)
