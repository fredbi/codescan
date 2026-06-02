# Go: Testing Generated Code Outside GOROOT with t.TempDir()

**Extracted:** 2026-02-08
**Context:** Testing Go code generators that produce compilable output in isolated temp directories

## Problem

When testing a Go code generator that writes `.go` files to `t.TempDir()`, two tools silently fail:

1. **`imports.Process()`** (from `golang.org/x/tools/imports`) — resolves and formats imports. It needs a valid Go module context to resolve import paths. Without `go.mod`, it can't find packages and either errors out or produces broken output.

2. **`go/packages.Load()`** — used by example parsers or any tool that does semantic analysis on generated code. It requires module-aware resolution to find dependencies.

Both work fine when generating code inside the project's own source tree (where `go.mod` already exists), but break when the target is an isolated temp directory outside the module root.

## Solution

Bootstrap a minimal Go module in the temp directory before running the generator. The pattern has 3 sequential steps:

1. **`go mod init`** — makes the temp dir a valid Go module
2. **`go mod edit -replace`** — redirects the dependency to the local source tree (avoids network fetches, tests against local code)
3. **`go get`** — resolves the dependency graph and populates `go.sum`

```go
//nolint:gosec // "tainted" args exec is actually okay in tests
func goModInit(t *testing.T, location, source string) {
    t.Helper()

    // Step 1: create go.mod in the temp directory
    mod := exec.CommandContext(t.Context(), "go", "mod", "init", path.Base(location))
    mod.Dir = location
    output, err := mod.CombinedOutput()
    if err != nil {
        t.Fatalf("go mod init at %s returned: %v: %s", location, err, string(output))
    }

    // Step 2: replace the module dependency with the local source tree
    replace := exec.CommandContext(t.Context(), "go", "mod", "edit",
        "-replace=github.com/example/mymodule="+source,
    )
    replace.Dir = location
    output, err = replace.CombinedOutput()
    if err != nil {
        t.Fatalf("go mod edit at %s returned: %v: %s", location, err, string(output))
    }

    // Step 3: resolve the dependency graph
    get := exec.CommandContext(t.Context(), "go", "get",
        "github.com/example/mymodule/pkg1",
        "github.com/example/mymodule/pkg2",
    )
    get.Dir = location
    output, err = get.CombinedOutput()
    if err != nil {
        t.Fatalf("go get at %s returned: %v: %s", location, err, string(output))
    }
}
```

### Usage in test:

```go
func TestGenerate(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    tmpDir := t.TempDir()
    targetRoot := filepath.Join(tmpDir, "output")
    os.MkdirAll(targetRoot, 0o700)

    cwd, _ := os.Getwd()
    goModInit(t, targetRoot, filepath.Join(cwd, ".."))

    // Now generate code into targetRoot — imports.Process and go/packages will work
    err := myGenerator.Generate(targetRoot)
    // ...
}
```

### Key details

- **`source` parameter**: absolute path to the local module root (where the real `go.mod` lives). Typically `filepath.Join(cwd, "..")` when tests run from a subdirectory.
- **`-replace` directive**: essential for offline/CI — avoids network fetches and tests against the working copy rather than a published version.
- **`go get` with specific packages**: only list the packages your generated code actually imports. This populates `go.sum` with the minimum required.
- **`testing.Short()` guard**: these tests shell out to `go` commands and are slower than unit tests. Skip them in short mode.
- **Non-parallel**: the test must NOT be `t.Parallel()` since the setup steps are sequential filesystem mutations.
- **`t.Context()`**: uses Go 1.21+ test context for proper cancellation.

## When to Use

- When testing a Go code generator that writes compilable `.go` files to a temp directory
- When the generated code imports packages from the parent module (or any module)
- When `imports.Process()` or `go/packages.Load()` fails with "could not import" or module resolution errors in tests
- When generated code needs to be syntactically valid and import-resolved outside the project's GOROOT
- Any integration test that produces Go source code in an isolated directory and needs it to compile or be analyzed
