# Type Dependency Analysis for Package Split

Date: 2026-03-24

## Dependency Layers (bottom-up)

### Layer 0: Pure data types (no dependencies on other codescan types)

```
Options          — config struct, only spec.Swagger
MetaSection      — just *ast.CommentGroup
ParsedPathContent — strings + *ast.CommentGroup
EntityDecl       — AST/types + regex helpers (modelOverride, etc.)
```

These four types have NO dependencies on other codescan types (only stdlib + spec).
`EntityDecl` calls helper functions like `modelOverride()`, `responseOverride()`,
`parametersOverride()` — these are thin wrappers over regex matching, also with no
type dependencies.

### Layer 1: Interfaces (depend on nothing from codescan)

```
SwaggerTypable              — references only spec types
ValidationBuilder           — references only primitives
OperationValidationBuilder  — extends ValidationBuilder
ValueParser                 — Parse/Matches on strings
Objecter                    — wraps *types.TypeName
```

These are pure interfaces with no concrete type references.

### Layer 2: Parser infrastructure (depends on Layer 1 interfaces)

```
TagParser          — wraps ValueParser (Layer 1)
SectionedParser    — contains []TagParser, ValueParser
YAMLSpecScanner    — standalone, no codescan type deps
yamlParser         — standalone
```

`SectionedParser` depends on `TagParser` and `ValueParser` (interface). It does NOT
depend on any builder types or scanner types.

### Layer 3: Setter types (depend on Layer 1 interfaces)

```
setMaximum, setMinimum, setMultipleOf, setMaxItems, setMinItems,
setMaxLength, setMinLength, setPattern, setUnique, setEnum
    → all hold: ValidationBuilder (interface)

setCollectionFormat
    → holds: OperationValidationBuilder (interface)

setDefault, setExample
    → holds: ValidationBuilder + *spec.SimpleSchema

matchOnlyParam, setRequiredParam
    → holds: *spec.Parameter

setReadOnlySchema, setDiscriminator, setRequiredSchema
    → holds: *spec.Schema

setDeprecatedOp
    → holds: *spec.Operation

setSchemes, setSecurity, multiLineDropEmptyParser
    → holds: func callbacks

setOpResponses
    → holds: func callbacks + maps of spec types

setOpExtensions
    → holds: func callback

setMetaSingle
    → holds: *spec.Swagger

setOpParams
    → holds: *spec.Parameter slice
```

ALL setter types depend ONLY on Layer 1 interfaces (ValidationBuilder,
OperationValidationBuilder) and spec types. They have ZERO dependencies on
scanner types, builder types, or each other.

### Layer 4: Scanner context (depends on Layer 0)

```
TypeIndex  — contains EntityDecl, MetaSection, ParsedPathContent
ScanCtx    — contains TypeIndex, Options
```

`ScanCtx` returns `EntityDecl`, `MetaSection`, `ParsedPathContent` from its methods.
`TypeIndex` stores them. Both depend only on Layer 0 types.

### Layer 5: Typable + Validation implementations (depends on Layer 1 interfaces + spec)

```
SchemaTypable      — implements SwaggerTypable, holds *spec.Schema
schemaValidations  — implements ValidationBuilder, holds *spec.Schema
paramTypable       — implements SwaggerTypable, holds *spec.Parameter
paramValidations   — implements ValidationBuilder+OperationValidationBuilder
itemsTypable       — implements SwaggerTypable, holds *spec.Items
itemsValidations   — implements ValidationBuilder+OperationValidationBuilder
responseTypable    — implements SwaggerTypable, holds *spec.Header + *spec.Response
headerValidations  — implements ValidationBuilder+OperationValidationBuilder
```

These depend on Layer 1 interfaces (they implement them) and spec types.
`paramTypable.Items()` and `responseTypable.Items()` call `bodyTypable()` which
returns `SchemaTypable` — this is the ONLY cross-dependency within Layer 5.

### Layer 6: Builders (depend on Layer 2-5)

```
SchemaBuilder
    → holds: *ScanCtx, *EntityDecl
    → creates: SectionedParser, TagParser, SchemaTypable, schemaValidations
    → calls: setter constructors (setMaximum, etc.)
    → uses: SwaggerTypable interface in method signatures

ParameterBuilder
    → holds: *ScanCtx, *EntityDecl
    → creates: SectionedParser, TagParser, paramTypable, paramValidations,
               itemsValidations, SchemaTypable, SchemaBuilder(!)
    → calls: setter constructors

ResponseBuilder
    → holds: *ScanCtx, *EntityDecl
    → creates: SectionedParser, TagParser, responseTypable, headerValidations,
               SchemaTypable, SchemaBuilder(!)
    → calls: setter constructors

OperationsBuilder
    → holds: *ScanCtx, ParsedPathContent
    → creates: YAMLSpecScanner

RoutesBuilder
    → holds: *ScanCtx, ParsedPathContent
    → creates: SectionedParser, TagParser, setter objects
    → calls: setter constructors (newSetResponses, newSetParams, etc.)
```

### Layer 7: Orchestrator (depends on everything)

```
SpecBuilder
    → holds: *ScanCtx, []*EntityDecl
    → creates: SchemaBuilder, ParameterBuilder, ResponseBuilder,
               OperationsBuilder, RoutesBuilder, SectionedParser
```

## The Dependency Graph (DAG)

```
Layer 0: Options  MetaSection  ParsedPathContent  EntityDecl
              \        |              |            /
Layer 1:      SwaggerTypable  ValidationBuilder  ValueParser
                    |               |               |
Layer 2:         TagParser     SectionedParser   YAMLSpecScanner
                    |               |
Layer 3:     setMaximum... setDefault... matchOnlyParam... setOpResponses...
                    |               |
Layer 4:         ScanCtx ← TypeIndex
                    |
Layer 5:  SchemaTypable  paramTypable  responseTypable  (+ validations)
                    |         |              |
Layer 6:    SchemaBuilder  ParameterBuilder  ResponseBuilder  OperationsBuilder  RoutesBuilder
                    \           |                |                  /            /
Layer 7:                          SpecBuilder
```

## Critical Cross-Cuts

### 1. ParameterBuilder → SchemaBuilder
`ParameterBuilder` instantiates `SchemaBuilder` at parameters.go:329,344 for sub-schema building.

### 2. ResponseBuilder → SchemaBuilder
`ResponseBuilder` instantiates `SchemaBuilder` at responses.go:212,239,292.

### 3. bodyTypable → SchemaTypable
`bodyTypable()` in schema.go returns `SchemaTypable`. Called from `paramTypable.Items()`
and `responseTypable.Items()`. This means param and response typables depend on schema typable.

### 4. Setter types are INSTANTIATED by builders but DEFINED in parser layer
All builders create setter objects (`&setMaximum{Builder: ..., Rx: ...}`) using
concrete struct types. The setter types live in the parser layer but are wired up
by the builder layer.

### 5. EntityDecl.Names() calls parser helpers
`EntityDecl.Names()`, `ResponseNames()`, `OperationIDs()` call `modelOverride()`,
`responseOverride()`, `parametersOverride()` which are regex-based comment matchers.

## Implications for Package Structure

### What CAN be cleanly separated (no circular deps):

1. **Layer 0 types** → `internal/scanner` (EntityDecl, MetaSection, ParsedPathContent, TypeIndex, ScanCtx)
2. **Layer 1 interfaces** → `internal/ifaces` or stay in root
3. **Layer 2+3 parser types** → `internal/parsers` (SectionedParser, TagParser, all setters, YAML parser)
4. **Layer 7 orchestrator** → `internal/builders/spec`

### What's ENTANGLED:

1. **Typables + validations** (Layer 5) cross-reference each other via `bodyTypable()`.
   If they go into separate builder packages, `paramTypable` and `responseTypable`
   would import from the schema package just for `bodyTypable()`.

2. **Builders** (Layer 6) instantiate setter types from Layer 3. If setters are in
   `internal/parsers`, builders import from parsers — this is fine (not circular).

3. **Builders** cross-reference: param→schema, response→schema. This means separate
   builder packages would have `parameter → schema` and `response → schema` dependencies.

### Recommended package structure:

```
codescan/                     — public API: Options, Run()
  internal/
    scanner/                  — ScanCtx, TypeIndex, EntityDecl, MetaSection, ParsedPathContent
                               (Layer 0 + Layer 4)
    parsers/                  — SectionedParser, TagParser, all setters, YAMLSpecScanner,
                               regex helpers, SwaggerTypable/ValidationBuilder interfaces
                               (Layer 1 + Layer 2 + Layer 3)
    builders/
      schema/                 — SchemaBuilder, SchemaTypable, schemaValidations, bodyTypable,
                               addExtension
                               (Layer 5 schema + Layer 6 schema)
      parameter/              — ParameterBuilder, paramTypable, paramValidations,
                               itemsTypable, itemsValidations
                               (Layer 5 param + Layer 6 param, imports schema)
      response/               — ResponseBuilder, responseTypable, headerValidations
                               (Layer 5 response + Layer 6 response, imports schema)
      operation/              — OperationsBuilder (Layer 6, standalone)
      route/                  — RoutesBuilder (Layer 6, standalone)
      meta/                   — meta parsing helpers, setMetaSingle (Layer 6, standalone)
      spec/                   — SpecBuilder (Layer 7, imports everything)
```

Dependency flow (all arrows point UP = no cycles):
```
scanner ← parsers ← schema ← parameter
                            ← response
                   ← operation
                   ← route
                   ← meta
                                     ← spec (imports all builders)
```
