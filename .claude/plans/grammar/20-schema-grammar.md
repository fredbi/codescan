# SchemaGrammar — schema-bearing annotations and their fields

**Imports from `SharedGrammar`:**
`PrefixTag`, `Prelude`, `ProseLine`, `BlankLine`, `Whitespace`, `LF`,
`Letter`, `Digit`, `IdentifierName`, `ItemsPrefix`,
`ExtensionsBlock`, `ExternalDocsBlock`, all value categories.

**Imports from `ClassifierGrammar`:** `EnumValueList` (reused as the
RHS of the `enum:` keyword body — see §"Validations" and Q14).

## Scope

`SchemaGrammar` defines the bodies of the three schema-bearing
annotations and the `UnboundBlock` that field docstrings produce:

| Block kind          | Annotation header               |
|---------------------|---------------------------------|
| `SchemaBlock`       | `swagger:model`                 |
| `SchemaBlock`       | `swagger:parameters`            |
| `SchemaBlock`       | `swagger:response`              |
| `UnboundBlock`      | (none — struct-field docstring) |

Schema-bearing means the body carries JSON Schema draft-4 validations
and OAS schema decorators. **No operation/meta keywords are legal in
`SchemaGrammar`.**

## Block productions

```ebnf
SchemaBlock      = Prelude , SchemaAnnotation , SchemaAnnotationBody ;

UnboundBlock     = Prelude , UnboundBlockBody ;
                  (* No annotation. Produced by member-level Go
                     declarations: struct fields, interface methods,
                     individual `const` / `var` declarations — i.e.
                     anything *not* the godoc headline of a top-level
                     type/function. Although a struct under
                     `swagger:model` is the canonical case, an
                     UnboundBlock may appear under any schema-bearing
                     parent annotation (e.g. const docstrings under a
                     `swagger:enum`-tagged type). The grammar is the
                     same regardless of the enclosing parent; the
                     analyzer determines the parent context from the
                     Go AST. *)
```

The two productions distinguish *contexts* even though they share
the same body items (see below). The split lets the analyzer apply
context-specific rules — e.g. validations are typically meaningful
under `UnboundBlockBody` (per-member) but vacuous under
`SchemaAnnotationBody` (per-annotation). The grammar **recognises**
both eagerly; legality is the analyzer's call.

## Annotation headers

Each schema-bearing annotation has its own header production so the
binding role of its name argument is explicit (`ModelName` /
`ResponseName` introduce names into their respective namespaces;
`OperationIDRef` references an operation-id introduced elsewhere).

```ebnf
SchemaAnnotation     = ModelAnnotation
                     | ResponseAnnotation
                     | ParametersAnnotation
                     | NameAnnotation ;

ModelAnnotation      = PrefixTag , "model" ,
                       [ Whitespace , ModelName ] , LF ;
                      (* Omitted → name inferred from the enclosing
                         Go type declaration. *)

ResponseAnnotation   = PrefixTag , "response" ,
                       [ Whitespace , ResponseName ] , LF ;
                      (* Omitted → name inferred from the enclosing
                         Go type declaration. *)

ParametersAnnotation = PrefixTag , "parameters" , Whitespace ,
                       OperationIDRef , { Whitespace , OperationIDRef } , LF ;
                      (* One or more references to operation IDs the
                         parameter set binds to. *)

NameAnnotation       = PrefixTag , "name" , Whitespace , MemberName , LF ;
                      (* Field-level rename. Dispatches under the
                         schema family because its body accepts the
                         full SchemaAnnotationBody (validation
                         keywords, raw blocks, etc.). v1 didn't
                         distinguish; grammar2 P7/S3 promoted it from
                         the classifier family for byte-stable parity
                         with v1's lenient interface-method handling.
                         See fixtures/.../classification/models for
                         the canonical interface-method shape. *)

(* --- name productions: defining vs referencing --- *)

ModelName            = IdentifierName ;        (* def: Model namespace *)
ResponseName         = IdentifierName ;        (* def: Response namespace *)
OperationIDRef       = IdentifierName ;        (* ref: OperationID namespace *)
MemberName           = IdentifierName ;        (* def: MemberName (per-container scope) *)
```

## Bodies

Both `SchemaAnnotationBody` and `UnboundBlockBody` recognise the same
set of items — the optimistic-parsing stance: parse what the line
looks like, let the analyzer decide whether it is meaningful in this
context.

```ebnf
SchemaAnnotationBody = { SchemaBodyItem } ;

UnboundBlockBody     = { SchemaBodyItem } ;

SchemaBodyItem       = Validation
                     | SchemaDecorator
                     | ExtensionsBlock
                     | ExternalDocsBlock
                     | ProseLine | BlankLine ;
```

## Validations (JSON Schema draft-4 + per-field markers)

`required:` and `readOnly:` join the validation family rather than the
decorator family — they constrain *whether* / *how* the value may be
present, not how it is rendered. See Q13.

```ebnf
Validation       = NumericValidation
                 | StringValidation
                 | ArrayValidation
                 | EnumValidation
                 | RequiredLine
                 | ReadOnlyLine ;

NumericValidation = [ ItemsPrefix ] , NumericKw ,
                    ":" , [ Whitespace ] , NumberValue , LF ;
NumericKw         = "maximum" | "max"
                  | "minimum" | "min"
                  | "multipleOf" | "multiple of" | "multiple-of" ;

StringValidation  = [ ItemsPrefix ] , StringKw ,
                    ":" , [ Whitespace ] , StringValidationValue , LF ;
StringKw          = MaxLengthKw | MinLengthKw | "pattern" ;
MaxLengthKw       = "maxLength" | "max length" | "max-length"
                  | "maxLen"   | "max len"   | "max-len"
                  | "maximum length" | "maximum-length"
                  | "maximumLength"
                  | "maximum len"    | "maximum-len" ;
MinLengthKw       = "minLength" | "min length" | "min-length"
                  | "minLen"   | "min len"   | "min-len"
                  | "minimum length" | "minimum-length"
                  | "minimumLength"
                  | "minimum len"    | "minimum-len" ;
StringValidationValue
                  = IntegerValue              (* for max/minLength *)
                  | StringValue ;             (* for pattern *)

ArrayValidation   = [ ItemsPrefix ] , ArrayKw ,
                    ":" , [ Whitespace ] , ArrayValidationValue , LF ;
ArrayKw           = MaxItemsKw | MinItemsKw | "unique" | CollectionFormatKw ;
MaxItemsKw        = "maxItems" | "max items" | "max-items" | "max.items"
                  | "maximum items" | "maximum-items" | "maximumItems" ;
MinItemsKw        = "minItems" | "min items" | "min-items" | "min.items"
                  | "minimum items" | "minimum-items" | "minimumItems" ;
CollectionFormatKw = "collectionFormat" | "collection format"
                   | "collection-format" ;
ArrayValidationValue
                  = IntegerValue              (* maxItems, minItems *)
                  | BooleanValue              (* unique *)
                  | StringEnumValue ;         (* collectionFormat *)

EnumValidation    = [ ItemsPrefix ] , "enum" ,
                    ":" , [ Whitespace ] , RawValueBody ;
                  (* `RawValueBody` is a lexer-emitted token — see
                     `00-overview.md` §"Grammar / parser boundary"
                     and Q15. The lexer accumulates body content
                     until the next sibling structural item, then a
                     downstream sub-parser interprets the captured
                     bytes as `EnumValueList` (aligned with
                     `swagger:enum`'s value list — see Q14). The
                     keyword targets an anonymous schema (the
                     enclosing field/property); no name component. *)

RequiredLine      = "required" , ":" , [ Whitespace ] , BooleanValue , LF ;
                  (* Field-level boolean marker. See Q13. *)

ReadOnlyLine      = ReadOnlyKw , ":" , [ Whitespace ] , BooleanValue , LF ;
ReadOnlyKw        = "readOnly" | "read only" | "read-only" ;
```

## Schema decorators (OAS-specific, non-validation)

```ebnf
SchemaDecorator   = DefaultLine
                  | ExampleLine
                  | DiscriminatorLine
                  | DeprecatedLine ;

DefaultLine       = [ ItemsPrefix ] , "default"       , ":" , [ Whitespace ] , RawValueBody ;
ExampleLine       = [ ItemsPrefix ] , "example"       , ":" , [ Whitespace ] , RawValueBody ;
DiscriminatorLine =                    "discriminator", ":" , [ Whitespace ] , BooleanValue , LF ;
DeprecatedLine    =                    "deprecated"   , ":" , [ Whitespace ] , BooleanValue , LF ;
                  (* Also legal under OperationFamilyBlock — see
                     21-operation-grammar.md's OperationDecorator. *)
                  (* Field-level boolean marker. See Q13. *)

(* `default:` and `example:` carry value bodies that v1 accepts as
   either a single-line literal or a multi-line block (object/array
   JSON, multi-line YAML, …). The lexer-emitted `RawValueBody` token
   captures the entire body uniformly; downstream sub-parsers
   interpret per the target Go type via
   `internal/parsers/helpers/ParseValueFromSchema`. See Q15. *)
```

`in:` is **not** a schema-grammar production. Although v1 keyed on
the `in:` line at the field level inside `swagger:parameters`, the
keyword is meaningful only there, and the schema grammar models the
field-level body without it. Recognition of `in:` happens in the
parameters dispatch path, outside `SchemaBodyItem`. See Q15.

`KW_DEPRECATED` (the `deprecated:` boolean keyword) is part of
`SchemaDecorator` here. It is **also** legal under
`OperationFamilyBlock` — see `21-operation-grammar.md`'s
`OperationDecorator`. The pre-token-vocabulary grammar split it out
as a `SchemaCrossover` non-terminal; with token-level productions
that level of indirection is no longer needed.

## `ItemsPrefix` semantics

`ItemsPrefix` injects a validation into a **nested array level** of
the target type. The grammar checks only its surface syntax — repeated
`items.` segments — and leaves type-compatibility to the analyzer.

| Annotation                       | Targets the validation at         |
|----------------------------------|-----------------------------------|
| `maxLength: 10`                  | The field itself                  |
| `items.maxLength: 10`            | Elements of `[]T` (one level deep) — `T`'s max length |
| `items.items.maxLength: 10`      | Elements of `[][]T` (two levels deep) |
| `items.items.items.maxLength: 10`| Elements of `[][][]T`             |

The analyzer (see
`internal/builders/schema/bridge.go` `collectItemsLevels`) walks the
declared Go type and matches each `ItemsDepth = N` to the schema at
the right nesting level, emitting an error when the depth doesn't
fit the type. The grammar production:

```ebnf
ItemsPrefix      = ItemsSegment , { ItemsSegment } ;
ItemsSegment     = "items" , "." ;
                  (* Imported from SharedGrammar — see 10-shared.md. *)
```

## Annotation crossover within a single comment block

A Go comment block may contain **more than one** `swagger:` annotation
line. Each new annotation **closes the preceding block** and starts a
fresh one of the new annotation's family. The grammar dispatcher walks
the comment block top-to-bottom, splitting it on every recognised
annotation header.

Practical consequence inside a struct body: a member-level comment
that opens with a classifier annotation (`swagger:strfmt`,
`swagger:enum`, `swagger:default`, `swagger:type`, `swagger:name`,
`swagger:allOf`, `swagger:ignore`) produces a `ClassifierBlock`, **not**
an `UnboundBlock`. The same comment may then contain further
`swagger:` lines that re-classify the remaining content; everything
between two annotations belongs to the earlier one.

A struct under `swagger:model` is therefore a mix of `ClassifierBlock`
(member with a classifier annotation), `UnboundBlock` (plain member
docstring, no annotation), and the parent `SchemaAnnotationBody`.

## Notes

- Schema-bearing bodies have no traditional `RawBlock` (no
  `consumes:` / `produces:` / etc.). The multi-line shapes in this
  grammar are:
  - `ExtensionsBlock` / `ExternalDocsBlock` — specified in `SharedGrammar`,
  - `RawValueBody` (used by `default:`, `example:`, `enum:`) — a new
    lexer-emitted body terminal documented in `00-overview.md`.
- Every `Validation` and most `SchemaDecorator` keywords carry a
  `ValueSpec`-shaped argument (see `10-shared.md`). The exact lexical
  type per keyword is given by the keyword's own RHS production.
- `required:` and `discriminator:` are **field-level boolean markers**
  in v1, not their JSON Schema / OAS schema-level forms. The grammar
  models the boolean form only — see Q13. Schema-level string-list
  `required:` and string-valued `discriminator:` are tracked as v2
  enhancements (§"Out of scope — v2 enhancements" below).

## Out of scope — v2 enhancements

The following keyword and argument forms are **not** in the round-1
grammar. They are noted here so v2 can revisit:

| Item                              | Round-1 stance                              | v2 motivation                                                            |
|-----------------------------------|---------------------------------------------|--------------------------------------------------------------------------|
| `type:` body keyword              | Not a body keyword — covered by `swagger:type` classifier annotation only | JSON Schema-aligned consistency; let body locally override inferred type |
| `format:` body keyword            | Not a body keyword — covered by `swagger:strfmt` classifier annotation only | JSON Schema-aligned consistency; inline strfmt without classifier dance  |
| `title:`, `description:`          | Not body keywords — derived from godoc prose | Explicit forms decouple spec text from godoc style                       |
| `allOf:` body keyword             | Not a keyword — composition is via `swagger:allOf` annotation at field/method level | Reassess once the annotation form is stable; may stay annotation-only    |
| `nullable:` / `x-nullable`        | Currently writable via `extensions:` block (or auto-emitted on pointer fields) | Promote to a first-class field marker                                    |
| `required: [a, b]` (string list)  | Boolean field-level marker only             | JSON Schema schema-level form on the parent schema                       |
| `discriminator: <propertyName>`   | Boolean field-level marker only             | OAS schema-level form on the parent schema                               |
| Distinct schema-flavour for `parameters` and response headers | Both reuse `SchemaBlock`; OAS v2's `SimpleSchema` constraints (subset of validations) are not modelled in the grammar | Split `SimpleSchemaBlock` from `SchemaBlock` once OAS v3 (which collapses the distinction) is on the roadmap |
