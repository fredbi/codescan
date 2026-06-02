# sync.Once Test Fixture Cache

**Extracted:** 2026-01-29
**Context:** Go integration tests that repeatedly parse the same package or load expensive resources

## Problem

Multiple parallel integration tests each independently perform the same expensive setup (e.g., `go/packages.Load`, file parsing, network calls). This wastes CPU and wall-clock time, especially in CI.

## Solution

Use `sync.Once` with package-level globals to parse the fixture once and share the result across all tests. This is safe because the parsed data is read-only after initialization.

Follow this project convention from `scanner/comments/extractor_test.go`:

## Example

```go
//nolint:gochecknoglobals // it is okay to use globals for [sync.Once] cache for parsed testdata
var (
    parseOnce sync.Once
    cached    parser.Examples
)

func loadTestExamples(t *testing.T) parser.Examples {
    t.Helper()

    parseOnce.Do(func() {
        examples, err := parser.New(
            "./testdata/examplespkg",
            parser.WithBuildTags("integrationtest"),
        ).Parse()
        if err != nil {
            t.Fatalf("Parse() error: %v", err)
        }

        cached = examples
    })

    return cached
}

func TestFoo(t *testing.T) {
    t.Parallel()
    examples := loadTestExamples(t)
    // ... assertions on examples
}
```

Key conventions:
- Group `sync.Once` and cache var in a single `var()` block
- Add `//nolint:gochecknoglobals` with an explanatory comment mentioning `sync.Once`
- The loader function takes `*testing.T` and calls `t.Helper()`
- Use `t.Fatalf` inside `sync.Once.Do` for setup failures

## When to Use

- Multiple test functions in the same file parse the same testdata package
- The parsed result is read-only (stateless parser, immutable output)
- The setup cost is non-trivial (package loading, file I/O, network)
- Tests are parallel (`t.Parallel()`) and share the same fixture
