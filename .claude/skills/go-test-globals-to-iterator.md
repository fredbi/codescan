# Go Test Pattern: Replace Global Accumulators with Iterator Functions

**Extracted:** 2026-01-29
**Context:** Go test files using the "global slice + append helper + init" anti-pattern

## Problem

A common legacy Go test pattern uses a global slice populated by many `addXxxTests()`
functions, each calling a shared `addTest()` helper that appends to the global. This
triggers `gochecknoglobals` lint errors and obscures data flow:

```go
var tests = make([]myTest, 0) // global accumulator

func addTest(in any, wants ...string) {  // global appender
    tests = append(tests, myTest{in, wants})
}

func addFooTests() { addTest(...); addTest(...) }
func addBarTests() { addTest(...); addTest(...) }

func TestAll(t *testing.T) {
    addFooTests()
    addBarTests()
    for i, test := range tests { ... }
}
```

## Solution

Three-step conversion:

### 1. Each add function returns a slice with a local closure

```go
func fooTests() []myTest {
    var tests []myTest
    add := func(in any, wants ...string) {
        tests = append(tests, myTest{in, wants})
    }

    add(...)
    add(...)

    return tests
}
```

The local `add` closure keeps the call sites unchanged (just drop the prefix), minimizing
diff noise inside each function.

### 2. Chain with an iterator using slices.Concat + slices.Values

```go
func allTestCases() iter.Seq[myTest] {
    return slices.Values(slices.Concat(
        fooTests(),
        barTests(),
        bazTests(),
    ))
}
```

### 3. Consume with external index counter

```go
func TestAll(t *testing.T) {
    i := 0
    for tc := range allTestCases() {
        // use tc and i
        i++
    }
}
```

## Handling function-value references

If any test captures a former `addXxxTests` as a function value (e.g., to test function
pointer dumping), replace it with `func() {}` or another function matching the expected
type, since the signature changed from `func()` to `func() []myTest`.

## Handling build-tag variants

For functions split across build-tag files (e.g., `addCgoTests` in both
`cgo_test.go` and `nocgo_test.go`), convert both:
- The real one returns its slice normally
- The stub returns `nil`

## When to Use

- Test files with `var xxxTests` global slices and `addXxxTest` helpers
- `gochecknoglobals` lint errors on test data variables
- Legacy test files following the go-spew / testify accumulator pattern
- Any test file where data initialization is split across 5+ builder functions
