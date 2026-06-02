# Rename Tooling Experience Report

Date: 2026-03-24

## Context

During Phase 3 preparation of the codescan package split, we needed to rename ~40 unexported
struct fields to exported names across ~25 types, with automatic update of all usage sites
(~52 positional struct literals + method body references). This is exactly the kind of
operation where AST-aware rename tooling should shine vs. regex-based sed.

## Tools Tried

### 1. `mcp__mcp-gopls__rename_symbol` (MCP tool)

**Expected**: Call with file URI + position + new name → edits applied to workspace.

**Actual**: Returns `{"edits":{}}` (empty edits object) for every position tried.

**Problems**:
- Position format unclear — the tool accepts `{"line": N, "character": N}` but doesn't
  document whether 0-indexed or 1-indexed. Tried both, neither produced edits.
- No error message when position doesn't match a symbol — silent failure with empty edits.
- No way to verify what symbol gopls thinks is at a given position (hover_info returned
  "position must be an object" error with the same position format).
- The `edits` field is always an empty object `{}`, even on apparent success — unclear if
  it's returning a diff to apply or if it's supposed to write files directly.

**Verdict**: Unusable in current form. The position resolution is broken or the return
format doesn't communicate results. Needs:
- A diagnostic mode: "what symbol is at this position?"
- Clear documentation of position format (0-indexed vs 1-indexed)
- Either apply edits directly or return them in a parseable format
- Error messages when no symbol is found at the given position

### 2. `gorename` (golang.org/x/tools/cmd/gorename)

**Expected**: `gorename -from '"pkg".Type.field' -to NewName` → renames in-place.

**Actual**: Scans the entire GOPATH/module cache, takes forever, produces pages of
irrelevant warnings about unrelated packages.

**Problems**:
- No way to scope to current module only — it tries to load every package visible in GOPATH
- The `-from` form works conceptually but the workspace scanning makes it impractical
- The `-offset` form (`-offset file:#byte`) also triggers full workspace scan
- Output is dominated by "While scanning Go workspace: Package X: no required module" noise
- On a machine with a large GOPATH, this tool is effectively unusable without manual scoping

**Verdict**: Correct semantics but terrible performance. Would work if it could be scoped
to `./...` only. The `-d` (dry-run) flag is useful but can't be evaluated through the noise.

### 3. `gopls rename` (CLI)

**Expected**: `gopls rename file.go:line:col NewName` → outputs renamed file.

**Actual**: Works correctly but only outputs ONE file to stdout per invocation.

**Problems**:
- Outputs the entire modified file to stdout — needs `-w` flag to write in place
- Only shows changes to one file at a time (the file containing the definition)
- Doesn't show what changes were made to OTHER files (e.g., if renaming a struct field
  that's used in 6 other files, you don't see those changes)
- `-w` flag writes in place but you can't verify what changed across the workspace
- Need to run once per field, sequentially — can't batch multiple renames
- Position is 1-indexed (line:col), which is different from LSP protocol (0-indexed)

**Verdict**: Actually works but clunky for batch operations. The single-file output
limitation means you need to trust it modified other files correctly.

### 4. `sed` (what we used in the previous failed attempt)

**Problems already documented in cascading-refactor-antipattern.md**:
- Can't distinguish positional vs named struct literals
- Can't distinguish field definitions from field accesses from method names
- Double-renames when run on already-modified text
- Multi-line patterns are unreliable

**Verdict**: Wrong tool for Go syntax transformations. Period.

## What a Good MCP Rename Tool Would Look Like

For `go-fred-mcp`, the ideal rename tool would:

1. **Accept symbol specification by name**, not position:
   ```
   rename_field(package="./...", type="setMaximum", field="builder", new_name="Builder")
   ```
   This avoids all the position-resolution issues.

2. **Batch multiple renames** in one call:
   ```
   rename_fields(package="./...", renames=[
     {type: "setMaximum", field: "builder", new_name: "Builder"},
     {type: "setMaximum", field: "rx", new_name: "Rx"},
     {type: "setMinimum", field: "builder", new_name: "Builder"},
     ...
   ])
   ```

3. **Return a summary of changes**, not the full file contents:
   ```json
   {
     "changes": [
       {"file": "parser.go", "line": 369, "old": "builder", "new": "Builder"},
       {"file": "schema.go", "line": 1535, "old": "&setMaximum{schemaValidations{ps}, rxf(...)}",
        "new": "&setMaximum{Builder: schemaValidations{ps}, Rx: rxf(...)}"},
       ...
     ],
     "files_modified": 8,
     "total_edits": 52
   }
   ```

4. **Apply edits in-place** and report success/failure per file.

5. **Dry-run mode** that shows the diff without applying.

6. **Scope to current module** — never scan GOPATH or module cache.

7. **Handle the positional-to-named struct literal conversion** automatically when
   exporting a field — this is the #1 use case for cross-package refactoring.

## Workaround Used

Given the tooling gaps, the practical approach for this session is:
- Use `gopls rename -w parser.go:LINE:COL NewName` for each field
- Run sequentially, verify compilation after each batch
- Accept that we can't preview cross-file changes before applying

This is slow (~30 seconds per rename due to gopls startup) but correct, unlike sed
which is fast but wrong.

## Key Insight

The fundamental issue is that **Go struct field renames are a cross-file operation**
but all available tools either:
- Think file-locally (sed)
- Think globally across the entire Go ecosystem (gorename)
- Think one-symbol-at-a-time (gopls rename)

What we need is **module-scoped batch rename** — rename N fields across M types,
applying all changes atomically within the current module. This is the tool to build
in go-fred-mcp.
