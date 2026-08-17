---
title: Naming from struct tags (form:, schema:, …)
stream: 9
origin: iii
status: done
release: v0.36
issues: [go-swagger#2912, go-swagger#1391]
prev: "§7"
---

# Naming from struct tags (`form:`, `schema:`, …)

**Status:** ✅ done (2026-06-23, branch `feat/feature-v0.36`). Option named
**`NameFromTags`**. First feature of the v0.36 streak.

**As built.**

- `Options.NameFromTags []string` — ordered list of struct-tag *types* the name
  is sourced from. The first listed type that yields a *usable* name wins; a tag
  that's absent, or present but nameless (`,omitempty` only) or `-`, is skipped
  and the next type is tried.
- **nil / unset ⇒ `["json"]`** (byte-identical to historic behaviour);
  **explicit empty slice ⇒** Go field name (no tag consulted). The nil-vs-empty
  distinction is the opt-out switch (getter `ScanCtx.NameFromTags()`).
- **Name only.** The encoding/json directives `-` (exclude), `,omitempty`,
  `,string` are ALWAYS read from the `json` tag, independent of the setting
  (decided with Fred: "we don't care about other attributes than the name"). So
  `json:"-"` still excludes even when the name comes from `form`; `form:"-"` does
  NOT exclude (its `-` is just a non-usable name → fall through).
- Applies across schema properties, parameters, response headers — one shared
  resolver `resolvers.ParseFieldTag(field, goName, nameTags)` (renamed from
  `ParseJSONTag`), threaded at all four call sites via `*.Ctx.NameFromTags()`.
- Multi-name groups (`R, G, B uint8`) keep each member's Go name whatever the tag
  type (a single rename can't name N members) — go-swagger#2638 invariant held.
- Targeted overrides unchanged and still win: `name:` keyword (params/headers),
  `swagger:name` / `swagger:model {name}` (schema). Embed-nesting still keyed off
  the json tag (`ExplicitJSONName`, a Go encoding/json structural concern).

**Tests.** `resolvers_test.go::TestParseFieldTag` (unit, incl. fallthrough /
empty-list / json-directive-independence); `fixtures/enhancements/name-from-tags/`
+ `internal/integration/coverage_name_from_tags_test.go` (e2e across the three
modes × schema/param/header).

---

_Original grooming notes (2026-06-23):_

**Origin.** go-swagger#2912 (backlog verification, 2026-06-13). Names derive from
the `json:` tag (then the Go field name); the `form:` tag is not consulted, so a
field tagged only `form:"sort_key"` is named `SortKey`. Also go-swagger#1391
(`schema:` tag). E.g. `form:` is used by gin.

**Decision (groomed 2026-06-23).** Add an ordered tag-precedence option — working
name `NameTags []string` (Fred's example spelled it `ParameterNameTags`, but the
setting is **global, not parameter-only**, so confirm the name).

- **Applies across the board** — schema properties, parameters, response headers
  — wherever a name is derived from a struct field.
- **Default `["json"]`** (current behaviour).
- **Precedence = the list order**, first present tag wins. The order the tags
  appear *on the struct field* is irrelevant. Example with
  `NameTags = []string{"form","json"}`:

  ```go
  X int `json:"x" form:"y"`  // named "y" — form: wins (first in the list)
  ```

- **Empty list ⇒ struct tags ignored entirely**; the name falls back to the Go
  field name.

**Targeted overrides are unchanged and still win.** Per-field rename via the
`name` keyword (parameters / response headers) and per-schema rename via
`swagger:model {name}` take precedence over any tag-derived name.

**When to revisit.** When a user need recurs, or alongside parameters builder work.
