---
title: swagger:description / swagger:title overrides — design & fixture harness
stream: —
origin: ii (Q30 close-out)
status: draft-for-review
release: v0.36
issues: []
supersedes-stub: swagger-description-override.md
---

# `swagger:description` / `swagger:title` overrides — design & harness

This is the reviewable design + fixture contract for the override annotations.
The backlog stub ([swagger-description-override.md](swagger-description-override.md))
groomed the *decision*; this doc locks the *grammar, scope, semantics, and
fixtures* before any code lands. **Read §3 (the multi-line fork) first — it is the
one genuinely open architectural question.**

---

## 0. Problem (grounding)

Every decl that surfaces in the spec carries its **godoc** as `title` /
`description`. The godoc is written for *Go* readers; sometimes it must diverge
from the *API* text (Q30: `time.Time`'s 2305-char monotonic-clock godoc leaks
into any spec that touches it). The rules are correct — the gap is the lack of an
**override affordance**. Groomed 2026-06-23 as *annotation-driven, not an option*.

What exists today (verified):
- Title/description are **purely prose-derived**:
  - model schema: `schema/walker.go:77-78` — `PreambleTitle()` / `PreambleDescription()`
  - field schema: `schema/walker.go:111-112` — `Prose()`
  - response decl / header: `responses/walker.go:74,126` — `Prose()`
- There is **no** `title`/`description` keyword or annotation today.
- Classifier annotations (`swagger:strfmt`, `swagger:type`, `swagger:name`, …)
  are grammar-owned: `AnnotationKind` enum (`grammar/annotations.go:14-43`), arg
  lexed in `classifyAnnotationArgs` (`lexer.go:284`), surfaced as
  `Block.AnnotationArg()`, consumed by builders via `ParseBlocks` +
  `scanFieldDoc` (`schema/walker_classifiers.go:306-351`).
- `swagger:patternProperties` already captures the **whole rest-of-line verbatim**
  as one `TokenRawValue` (`lexer.go:303-307`) — the precedent for free-text args.
- The multi-line **body accumulator** (`collectRawBlock`/`collectRawValue`,
  `lexer.go:618+`) is **keyword-driven** (`ShapeRawBlock`/`ShapeRawValue`), not
  annotation-driven. `swagger:enum`'s multi-line body rides the `enum` keyword's
  raw-value path; `EnumDeclBlock.BodyValues` is the only classifier-block with a
  multi-line body, and it is special-cased in `parseClassifierBlock`
  (`parser.go:558-562`).

---

## 1. Decisions locked (this review)

| # | Decision | Source |
|---|----------|--------|
| D1 | **Annotation form** `swagger:title` / `swagger:description`, not a keyword, not an Option. | groomed stub + Fred 2026-06-25 |
| D2 | **Scope = schemas (models + fields) + responses & headers.** Not parameters, not operations (this round). | Fred 2026-06-25 |
| D3 | `swagger:title` is **schema-only** (models + fields). Responses/headers have no `title` field in spec 2.0 → title there is a context error. | spec 2.0 shape |
| D4 | `swagger:description` applies to models, fields, response decls, response headers. | D2 |
| D5 | **Precedence**: explicit override *wins*; prose remains the fallback when the override is absent. No behavior change for any existing fixture (no override ⇒ identical output). | groomed |
| D6 | **Override replaces, does not append.** When `swagger:description`/`swagger:title` is *present* (empty or not), the prose-derived value is discarded for that decl (the whole point: decouple from godoc). | this doc |
| D7 | **Empty override ⇒ wipe + warn.** Combined `trim(inline + body)` empty (bare `swagger:description`/`swagger:title`, or whitespace/blank-only body) → **apply the empty value** (godoc text is suppressed) **and** emit a `scan.empty-override` **warning** (probably unintended). Empty thus *is* the suppression affordance — it serves the Q30 godoc-leak origin directly. The warning flags the case in case it was accidental. | Fred 2026-06-25 |
| D8 | **§3 fork resolved: Option B, blank-line terminator.** | Fred 2026-06-25 |

Deliberately **out of scope** (noted, not built): parameters, operation
summary/description, and stdlib types reached via embed/field (no
user-controlled decl to annotate — Q30 residual).

---

## 2. Grammar — annotation registration

New `AnnotationKind` members (all mechanical, mirrors existing classifiers):

- `grammar/annotations.go`: add `AnnTitle`, `AnnDescription`; labels
  `"title"` / `"description"`; wire `String()` + `AnnotationKindFromName()`;
  `family()` → `familyClassifier`.
- `lexer.go classifyAnnotationArgs`: add a `case AnnTitle, AnnDescription` that
  captures the **whole rest-of-line verbatim** as a single `TokenRawValue`
  (identical to the `AnnPatternProperties` arm, `lexer.go:303-307`). This gives
  the single-line arg for free; `ClassifierBlock.AnnotationArg()` returns it.
- `parser.go parseClassifierBlock`: **no** `CodeMissingRequiredArg` parse-error
  for a missing inline arg — under Option B (§3) a body on the following lines may
  legitimately supply the text, so emptiness can only be judged *after* body
  accumulation. The empty case is a **warning at the grammar layer** once combined
  text is known (D7), not a parse error.
- Scanner classifier closed-set: adding to `AnnotationKindFromName` is sufficient
  for recognition (no new `detectNode` bit needed — these are field/decl-level
  decorations, like `strfmt`/`type`). Confirm `index_test.go`'s recognized-token
  list is extended so the classifier does not raise "unknown annotation".

---

## 3. Multi-line description capture — LOCKED: Option B, blank-line terminator

**Resolved (Fred 2026-06-25): Option B with a blank-line terminator.** The
options below are retained for the record; the live design is **Option B** plus
the empty-override rule (D7) and the §3.1 contract.

The groomed stub says `swagger:description [raw block]` *"may be followed by a
multiline comment block (the raw block carries through)."* `swagger:title` is
single-line by nature. So only **description** needs a multi-line story, and the
question is purely *how the lexer folds the following comment lines into the
description body*. Three candidate mechanisms:

### Option A — single-line only (rest-of-line)
`swagger:description <text>` on one line; no body. Simplest (zero accumulator
work — just the §2 `TokenRawValue` arm). **Contradicts the groomed "raw block
carries through".** Multi-paragraph text stays in godoc (the default source);
the override is a one-line curated string.
- 👍 trivial, no new lexer surface, no ambiguity.
- 👎 a long curated description must be one physical comment line.

### Option B — inline head + trailing raw body, enum-style  *(recommended)*
Mirror `EnumDeclBlock`: the annotation's inline arg is the *first* line; following
comment lines accumulate into a raw body until a **blank line, the next
`swagger:` annotation, or EOF**. Combined text = inline arg + body lines joined.
Special-case in `parseClassifierBlock` exactly as the `enum`/`RAW_VALUE_ENUM` arm
already is (`parser.go:558-562`); expose via a `DescriptionBlock.Body` field (or
reuse the base-block prose, see open Q1).
- 👍 faithful to the groomed contract; reuses the enum precedent; intuitive
  ("write the description under the annotation").
- 👎 needs a small lexer hook so `swagger:description` opens a raw body (the
  accumulator is keyword-driven today); must define the terminator precisely.

### Option C — promote `title`/`description` to raw-body *keywords*
Register `title:` (ShapeString) and `description:` (ShapeRawBlock) as **keywords**
and let the existing body accumulator do all the work; the `swagger:`-prefixed
annotation becomes a thin synonym. Reuses `collectRawBlock` with zero new
accumulator code.
- 👍 maximal reuse; multi-line "just works".
- 👎 reopens the keyword-vs-annotation choice D1 explicitly closed; two surfaces
  to document; `description`/`title` as bare keywords risk colliding with prose.

**Recommendation: Option B.** It honours D1 (annotation form) and the groomed
multi-line contract with the least new vocabulary, and it has a direct precedent
(`swagger:enum` body). Option A is the fallback if we want to ship the 80% case
this week and defer multi-line; Option C is rejected (fights D1).

### 3.1 Locked contract for Option B

- **Terminator: blank line.** The body is the run of contiguous, non-blank
  comment lines immediately following the `swagger:description` head. The **first
  blank line ends the body**; so does the next `swagger:` annotation or EOF
  (a blank line is the primary terminator). A paragraph break (blank line)
  therefore *cannot* appear inside an override description — multi-paragraph text
  stays in the godoc. This keeps the rule trivial to read at the call site.
- **Combined text** = `trim(inlineArg + "\n" + join(bodyLines, "\n"))`. Inline
  arg (rest of the head line) and body lines both contribute; either may be
  empty.
- **Empty ⇒ wipe + warn (D7).** Combined text empty → the override is still
  applied (godoc suppressed) and a `scan.empty-override` warning is emitted at the
  grammar layer. On a schema an empty description simply omits the field
  (`omitempty`); on a **response** the (spec-2.0-required) `description` becomes
  empty — the warning is especially apt there.
- **`swagger:title` is single-line only.** It takes the inline rest-of-line; it
  does **not** open a body. A blank-separated block under `swagger:title` is not
  part of the title (it reverts to ordinary comment prose). Bare/empty title →
  D7 (wipe + `scan.empty-override`).
- **Body storage:** dedicated field on the produced block (e.g.
  `DescriptionBlock{Body string}`), **not** `baseBlock` prose — because
  `parseClassifierBlock` runs `extractTitleDesc` over the *whole* comment-group
  token stream, so all sibling blocks share the same `Prose()`; a dedicated field
  avoids that cross-talk. Special-case the body accept in `parseClassifierBlock`
  exactly as the `enum`/`RAW_VALUE_ENUM` arm already is (`parser.go:558-562`).
- **Placement after the decl** (via
  [comment-source-filtering](comment-source-filtering.md)) must behave
  identically — the blank-line terminator is position-independent.

### 3.2 Keyword co-location — schema family (resolved Fred 2026-06-25)

Surfaced during P4: `swagger:title` / `swagger:description` were first wired as
**classifier**-family annotations. But the classifier family *rejects* any
co-located schema keyword — so a field carrying both an override and an inline
validation (`swagger:description … ` + `maximum: 1000`) would drop the
`maximum:` with a `parse.context-invalid` warning. This is the universal
classifier-family behaviour (verified: `swagger:strfmt`/`swagger:type` reject
co-located keywords identically); the lone exception is `swagger:name`, which
was deliberately routed to the **schema** family precisely so a renamed field
can still carry validations.

**Decision:** route `title`/`description` to the **schema** family too (they are
orthogonal decorations like a rename, naturally co-occurring with validations).
`parseSchemaBlock` returns a `ClassifierBlock` carrying the free-text arg
(`AnnotationArg`) plus any harvested body keywords; `allowedContexts` returns
nil for these kinds (permissive, like `swagger:name`). Residual obscure
limitation: a field with *both* `title` *and* `description` *plus* keywords
splits into two annotation blocks and the schema walker only dispatches the
first, so keywords in the second block are still dropped — field titles are
themselves rare, so this is a corner of a corner.

---

## 4. Builder consumption (per target)

All sites already iterate sibling classifier blocks (`ParseBlocks`) or have a
block in hand. Apply the override **after** the prose-derived assignment, and
apply whenever the override is *present* — empty included (D6/D7). The
`scan.empty-override` warning is raised once at the grammar layer (D7), so
builders apply unconditionally when the override block is present.

| Target | Site | Action |
|--------|------|--------|
| model schema | `schema/walker.go applyDeclCommentBlock` (`:77-78`) | after setting `Title`/`Description` from preamble, scan sibling blocks; `AnnTitle` → `schema.Title = combined`; `AnnDescription` → `schema.Description = combined` |
| field schema | `schema/walker_classifiers.go scanFieldDoc` (`:306-351`) → applied in `applyBlockToField` (`:111-112`) | add `AnnTitle`/`AnnDescription` arms to `fieldDoc`; apply after `Prose()` |
| response decl | `responses/walker.go applyBlockToDecl` (`:74`) | `AnnDescription` → `resp.Description = combined`; **`AnnTitle` here → `CodeContextInvalid` diag** (D3), not applied |
| response header | `responses/walker.go applyBlockToHeader` (`:126`) | `AnnDescription` → `header.Description = combined`; `AnnTitle` → `CodeContextInvalid` diag, not applied |

**Diagnostic codes (resolved):** `scan.empty-override` (warning, D7) for an empty
combined value; `CodeContextInvalid` for `swagger:title` on a response/header
(D3). **Fail-loud (per [[feedback_go_types_defensive_guards]] /
fail-loud-diagnostics):** an override annotation that can't be consumed in its
context emits a located diagnostic, never a silent drop.

---

## 5. Fixtures & goldens

One fixture dir under `fixtures/enhancements/description-title-override/`,
exercising every consumed target + the context-error path. Golden:
`fixtures/integration/golden/enhancements_description_title_override.json`.

Cases (one decl each, so the golden reads cleanly):
1. **Model**: godoc title+desc present; `swagger:title` + `swagger:description`
   override both → spec shows the overrides, not the godoc.
2. **Field**: a struct field whose godoc would leak; `swagger:description`
   overrides → property description is the override; sibling field with
   `swagger:title` → property title set.
3. **Response decl**: `swagger:description` overrides the response description.
4. **Response header**: `swagger:description` on a header field.
5. **Multi-line** (gated on §3 Option B): a `swagger:description` with a trailing
   raw body → multi-line description carries through.
6. **Context error**: `swagger:title` on a response → emitted as
   `scan.context-invalid` (or chosen code), description-only response otherwise
   intact. Assert the diagnostic via the coverage test, not the golden.
7. **No-override regression**: a decl with godoc and no override → byte-identical
   to today (guards D5).
8. **Empty override (D7)**: a decl with godoc + a bare `swagger:description` (no
   inline, no body) → description **suppressed** (omitted from the schema), one
   `scan.empty-override` warning emitted. A second decl with `swagger:description`
   whose only body is a blank line → same. Assert the warning via the coverage
   test; golden shows the godoc text **gone**.

Coverage test `internal/integration/coverage_description_title_override_test.go`
following the shared-parameters pattern: golden assertion + explicit diagnostic
assertions for cases 6 and 8.

---

## 6. Phasing (P1–P5)

- ✅ **P1 — grammar registration (§2)** (`bc10ad7`, 2026-06-25). `AnnTitle`/
  `AnnDescription`, rest-of-line arg (verbatim `TokenRawValue`, keeps trailing
  `.`), scanner closed-set. Single-line only. Grammar + scanner unit tests.
  *Refinement landed in P2:* the empty-override warning is emitted at the
  builder consumption point, not the parser (sibling classifier blocks are not
  `Walk`-ed, so a grammar-stored diagnostic would not reach `OnDiagnostic`); the
  grammar treats a bare override as well-formed.
- ✅ **P2 — schema consumption** (`883e835`, 2026-06-25). `schema.overridesFor`
  harvests the overrides from sibling blocks; applied on the model decl and the
  inline field path. `$ref` fields: title/description are symmetric siblings
  riding description's fate (kept under EmitRefSiblings / forced compound,
  dropped to bare `$ref` under default flags), threaded through
  `applyToRefField`/`applyRefSiblingDrop`. Empty override → `scan.empty-override`
  + suppression. Fixture `enhancements/description-title-override` + two goldens
  (default + EmitRefSiblings) cover cases 1, 2, 5(single-line), 7, 8, and the
  bare-$ref drop. Full suite green, lint clean.
- ✅ **P3 — response/header consumption + context error** (`26dec7a`).
  `responses.overriddenDescription` applies the description override on the
  response decl + each header, rejects `swagger:title` with
  `parse.context-invalid`. Shared harvest lifted to `common.Builder`
  (`HarvestOverrides`/`WarnEmptyOverride`/`OverrideValue`). Fixture
  `description-title-override-responses` + golden.
- ✅ **P4 — multi-line body (Option B) + keyword co-location** (`5f87d0e`).
  Lexer `collectDescriptionBody` folds prose lines after the annotation
  (blank/keyword/annotation/EOF terminated, `\n`-joined). **Family change:**
  `title`/`description` moved to the **schema** family (like `swagger:name`) so
  co-located validation keywords on the same field survive (Fred's call,
  2026-06-25 — see §3.2). Witnessed: `notes` (multi-line) + `capacity`
  (override + `maximum:`).
- ✅ **P5 — README + close-out** (this commit). `schema/README.md`
  §user-overrides documents the overrides; `scanner/README.md` classifier vocab
  lists `title`/`description`. Backlog stub flipped to done.

One commit per phase on `feat/feature-v0.36`, full suite green + lint clean
(`--new-from-rev master`) each, no merge before review (per
[[feedback_wait_for_review_before_merge]]).

---

## 7. Review status

**All design questions resolved (Fred 2026-06-25).** §3 fork → Option B (D8);
terminator → blank line (§3.1); empty override → wipe + warn (D7); codes →
`scan.empty-override` + `CodeContextInvalid` (§4); scope → schemas +
responses/headers, parameters & operations deferred (§1). **Cleared for
implementation — P1.**
