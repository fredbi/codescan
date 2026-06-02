# W3 — Alias workshop, cycle 4 (parameters)

Date: 2026-06-10
Status: ⬜ cycle 4 kickoff (calibration fixture + pre-patch goldens)
Companion: `alias-handling.md` (overall workshop), `alias-ledger.md`
(judgment ledger; cycle 4 entries appended here), `alias-matrix.md`
(matrix audit), `fix-quirks.md` C1 / C2 (the formal task slots).
Branch: `fix/aliased-types` (after the cycle 1-3 R6 wave).

This doc anchors the parameters analogue of cycles 1-3. The schema
builder converged on **R6**: annotation gates whether an alias
surfaces as a first-class spec entity at use sites. The same rule
needs to land in the parameters builder, with **R7** as its
parameters-flavoured restatement plus a top-level-decl-specific
clause that doesn't have a schema analogue.

This document is **workshop input**, not a decision record.

---

## 1. Why a cycle 4 at all

Fred's read (2026-06-10):

> The parameters builder's alias handling was never really
> implemented even though there is a switch/case that catches an
> alias, I don't think anything reliable is built on top of that.

The current state of `enhancements_alias_expand.json` confirms it
empirically:

- **`paths: {}`** — the alias-expand fixture defines four
  `swagger:parameters` annotations (`aliasedRequest`, `aliasedTop`,
  `aliasedTop2`, plus the response side) and *no* `swagger:route` —
  so paths is correctly empty for that fixture, but the parameter
  sets themselves are wrongly registered as models in
  `definitions`. Q12 leaking on every annotated decl.
- **Definitions polluted**: `AliasedTopParams`, `AliasedTopParams2`,
  `PayloadAlias`, `PayloadAlias2`, `QueryIDAlias`, `exportedParams`
  — none of these should be top-level model definitions. The
  parameters builder's `buildAlias` calls `AppendPostDecl(decl)`
  unconditionally on both the alias and its target, with no
  annotation gate.

This is the same shape of bug that R6 fixed in the schema builder,
plus a top-level-`swagger:parameters` clause: a parameters
declaration aliased to a backing struct should never produce a
model entry for either layer of the chain.

---

## 2. R7 — the candidate rule

R7 is R6 lifted to the parameters layer, with one extra clause for
the top-level-decl case that has no schema analogue.

| Reach context | R7 |
|---|---|
| **Top-level alias annotated `swagger:parameters`** | The alias is **transparent re: model creation**. Neither the alias decl nor any chain link of its target appears in `definitions`. The fields of the underlying struct become the operation's parameters. |
| **Alias as a body field type within a parameters struct** | **R6 applies.** Annotation gates first-class status: annotated body alias → `$ref: <AliasName>` (under default / Ref); unannotated → dissolve to target. |
| **Alias as a non-body field type within a parameters struct** | Always expand to the unaliased target (SimpleSchema can't carry `$ref`). Annotation has no effect at non-body sites — the SimpleSchema constraint already forces inline. |
| **`TransparentAliases=true`** | Supersedes — dissolves at every use site (existing behaviour). |
| **`RefAliases=true`** | Body fields chain via `$ref` (existing behaviour); non-body fields still expand. The annotation gate runs *before* the chain decision. |

The top-level clause is the genuinely new rule: even when the user
explicitly opts in via `swagger:parameters X = Y`, the alias and
its backing struct must not produce model definitions. The
`swagger:parameters` annotation does NOT promote anything to
`definitions`; it declares a parameter SET, not a model.

If the same decl carries BOTH `swagger:parameters` and
`swagger:model`, the model annotation still wins for the
schema-side surfacing — that's two orthogonal annotations doing
two orthogonal jobs.

---

## 3. Axes for the parameters builder

Derived from `parameters/parameters.go` decision branches. Some
overlap with the schema axes in `alias-matrix.md` §1, others are
parameters-specific.

### 3.1 Reach context (parameters-specific)

| Value | Code path |
|---|---|
| **Top-level decl** — `type P = Q` annotated `swagger:parameters` | `parameters.buildAlias` (line 102) |
| **Body field** — body parameter typed as an alias | `parameters.buildFieldAlias` (line 272), `In() == body` |
| **Non-body field** — query / path / header / form-data | `parameters.buildFieldAlias`, `In() != body` |
| **Chain** — `type A = B = Target` for either of the above | recursion through the same paths |

4 cells. Body vs non-body matters for the chain shape (`RefAliases`
only applies to body).

### 3.2 Annotation context

| Value | Effect |
|---|---|
| `swagger:parameters` on the alias decl (top-level) | Triggers the parameters builder entry. R7 clause 1. |
| `swagger:model` on a body-field alias decl | R6 applies (clause 2). |
| Unannotated alias as body field | R6 dissolve (clause 2). |
| Unannotated alias as non-body field | Forced expand (clause 3). |

### 3.3 RHS kind (subset of schema's axes)

The parameters builder rejects `any` and `error` outright (line
104). RHS kinds in play:

- `*types.Named` regular (struct backing a parameters set; named
  primitive backing a non-body field)
- `*types.Alias` (chain links)
- `*types.Basic` (only via alias-of-basic at body field site —
  unusual but possible)
- `*types.Struct` (anonymous — `type X = struct{…}` annotated
  `swagger:parameters` — degenerate)

### 3.4 Mode

Same as schema: Default, RefAliases, TransparentAliases. The R7
rule applies in all three for the unannotated cases (R6 ports
cleanly). For annotated body fields, mode controls the chain shape.

---

## 4. Calibration fixture sketch

The cycle-4 calibration fixture mirrors the cycle-3 design: bounded
canvas, side-by-side annotated vs unannotated variants, real
operation wiring so `paths` populates and we can see body/query
semantics in the goldens.

**Decls** (`fixtures/enhancements/alias-parameters-calibration/`):

- `Payload` — annotated `swagger:model`, canonical body model.
- `QueryID` — named string, non-body alias target.
- `PayloadAlias = Payload` — unannotated.
- `PayloadAliasModeled = Payload` — annotated `swagger:model`.
- `PayloadAlias2 = PayloadAlias` — 2-link unannotated chain.
- `QueryIDAlias = QueryID` — unannotated, non-body backing alias.
- `internalParams` — unexported backing struct for the top-level
  aliased parameter set; one body field, one query field. Q12
  witness.
- `AliasedTopParams = internalParams` — annotated `swagger:parameters
  aliasedTop`. Top-level R7 witness.
- `DirectParams` — annotated `swagger:parameters directParams`,
  the control case. Three body fields (direct, plain alias,
  annotated alias) + one non-body alias field.

**Routes** (`handlers.go`):

- `swagger:route GET /aliased-top aliasedTop` — binds
  `AliasedTopParams` to an operation so `paths` populates.
- `swagger:route POST /direct directParams` — binds
  `DirectParams`.

**Cells exercised:** 9 decls × 3 modes = 27 cells across cycle 4.

---

## 5. What the workshop has to decide

§3 / §4 enumerate the surface; the corpus walk decides the shape.
Pre-patch goldens capture the current state; post-patch goldens
will pin the converged R7 shape. Mid-cycle judgments accumulate in
`alias-ledger.md` cycle 4.

Open questions to clear during the walk:

- **Q-G — top-level decl shape.** Confirm R7 clause 1 visually:
  does the user expectation match "no definitions for
  AliasedTopParams or internalParams"? Or do users want the alias
  chain visible in the spec for documentation purposes (despite
  the leak diagnosis)?
- **Q-H — chain dissolve depth.** When `AliasedTopParams =
  internalParams` and `internalParams` itself is unexported,
  should the dissolve walk land at `internalParams`'s fields, or
  is there a case for treating `internalParams` as a transient
  step too? Probably the former (it's where the field declarations
  live), but worth checking on the canvas.
- **Q-I — what about `swagger:response` aliases?** Same rule
  shape, different builder. Likely Q-G's answer ports
  symmetrically. Defer to cycle 5 unless the cycle 4 walk surfaces
  a coupling.

---

## 6. Methodology checklist

Following the empirical corpus walk methodology validated in
cycles 1-3 (see `feedback_workshop_methodology` in the project's
memory index):

- [x] Code-derived axes (§3)
- [x] Calibration fixture sketch (§4)
- [ ] Pre-patch goldens captured (cycle 4 kickoff commit)
- [ ] Fred's TUI side-by-side judgment on the pre-patch state
- [ ] Cycle 4 ledger entries with ✓ / ✗ / ? per cell
- [ ] R7 patch in `parameters.buildAlias` / `buildFieldAlias`
- [ ] Post-patch goldens
- [ ] Bidirectional witness variant (annotated vs unannotated, like
      cycle 3)
- [ ] Cycle 4 summary appended to the ledger; Q-G/H/I resolved

---

## 7. Out of scope for cycle 4

- **`swagger:response` alias handling** — Q-I above. Same shape
  of bug in the responses builder; defer to cycle 5.
- **Q11 chain duplication for annotated chains** — schema-layer
  question; orthogonal to parameters cleanup.
- **Q-F (annotated alias decl shape under Default vs Ref)** —
  schema-layer judgment call still pending; not blocking.
- **Phase D wrap-up** — pre-PR cleanup, runs after the cluster
  closes.
