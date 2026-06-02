# Fixing dupl Linter in Tests: Extract Shared Test Table as Iterator

**Extracted:** 2026-03-17
**Context:** When golangci-lint reports `dupl` on structurally identical table-driven tests that differ only in types/functions

## Problem
Two or more table-driven tests have identical structure (same case names, same expected behavior patterns) but test different functions or types. The `dupl` linter flags them as duplicates.

## Solution
Extract the shared test case table into a function returning `iter.Seq[T]`, then have each test function supply only the type-specific inputs.

### Step 1: Identify the shared structure
Look at the duplicate tests side by side. The shared parts are:
- Case names ("both null", "greater", "less", etc.)
- Expected behavior pattern (which change code to expect)

The differing parts are:
- Input types (`*float64` vs `*int64`)
- The function under test (`CompareFloatValues` vs `CompareIntValues`)

### Step 2: Extract shared cases as an iterator

```go
type compareValueCase struct {
    name       string
    fieldName  string
    wantChange SpecChangeCode
}

func compareValueCases() iter.Seq[compareValueCase] {
    return slices.Values([]compareValueCase{
        {name: "both null", fieldName: "bob", wantChange: NoChangeDetected},
        {name: "greater", fieldName: "bob", wantChange: WidenedType},
        // ...
    })
}
```

### Step 3: Each test provides type-specific inputs via a map

```go
func TestCompareFloatValues(t *testing.T) {
    floatInputs := map[string]struct{ field1, field2 *float64 }{
        "both null": {nil, nil},
        "greater":   {floatPointerOf(1.0), floatPointerOf(2.0)},
        // ...
    }

    for tc := range compareValueCases() {
        in := floatInputs[tc.name]
        t.Run(tc.name, func(t *testing.T) {
            got := CompareFloatValues(tc.fieldName, in.field1, in.field2, WidenedType, NarrowedType)
            // assert on tc.wantChange
        })
    }
}
```

This first step of extracting the table often reveals that the duplication was mostly in the test scaffolding, not the logic. Sometimes just this extraction is enough to satisfy the linter.

## When to Use
- `dupl` linter flags on table-driven tests
- Two+ test functions with identical case names/structure but different types
- First step before considering generics — often sufficient on its own
