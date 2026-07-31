# `swagger:omit` — author-resolved embed conflicts

**Branch:** `feat/swagger-omit` — ✅ merged to master via PR #67 (`e5037f1`); worktree removed.
**Origin:** go-swagger#1992 + the `DefaultAllOfForEmbeds` override defects found 2026-07-30.
**Status:** ✅ done & merged (2026-07-30): `bc0fcb0` feature + `c3e225a` doc-site.

---

## 1. Why

Promoting an embed can produce a schema the author never intended, and codescan cannot decide the
intent for them:

| shape | Go marshals | inlined | `DefaultAllOfForEmbeds` |
|---|---|---|---|
| `Decorated`: outer `ID` re-declared to add `readOnly` | one `ID` | ✅ one, decorated | ⚠️ `ID` in **both** members |
| `Retyped`: outer `ID string` over `ID int64` | one `ID`, string | ✅ string | ❌ integer **and** string — unsatisfiable |
| go-swagger#1992: `Body struct{ models.User }` | all fields | all fields | all fields |

These are limit cases of a shared type being reused where it does not quite fit. The scanner will
not guess: it gives the author an explicit, in-band escape hatch and keeps mirroring the code
otherwise.

## 2. The annotation

```go
// swagger:omit <name>[,<name>…]
```

Names are **Go field names**, never JSON aliases — the annotation acts before names are computed.

Two placements:

- **on an embed** (ergonomic form) — targets are fields of *that* embedded type, no qualification:

  ```go
  Body struct {
      // swagger:omit ID,Created
      models.User
  }
  ```

- **on the enclosing type** (power form) — for embeds you cannot annotate (a type you do not own,
  or a deeply nested one). A bare name resolves against the promoted set; a dotted path names the
  embed chain:

  ```go
  // swagger:omit Base.ID,Created
  type Decorated struct { … }
  ```

**Embeds only.** Every path segment except the last must name an embedded field. A target naming
the struct's *own* field is a diagnostic, not a feature — `swagger:ignore` is that tool. This keeps
`omit` to what the schema owns (promoted content) rather than a general spec editor.

## 3. Semantics — a pre-filter, not a post-hoc delete

`omit` means **"do not promote this field when walking the embed"**, not "delete this property
afterwards". One rule then covers both renderings:

- **inlined**: the field is never written; an outer re-declaration wins as it already did (so
  `omit Base.ID` on `Decorated` is correctly a no-op there);
- **allOf**: the base member is built without the field, which removes `Decorated`'s duplicate and
  `Retyped`'s unsatisfiable pair.

No mode-awareness, and no Go-name↔JSON-name reconciliation, because filtering happens before names
exist.

## 4. Resolution — against the type, not the walk

Each target is resolved **once, at the annotation site**, with
`types.LookupFieldOrMethod(embedded, false, pkg, name)` — Go's own promoted-field lookup, so depth
and ambiguity rules come for free. What the scope carries is the resulting set of `*types.Var`, and
the filter is pointer identity, not string matching.

Consequences:

- no hit-tracking / "actually omitted" bookkeeping — a target either resolves against the type or
  it does not, and that is known before the walk starts;
- a target already excluded for another reason (unexported, `swagger:ignore`) resolves fine and is
  silently redundant — correctly not a diagnostic;
- no name-collision risk between depths: distinct fields are distinct objects.

## 5. Diagnostics (all Hints — informational, never blocking)

| code | fires when |
|---|---|
| `scan.omit-unresolved` | the target names no (promoted) field of the embedded type — typo, or renamed upstream |
| `scan.omit-behind-ref` | the target resolves, but the embed is emitted as `$ref` (a `swagger:model`), where OAS2 cannot subtract a property |
| `scan.shadowed-embed-field` | a field re-declared with `json:"-"` shadows a promoted field — suggests `swagger:omit` (see §7) |

`omit` is the only construct whose output depends on a hand-written name the compiler never checks;
everything else is derived from types. `scan.omit-unresolved` is what stops it rotting silently.

## 6. Implementation surface

| layer | change |
|---|---|
| `grammar/annotations.go` | `AnnOmit` kind + `labelOmit`, name↔kind arms |
| `grammar/lexer.go` | raw-remainder arg capture (the `AnnPatternProperties` precedent — commas and spaces) |
| `grammar/parser.go` | classifier arm (optional args, like `AnnAllOf`) |
| `scanner/index.go` | `"omit"` in the recognised-but-bitless annotation list |
| `schema/walker_classifiers.go` | `fieldDoc.OmitTargets` from `scanFieldDoc` |
| `schema/omit.go` (new) | target parsing, type resolution, the scope stack |
| `schema/allof.go` | gather + push the scope around each embed walk |
| `schema/struct.go` | the pre-filter itself (skip a field in the active scope) |

`parameters` needs no change for the embed-level form: a `swagger:parameters` body struct is built
through `schema.Builder`, so the embed walk is shared.

## 7. `json:"-"` eviction — related, decided separately

`fields.go` currently deletes a promoted property when an outer field re-declares it with
`json:"-"`. That is unfaithful to `encoding/json`, which ignores a `-` field *entirely* so it never
shadows anything — Go still marshals the promoted one. `swagger:omit` supersedes the hack.

This pass **only adds the Hint** (`scan.shadowed-embed-field`). Removing the eviction changes
existing output and breaks `TestOverridingOneIgnore`, so it is a separate decision, not folded in
here.

Registered as **Q31** in `archive/observed-quirks.md` (STILL PRESENT, decision open) with the four options
laid out. It also blocks repointing `fixtures/bugs/1992` at the mechanism the issue is really
about — that fixture currently witnesses `readOnly`, which belongs to #1063.

## 8. Phases

| # | Phase | Status |
|---|-------|--------|
| P1 | Grammar + classifier: `AnnOmit`, args, scanner recognition | ✅ done |
| P2 | Embed-level form: scope, resolution, pre-filter, both diagnostics | ✅ done |
| P3 | Type-level form: bare + dotted paths, own-field diagnostic | ✅ done |
| P4 | `json:"-"` shadow Hint | ✅ done |
| P5 | Fixtures + witnesses (inlined / allOf / parameters body / diagnostics) | ✅ done |
| P6 | Docs: annotation page + both indexes + allOf how-to section + witness; verified by a full hugo build | ✅ done |
