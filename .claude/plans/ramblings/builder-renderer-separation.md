# Builder / Renderer Separation: Decoupling Scanning from Output Format

Date: 2026-03-25

## Context

In go-openapi v2, the `spec` package (which maps OpenAPI specs to Go structs) is being
replaced by immutable JSON Documents with navigation APIs (JSONPointer, JSONPath). The
entire approach of "build a `spec.Schema` struct, set its fields" disappears.

This creates both a problem and an opportunity for codescan.

## The Problem: Deep Coupling to spec Types

The current builders are saturated with `spec.*` types:

```go
// schema builder writes directly to spec.Schema
ps.Description = block.Description
ps.Maximum = &val
ps.ExclusiveMaximum = exclusive
ps.Ref = ref

// response builder writes directly to spec.Response
response.Description = JoinDropLast(lines)
resp.Headers[name] = ps  // ps is spec.Header

// parameter builder writes directly to spec.Parameter
param.In = "query"
param.Schema.Minimum = &v
param.Required = true

// route builder writes directly to spec.Operation
op.Summary = JoinDropLast(lines)
op.Tags = route.Tags
op.Consumes = consumes
```

Every builder imports `github.com/go-openapi/spec`. Every validation setter receives a
`spec.Schema` or `spec.Parameter` or `spec.Header` to mutate. The `ifaces.SwaggerTypable`
interface is shaped exactly around `spec.Schema`'s mutation API.

When `spec` becomes a JSON Document with no settable Go struct fields, **every builder
breaks**. Rewriting them all to target the new Document API would couple them to that new
API just as tightly. And we'd be back to square one when we want to target protobuf, or
gRPC, or GraphQL, or AsyncAPI.

## The Insight: There Are Two Distinct Concerns

### Concern 1: Understanding the Go code (building)

"This struct has a field `Name string` with a `json:"name"` tag and a comment saying
`required: true, min length: 1`. It references type `Address` in package `models`."

This is **format-independent**. Whether the output is OpenAPI 2.0, 3.1, protobuf, or
GraphQL, the Go code says the same thing.

### Concern 2: Expressing it in a target format (rendering)

"In OpenAPI 3.1, this field becomes a `properties.name` entry with `type: string`,
`minLength: 1`, inside a `required: [name]` array at the schema level. In protobuf, it
becomes `string name = 1;` with no validation (protobuf doesn't have minLength). In
OpenAPI 2.0, the `required` array lives on the parent schema, not the field."

This is **entirely format-specific**. The structural rules, naming conventions, where
`required` lives, how `$ref` works, what validations are expressible — all differ per format.

**The current builders conflate both concerns.** They understand Go types AND write
spec.Schema fields. The v2 should split them.

## The Intermediate Representation (IR)

Between building and rendering, an **abstract API model** captures what the scanner
understood from the code, without committing to any output format.

### What the IR represents

Every HTTP API (regardless of spec format) has these concepts:

```
API
├── Info (title, version, description, contact, license)
├── Servers / Host+BasePath
├── SecuritySchemes
├── Models (named types → schemas)
│   └── Model
│       ├── Name, Description
│       ├── Fields
│       │   └── Field
│       │       ├── Name, GoName, Description
│       │       ├── Type (primitive, ref, array, map, any)
│       │       ├── Validations (min, max, pattern, enum, ...)
│       │       ├── Required, ReadOnly, Nullable
│       │       └── Extensions
│       ├── AllOf / Composition
│       └── Discriminator
├── Operations (handler functions → endpoints)
│   └── Operation
│       ├── ID, Summary, Description
│       ├── Method, Path
│       ├── Tags
│       ├── Deprecated
│       ├── Security
│       ├── Consumes, Produces (or Content negotiation)
│       ├── Parameters
│       │   └── Parameter
│       │       ├── Name, In (query/path/header/cookie/body)
│       │       ├── Type, Validations
│       │       ├── Required
│       │       └── Description
│       ├── RequestBody (OAI 3.x concept, mapped from body param in 2.0)
│       └── Responses
│           └── Response
│               ├── StatusCode (or "default")
│               ├── Description
│               ├── Schema (ref or inline)
│               └── Headers
└── Extensions (x-* vendor properties)
```

This is **not** an OpenAPI struct. It's not a protobuf descriptor. It's the **scanner's
understanding of the API**, expressed in terms that are universal to API description.

### IR design principles

1. **Immutable after building.** The builders produce IR nodes; the renderers only read them.
   No mutation during rendering. This enables concurrent rendering to multiple formats.

2. **No $ref.** The IR uses direct Go pointers/references. A field of type `Address` points
   to the `Address` IR Model node. The renderer decides whether to inline it, produce a
   `$ref`, or produce a protobuf `import`.

3. **Superset of expressible properties.** The IR can represent things that not every format
   supports. `minLength` exists in the IR even though protobuf can't express it. The renderer
   decides what to keep and what to drop (or warn about).

4. **Carries provenance.** Each IR node knows where it came from: file, line, column,
   annotation text. This enables: (a) error messages that point to source, (b) LSP
   integration, (c) a `codescan explain` command that shows which code produced which
   spec entry.

5. **No dependency on any spec library.** The IR is defined entirely within codescan.
   No imports from `go-openapi/spec`, no imports from protobuf libraries. It's plain Go
   structs.

### IR sketch

```go
package ir

// Model represents a named type that becomes a schema/message/type definition.
type Model struct {
    Name        string
    GoName      string        // original Go type name (may differ from Name if overridden)
    Package     string        // Go package path
    Description string
    Fields      []Field
    AllOf       []*ModelRef   // composition (embedded structs)
    Discriminator *Discriminator
    Extensions  map[string]any
    Source      SourcePos     // where in Go source this was defined
}

// Field represents a struct field.
type Field struct {
    Name        string        // JSON/spec name
    GoName      string        // Go field name
    Description string
    Type        TypeRef       // what type this field is
    Validations Validations   // min, max, pattern, enum, ...
    Required    bool
    ReadOnly    bool
    Nullable    bool
    Deprecated  bool
    Extensions  map[string]any
    Source      SourcePos
}

// TypeRef describes a field's type abstractly.
type TypeRef struct {
    Kind      TypeKind       // Primitive, Model, Array, Map, Any
    Primitive *PrimitiveType // non-nil when Kind == Primitive
    Model     *ModelRef      // non-nil when Kind == Model (pointer into IR graph)
    Items     *TypeRef       // non-nil when Kind == Array
    MapKey    *TypeRef       // non-nil when Kind == Map
    MapValue  *TypeRef       // non-nil when Kind == Map
}

type TypeKind int
const (
    KindPrimitive TypeKind = iota
    KindModel
    KindArray
    KindMap
    KindAny
)

type PrimitiveType struct {
    Name   string // "string", "integer", "number", "boolean"
    Format string // "int32", "int64", "float", "double", "date-time", ...
}

// ModelRef is a reference to another Model in the IR graph.
type ModelRef struct {
    Model *Model  // direct pointer — no $ref resolution needed
}

// Validations captures all validation constraints expressible across formats.
type Validations struct {
    Minimum          *float64
    Maximum          *float64
    ExclusiveMinimum bool
    ExclusiveMaximum bool
    MultipleOf       *float64
    MinLength        *int64
    MaxLength        *int64
    Pattern          string
    MinItems         *int64
    MaxItems         *int64
    UniqueItems      bool
    Enum             []any
    Default          any
    Example          any
}

// Operation represents an API endpoint.
type Operation struct {
    ID          string
    Method      string
    Path        string
    Summary     string
    Description string
    Tags        []string
    Deprecated  bool
    Consumes    []string        // media types accepted
    Produces    []string        // media types produced
    Security    []SecurityReq
    Parameters  []Parameter
    RequestBody *RequestBody    // OAI 3.x concept; in 2.0, rendered as body parameter
    Responses   []Response
    Extensions  map[string]any
    Source      SourcePos
}

// Parameter represents an operation parameter.
type Parameter struct {
    Name        string
    In          ParamLocation   // Query, Path, Header, Cookie, Body (body is legacy)
    Description string
    Type        TypeRef
    Validations Validations
    Required    bool
    AllowEmpty  bool
    Extensions  map[string]any
    Source      SourcePos
}

type ParamLocation int
const (
    InQuery ParamLocation = iota
    InPath
    InHeader
    InCookie   // OAI 3.x
    InBody     // OAI 2.0 legacy; rendered as RequestBody in 3.x
    InFormData // OAI 2.0 legacy; rendered as RequestBody multipart in 3.x
)

// Response represents an operation response.
type Response struct {
    StatusCode  int            // 0 = default
    Description string
    Schema      *TypeRef       // response body type
    Headers     []Header
    Extensions  map[string]any
    Source       SourcePos
}

// SourcePos links an IR node back to Go source for error reporting and LSP.
type SourcePos struct {
    File   string
    Line   int
    Column int
}
```

## The Renderer Interface

```go
package render

// Renderer transforms an IR API into an output format.
type Renderer interface {
    // RenderAPI produces the complete output for an API.
    RenderAPI(api *ir.API) ([]byte, error)
}
```

Concrete renderers:

```go
// OpenAPI 2.0 (Swagger) — produces JSON or YAML
type Swagger20Renderer struct { ... }

// OpenAPI 3.0.x — produces JSON or YAML
type OpenAPI30Renderer struct { ... }

// OpenAPI 3.1.x — produces JSON or YAML (uses JSON Schema 2020-12)
type OpenAPI31Renderer struct { ... }

// Protobuf — produces .proto file content
type ProtobufRenderer struct { ... }

// Markdown — produces human-readable API documentation
type MarkdownRenderer struct { ... }
```

Each renderer makes format-specific decisions:

| IR concept | OAI 2.0 | OAI 3.1 | Protobuf |
|------------|---------|---------|----------|
| Model | `definitions.Name` | `components.schemas.Name` | `message Name` |
| Field.Required | parent schema `required: [name]` | same | not expressible (all optional in proto3) |
| Field.Nullable | `x-nullable: true` | `type: ["string", "null"]` | wrapper type or `optional` |
| Parameter InBody | `in: body` parameter | `requestBody` | RPC input message |
| Parameter InFormData | `in: formData` | `requestBody` multipart | not applicable |
| ModelRef | `$ref: "#/definitions/X"` | `$ref: "#/components/schemas/X"` | field type or `import` |
| Validations.MinLength | `minLength: N` | `minLength: N` | dropped (no equivalent) |
| Consumes/Produces | top-level arrays | per-operation `content` keys | not applicable |
| Extensions | `x-*` keys | `x-*` keys | custom options |

The renderer also handles:
- **Cycle breaking**: The IR uses direct pointers. The renderer must detect cycles and produce
  `$ref` (or protobuf `import`) as appropriate.
- **Name collision**: Two Go types `pkg1.Pet` and `pkg2.Pet` both become `Pet` in the IR.
  The renderer disambiguates (e.g., `Pet` and `Pkg2Pet`, or uses component paths).
- **Capability gating**: A protobuf renderer encountering `minLength` on a field can warn
  "validation constraint dropped: protobuf does not support minLength" — using `Source` to
  point the user at the annotation that will be lost.

## What Changes in the Builder

The builders no longer import `spec` or any output format library. They produce IR nodes:

```go
// BEFORE (current):
func (s *SchemaBuilder) Build(definitions map[string]spec.Schema) error {
    ps := new(spec.Schema)
    ps.Description = "..."
    ps.Maximum = &val
    definitions[name] = *ps
}

// AFTER (v2):
func (s *SchemaBuilder) Build() (*ir.Model, error) {
    return &ir.Model{
        Name:        name,
        Description: "...",
        Fields:      fields,   // []ir.Field, built from struct fields
    }, nil
}
```

The builder's job simplifies: it reads Go types and annotations, and produces a format-neutral
description. It doesn't need to know:
- Where `required` lives in the output (on the field? on the parent? in an array?).
- How `$ref` is spelled.
- Whether the output has `definitions` or `components/schemas`.
- Whether `nullable` is a keyword or a type union.

**The ifaces.SwaggerTypable interface disappears entirely.** It was an abstraction over
"things that can be typed in a Swagger schema" — but the IR's TypeRef replaces it with
a cleaner, format-neutral representation.

**The ValidationBuilder interface simplifies.** Instead of 13 setter methods that each
write to a different spec struct field, the builder just populates `ir.Validations` —
a plain struct with no behavior.

## The Pipeline

```
Go source
    │
    ▼
AnnotationScan          ← lightweight, comments only
    │ roots
    ▼
DemandLoader            ← lazy package loading
    │ types + AST
    ▼
Builders                ← produce ir.Model, ir.Operation, ir.Parameter, ...
    │ IR graph
    ▼
Renderer                ← format-specific: OAI 2.0, 3.1, protobuf, ...
    │
    ▼
Output (JSON, YAML, .proto, ...)
```

Each stage has a clean boundary:
- AnnotationScan → DemandLoader: "load this package"
- DemandLoader → Builders: type info + AST
- Builders → IR: format-neutral API description
- IR → Renderer: format-specific output

No stage knows about the stages above or below it (beyond the interface boundary).

## Why This Is Worth the Rewrite

### 1. Multi-format output from one scan

Run the scanner once, render to OAI 2.0 for legacy consumers AND OAI 3.1 for modern ones
AND markdown for documentation. The builders run once; only the renderers differ.

### 2. The spec dependency disappears from the core

`go-openapi/spec` is no longer imported by the scanner or builders. Only the OAI 2.0
renderer imports it (or its v2 successor, the JSON Document library). The scanner becomes
independent of any specific spec library version.

### 3. Testing simplifies dramatically

Testing a builder means: "given this Go type, does the IR look right?" — no need to
construct spec.Schema objects in test assertions. The IR is plain structs, easy to compare.

Testing a renderer means: "given this IR, does the JSON output look right?" — golden-file
tests against expected JSON/YAML. No Go type system involved.

The current tests conflate both: they load Go packages, run the full pipeline, and assert
on spec.Schema fields. These are integration tests disguised as unit tests.

### 4. Protobuf (and beyond) becomes feasible

Generating `.proto` files from Go source has real value: teams that have a Go API and want
to add gRPC support. The IR already contains everything needed — types, fields, nesting,
descriptions. A protobuf renderer is ~300 lines that maps IR concepts to proto3 syntax.

Similarly: GraphQL schema generation, AsyncAPI for event-driven APIs, JSON Schema standalone
(not embedded in OpenAPI), TypeScript type definitions for frontend consumers.

### 5. The IR is the integration point for smart detection

When the handler scanner infers "this function reads `limit` from the query string," it
produces an `ir.Parameter{Name: "limit", In: InQuery, Type: string}`. The annotation
parser produces the same IR type when it reads `// in: query`. Both feed into the same
IR graph. The renderer doesn't know or care whether the parameter was inferred or annotated.

## Risks

**Risk**: The IR becomes a lowest-common-denominator that can't express format-specific
features.
**Mitigation**: The IR is a **superset**, not an intersection. It has fields for OAI 3.x
concepts (RequestBody, Cookie parameters) even though OAI 2.0 can't render them. The
renderer decides what to drop. For truly format-specific concepts (protobuf `oneof`,
GraphQL `union`), the IR can carry `Extensions` that specific renderers understand.

**Risk**: Two layers of abstraction (IR + renderer) instead of one (direct spec building)
adds complexity.
**Mitigation**: The IR layer is simple — plain structs, no behavior, no interfaces. The
complexity it adds is far less than the complexity it removes (no more SwaggerTypable,
ValidationBuilder, 4 different tagger setup functions per builder, postDecls, etc.).

**Risk**: The IR graph uses direct pointers, which makes serialization/debugging harder
than `$ref` strings.
**Mitigation**: The IR is not serialized. It exists only in memory during the scan. An
`ir.Dump()` debug printer can show the graph with cycle detection for debugging. The
renderer handles serialization into the target format.

**Risk**: Mapping OAI 2.0 ↔ 3.x concepts in the IR (e.g., body parameter vs requestBody)
requires the IR to understand version-specific semantics.
**Mitigation**: The IR uses the richer representation (requestBody) as the canonical form.
The OAI 2.0 renderer knows how to degrade it to a body parameter. The builder doesn't
need to choose — it always produces the richer form.

## Clarification: Where the IR Sits in the Ecosystem

The go-openapi v2 ecosystem has a shared core component: the **JSON Document** — an immutable
indexed hierarchy of nodes with navigation (JSONPointer, JSONPath) and builders for construction
and copy-with-mutation. This core has its own JSON parser/writer designed for low memory usage
(large documents). It is the universal representation for specs (Swagger 2.0 Document, OAI 3.x
Document, JSON Schema Document, etc.).

The two main tools that produce/consume these Documents are:

1. **codescan** (this repo): Go source → spec Document (code scanning → spec generation)
2. **Code generator**: spec Document → Go source (spec → code generation)

These two directions **don't share much**. The code generator's internal representation is
necessarily complex (it must reason about language features, templates, package layout, import
graphs, naming conventions). The scanner's internal representation can stay simple — it only
needs to capture "what did the Go code say about the API?"

### Revised architecture

```
Go source
    │
    ▼
┌──────────┐
│ codescan │  AnnotationScan → DemandLoader → Builders
│          │      │
│          │      ▼
│          │  IR (simple: models, operations, fields, validations)
│          │      │
│          │      ▼
│          │  Renderer(s)
└──────────┘      │
                  ▼
           JSON Document (immutable, indexed)
                  │
           ┌──────┴──────┐
           ▼              ▼
    OAI 2.0 Document  OAI 3.1 Document  (protobuf, etc.)
```

The **renderer's output is a JSON Document**, not raw JSON text. The Document library handles
serialization (JSON, YAML). This means:

- The renderer calls Document builder APIs: `doc.Set("/paths/~1pets/get/summary", "List pets")`.
- The renderer doesn't `json.Marshal` anything — the Document library does that.
- The Document is the boundary between codescan and the rest of the ecosystem.

The **IR is internal to codescan**. It is not a shared data structure. It doesn't need to be
versioned as a public API. It exists only between the builders and the renderers, within a
single process, within a single scan invocation.

This keeps the IR deliberately simple:

- Plain Go structs, no interfaces, no methods beyond accessors.
- Direct pointers for references (no `$ref` resolution, no indexing).
- No JSON serialization capability needed.
- No stability guarantee beyond "all renderers in this repo consume it."

The complexity lives in two places where it belongs:

1. **The JSON Document library** (shared core) — handles immutability, indexing, navigation,
   memory-efficient storage, JSONPointer/JSONPath, copy-with-mutation builders. This is a
   separately maintained, well-tested foundation.

2. **The code generator's IR** (separate repo) — handles the reverse direction: spec Document
   → language-specific code model → template rendering. This IR is necessarily complex because
   code generation is complex.

Codescan's IR is the simple one. It captures "Go types + annotations → abstract API model"
and nothing more. The renderer's job is to translate that simple model into the Document
builder calls that produce the correct spec structure.

### What this means for the IR design

The IR proposed earlier in this document is roughly the right shape, but with these
refinements:

- **No need for round-tripping.** The IR is write-once (by builders), read-once (by
  renderers). It's never persisted, never loaded from disk, never shared across processes.
  The `InputSpec` (base spec overlay) feature works differently: the base spec is loaded
  as a Document, and the renderer merges scanned IR into it at the Document level — not
  by converting the base spec into IR.

- **The renderer produces Document builder calls, not JSON.** Instead of:
  ```go
  func (r *OAI31Renderer) RenderAPI(api *ir.API) ([]byte, error)
  ```
  It's:
  ```go
  func (r *OAI31Renderer) RenderAPI(api *ir.API) (*jsondoc.Document, error)
  ```
  The caller can then serialize the Document to JSON or YAML, or navigate it, or merge it
  with another Document.

- **Protobuf and other non-JSON formats** are renderers that produce their own output type
  (e.g., `[]byte` for `.proto` text). They don't go through the Document library since
  protobuf isn't JSON. The Renderer interface should be generic enough:
  ```go
  type Renderer[T any] interface {
      RenderAPI(api *ir.API) (T, error)
  }
  ```
  Where `T` is `*jsondoc.Document` for OpenAPI renderers and `[]byte` for protobuf.

## Open Questions

- **Extensions passthrough**: Vendor extensions (`x-*`) from annotations should flow through
  to the output. The IR carries them as `map[string]any`. But what about extensions that are
  format-specific? E.g., `x-go-type` makes sense in OpenAPI but not in protobuf. Should the
  renderer filter extensions? Or should the IR tag them with intended audience?

- **Base spec merging**: The current `InputSpec` loads a base Swagger spec and overlays
  scanned annotations onto it. In v2, the base spec is a Document. The renderer must merge
  IR-derived entries with existing Document entries. This is Document-level merging, not
  IR-level — the IR never sees the base spec. The Document library's copy-with-mutation
  builders should make this natural, but the merge semantics need defining (e.g., does a
  scanned schema replace or deep-merge with a base schema?).

- **Schema-only mode**: For users who want JSON Schema output without the API layer
  (operations, routes), the IR should support a "models only" mode. The builders already
  know this (`scanModels` flag). The IR's Operation/Parameter types are simply not populated.
  A JSON Schema renderer consumes only `ir.Model` nodes. This suggests the IR should be
  structured as composable layers: `ir.Schema` (models, fields, types) and `ir.API`
  (operations, routes, security — which embeds `ir.Schema`).
