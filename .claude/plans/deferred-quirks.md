# Deferred quirks — holding area for v2 or later

Items that surfaced during the `observed-quirks.md` cleanup pass but were
deliberately set aside because the fix either (a) touches semantics that
need broader design input, (b) has cross-cutting ramifications the current
architecture makes hard to bound, or (c) the v2 grammar-parser rewrite
will obsolete the current code path anyway.

Each entry names the fix that was *attempted or considered*, what made it
risky, and the fixture that demonstrates the behavior so future work can
pick up from a failing test.

---

## D1 — Alias-expand parameters drop the body/query semantic (was Q7)

**Status:** Not attempted. Logged during the Q-pass as known-hard.

**Symptom.** When `swagger:parameters` is declared on a top-level alias
(`type P = Q`) in the default (non-`RefAliases`, non-`TransparentAliases`)
mode, the expand path builds `Q` as a plain schema rather than dispatching
through the parameter builder. Consequence:

- `data` field loses `in: body` semantic and becomes a property, not a body parameter.
- `search` field loses `in: query` semantic.
- Both fields leak the directive text into their `description` (pre-Q5 baseline; Q5 now strips it at the schema level).

**Fixture.** `fixtures/enhancements/alias-expand/api.go` —
`type AliasedTopParams = exportedParams`. Golden:
`fixtures/integration/golden/enhancements_alias_expand.json`.

**Why deferred.** The alias-expand dispatch for annotated-parameter types
needs to route through the parameter builder for the underlying type, not
the schema builder. That crosses the parameter/schema boundary and
interacts with `RefAliases`, `TransparentAliases`, and the two expand
paths in `buildDeclAlias`. Not a localized fix.

**Pointer to compare.** Same fixture under `RefAliases=true` —
`enhancements_alias_ref.json` — produces correct body/query parameters.

---

## D2 — Embed-as-allOf vs. embed-as-inline asymmetry (was Q8)

**Status:** Not attempted. Flagged as "audit before changing" in the
original plan.

**Symptom.** A struct that embeds a plain named struct `Base` emits
`allOf: [$ref: Base, {inline object}]`. A struct that embeds a named
*interface* inlines the interface's methods as properties of the outer
schema directly (no `$ref`, no `allOf`).

**Fixture.** `fixtures/integration/golden/enhancements_embedded_types.json`:
- `EmbedsAlias` → `allOf` shape.
- `EmbedsNamedInterface` → flat properties, no allOf.

**Why deferred.** The asymmetry appears intentional (interfaces are
method-set shapes in Swagger's view), but reconciling it would require
deciding whether embedded-interface should also be an allOf member. That's
a design call — not a bug fix. v2's schema model (OAI 3.x) can answer it
more naturally than Swagger 2.0's toolkit.

---

## D3 — `swagger:strfmt` + `swagger:model` named-strfmt inconsistency

**Status:** Attempted during the Q-pass as a follow-up to Q10 ("Option 1"
in the Q10 discussion). Reverted before merge.

**Symptom.** When a type carries both annotations:

```go
// swagger:strfmt phone
// swagger:model
type PhoneNumber struct {
    CountryCode string
    Number      string
}
```

and a field references it (`Contact.Phone PhoneNumber`), the scanner
produces:

- **Field site:** `{type: "string", format: "phone"}` — strfmt wins.
- **Top-level definition:** `{type: "object", properties: {CountryCode, Number}}` — the struct walk wins; the strfmt annotation is ignored at decl time.

So the author asked for "named strfmt" (reusable `PhoneNumber` definition
rendered as a formatted string) but gets an inconsistent pair: the field
says string, the definition says object.

**Fixture.** `fixtures/enhancements/named-struct-tags-ref/types.go` —
`PhoneNumber` with both annotations, used by `Contact.Phone`. Golden:
`enhancements_named_struct_tags-ref.json`.

**What was tried.** The Option 1 fix:
1. `buildDeclNamed`: detect `swagger:strfmt` on the decl and emit
   `{string, fmt}` instead of walking the struct body.
2. `buildNamedStruct`: when the target also has `swagger:model`, emit
   `$ref` instead of inlining the strfmt.

**Why reverted.** Ramifications were larger than expected:

- Pre-existing fixtures in `fixtures/goparsing/classification/transitive/mods/aliases.go`
  use the same `swagger:strfmt + swagger:model` combination on
  defined-from-`time.Time` types (e.g. `SomeTimeType time.Time`). Tests
  there — `TestAliasedTypes` line 733, `TestAliasedModels` — assert the
  *inline* baseline (`scantest.AssertProperty(..., "string", ...)`) rather
  than a `$ref`. Option 1 flips these to `$ref`, requiring coordinated
  test updates.
- The decl-level `StrfmtName` check also over-fires on slice/array/map
  underlyings: `type SomeTimesType []time.Time` with `swagger:strfmt date-time`
  should emit `{array, items: {string, date-time}}`, not flatten to
  `{string}`. A correct fix would gate the check on struct underlying,
  then symmetrically consider whether `buildNamedSlice` /
  `buildNamedArray` / `buildNamedMap` should also route through `$ref`
  under the `swagger:model` combination.

The surface area is wider than the Option 1 code change suggested, and
the existing test coverage of the combination is entangled with the
inconsistency itself.

**Why deferred.** The combination is niche, the footgun is narrow (you
get what you asked for on one side of the indirection, not both), and
v2's annotation redesign can reshape the contract without carrying this
legacy. A focused decision on "named strfmt" semantics belongs in the v2
design, not a bug-fix pass.

**Trace it leaves.** The `named-struct-tags-ref` fixture and its golden
are checked in as a deliberate marker. The golden captures the observable
inconsistency (inline field + struct-body definition) so future work on
this decision has a failing test to anchor against.

---

## D4 — Expand-mode alias declarations are verbose (was Q11)

**Status:** Reviewed during the Q-pass and classified as working as designed.

**Symptom.** With the default scanner options (non-`RefAliases`,
non-`TransparentAliases`), each alias layer in a chain like
`PayloadAlias2 = PayloadAlias = Payload` produces its own top-level
definition with a full copy of `Payload`'s struct body. A 2-level chain
yields three identical bodies.

**Fixture.** `fixtures/enhancements/alias-expand/api.go` — `PayloadAlias`,
`PayloadAlias2`. Golden:
`fixtures/integration/golden/enhancements_alias_expand.json`.

**Why deferred.** Not a bug; it is the defining feature of expand mode.
Any "fix" collapses into one of the two existing modes:

- Emit `$ref` chains at top level → identical to `RefAliases=true`.
- Elide alias decls entirely, emit only terminals → identical to `TransparentAliases=true`.

A conditional elision (e.g. "drop aliases that have no doc comment")
would be an arbitrary rule that does not reflect author intent. Field
sites already emit `$ref` under expand mode, so the verbosity is limited
to the top-level decls the user explicitly named.

The broader honest framing: the alias-handling surface is a recently
added feature, still fundamentally incomplete. The three modes cover a
reasonable space of user intents, but they expose the incompleteness —
the default mode has a visible cost the other two modes avoid, each with
their own trade-offs. A v2 design can revisit whether the scanner can
infer alias intent without requiring a user flag at all.

---

## D5 — Unexported parameter-alias backing struct leaks as definition (was Q12)

**Status:** Not attempted. Classified as alias-theme during the Q-pass.

**Symptom.** When `swagger:parameters` is declared on an exported alias
whose target is an unexported struct (`type AliasedTopParams = exportedParams`),
the unexported backing struct (`exportedParams`) ends up in the spec's
`definitions` under its lowercase Go identifier.

**Fixture.** `fixtures/enhancements/alias-expand/api.go` — `exportedParams` +
`AliasedTopParams`. Golden:
`fixtures/integration/golden/enhancements_alias_expand.json` shows
`"exportedParams": { ... }` alongside the exported aliases.

**Origin.** `internal/builders/parameters/parameters.go:120, 279` register
the alias's backing decl in `postDecls`, which later surfaces as a
top-level definition.

**Why deferred.** The fix has to answer: when the alias target is
Go-unexported, should the spec (a) elide the backing struct entirely,
(b) rename it after the exported alias, or (c) emit both but rename the
backing one? Each option entangles with how parameter-alias expansion
should dispatch (see D1). Part of the alias-handling cluster.

---

## D6 — `type X = any` emits a bare stub (was Q13)

**Status:** Not attempted. Classified as alias-theme during the Q-pass.

**Symptom.** An alias to the predeclared `any` (e.g. `type Wildcard = any`)
under `RefAliases=true` short-circuits to `_ = tgt.Schema()` in
`buildDeclAlias`, creating an empty schema. The resulting top-level
definition carries only `x-go-package` — no `type`, no `format`. Fields
referencing it become `$ref`-only with no schema information.

**Fixture.** `fixtures/enhancements/ref-alias-chain/types.go` — `type Wildcard = any`
and `Envelope.meta Wildcard`. Golden:
`fixtures/integration/golden/enhancements_ref_alias_chain.json`:

```json
"Wildcard": { "x-go-package": "..." }   // no type
"Envelope.meta": { "$ref": "#/definitions/Wildcard" }
```

**Why deferred.** The fix candidates — emit `{}` (Swagger 2.0 has no
equivalent of JSON Schema `true`), emit `{type: object}` as a catch-all,
or elide the definition entirely — each encode a different design
decision about aliases to predeclared-primitives. Same design question
surfaces in D4's mode discussion. Part of the alias-handling cluster.

---

## Aliasing as a theme

D1, D3, D4, D5, and D6 are all surfaces of the same underlying issue:
alias handling was added relatively late and the three modes (default /
`RefAliases` / `TransparentAliases`) don't fully carve up the decision
space. Each defers cleanly to v2 rather than accumulating rules on the
current two-mode dispatch in `buildDeclAlias` / `buildAlias`.

The Q-pass patches D1, D5, D6 would each require the same kind of
design decision ("what does an alias to X mean in the spec?") being
re-answered per case. v2 is the right place to answer once.

---

## When to pick these up

Either of the following makes D1–D3 actionable:

1. v2 grammar parser lands and redefines the annotation contract — these
   quirks dissolve or reappear in a form specific to the new model.
2. A user surfaces a concrete, non-hypothetical need that the current
   behavior blocks — the discussion above gives a starting diagnosis and
   a fixture to reproduce against.

Until then: leave alone. The tests that currently pin the baseline do
their job as change-detectors.
