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

**Status:** ⬜ open.

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
  [default-allof-for-embeds](default-allof-for-embeds.md)).
- The `swagger:model` decl-side semantics are unaffected — an annotated alias
  decl carries its own `definitions` entry unconditionally (R2 in the workshop
  ledger), and this option does not alter that.

**Out of scope.** The parameters and responses builders have not yet received the
R6 treatment (Q7 / Q12 work, fix-quirks Phase C1/C2). When that lands, the same
option should govern those layers uniformly — the gate moves from
schema-builder-internal to a cross-layer concern. Until then this option only
affects the schema builder.

**When to revisit.** First time a user requests "I want every alias to be a type"
without per-decl annotation churn, or as part of an Options surface review.
