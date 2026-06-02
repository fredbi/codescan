# TypeIndex / SpecBuilder Statefulness and the Path to Demand-Driven Building

Date: 2026-03-25

## The Current Architecture

The scan-to-spec pipeline has two phases, both accumulative:

### Phase 1: TypeIndex (scanner)

`NewTypeIndex` loads **all** packages via `golang.org/x/tools/go/packages` with full type
information (`NeedTypes + NeedSyntax + NeedTypesInfo`), walks every file, and populates:

- `AllPackages map[string]*packages.Package` — every loaded package, including transitive deps
- `Models map[*ast.Ident]*EntityDecl` — types annotated with `swagger:model`
- `Parameters []*EntityDecl` — types annotated with `swagger:parameters`
- `Responses []*EntityDecl` — types annotated with `swagger:response`
- `Routes []ParsedPathContent` — `swagger:route` annotations
- `Operations []ParsedPathContent` — `swagger:operation` annotations
- `Meta []MetaSection` — `swagger:meta` blocks
- `ExtraModels map[*ast.Ident]*EntityDecl` — discovered during Phase 2, stored back here

Everything is held in memory. For a large project, `AllPackages` alone can be hundreds of MB
because each `*packages.Package` retains the full AST and type-checker state for every file.

### Phase 2: SpecBuilder (builders/spec)

`Build()` processes the index in a fixed order:

```
buildModels()      → for each Model, build schema → joinExtraModels() loop
buildParameters()  → for each Parameter, build → append postDecls to discovered
buildResponses()   → for each Response, build → append postDecls to discovered
buildDiscovered()  → completion loop: process transitive dependencies
buildRoutes()      → match routes to operations
buildOperations()  → parse YAML operation blocks
buildMeta()        → parse swagger:meta blocks
```

The **completion loop** is the critical piece. When a SchemaBuilder processes a struct and
encounters a field of type `Pet`, it calls `ScanCtx.FindModel("Pet")`. If `Pet` wasn't
explicitly annotated with `swagger:model`, it gets added to `ExtraModels`. The loop:

1. `joinExtraModels()` promotes ExtraModels → Models, builds them.
2. Building those may discover more types → more ExtraModels.
3. Recurse until fixpoint.

Then `buildDiscovered()` does a similar loop for types discovered by parameter/response
builders via their `postDecls` queues.

**Two separate queues (`ExtraModels` and `discovered`) doing essentially the same thing** —
a sign that the discovery mechanism grew organically.

## Why This Design Is Problematic

### 1. Memory: everything loaded, everything retained

`packages.Load` with `NeedSyntax + NeedTypesInfo` loads the full AST and type-checker output
for every package in the dependency graph. For a project like the Kubernetes API:

- ~2000 Go packages in the dependency graph
- Each package has dozens of files with hundreds of type declarations
- The type checker retains cross-references between all of them
- `AllPackages` holds references to all of this until the process exits

The scanner only needs a fraction of these packages — the ones containing annotated types
and their transitive type dependencies. But it loads everything upfront because it doesn't
know which packages matter until after scanning.

### 2. The fixpoint loop is inherently batch

The completion loop can't start until `buildModels()` finishes. `buildDiscovered()` can't
start until `buildParameters()` and `buildResponses()` finish. Each phase must complete
before the next begins.

This means peak memory = all packages + all models + all parameters + all responses + all
discovered types + all schemas built so far. Nothing can be released early.

### 3. No incrementality

Change one comment annotation → reload all packages, rebuild entire index, rebuild entire
spec. There's no way to rebuild just the affected schema.

### 4. The discovery mechanism fights the index

`FindModel` is called during schema building. When it finds an unannotated type, it reaches
back into the TypeIndex to add it to `ExtraModels`. This breaks the clean phase separation:
Phase 2 (building) mutates Phase 1's data structure (the index). The builder isn't just
reading the index — it's growing it.

### 5. Doesn't scale to handler/route scanning

If the scanner starts analyzing handler functions (detecting `r.URL.Query().Get("limit")`,
recognizing framework routing patterns), it needs to process function bodies — not just
type declarations. The current TypeIndex has no concept of "functions of interest." Adding
handler scanning to the batch model means loading and retaining even more AST data.

## The Alternative: Demand-Driven Building

Instead of "index everything, then build everything," flip to "discover what's needed, build
it, follow references":

### Core idea: build from the API surface inward

The **API surface** is the set of explicitly annotated types and routes. These are the roots.
Everything else (schemas for referenced types, parameter types, response bodies) is discovered
by following references from the roots — not by pre-indexing the entire codebase.

```
API surface (roots):
  swagger:route GET /pets → handler → parameters, responses
  swagger:model Pet       → schema  → field types → nested schemas
  swagger:response Error  → schema  → field types

Build order: demand-driven, not phase-driven
  1. Find all roots (scan annotations — lightweight, comments only)
  2. For each root, build its spec fragment
  3. When building encounters a type reference, resolve it on demand
  4. Cache built schemas so each type is built at most once
```

### Architecture sketch

```
┌─────────────────┐
│  AnnotationScan │  Lightweight: scan comments only, no type-checking
│  (find roots)   │  Output: list of (file, position, annotation) tuples
└────────┬────────┘
         │ roots
┌────────▼────────┐
│  DemandLoader   │  Loads packages on demand, caches loaded packages
│  (lazy loading) │  Only loads packages that contain referenced types
└────────┬────────┘
         │ type info (on demand)
┌────────▼────────┐
│  SchemaCache    │  Builds and caches schemas
│  (build once)   │  Key: (package path, type name) → Schema
└────────┬────────┘
         │ schemas
┌────────▼────────┐
│  SpecAssembler  │  Assembles the final spec from built fragments
│  (stateless)    │  No fixpoint loop — everything was built on demand
└─────────────────┘
```

### DemandLoader: lazy package loading

The key change. Instead of `packages.Load` on everything upfront:

```go
type DemandLoader struct {
    cfg      *packages.Config
    loaded   map[string]*packages.Package  // cache of already-loaded packages
    mu       sync.RWMutex                  // safe for concurrent access
}

// LoadType loads just enough to resolve a type by package path + name.
// If the package isn't loaded yet, load it (and only it).
func (d *DemandLoader) LoadType(pkgPath, typeName string) (*types.Named, error) {
    pkg, err := d.ensurePackage(pkgPath)
    if err != nil { return nil, err }
    obj := pkg.Types.Scope().Lookup(typeName)
    // ...
}
```

**Memory impact**: Only packages that contain types actually referenced by the spec are
loaded. For a large project where the spec covers 50 types across 10 packages, but the
full dependency graph has 2000 packages — that's a 200x reduction in loaded packages.

**Trade-off**: Multiple `packages.Load` calls vs one big one. Each call has overhead (Go
module resolution, type-checking). Mitigation: batch loads when the reference graph is
known (e.g., load all packages referenced by a struct's fields in one call).

### SchemaCache: build-once, reference-many

```go
type SchemaCache struct {
    loader  *DemandLoader
    schemas map[typeKey]*spec.Schema       // built schemas
    pending map[typeKey]struct{}            // cycle detection
}

type typeKey struct {
    PkgPath  string
    TypeName string
}

// Get returns the schema for a type, building it on first access.
func (c *SchemaCache) Get(pkgPath, typeName string) (*spec.Schema, error) {
    key := typeKey{pkgPath, typeName}
    if s, ok := c.schemas[key]; ok {
        return s, nil  // already built
    }
    if _, ok := c.pending[key]; ok {
        return nil, ErrCyclicReference  // or produce a $ref
    }

    c.pending[key] = struct{}{}
    defer delete(c.pending, key)

    schema, err := c.build(pkgPath, typeName)
    if err != nil { return nil, err }

    c.schemas[key] = schema
    return schema, nil
}
```

**This replaces the entire ExtraModels + discovered + postDecls + fixpoint loop machinery.**
When building a schema for `Pet` and encountering a field of type `Owner`, the builder
calls `cache.Get("pkg", "Owner")`. If `Owner` hasn't been built yet, it's built now —
recursively. The cache prevents double-building. The pending set detects cycles.

No fixpoint loop. No two-phase completion. No mutation of the index during building.

### AnnotationScan: lightweight root discovery

The initial scan doesn't need type information. It only needs to find comments containing
`swagger:model`, `swagger:route`, etc. This can be done with `go/parser` in `ParseComments`
mode — much cheaper than `packages.Load`:

```go
func ScanAnnotations(patterns []string) ([]AnnotatedDecl, error) {
    fset := token.NewFileSet()
    // Parse only comments, no type-checking
    pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
    // Walk comments, find annotations, record (file, pos, kind, args)
}
```

This is O(files) with tiny memory — just comment strings, no AST retention.

For the v2 handler/route detection feature, this phase would also scan for function
signatures and routing registration patterns. Still lightweight — pattern matching on
AST nodes, not full type-checking.

### What about handler scanning?

Demand-driven building plays well with handler scanning:

1. **AnnotationScan** finds `swagger:route` and also detects handler functions by
   signature (e.g. `func(http.ResponseWriter, *http.Request)`).

2. For detected handlers, a **HandlerAnalyzer** (framework-specific plugin) examines
   the function body to infer parameters and responses. This needs the function's AST
   and type info — loaded on demand via `DemandLoader`.

3. Inferred parameters/responses reference types → resolved on demand via `SchemaCache`.

4. The handler's package is loaded only if the handler is detected. Packages that contain
   no handlers and no referenced types are never loaded.

```
Route detected: mux.HandleFunc("/pets/{id}", getPet)
  → load package containing getPet
  → analyze getPet body:
      r.URL.Query().Get("limit") → query param "limit", type string
      json.NewDecoder(r.Body).Decode(&input) → request body, type Input
      json.NewEncoder(w).Encode(pets) → response body, type []Pet
  → cache.Get("pkg", "Input") → build Input schema (loads Input's package if needed)
  → cache.Get("pkg", "Pet") → build Pet schema
```

Each step loads only what it needs. A handler that only uses stdlib types (`string`, `int`)
doesn't trigger any additional package loads.

## Comparison

| Aspect | Current (batch) | Demand-driven |
|--------|----------------|---------------|
| Initial load | All packages, full type info | Comments only (lightweight) |
| Package loading | Upfront, all at once | On demand, cached |
| Peak memory | O(all packages + all types) | O(referenced packages + built schemas) |
| Discovery | ExtraModels + discovered + fixpoint loop | Recursive cache.Get() |
| Index mutation during build | Yes (ExtraModels written back) | No (cache is append-only) |
| Incrementality | None (full rebuild) | Possible (invalidate cache entry) |
| Handler scanning | Doesn't fit (no function indexing) | Natural extension (load handler on demand) |
| Parallelism | Hard (shared mutable index) | Easier (cache is concurrent-safe, builds are independent) |
| Determinism | Non-deterministic (map iteration) | Deterministic (build order follows reference graph) |

## Migration Path

### Step 1: SchemaCache (can be done in v1.x)

Introduce `SchemaCache` as a layer between SpecBuilder and the individual builders. The
cache wraps the existing build logic but deduplicates and handles the fixpoint internally.
The TypeIndex still exists but the completion loops simplify.

This is a refactoring, not a rewrite. The builders' `postDecls` pattern is replaced by
cache lookups.

### Step 2: DemandLoader (v2)

Replace the upfront `packages.Load` with lazy loading. This is a bigger change because
the TypeIndex goes away — the DemandLoader replaces it.

The AnnotationScan phase can still use the existing comment-scanning logic (it's just
the initial file walk from `TypeIndex.build()` without the type-checking).

### Step 3: Handler scanning (v2+)

Add framework-specific handler detection to AnnotationScan, and HandlerAnalyzer plugins
that use DemandLoader to analyze function bodies.

## Risks

**Risk**: Multiple `packages.Load` calls may be slower than one big call.
**Mitigation**: Batch loading — when building a struct, collect all field type packages
and load them in one call. The Go module cache means repeated resolution is cheap.

**Risk**: Cycle detection in SchemaCache is more complex than the fixpoint loop.
**Mitigation**: The `pending` set handles simple cycles. For mutual recursion (A → B → A),
produce a `$ref` — which is already what the current builder does when it detects a known
definition name.

**Risk**: Demand-driven loading makes the build order dependent on reference traversal,
which could cause non-determinism if references are followed in map-iteration order.
**Mitigation**: Sort struct fields by name before processing. Sort annotation roots
by file position. The spec output should be deterministic regardless of discovery order
(the final spec is a JSON object with sorted keys).

**Risk**: Losing the "global view" — the batch approach knows about all annotated types
upfront, which is useful for validation ("this swagger:parameters refers to operationId X
but no route defines X").
**Mitigation**: The AnnotationScan phase still has the global view of all annotations.
Cross-reference validation happens there, before building begins.

## Open Questions

- **Streaming the spec**: Could the spec be emitted incrementally (e.g., stream schemas
  to a JSON encoder as they're built, without holding the entire spec in memory)? For very
  large APIs this could reduce peak memory further. The spec structure (definitions are a
  map, paths are a map) doesn't strictly require all entries to be in memory simultaneously.
  But JSON serialization of nested `$ref` structures may need the full definition set for
  validation. Worth exploring for extreme cases.

- **Caching across runs**: If the DemandLoader can fingerprint packages (go module version +
  file content hash), built schemas could be cached on disk. Subsequent runs only rebuild
  schemas whose source changed. This would make the tool fast enough for CI pre-commit hooks
  on large repos.

- **Parallel building**: With a concurrent-safe SchemaCache, multiple roots could be built
  in parallel. The DemandLoader would need a `singleflight`-style mechanism to avoid
  loading the same package concurrently from multiple goroutines. The individual schema
  builders are already stateless (they read from the cache and write to their own schema
  object) — parallelism is natural once the shared state (cache + loader) is safe.
