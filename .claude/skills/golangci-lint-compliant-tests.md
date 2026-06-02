# Writing golangci-lint Compliant Tests with Minimal Duplication

**Extracted:** 2026-01-26
**Context:** Go test files that need to pass golangci-lint checks while maintaining readability

## Problem

Test files often accumulate linter violations:
- **gochecknoglobals**: Package-level test data variables
- **unparam**: Helper functions with parameters that always receive the same value
- **gocognit/cyclop**: High complexity in test case generators
- **dupl**: Excessive code duplication in validation logic

These issues reduce maintainability and violate project linting standards.

## Solution

Apply a layered refactoring approach:

### 1. Extract Helper Functions (Eliminate Duplication)

**Before (160+ lines of duplicated validation):**
```go
// Test case 1
if o.EqualPrinter == nil {
    t.Error("EqualPrinter should not be nil")
} else {
    printer := o.EqualPrinter(w)
    if err := printer("test"); err != nil {
        t.Errorf("EqualPrinter error: %v", err)
    }
    w.Flush()
    expected := greenMark + "test" + endMark
    if buf.String() != expected {
        t.Errorf("EqualPrinter expected %q, got %q", expected, buf.String())
    }
}
// ... repeated 7 more times for different printers/themes
```

**After (single helper function):**
```go
// Helper function (reused 8+ times)
func validatePrinter(t *testing.T, builder PrinterBuilder, expectedMark, printerName string) {
    t.Helper()

    if builder == nil {
        t.Errorf("%s should not be nil", printerName)
        return
    }

    var buf bytes.Buffer
    w := bufio.NewWriter(&buf)

    printer := builder(w)
    if err := printer(testInput); err != nil {
        t.Errorf("%s error: %v", printerName, err)
        return
    }

    if err := w.Flush(); err != nil {
        t.Errorf("%s flush error: %v", printerName, err)
        return
    }

    expected := expectedMark + testInput + endMark
    if buf.String() != expected {
        t.Errorf("%s expected %q, got %q", printerName, expected, buf.String())
    }
}

// Usage
validatePrinter(t, o.EqualPrinter, greenMark, "EqualPrinter")
validatePrinter(t, o.DeletePrinter, redMark, "DeletePrinter")
// ... 6 more calls instead of 160+ lines
```

### 2. Use Data Structures Instead of Globals (Fix gochecknoglobals)

**Before:**
```go
var (
    lightThemeColors = colorSpec{
        equal:  greenMark,
        delete: redMark,
        update: cyanMark,
        insert: yellowMark,
    }

    darkThemeColors = colorSpec{
        equal:  brightGreenMark,
        delete: brightRedMark,
        update: brightCyanMark,
        insert: brightYellowMark,
    }
)
```

**After (define inline in closures):**
```go
func testCases() iter.Seq[testCase] {
    return slices.Values([]testCase{
        {
            name: "light theme",
            validate: func(t *testing.T, o *Options) {
                t.Helper()
                spec := colorSpec{
                    equal:  greenMark,
                    delete: redMark,
                    update: cyanMark,
                    insert: yellowMark,
                }
                validateAllColors(t, o, spec)
            },
        },
        // ... more cases
    })
}
```

### 3. Remove Unused Parameters (Fix unparam)

**Before:**
```go
func validateColorizer(t *testing.T, colorizer StringColorizer, input, expectedOutput, name string) {
    t.Helper()
    got := colorizer(input)
    if got != expectedOutput {
        t.Errorf("%s expected %q, got %q", name, expectedOutput, got)
    }
}

// All callers pass the same constant:
validateColorizer(t, c.expected, testInput, greenMark+testInput+endMark, "expected")
validateColorizer(t, c.actual, testInput, redMark+testInput+endMark, "actual")
```

**After:**
```go
const testInput = "test"

func validateColorizer(t *testing.T, colorizer StringColorizer, expectedOutput, name string) {
    t.Helper()
    got := colorizer(testInput)  // Use constant directly
    if got != expectedOutput {
        t.Errorf("%s expected %q, got %q", name, expectedOutput, got)
    }
}

// Simpler calls:
validateColorizer(t, c.expected, greenMark+testInput+endMark, "expected")
validateColorizer(t, c.actual, redMark+testInput+endMark, "actual")
```

### 4. Reduce Complexity with Composition

Create high-level validators that compose lower-level ones:

```go
// Low-level: validate one printer
func validatePrinter(t *testing.T, builder PrinterBuilder, expectedMark, name string)

// Mid-level: validate all printers for a theme
func validateAllPrinters(t *testing.T, o *Options, spec printerSpec) {
    t.Helper()
    validatePrinter(t, o.EqualPrinter, spec.equal, "EqualPrinter")
    validatePrinter(t, o.DeletePrinter, spec.delete, "DeletePrinter")
    validatePrinter(t, o.UpdatePrinter, spec.update, "UpdatePrinter")
    validatePrinter(t, o.InsertPrinter, spec.insert, "InsertPrinter")
}

// High-level: test case just passes the spec
validateAllPrinters(t, o, spec)
```

## Verification

Always verify with golangci-lint:

```bash
# Run on specific test file
golangci-lint run ./path/to/package/file_test.go

# Run on entire package
golangci-lint run ./path/to/package/

# Check for specific linters
golangci-lint run --enable-only=gochecknoglobals,unparam,dupl,gocognit ./path/
```

## When to Use

**Triggers:**
- Writing new test files with repetitive validation logic
- Refactoring existing tests that fail golangci-lint
- Test files with >200 lines and visible duplication patterns
- Multiple test cases validating similar structures (printers, colorizers, validators)
- Package-level variables in test files

**Red flags indicating refactoring needed:**
- Copy-pasted validation blocks with minor variations
- `gochecknoglobals` linter errors
- `unparam` warnings about parameters that never vary
- `dupl` warnings about duplicate code blocks
- High cyclomatic complexity in test case generator functions

## Benefits

- **Compliance**: Passes golangci-lint without exceptions
- **Maintainability**: ~80% reduction in test code (347 → 267 lines in example)
- **Readability**: Clear separation between test data and validation logic
- **Coverage**: Maintains 100% coverage while reducing code
- **Reusability**: Helper functions work across multiple test cases

## Project Context

This pattern is especially valuable in codebases that:
- Enforce strict linting standards (like go-openapi/testify)
- Use table-driven tests with iter.Seq patterns
- Test multiple variants of similar functionality
- Generate code and want consistent test coverage
