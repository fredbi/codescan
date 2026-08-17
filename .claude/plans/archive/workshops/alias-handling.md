# W3 — Alias-handling design

Date: 2026-06-03
Status: ⬜ **draft, awaiting discussion**
Companion: `.claude/plans/fix-quirks.md` C0 (workshop gate);
`.claude/plans/observed-quirks.md` Q7, Q8, Q11, Q12, Q13.
Branch context: `fix/quirks` after A1+A2+A3+B1+B2 (commits
`b3257a9` … `75b077f`).

This document is **workshop input**, not a decision record. It exists
to anchor the conversation: state the problem, list what we already
know empirically, and pose the questions whose answers will determine
the model. The decision sections at the end are deliberately empty
until the discussion fills them.

---

## 1. Scope and framing

### 1.1 The six quirks in scope

| Q | Symptom (one line) | Surface |
|---|---|---|
| **Q3** | `type Timestamp = time.Time` annotated `swagger:model` emits `{type: object}` in default Expand mode — date-time format silently lost. RefAliases mode improves to a 2-hop chain reaching `Time`. TransparentAliases inlines. | schema buildDeclAlias |
| **Q7** | `swagger:parameters` on a top-level alias in default mode emits the alias as a `definitions` object instead of merging into the operation's parameters. `data` loses `in: body`, `search` loses `in: query`. | parameters builder |
| **Q8** | `struct { BaseAlias; Extra }` emits `allOf: [$ref:BaseAlias, {Extra inline}]`, but `struct { Base; Extra }` emits a flat `{id, name, extra}`. Two literally indistinguishable Go types produce different schema shapes. | schema embed dispatch |
| **Q11** | 3-link alias chain `A = B = Target` produces three definitions (refs-not-copies since M-stream, but the count is still three). | schema alias-expand |
| **Q12** | A lowercase-named struct (`exportedParams`) declared only as the RHS of an exported alias still ends up in `definitions`. Unexported leak via alias. | parameters builder (Q7-coupled) |
| **Q13** | `type X = any` emits `{title, x-go-package}` only — no `type`, no `format`. Spec-edge but blank. | schema buildDeclAlias |

**Q3/Q13 family.** Same root cause — `buildDeclAlias` Expand-branch
(`schema.go:168-170`) falls through to `buildFromType(Underlying)`
without consulting `applyStdlibSpecials` first. For `time.Time` the
Underlying is the unexported-fields stdlib struct (→ `{type: object}`);
for `any` the Underlying produces nothing typeful. The patch is
the same one-line check in both cases; the *semantic* answer ("what
should Expand mean for a stdlib-special RHS?") is what §3.5 has to
answer.

### 1.2 Fred's framing (2026-06-03)

> The alias model landed cleanly for the schema builder during Stream M
> but not for the allOf path, the parameters builder, or the responses
> builder. Essentially an **unfinished job**.

The workshop must answer whether to **propagate** the schema-builder
model to the other three sites, or **design a different model** that
fits all four uniformly.

### 1.3 What B1 + B2 surfaced

Two empirical findings from the fix-quirks B-phase that change the
shape of this workshop:

- **B1 (Q9 close-out):** the principled struct-vs-interface asymmetry
  rationale is documented in `schema/README.md` §method-mangler.
  *Names* have a defensible asymmetry (struct mirrors `encoding/json`,
  interface invents). The workshop should NOT relitigate naming.
- **B2 (Q8 re-classification):** the Q8 author thought the asymmetry
  was struct-embed-vs-interface-embed, but the actual dispatch
  (`embedded.go:40-53`) inlines for both named-struct AND
  named-interface embeds. The outlier is the **alias path**, which is
  the ONLY embed form producing `allOf` + `$ref`. So Q8 is not an
  interface question at all — it's an alias question.

---

## 2. Empirical landscape — what the code does today

### 2.1 The three modes

Three `Options` flags govern alias treatment:

| Flag | Default | Effect (schema builder) |
|---|---|---|
| `TransparentAliases` | `false` | **Dissolve.** No definition for the alias LHS. Build from the RHS directly into the field/element/decl position. |
| `RefAliases` | `false` | **$ref.** LHS gets a definition that is a `$ref` to the RHS's target. |
| (neither) | the implicit default | **Expand.** LHS gets a structural definition mirroring the underlying. |
| `DescWithRef` | `false` | Tangential. Controls whether description text accompanies a `$ref` (the JSON Reference spec says siblings of `$ref` are ignored; this flag bypasses that). Not in alias scope per se. |

Selection logic (`schema/schema.go:155-170`):

```go
// buildDeclAlias for a top-level *types.Alias decl:
if Ctx.TransparentAliases() { return buildFromType(rhs, target) }   // Dissolve
AddDiscoveredModel(s.Decl); AppendPostDecl(s.Decl)                  // Register LHS
if !Ctx.RefAliases()       { return buildFromType(Underlying, target) } // Expand
// else fall through to $ref
```

So **Dissolve** wins over **Expand** wins over **$ref** in the
order-of-checks. The empirical default (neither flag set) is
**Expand**.

### 2.2 Schema builder — three call sites

The schema builder has three places it can see an alias. Each runs a
slightly different model.

**Call site A — `buildDeclAlias` (top-level alias decl).** Full
three-mode dispatch as above (`schema.go:149-205`). On the `$ref`
branch, peels `rhs.(type)`:
- `*types.Named`: ref to RHS's named target.
- `*types.Alias`: ref to the *next* alias's LHS — Q11's "chain"
  shape.
- default: `buildFromType(rhs, target)`.

**Call site B — `buildAlias` (alias reached as field/element type).**
Two-mode dispatch (`schema.go:264-285`): Transparent → dissolve;
else → $ref to RHS. **No Expand branch** here — fields always either
dissolve or $ref. (Q13 lands here for `type X = any`.)

**Call site C — `buildEmbedded` (alias reached as embed type).**
Always calls `buildAlias` (`embedded.go:46-48`). So embed of
`*types.Alias` flows through call site B's two-mode dispatch, which
produces a `$ref` (default) or dissolve. The outer schema then
absorbs it as one allOf member.

**Compare with non-alias embed:** `buildNamedEmbedded`
(`embedded.go:66-105`) for `*types.Named` (struct or interface) ALWAYS
inlines (`buildFromStruct` / `buildFromInterface` write directly
into the outer `schema`, never produce an allOf member). The modes are
not consulted.

So **Q8's asymmetry is here**: the alias embed path honors the modes
and emits `$ref` by default; the named-direct embed path ignores the
modes and always inlines.

### 2.3 Parameters builder — `parameters/parameters.go:102` `buildAlias`

Two-mode dispatch:
- Transparent → dissolve to RHS.
- else → `AppendPostDecl(LHS)`, then walk the RHS (Named / Alias /
  default).

**No `RefAliases` branch in this function.** The flag is consulted in
a *different* helper, `processFieldOnAlias` at line 289, for body
parameters with full schema treatment. That helper does:

```go
if typable.In() != inBody || !Ctx.RefAliases() {
    // expand: unaliased := types.Unalias(tpe); buildFromField(...)
}
// otherwise: $ref the alias to RHS
```

So in the parameters builder, the `$ref` mode is gated on
`in: body` AND `RefAliases=true`. For non-body parameters
(query/path/header/formData), `RefAliases` is **never honored** —
they always expand.

**Q7 lands here.** When `swagger:parameters` is on a top-level alias
*itself* (not a field of one), the dispatch routes through a path
that builds the alias as a "plain schema" rather than as a parameter
envelope. The fix-quirks plan calls out the path crossing the
parameter/schema boundary as the structural reason.

**Q12 is coupled to Q7.** If Q7's fix routes alias parameters through
the parameter builder, the unexported RHS target may stop being
reachable as a top-level schema definition, and the unexported leak
collapses to zero work.

### 2.4 Responses builder — `responses/responses.go:215`

Same three-mode skeleton as schema (`TransparentAliases` checked
first), plus a `RefAliases`-gated path at line 225 mirroring the
parameters body-vs-non-body distinction at line 295/306.

The shape is structurally parallel to the parameters builder. Same
"in: body + RefAliases → $ref; everything else expands" rule.

### 2.5 Discovery layer — `AddDiscoveredModel` / `AppendPostDecl`

The spec orchestrator's discovery loop picks up any decl that lands
in `ExtraModels` (via `AddDiscoveredModel`) or in any builder's
`PostDeclarations` queue (via `AppendPostDecl`). Both eventually
produce `definitions` entries.

The schema builder's `buildDeclAlias` registers the LHS decl on every
mode except Transparent. The parameters builder's `buildAlias`
registers the LHS decl on every mode except Transparent. The responses
builder's mirrors that.

This is the **plumbing** by which Q12's unexported-leak happens: the
parameters builder marks the unexported RHS struct as discovered, the
orchestrator visits it during its post-walk, and a definition pops out
even though the user only ever exposed an alias.

### 2.6 The dispatch table for embed (the Q8 surface)

Empirical 5-row probe from B2 (`fix-quirks.md` B2 commit notes):

| Embed shape | Go type token | Code path | Result |
|---|---|---|---|
| `Base` (named struct, direct) | `*types.Named` | `buildNamedEmbedded` | **FLAT** (inline) |
| `BaseAlias` (alias of struct) | `*types.Alias` | `buildAlias` → `MakeRef` | **allOf** with `$ref` |
| `*Base` (pointer to named struct) | `*types.Pointer` → recurses | `buildNamedEmbedded` | **FLAT** (inline) |
| `Iface` (named interface) | `*types.Named` (interface underlying) | `buildNamedEmbedded` → `buildFromInterface` | **FLAT** (inline) |
| `interface{Audit() string}` (anonymous, struct-embed) | (Go-illegal) | n/a | discard |

`BaseAlias = Base` is a transparent rename — the two are literally
the same `*types.Named` in the Go type system — but produce different
spec output.

### 2.7 Per-quirk current behaviour summary

| Q | Fixture / golden | Pre-workshop state |
|---|---|---|
| Q3 | `fixtures/enhancements/ref-alias-chain/` → `enhancements_ref_alias_chain.json` (RefAliases=true only) | Default mode: `Timestamp = time.Time` model emits `{type: object}` — format LOST. RefAliases: 2-hop chain reaches `Time`. TransparentAliases: inline. The existing golden only covers RefAliases — no default-mode witness for the bug. |
| Q7 | `fixtures/enhancements/alias-expand/api.go` → `enhancements_alias_expand.json` | `AliasedTopParams` lands in `definitions` (wrong); compare to `enhancements_alias_ref.json` (`RefAliases=true`) which routes correctly. |
| Q8 | `fixtures/enhancements/embedded-types/types.go` → `enhancements_embedded_types.json` | `EmbedsAlias` has allOf, `EmbedsNamedInterface` has flat. Missing: `EmbedsBase` (direct-named-struct) case. |
| Q11 | `fixtures/enhancements/alias-expand/` (shared with Q7) | 3-link chain produces 3 definitions. M-stream made them ref-chains (byte cost lower) but the count is unchanged. |
| Q12 | `enhancements_alias_expand.json` | `exportedParams` (lowercase) present in `definitions`. |
| Q13 | `enhancements_ref_alias_chain.json` | `Wildcard` definition is `{title, x-go-package}` only. |

---

## 3. The questions the workshop must answer

Seven questions. Each is structured the same way:

- **Question.** What the workshop is being asked.
- **Options.** Concrete candidate answers, labelled A/B/C/…
- **Tells.** What shifts the answer one way or another — questions of
  fact you can probe, or value judgments you can articulate.
- **Couplings.** Which other §3.X questions move with this one.
- **Empirical check.** Existing fixture / golden / source pointer
  whose state should be inspected before deciding.

§5 collects the decisions in prose. §6 turns each decision into a
per-quirk action item.

### 3.1 What is the **intent** of `type X = Y` per target?

This is the most abstract framing. The answer to every other question
is partly derived from this.

#### Question

For each of the six contexts an alias can appear in, what does the
*user typically mean* when writing it? Spec output should reflect
that intent.

#### Options

| Context | Option A | Option B | Option C |
|---|---|---|---|
| `swagger:model X = Y` | X is a *named alias* of Y. Emit X as its own definition (ref or expand) — both X and Y visible. | X is a *rename* of Y for clarity. Collapse to one definition under the more canonical name. | X is a *synonym*. Don't even mention X in the spec — only Y. |
| `swagger:parameters X = Y` | X is the parameter-set name; Y is incidental backing. | X is a synonym; walk Y as the parameter set. | Refuse — parameter sets shouldn't be aliasable. |
| `swagger:response X = Y` | Same as parameters: X is the response-name; Y backs it. | Synonym; walk Y. | Refuse. |
| Field type `F X` (X is an alias) | Field is typed by alias name. Field schema → $ref to X. | Field is typed by what X points to. Field schema → walk Y. | Mixed — depends on whether X is annotated `swagger:model`. |
| Embed `struct { X; … }` | Embed of X = embed of Y (same shape). | Embed of X retains alias identity (different shape). | Mixed — depends on annotation. |
| Unannotated, reached transitively | Discover X, emit a definition for it. | Dereference; no definition for X. | Discover only if also reachable via a field, not via embed. |

> **FRED**
> Aliased types are syntactic sugar in go and are undistinguable as types.
>
> In a schema (spec), they express an idea similar to a "$ref".
>
> For example, the design I chose some time ago to generate a json schema like:
>
> myString:
>   $ref: otherstring
>
> otherString:
>   type: string
>
> was to capture this as a type alias in generated go:
>
>
> type MyString = OtherString
> type OtherString string
>
> Scanned code is of course not generated in general, but I think the idea is to create an indirection.
>
> Since there is no "one-size-fits-all" approach, there is an option to expand the type as a standalone synonym like so:
> (the option is only global now IIRC).
>
> myString:
>   type: string # just duplicates (expand)
>
> otherString:
>   type: string
>
> This stands for models of course.
>
> I believe that an aliased type tagged operation or routes would just duplicate (and produce something weird I assume).
>
> For parameters and responses, I think the $ref should work too (if I am correct we only support operation-level params,
> not spec level shared params / responses - known shortcoming -).
>
> We should probably reason with concrete examples.


#### Tells

- **Does `encoding/json` treat X and Y identically?** Yes — they are
  the same type at runtime. Argues for B everywhere (rename for
  clarity, one definition).
- **Why did the user bother to write the alias?** If for clarity of
  intent (`type UserID = int64`), B; if to introduce a named entity
  the spec should expose (`type ApiKey = string` with
  `swagger:strfmt`), A.
- **Did the user annotate the alias with `swagger:model`?** Strong
  signal of intent A (they want it in the spec). No annotation → B
  or C is defensible.
- **Does Y already have a definition?** If yes and X is annotated,
  A is the natural answer ($ref). If Y has no definition, C may be
  cleaner than emitting both.

> **FRED**
> Yes all 4 are true.

#### Couplings

- §3.2 — the mode factoring exists to surface these intents as
  user-controllable knobs (A ↔ `$ref`/`Expand`, B ↔ no separate
  decision, C ↔ `Transparent`).
- §3.3 — if intent is B (rename), §3.3 must be "identical schemas."
- §3.7 — option B for unannotated transitive case is a leak.

#### Empirical check

Inspect `fixtures/enhancements/ref-alias-chain/types.go` and the
`alias-expand` fixture: are the existing `swagger:model` annotations
on the alias decls a deliberate signal of intent A, or just routine
boilerplate added by the fixture author to make sure the alias gets
discovered? The answer informs whether annotation is the right
trigger.

---

### 3.2 Are the three modes the right factoring?

#### Question

`Options.TransparentAliases` × `Options.RefAliases` produces three
effective modes (Dissolve / Expand / $ref). Is this the right
factoring of user-controllable behaviour?

> **FRED**
> I don't know. We should reason on code examples and side them with spec.
> This is just too hard to reason in plain english on this kind of thing (at least for now, and for me).

#### Options

- **A. Keep all three modes as today.** Fix the bugs inside each; do
  not change the user-facing surface.
- **B. Collapse to two.** Drop Expand (the bug magnet) and require
  users who want it to opt in explicitly. Default becomes either
  Transparent or $ref.
- **C. Single enum option.** Replace the two booleans with
  `Options.AliasMode = Transparent | Expand | Ref` (default
  TBD). Same modes, cleaner API.
- **D. Pick a canonical mode** and demote the others to legacy
  back-compat. New code should use the canonical mode; the others
  emit a deprecation note.

#### Tells

- **Which mode does the M-stream-resolved schema-builder model
  prefer?** Per the empirical landscape (§2.2), schema builder
  treats `RefAliases=true` as the well-behaved path; other modes
  have edge-case bugs. Argues for D with Ref canonical.
- **Which mode best matches Go semantics?** Transparent — the alias
  is invisible at the type level. Argues for D with Transparent
  canonical.
- **Which mode is most common in user code?** Default Expand (no
  flags set). Argues for A (don't break existing users) or D with
  Expand canonical.
- **Is `Expand` salvageable?** It needs a `applyStdlibSpecials`
  carve-out (§3.5) to fix Q3/Q13. With the carve-out, it becomes
  defensible; without, it's a bug magnet.
- **Are the two booleans semantically independent or coupled?**
  Today's code treats them as mutually exclusive
  (`TransparentAliases` short-circuits `RefAliases`). That's a
  smell — boolean independence is a lie. Argues for C.

#### Couplings

- §3.1 — modes are knobs surfacing the intents from §3.1. If §3.1
  picks one intent as obviously dominant, modes collapse.
- §3.3 — if alias and named must be identical (§3.3 = A), the modes
  are about how the *combined* alias-or-named decl renders, not
  about the alias vs named distinction.
- §3.4 — uniform-vs-context-gated `RefAliases` interacts with
  whether `RefAliases` survives as a separate flag at all.

#### Empirical check

`grep -rn "TransparentAliases\|RefAliases" internal/builders/` for
the full consumer surface (already partially done in §2.1-§2.4 of
this doc). Verify no consumer relies on the booleans being
independently set (e.g. `TransparentAliases=true && RefAliases=true`
— what does that mean today?).

---

### 3.3 Should `*types.Alias` and `*types.Named` produce identical schemas?

The Q8 core question, but with consequences across all the other
quirks.

#### Question

In Go, `type BaseAlias = Base` is a transparent rename — `Base` and
`BaseAlias` are literally the same type (`types.Unalias(BaseAlias) ==
Base`; reflection cannot distinguish them). Should the spec output
also be indistinguishable?

Specifically at the embed site:
- `struct { Base; Extra }` emits `{type:object, properties:{id,name,extra}}` today (flat inline).
- `struct { BaseAlias; Extra }` emits `allOf: [{$ref:BaseAlias}, {properties:{extra}}]` today.

#### Options

- **A. Identical schemas everywhere.** `BaseAlias = Base` produces
  identical spec output as `Base` in every context.
  - **A1.** Make the alias path inline (match named — drop the
    allOf+$ref shape).
  - **A2.** Make the named path emit allOf+$ref (match alias —
    introduce a Base definition + reference it).
- **B. Different schemas — principled.** Articulate a structural
  reason a transparent Go rename should change the spec output.
- **C. Annotation-gated.** Identical when neither is annotated;
  different when the alias carries `swagger:model` (the annotation
  is the signal of intent A from §3.1).

#### Tells

- **Go type-system fact:** `types.Unalias(BaseAlias) == Base`. The
  type system cannot tell them apart. Argues hard for A.
- **Q9 precedent.** When B1 closed Q9, we accepted the
  struct-vs-interface asymmetry because the two have *genuinely
  different runtime serializations* (`encoding/json` mirrors struct
  fields; can't marshal interface methods). Aliases have NO
  different runtime serialization — `encoding/json` sees one type.
  The Q9 precedent argues against B unless a similar structural
  reason exists.
- **What does code generation downstream do with allOf?** Generators
  produce a base-type / extension-type relationship. The user who
  wrote `type BaseAlias = Base` was probably NOT expressing a
  composition intent — but A1 (inline) makes that go away.
- **What about explicit allOf?** The `swagger:allOf` annotation
  exists for explicit composition. If implicit allOf-from-alias goes
  away (A1), users who want allOf explicitly can still use the
  annotation. Argues for A1.
- **Migration cost of A vs B.** A1 (inline alias) breaks fixtures
  using the alias-embed shape; A2 (refify named) breaks fixtures
  using direct-embed (more common). Argues A1.

#### Couplings

- §3.1 — if the intent of embed-of-X is "embed of underlying"
  (option B everywhere in §3.1), §3.3 = A.
- §3.6 — chain length follows directly. If alias ≡ named, a chain
  collapses to one definition.
- §3.5 — Q3/Q13 stop being alias-specific bugs if alias-as-decl
  collapses to "emit the underlying's natural representation."
- §3.7 — Q12's unexported leak depends on whether the alias decl
  creates its own definition or just defers to the underlying.

#### Empirical check

The 5-row probe from B2 (§2.6) — confirm that no existing fixture
golden depends on the `EmbedsAlias` allOf shape semantically (a
generated client code expecting BaseAlias as a base type). Likely
the goldens are descriptive of current behaviour, not normative —
but worth confirming before A1 lands.

---

### 3.4 Should `RefAliases` apply uniformly, or remain context-gated?

#### Question

Today the `RefAliases` flag has three different scopes:
- Schema: controls Expand-vs-$ref for top-level alias decls.
- Parameters: only applies to `in: body` parameters; non-body always
  expands.
- Responses: same body-only gating.

Is this a UX heuristic, a technical constraint, or accidental drift?

#### Options

- **A. Uniform — drop the gate.** `RefAliases` applies to every
  context where a `$ref` is syntactically valid.
- **B. Keep the body-only gate as a documented contract.** Articulate
  the rationale (it does exist — see Tells).
- **C. Drop `RefAliases` entirely.** Replace with §3.2's enum mode;
  `Ref` mode applies wherever it's syntactically valid; otherwise
  reverts to Expand.

#### Tells

- **Technical fact:** non-body parameters in OAS 2.0 use SimpleSchema,
  which does NOT support `$ref`. So `RefAliases` on a non-body
  parameter cannot produce a `$ref` shape — the body-only gate is
  technically forced, not a UX choice. Argues for B (or for C with
  "applies where syntactically valid").
- **Response headers** are also SimpleSchema. Same constraint.
- **Is the gate documented?** Today, no — it's an implementation
  detail in `parameters.go:306` and `responses.go:308`. If we keep
  the gate, document it.

#### Couplings

- §3.2 — if §3.2 picks C (enum mode), the gate becomes "Ref mode
  applies wherever $ref is valid."

#### Empirical check

`grep -rn "in == \"body\"\\|InBody" internal/builders/parameters/
internal/builders/responses/` — confirm the only place "body" gates
ref-behaviour is around `RefAliases`, not other features.

---

### 3.5 What does Expand mode mean for a stdlib-special RHS? (Q3, Q13)

Same root cause for both — `buildDeclAlias` Expand branch
(`schema.go:168-170`) walks `Underlying()` without first checking
`applyStdlibSpecials`. The behaviour on Q3's `time.Time` and Q13's
`any` is different but produces the same shape of bug.

#### Question

Q3 — `type Timestamp = time.Time` annotated `swagger:model`, default
mode, today emits:

```json
{"type": "object", "title": "...", "x-go-package": "..."}
```

The `format: date-time` known for `time.Time` is lost. Inline-field
references to the same `time.Time` still produce `{type: string,
format: date-time}` correctly — only the alias decl loses it.

Q13 — `type X = any` annotated `swagger:model`, default mode,
emits:

```json
{"title": "...", "x-go-package": "..."}
```

Bare; no `type`, no `format`.

#### Options

**Q3-specific:**

- **A. Inline `{type: string, format: date-time}` on the alias
  definition** (matches TransparentAliases mode; matches the
  inline-field path).
- **B. Take the RefAliases 2-hop chain shape** (`Timestamp → Time`)
  even when `RefAliases=false`. Contradicts what the Expand label
  promises.
- **C. Refuse with a diagnostic.** "Expand mode cannot expand a
  stdlib-special type; use TransparentAliases or RefAliases."

**Q13-specific:**

- **A. Omit the definition.** No ref source; field carrying the
  alias still emits something useful.
- **B. Emit `{type: object}`.**
- **C. Emit `{type: object, additionalProperties: true}`** — the
  OAS-2 idiom for "anything."

#### Tells

- **Should Q3 and Q13 land the same fix?** Same root cause in code,
  but the semantic answers differ — `time.Time` HAS a canonical
  spec shape (`{type:string, format:date-time}`), `any` does NOT.
  Argues for "same code patch, different output."
- **One-line patch shape:** `if applyStdlibSpecials(...) { return
  nil }` before the Expand fallthrough. `applyStdlibSpecials`
  already knows what each stdlib-special should emit. The decision
  is whether the patch's *semantic* matches what the user wants.
- **Does Expand mode have a coherent contract today?** "Walk
  Underlying" works for plain user structs. It DOESN'T work for
  types with structural stdlib-specials. So either Expand's
  contract is wrong, or it's right but incomplete.

#### Couplings

- §3.2 — if §3.2 drops Expand entirely, Q3/Q13 dissolve.
- §3.1 — if intent of `swagger:model X = Y` for stdlib Y is
  "expose Y's canonical spec form under name X" (option A from
  §3.1), then Q3 = A, Q13 = C.

#### Empirical check

`applyStdlibSpecials` — read its current implementation. What does
it recognize? `time.Time`, `error`, anything else? Does it have a
notion of "this stdlib type expands to X spec shape"?
`grep -rn "applyStdlibSpecials\|isStdTime\|isStdError\|isAny"
internal/builders/schema/`.

---

### 3.6 How many definitions for an N-link alias chain? (Q11)

#### Question

`type A = B; type B = Target` — should the spec contain:

- **A.** One definition (the terminal `Target`).
- **B.** N refs to one (current: `A → B → Target` chain of `$ref`s,
  three definitions).
- **C.** N copies (legacy v1 pre-Stream-M behaviour).

#### Tells

- **§3.3 answer dominates.** If alias ≡ named, the chain has no
  identity — collapse to A.
- **Does the user write `swagger:model` on each link?** Today the
  fixture does. That's a strong signal of "I want each in the
  spec" — argues for B (preserve names as refs).
- **What does v1 do?** C (copy). Stream M's middle ground (B) was a
  byte-cost optimisation, not a semantic decision.

#### Couplings

- §3.3 — A here requires §3.3 = identical-schemas.
- §3.1 — if intent of `swagger:model X = Y` is "expose X as a named
  entity" (§3.1 option A), then B is the right answer regardless of
  §3.3.

#### Empirical check

`fixtures/enhancements/ref-alias-chain/` and
`fixtures/enhancements/alias-expand/` — confirm the chain shapes
they exercise are deliberate test of behaviour B, not just
incidental.

---

### 3.7 Unexported-leak (Q12) — emergent or designed?

#### Question

A lowercase-named struct (`exportedParams`) declared only as the RHS
of an exported alias still ends up in `definitions`. Is this a
feature (the type is reachable, so we expose it) or a bug (the user
chose lowercase precisely to hide it)?

#### Options

- **A. Filter unexported names from `definitions` output.**
  Regardless of how reached.
- **B. Allow them; user opts out with `swagger:ignore`.** Status
  quo.
- **C. Q7's fix dissolves the leak.** If §3.1 routes parameter
  aliases through the parameter builder (not the schema builder),
  the unexported RHS may stop being reachable as a top-level
  schema; no separate filter needed.

#### Tells

- **Go semantics:** lowercase ≡ package-private ≡ "callers from
  outside this package should not see this." Surfacing it in the
  spec is arguably a leak per Go semantics. Argues for A.
- **Counter:** the user wrote `type Exported = exportedParams` —
  an EXPLICIT bridge from public name to private struct. They knew
  what they were doing; filtering would lose information they
  wanted exposed. Argues for B.
- **Does C work?** Depends on whether the unexported struct is
  reachable via any other path (field type elsewhere, embed, …).
  Often yes — so C alone may not suffice.

#### Couplings

- §3.1 (parameters path), §3.3 (alias-vs-named for the RHS).
- If §3.1 option B is picked for parameters ("walk Y as the
  parameter set"), the unexported RHS is walked but not
  necessarily registered as a definition — C is plausible.

#### Empirical check

Probe: what other contexts emit unexported types into `definitions`
today? If alias-via-exported is the only path, A is cheap. If there
are other paths (embed-of-unexported, field-typed-by-unexported),
A is a broader policy decision.

---

### 3.8 Cross-cutting consideration — back-compat

Not a question per se but a constraint to bear in mind for every §3.X
answer. Every fixture under `fixtures/integration/golden/` is a
user-visible contract; changes that break those goldens need either
(a) a migration story, (b) an opt-out flag, or (c) explicit acceptance
that the old shape was wrong and the new shape is right.

Track which §3.X answers are golden-breaking as we go.

---

## 4. Constraints to honor

The decision must respect these. Listed up front so the discussion
doesn't accidentally propose violations.

### 4.1 Back-compat — existing user code

Every fixture under `fixtures/integration/golden/` is a user-facing
contract. Behaviour changes that break a non-trivial existing golden
need an explicit migration story or an opt-out flag.

### 4.2 Go type-system semantics

`type X = Y` is a **transparent rename** in Go's type system. `x` of
type `X` and `y` of type `Y` are assignable in either direction
without conversion; reflection cannot distinguish them. Any spec
output that differs *based on which name was written* is making a
choice the type system does not support.

This is a strong argument for §3.3 = "behave identically." It does
not foreclose the choice — codescan reads source files, not just the
type graph, and can make documentation choices the type system can't
— but the workshop should articulate WHY if it picks asymmetric.

### 4.3 Downstream consumers — code generators

The spec output is consumed by tools that generate client code, server
stubs, validators. They care about:
- Stable definition names across runs.
- $ref consistency (every $ref resolves; nothing dangling).
- AllOf composition staying decomposable (no inlining of refs that
  consumers tried to use as base types).

Changes that shuffle definition names or refs will ripple through
downstream-generated code. Where this is hard to avoid, the new
behaviour should produce *cleaner* refs than the old (e.g. fewer
duplicates), not more.

### 4.4 Q9's principled-asymmetry precedent

When B1 closed Q9 we accepted an asymmetry between struct field
naming and interface method naming because the two have genuinely
different runtime serializations (`encoding/json` mirrors struct
fields verbatim but can't marshal interface methods). The workshop
should be wary of inventing a similar "principled" justification for
the alias asymmetries without the same kind of structural reason.
Reverse-engineered post-hoc justifications are a smell.

### 4.5 OAS 2.0 only

This workshop is bounded to Swagger 2.0 / OpenAPI 2.0 spec output.
OAS 3.x's discriminator, anyOf/oneOf, and richer composition are NOT
in scope — see `.claude/plans/forthcoming-features.md` and the v2
vision memory.

---

## 5. Decisions (to be filled during discussion)

### 5.1 Intent per target (answers §3.1)

*[empty — to fill]*

### 5.2 Mode factoring (answers §3.2)

*[empty — to fill]*

### 5.3 Alias-vs-named symmetry (answers §3.3)

*[empty — to fill]*

### 5.4 RefAliases scope (answers §3.4)

*[empty — to fill]*

### 5.5 Expand mode with stdlib-special RHS (Q3, Q13, answers §3.5)

*[empty — to fill]*

### 5.6 Chain length (Q11, answers §3.6)

*[empty — to fill]*

### 5.7 Unexported visibility (Q12, answers §3.7)

*[empty — to fill]*

---

## 6. Action items (to be filled during discussion)

Once §5 is filled, each quirk gets a concrete action item:

- **Q3 implementation plan:**
- **Q7 implementation plan:**
- **Q8 implementation plan:**
- **Q11 implementation plan:**
- **Q12 implementation plan:**
- **Q13 implementation plan:**

Per-Q witness fixture roster (some are pinned today, some need adding
to make the alias-vs-direct comparison visible — e.g. an `EmbedsBase`
case alongside `EmbedsAlias` in the embedded-types fixture):

- *[to fill]*

---

## 7. Out of scope

Do **not** relitigate these during the workshop. Each has its own
landing zone.

- **Struct-vs-interface naming asymmetry** (Q9) — closed by B1 with a
  principled rationale documented in `schema/README.md`
  §method-mangler.
- **OAS 3.x model design** — see `project_v2_vision` memory and
  `.claude/plans/forthcoming-features.md`. The alias question is one
  of the items that v2's richer composition might dissolve, but v2 is
  not the answer here.
- **The `swag/mangling` library or any spec-rewriting downstream of
  codescan** — codescan emits, it does not rewrite. Whatever the
  workshop decides, the implementation lives inside the four builders.
- **DescWithRef** — a tangential flag about description text next to
  `$ref`. Out of scope unless the chosen alias model genuinely
  collides with it.

---

## 8. References

- `internal/builders/schema/embedded.go:40-105` — three-arm embed
  dispatch.
- `internal/builders/schema/schema.go:143-205` — `buildDeclAlias`
  three-mode dispatch.
- `internal/builders/schema/schema.go:257-285` — `buildAlias`
  two-mode dispatch.
- `internal/builders/parameters/parameters.go:102-140` — parameters
  `buildAlias`.
- `internal/builders/parameters/parameters.go:289-310` —
  `processFieldOnAlias` body-vs-non-body gating.
- `internal/builders/responses/responses.go:215-310` — responses
  alias dispatch.
- `internal/builders/schema/README.md` §aliases — current
  documentation.
- `fixtures/enhancements/alias-expand/` — Q7/Q11/Q12 fixture.
- `fixtures/enhancements/alias-findmodel-witness/` — A1-era alias
  witness (different concern but adjacent).
- `fixtures/enhancements/embedded-types/` — Q8 fixture (needs
  `EmbedsBase` direct case adding).
- `fixtures/enhancements/ref-alias-chain/` — Q3, Q11, Q13 witness
  (RefAliases mode only; needs a default-mode companion for Q3
  per §6 roster).

---

## 9. How to use this document

1. Read §1, §2 to load context.
2. Read §3 to see the questions on the table.
3. Read §4 to know the rails.
4. Conduct the discussion. As decisions land, write them into §5 in
   prose (full sentences, not just bullets — the *why* matters for
   future archaeology).
5. Once §5 is complete, fill §6's action items.
6. Mark Status: **decided**. Close C0 in `fix-quirks.md`. C1-C4 may
   now proceed in order.
