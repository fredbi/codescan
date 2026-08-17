---
title: After-declaration annotation comments — design & plan
stream: 8
origin: i
status: in-progress
release: v0.36
issues: []
supersedes-stub: comment-source-filtering.md
---

# After-declaration annotation comments — design & plan

Opt-in (`Options.AfterDeclComments`) that lets swagger annotations live **inside**
a declaration (the leading comment of a func/struct body) or **inlined** as a
trailing comment, so the godoc *above* the decl stays clean and human-facing.

**Hard constraint (Fred):** this is **solely a scanner concern** — no grammar
change, no builder change. The scanner enriches the comment fields the builders
already read; the existing parse/build pipeline does the rest.

Backing stub: [comment-source-filtering.md](comment-source-filtering.md). Clean-
godoc cluster with [swagger-description-override](swagger-description-override.md)
(landed) and [godoc-filter](godoc-filter.md) (separate).

---

## 0. The shapes

```go
// ThisFunction does the job.            ← clean godoc (title)
func ThisFunction() error {
    // thisFunction is the API operation. (description)
    //
    // swagger:route GET /job doJob
    return doJob()
    // anything here is IGNORED (not a leading body comment)
}

// ThisType is the best type.            ← clean godoc (title)
type ThisType struct {
    /* Why best? minimalism. (description)
       swagger:model bestType */
    Anything AnyType                      // ordinary field doc still works
}

const X = "v"  // swagger:enum            ← inlined trailing
type  Y = int  // swagger:model apiType   ← inlined trailing
```

---

## 1. Architecture — enrich what the builders already read

Audit of every comment read in the builders: they consume comment text through
**exactly two sources** —

1. **`EntityDecl.Comments`** — the decl-level group (`s.Decl.Comments`,
   `ParameterRef.Comments`). The scanner *constructs* this (`index.go:271`:
   `comments := ts.Doc ?? gd.Doc`).
2. **`ast.Field.Doc`** — the struct field's doc group (`afld.Doc`), read directly
   by every field walker.

…plus the scanner-side, location-agnostic `file.Doc` (meta) and `file.Comments`
(routes/operations). **No builder reads `.Comment` / `TypeSpec.Comment` /
`ValueSpec.Comment`.** So the feature reduces to: when the option is on, the
scanner makes the located inside/trailing comments **part of `EntityDecl.Comments`
or `ast.Field.Doc`**, and everything downstream is unchanged.

Two enrichment styles, by target:

- **`EntityDecl.Comments` — pure construction (no shared mutation).** The scanner
  already builds this value; for the decl-level shapes it builds a *merged*
  `*ast.CommentGroup` (`docAbove.List ++ located.List`, original nodes untouched).
  Idempotent by construction — `ts.Doc` is never mutated.
- **`ast.Field.Doc` — shared-AST rewrite.** The builder reads the real
  `*ast.Field.Doc`, so field-level enrichment must append `Field.Comment` onto
  `Field.Doc` in place. This is the only mutation of the shared tree, and the
  only case that needs an **idempotency guard** (see §4).

Positions stay ascending in every merge (doc above < inside/trailing below), so
the grammar's line reconstruction sees a blank-line gap between the clean godoc
and the annotation block and parses it without any grammar change. (Verified by
AST probe + confirmed against `Preprocess` in P1.)

---

## 2. Locating the comments (per shape)

The "leading body comments" rule: **every comment group before the first body
element** (first field for a struct, first statement for a func) that is **not a
field's `.Doc`**. Comments interspersed with / after code or fields are not
leading → ignored. (A leading group with no annotation is harmless — it just adds
inert prose; "ignored" in the examples means "no effect", not "specially
excluded".)

| Shape | Located comment(s) | From |
|---|---|---|
| struct type | leading inside-body groups (`Fields.Opening` < pos < first field; not a field `.Doc`) | `file.Comments` position scan |
| func (route/op) | leading inside-body groups (`Body.Lbrace` < pos < first stmt) | already discovered — see §3 |
| alias / non-struct type | trailing `TypeSpec.Comment` | direct AST field |
| const | trailing `ValueSpec.Comment` | direct AST field |
| struct field | trailing `Field.Comment` (and/or a leading comment below the field) | direct AST field |

---

## 3. Routes / operations already work

`collectRoutePathAnnotations` / `collectOperationPathAnnotations`
(`index.go:176-182`) iterate **all** `file.Comments` and parse a path annotation
from each — they are already location-agnostic. A `swagger:route` inside a func
body is therefore already discovered today (its description comes from that same
comment group). Scope here is **verify + fixture**, no code.

---

## 4. Idempotency (Phase B)

Field-level enrichment mutates the shared `*ast.Field.Doc`. To guarantee a given
field is enriched at most once (defensive against any revisit), `TypeIndex`
carries a guard set — `enrichedFields map[*ast.Field]struct{}` — checked before
the append. Decl-level (Phase A) needs no guard (fresh `EntityDecl.Comments`,
`ts.Doc` never mutated).

---

## 5. Scope / phasing

- ✅ **P1 (Phase A) — decl-level type carriers** (`b8329e9`).
  `Options.AfterDeclComments`; `afterDeclSource`/`leadingBodyComments`/
  `mergeCommentGroups` in `index.go` merge struct inside-body groups + alias/
  non-struct trailing `TypeSpec.Comment` into a fresh `EntityDecl.Comments`
  (ts.Doc never mutated → idempotent). Discovery + naming + validation +
  clean-godoc title/description all flow through unchanged. Routes/operations
  confirmed already position-agnostic (route inside func body discovered).
  Fixture `enhancements/after-decl-comments` + golden + off/on coverage test;
  off ⇒ byte-identical. Suite green, lint clean.
- ✅ **P2 (Phase B) — struct fields** (`c09e382`). `enrichStructFields` appends
  each `Field.Comment` onto `Field.Doc` during indexing, guarded by
  `TypeIndex.enrichedFields` (the one shared-AST mutation). Fixture `created`
  field (`// swagger:strfmt date`) → `format: date`.
- ❌ **P3 (Phase C) — const enum: DEFERRED.** `swagger:enum` is type-based
  (resolves a *type*, collects its consts via `FindEnumValues`), so a standalone
  `const X = … // swagger:enum` is not an enum carrier and has no builder
  semantics today. Supporting it would require new builder behaviour, breaking
  the scanner-only constraint — deliberately out of scope. (The type-alias
  inlined forms Fred listed — `type X = Y // swagger:model` — are covered by
  Phase A via `TypeSpec.Comment`; witnessed by the `stampType` fixture case.)
- ✅ **P4 — README + stub** (this commit). `scanner/README.md` §after-decl +
  stub flipped to done.

One commit per phase on `feat/feature-v0.36`; suite green + lint clean each; no
merge before review.

## 6. Option

`Options.AfterDeclComments bool` (default false), re-exported via
`codescan.Options`. Threaded to `TypeIndex` where `processDecl` runs.

## 7. Fixtures

`fixtures/enhancements/after-decl-comments/` scanned **off** and **on**:
1. struct model: clean godoc above, inside-body `swagger:model name` +
   `maxProperties:` ⇒ discovered/named/validated; godoc = description. off ⇒ not
   discovered.
2. parameters / response carriers: inside-body annotation.
3. alias trailing: `type X = Y // swagger:model apiType`.
4. route inside func body (§3 verify).
5. (Phase B) field trailing `// swagger:strfmt date`.
6. (Phase C) const trailing `// swagger:enum`.
7. off-regression: whole fixture with the option off ⇒ inert.

Coverage test `internal/integration/coverage_after_decl_comments_test.go`
(off vs on) + golden(s).
