# Doc-site quirks — F1–F5

Date surfaced: 2026-06-13
Surfaced by: the `doc-site-reference` branch (reference + tutorial build).
Intended fix branch: **a dedicated scanner branch off master, to land BEFORE
the doc site goes live** — so the published docs describe consistent, fixed
behavior rather than today's quirks.

Companion docs: [`observed-quirks.md`](observed-quirks.md) (Q-description
source-of-truth), [`deferred-quirks.md`](deferred-quirks.md) (v2/holding area),
[`fix-quirks.md`](fix-quirks.md) (the closed post-Stream-M pass this mirrors in
shape). The doc-site work that produced these is tracked in the doc-site
worktree at `.claude/plans/doc-site-reference.md` §Behavioral findings; the
durable cross-worktree pointer is the memory `project_scanner_quirks_from_docs`.

## Ground rules (carry-over from fix-quirks.md)

- **One commit per F** (DCO `-s`, `Co-Authored-By: Claude`). Fix + golden in the
  same commit.
- **Witness-then-fix**: capture the pre-change golden, regenerate post-fix, the
  golden *diff* is the audit trail (load-bearing on schema-discovery paths —
  see `feedback_schema_discovery_verify_with_witness.md`).
- `UPDATE_GOLDEN=1 go test ./...` to regenerate; `golangci-lint run
  --new-from-rev master` before push.
- Status icons: ⬜ todo · 🟡 in progress · ✅ done · 🔵 blocked · ⏸ re-deferred.

All five were verified empirically against `docs/examples/concepts/models`
scans (in the doc-site worktree) plus the builder code.

---

## F1 — `swagger:strfmt`+`swagger:model` does not produce a `$ref` definition ✅ (fix/quirks-model-override `8e20d2f`)

**Symptom.** A named string-format type annotated with **both** `swagger:strfmt X`
and `swagger:model`: referencing fields still **inline** `{type:string,
format:X}`, and the type emits an **orphan `{type:string}` definition with the
format dropped**. No `$ref` to a named strfmt definition is produced.

**Expected.** `+swagger:model` should register the strfmt type as a first-class
definition `{type:string, format:X}` that referencing fields point at via
`$ref` (consistent with the general "swagger:model ⇒ definition + $ref" rule).

**Root.** The strfmt classifier (`classifierNamedStructStrfmt` / basic-arm
strfmt path) returns handled **before** the `$ref` pivot `resolveRefOr`
(`internal/builders/schema/schema.go` named-type dispatch; classifier in
`internal/builders/schema/walker_classifiers.go`). The format is set on the
field target inline; the model registration produces a degenerate definition.

**Repro.** `type UUID string` with `// swagger:strfmt uuid` + `// swagger:model`,
referenced by a struct field. (Variant was prototyped then parked in the
doc-site `concepts/models` package.)

---

## F2 — `swagger:type`+`swagger:model` does not produce a `$ref` definition ✅ (fix/quirks-model-override `8e20d2f`)

**Symptom.** A named type annotated with both `swagger:type T` and
`swagger:model`: referencing fields **inline** the overridden type; the type
emits an **orphan, unreferenced `{type:T}` definition**. No `$ref`.

**Expected.** Same as F1 — `+swagger:model` should yield a definition fields
`$ref`. Same classifier-short-circuit root (`classifierNamedTypeOverride` at
`internal/builders/schema/walker_classifiers.go:87` returns handled before
`resolveRefOr`).

**Repro.** `type RawID [12]byte` with `// swagger:type string` + `// swagger:model`,
referenced by a struct field.

---

## F3 — `swagger:type` accepts unsupported values silently ✅ (fix/quirks-F-series: witness `cc59e88` + reconciliation `80300b6`)

**Symptom.** `swagger:type` performs **no closed-vocabulary validation**.
`internal/builders/schema/walker_classifiers.go:87` calls
`resolvers.SwaggerSchemaForType` (`internal/builders/resolvers/resolvers.go:32`),
whose `default:` returns an error; the classifier then returns
`(handled=true, fallthroughUnderlying=true)`, i.e. the bad value is **silently
ignored** and the field falls back to the underlying Go type — no warning, no
diagnostic. Accepted set: `string/integer/number/boolean/object` + Go builtins;
notably **`array` and `file` are NOT accepted** via `swagger:type`.

**Expected.** Emit a diagnostic/warning on an unsupported `swagger:type`
argument (via `Options.OnDiagnostic` / the logger), rather than silently
dropping the override. Decide whether `array`/`file` should be supported.

---

## F4 — `swagger:enum` is inline-only; no `$ref` opt-in; bare form misses consts ✅ (fix/quirks-model-override `8e20d2f` + `716d8c1`)

**Resolution (2026-06-15).** (a) `+swagger:model` now makes the enum a
first-class definition (carrying the `enum` + `x-go-enum-desc`) that fields
`$ref` — the `$ref` opt-in (`8e20d2f`, with F1/F2). (b) bare `swagger:enum` (no
name) on a type decl is now valid and infers the name from the declaration,
collecting its consts (`716d8c1`). `annotations.md` updated. Original report
below.

**Symptom.** `swagger:enum <Name>` + matching consts: the values inline
(`enum` + `x-go-enum-desc`) on each **referencing field**; the enum type is
**not** a standalone definition. `+swagger:model` only yields an orphan
value-less `{type:string}` definition (still inlines on the field). Separately,
**bare `swagger:enum` (no name arg) failed to collect consts** in testing; the
named form `swagger:enum TypeName` works. The enum-with-no-consts case
(Case D in `fixtures/enhancements/enum-overrides`) becomes a `$ref` definition
with no `enum`.

**Expected.** (a) Decide whether an enum-as-`$ref`+definition opt-in is wanted
(parallel to F1/F2). (b) Confirm/fix bare-`swagger:enum` const collection or
document the name as required. The migrated `maintainers/annotations.md` has
already been corrected to describe the inline reality on the doc-site branch.

**Decision input.** `enhancements_enum_overrides.json` is the factual reference.

---

## F5 — `swagger:name` ignored on struct fields ✅ (fix/quirks-F-series `5a05231`)

**Symptom.** `swagger:name X` works on **interface methods** (every corpus use
is a method — `fixtures/goparsing/classification/models/nomodel.go`,
`fixtures/enhancements/interface-name-verbatim`). On a **struct field** the
override is **silently ignored** (the property keeps the Go field name / json
tag). `maintainers/annotations.md` claims it works on "field OR method".

**Expected.** Either support `swagger:name` on struct fields (so it overrides
the json-tag/field-name derivation) or correct the reference to "interface
methods only". Verify intended scope, then align code + doc.

**Repro.** A `swagger:model` struct with a tag-less field carrying
`// swagger:name balance` — property stays the Go field name.

---

## F6 — `deprecated:` on a model field is silently dropped ✅ RESOLVED (go-swagger#3138: model/field deprecation → `x-deprecated`, incl. godoc `Deprecated:`)

**Symptom.** `deprecated: true` on a `swagger:model` struct field produces **no**
output — no `deprecated`, no `x-deprecated` — and **no diagnostic**. It works
correctly on operations (`swagger:route`/`swagger:operation` body →
`operation.deprecated: true`).

**Context.** OpenAPI 2.0's Schema object has no `deprecated` (that arrives in
OAS 3.0), so dropping it is arguably correct. The question is only whether to
(a) emit a diagnostic (consistent with the F3 "warn on silently-dropped
keyword" theme), or (b) emit `x-deprecated` as a vendor extension for
round-tripping. Verified via `docs/examples/concepts/decorators` (probe removed
after confirmation). Low priority.

## F7 — gofmt-canonical `swagger:meta` YAML fails to parse ✅

**Provenance.** Surfaced during the go-swagger backlog verification pass (not
the doc-site build). A well-scoped robustness bug, given its own branch
**`fix/scanner-gofmt-meta-yaml`** — ready ahead of the design-laden F1–F6.

**Symptom.** A `swagger:meta` YAML block written in the natural column-0 form
(top-level key at column 0, indented children) fails to scan with `found
character that cannot start any token` **once gofmt runs**. gofmt inserts a
mandatory blank `//` line under the column-0 key (its doc-comment code-block
rule); the tab-indented children after that blank line then reach the YAML
sub-parser, which rejects tab indentation. Empirically:

- **A** col-0 key + blank `//` + tab children — *gofmt-clean* → **errors**.
- **B** col-0 key + tab children, no blank — *not gofmt-clean* → parses.
- **C** whole block uniformly tab-indented + blank — *gofmt-clean* → parses.

gofmt rewrites **B → A**, so a user who writes the natural form and formats
their code ends up with the failing source.

**Expected.** The meta-block preprocessor tolerates the gofmt-inserted blank
line / normalises tab indentation so the gofmt-canonical form (A) parses
identically to C.

**Root.** Pinpointed to `yaml.RemoveIndent` (`internal/parsers/yaml/dedent.go`):
it keyed the dedent strip width off the *literal* first body line. gofmt's
inserted blank line made that width zero, so RemoveIndent returned early
without stripping or retabbing, leaving the children's tab indentation for the
YAML parser to reject.

**✅ Fixed (`fix/scanner-gofmt-meta-yaml`, commit `0545527`).** RemoveIndent now
keys the strip width off the first NON-BLANK line, so a leading blank line is
skipped and the gofmt-canonical form dedents + retabs identically to the
uniformly-indented form. Behaviour is unchanged for bodies without leading
blanks (operations / meta goldens untouched). `TestQuirk_GofmtMetaYAML` flipped
from documented-error to asserting the parsed securityDefinitions (golden
`quirk_gofmt_meta.json`); direct `RemoveIndent` unit tests added
(`internal/parsers/yaml/dedent_test.go`). Full suite + lint clean. This closes
the residual behind go-swagger#2959 (previously "fixed" only for non-gofmt'd
meta forms).

**✅ Extended to `swagger:operation` bodies (2026-06-15, branch
`fix/gofmt-operation-yaml-tabs`).** The first-non-blank-line fix above covered
swagger:meta because a meta body is a SINGLE uniformly tab-prefixed block. A
swagger:operation `---` body, by contrast, interleaves column-0 top-level keys
(`responses:`, `x-*:`) — rendered by gofmt as **prose** lines (one leading
space) — with **tab-prefixed** value blocks. RemoveIndent stripped the prose-key
width (1) off the children's lone leading tab, flattening the nesting
(`responses` → bare `default` with empty description; vendor extension →
`null`). Fix: expand every non-blank line's leading tabs to two spaces BEFORE
the first-line strip (replacing the strip-then-retab-remainder order), so tab-
and space-indented lines are comparable. Witness `fixtures/quirks/gofmt-operation`
+ `TestQuirk_GofmtOperationYAML`; `RemoveIndent` interleaved-shape unit test;
`bugs_3138_schema.json` regenerated (that fixture was itself tab-form and had
silently dropped its `200` response). Affects backlog #1867 / #1958.

## F8 — string `example`/`default` values retain surrounding quotes ✅ RESOLVED (go-swagger#2547: surrounding double quotes stripped; `example: ""` → empty string)

> ⚠️ **Numbering collision:** there are two `## F8` headings in this file — this
> one (example/default quote retention) and the `swagger:alias` no-op below
> (which landed ✅ `cfd8917`). The F1–F9 fix wave closed the *alias* F8; **this
> quote-retention F8 was NOT fixed** (verified 2026-06-15: `coerce.go` still has
> no string-unquote case; `coverage_bug_2899_test.go` still asserts the quoted
> value; `fix/go-swagger-2547`'s `TestCoverage_Bug2547` is RED on `master`).
> Tracked as backlog **#2547** (🛠, branch `fix/go-swagger-2547`). Renumber one of
> the two F8s when convenient.

**Provenance.** Surfaced during the go-swagger backlog verification pass
(go-swagger#2899).

**Symptom.** A quoted string literal in an `example:` (or `default:`)
annotation keeps its surrounding quotes in the emitted value: `example:
"123456"` → `"123456"` (an 8-char string *including* the quotes) rather than
`123456`. The value IS now carried onto the schema (the #2899 fix); only the
quoting is wrong.

**Expected.** For a string-typed `example:` / `default:`, strip a single pair
of surrounding double quotes so `example: "123456"` yields `123456` (while
still permitting intentional inner quotes).

**Root.** String coercion has no unquote step:
`internal/builders/validations/coerce.go` `CoerceValue` has no `case "string"`
— string targets fall to `default`, which returns the raw annotation token
verbatim (quotes included).

**Repro.** `internal/integration/coverage_bug_2899_test.go` (on
`fix/backlog-lot1`) asserts the current quoted value with a `TODO F8`; flip the
assertion and regenerate when fixed.

**Backlog manifestation + RED witness:** go-swagger#2547 ("unable to set a field
value as empty string") is the same bug — `example: ""` yields the 2-char string
`""` instead of the empty string. Branch **`fix/go-swagger-2547`** carries a RED
witness (`fixtures/bugs/2547` + `TestCoverage_Bug2547`) asserting the
quote-stripped values for an empty-string example, a quoted example, and a quoted
default. The F8 fix (the `case "string"` unquote in
`validations/coerce.go CoerceValue`) lands on that branch and turns both the
#2547 witness and the #2899 assertion green.

## F8 — `swagger:alias` appears to be a no-op (dissolves like unannotated) ✅ DEPRECATED (fix/quirks-F-series `cfd8917`)

**Resolution (2026-06-15).** The report's premise was inaccurate: `swagger:alias`
was not a no-op — on a named **primitive** it force-inlined the scalar
(`{type:string}`) instead of the `$ref` a named type gets; it was inert only on
non-primitive aliases (the observed case). It never produced the
docs-claimed `$ref`-to-target. Per the agreed reconciliation it is now
**deprecated** as an empty sink: emits a `validate.deprecated` diagnostic, no
effect on output (primitives now `$ref` their definition like any named type).
Migration: `swagger:type inline` / `swagger:model` / the alias Options.
`annotations.md` rewritten. Locked by `fixtures/quirks/alias-deprecated` +
`TestQuirk_AliasDeprecated`. Original report below.

**Symptom.** A type alias annotated `swagger:alias` (`type Price = Money //
swagger:alias`) behaves exactly like an **unannotated** alias: at use sites it
dissolves to the target (`$ref: Money`) and produces **no definition of its
own** — in all of default / RefAliases / TransparentAliases modes. The
first-class-alias opt-in that actually works is `swagger:model` (per
`alias-calibration-embed`'s `BaseAliasModeled`).

**Conflict.** `maintainers/annotations.md`'s `swagger:alias` section describes it
as publishing a `$ref` to the target / a first-class entity — which is not what
happens. Either `swagger:alias` should do something distinct from
`swagger:model`-on-alias, or the reference (and the swagger:alias annotation
itself) should be reconciled/removed. Verified via `docs/examples/shaping/aliases`
(the `swagger:alias` variant, since replaced by the safe unannotated form).

**Note.** Closely related to F9 — fix together in the alias-handling pass.

## F9 — `swagger:model` on a type alias hangs the scanner (infinite loop) ✅ RESOLVED (fix/quirks-F-series `5a56679`, locked)

**Resolution (2026-06-15).** The hang **no longer reproduces** on the current
lot1 base: the exact doc-site trigger (bare `swagger:model` on `type Price =
Money`, both annotated, `Invoice` referencing `Price`) completes in
milliseconds with correct output in all three modes (default → inlined
structural copy, RefAliases → `$ref` to target, TransparentAliases → inlined).
It was dissolved by the intervening builder/discovery/grammar work accumulated
in the lot branch since this quirk was filed, not a single named fix. Locked by
`fixtures/quirks/alias-model` + `TestQuirk_AliasModelNoHang` (scans all three
modes; a regression resurfaces as a CI `-timeout`). The doc-site may now
re-enable its first-class-alias how-to. Original report below.

**Symptom.** A `swagger:model`-annotated Go **type alias** (`type Price = Money`
where `Money` is itself `swagger:model`) sends the scan into an **infinite loop**
once the alias is actually built — a 10-minute test timeout, goroutine stuck in
`golang.org/x/tools/go/ast/astutil.PathEnclosingInterval` / `go/ast.(*File).End`.

**Trigger isolation (verified, `docs/examples/shaping/aliases`):** the alias is
"built" when it is reachable (referenced from an operation/model) **or** picked
up as a `swagger:model` root under `ScanModels`. Across modes:

- default (expand) → **HANG**
- `RefAliases: true` → **HANG**
- `TransparentAliases: true` → completes (dissolves the alias to its target
  before the build path that loops)
- alias not reached at all (nothing references it, no `ScanModels`) → completes
  (never built)

So the loop is in the **first-class-alias build path** (expand / ref structural
copy), which `TransparentAliases` bypasses. The `alias-calibration-embed`
fixture exercises `BaseAliasModeled` without hanging, so the trigger is
form-specific — narrow this against that fixture during the fix.

**Severity.** High — an infinite loop on valid input. Blocks shipping a live
"Alias rendering" how-to with a default/RefAliases example; that page currently
documents the modes conceptually + a safe unannotated-alias dissolve example
only. Highest-priority of the doc-site batch.

**Repro.** `type Money struct{…} // swagger:model`; `type Price = Money //
swagger:model`; a `swagger:model` `Invoice` with a `Price` field reachable from
a route → scan in default or RefAliases mode.

## After the fix branch lands

Re-enable the parked doc-site example variants (strfmt-ref, type-ref,
enum-ref panes; struct-field `swagger:name`) and revisit
`maintainers/annotations.md` F3/F5 wording so the published docs match the
fixed scanner. See `project_scanner_quirks_from_docs` memory for the trigger.

---

# Round 2 — doc-vantage findings (G-series, 2026-06-16)

Surfaced while keeping the doc-site aligned with `fix/backlog-lot1` (boundary
~`7334f59`), each reproduced with a probe scan of a minimal example package.
These are *consistency* smells rather than crashes — judgment calls for the
fixers, not necessarily bugs.

## G1 — `body:<Type>` response description defaults to the type name ✅ (fix/quirks-g1-body-response-desc, merged 4121b42)

**Resolution.** `buildRouteResponse` now derives a fallback description (only
when no trailing text is given) via a 3-tier `bodyResponseDescription`: (1) the
referenced model's godoc title/description; (2) the HTTP status reason phrase for
a numeric code (404 → "Not Found"); (3) a neutral "default response" placeholder.
Applied uniformly to the `body:<Type>` and definition-fallback forms; never
leaks the Go token. Reviewed + approved from the doc-vantage. Doc-site follow-up:
the `inline-response-bodies` how-to can drop its trailing-description workaround
to showcase the derived default.

**Symptom.** In a `swagger:route` `Responses:` block, a `body:` entry with no
trailing description fills `response.description` with the **type name**, not an
empty string or the type's godoc:

- `200: body:Pet`    → `{description: "Pet",    schema:{$ref Pet}}`
- `404: body:string` → `{description: "string", schema:{type:string}}`

OAS2 requires a non-empty `description`, so something must be emitted — but
"Pet" / "string" is a type token, not a human description, and leaks the Go type
name into the contract. The doc-site `inline-response-bodies` how-to works around
it by always adding trailing description words. Candidate: default to the named
type's title/godoc, or to a neutral placeholder.

**Repro.** `// swagger:route GET /q things op` with `Responses:` `200: body:Pet`,
no trailing text; inspect `op.responses["200"].description`.

**✅ Fixed.** `internal/builders/routes/walker.go` now derives the fallback
description (new `bodyResponseDescription`) in three tiers — same tiering for
the `body:<Type>` form and the definition-fallback `<Model>` form, so they agree:

1. the referenced model's own godoc (`Title`, then `Description`) — mirroring
   how a named `swagger:response` takes its description from prose, instead of
   echoing the type token;
2. the HTTP status reason phrase for a numeric code (`200`→"OK",
   `404`→"Not Found", `500`→"Internal Server Error");
3. a neutral `"Default response"` placeholder (one const) for the `default`
   catch-all and non-standard numeric codes with no godoc — kept neutral since
   OAS2 `default` covers any undeclared code, not necessarily an error.

Never leaks the Go token; always non-empty. Witness
`fixtures/quirks/body-response-description` + `TestQuirk_BodyResponseDescription`
exercise all three tiers. 14 integration goldens drifted — **all** changes are
`description` lines (no schema/`$ref`/type/code drift); e.g. `"Pet"`→
"Pet is a pet on offer.", `validationError` body at 422 → "Unprocessable Entity",
`genericError` body at `default` → "Default response".

## G2 — `name:` keyword and `swagger:name` annotation have disjoint scopes ✅ (fix/quirks-g2-name-scopes)

**✅ Fixed — exactly per the decision + implementation spec below.** Two small,
additive changes: (1) `internal/parsers/grammar/keywords.go` adds `CtxSchema` to
`KwName`'s contexts — `emitInlineKeyword` already *stored* the `name:` Property
regardless of context (so it was always stripped from prose), so this only
silences the spurious `parse.context-invalid` warning and declares the keyword
legal on schema fields; the full-Schema walker already ignores it safely (no
`schemaStringHandler` case, and `IsLegalForType` has no rule for `name` so
`checkShape` passes). (2) `internal/builders/schema/walker_classifiers.go`
`scanFieldDoc` reads `name:` via `GetString(KwName)` across the field's blocks
and lets it win over the `swagger:name` annotation (`fd.JSONName`); `fields.go`
is untouched — both the struct-field and interface-method carriers already
derive from `fd.JSONName`. Precedence `name:` > `swagger:name` > json tag > Go
field name holds in every context. **Symmetry diagnostic (added per Fred):** a
`swagger:name` annotation in a parameter or response-header context is still
rightfully dropped (the canonical form there is `name:`), but now emits a
`parse.context-invalid` warning pointing the author at the `name:` keyword —
catching the mirror-image mistake. `internal/builders/parameters/parameters.go`
+ `internal/builders/responses/responses.go` (the response one gated on
`in: header`, since a body field never consults the field name). Witness
`fixtures/quirks/name-keyword-universal` + `TestQuirk_NameKeywordUniversal` cover
a tag-bearing struct field, the keyword-beats-annotation precedence, legacy
`swagger:name` alone, a tag-less interface method (both `name:` and legacy
forms), a param-field regression guard, **and** the two new diagnostics (with
the misused `swagger:name` fields keeping their Go-derived names). **No golden
drift** beyond the new witness — nothing in the corpus had ever used `name:` on a
schema field (it was inert), so the change is purely additive. Full suite + lint
clean. **✅ Doc-site edits applied (fifteenth pass, commit c52bd1e):** items 1–4
below all done — keywords.md `### name` rewritten to universal (+ the param/header
warning note), annotations.md params-note + swagger:name section reframed as
legacy, model-definitions.md reframe + a new test-backed `namekeyword` witness
(concepts/models Account: tag-less `name:` rename + the precedence ladder).

**Symptom.** The two naming mechanisms are complementary, not interchangeable —
each is silently ignored where the other applies:

| field site | `swagger:name` annotation | `name:` keyword |
|---|---|---|
| `swagger:model` property | **honoured** (F5) | ignored (stripped from desc, no effect) |
| interface method | honoured | n/a |
| `swagger:parameters` field | **ignored** | honoured (feat f5845ea) |
| `swagger:response` header | — | honoured (feat 4de9680) |

A user who learns `swagger:name` for models reasonably tries it on a parameter
(no effect); a user who learns `name:` for params tries it on a model field (no
effect). The `name:` keyword is even *parsed* on a model field (it is stripped
from the description) but its rename is dropped — a **silent no-op with data
loss** (the comment line vanishes from the description and the rename never
happens). `swagger:name` on a param is, by contrast, a clean no-op.

The split is **historical, not semantic** — there is no reason a `name:` keyword
cannot name a model property (it is just the `properties` map key); it was only
ever wired into the param/header builder.

**Decision (2026-06-16, with Fred).** Make **`name:` the one canonical
field-naming keyword that works everywhere** — model fields, interface methods,
parameters, response headers. Keep **`swagger:name` as a retained back-compat /
legacy form** (still honoured, idiomatic for interface methods; not removed),
documented as "the older annotation; `name:` is the universal keyword." NOT the
diagnose-the-split option, and NOT a redundant `swagger:name` alias on the param
side (zero capability).

**Implementation spec.**

- Wire `name:` into the **schema field-name derivation** (the load-bearing
  schema-discovery path — land with a witness fixture + golden A/B per
  `feedback_schema_discovery_verify_with_witness`).
- Precedence, most-explicit-wins, **identical in every context**:
  `name:` keyword > `swagger:name` annotation > `json` tag > Go field name.
- Verify on a **tag-less interface method** (no `json` tag to compete) and on a
  tag-bearing struct field.
- `swagger:name` stays functional everywhere it is today; no deprecation
  diagnostic.

**Doc-site changes to apply on merge** (pre-drafted; flip when the branch lands):

1. `keywords.md` `### name` — rescope to "names **any** field (model property,
   interface method, parameter, response header)"; drop the "complementary, not
   interchangeable" split; state the precedence; note `swagger:name` is the
   legacy annotation form.
2. `annotations.md` `swagger:parameters` note — replace the "`swagger:name` is
   not consulted on a parameter field" caveat with the universal `name:`.
3. `annotations.md` / `model-definitions.md` `swagger:name` sections — reframe as
   the legacy annotation form, pointing at the `name:` keyword as canonical;
   keep the interface-method example (where it remains idiomatic).
4. Add a test-backed witness to the doc-site: a `name:` on a tag-less
   `swagger:model` field in `concepts/models` (or `concepts/refoverride`) so a
   golden proves the rename, mirroring the `swagger:name` interface example.

**Repro (pre-fix).** A `swagger:parameters` struct field `A string \`json:"a"\``
with `// swagger:name a_name` → param name stays `a`. A `swagger:model` field
with `// name: x` → property name unchanged.

## G3 — a JSON-literal example/default on a `$ref`'d field stays a raw string ✅ (fix/quirks-g3-ref-example-coercion)

**✅ Fixed.** The `$ref` override arm (`refOverrideCollector` in
`internal/builders/schema/walker.go`) carries no `Type` of its own — the type
lives on the referenced definition, not on the `allOf` sibling — so
`ParseDefault(val, SchemaTypeOf(&c.override), …)` dispatched on an empty type
and returned the literal verbatim. New `validations.CoerceJSONOrString`
(`internal/builders/validations/coerce.go`) coerces a JSON object/array literal
structurally and falls back to the unquoted string for anything else; the
`onString`/`onRaw` `default:`/`example:` arms now call it. **Scalars are
deliberately NOT coerced** (a bare `example: 42` against an unknown referenced
type must not silently become a number) — only `{…}`/`[…]` literals, the
witnessed case. Witness `fixtures/quirks/ref-example-coercion` +
`TestQuirk_RefExampleCoercion` (object example, object default, array example,
and the scalar-stays-string guard). Golden drift was confined to four
pre-existing silent witnesses of the same quirk — `bugs_2652` (PodSelector) and
the four `bugs_3125` variants (Value2) — each flipping a raw-string example to
the structured object; their inline assertions (`TestCoverage_Bug2652`,
`TestEmbeddedDescriptionAndTags`, `TestIssue2540`) were updated. No schema /
`$ref` / type / code drift. Note: a *model-level* (non-`$ref`) `example:` /
`default:` JSON literal (e.g. Book in `bugs/2540`) is a separate surface and
still rides as a raw string — out of G3's scope. Full suite + lint clean.
**✅ Doc-site edit applied (fifteenth pass, commit 0e57898):** added a test-backed
`refstructured` witness (concepts/examples Place — JSON-object example coercing
on a `$ref`'d field's allOf arm) to examples-and-defaults.md "On a defined-type
field", with the scalar-stays-string exception spelled out.

**Symptom.** On a *plain* field a JSON literal example is parsed into a
structured value (`example: {"k":"v"}` → an object). On a field whose type is a
`$ref` (so the example rides the `allOf` override arm), the same literal is
carried as a **raw string**:

```
"pet": { "allOf": [ {"$ref": "Pet"}, { "example": "{\"name\":\"x\"}" } ] }
```

i.e. `"{\"name\":\"x\"}"` instead of `{"name":"x"}`. Coercion runs on the
direct-field path but not on the override-arm collector, so the same annotation
yields an object on one field shape and a string on another. (`readOnly`,
`minimum`, etc. *do* ride the override arm correctly — only the JSON-literal
coercion is missing there.) Candidate: run the same example/default coercion when
lifting siblings onto the override arm.

**Repro.** `swagger:model` `Holder` with a field typed by another `swagger:model`
carrying `// example: {"k":"v"}` → inspect the property's `allOf[1].example`.
