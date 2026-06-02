# Forthcoming features — parked, not forgotten

Date: 2026-04-21
Status: living catalog
Companion: `.claude/plans/grammar-parser-architecture.md` /
`.claude/plans/grammar-parser-tasks.md` /
`.claude/plans/workshops/*.md`

This document exists because small flagged items surface during
design/implementation and tend to evaporate unless collected. The
large architectural deferrals (OAI 3.x, pluggable annotation
styles, incremental parsing, …) live in the architecture doc's
A.4 section — those are durably tracked and are NOT duplicated
here.

This is the smaller backlog: implementation-level flags, sub-feature
seams we preserved room for, and capability audits pending.
Everything here is real work the project wants eventually; none is
a dependency for the v2 parser-migration release (v0.34 / v0.35).

**When a new flag surfaces during a commit or workshop**, add it
here and cite the origin. When a flag is picked up and landed,
remove the entry and note the commit in the change history.

---

## 1. Enum (W2 §4 forthcoming)

### 1.1 `swagger:enum` on `var` and richer Go expressions

**Origin:** W2 §4.1 (workshop 2026-04-21); raised by Fred as a real
use case (enum values that aren't const-representable — `[]byte`,
typed structs, typed maps).

**Scope:** lift the hard `token.CONST` gate at
`internal/scanner/scan_context.go:312` and extend `findEnumValue`
(`scan_context.go:337`) to accept value expressions beyond
`*ast.BasicLit`:

- `var (...)` declarations (the immediate need).
- Composite literals (`Status{Label: "foo", Code: 1}`).
- Call-result initializers — `go/constant` where compile-time-known.

**Prerequisites:** parity harness (P4) must already lock current
const-only behavior so the expansion doesn't regress fixtures.

**When to revisit:** after P5 lands — this is a scanner feature, not
a parser feature, so it's logically independent of the migration.

### 1.2 Inline YAML-block enum values

**Origin:** W2 §4.2 — recalled by Fred as "a hot debate at some
point"; no code, no fixture, no doc ever landed.

**Proposed surface:**

```go
// enum:
//   - complex: {object: value}
//   - another: [array, element]
```

**Scope:** in `internal/parsers/enum/`, add a third shape-detection
branch — if the `enum:` keyword has a block body (P2.3 captures
it as `Property.Body`), hand the body to `internal/parsers/yaml/`
and fold the result into the returned `[]any`.

**When to revisit:** after non-scalar scalar enums are actually
needed by real users; not before.

### 1.2a v1 enum quirks (concrete fixes, not parity items)

**Origin:** surfaced by `TestCoverage_EnumOverrides` golden
(`fixtures/integration/golden/enhancements_enum_overrides.json`,
commit `4cf0c41`). These are **v1 bugs** P5.1 will fix during the
schema-builder migration — listed here as a crossref so they're
visible from the feature-backlog side too. Primary home for these
items is P5.1's task entry in `grammar-parser-tasks.md`.

- Comma-list value parsing preserves literal leading whitespace
  (case B renders `["low", " medium", " high"]`). Fix: trim each
  value in `internal/parsers/enum/`.
- `swagger:enum TypeName` with no matching consts silently drops
  the annotation (case D: type-less schema property). Fix: emit
  `parse.context-invalid` diagnostic + surface the zero-values
  outcome.
- Stale `x-go-enum-desc` retained when the inline override wins
  (case E). Fix: drop the vendor extension when const-derived
  values are superseded.

**When to revisit:** P5.1 — these are migration-commit obligations,
not deferred items.

### 1.3 Non-scalar enum emission parity audit

**Origin:** W2 §4.3 — v1's `ParseEnum` JSON path accepts objects,
arrays, null, but no fixture demonstrates emission. Before any
release claims "OpenAPI-complete enum support", add object-valued
and array-valued enum fixtures and confirm the full pipeline
(parser → sub-parser → schema builder → spec) round-trips them.

**Blocker for:** marketing "spec-complete" enum handling. Not a
blocker for v0.34 which inherits v1's scalar-only corpus.

---

## 2. Example / examples (W3 — deferred)

**Workshop decision 2026-04-21:** W3 is **not held** as a pre-P5.1
workshop. Rationale: the OpenAPI `example:` (single value) and
`examples:` (array / OAI 3 map keyed by name) have essentially the
same shape as `enum:` — an array of values of any type — and
neither was supported in v1 beyond a single scalar raw-value-per-
field. The parser-migration release inherits v1 parity exactly
(`example:` stays `ValueRawValue` per `keywords_table.go`); W3 is
deferred to when the project actually ships richer example support.
When W3 does get held, it inherits W2's pattern directly:

### 2.1 Array-of-any-type example values

**Origin:** W3 deferred 2026-04-21; OAI 2 + OAI 3 `example:` /
`examples:` distinction.

**Expectation:** W3 will follow W2's
`raw-string → sub-parser → field-type-coerced-at-bridge-tagger`
pattern. Natural home: a shared `internal/parsers/values/`
subpackage (or dedicated `internal/parsers/example/`), reusing
the shape-detection logic W2 establishes for enum.

**OpenAPI quirks to address at that time:**
- `example:` (singular) — one value bound to a schema / parameter.
- `examples:` (plural) — array in OAI 2; in OAI 3 a map keyed by
  example-name (`examples: { good: {...}, bad: {...} }`). The
  two OAI versions disagree on shape; the sub-parser has to decide
  what to emit at the IR level.
- Example values may be any JSON type including objects, arrays,
  null, and `$ref` in OAI 3. Non-scalar support inherits W2 §1.3's
  audit requirement.

**When to revisit:** alongside OAI 3.x keyword onboarding (§C14
deferred item in the architecture) or when a concrete user request
surfaces. Not on the P5 path.

---

## 3. Sub-parser layer

### 3.1 `goccy/go-yaml` POC for token-level YAML positions

**Origin:** architecture §5.1 / A.4; flagged 2026-04-21 during P2.5.

**Scope:** today `internal/parsers/yaml/` wraps `go.yaml.in/yaml/v3`,
which gives coarse line/column in errors only. If LSP wants to
highlight the *specific* key inside a 20-line embedded YAML block,
the incumbent library is insufficient. `goccy/go-yaml` exposes a
token-level lexer.

**When to revisit:** when the LSP server (out of scope for v2.0)
actually needs per-token YAML positions. A POC comparing both
libraries is the first step.

**Companion follow-on — duplicate-key diagnostic emission.**
Tracked here because it shares the same underlying capability —
walking the YAML AST with per-token positions. Q28 (resolved in the
fix-quirks wave, 2026-06) silently dedupes duplicate mapping keys
last-wins; that fix has no warning surface because
`go.yaml.in/yaml/v3` exposes no public `UniqueKeys(bool)` knob and
building a parallel diagnostic infra around the AST-mutate workaround
was judged too hacky for now. When the position-tracking yaml library
lands, wire it: each duplicate key drops with a
`CodeDuplicateMappingKey` diagnostic carrying the file:line:col of
both occurrences. The dedupe semantics (last-wins) stays — only the
warning becomes visible. Re-check the Q28 witness fixture
(`fixtures/enhancements/meta-securitydefs-duplicate-keys/`) at the
same time.

### 3.2 `splitCommaList` respect quoted strings

**Origin:** `internal/parsers/grammar/ast.go` godoc — flagged during
P3.3.

**Scope:** today `splitCommaList("a, \"hello, world\", b")` yields
four entries because the comma inside the quoted string isn't
protected. Architecture §2.1 explicitly called out respecting
quoted strings. No fixture exercises this; parity harness will
reveal if a real fixture needs it.

**When to revisit:** first time the parity harness flags a
discrepancy, or pre-emptively if we land a JSON-escaping-aware
enum sub-parser that shares the same concern (W2 §1.2 route).

### 3.3 Skip-jsonify-interfaces — opt out of the interface-method mangler

**Origin:** Q9 close-out (2026-06-03, fix-quirks B1).

**Scope:** today `internal/builders/schema/` runs the
`swag/mangling` `ToJSONName` transform on every interface-method
property name when the author didn't provide a `swagger:name`
override. This is a one-size-fits-all convention
(`CreatedAt → createdAt`, `ID → id`) — see schema/README.md
§method-mangler for the rationale and the principled struct-vs-
interface asymmetry.

It will not always be what the author wanted. Examples:
- An interface that's already named for its JSON shape (`Get_user`
  or similar non-Go-idiomatic spellings).
- A codebase that has its own canonical-name discipline and wants
  codescan to stay out of the renaming business entirely.

**Shape:** a global option
(`Options.SkipJSONifyInterfaceMethods bool` or similar), defaulting
to false so existing behaviour is preserved. When true,
`fields.go:methodCarrier` skips the
`s.interfaceJSONName(fld.Name())` call and falls back to the Go
method name verbatim. `swagger:name` continues to win regardless of
the option.

A per-decl annotation (`swagger:no-mangle` on the interface or
package) would also work and is more granular, but adds a new
annotation surface — the global option is simpler for v2.0.

**When to revisit:** first time a user requests it, or when v2.0's
Options surface gets its next pass.

**Witness fixture (locks current behaviour):**
`fixtures/enhancements/interface-name-verbatim/` —
the `swagger:name` verbatim contract that any opt-out implementation
must still respect.

### 3.4 `swagger:description` — discretionary description override

**Origin:** Q30 close-out (2026-06-04, W3 alias workshop cycle 2).

**Scope:** today, every type that surfaces in `definitions` carries
its godoc as the `description` field — including reachable stdlib
types the user doesn't control. `time.Time` is the recurring
example: a 2305-character description on monotonic clocks, location
handling, and `==` semantics drags into any spec whose user code
embeds, references, or aliases `time.Time`. The user has no way to
shorten or replace this without forking and modifying the source.

The discovery + godoc-annotation behaviour is correct per the
current rules — every reachable named type becomes a definition
annotated with its godoc. The issue is the **lack of an override
affordance** for the user to bring spec quality back under their
control.

**Shape:** a per-decl annotation that replaces the godoc-derived
description on the annotated type's emitted definition. Several
phrasings worth weighing:

- **`// swagger:description <text>` on the user's own type that
  ALIASES the stdlib type.** E.g. `type Timestamp = time.Time` with
  `swagger:description "ISO 8601 timestamp"`. Works for the alias
  path (Q-C already inlines stdlib aliases, so the alias's
  description survives onto the canonical shape). Does NOT help
  when the user reaches `time.Time` via embed or field type
  directly — there's no user-controlled decl to attach the
  annotation to.
- **An `Options.DescriptionOverrides map[string]string` field.**
  Maps fully-qualified Go type names (`"time.Time"`,
  `"encoding/json.RawMessage"`) to replacement descriptions. Works
  for every usage path because the override is consulted at the
  emit site, not at the decl-discovery site. Coarser than the
  per-decl annotation but covers the stdlib-noise case
  comprehensively.
- **Both.** Per-decl annotation for ergonomics on user-controlled
  types; global option for stdlib + third-party types the user
  can't annotate.

**When to revisit:** when a user surfaces a real complaint about
spec noise (or when v2 reopens the Options surface).

**Companion finding** — Q30 in `observed-quirks.md` documents the
empirical observation (Time def with full godoc appearing across
field / embed / slice / map / pointer usages of `time.Time`). The
feature work here is the documented resolution path.

### 3.5 `DefaultAllOfForEmbeds` — opt-in allOf-by-default for embeds

**Origin:** W3 alias workshop Q-D close-out (2026-06-04).

**Scope:** today, the documented rule is "embeds always inline
properties, never `$ref` unless the embed is `swagger:allOf`-tagged"
(`internal/builders/schema/embedded.go`). The Q-D patch enforced this
rule for aliased embeds too (previously, aliased embeds were
silently promoted to allOf even without `swagger:allOf` — that's the
bug Q-D fixed).

Some users prefer the opposite default. Generating client code from
a spec where every embed produces allOf composition gives
downstream a clean inheritance hierarchy — each embedded type
becomes a base type the generator can reuse. The inline default
loses that structure: every embedding struct emits a flat copy of
the embedded fields, and the generator has no way to recover the
"this composes Y" relationship.

**Shape:** a new option `Options.DefaultAllOfForEmbeds bool`
(default `false` so existing users see no change). When `true`,
plain (non-`swagger:allOf`-tagged) embeds are emitted as allOf
members containing `$ref` to the embedded type's definition — the
shape the per-field `swagger:allOf` annotation produces today.
Plain non-embed fields are unaffected; this is purely about how
struct embeds render.

Interactions:
- `swagger:allOf` annotation continues to win (already in allOf
  shape; the option just makes that the default).
- Pointer embeds (`*Base`) take the same path — the pointer is
  peeled to its named target.
- Aliased embeds resolve to their unaliased type (the Q-D contract)
  and then enter the allOf path under this option.
- Embeds of stdlib-special types (`error`, `time.Time`) interact
  with `applyStdlibSpecials` first; the recognizer's canonical shape
  takes precedence over any composition shape.

**When to revisit:** first time a user requests a flatter spec
output from a deeply-nested embed graph, or as part of the v2
Options surface review.

### 3.7 `DiscoverAliasesAsTypes` — opt-in pre-R6 alias discovery

**Origin:** W3 alias workshop Q-E close-out (2026-06-10).

**Scope:** the schema builder's rule today is "an unannotated alias
is a Go implementation detail; it dissolves to its unaliased target
at every use site, and never produces its own `definitions` entry"
(R6, enforced in `internal/builders/schema/schema.go` `buildAlias`
via the `decl.HasModelAnnotation()` gate). The patch fixed Q-E —
unannotated aliases no longer manufacture dangling definitions, and
field-site `$ref`s point at the underlying target instead of the
alias name. `swagger:model` is the user's explicit opt-in for
"expose this alias as a first-class spec entity."

Some users prefer the opposite default. When a codebase pervasively
uses aliases as domain-modelling vocabulary (`type UserID =
int64`, `type Email = string`), they'd rather every discovered
alias surface as its own definition without having to annotate
each one individually. The pre-R6 behaviour did exactly that — the
discovery loop pulled each referenced alias into `ExtraModels` as a
side effect of the field-site `MakeRef`.

**Shape:** a new option `Options.DiscoverAliasesAsTypes bool`
(default `false` so existing users see the R6 behaviour). When
`true`, every alias reachable through the discovery walk gets a
`definitions` entry and field-site `$ref`s point at the alias name
rather than dissolving to the target. The patch reverses the R6
gate at the same `buildAlias` site:

  - `false` (default): annotation gates first-class status (R6)
  - `true`:  every discovered alias is first-class (pre-R6 behaviour)

Interactions:
- `TransparentAliases=true` always wins. Transparent dissolves
  aliases at every use site by definition, so this option is
  inert under Transparent (the dissolve happens before the R6
  gate is consulted).
- `RefAliases=true` × `DiscoverAliasesAsTypes=true` reproduces the
  pre-R6 RefAliases chain shape (alias decl chains to its target
  via `$ref`; alias-name `$ref`s appear at field sites).
- `swagger:model` annotation is still honoured; this option just
  removes the requirement to annotate every alias to get it into
  `definitions`.
- The Q-D embed contract is independent: `swagger:allOf` still
  governs composition shape at embed sites, regardless of whether
  the embedded alias is annotated or auto-discovered under this
  option.
- The `swagger:model` decl-side semantics are unaffected — an
  annotated alias decl carries its own `definitions` entry
  unconditionally (R2 in the workshop ledger), and this option
  does not alter that.

**Out of scope:** the parameters and responses builders have not
yet received the R6 treatment (Q7 / Q12 work, in flight as
fix-quirks Phase C1/C2). When that work lands, the same option
should govern those layers uniformly — the gate moves from
schema-builder-internal to a cross-layer concern. Until then this
option only affects the schema builder.

**When to revisit:** first time a user requests "I want every
alias to be a type" without per-decl annotation churn, or as part
of the v2 Options surface review.

### 3.6 Godoc-identifier prefix — scope review

**Origin:** P1.6 decision — only `swagger:route` accepts a leading
godoc identifier (`DoFoo swagger:route ...`), everything else must
start the comment line.

**Scope:** revisit whether `swagger:operation` (or future
annotations under an `openapi:` prefix per C9) should inherit the
same godoc-friendliness. v1 does NOT extend the exception; we
preserved that for parity. But from an author-experience angle,
the restriction is arguably surprising.

**When to revisit:** C9 (pluggable styles) work in v2.x, which
will also want to think about style-prefix choices.

---

## 4. Parser internals

### 4.1 Column precision under `/* */` continuation beyond ASCII

**Origin:** P1.1 preprocessor — current column math uses byte
offsets; works correctly for ASCII godoc but overstates columns
when a continuation line contains multi-byte runes before the
content start.

**Scope:** switch to rune-based column counting (or UTF-16 for
LSP interop) in `stripLine`. Small change, affects all
`Line.Pos.Column` values.

**When to revisit:** when LSP consumers actually show non-ASCII
comments — probably never for English docs, but an eventual
concern for i18n annotations or URL-embedded chars.

### 4.2 Bullet-list dash preservation — analyzer-side tolerance

**Origin:** P1.10 decision to keep leading `-` in Text (v2
divergence from v1's silent strip).

**Scope:** the Block's Title/Description now contains `- foo -
bar` literally. If any downstream renderer trims leading `-` for
cosmetic reasons (v1 did so silently), we keep parity by default
at the parser layer but may need an opt-in `StripBulletDashes`
option on the bridge-tagger. No known consumer yet.

**When to revisit:** first time a P5 builder migration turns up
fixture diffs involving bullet-list description formatting.

---

## 5. Test infrastructure

### 5.0 Parity suite removal at P6 cutover

**Origin:** P5.1 step 3 decision 2026-04-21 — `internal/integration/
parity_test.go` is tactical migration infrastructure. Once the
`UseGrammarParser` flag is removed at P6 and grammar-parser is the
only path, the suite has no dual path to compare against and
becomes pure CI burden.

**Scope:** delete the file; clean up any fixture entries in
`parityFixtures`; remove the helper shims. One commit, bundled
with P6.4 flag removal.

**When to revisit:** P6 cutover. This is an explicit migration-
design obligation, not a deferral. See
`grammar-parser-tasks.md` P6.4 and
`p5-builder-migrations.md` §5.3.

### 5.1 Parity harness → v1-side adapter

**Origin:** P4.1 shipped as v2-regression-only; the v1 comparison
side is deferred to P5.1 (the first bridge-tagger commit).

**Scope:** add a v1 adapter at `internal/parsers/grammar/grammar_test/`
that intercepts the legacy tagger writes and produces a
`NormalizedCommentView`. The harness then runs per-fixture:
v1-view vs v2-view, diffs, fails on mismatch that isn't explicitly
allowlisted.

**When to revisit:** first task of P5.1 — literally the first
thing the schema-builder migration needs.

### 5.2 Annotation surface-form documentation (P7.7)

**Origin:** Q4 design debate 2026-04-22 — during the multi-line-
block-body discussion it emerged that the expected contract was
not documented anywhere user-facing. Three participants had three
different mental models of what `consumes:`/`produces:`/etc.
bodies should accept.

**Scope:** when P7 hardening runs, write a surface-forms doc
covering:
- YAML-list contract for multi-line block bodies (strict; scalar
  emits diagnostic).
- Every keyword whose body follows the contract.
- YAML-list vs bare-form migration guidance.
- Enum surface forms (W2 §1 three variants).
- Cross-refs to the auto-generated keyword table.

**When to revisit:** P7.7 — post-cutover hardening. Tracked in
`grammar-parser-tasks.md` P7.7.

### 5.3 Property-based Block generator (P7.4 preview)

**Origin:** architecture §5.3 (go-openapi/testify pattern) / task
P7.4.

**Scope:** generator for synthetic `Block` values used in builder
tests. Produces thousands of plausible shapes, assertion runs over
the spec output.

**When to revisit:** P7 hardening — after cutover. Tracked as
P7.4 in the tasks plan; this entry exists so it's visible here
alongside its siblings.

---

## 6. Process

### 6.1 P5 per-builder catch-up tasks

**Origin:** the cross-cutting plan rule "no 'we'll check in P4
parity' deferrals — every phase ends with a catch-up task".

**Scope:** each P5.x bridge-tagger migration will surface its own
flags (e.g., schema-specific surface quirks, operation-specific
YAML oddities). Those flags either get resolved inside their P5.x
commit or a P5.x.catch commit before the migration is considered
done.

**When to revisit:** during P5 migration — this is a reminder, not
a deliverable.

---

## Change history

| Date       | Change |
|------------|--------|
| 2026-04-21 | Initial catalog (post-W2). Entries: W2 §4 forthcoming (1.1–1.3), W3 placeholder (2.1), goccy POC (3.1), splitCommaList quoting (3.2), godoc-ident scope (3.3), column unicode (4.1), bullet-dash tolerance (4.2), parity v1 adapter (5.1), property-based generator (5.2), P5 catch-up rule (6.1). |
| 2026-06-10 | Added §3.7 `DiscoverAliasesAsTypes` — opt-in toggle to disable the Q-E/R6 annotation gate and restore pre-R6 "every discovered alias is a type" behaviour. Inert under `TransparentAliases`; cross-layer extension (parameters/responses) gated on the C1/C2 work. |
