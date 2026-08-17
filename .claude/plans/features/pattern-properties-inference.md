---
title: withPatternProperties auto-inference
stream: 9
origin: iv
status: open
release: unscheduled
issues: []
prev: "§18.1"
---

# `withPatternProperties` auto-inference

**Status:** ⬜ open · **low priority (deferred)** — the output-contract decision
(replace `additionalProperties` vs emit both + `additionalProperties:false`) and
the `TextMarshaler` carve-out remain open.

**Origin.** (iv) deferred refinement on the landed additionalProperties work —
see [map-additionalproperties-keys](archive/map-additionalproperties-keys.md) and
[additionalproperties-control](archive/additionalproperties-control.md) as the shipped
parents. The unsolved tail recorded in `.claude/plans/additional-properties.md`
§7.

**Scope.** Auto-derive `patternProperties: {"^-?\d+$": V}` from a
non-string-keyed map (`map[int]V`) instead of the loose `additionalProperties: V`
we emit today — a *tighter* schema that constrains keys to integer-shaped strings.

**Why it's hard (why it's parked).**

- The regex is only synthesizable for the **integer family** (`^-?\d+$` signed,
  `^\d+$` unsigned, named integer types follow their underlying). A `TextMarshaler`
  key — also a legal JSON map key — has a key string we *cannot know statically*,
  so it can't participate; the knob would be inconsistent.
- It's a **behaviour choice, not an addition**: `additionalProperties: V` and
  `patternProperties: {…}` say different things. To actually mean "only
  integer-shaped keys" you need `patternProperties` **plus**
  `additionalProperties: false` — combining three keywords coherently.
- **Tooling reach:** `patternProperties` isn't OAS2 vocabulary (we emit it under
  the JSON-Schema-over-Swagger policy); making it the default for integer maps
  could regress strict OAS2 consumers — so it must be an opt-in knob with careful
  interaction rules.

**Shape (when taken up).** An opt-in `Options.WithPatternProperties` (working
name); decide whether it *replaces* `additionalProperties` with
`patternProperties`, or emits both with `additionalProperties: false`; carve out
`TextMarshaler` keys explicitly.

**When to revisit.** When integer-keyed-map precision is demanded, or alongside a
broader JSON-Schema-output pass.
