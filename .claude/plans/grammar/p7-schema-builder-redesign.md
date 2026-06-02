# P7 — Schema Builder Redesign on grammar2

**Status:** draft for review (revised 2026-05-01 — Walker design)
**Author:** Claude (Opus 4.7) + Fred
**Date:** 2026-05-01
**Scope:** `internal/builders/schema/` and the helper migration out of
`internal/parsers/helpers/`. Items / parameters / responses / routes
builders are addressed where they share dispatch with schema (the
`Walker` design forces a four-builder convergence on dispatch).
Operations / meta builders follow the same template separately.

## 0. Headline change vs. v0 of this plan

**v0** proposed a shared `internal/builders/validations/` package owning a
polymorphic `ApplyValidation(p, target, sink)` function. Three bridges
(schema, parameters, responses) would call it via the
`ifaces.ValidationBuilder` interface, plus the items package would
shed its near-duplicate dispatcher.

**v1 (this revision)** replaces that with a **Walker** in grammar2:

```go
type Walker struct {
    Title       func(s string)
    Description func(s string)
    Number      func(p Property, val float64, exclusive bool)
    Integer     func(p Property, val int64)
    Bool        func(p Property, val bool)
    Raw         func(p Property)            // pattern/default/example/enum-raw — caller types it
    Unknown     func(p Property)
    Extension   func(ext Extension)
    Diagnostic  func(d Diagnostic)
    FilterDepth int                          // -1 for all depths; otherwise only fire on Property.ItemsDepth == FilterDepth
}

func (b Block) Walk(w Walker)
```

Why the change: the same dispatch logic was about to land in
`validations.ApplyValidation`. That function was 110 LOC of switch
mirroring grammar2's keyword table — i.e. mirroring the same data the
**lexer already** uses to value-type Properties. A Walker fronts the
table once, in grammar2; consumers wire 4–6 small handler functions
per call site instead of writing the same 110-LOC switch four times.

The existing iterator (`Block.Properties() iter.Seq[Property]`) is
**kept** for ad-hoc reads (LSP "what's at line 42", test introspection).
Walker and iterator coexist for the same reason `ast.Inspect` and
`ast.File.Decls` coexist in `go/ast`.

## 1. Objective

Reduce the bridge layer in `internal/builders/schema/bridge.go` to zero by
exploiting grammar2's richer Block API. Reincorporate the dispatch logic
directly into `Builder`, migrate misplaced helpers out of
`internal/parsers/helpers/`, and add the validation-with-diagnostics
behaviour that the bridge currently lacks.

The new objective is explicitly:

- **Strip the bridge.** Bridge helpers were a v1-grammar parity scaffold;
  with grammar2's typed Blocks the indirection earns its keep nowhere.
- **Inline logic into the Builder.** `applyBlockToField` /
  `applyBlockToDecl` become methods on `Builder`. `bridge.go` ceases to
  exist as a file.
- **Eliminate `parsers/helpers` calls from the schema builder.** Every
  helper currently consumed by the schema builder is either a builder
  concern misplaced under `parsers/` (it should move into `schema/`) or
  a parser concern that grammar2 now owns (it should be deleted).
- **Add validation.** Errors remain non-blocking — invalid constructs
  (e.g. `required: foo` where `foo` is not a boolean) are silently
  dropped from the output spec, but they accumulate as diagnostics so
  the caller can choose to surface them.

## 2. Inventory — what the bridge does today

`internal/builders/schema/bridge.go` is 423 LOC. It does five jobs.

| Concern | LOC | Justification today |
|---|---|---|
| **Dispatch loop** (`applySchemaBlock` + 4 `dispatch*` funcs) | ~110 | Maps `Property.Keyword.Name` × `Typed.Type` to validation setters |
| **Items recursion** (`collectItemsLevels`) | ~35 | Walks AST `*ast.ArrayType` to pair nesting depth → schema for `items.maximum:` etc. |
| **Extensions YAML re-parse** (`applyExtensionsBody`) | ~25 | YAML→JSON pipeline because grammar's `Extensions()` flattens to strings |
| **`required:` / `discriminator:` writers** (`setRequired`, `setDiscriminator`) | ~30 | Mutate enclosing schema's slice |
| **Orchestrators** (`applyBlockToField`, `applyBlockToDecl`) | ~55 | Parse comments, write title/description, dispatch, recurse items |
| **Quirk shims** (`schemeFromPS`) | ~15 | Drop Format from SimpleSchema for `default:` parity |

Helper calls escaping into `parsers/helpers`:

- `JoinDropLast` (3×) — `\n`-join with trailing-blank elision
- `CollectScannerTitleDescription` (1×) — title/description split heuristic
- `ParseValueFromSchema` (2×) — `default:` / `example:` raw → typed value
- `ParseEnum` (1×, in `typable.go`) — comma-split + per-item type coercion
- `GetEnumDesc` / `EnumDescExtension` (3×) — `x-go-enum-desc` extension I/O

## 2bis. Inventory — what the items package does today (newly in scope)

`internal/builders/items/bridge.go` (121 LOC) provides
`ApplyBlock(b grammar.Block, target ifaces.ValidationBuilder, level int)`
— a polymorphic dispatcher consumed by **three** builders:

```
items.ApplyBlock
  ← schema/bridge.go:376       (level ≥ 1, target = *oaispec.Schema)
  ← parameters/bridge.go:121   (level ≥ 1, target = *oaispec.Items)
  ← responses/bridge.go:103    (level ≥ 1, target = *oaispec.Items)
```

Its dispatch table is the items-depth subset of schema's level-0 table
(no `required:` / `discriminator:` / `readOnly:` / `extensions:`, plus
`collectionFormat:` for parameter/response targets via interface
upcast). Today this is 110 LOC of switch in `items/`, near-identical
to the 110 LOC of switch in `schema/bridge.go` for level-0.

`helpers.ParseEnum` is called from **two** sites:

- `schema/typable.go:133` — `schemaValidations.SetEnum`
- `items/validations.go:41` — `items.Validations.SetEnum`

Same function, same `Type+Format` driver, same need.

The Walker subsumes both dispatchers. `items/bridge.go` deletes; the
items package keeps `Typable` + `Validations` (the `*oaispec.Items`
ValidationBuilder impl, still legitimately needed).

## 3. What grammar2 already gives us (vs. v1 grammar)

Compared to v1's flat `Property{Keyword, Value, ItemsDepth, Typed}`,
grammar2 offers:

- **Block typing as a real discriminator.** `ModelBlock`, `ResponseBlock`,
  `ParametersBlock`, `RouteBlock`, etc. The schema-side bridge entry
  points (`applyBlockToField`, `applyBlockToDecl`) currently assume the
  doc is schema-shaped; grammar2 makes that assumption explicit at the
  type level.
- **Diagnostics as a first-class output channel.** `Block.Diagnostics()
  []Diagnostic` — the bridge currently silently ignores invalid
  constructs. Grammar2's lexer already emits `CodeInvalidNumber` /
  `CodeInvalidInteger` / `CodeInvalidBoolean` / `CodeContextInvalid` /
  `CodeInvalidExtension`. We can collect those at the builder level
  instead of inventing parallel checks.
- **`Title()` / `Description()` accessors** on every Block — replaces
  `JoinDropLast(block.ProseLines())` + `CollectScannerTitleDescription`.
- **Typed extensions iterator.** `Block.Extensions() iter.Seq[Extension]`
  already extracts `x-*` lines (currently flat strings). With one
  schema enhancement (preserve YAML-typed values), `applyExtensionsBody`
  disappears.
- **Per-property position fidelity.** Every `Property` carries `Pos`, so
  diagnostics pinpoint the source line. The bridge currently has no
  source-position awareness.

## 4. Target shape — what to inline, what to keep, what to delete

### 4.1 Delete entirely (5 things)

1. **`applyExtensionsBody`** (~25 LOC, `schema/bridge.go`) — replace
   by extending `Block.Extensions()` to yield typed values (`any`,
   not `string`). The YAML→JSON pipeline moves into the lexer's body
   accumulator (it already collects RAW_BLOCK bodies; typing them is
   one more pass). Allowed-extension filtering moves to the builder
   where it belongs (output-policy decision, not parser concern).
   Walker fires `Extension(ext)` per typed extension.
2. **`collectItemsLevels`** (~35 LOC, `schema/bridge.go`) — grammar2
   already carries `ItemsDepth` on each Property. The recursion is
   gone; the schema builder calls `Block.Walk(w)` once per items
   depth, with `Walker.FilterDepth = depth`. The AST walk on
   `*ast.ArrayType` to find each nested `*oaispec.Schema` stays
   (Go-types→spec mapping concern) but produces a flat
   `[]*oaispec.Schema` indexed by depth. Net: ~10 LOC of `flattenItemsTargets`
   in `schema.go`.
3. **`schemeFromPS`** (~15 LOC) — v1 parity shim. For grammar2 we own
   the value-parsing path: `parseDefault(raw, schemaType, schemaFormat)`
   directly with `strconv`. Format-aware where it should be (`int32`
   defaults parse as int32 not float64). Delete the helpers call.
4. **`dispatchSchemaKeyword` + 4 sub-dispatchers** (~110 LOC,
   `schema/bridge.go`) — replaced by 4–6 small Walker handler
   functions wired into a `Walker{}` literal. The dispatch table
   itself moves into grammar2's `Block.Walk`. ~90 LOC saved here;
   ~50 LOC of dispatch table added to grammar2 (paid once).
5. **`items/bridge.go`** (121 LOC) — `ApplyBlock` +
   `dispatchItemsKeyword` are the items-depth twin of #4 and disappear
   for the same reason. The items package retains `Typable`,
   `Validations`, and `errors.go` (~120 LOC kept).

### 4.2 Inline into Builder (keep the logic, remove the indirection)

6. **`applyBlockToField` / `applyBlockToDecl`** become methods on
   `Builder` and live in `schema.go`. They now build a `Walker{}`
   literal and call `block.Walk(w)`. `bridge.go` ceases to exist.
7. **`setRequired` / `setDiscriminator`** stay (logic-bearing,
   ~30 LOC) but as private methods on `Builder`. Currently free
   functions in `bridge.go`. Called from `Walker.Bool`.

### 4.3 Cannot be eliminated — but should move into `schema/`

8. **`ParseEnum`** (`helpers.ParseEnum`, called from
   `schema/typable.go:133` and `items/validations.go:41`) —
   schema-format-aware enum-item coercion. Not parser business:
   `[1, 2, 3]` for a `string[]` schema is type-error territory the
   parser doesn't know about. Move into
   `internal/builders/schema/enum_value.go`. The items-side caller
   becomes a Walker.Raw handler in the consumer that already imports
   schema (parameters/responses do **not** import schema today —
   resolved by promoting the helper to a small standalone func in
   either `internal/builders/resolvers/` or — preferred — a tiny
   `internal/builders/enumcoerce/` package consumable by both items
   and schema callers without a circular dep). **Decision needed.**
9. **`ParseValueFromSchema`** (`default:` / `example:`) — same
   justification. The value's *type* depends on the resolved Go type,
   which only the builder has. Move into the same place
   `ParseEnum` lands.
10. **`x-go-enum-desc`** read/write — `EnumDescExtension()` constant +
    `GetEnumDesc` accessor. This is a schema-builder concept (it's
    ours, not OpenAPI's). Move the constant into
    `internal/builders/schema/extensions.go`.

### 4.4 New responsibility — diagnostics

11. **Builder collects parser diagnostics.** Today: silently dropped.
    Target: `Builder` exposes `Diagnostics() []grammar2.Diagnostic`,
    populated by every `Walk` call (via `Walker.Diagnostic`). Two
    policies:

    - `Severity == SevError`: skip the offending property (matches
      "errors remain non-blocking, invalid construct ignored from
      output").
    - All severities: bubbled up to `*Options.OnDiagnostic` callback
      (new field) or accumulated for top-level reporting.

    This is **the new logic** — ~20 LOC plumbing, no parallel
    validation code because the parser already validates.

### 4.5 New responsibility — builder-side validation (what the parser cannot do)

The parser validates *lexical / syntactic* shape. The builder still
validates *semantic* shape against the resolved Go type:

- `pattern: ^...$` only valid on string-typed schemas (parser doesn't
  know the type).
- `multipleOf: 5` invalid on a schema typed as `string` (same).
- `enum:` items must coerce to schema type.
- `required: true` on a property whose enclosing isn't an object schema.

This is ~40 LOC of new Builder code, replacing zero LOC today. Net add,
but well-justified — the parser cannot and should not know the resolved
Go type.

## 5. Net file shape — proposal

```
internal/parsers/grammar2/
├── walker.go            (~120 LOC, new — Walker struct + Block.Walk + dispatch table)
└── (everything else unchanged — additive change)

internal/builders/schema/
├── schema.go            (~1330 LOC, was 1368 — slight shrink from inlined orchestrators)
├── extensions.go        (~40 LOC, new — owns x-go-enum-desc constant + filter + add)
├── diagnostics.go       (~25 LOC, new — Builder.Diagnostics + recordDiag)
├── typable.go           (~120 LOC, was 149 — SetEnum simplifies once dispatch is via Walker)
├── errors.go            (unchanged, 9 LOC)
├── schema_test.go, …    (existing tests stay)
└── bridge.go            DELETED (was 423 LOC)

internal/builders/items/
├── typable.go           (unchanged, 59 LOC — legitimate Items typable provider)
├── validations.go       (~40 LOC, was 44 — SetEnum simplifies; uses validations/coerce)
├── errors.go            (unchanged, 15 LOC)
└── bridge.go            RETAINED (was 121 LOC) — kept alive for parameters/responses on grammar v1; deletes in P8 once they migrate

internal/builders/validations/  (new — see §4.3 #8 + Q3/Q4 decisions)
├── coerce.go            (~130 LOC — Enum + ParseValueFromSchema, format-aware)
├── shape.go             (~80 LOC, S5.5 — IsLegalForType keyword × type/format)
└── context.go           (~60 LOC, S5.5 — IsLegalIn keyword × ContextKind)

internal/builders/parameters/bridge.go   (UNCHANGED through S1–S7 — still on grammar v1; migrates in P8)
internal/builders/responses/bridge.go    (UNCHANGED through S1–S7 — same)
```

`internal/parsers/helpers/` shrinks dramatically:

- `enum.go`, `body.go` (ParseValueFromSchema) migrate to
  `internal/builders/enumcoerce/`.
- `title_desc.go` (CollectScannerTitleDescription) — replaced by
  `Block.Title()` / `Block.Description()` returning joined strings.
- `JoinDropLast` (lines.go) — same; replaced by `Block.Description()`.
- After migration, `helpers/` may have 0–50 LOC left (consumed only by
  operations/meta builders, not yet in scope).

> Note: this frees the **schema, items, parameters, and responses**
> builders from `parsers/helpers`. Operations / meta builders are not
> in scope and need their own P7-style passes before `helpers/` can be
> deleted entirely.

## 6. Sequencing — incremental migration

To avoid a megacommit, phase the work. Each phase ships independently
and passes the test suite before the next starts.

| Phase | Action | Touches | Ship-ready? |
|---|---|---|---|
| **S1** | Add `Walker` + `Block.Walk(Walker)` to grammar2. Add `Property.IsTyped()`. Add `Block.Title()` / `Block.Description()` returning joined strings. Switch `Block.Extensions()` to yield typed values. Cover with unit tests, no consumer changes yet. | grammar2 only (additive) | yes — bridges unchanged |
| **S2** | Create `internal/builders/validations/coerce.go`. Move `helpers.ParseEnum` → `validations.Enum`; move `helpers.ParseValueFromSchema` → `validations.Value`. Update `schema/typable.go` and `items/validations.go` to import from the new package. | helpers shrinks, new package | yes |
| **S3** | Schema-side: inline `applyBlockToField` / `applyBlockToDecl` as Builder methods using `block.Walk(Walker{...})`. Delete `applyExtensionsBody` (Walker.Extension subsumes it). Delete the 4 schema dispatchers. Move `EnumDesc*` constants + `clearStaleEnumDesc` strip helper to `schema/extensions.go`. **Fix the v1 Format-drop bug** as part of this phase: replace `schemeFromPS` + `helpers.ParseValueFromSchema` with a `parseDefault(type, format, raw)` that handles `float`/`float64`/`float32` correctly. Add a regression test on `fixtures/enhancements/defaults-examples/types.go`. | bridge.go → schema.go | yes |
| **S4** | Schema-side items: schema stops calling `items.ApplyBlock`. Replace `collectItemsLevels` with `flattenItemsTargets` returning `[]*oaispec.Schema` (depth-indexed). Schema invokes `block.Walk(Walker{FilterDepth: N, ...})` per nesting level for its own items walk. **`items/bridge.go` is retained**: parameters and responses remain on grammar v1 + `items.ApplyBlock` through P8. | schema only; items/bridge.go untouched | yes |
| **S5** | Add `Builder.Diagnostics()` + `Options.OnDiagnostic` plumbing; wire grammar2 diagnostics through `Walker.Diagnostic`. | new feature | yes |
| **S5.5** | Build out `internal/builders/validations/`: add `shape.go` (`IsLegalForType(keyword, schemaType, schemaFormat)`) and `context.go` (`IsLegalIn(keyword, ctx)`). Port semantic-rule unit tests. No consumer wiring yet. | new package | yes |
| **S6** | Wire `validations.IsLegalForType` / `IsLegalIn` into each builder's Walker callbacks. Failed checks fire `Walker.Diagnostic` with `CodeShapeMismatch` / `CodeContextInvalid` and drop the property. | wiring only | yes |

After S4, `schema/bridge.go` is gone and the schema builder is fully on
grammar2 (level-0 + items walks). `items/bridge.go` is **retained** —
parameters and responses still consume `items.ApplyBlock` through
grammar v1 until P8 migrates them. `helpers/` retains everything the
non-schema builders still consume.

After S6, semantic validation is wired and diagnostics flow on the
schema builder. Behaviour matches v1 baseline byte-for-byte (golden
fixtures unchanged); the only new observable is the diagnostic stream
on schema-emitted output.

**S7 (polishing)** is post-bulk and produces intentional golden drift
on three v1 bugs. See §6.1 for the bug list and the shared
`mergeIntoAllOf(ps, override)` helper that two of the three fixes
ride on.

## 6.2 Partial-migration invariant

Through S1–S7, the codebase is in a **mixed-grammar state** by design:

- Schema builder: grammar2 + Walker (post-S3) + `validations/`.
- Operations / parameters / responses / routes / meta builders:
  unchanged — still on grammar v1, still consuming
  `parsers/helpers/`, still calling `items.ApplyBlock` (parameters
  and responses).

Both grammars run in the same process during a single `codescan.Run`
call, parsing the same comment groups independently. They have
disjoint state (each builder constructs its own parser per call),
disjoint output targets, and any spec-field overlap produces
identical writes — so the integration suite stays green throughout.

Subsequent P8/P9/etc. migrate the remaining builders one at a time;
each follows the same S1–S7 template scoped to its bridge. Once all
builders are on grammar2, `parsers/grammar/`, `parsers/helpers/`, and
`items/bridge.go` delete in a final cleanup pass.

## 6.1 Polishing phase (post-migration v1-bug fixes)

After the bulk migration (S1–S6) lands, a follow-up phase **S7
("polishing")** addresses three v1 bugs that were classified as quirks
in earlier drafts but are now scheduled for proper resolution. All
three are deferred to keep the bulk migration scope tight; they touch
golden fixtures so they can land independently.

| Bug | Fix | Effort | Fixture impact |
|---|---|---|---|
| **Format-drop default-coercion** | In S3 already — replace `helpers.ParseValueFromSchema` + `schemeFromPS` with `parseDefault(type, format, raw)`, dispatching on `type` first with format-aware branches where helpful. Drop the v1 `TypeName()` Format-precedence accident. | Small (~30 LOC + test) | `defaults-examples` (float32 case will produce `1.5` as number instead of string `"1.5"`). |
| **Field-level enum override on a $ref'd field** | Produce an `allOf` compound: arm 1 = `{ "$ref": "#/definitions/Parent" }`, arm 2 = inline schema with the field-level `enum`. Solves the JSON-Schema-draft-4 sibling-override constraint correctly. | Medium (~80–120 LOC + test) | `enum-overrides` Case E — the rendered schema for `NotificationE.Priority` becomes `{allOf: [{$ref: PriorityE}, {enum: [urgent, normal]}]}` instead of inheriting then stripping. |
| **Field-level validation on any $ref'd field (`example`, `default`, `pattern`, etc.)** | Same `allOf` compound shape as above, generalised across any field-level keyword that conflicts with the parent $ref. Replaces today's `DescWithRef=false` short-circuit-drop and the `DescWithRef=true` silent-sibling-merge. | Medium-to-large (~120–180 LOC + tests) | `bugs/2540` — Author field will produce `{allOf: [{$ref: Author}, {example: {Name: Tolkien}}]}` rather than dropping the example or attaching it as a sibling. |

These three sit in a phase **S7** because they share two properties:

1. They produce **justifiable golden-fixture drift** — the rendered
   spec changes for a small set of fixtures, but each diff has a
   reasoned explanation tied to JSON-Schema-draft-4 semantics.
2. They share a common helper: a `mergeIntoAllOf(ps, override)` builder
   that takes a $ref'd schema and an override fragment, producing the
   two-armed allOf. The first bug (Format-drop) doesn't need this
   helper; the latter two do — so once the helper lands, both fixes
   ride on it.

**S7 stays out of the critical path.** S1–S6 land first, all behaviour
matches v1 baseline (modulo the now-active diagnostics from S5–S6),
and golden fixtures are unchanged. S7 is announced as a planned drift
in its own PR, with a per-fixture diff explanation.

**Deferred to the S7 implementation pass** (these don't block S1–S6
since S7 is post-bulk; record here so they aren't lost):

- **Anonymous-vs-named field distinction.** Per fred's annotation,
  anonymous-typed fields (`string` / `int` / `struct{…}`) are handled
  correctly in v1 — the produced schema is inlined with an inline
  enum/validation, no $ref involved. The allOf compound applies only
  to named-typed fields whose schema is rendered as `$ref`. S7's
  $ref-override fix must gate on field-type kind (named vs.
  anonymous) before invoking `mergeIntoAllOf`.
- **`DescWithRef` option fate.** Three candidate dispositions: drop
  the field outright (public-API break), keep as no-op with
  deprecation godoc, or repurpose as a force-flatten escape hatch.
  Decide during the S7 PR with a short ADR; default lean is
  no-op + deprecation (lowest blast radius).
- **`mergeIntoAllOf(ps, override)` helper signature.** Placeholder
  in this plan — designed during the S7 PR alongside the first
  fixture rewrite. The signature will likely take the existing
  schema (with `$ref` set) plus an override fragment, and return
  the rewritten parent with `AllOf: [{$ref}, {override}]`. Edge
  cases (existing AllOf on the parent, nested arrays, items chains)
  shaken out during implementation.

These three items are explicitly **out of scope for S1–S6** and
are recorded here as a hand-off to whoever picks up S7.

## 7. Caveats worth flagging

### 7.1 Default/example value-parsing call ordering

The bridge currently routes `default:` / `example:` through
`helpers.ParseValueFromSchema` *before* the schema's resolved type is
settled in some construction orders. If we inline this into Builder, we
need to be sure the call site fires after `buildFromType` has populated
`ps.Type` / `ps.Format`. Spot-checked the three call sites in
`schema.go` (lines 714, 939, 1192 — all after type construction); should
be fine, but worth a small invariant test in S3.

### 7.2 Extension typing change is a parser API change

S1's "yield typed values" change to `Block.Extensions()` widens the
return type from `string` to `any`. That's a breaking signature change.
Since grammar2 has no external consumers yet (the bridge is the only
caller), the cost is contained. Worth landing this in S1 cleanly with a
minor godoc note rather than introducing a parallel `ExtensionsTyped()`
accessor that we'd have to deprecate later.

### 7.3 `Options.OnDiagnostic` is a new public API surface

Adding a callback field to `Options` is a public-API change. We may
want to gate it behind v2 of the public package, or add it as
`(experimental: subject to change)` in godoc and stabilise once LSP
diagnostics are wired up. **Decision needed before S5.**

### 7.4 v1-parity quirks — disposition

What started as "quirks to preserve" turned out, on inspection, to be
v1 bugs. Three are scheduled for fix (see §6.1 / S7 polishing phase);
none are preserved as-is across S1–S6 except for short-circuiting
behaviour where deletion would itself break fixtures during the
intermediate phases.

The annotated discussion below is preserved as the audit trail behind
the §6.1 disposition.

- **`schemeFromPS` Format-drop.** Float32 fields with `default: 1.5`
  rely on the SimpleSchema constructed without Format. The replacement
  `parseDefault` must reproduce this for the `number` / `integer` cases
  (no format-specific paths) until we deliberately fix it as a separate
  PR. Track as a follow-up in S3.

  > **fred** I don't understand this quirk.
  >
  > Could you please elaborate a bit with an example of such annotation?
  >
  > Do we have already a fixture that exhibits this?

  > **claude (2026-05-01):** Walkthrough with the actual code paths.
  >
  > `helpers.ParseValueFromSchema(s, schema)` (in `parsers/helpers/enum.go:33`)
  > switches on `strings.Trim(schema.TypeName(), "\"")`. The
  > `SimpleSchema.TypeName()` accessor in `go-openapi/spec` returns
  > **`Format` if non-empty, otherwise `Type`** — i.e. Format wins.
  > The switch arms are: `integer/int/int64/int32/int16` → `Atoi`,
  > `bool/boolean` → `ParseBool`, `number/float64/float32` →
  > `ParseFloat`, `object` → JSON-unmarshal map, `array` → JSON-unmarshal
  > slice, otherwise raw string.
  >
  > Concrete fixture: `fixtures/enhancements/defaults-examples/types.go`
  > defines `Ratio float32 \`json:"ratio"\`` with `// default: 1.5`. After
  > `buildFromType`, the property schema has `Type=["number"]`,
  > `Format="float"`. If `schemeFromPS` passed Format through, then
  > `TypeName()` → `"float"` → switch falls through to default →
  > `default` is stored as the literal string `"1.5"` instead of the
  > number `1.5`. By dropping Format, `TypeName()` → `"number"` → matches
  > the float case → ParseFloat → stored as `float64(1.5)`. (Note: the
  > switch does include `float32` and `float64` literally as arms, but
  > NOT `float` alone, which is the canonical Swagger format suffix for
  > a Go `float32` mapped to OpenAPI's `number` type.)
  >
  > Same goes for `int32` — Go `int32` field maps to `Type=integer`,
  > `Format=int32`. The switch DOES have an `int32` arm, so dropping
  > Format isn't strictly necessary in that case; it just doesn't harm.
  > The quirk is load-bearing specifically for `float`-format numeric
  > fields. Verified against
  > `fixtures/enhancements/defaults-examples/types.go` line 14–17.
  >
  > Disposition: not a hidden invariant — it's a v1 oversight in the
  > `TypeName()` semantics that we should fix, not preserve. The
  > replacement `parseDefault` in S3 should accept `(type, format)`
  > separately and dispatch on `type` first (with format-aware paths
  > only where they help, e.g. `int32` → int32 not int64). I'll demote
  > this from "quirk to preserve" to "v1 bug to fix in S3 with a
  > regression test on the float32 fixture".

  **fred** now I understand and agree that this is not a quirk to maintain but a bug to fix in our version.
  During the test stage, we'll expect a justifiable difference on the rendring of this example.

- **Enum-desc strip on field-level enum override.** `typable.go:142`
  has a stale-suffix removal that runs when a field-level `enum:`
  overrides a type-level `swagger:enum`. The behaviour is correct but
  load-bearing — ensure the test fixture in `fixtures/enhancements/`
  for this case still passes.

  > **fred** could you please point me to the exact fixture that exhibits this behavior?
  >
  > [example struct snipped]
  >
  > I don't think this behaviour is currently supported and I am fine with leaving this kind of special case to be handled in v2 only.

  > **claude (2026-05-01):** Push back: the behaviour IS supported in v1
  > and there's a fixture. Your example is a different shape — let me
  > distinguish.
  >
  > **Your example** (`swagger:enum fool` on a struct type) is the
  > "swagger:enum on a struct" case. Today's parser doesn't accept
  > swagger:enum on struct types — swagger:enum binds to a type alias
  > (`type Priority string`) whose const declarations supply the values.
  > Your interpretation (use the field-level `enum:` to populate the
  > type-level enum) is novel and v2-territory. Skip it.
  >
  > **The actual quirk** is in
  > `fixtures/enhancements/enum-overrides/types.go` line 88–112 — Case E:
  >
  > ```go
  > // swagger:enum PriorityE
  > type PriorityE string
  >
  > const (
  >     PriorityELow  PriorityE = "low"
  >     PriorityEMed  PriorityE = "medium"
  >     PriorityEHigh PriorityE = "high"
  > )
  >
  > // swagger:model NotificationE
  > type NotificationE struct {
  >     // Inline enum provides a narrower set than the const block.
  >     //
  >     // enum: urgent, normal
  >     Priority PriorityE `json:"priority"`
  > }
  > ```
  >
  > Type-level scan finds PriorityE's three consts (low/medium/high) and
  > stores them on the schema definition; it also writes a description
  > to PriorityE explaining each value, plus the same text into
  > `x-go-enum-desc`. When the `Priority` field on NotificationE is
  > built, it inherits PriorityE's schema fragments (including the enum
  > [low, medium, high] and the x-go-enum-desc). Then the field-level
  > `enum: urgent, normal` overrides — but the inherited
  > x-go-enum-desc + the appended description suffix on the field's own
  > Description still describe low/medium/high. That's stale and
  > misleading: the rendered docs would say "Priority must be urgent or
  > normal — Low: …, Medium: …, High: …".
  >
  > `typable.go:142` is the strip: when SetEnum runs, if the current
  > schema has an x-go-enum-desc, drop it AND drop the description
  > suffix that matches it.
  >
  > Disposition for P7 (initial draft): preserve the strip in
  > `schemaValidations.SetEnum` via a `clearStaleEnumDesc(ps)` helper,
  > don't refactor.

  > **claude (revised after fred's annotation):** Promoted from
  > "preserve" to "**fix in S7 polishing**". The strip is treating
  > the symptom (stale x-go-enum-desc) of a deeper modelling error
  > (sibling validation overriding a $ref'd schema). The right model
  > is the allOf compound described in §6.1. For S1–S6 the strip
  > stays in place to keep golden fixtures green; in S7 the strip
  > deletes alongside the broader $ref-override fix.

  **fred** now I understand and I see this as a bug in v1 when such inconsistent"
  annotations are found.
  
  I say this is inconsistent because if we have another type "model" referring to the
  same enum, then we'll have another (possibly not-deterministic) enum override.

  So we have 2 bugs (not mere quirks to be maintained):
  * one is that field-level values are not appropriately interpreted.
      * if the field is an anonymous type (e.g. string, int, struct{...}):
         then the produced schema for this field is anonymous (i.e. inlined schema in json),
         with an enum validation. This case should be handled properly already in v1.
     * if the field is a named type, then:
        * if we map it to $ref then we produce an allOf construct:
          * schema, with allOf, then 1 arm with the $ref
          * one arm: empty schema with the "enum" defined at the field level
          * in your example, this would produce a schema that is never validated (field level values have no
            intersection with the values defined at the parent level) and that's okay because this is a mis-specification.
          * another example in which the field-level values define a subset of the parent work ok.
        * if we inline it treat as the above case for anonymous
  * the other is that the "x-go-enum-desc" extension should remain aligned with the
    values of the enum validation.

  The go-enum-desc should be fixed with the migration.
  The allOf fix is complex and we may defer to a "polishing" stage after the bulk of the migration.
  This brings a significant code change, but would only trigger in a handful of tests.
  During the test stage, we'll expect a justifiable difference on the rendring of this example.

- **`refSchemaTaggers` $ref-mode.** When `ps.Ref` is set and
  `DescWithRef=false`, only `required:` applies. `applyBlockToField`
  has this short-circuit at lines 344–351; preserve it as the first
  branch of the inlined method.

  > **fred** Again a code example would be welcome.

  > **claude (2026-05-01):** Concrete fixture:
  > `fixtures/bugs/2540/foo/examples.go`. Tested by
  > `TestIssue2540` in `internal/builders/schema/schema_test.go:1424`.
  >
  > The fixture contains:
  >
  > ```go
  > // swagger:model
  > type Book struct {
  >     // example: 2021
  >     Published int
  >     // example: { "Name": "Tolkien" }
  >     Author Author     // <- field with $ref: #/definitions/Author
  > }
  > ```
  >
  > `Author` is a struct, so the `Author` field on `Book` is a $ref to
  > the Author definition (not an inline schema). The field-level
  > comment carries `example: { "Name": "Tolkien" }`. The question is:
  > does the `example:` apply to the $ref?
  >
  > JSON Schema draft-4 says no — `$ref` predates and replaces every
  > sibling keyword on the schema fragment. v1 honors this by default:
  > with `DescWithRef=false` (the default), every keyword on a $ref'd
  > field is dropped EXCEPT `required:` (which writes to the enclosing
  > schema's required slice, not to the $ref'd schema, so it's safe).
  > The `TestIssue2540` first sub-test (`descWithRef=false`) asserts the
  > Author field has only `$ref` — no example. The second
  > (`descWithRef=true`) asserts Author has both `$ref` AND
  > `example: { "Name": "Tolkien" }`.
  >
  > Disposition for P7 (initial draft): preserve the inlined
  > short-circuit in `applyBlockToField`, walking only `required:`.

  > **claude (revised after fred's annotation):** Same trajectory as
  > the enum-override case above — promoted to "**fix in S7
  > polishing**". S1–S6 preserve the existing `DescWithRef` branch so
  > golden fixtures stay green; S7 replaces both the
  > `DescWithRef=false` drop and the `DescWithRef=true` sibling-merge
  > with a single allOf compound that respects JSON-Schema-draft-4
  > semantics. The `DescWithRef` option becomes a no-op (or is
  > deprecated) once the allOf model is in place — its raison d'être
  > was to opt into the broken sibling-merge behaviour, which the new
  > model makes correct unconditionally.

  **fred** this is of the same nature as the enum override above.
  It applies to ANY validation with a field level explicit validation on something that points to a $ref.
  In JSON schema draft-4 the $ref always predates its sibling (more recent jsonschema specs have a slightly
  different spec). The right way to model this is to produce a allOf compound with 2 arms: 1 with the $ref
  and 1 with the validation overrides.

  This is a complex fix that we may defer to a "polishing" stage after the bulk of the migration.
  During the test stage, we'll expect a justifiable difference on the rendring of this example.

## 8. Validation strategy

Each phase must pass:

- `go test ./...` — no regressions in existing fixtures.
- `golangci-lint run --new-from-rev master` — no new lint findings.
- `UPDATE_GOLDEN=1 go test ./...` — golden fixtures unchanged.

Phases S5 and S6 add new test cases:

- S5: a fixture with deliberately invalid `required: foo` produces a
  non-empty `Builder.Diagnostics()` and the property is dropped from
  the output spec.
- S6: a fixture with `pattern: ^...$` on an integer field produces a
  diagnostic and the pattern is dropped.

## 9. Out of scope

- Operations / parameters / responses / routes / meta builders. Each
  needs its own redesign doc following the same template. They are
  the prerequisite to deleting `internal/parsers/helpers/` entirely.
- Lexer/parser refactor. Grammar2 is treated as stable for this work
  (only additive changes in S1).
- Public API redesign of `codescan.Options`. The `OnDiagnostic`
  callback is the only addition contemplated; broader `Options`
  changes are v2-territory.
- The v2 public package surface (re-exports, package layout). Tracked
  separately under `project_v2_vision`.

## 9.1 Cross-builder factorisations to consider after P7

Once all builders have migrated to grammar2 + Walker, several
schema-local helpers will likely have direct counterparts in the
other builders. Hold these off until duplication is concrete (don't
abstract from one site); fold once parameters/responses/etc. land
their own versions:

- **`getEnumDesc` / `clearStaleEnumDesc` / `extEnumDesc`** in
  `schema/extensions.go` — the `x-go-enum-desc` extension is a
  go-swagger concept and applies wherever an enum-typed field can
  appear. Move to `internal/builders/resolvers/` (which already
  houses cross-builder concerns like `AddExtension`) once at least
  one other builder needs it.
- **Diagnostic plumbing** — `Builder.Diagnostics()` +
  `recordDiagnostic` will repeat verbatim in every builder. Consider
  a shared base type (e.g. `resolvers.Diagnosticable` or a
  `resolvers.DiagnosticSink` mixin) that builders embed. The exact
  shape — interface, embedded struct, generic — should be picked
  when the second builder lands its diagnostic plumbing.
- **`splitTitleDesc` / `joinDropLast` / `endsWithUnicodePunct` /
  `isATXHeading` / `stripATXMarker`** in `schema/title_desc.go` —
  the v1 SectionedParser title/description split is a cross-builder
  concern (same heuristics for meta/operations/routes). When the
  next builder needs them, hoist into a shared `internal/builders/...`
  helper.
- **`ScanCtx.FindComments`** is unused after the post-S7 schema
  migration replaces the four regex-based `parsers.{ModelOverride,
  StrfmtName, TypeName, EnumName}` lookups with grammar2-driven
  helpers (each grammar2 parse returns the typed Block, from which
  the schema builder reads the relevant arg directly — no need for
  the up-front `*ast.CommentGroup` lookup). When the migration lands,
  schema's last `FindComments` call site goes away; since
  `FindComments` was only ever consumed by the schema builder, the
  method drops from `ScanCtx`. Track here so the deletion lands in
  P9 alongside the other cross-builder cleanups, after every builder
  is on grammar2 (no risk of a non-schema builder needing to bring
  it back).

These belong in a future `P9: builder-shared infrastructure` pass,
**not** in the P7 series, to avoid speculative abstractions that may
not match the second builder's actual shape.

## 10. Complexity assessment

On the 1–5 scale used in design experiments:

- **Score: 4/5 for S1–S6 (bulk migration).** Multi-file inlining, new
  Builder responsibilities (diagnostics + semantic validation),
  ~7-phase migration with S5.5. Not 5 because no pipeline-stage
  change in grammar2 itself; the grammar API gains accessors but its
  internals don't move. Golden fixtures unchanged at end of S6.
- **Score: 3/5 for S7 (polishing).** Localised — three v1 bugs in
  `schema/`, sharing one `mergeIntoAllOf` helper. Lower complexity
  than the bulk migration but produces **justifiable golden-fixture
  drift** that requires a per-fixture diff explanation in the PR.
  S7 stays out of the critical path so the bulk migration ships
  with byte-stable behaviour.

Combined effort across S1–S7: still recoverable in two reviewable
chunks. S1–S6 is one PR series ("migration, no behavioural change");
S7 is a follow-up PR series ("v1-bug fixes, with explained golden
drift").

Token spend on the analysis that produced this plan:

- Tool calls: 8 (1 ls, 1 wc, 2 reads, 4 greps).
- Total spend: ~3,000 tokens with I/O.

Recorded for the design-experiment baseline: pipe-stage redesigns are
~3k-token-with-I/O exercises in this codebase; single-keyword adds are
~750 zero-I/O.

## 11. Open questions for review

- **Q1.** Is the S5 `Options.OnDiagnostic` callback the right shape, or
  do we want diagnostics surfaced via a separate method on the returned
  `*spec.Swagger` (or an out-parameter)? Affects the public API.

  > **fred**: yes this callback is good.

- **Q2.** Do we want `Builder.Diagnostics()` to be deduplicated across
  Walk() calls (same lexer error reported twice if the same comment
  group is parsed twice), or always append?

  > **fred**: no deduplication for now. We'll watch and improve whatever is
  > necessary when we start doing some serious work with diagnostics.

- **Q3.** S6 semantic validations — is the right list:
  pattern-on-non-string, multipleOf-on-non-numeric,
  enum-item-coercion, required-on-non-object? Should we add
  format-on-wrong-type (e.g. `format: date` on an integer)?

  > **fred**: yes mostly.
  >
  > Required in schema applies to an object properties so YES. Beware that the
  > semantics of "required" are not the same for parameters and response headers.
  >
  > Also:
  > {maximum,minimum,maxItems}-on-non-numeric,
  > {required,readOnly,discriminator,deprecated}-on-non-boolean
  >
  > Basically everything is made rather explicit in keywords.go.
  >
  > Move as much as possible to the new "validations" package.

  > **claude (2026-05-01):** Decision logged. Two implications for the
  > plan:
  >
  > **(a) The semantic-validation list expands** — beyond what S6 had:
  > - shape-mismatch (keyword.Shape × schema.Type/Format mismatch):
  >   pattern/format on non-string; maximum/minimum/multipleOf on
  >   non-numeric; min/maxItems/uniqueItems on non-array;
  >   minLength/maxLength on non-string; min/maxProperties on
  >   non-object; required/readOnly/discriminator/deprecated on
  >   non-boolean (this last set is parser-side already — the lexer
  >   rejects `required: foo`; what's missing is enforcement that the
  >   keyword applies in the resolved type's context).
  > - context-mismatch: required's semantics differ across schema /
  >   parameters / response headers. Schema: name → enclosing.Required.
  >   Parameters: writes Parameter.Required. Headers: writes
  >   Header.Required. Three different writers; the Walker.Bool
  >   handler in each builder owns its own behaviour.
  > - enum-item coercion failure (already discussed; emit
  >   `CodeInvalidEnumItem` per failed item, drop the offending item).
  >
  > **(b) The "validations" package returns** — but with a different
  > role than v0. v0 had it own dispatch (replaced by Walker). v1 has
  > Walker do dispatch, leaving validations free to own **semantic
  > validation rules** (the list above) as a shared cross-builder
  > library. Proposed shape:
  >
  > ```
  > internal/builders/validations/
  > ├── shape.go    — IsLegalForType(keyword Keyword, schemaType, schemaFormat string) (ok bool, hint string)
  > ├── context.go  — IsLegalIn(keyword Keyword, ctx ContextKind) (ok bool, hint string)
  > └── coerce.go   — moved from internal/builders/coerce/ (per Q4 below — single home)
  > ```
  >
  > Each Walker callback in each builder calls `validations.IsLegalForType`
  > before applying. Failure → `Walker.Diagnostic` fires with a
  > `CodeContextInvalid` or `CodeShapeMismatch` and the property is
  > dropped. The keyword table in `grammar2/keywords.go` is the source
  > of truth; `validations.IsLegalForType` reads `Keyword.Shape` and
  > the schema type/format and returns a structured verdict.
  >
  > **(c) Plan changes** — promoting "validations" from declined-to-revived:
  > - Rename the proposed `internal/builders/coerce/` to `internal/builders/validations/`
  >   (per Q4: name "coerce" was Fred-fine, but if we already have
  >   `validations/` for semantic rules, fold coercion into it as
  >   `validations/coerce.go`).
  > - Add a new phase **S5.5** between S5 and S6: create
  >   `validations/` with `shape.go` + `context.go`, ported tests, and
  >   wire each builder's Walker callbacks through it.
  > - S6 becomes "wire validations into Walker.Diagnostic in each of
  >   the four bridges" — no new validation logic, just plumbing.

- **Q4.** S2 enum-coercion location — `internal/builders/enumcoerce/`
  is one option. Alternatives: put it in
  `internal/builders/resolvers/` (already houses cross-builder shared
  concerns) or in a new `internal/builders/coerce/` (broader name).
  Mild preference question.

  > **fred**: "coerce" is fine. In v2 type coercion would be attempted for other things
  > than pure enum (see below).

  > **claude (2026-05-01):** Decided per Q3 above — fold coerce into
  > `internal/builders/validations/coerce.go` so the v2 broader-coercion
  > scope and the semantic-validation rules share one home. If at v2
  > time coerce grows past ~150 LOC we can split it back out.

- **Q5.** Walker fields — `Number`/`Integer`/`Bool` are pre-typed but
  `Raw` is not (because `pattern`/`default`/`example`/`enum` need the
  resolved schema type, which lives outside grammar2). Acceptable
  asymmetry, or should we add `String func(p Property, val string)`
  for `pattern` (which IS pre-typed) and only leave `default`/
  `example`/`enum` on `Raw`? Cleaner taxonomy at cost of one more
  field.

  > **fred**: shouldn't "pattern" be typed as "String"? I don't quite get the usage of "pattern" returning "Raw"...
  >
  > "default", "example", "enum" can indeed be of any type, hence Raw. The asymmetry is acceptable.
  > What we could do is to add a utility method "IsTyped() bool" to avoid the explicit switch over block kind by the caller.

  > **claude (2026-05-01):** Resolved both ways:
  >
  > **Pattern is now String.** Already done in the landed Walker:
  > `walkerDispatchProperty` dispatches by `Keyword.Shape` (the table
  > value), not `Typed.Type` (the lexer-populated value). `pattern` is
  > `Shape=ShapeString` in `keywords.go:177`, so it fires
  > `Walker.String(p, p.Value)` with the raw regex source. Cleaner
  > taxonomy. Verified by `walker_test.go:TestWalker_LevelZeroFiresPerShape`.
  >
  > **`Property.IsTyped() bool` — TODO in S1.** Add to grammar2's
  > `ast.go` so consumers don't switch on `Keyword.Shape` to know
  > whether `default:` / `example:` / `enum:` need coercion.
  > Definition: `p.IsTyped() == p.Typed.Type != ShapeNone`. So:
  > - `maximum: 5` → IsTyped=true, Typed.Number=5
  > - `default: 1.5` → IsTyped=false, Value="1.5" (caller coerces)
  > - `pattern: ^a$` → IsTyped=false, Value="^a$" (no coercion needed,
  >   but caller treats it as raw string from the Walker.String
  >   callback's perspective).
  >
  > Schedule: I'll add `Property.IsTyped()` plus a unit test as part of
  > closing-out this turn (small enough to not need its own phase).

- **Q6.** Walker iteration order — properties fire in source order. Do
  we promise this in the godoc, or leave it unspecified? LSP and
  diagnostics ordering both want stable order, so promising it costs
  nothing.

  > **fred**: yes this is source order within a single file, then top-down for discovered dependencies (if applicable).

  > **claude (2026-05-01):** Acknowledged. Caveat for the godoc: a
  > single Walker call only walks ONE Block (one comment group from one
  > file). The "top-down for discovered dependencies" ordering is a
  > builder-level concern (the `Builder` walks its declarations in
  > scanner discovery order). I'll codify "source order within the
  > Block" in `walker.go`'s godoc — TODO this turn.

- **Q7.** Walker.Unknown — fires for keywords grammar2 doesn't
  recognise (typos, future keywords). Should it always fire (so
  consumers can record diagnostics), or only when the consumer wires
  it up (default = silent drop)? Today's bridges silently drop, so
  default-silent matches behaviour but loses the diagnostic
  opportunity.

  > **fred**: fires when wired only - silent drop is the current default.
  > Once we land a working version, we'll inject a logger there to spew warnings.
  > And for v2, we'll inject smarter diagnotic tools.

  > **claude (2026-05-01):** Acknowledged. Already implemented this way
  > — `walkerDispatchProperty` checks `if w.Unknown != nil` before
  > firing, so the default for an unset field is silent drop. Verified
  > by `walker_test.go:TestWalker_NilCallbacksAreNoops`.
