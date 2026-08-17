---
title: DiscoverAliasesAsTypes — opt-in pre-R6 alias discovery
stream: —
origin: ii
status: open
release: v0.37
issues: []
prev: "§3.7"
---

# `DiscoverAliasesAsTypes` — opt-in pre-R6 alias discovery

**Status:** ⬜ open · spec refreshed 2026-08-02 against the post-Q32 landscape (the old
"out of scope" note was stale — see the end).

**Origin.** W3 alias workshop Q-E close-out (2026-06-10). A core-product
enhancement with no dedicated stream; it rides whichever stream next touches the
alias-handling Options surface.

**Scope.** The schema builder's rule today is "an unannotated alias is a Go
implementation detail; it dissolves to its unaliased target at every use site,
and never produces its own `definitions` entry" (R6, enforced in
`internal/builders/schema/schema.go` `buildAlias` via the
`decl.HasModelAnnotation()` gate). The patch fixed Q-E — unannotated aliases no
longer manufacture dangling definitions, and field-site `$ref`s point at the
underlying target instead of the alias name. `swagger:model` is the user's
explicit opt-in for "expose this alias as a first-class spec entity."

Some users prefer the opposite default. When a codebase pervasively uses aliases
as domain-modelling vocabulary (`type UserID = int64`, `type Email = string`),
they'd rather every discovered alias surface as its own definition without
annotating each one. The pre-R6 behaviour did exactly that — the discovery loop
pulled each referenced alias into `ExtraModels` as a side effect of the
field-site `MakeRef`.

**Shape.** A new option `Options.DiscoverAliasesAsTypes bool` (default `false` so
existing users see the R6 behaviour). When `true`, every alias reachable through
the discovery walk gets a `definitions` entry and field-site `$ref`s point at the
alias name rather than dissolving to the target. The patch reverses the R6 gate
at the same `buildAlias` site:

- `false` (default): annotation gates first-class status (R6)
- `true`: every discovered alias is first-class (pre-R6 behaviour)

Interactions:

- `TransparentAliases=true` always wins. Transparent dissolves aliases at every
  use site by definition, so this option is inert under Transparent (the dissolve
  happens before the R6 gate is consulted).
- `RefAliases=true` × `DiscoverAliasesAsTypes=true` reproduces the pre-R6
  RefAliases chain shape (alias decl chains to its target via `$ref`; alias-name
  `$ref`s appear at field sites).
- `swagger:model` annotation is still honoured; this option just removes the
  requirement to annotate every alias to get it into `definitions`.
- The Q-D embed contract is independent: `swagger:allOf` still governs
  composition shape at embed sites, regardless of whether the embedded alias is
  annotated or auto-discovered under this option (see
  [default-allof-for-embeds](archive/default-allof-for-embeds.md)).
- The `swagger:model` decl-side semantics are unaffected — an annotated alias
  decl carries its own `definitions` entry unconditionally (R2 in the workshop
  ledger), and this option does not alter that.

**~~Out of scope.~~ Superseded 2026-08-02 — the cross-layer landscape is now known.**

The previous revision said "the parameters and responses builders have not yet received the R6
treatment … until then this option only affects the schema builder". That is **stale**: all three
builders now state and implement the shared contract. Each README says so
(`parameters/README.md` §alias-handling: *"shares the alias-handling contract with the schema and
responses builders"*; same in `responses/README.md`), and the gate exists in all three:

| builder | R6 gate | use-site handler |
|---|---|---|
| schema | `schema.go:440` (`refModel`), `schema.go:460` (dissolve) | `buildAlias` |
| parameters | `parameters.go:445` (`refModel`), `:487` (`$ref`) | `buildFieldAlias` |
| responses | `responses.go:390` (`refModel`), `:431` (`$ref`) | `buildFieldAlias` |

So the option is **already a cross-layer concern**, not a schema-builder-internal one, and it must
be honoured at three independent sites rather than one.

**What the Q32 work (`fix/strfmt-dispatch-symmetry`) adds to the design:**

1. **`buildAlias` has four callers**, not one — `buildFromType`, the `swagger:allOf` walk, the
   interface-side embed walk, and the stdlib-specials routing. An option that changes "does this
   alias keep its identity" has to be evaluated wherever the dissolve happens, which is all four,
   plus the two sibling `buildFieldAlias` implementations.
2. **The dissolve must be reached with the declaration already in hand.** The lookup used to sit
   *below* the `TransparentAliases` early return in all three builders, so that mode dissolved
   without ever reading the decl. It is now above it everywhere — which this feature depends on,
   since `DiscoverAliasesAsTypes` is precisely a decision about a declaration taken at a use site.
3. **There is now a precedent for a cross-layer alias concern**: `common.Builder.ClassifierAliasStrfmt`
   lives on the shared builder because all three needed identical behaviour before their dissolves.
   A `DiscoverAliasesAsTypes` gate belongs in the same place, for the same reason — three copies of
   the rule is how the strfmt defect happened.

**Interactions — one correction.** The previous revision said `TransparentAliases=true` "always
wins … so this option is inert under Transparent (the dissolve happens before the R6 gate is
consulted)". The *conclusion* still holds — Transparent dissolves at every use site by definition —
but the *reason* given is no longer accurate: the decl lookup now precedes the Transparent return, so
the gate is reachable there. Inertness under Transparent is therefore a **choice the implementation
must make explicitly**, not something the control flow enforces for free.

**Effort re-estimate.** Higher than the previous revision implied: six sites, not one, and the gate
belongs on `common.Builder` rather than in `schema.buildAlias`. Against that, the plumbing it needs
(decl-before-dissolve, in all three builders) now exists.

**When to revisit.** First time a user requests "I want every alias to be a type" without per-decl
annotation churn, or as part of an Options surface review. Still no user demand recorded.
