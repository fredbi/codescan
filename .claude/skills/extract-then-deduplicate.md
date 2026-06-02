---
name: extract-then-deduplicate
description: Refactoring pattern to fix nestif/gocognit/dupl linting issues by extracting nested blocks into functions, moving them to a common file, then deduplicating
user_invocable: false
---

# Extract-Then-Deduplicate Refactoring Pattern

Use this pattern when facing deeply nested code blocks (nestif, gocognit, gocyclo)
that are also duplicated across files (dupl).

## Steps

### 1. Extract nested blocks into standalone functions

Instead of inlining complex logic inside `if/else` branches or loop bodies,
extract each branch into its own function.

- Pass captured variables as explicit parameters instead of relying on closures.
- The extracted function should have a clear name describing what it configures
  (e.g. `setupInlineParamTaggers`, `setupResponseHeaderTaggers`).
- If the original block returned an error mid-way, the extracted function returns `error`.

**Before:**
```go
if condition {
    // 80 lines of deeply nested logic using local vars x, y
} else {
    // 5 lines
}
```

**After:**
```go
if condition {
    if err := setupComplexThing(sp, x, y); err != nil {
        return err
    }
} else {
    setupSimpleThing(sp, x)
}
```

### 2. Move extracted functions to a common file

Move all the newly created functions into their own file (e.g. `taggers.go`,
`builders.go`). This makes duplicated logic visually obvious when the functions
sit side by side.

### 3. Deduplicate shared logic

Once the functions are in the same file, identical or near-identical closures
and helper logic become easy to spot and factor out into shared functions.

- Closures that were duplicated across files become a single top-level function
  with the previously-captured variables passed as parameters.
- Recursive closures (like `parseArrayTypes`) become regular recursive functions
  with the extra context (e.g. `sp`, `name`) passed explicitly.

## When to apply

- `nestif` complaints about complex nested blocks inside loops
- `gocognit` / `gocyclo` on functions that mix orchestration with detail
- `dupl` flagging identical blocks across different files
- Any combination of the above (common in code that processes similar structures
  in slightly different contexts, like parameters vs responses)

## Key insight

Closures that capture local variables prevent deduplication. Converting them to
regular functions with explicit parameters is the unlock that makes sharing possible.
