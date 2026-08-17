# `additionalProperties` / `patternProperties` — design

**Branch:** `feat/additional-properties` (worktree `.worktrees/feat/additional-properties`)
**Status:** ✅ Phases 1, 2a, 2b, 2c landed (2026-06-17). Phase 1 committed + rebased onto
lot1; 2a/2b/2c committed on `feat/additional-properties`, awaiting review before lot1 merge.
**Folds in:** go-swagger #2251 (§18), #3005, #2539 (§17), and typed `patternProperties`
of `forthcoming-features.md`.

This is a project-mode track (design → debate → execution), not a one-off backlog
fix. The seeded RED/locking witnesses live on the branch tip
(`coverage_bug_2251_test.go`, `coverage_bug_3005_test.go`,
`coverage_pattern_properties_test.go`).

---

## 1. Current state (grounded)

Structural derivation (map → `additionalProperties`) lives entirely in
`buildFromMap` (`internal/builders/schema/schema.go:506`):

```go
key := titpe.Key()
if key.Underlying().String() == "string" || resolvers.IsTextMarshaler(key) {
    return s.buildFromType(titpe.Elem(), eleProp.AdditionalProperties())
}
return nil   // every other key kind → silently dropped to a typeless property
```

- The **only** way to get `additionalProperties` today is a map-typed entity with a
  string-underlying or `TextMarshaler` key. There is **no annotation-driven control**.
- `patternProperties` is already wired (`KwPatternProperties` keyword →
  `handlers.ApplyPatternProperties` → `schema/walker.go:252`), RE2-hygiene-checked with
  diagnostics — **but its value schema is always the empty `{}`** (`dispatch_schema.go:78`),
  and it is `CtxSchema`-only. No typed values.
- Diagnostics infra is ready: `Options.OnDiagnostic`, `grammar.Diagnostic`,
  `CodeInvalidAnnotation`, `RecordDiagnostic`, severities.
- Test status at branch tip: `Bug2251` **RED**; `Bug3005` **GREEN, locking the buggy
  behaviour** (`AdditionalProperties == nil`); `PatternProperties` **GREEN**.

Good news: we are not starting from zero. The existing machinery is imperfect but works
for the basic string-keyed case.

### How a `map[K]V` flows

- **Top-level `swagger:model M map[K]V`** → `buildFromType` → `*types.Map` →
  `buildFromMap`: emits `{type: object, additionalProperties: <V rendered recursively>}`,
  gated on the key. A `*types.Named` whose underlying is a map goes through
  `buildNamedType` → `resolveRefOr` (so it may become a `$ref`).
- **Map struct field** → `buildFromStruct` → `structFieldCarrier` → `applyFieldCarrier`
  → `buildFromType` → same `buildFromMap`. The field becomes one property.
- **`json:"-"` map field** → `structFieldCarrier:212` hits the `ignore` branch → the field
  is **muted** (and evicts any inherited same-Go-name property). This is exactly why
  #3005 loses its map. Muting is by design and is the show-stopper for auto-detection.
- **Map alias** → `buildAlias` honours `RefAliases`/`TransparentAliases`, else recurses to
  the underlying map.

## 2. The authoritative map-key rule (`encoding/json`)

Verified against `encoding/json/encode.go` `newMapEncoder` and by empirical
`json.Marshal` probes (2026-06-16). A Go map marshals to a JSON object **iff** its key
type is:

- a **string** kind, **or**
- an **integer / unsigned-integer** kind — `int`,`int8`…`int64`, `uint`,`uint8`…`uint64`,
  **and `uintptr`** (all stringified, e.g. `map[int]int{-2:9}` → `{"-2":9}`), **or**
- implements **`encoding.TextMarshaler`** (pointer receiver counts when the key is `*T`;
  when a type has both `MarshalText` and `MarshalJSON`, the **key uses `MarshalText`**).

Everything else **errors at runtime**: `float`, `bool`, `struct` without `TextMarshaler`,
`func`, `map[any]any` (the static key type is `interface{}`), and — importantly —
a key implementing **only `json.Marshaler`** (`json.Marshaler` is *never* consulted for
keys). `fmt.Stringer` does **not** qualify.

This set is **statically decidable** at the `go/types` level:
`IsString(key) || IsInteger(key) || resolvers.IsTextMarshaler(key)`
(`*types.Basic.Info() & types.IsInteger` already covers every int/uint/uintptr kind).
So we can detect the illegal-key cases **perfectly** for the static key type. The
*imperfect* detection problem is on the **value** side (`map[string]func() error` has a
legal key and an unmarshalable value) — partial, best-effort only.

## 3. Design — the unified annotation surface

Two surfaces share one value grammar.

### 3.1 Value grammar (`<spec>`)

`<spec>` ::= `true` | `false` | `<TypeSpec>`

where `<TypeSpec>` reuses the **`swagger:type` spec resolver** so authors can inline
anything `swagger:type` already accepts — a primitive (`string`, `integer`, `number`, …)
or a model name resolving to a `$ref`. This keeps the surface consistent with the
recently-introduced `swagger:type` inlining.

- `true`  → `additionalProperties: true` (explicit allow; OAS/JSON-Schema default, but
  authors may want it stated).
- `false` → `additionalProperties: false` (forbid extra keys; `SchemaOrBool{Allows:false}`).
- `<TypeSpec>` → `additionalProperties: {<schema for TypeSpec>}` (primitive schema or `$ref`).

### 3.2 Type/model marker — `swagger:additionalProperties <spec>`

Marker form (space, no colon — matching `swagger:type X`), placed on a `swagger:model`
type decl. **Semantics depend on the underlying Go type:**

- **On a map** → it *defines* the value schema, overriding the inferred element type.
  ```go
  // swagger:model
  // swagger:additionalProperties Thing
  type Index ordered.Map
  ```
  ```yaml
  Index: { type: object, additionalProperties: { $ref: '#/definitions/Thing' } }
  ```
  Works even when the Go type is **not** a map (e.g. a custom `ordered.Map` wrapper):
  the annotation beats the actual type.

- **On a struct** → it *complements* the object (scope = **this type only**, never the
  contained fields/elements):
  ```go
  // swagger:model
  // swagger:additionalProperties false
  type Thing struct { Bag, Of, Stuff string }
  ```
  ```yaml
  Thing: { type: object, properties: {…}, additionalProperties: false }
  ```

  Multi-annotation combinations need a consistent rendering and form a core part of
  the test matrix (§4 carries the precedence principle). The contradiction cases all
  resolve to "the other annotation wins, AP is dropped with a diagnostic":

  | Combination | Outcome |
  |---|---|
  | `swagger:additionalProperties true` + `swagger:strfmt string` | strfmt wins → AP inconsistent → diagnostic |
  | `swagger:additionalProperties true` + `swagger:type uint8` | type wins → AP inconsistent → diagnostic |
  | `swagger:additionalProperties true` + `swagger:type []T` | type wins → AP inconsistent → diagnostic |
  | `swagger:additionalProperties true` + `swagger:type object` | consistent → both apply |

### 3.3 Field keyword — `additionalProperties: <spec>`

Grammar keyword (colon) on a struct field, scoped to the field's own schema. **Beats the
field's Go type** — even a non-map field becomes `{type: object, additionalProperties: …}`
(again, the `ordered.Map`-wrapper case). Same `<spec>` grammar.

`json:"-"` still mutes the field outright; an author who wants a muted map to become the
model's `additionalProperties` uses the **type-level marker** (§3.2) with an explicit
`<TypeSpec>` instead — this is the #3005 resolution (the value type is restated in the
annotation; the `json:"-"` map stays a muted implementation detail).

`additionalProperties` does not preclude other object-compatible validations on the same
schema — `maxProperties`, `minProperties`, `patternProperties` may all co-exist with it.

### 3.4 Typed `patternProperties` — `swagger:patternProperties "<re>": <spec>, …`

Multi-pair marker (type/model and field level). The regex is **quoted** so it may contain
commas/spaces; each pair is `"<regex>": <TypeSpec>` (same value grammar, sans the
bare-bool forms):
```go
// swagger:model
// swagger:patternProperties "^[a-z]-": Thing, "^\d+$": integer
type Index ordered.Map
```
```yaml
Index:
  type: object
  patternProperties:
    "^[a-z]-": { $ref: '#/definitions/Thing' }
    "^\d+$":   { type: integer }
```
The existing `patternProperties: <regex>` field keyword (regex → empty `{}` schema) stays
for back-compat; the typed marker is the richer form. Each regex keeps the RE2-hygiene
check + `CodeInvalidAnnotation` diagnostic already in place.

### 3.5 Auto-detection (structural, no annotation)

A map type or a **non-muted** map field auto-emits `additionalProperties: V` when the key
is JSON-legal per §2 (this is the existing string path **widened** to integer/uint keys —
§18 / #2251). `json:"-"` is the show-stopper.

### 3.6 Diagnostics

A map whose static key type is JSON-illegal per §2 (and that is *not* muted by `json:"-"`)
emits a **warning** diagnostic (the type genuinely fails `json.Marshal`) and the property
is dropped (no bogus typeless schema), e.g.:
```go
// swagger:model
type M map[uint]IndexedThing   // OK — uint is legal, emits additionalProperties
type N map[float64]IndexedThing // WARN — float64 key never marshals
```
Severity is **warning** (not info): these error at runtime, and the `tries-hard-to-
stringify` intuition does not apply to the illegal static key types (empirically
disproven). `json:"-"` fields never reach `buildFromMap`, so muted maps never warn.

## 4. Precedence / combination matrix

**Governing principle:** `swagger:additionalProperties` is the **lowest-priority**
annotation in any mix. If any other rule has already resolved the schema to something
other than an `object` — `swagger:type` on a non-object, `swagger:strfmt`, a special/known
type, whatever — then `additionalProperties` is **dropped with a warning diagnostic**. It
only ever applies on top of an `object`.

| Situation | Resolution |
|---|---|
| `swagger:type T` present | wins the **type** axis as today; if `T` is not `object`, `additionalProperties`/`patternProperties` on the same decl is **warn-and-dropped** |
| `swagger:strfmt …` present | strfmt forces `{type:string,…}` → `additionalProperties` is **warn-and-dropped** (non-object) |
| `swagger:additionalProperties` on a **map** | annotation value schema **overrides** the inferred element type |
| `swagger:additionalProperties` on a **struct** | **complements** — adds AP alongside named properties; scope is this type only |
| `swagger:additionalProperties false` on a **map** | contradiction → **warn-and-ignore** (a map inherently has additional properties) |
| field `additionalProperties: <spec>` | applies to that field's schema; **beats the field's Go type** (resolves it to `object`) |
| `json:"-"` field | muted (structural + keyword); use the type-level marker to express AP |
| auto-detected map (legal key) + explicit annotation | explicit annotation wins |

## 5. OAS2 validity stance

`patternProperties` (and, later, `additionalItems`) are **not** in the Swagger-2.0 Schema
Object subset — they are JSON-Schema draft-4 keywords. go-openapi has always favoured
JSON Schema over Swagger when the two conflict, `go-openapi/spec` models these fields, and
the broader ambition is full JSON-Schema support (draft-4 now, later drafts in future
codescan versions). We therefore **keep emitting them, ungated** — this also unblocks a
forthcoming "schema-only" generator (definitions without a full spec). No option flag.

`patternProperties` is always something the author explicitly calls for: no gate, and no
auto-inference (for now — see §7).

## 6. Execution plan

Land the structural fix first (isolated, low-risk, no annotation design), then the
annotation trio as a unified second phase.

### Phase 1 — §18 / #2251: widen the key gate + fail-loud (structural) ✅
- ✅ New `resolvers.IsJSONMapKey` (`IsString || IsInteger || IsTextMarshaler`, covers
  int/uint/uintptr); `buildFromMap` routes through it. Element schema unchanged.
- ✅ JSON-illegal static key (non-muted) → `grammar.Warnf(…, CodeUnsupportedType, …)`
  ("additionalProperties dropped"), no AP emitted. `json:"-"` maps never reach
  `buildFromMap`, so muted maps never warn.
- ✅ `coverage_bug_2251_test.go` GREEN; added the float-key fail-loud witness
  (`BadKey map[float64]int`) asserting no-AP + one `CodeUnsupportedType` warning.
- ✅ Repurposed the obsolete `index_map`/`UnsupportedMap` special-schema witness
  (`map[int]struct{}` is now genuinely supported) → renamed `IndexedMap`, subtest
  asserts object+additionalProperties. One golden drifted (`go123_special_spec.json`).
- ✅ Full suite green; `golangci-lint --new-from-rev master` → 0 issues.

### Phase 2a — `swagger:additionalProperties` marker (#3005, §17, #2539) ✅
- ✅ Value-grammar `<spec>` = `true | false | <TypeSpec>` (`resolveAdditionalPropertiesType`)
  reusing the `swagger:type` arg grammar (`[]T` layers / primitives) but resolving a
  type-name to a **`$ref`** (not inline), with discovery registration.
- ✅ `AnnAdditionalProperties` classifier wired through the grammar (`annotations.go`
  const/label/String/FromName/family; `lexer.go` arg → `argTypeRef`) and the scanner
  classifier allow-list (`index.go`). Grammar round-trip + classifier unit tests added.
- ✅ `classifierAdditionalProperties` (`schema/additional_properties.go`), applied in
  `Build` after `buildFromDecl`: struct = complement, map = override, bare `$ref` = define
  (reset). Lowest-priority precedence — non-object → `CodeShapeMismatch` warn-and-drop.
- ✅ #3005 reseeded with `swagger:additionalProperties number`; witness flipped to fix-lock;
  `bugs_3005_schema.json` regenerated.
- ✅ `fixtures/enhancements/additional-properties` + golden + `coverage_additional_properties_test.go`
  cover `true`/`false`/typed/`$ref`, map-override, `maxProperties` coexistence, and the
  `swagger:type` contradiction (dropped + diagnostic). READMEs updated (`schema`
  §additional-properties, `grammar` classifier lists). Full suite + lint + markdown clean.

### Phase 2b — `additionalProperties:` field keyword ✅
- ✅ New `KwAdditionalProperties` keyword (`CtxSchema`, akas `additional properties`/
  `additional-properties`). Resolver refactored: `resolveAdditionalPropertiesValue`
  (pure `true|false|TypeSpec` → `SchemaOrBool`) shared by marker, field keyword, and the
  $ref-sibling path.
- ✅ Two landing paths in `walker.go`: non-`$ref` field → post-scan
  `applyAdditionalPropertiesSpec` (override map element / warn-drop on primitive);
  `$ref`'d field → `refOverrideCollector` emits an **allOf sibling**
  (`{allOf:[{$ref},{additionalProperties:…}]}`, ref preserved).
- ✅ `fixtures/enhancements/additional-properties-field` + golden +
  `coverage_additional_properties_field_test.go` (map-override, `true`, typed `$ref`,
  allOf-sibling close, primitive contradiction + diagnostic). README §additional-properties
  extended. Full suite + lint clean.

### Phase 2c — typed `swagger:patternProperties "<re>": <spec>, …` marker ✅
- ✅ `AnnPatternProperties` classifier (grammar + scanner allow-list); lexer captures the
  whole `"<re>": <spec>, …` remainder as one raw arg token, read via the non-filtering
  `findRawAnnotationArg`, split by `parsePatternPropertyPairs` (quote-aware; `\d` preserved,
  only `\"` escaped). Values resolve through the shared `<TypeSpec>` resolver (→ `$ref`).
- ✅ `classifierPatternProperties` (`schema/pattern_properties.go`), applied in `Build`:
  object-only precedence, `$ref`-reset, per-regex RE2-hygiene check (CodeInvalidAnnotation),
  malformed-list diagnostic. Coexists with the regex-only `patternProperties:` keyword.
- ✅ `fixtures/enhancements/pattern-properties-typed` + golden + integration test (multi-pair
  primitive/`$ref`, invalid-regex diagnostic, contradiction). Grammar round-trip + pair-parser
  unit tests. README §pattern-properties. Full suite + lint clean.
- Note: implemented at type/model level (Fred's documented example). Field-level typed
  patternProperties not wired (the regex-only field keyword remains); revisit if demand.

### Phase 2 wrap ✅
- ✅ `forthcoming-features.md` §17 and §18 marked ✅ DONE with pointers here.

## 7. Deferred

- **`withPatternProperties` inference knob** — auto-deriving `patternProperties:
  {"^-?\d+$": V}` from a non-string-keyed map. Over-ambitious for now; explicit
  patternProperties (§3.4) ships first, auto-detected version revisited in a future
  release.
- **Value-side unmarshalability detection** (`map[string]func() error`) — partial only;
  not attempted in Phase 1.
- **`additionalItems`** — same JSON-Schema-over-Swagger rationale; not in this track.
- **Infer the value type for a type-level marker from a `json:"-"` map field** — possible
  convenience for #3005, but ambiguous (which map?); keep the explicit `<TypeSpec>`.

## 8. Change history

| Date | Note |
|---|---|
| 2026-06-16 | Initial design. Grounded investigation + empirical `encoding/json` map-key rule; decisions settled with Fred (keep patternProperties ungated; `swagger:additionalProperties`/field-keyword surface with `true\|false\|TypeSpec` grammar; typed `swagger:patternProperties`; warn-and-ignore contradictions; warning severity for illegal keys; `withPatternProperties` deferred; #2251 first). |
| 2026-06-16 | Folded in Fred's review notes: AP is the lowest-priority annotation (warn-and-drop on any prior non-object resolution; strfmt/type contradiction grid in §3.2); AP co-exists with other object validations (`maxProperties`, …); patternProperties stays author-explicit, no gate/inference. |
