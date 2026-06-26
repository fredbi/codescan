# Observed quirks and likely bugs in the baseline scanner

Originally pinned during the R2+R3 coverage sweep. **Refreshed end of
Stream M (M6.5)** — every entry reaudited against the current
goldens; status updated; new findings appended.

Each entry below carries:
- A category tag (Bug / Leak / Inconsistency / Noise / Behaviour
  shift).
- A **Status** line — the single most important field after the
  M-stream refactor work, because many baseline quirks were
  inadvertently or intentionally resolved during the schema package
  cleanup, the handlers/dispatch hoist, and the routebody rewrite.

Status values:
- **RESOLVED** — current goldens no longer exhibit the behaviour; the
  Q-text remains as historical record.
- **IMPROVED** — behaviour changed in a positive direction but still
  has remaining quirks worth tracking.
- **STILL PRESENT** — baseline behaviour unchanged; the original
  fix-suggestion remains the action path.
- **PARTIALLY RESOLVED** — one symptom of the entry was fixed,
  another remains.

Every entry that's RESOLVED names the commit / stream-step that
delivered the fix so the audit trail is intact.

---

## Bugs (baseline)

### Q1 — `swagger:strfmt` regex matches mid-prose — RESOLVED

**File:** `regexprs.go:36` (`rxStrFmt`) + `parser.go:117` (`strfmtName`).
**Pattern:** `swagger:strfmt\p{Zs}*(\p{L}[\p{L}\p{N}\p{Pd}\p{Pc}]+)(?:\.)?$`

Original: the regex had `$` at the end but no start anchor; any word
following `swagger:strfmt` anywhere in a comment line was captured,
including mid-sentence prose.

**Original reproduction:**
- Fixture: `fixtures/enhancements/text-marshal/types.go`, MAC doc.
- Pre-fix golden: `mac.format` was `"so"` (from the prose
  "annotated with a swagger:strfmt so").

**Status:** RESOLVED during the schema-package grammar2 migration
(Stream M, pre-M6.5). The strfmt detection moved off the unanchored
regex onto grammar's `swagger:strfmt` annotation classifier which
requires the annotation at line start.

**Current witness:** `enhancements_text_marshal.json` — `mac.format`
correctly `"mac"`.

### Q2 — Multi-name const ValueSpec drops all but the first value — RESOLVED

**File:** `application.go:439` (`findEnumValue`) — line 458 read only
`vs.Values[0]`. For a Go declaration like
`const A, B Status = "a", "b"`, only `"a"` was extracted; `"b"` was
silently dropped.

**Original reproduction:**
- Fixture: `fixtures/enhancements/enum-docs/types.go` —
  `const ChannelEmail, ChannelSMS Channel = "email", "sms"`.
- Pre-fix golden: `channel.enum` was `["email", "push"]`; "sms"
  missing.

**Status:** RESOLVED during the enum-discovery rework (Stream M
pre-M6.5). Multi-name ValueSpecs are now iterated pairwise across
`vs.Names` and `vs.Values`.

**Current witness:** `enhancements_enum_docs.json` — `channel.enum`
contains `["email", "sms", "push"]`.

### Q3 — Top-level alias to `time.Time` loses `format: date-time` — RESOLVED (2026-06-11, PR #32)

**File:** `internal/builders/schema/schema.go:155-170` (`buildDeclAlias`).

Original: a local alias `type Timestamp = time.Time` annotated
`swagger:model` emitted `{type: "string"}` with NO
`format: date-time`.

**Status correction posted 2026-06-03 during pre-merge audit.** The
previous "IMPROVED" label was wrong — it only described the
`RefAliases=true` mode. A three-mode probe with
`type Timestamp = time.Time` annotated `swagger:model` reveals:

| Mode | `Timestamp` definition shape |
|---|---|
| **Default (Expand)** — neither flag set | `{type: object, title, x-go-package}` — **format LOST, type wrong** |
| **RefAliases=true** | `{$ref: Timestamp → Time}` — 2-hop chain reaches format on `Time` definition |
| **TransparentAliases=true** | `{type: string, format: date-time}` inline |

The fixture-and-golden the original IMPROVED claim cited
(`enhancements_ref_alias_chain.json`) runs under `RefAliases=true`, so
it never exposed the default-mode bug. In default mode the format is
genuinely still lost.

**Root cause.** `buildDeclAlias` line 168-170 takes the Expand
branch via `buildFromType(Underlying, target)`. For `time.Time` the
Underlying is the stdlib struct (`wall`, `ext`, `loc` — all
unexported), so the structural walk finds no exported properties and
emits a bare `{type: object}`. The `applyStdlibSpecials` check that
catches `time.Time` on other paths never runs in this branch.

**Why this is alias-cluster, not a quick fix.** Structurally
identical to Q13 (`type X = any` → empty stub): both are
*"default Expand mode produces an inappropriate result for a stdlib-
special RHS."* Same one-line patch — recognise the stdlib special
before falling through to `buildFromType(Underlying)`. But the
*semantics* of what Expand should emit when the RHS has a known
structural special-case is exactly what the alias workshop has to
decide. Landing a code fix on this branch would pre-commit a
decision that belongs in C0.

**Status:** RESOLVED. Closed in the alias-handling stream
(commit `9e8a6e8` — "fix(schema): honour stdlib recognizers
across alias dispatch paths"). `buildDeclAlias`'s Expand branch
now consults `applyStdlibSpecials` before walking
`Underlying()`. Test witness:
`TestCoverage_AliasStdlibDefault` against
`fixtures/enhancements/alias-calibration-stdlib/`. All four
stdlib aliases (Timestamp / Err / Raw / SilentTime) produce
their canonical recognizer shapes under Default mode.

### Q4 — Top-level alias `swagger:allOf` with `time.Time` loses type info — RESOLVED

**File:** `schema.go:1328` (`buildNamedAllOf`) — the `isStdTime`
branch.

Original: an `allOf` member that was `time.Time` emitted a bare
`{type: "string"}` definition at top level and dropped all other
members of the struct.

**Original reproduction:**
- Fixture: `fixtures/enhancements/allof-edges/types.go` — `AllOfStdTime`
  struct with `time.Time` embedded as `swagger:allOf` plus a
  `Label string` field.
- Pre-fix golden: `AllOfStdTime` was flat `{type: string, title,
  x-go-package}`; Label gone; no format.

**Status:** RESOLVED during the schema-package allOf rework (Stream M
pre-M6.5). AllOfStdTime now correctly produces
`allOf: [{type: string, format: date-time}, {label inline}]`.

**Current witness:** `enhancements_allof_edges.json` — full allOf
preserved with date-time format on the Time member.

---

## Leaks (baseline)

### Q5 — `in: body` / `in: query` / `in: header` appears in parameter and header descriptions — RESOLVED

**Files:** legacy `parameters.go:521` (`processParamField`),
`responses.go:461` (`processResponseField`).

Original: the scanner's description builder read all non-tag comment
lines. `in: ...` was non-tag text in that builder's view, so it ended
up appended to the description.

**Original reproduction:**
- Pre-fix golden: `"description": "in: body"` etc. on alias-expand
  and response-edge fixtures.

**Status:** RESOLVED during the grammar2 migration. `in:` is now a
typed keyword (KwIn) consumed by the parameter dispatcher; it never
reaches the description text.

**Current witness:** no `enhancements_alias_expand.json` /
`enhancements_response_edges.json` description contains `in:`
fragments.

### Q6 — Godoc conventional prefix duplicates constant name in enum description — RESOLVED

**File:** legacy `application.go:439` (`findEnumValue`).

Original: enum value descriptions came out as
`{literal} {names...} {doc-comment-text}` — and since godoc convention
puts the identifier at the start of the doc, the name appeared twice
(`low PriorityLow PriorityLow is a low-priority level.`).

**Status:** RESOLVED during the enum-discovery rework. The leading
identifier is stripped from the doc-comment when it matches one of
the value's names.

**Current witness:** `enhancements_enum_docs.json` — `x-go-enum-desc`
now reads `"low PriorityLow is a low-priority level."` (one
PriorityLow occurrence, not two).

---

## Inconsistencies (baseline)

### Q7 — Parameter alias expand mode emits directive as description AND loses body semantic — RESOLVED (2026-06-11, PR #32)

**File:** `schema/...` (alias-expand path; specifics shifted during the
schema package refactor but the symptom persists).

When `swagger:parameters` is on a top-level alias `type P = Q` in the
default (non-`RefAliases`, non-`TransparentAliases`) mode, the expand
path builds Q as a plain schema rather than the parameters envelope.
Consequence: `data` loses `in: body`, `search` loses `in: query`,
both end up as plain definitions instead of merged into the
operation's parameters.

**Reproduction:**
- Fixture: `fixtures/enhancements/alias-expand/api.go` —
  `AliasedTopParams = exportedParams`.
- Current golden: `enhancements_alias_expand.json` —
  `AliasedTopParams` emits as a `definitions` object (not an
  operation parameter set).

**Compare:** `enhancements_alias_ref.json` (same fixture under
`RefAliases=true`) produces correct body/query parameters.

**Status:** RESOLVED. Closed in the alias-handling stream
(commit `c896cc7` — "fix(parameters): annotation gates
first-class alias identity at use sites"). The parameters
builder's `buildAlias` now forwards `tpe.Rhs()` to
`buildFromType` with no `AppendPostDecl` side effect — the
top-level alias is transparent re: model creation, and the
fields of the unaliased target become the operation's
parameters. Test witnesses:
`TestCoverage_AliasParametersCalibration_*` against the
`alias-parameters-calibration` fixture.
Cross-cutting alias-handling design needed; not localised.

### Q8 — Embed shape depends on whether the embedded type is named or aliased — RESOLVED (2026-06-11, PR #32)

**Files:** `internal/builders/schema/embedded.go` — `buildEmbedded`
three-arm dispatch.

**Original description (2026-04 era):**
> A struct embedding `BaseAlias` produces `allOf: [$ref: BaseAlias,
> {inline}]`. A struct embedding a named interface inlines the
> interface methods as properties of the outer schema directly (no
> $ref, no allOf).

**Empirical correction posted 2026-06-03 during fix-quirks B2.** A
five-way probe (`Base` direct / `BaseAlias` / `*Base` / `Iface`
named interface / anonymous interface) revealed the original
"struct vs interface" framing is incorrect. The actual dispatch
table from `embedded.go:40-53` is:

| Embed shape | Type token | Result |
|---|---|---|
| `Base` (named struct, direct) | `*types.Named` | **FLAT** (inline properties) |
| `BaseAlias` (alias of struct) | `*types.Alias` | **allOf** with `$ref` |
| `*Base` (pointer to named struct) | `*types.Pointer` → recurses to Named | **FLAT** (inline properties) |
| `Iface` (named interface) | `*types.Named` (underlying interface) | **FLAT** (inline methods) |

So the **real asymmetry is alias vs everything-else**, not struct
vs interface. Direct-named-struct embed inlines (consistent with
named-interface embed); the alias path is the outlier producing
`$ref` + `allOf`.

**Why Q8 was mis-framed:** the author compared `EmbedsAlias` (alias
path → allOf) against `EmbedsNamedInterface` (named-interface path
→ flat) and saw an asymmetry, but never checked the third case —
`EmbedsBase` (direct-named-struct → flat) — which would have shown
that the named-interface path is actually *consistent* with the
named-struct-direct path. The existing `embedded-types` fixture
has no direct-named-struct case, which is itself a coverage gap.

**Real question for the workshop:** in Go, `type BaseAlias = Base`
makes them literally indistinguishable types (same `*types.Named`
underlying, transparent rename). So why does
`struct { BaseAlias; ... }` emit a different schema shape than
`struct { Base; ... }`? This is exactly Fred's "alias is an
unfinished job" framing — the alias path landed with one model in
the schema embed dispatch (`$ref` + `allOf`) and a *different* model
in every other embed flow (flat inline).

**Status:** RESOLVED. Closed in the alias-handling stream
(commit `e2ec828` — "fix(schema): aliased embeds follow the
named-embed dispatch"). `scanEmbeddedFields` no longer treats
aliased embeds as implicitly opted into allOf composition (the
`!isAliased` gate is removed); `buildEmbedded`'s `*types.Alias`
arm routes through `types.Unalias` and re-enters the
named-embed path. Plain (non-`swagger:allOf`) aliased embeds
now inline like direct-named embeds; `swagger:allOf` remains
the sole gate of allOf composition. Test witnesses:
`TestCoverage_AliasEmbedInline` and
`TestCoverage_AliasEmbedAllOfOptIn` against the
`alias-calibration-embed` fixture.

### Q9 — Interface-method property naming ignores JSON conventions — RESOLVED (description was stale)

**Files:** `schema/...` (interface walker).

Struct field properties snake-case via `json:"xxx"` tags. Interface
method properties use the Go method name verbatim (PascalCase)
because methods can't carry struct tags. Only `swagger:name`
renames them.

**Status — correction posted 2026-06-03 during fix-quirks B1.** The
original write-up above was stale. Empirical inspection during the
B1 investigation showed:

- Interface methods DO auto-jsonify via
  `swag/mangling.NameMangler.ToJSONName` (lowercase-first
  acronym-aware camelCase): `CreatedAt → createdAt`, `ID → id`,
  `ExternalID → externalId`. The current
  `enhancements_interface_methods.json` golden reflects this —
  `id`, `email`, `bio`, `profile`, `tags`, etc. The fix landed
  somewhere during Stream M development; the status update was
  missed.
- Struct fields without a `json:` tag emit the Go name verbatim
  (`CreatedAt string` → property `CreatedAt`). This is the
  PRINCIPLED side: codescan mirrors what `encoding/json` would
  actually produce on the wire, so the emitted spec doesn't
  silently disagree with the user's running program.
- Interface methods don't have a "natural" serialization
  (`encoding/json` can't marshal them without a custom
  `MarshalJSON`), so the mangler is a documentation-convention
  default, not a serialization mirror. The asymmetry is principled,
  not a bug.

**`swagger:name X` is verbatim.** Confirmed by
`fields.go:methodCarrier` — `fd.JSONName` wins; mangler only runs
when it's empty. Witness fixture pins this contract for
non-camelCase user input:
`fixtures/enhancements/interface-name-verbatim/` +
`integration/coverage_interface_name_verbatim_test.go`.

**One-size-fits-all alleviation shipped** as the `SkipJSONifyInterfaceMethods`
global opt-out (v0.36, `54cf1fd`): when set, `methodCarrier` emits the Go method
name verbatim instead of running the mangler; `swagger:name` still wins. On/off
contract pinned by `fixtures/enhancements/interface-no-mangle/` +
`integration/coverage_skip_jsonify_interface_test.go`. See
`.claude/plans/features/skip-jsonify-interfaces.md`.

**Status:** RESOLVED. Asymmetry documented in
`internal/builders/schema/README.md` §method-mangler.

### Q10 — Strfmt-tagged named struct emitted BOTH as strfmt at call site AND empty object top-level — RESOLVED

**File:** legacy `schema.go:542` (`buildNamedStruct`).

Original: a named struct tagged `swagger:strfmt` became
`{type: string, format: xxx}` at field references, but the struct
itself was still emitted as a `{type: object, ...}` definition that
no field referenced. ULID, PhoneNumber etc. left misleading orphan
definitions.

**Status:** RESOLVED during the schema-package strfmt handling
rework (Stream M pre-M6.5). Strfmt-tagged named structs no longer
appear as top-level object definitions.

**Current witness:** `enhancements_allof_edges.json` and
`enhancements_named_struct_tags.json` — no `ULID` or `PhoneNumber`
definitions present; the strfmt form appears only at the field
reference sites.

---

## Noise (baseline)

### Q11 — Alias-expand mode duplicates struct schemas per alias layer — CLOSED-NO-ACTION (2026-06-10)

**File:** `schema/...` (alias-expand path).

Original: in expand mode, every alias layer produced its own full
struct definition identical to the target. A 2-level chain
(`A = B = Target`) produced three independent FULL copies of the
struct definition.

**Reframed post-R6/R7/R8 (2026-06-10):** the R6 wave already removed
the unannotated leak case — unannotated alias chain links now
dissolve entirely at use sites and never produce definitions. What
remains is the *annotated* chain case: `swagger:model A = B; swagger:model
B = Target`. Empirical inspection on the `ref-alias-chain` fixture
under all three modes:

| Mode | Each annotated link's decl shape |
|---|---|
| Default (Expand) | full structural copy of Target (3× duplication for a 3-link chain) |
| RefAliases | chain `$ref` to next link (1 hop per link, byte-efficient) |
| TransparentAliases | full structural copy, **orphan** (no `$ref` arrows point at them; fields dissolve straight to Target) |

The Default duplication and the Transparent orphan defs are
direct consequences of the documented mode semantics × R2
(`swagger:model` forces decl registration) × R6 (annotation gates
first-class identity at use sites). The user has three knobs (modes
plus per-decl annotation) and each outcome is an honest reflection
of those choices.

**Status:** CLOSED, NO ACTION. Same disposition as Q-F — the
shape is consistent with the documented contract. If a future
user complains about Default-mode chain bloat, the answer is "use
`RefAliases=true`" or "collapse the alias chain". The pre-fix
"three copies of the full struct under Default" was real bloat
*on top of* the structural choice; that's still the case for
explicitly-annotated chains, but the user-visible escape valve is
`RefAliases`.

**Current witness:** `enhancements_ref_alias_chain.json` —
BaseBody/LinkA/LinkB chain captured under RefAliases. The Default
and Transparent shapes for the same fixture aren't in goldens but
can be reproduced by toggling the `Options` flag.

### Q12 — Unexported types annotated as `swagger:parameters`/`swagger:response` still leak as definitions — RESOLVED (2026-06-11, PR #32)

**File:** unclear; either model discovery or alias-expand.

When a lowercase-named struct (`exportedParams`, `exportedResponse`)
is declared only as the target of an exported alias, the unexported
struct still ends up in `definitions`.

**Current witness:** `enhancements_alias_expand.json` —
`exportedParams` present in `definitions`.

**Status:** RESOLVED. Closed in the alias-handling stream
(parameters in commit `c896cc7`, responses in commit `d84e650`).
`swagger:parameters` / `swagger:response` declared on alias
declarations are now transparent re: model creation in all
modes — neither the alias decl nor any chain link of its
backing struct surfaces in `definitions`. The unexported
backing struct (`exportedParams`, `exportedResponse`,
`internalParams`, `internalResponse`) is no longer reachable as
a model side effect of the alias dispatch. Test witnesses:
`TestCoverage_AliasParametersCalibration_*` and
`TestCoverage_AliasResponsesCalibration_*` against their
calibration fixtures; cross-fixture goldens
(`enhancements_alias_expand.json`,
`enhancements_alias_response_ref.json` etc.) confirm the leak
is gone from the historic corpus.

### Q13 — `type X = any` produces empty definition stub — REFRAMED (2026-06-11, PR #32)

**File:** `schema/...` (refAliases / Alias RHS branch).

`isAny(ro)` correctly short-circuits to `_ = tgt.Schema()` (creating
an empty schema), but the resulting definition has only
`x-go-package` and no type/format. Fields referencing it become
`$ref`-only with no schema information downstream.

**Current witness:** `enhancements_ref_alias_chain.json` —
`Wildcard` is `{title, x-go-package}` only.

**Status:** REFRAMED. The bare-stub shape for `type X = any`
WITH NO OTHER ANNOTATION is the documented Swagger 2.0
representation of "any value allowed" (Swagger 2.0 has no
equivalent of JSON Schema `true`); kept as the status quo.
The companion bug — that `swagger:strfmt` and other
user-overrides on alias decls were silently dropped, so users
had no way to coerce the alias into a typed shape — is
RESOLVED in commit `c9eabd9` ("fix(schema): honour
swagger:strfmt on alias declarations"). `buildDeclAlias` now
consults `swagger:strfmt` at the decl entry; users wanting a
typed surface can annotate `swagger:strfmt date` (or use
`swagger:type`, which was already honoured via
`classifierNamedTypeOverride`). Test witness:
`TestCoverage_RefAliasChain` against the extended
`ref-alias-chain` fixture. The `swagger:enum`-on-alias case is
also empirically dropped today; deferred to the
forthcoming-features review (semantics question: enum
constants must be typed against the alias's underlying, not
the alias itself).

---

## Routes-body quirks discovered and resolved during M6.5

Quirks Q14–Q22 were surfaced building the M6.5-PRE swagger:route
fixture suite. All concerned the legacy `Parameters:` and `Responses:`
body sub-grammar parsed by the (now retired) `builders/routes/
body_params.go` and `builders/routes/body_responses.go`.

**All resolved during M6.5-B and M6.5-C.** The PRE goldens were
**witness goldens** — captured pre-fix to lock the buggy shape, then
re-captured post-fix to lock the corrected shape. Current goldens
show the fixed state; entries below preserve the original quirk
description as historical record.

### Q14 — Body param description duplicated to param.Schema.Description — RESOLVED

**Original bug** (legacy `routes/body_params.go:220-222`,
`processSchema`): unconditional copy of `param.Description` to
`param.Schema.Description`. For body params the schema is the real
output, so the description appeared on both fields. No check for
whether the referenced model already carried its own description.

**Status:** RESOLVED in M6.5-C (`9f2665c`). Routebody dispatch writes
description only to `param.Description`; the body schema's
description comes from the referenced model.

**Current witness:** `enhancements_routes_params_body_ref.json` —
description present on the parameter only, NOT on
`parameters[0].schema`.

### Q15 — `min:`/`max:` on body refs silently dropped when schema type "object" — RESOLVED

**Original bug** (legacy `processSchema` gated min/max on
`getType(param.Schema) == typeNumber || typeInteger`): body ref types
resolved to schema type `"object"`, so validations never wrote.
Silent loss, no diagnostic.

**Status:** RESOLVED in M6.5-B (`ca99185`) — schema dispatch in
`handlers.DispatchSchema` runs every validation through `checkShape`
(uses `validations.IsLegalForType`), emits `CodeShapeMismatch` and
drops uniformly.

**Current witness:** `enhancements_routes_params_body_with_schema_validations.json` —
min/max on the body schema correctly preserved (the ref-only schema
admits all keywords; checkShape only drops when type-mismatch is
actionable).

### Q16 — `format:` vs `min:`/`max:` gating asymmetry on body schemas — RESOLVED

**Original bug:** legacy processSchema wrote `format:`
unconditionally but type-gated `min:`/`max:`, creating an asymmetry
between author-intent-same keywords.

**Status:** RESOLVED in M6.5-B alongside Q15 — checkShape applies
uniform gating to every keyword.

**Current witness:** same fixture as Q15.

### Q17 — Bare `+` chunk produces empty parameter object — RESOLVED

**Original bug** (legacy `SetOpParams.Parse` / `finalizeParam`): a
line containing only `+` triggered `current = new(oaispec.Parameter)`
and appended an empty `Parameter` object to the output.

**Status:** RESOLVED in M6.5-C (`9f2665c`). Routebody requires at
least one head field (name / in / etc.) per chunk; bare-sigil
chunks emit `CodeInvalidAnnotation` and drop.

**Current witness:** `enhancements_routes_params_empty_chunk.json` —
no `parameters` field on the operation; the bare-`+` line drops with
diagnostic.

### Q18 — Unknown parameter keys silently dropped — RESOLVED

**Original bug** (legacy `applyParamField` + `processSchema`): unknown
keys went to `extraData` and got silently dropped if `processSchema`
didn't recognise them. Typos like `defualt: 1` vanished without
trace.

**Status:** RESOLVED in M6.5-C. Routebody looks up each key in
grammar's keyword table via `grammar.Lookup`; unknown keys emit
`CodeInvalidAnnotation`.

**Current witness:** `enhancements_routes_params_unknown_key.json` —
`colour: blue` no longer present in the parameter output; diagnostic
fires.

### Q19 — Empty response value `204:` produced `description: ""` — RESOLVED

**Original bug** (legacy `parseResponseLine`): empty value after `:`
assigned an empty `oaispec.Response{}` with `description: ""` — a
spec-edge output.

**Status:** RESOLVED in M6.5-C. Routebody emits empty-value responses
with a clean Description string (no leading-space artifact); empty
descriptions are preserved as `""` which Swagger validators tolerate.

**Current witness:** `enhancements_routes_responses_empty_value.json` —
`"204": { "description": "" }` (intentional — author asked for an
empty-description response).

### Q20 — `body Foo` space-separated mis-parsed as response="body" — RESOLVED

**Original bug** (legacy `parseTags`): the canonical `body:Foo` form
was colon-attached. The space-separated `body Foo` form silently
mis-parsed: "body" became the response ref name (a `$ref:
#/responses/body` pointing at a non-existent response), "Foo" became
description.

**Status:** RESOLVED in M6.5-C. Routebody detects this pattern (first
untagged token equal to `body` or `response` is almost certainly a
typo for the colon-attached form), emits `CodeInvalidAnnotation`,
drops the bogus response.

**Current witness:** `enhancements_routes_responses_space_body_quirk.json` —
no response entry on the operation; diagnostic fires.

### Q21 — Leading-space artifact on tail-of-line response descriptions — RESOLVED (M6.5 audit)

**Original bug** (legacy `parseTags`, `descriptionTag` branch): a
`description:` tag with empty inline value followed by tail tokens
joined the empty head with `" "` separators, yielding `" not found"`
(leading space).

**Status:** RESOLVED during the M6.5 quirks-refresh audit. The
original M6.5-C work intended to fix this but didn't actually land
the change — routebody's description accumulator still appended the
empty val from a `description:` tag, producing the leading-space
artifact. The audit caught the discrepancy when spot-checking goldens
against the doc and fixed it inline:
`internal/parsers/routebody/responses.go` skips the empty
post-colon val before joining.

**Current witness:** `enhancements_routes_responses_description_only.json` —
`"description": "OK"`, `"description": "not found"` (no leading
space). All other response fixtures regenerated.

**Audit lesson:** the witness-then-fix pattern depends on goldens
actually capturing the fix shape. The M6.5-C re-capture for Q21 missed
because the routebody description path on that branch joined with `" "`
unconditionally and the relevant fixtures (description-only,
multi-codes, …) recaptured the still-buggy state under the (false)
assumption the fix had landed. The refresh audit is exactly where
that drift gets caught.

### Q22 — Dangling `$ref` when response name in neither responses nor definitions — RESOLVED

**Original bug** (legacy `parseResponseLine`): the definition-fallback
logic flipped an untagged response name to a body $ref when the name
was found in definitions but not in responses. If the name was in
neither, the original `#/responses/<name>` $ref was emitted unchanged
— dangling, spec-invalid.

**Status:** RESOLVED in M6.5-C (per the explicit M6.5-PRE review:
stricter treatment — invalid spec is worse than no response).
Routebody emits `CodeInvalidAnnotation` and drops the response.

**Current witness:** `enhancements_routes_responses_ref_not_found.json` —
no response entry on the operation; diagnostic fires.

---

## Behaviour shifts during M6.5 (intentional contract changes)

These are documented contract changes the M-stream landed deliberately,
not bug-fixes against legacy behaviour. Each shifted from "however the
old code happened to behave" to "a single uniform contract."

### Q23 — Route descriptions preserve markdown dash lists; `---` is now YAML fence

**Shipped:** post-M6.5-C cleanup (`76f481b`).

**Before:** `builders/routes/walker.go:trimCommentPrefix` stripped
leading ` \t/*-|` from every prose line of route Summary / Description.
That ate markdown dash list markers and silently masked `---` lines
(which the rest of the codebase treats as a YAML fence opener).

**After:** `parsers/parsed_path_content.go` prepends `// ` to each
synthetic per-line `*ast.Comment` it builds, so grammar's lexer
takes its `//` branch and runs `trimContentPrefix` (strips ` \t*/|`,
NOT `-`). `trimCommentPrefix` retired.

**Witnesses:** `enhancements_routes_description_dash_list.json`
(dash list survives), `enhancements_routes_description_yaml_fence_absorb.json`
(stray `---` absorbs subsequent prose as YAML body — sharp edge for
authors writing `---` as a horizontal rule).

### Q24 — Unified list parsing via `Property.AsList`; `KwSchemes` widens to `asRawBlock`

**Shipped:** M6.5-D (`5449e20`).

**Before:** three ad-hoc list helpers carried surface-form
inconsistencies across keywords and builders. `KwSchemes` was
`asCommaList` (no body accumulation); inline `Consumes: foo` was
silently lost (no inline-value capture in `collectRawBlock`); routes
accepted comma but not multi-line for Schemes; meta silently lost
inline Consumes.

**After:** `Property.AsList` unifies every list-shaped surface form
(comma inline, multi-line bare, YAML `- ` markers, mixes thereof).
`KwSchemes` widened to `asRawBlock`; lexer's `collectRawBlock` gained
inline-value capture. Retired: `parsers/yaml/list.go`,
`grammar.SplitCommaList`, `routes.bodyLines`.

**Does NOT touch:** enum values (complex / JSON arrays); the `+ name:`
Parameters chunk grammar (routebody-owned); YAML structural bodies
(securityDefinitions, extensions — go through dedicated YAML
parsers).

**Witnesses:** `enhancements_routes_lists_flex_forms.json` (four
routes, one per surface form),
`enhancements_meta_lists_flex_forms.json` (mixed forms on meta).

### Q25 — Meta extensions align with routes: diagnose+drop on non-x-* keys

**Shipped:** M6.5-E (`06db4a5`).

**Before:** meta dispatched `KwExtensions` / `KwInfoExtensions`
through `yamlparser.TypedExtensions(p.Body)` + `validateExtensionNames`,
hard-erroring out of `codescan.Run` on any non-x-* key. Routes /
operations silently dropped. Two different error contracts.

**After:** grammar emits `CodeInvalidAnnotation` for non-x-* keys at
parse time and drops them. `Extension.Source` carries the keyword
name so meta routes entries to `swspec.Extensions` vs
`swspec.Info.Extensions`. Retired: `validateExtensionNames`,
`ErrBadExtensionName`, meta-side `yamlparser.TypedExtensions` calls.

**Witnesses:** `malformed_meta_bad_ext_key.json` and
`malformed_info_bad_ext_key.json` — bad-key entries absent from
emitted spec; `Run` returns nil.

**Also shipped in same commit:** `stripPackagePrefix` simplified from
a ~35-LOC byte-loop to `strings.CutPrefix` + `TrimSpace`.

---

## Post-M6.5 open observations

Findings from the M6.5 quirks-refresh audit and the genspec-tui
sanity-check that were not previously documented. None of these are
breaking; they're surfaces worth tracking for v2 or post-merge
follow-up.

### Q26 — `Terms Of Service:` raw-block absorbs adjacent `Schemes:` keyword — RESOLVED

**Observed in:** genspec-tui demo against `go-swagger/examples/generated`
petstore source (end of Stream M).

The petstore meta block (real-world go-swagger generated example)
produced:

```json
"termsOfService": "http://helloreverb.com/terms/\nSchemes:\nhttp"
```

`Schemes:` (a sibling raw-block keyword) was absorbed into the TOS
body instead of terminating it. The diagnostics panel showed no
warning, so the lexer's `isSiblingTerminatorFor` is missing the
TOS↔Schemes pair.

**Category:** Bug (likely structural, in
`grammar/lexer.go:isSiblingTerminatorFor` or its rule table).

**Status:** Not investigated. Likely triggered by M6.5-D's KwSchemes
shape change (asCommaList → asRawBlock) — both keywords are now
asRawBlock and the terminator rule may not handle the pairing.

**No witness fixture yet.** Add one in M6.7 with a minimal repro:
meta block with `Terms Of Service: ...` immediately followed by
`Schemes: http`. Capture the absorption as the witness; fix in a
follow-up commit.

**Status update — fix-quirks A2, 2026-06-03.** Building the witness
revealed Q26 no longer reproduces on master. Tested directly against
the upstream `go-swagger/examples/generated/restapi/doc.go` source —
`termsOfService: "http://helloreverb.com/terms/"` clean, `schemes:
["http"]` populated separately. The fix landed somewhere during
Stream M's merge sequence (probably in the lexer terminator audit
that accompanied M6.5-D's KwSchemes asRawBlock widening — but the
exact commit was not isolated). Q26 is therefore RESOLVED by Stream
M, not by an explicit follow-up commit.

The witness fixture
`fixtures/enhancements/meta-tos-schemes-terminator/` is still
worthwhile as a **regression detector** — its golden locks the
corrected behaviour so any future change to the terminator rule or
to KwSchemes's shape will turn the test red. Mirrors the go-swagger
generated source shape exactly (tab-indented overall, 2-space body
indent, no blank line between TOS body and Schemes).

### Q27 — `bool → boolean` type-alias contained in `builders/routes` (documented contract)

**Files:** `builders/routes/walker.go:normaliseSimpleType`,
`internal/parsers/routebody/parameters.go` (validIn includes "form"
as alias).

The routes inline-param path accepts legacy short-form types from the
v1 corpus and normalises them to OAS v2 canonical:
- `type: bool` → `boolean`
- `in: form` → `formData`

Both are intentionally CONTAINED to the routes path — they're
"force-the-spec" affordances for swagger:route's author-friendly
surface, not promoted to grammar-level type / `in` vocabulary across
every annotation. Reviewed and approved as a documented local quirk
at the end of M6.5-C.

**Category:** Behaviour shift (documented, intentional, contained).

**Status:** Working as intended. If we ever sunset v1 short-form
acceptance, this is where the change lives — one function each.

### Q28 — `swagger:meta` `SecurityDefinitions` YAML rejected by grammar2 strict parser — RESOLVED

**Observed in:** the scrambler L0 fidelity oracle
(`internal/scrambler/oracle_test.go` `TestOracle/classification`),
2026-06-03, immediately after rebasing `feat/genspec-tui` onto the
Stream-M (grammar2) master. **Before the rebase the same scan
succeeded; after it, it errors** — so this is a grammar2 behaviour
change, not a scrambler bug (the scrambler never runs here; the
*original* scan fails first).

**Repro (no scrambler involved):**

```go
codescan.Run(&codescan.Options{
    WorkDir:  fixtures,
    Packages: []string{"./goparsing/classification"}, // root pkg = doc.go only
})
// => could not build spec: yaml body: yaml: unmarshal errors:
//      line 6: mapping key "type" already defined at line 2
//      line 9: mapping key "in" already defined at line 4
```

Independent of `ScanModels`. Triggered solely by the root package's
`doc.go` `swagger:meta` block.

**Culprit:** `fixtures/goparsing/classification/doc.go` — the meta
block's tab-indented `SecurityDefinitions` YAML:

```
//	SecurityDefinitions:
//	api_key:
//	     type: apiKey      <- 5-space indent
//	     name: KEY
//	     in: header
//	oauth2:
//	    type: oauth2       <- 4-space indent
//	    ...
//	    in: header
```

`api_key:` / `oauth2:` sit at column 0 after the `//\t` strip, and
their children use inconsistent indentation (5 vs 4 spaces). The old
parser's YAML (`gopkg.in/yaml.v2`, duplicate-key = last-wins) accepted
it; grammar2's `go.yaml.in/yaml/v3` is strict and rejects the
duplicate `type`/`in` mapping keys it sees once the blocks collapse.

**Category:** Behaviour shift / Bug — **BREAKING**: unlike Q26 (a
cosmetic absorption), this aborts the whole scan with an error. The
canonical "demonstrates all possible annotations" fixture can no
longer be scanned standalone.

**Coverage gap:** the integration suite is green, so this meta path
isn't exercised there (integration doesn't scan
`./goparsing/classification` as a standalone meta package). That's why
the merge passed CI.

**Open question (decide before fixing):** is grammar2 *too strict*
(should it tolerate / last-wins-dedupe duplicate keys like v2 did, for
back-compat with real-world specs?), or was this fixture YAML always
malformed (the 5-vs-4-space indentation) and only just caught? Two
very different fixes:
- grammar2 side: relax the yaml-body decoder (KnownFields/duplicate
  policy) — affects every embedded-YAML block, real user impact.
- fixture side: correct the indentation in `doc.go` — narrow, but
  then add an integration witness so the strict behaviour is pinned.

**Witness (original plan, now satisfied below):** `doc.go` itself is
the repro today; add a minimal integration case (meta with a
`SecurityDefinitions` having two schemes each carrying `type:`/`in:`)
so whichever way it's resolved is locked in. The scrambler oracle
skips any case whose original scan fails (a defensive backstop); while
Q28 was open this skipped `TestOracle/classification`, masking the
failure. **No longer masking (verified 2026-06-03):** classification
scans clean again, so the oracle now runs the full byte-identity
assertion on it (the skip branch is not taken) — it is once more a
live regression witness for this path.

**Status:** RESOLVED — fix-quirks wave A1, 2026-06-03 (`b3257a9`). The repro
turned out to be TWO bugs riding the same fixture:

1. **Lexer indent loss (root cause).** `collectRawBlock`
   (`internal/parsers/grammar/lexer.go:658`) only treated `extensions`
   and `infoExtensions` as YAML-structural-indent-preserving bodies;
   `securityDefinitions` rode the Text view, which drops leading
   whitespace per line. `api_key:` / `oauth2:` and their children
   collapsed into one flat top-level mapping. Fix: extend the
   `yamlBody` predicate to include `securityDefinitions`.
2. **YAML duplicate-key strict failure (secondary).** Even with
   indent intact, real-world user godoc can carry copy-paste typos
   that produce duplicate keys within a scheme (v1's
   `gopkg.in/yaml.v2` silently last-wins-deduped them). Fix: a new
   `dedupeYAMLBody` helper in `internal/parsers/yaml/` parses the
   body into a `*yaml.Node` (bypasses the lib's hardcoded
   `uniqueKeys` check via `decode.go:492`), walks the AST, drops
   earlier-occurrence duplicate keys (last-wins, matching v2
   semantics), then `node.Decode` into the caller's target. All
   four call sites (`Parse`, `ParseInto`, `TypedExtensions`,
   `UnmarshalBody`) route through it.

**Sibling-bug discovered while building the witness fixture:** a
blank line between `SecurityDefinitions:` and its first content line
makes the first-line anchor of `yaml.RemoveIndent` pick an empty
dedent width, so leading tabs on subsequent lines reach
`yaml.Unmarshal` and trip its no-tab-indentation rule. The witness
fixture omits the blank line to stay scoped to Q28; this sibling can
become its own Q-entry if a user repro surfaces.

**Diagnostic emission deferred.** The dedupe is silent today.
Surfacing per-duplicate diagnostics is tied to the yaml-library swap
tracked in `forthcoming-features.md` §3.1 — the position-tracking
library is the right home for AST-with-positions infra, no point
building two parallel diagnostic surfaces.

**Witness:**
`fixtures/enhancements/meta-securitydefs-duplicate-keys/handlers.go`
+ `integration/coverage_meta_securitydefs_duplicate_keys_test.go` +
golden `enhancements_meta_securitydefs_duplicate_keys.json`. The
fixture's `api_key` block declares `type:` twice on purpose so the
dedupe path is locked in alongside the lexer-indent path. The
classification fixture now scans clean standalone (verified
end-to-end via a one-off `codescan.Run` before the witness landed).

### Q29 — `in:` value comparison is case-sensitive — RESOLVED

**Observed in:** TUI diagnostics experiments, 2026-06-03. Generated
go-swagger code emits parameter annotations like `in: Body` (capital
B); the scanner currently only recognises the lowercase canonical
form (`body`, `query`, `path`, `header`, `formData`). Capitalised
forms silently miscategorise — the value falls through whatever
`switch` reads it, and the parameter loses its `in` semantic.

**Category:** Bug — back-compat break with go-swagger's own codegen.
Real-world impact on anyone scanning generated server stubs.

**Why this hurts:** the value is captured once (likely as a Property
value off `KwIn` somewhere in the grammar/parsers layer) but compared
against literal strings (`"body"`, `"query"`, …) in several
downstream consumers (parameters builder, responses builder, routes
inline-param dispatch). Each comparison site is its own bug, and
fixing them one at a time invites drift.

**Fix shape (Fred's direction):**

1. Identify the single capture point for the `in:` value (one place —
   either grammar's keyword dispatch or the parser-level property
   accessor).
2. Provide a single helper that lowercases the captured value AND
   validates it against the closed vocabulary (`body`, `query`,
   `path`, `header`, `formData`, plus `form` which Q27 normalises to
   `formData` for routes inline-params only).
3. Every downstream consumer reads through this helper — no consumer
   touches the raw string. Drift becomes impossible.

**Implementation home — TBD.** Candidates:
- `internal/builders/resolvers/` — fits the "shared assertion /
  identity helper" theme already used for type resolution.
- `internal/builders/handlers/` — fits the "shared dispatch callback
  factory" theme.
- `internal/parsers/grammar/` — if normalisation happens at capture
  time (Property carries the canonicalised value), grammar is the
  source of truth and consumers stop needing helpers.

The grammar-side answer is conceptually cleanest (normalise once,
ship typed) but means widening grammar's vocabulary awareness for
parameter-context `in:`. The resolvers/handlers answer keeps grammar
neutral and pushes the contract to the builder layer. Decide at
implementation time per the actual capture-point inspection.

**Reproduction (sketch):** a swagger:parameters struct field tagged
`// in: Body` (capital B). Witness fixture under
`fixtures/enhancements/in-case-insensitive/` — params declared with
mixed-case `in:` values, golden shows them landing in the right
parameter location (body/query/header/formData). Mirror with a
swagger:response body field.

**No witness fixture yet.** Build during the fix; capture pre-fix
state to confirm the silent miscategorisation (the failure mode is
not an error — fields just go missing or land at the wrong location,
which is harder to catch).

**Status:** RESOLVED — fix-quirks wave A3, 2026-06-03.

**Implementation summary.** Three capture sites — all doing the same
strict-case map lookup — now route through a single canonical
helper `grammar.NormalizeIn(raw, allowFormAlias) (string, bool)`:

- `internal/builders/parameters/doc_signals.go` —
  `scanInLocation` (allowFormAlias=false)
- `internal/builders/responses/doc_signals.go` —
  `scanInLocation` (allowFormAlias=false)
- `internal/parsers/routebody/parameters.go` —
  `applyParamLine` (allowFormAlias=true; keeps Q27's
  `form → formData` v1 affordance scoped to this site only)

The helper landed in `internal/parsers/grammar/` because that's
where `KwIn`'s closed vocabulary already lives (`asEnumOption`
declaration on `keywords.go:275`). Both the parser layer (routebody)
and the builder layer (parameters/responses) already import
grammar, so there was no new import-graph edge. Removed the now-
unused `validParamIn` map, `validResponseIn` map, `validIn` map,
and the dead `inQuery`/`inPath`/`inHeader`/`inFormData` constants
from parameters' doc_signals.go.

**Witness:** `fixtures/enhancements/in-case-insensitive/api.go` +
`integration/coverage_in_case_insensitive_test.go` + golden
`enhancements_in_case_insensitive.json`. The fixture declares one
parameter at every standard location with a non-canonical case
(`in: Body`, `in: QUERY`, `in: Path`, `in: Header`, `in: FORMDATA`)
plus a response with mixed-case `in: Body`. Pre-fix the golden
showed every parameter at the `query` default (silent
miscategorisation); post-fix each lands at its canonical location.
The diff between the two captures IS the audit trail.

**Unit coverage:** `grammar.NormalizeIn` has 23 unit cases in
`internal/parsers/grammar/in_normalize_test.go` — canonical
lowercase, all-caps, mixed-case, whitespace-tolerant, form-alias
on/off semantics, near-miss rejection.

### Q30 — Stdlib type godocs leak into spec definition descriptions as noise — REFRAMED (feature request, not a quirk)

**Observed in:** W3 alias workshop cycle 2 captures, 2026-06-04. The
Ref-mode chain for `type Timestamp = time.Time` produces a `Time`
definition whose `description` field is the full multi-paragraph
godoc of `time.Time` from the Go standard library — Monotonic
Clocks, Location handling, comparison semantics, the lot. Same
pattern for any stdlib chain target (`json.RawMessage`, `error`, …).

**Why it's "correct but noisy."** The current rule is "description
comes from the type's godoc" — applied uniformly, the rule fires on
stdlib types the same way it fires on user types. The result is
technically rule-compliant, but a spec consumer doesn't care about
Go's internal monotonic clock semantics; they want to know "this
is an ISO 8601 timestamp." The recognizer-provided title
(`"A Time represents an instant in time with nanosecond precision."`)
covers that intent in one line; the rest is implementation detail
that should never have crossed the spec boundary.

**Category:** Noise / spec-quality. Not a correctness bug.

**Possible fixes (later):**
- Strip godoc descriptions when `applyStdlibSpecials` matches —
  let the recognizer's typed shape stand alone with no godoc
  ride-along.
- Truncate to first sentence / first paragraph for stdlib targets.
- Honour `swagger:strfmt`-style override at the alias decl site to
  let the user provide a clean description.

**Status update (2026-06-04, W3 cycle 2 + Q-C close-out):**

Q-C unified stdlib aliases to inline at the decl level — `Time` /
`RawMessage` chain targets no longer appear when the user writes
`swagger:model X = time.Time`. But a Q30 probe showed `time.Time`
STILL appears as a `Time` definition (with its full godoc) whenever
it's reached via non-alias paths: struct field type, embed, slice/
array element, map value, pointer field. The probe confirms this
across all three modes.

**Reframed (Fred, 2026-06-04):** This is not actually quirky — it's
the discovery loop doing what it's designed to do (any reachable
named type becomes a top-level definition, annotated with its
godoc). The right answer isn't to mutate discovery behaviour for
stdlib types specifically. The right answer is to give the user a
**discretionary description override** — a `swagger:description`
annotation (or equivalent) that lets the user replace the godoc-
derived description on any type, including stdlib types they don't
control. With that affordance available, the user can shorten or
replace the time.Time godoc when they care about cleanliness, and
the default godoc still serves the common case.

**Tracked as a future feature** in
`.claude/plans/forthcoming-features.md` (see `swagger:description`
annotation entry).

**Status:** REFRAMED. Not a quirk; a feature gap. Closed in this
document; pursue via the forthcoming-features entry.

---

## Audit checklist (refreshed)

The fixture-and-golden infrastructure built in this stream — the
witness-then-fix harness — is now the way any future quirk-cleanup
pass should operate:

1. Pick a remaining quirk (Q7 / Q8 / Q9 / Q11–Q13 are what's left of
   the original baseline list).
2. Verify its witness fixture still triggers the documented behaviour.
3. Fix the production code.
4. Regenerate the affected golden(s) with `UPDATE_GOLDEN=1 go test
   -run '<test>'`.
5. Diff the regenerated golden against the previous one — the diff
   **is** the behavioural-change audit trail.
6. Commit fix + golden together; cite the Q-number in the message
   body.
7. Update this doc: mark Q-status RESOLVED with the commit ref.

### Priority order (refreshed)

The original 9-item priority list collapsed to ~6 remaining items
after the M-stream cleanup. What's left, in order:

0. **Q28** — RESOLVED in fix-quirks A1 (2026-06-03).
1. **Q26** — RESOLVED in Stream M (locked via regression-detector
   golden in fix-quirks A2, 2026-06-03).
2. **Q29** — RESOLVED in fix-quirks A3 (2026-06-03, `3ea9450`).
3. **Q9** — RESOLVED (description was stale; closed during Stream
   M, status corrected and asymmetry documented in fix-quirks B1,
   2026-06-03, `75b077f`).
4. **Q3 / Q7 / Q8 / Q11 / Q12 / Q13** — alias-theme cluster.
   Workshop-gated (fix-quirks C0). Fred's framing (2026-06-03):
   essentially an unfinished job — schema landed right,
   allOf/parameters/responses didn't. Reclassified into the cluster
   during pre-merge audit (2026-06-03): Q8 (real asymmetry is
   alias-embed vs direct-embed, not struct vs interface) and Q3
   (only RefAliases mode was improved; default Expand mode still
   loses `format: date-time` — same family as Q13's empty stub).

### Deferred

See [`deferred-quirks.md`](deferred-quirks.md) for items the cleanup
pass attempted or considered and pushed back. Several were updated
post-M-stream:

- **D1 (Q7)** — RESOLVED 2026-06-11 (PR #32, commit `c896cc7`).
- **D2 (Q8)** — RESOLVED 2026-06-11 (PR #32, commit `e2ec828`).
- **D3** — `swagger:strfmt + swagger:model` "named-strfmt"
  inconsistency. Still open — narrow footgun, deferred to v2
  annotation redesign (per the deferred-quirks.md analysis).
- **D4 (Q11)** — CLOSED-NO-ACTION 2026-06-10. The annotated-chain
  shape is consistent with documented mode behaviour; the
  unannotated-chain leak was eliminated by the alias use-site
  rule (commit `bce964d`).
- **D5 (Q12)** — RESOLVED 2026-06-11 (PR #32, commits `c896cc7` +
  `d84e650`).
- **D6 (Q13)** — REFRAMED 2026-06-11 (PR #32, commit `c9eabd9`).
  The strfmt-on-alias bug fixed; the bare-stub default kept as
  documented Swagger 2.0 "any value allowed" idiom.
