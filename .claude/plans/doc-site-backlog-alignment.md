# Plan — doc-site alignment with `fix/backlog-lot1`

Branch under audit: `doc-site-update` (= `fix/backlog-lot1` HEAD, moving target).
Audit date: 2026-06-14. Method: combed all 60 commits `master..fix/backlog-lot1`,
mapped each to its go-swagger issue + the tracker's **Need doc** (📖) flag, then
verified the doc-site against the fixture/golden each commit locked in.

Status legend: ✅ applied · 🟦 deferred (needs backing example + golden) · ⬜ todo · 💤 optional

---

## A. Prose-only corrections (no golden regeneration needed) — applied

These fix factual errors or document existing features in the **reference**
(`maintainers/*`) and **tutorial** prose. Verified against code + goldens.

| # | Issue | File | Fix | Status |
|---|-------|------|-----|--------|
| A1 | 3138 | `tutorials/other-type-decorators.md` | Deprecated notice was **factually wrong** ("model field deprecated produces no `x-deprecated`"). Golden `bugs_3138` emits `x-deprecated:true` on models+fields, incl. via godoc `Deprecated:`. Rewrote notice. | ✅ |
| A2 | 3138 | `maintainers/keywords.md` (`deprecated`) | State model/field `deprecated:` (and a godoc `Deprecated:` paragraph) emit `x-deprecated:true`, even under `SkipExtensions`. | ✅ |
| A3 | 2599/3069 | `maintainers/annotations.md` (`swagger:type`) | Argument-shape list wrongly included `array`. Code rejects `array` (walker_classifiers.go:95,191 → falls through). Removed it; matches tutorial. | ✅ |
| A4 | 2922 | `maintainers/annotations.md` (`swagger:enum`) | JSON sample omitted the `description`-folded const mapping (misleading — that folding is the whole issue). Added it + `SkipEnumDescriptions` note. | ✅ |
| A5 | 2922 | `getting-started/usage-as-a-library.md` | New `SkipEnumDescriptions` option missing from Options table. Added. | ✅ |
| A6 | 2922 | `shaping-the-output/vendor-extensions.md` | Document `SkipEnumDescriptions` (independent of `SkipExtensions`). | ✅ |
| A7 | 2922 | `maintainers/keywords.md` (`enum`) | Note default const→value folding into description + the knob. | ✅ |
| A8 | 2687/3007 | `maintainers/sub-languages.md` | "Comment-marker noise stripping" never said WHICH markers drop. Added "Tool-directive markers are dropped" (Go directives + `+kubebuilder`/`+marker`). Closes 📖 3007. | ✅ |
| A9 | 619/2898 | `maintainers/annotations.md` | `interface{}`/`[]any` → empty (any) schema by design; field with no `in:`/`body` → header. | ✅ |
| A10 | 874 | `maintainers/annotations.md` (`swagger:model`/`response`) | Name must be a plain identifier; dotted `pkg.Type` warns + drops; cross-pkg resolves automatically. | ✅ |
| A11 | 2761 | `maintainers/annotations.md` (`swagger:allOf`) | allOf inside a response body emits `$ref` only when the embedded base is a `swagger:model`. | ✅ |
| A12 | 2912 | `maintainers/annotations.md` (`swagger:parameters`) | Param name comes from `json:` tag (fallback Go field); `form:` tag not consulted. | ✅ |
| A13 | 2804 | `tutorials/validations.md` | Map field has no simple-schema form → skipped with `validate.unsupported-in-simple-schema`; only valid on body schema. | ✅ |
| A14 | 3213 | `maintainers/annotations.md` (How annotations attach) | Grouped `type ( … )` decls: each spec's comment honoured independently. | ✅ |
| A15 | 3214 | `maintainers/keywords.md` (`enum`) | Referenced named primitive: `enum:` in its doc parses to enum values; prose → title/desc, not swallowed. | ✅ |
| A16 | 301 | `tutorials/model-definitions.md` | Cross-ref: an unreferenced `swagger:model` needs `ScanModels`/`-m`. | ✅ |
| A17 | 334 | `maintainers/annotations.md` (How annotations attach) | `swagger:route`/`operation` detected in func body; model title/desc need prose-before-annotation. | ✅ |
| A18 | 2801/2802/413 | `shaping-the-output/type-discovery.md` | Generic structs + external embedding resolution note. | ✅ |
| A19 | 2746 | `maintainers/annotations.md` (`swagger:strfmt`) | Format propagates to slice items (`[]MAC` → items format). | ✅ |

## B. Exemplification — new-feature panes (test-backed example regions + goldens) — applied

All backed by new regions in the `docs/examples` module, golden-regenerated
(`UPDATE_GOLDEN=1`), wired with the `{{< example >}}` shortcode. `go test
./docs/examples/...` green, `golangci-lint --new-from-rev master` clean, Hugo
build 52 pages 0 warn/err.

| # | Issue | What landed | Status |
|---|-------|-------------|--------|
| B1 | 2985 | `concepts/validations` region `object` (`Attributes` map: minProperties/maxProperties/patternProperties) + `## On an object` pane in `validations.md`. | ✅ |
| B2 | 2599 | `concepts/models` region `typefield` (`Coupon.Code` field-level `swagger:type string`) + pane in `model-definitions.md`. | ✅ |
| B3 | 2872 + a431dea | meta `ExternalDocs:` block in `concepts/meta` (shows in meta pane); `concepts/routes` region `externaldocs` (operation + `CatalogEntry` schema) + `## externalDocs` section in `routes-and-operations.md`; prose in `document-metadata.md`. | ✅ |
| B4 | 3138 | `concepts/decorators` region `deprecatedmodel` (`Gadget` — model + field via godoc `Deprecated:` → `x-deprecated`) + pane in `other-type-decorators.md`. | ✅ |
| B5 | 3035/3013/2652/958 | `concepts/examples` regions `reffield` (`Price.Unit` defined-type → allOf override arm) + `responseexample` (`NTPServers` array response example) + two panes in `examples-and-defaults.md`. | ✅ |

**Decision (B4 / gocritic) — confirmed by maintainer:** the godoc `Deprecated:`
paragraph is a deliberate **synonym** for `deprecated: true`, supported in any
context *precisely so* a declaration doc comment can mark deprecation without a
bare `// deprecated: true` line (which gocritic `deprecatedComment` reads as a
malformed godoc marker, and which cannot be `//nolint`-suppressed since the
issue lands on the comment line). In indented route/operation bodies
`deprecated: true` is a free-floating/indented line and gocritic stays quiet.
So: `deprecated: true` in operation bodies, `Deprecated:` paragraph on model/
field doc comments — same result. Docs (`other-type-decorators.md`,
`keywords.md`) now state this synonym + idiom explicitly.

**Note:** `docs/examples/go.mod` tidied — `spec` v0.22.5 → v0.22.6 (required by
the current root: commit b0ad37d's `x-go-enum-desc` on response headers needs
spec ≥ v0.22.6) + indirect bumps.

## C. Optional / low priority — 💤

- 361/2625: generic-`interface{}`-envelope worked example in routes tutorial.
- 3119: YAML-body indentation hazard notice (generic YAML caveat).
- 618: one-line "bare slice carries no in/type" note (arguably already implied).
- 1109: note the definitions-fallback requires `ScanModels`.

## E. Second pass — new commits `58130ef..0dd122b` (34 commits) — applied

Rebased `doc-site-update` onto the advanced `fix/backlog-lot1` (clean, no
conflicts; example tests stayed green — no golden drift from the new engine
commits). Then the same alignment pass on the new range.

**New feature exemplified — swagger Tags (2655, closes 📖 1121):**
- meta `Tags:` region added to `concepts/meta/doc.go` (name/description/nested
  externalDocs/`x-*` per tag) → top-level `tags`; `## Tags` section in
  `document-metadata.md`; `tags` keyword entry + summary row in `keywords.md`;
  added to `swagger:meta` and `swagger:route` legal-keyword lists in
  `annotations.md` (route/operation body `tags:` list, unioned with header tags).
- **externalDocs completion (2655):** field-level + `$ref`'d-field (allOf-lift)
  exemplified by extending the `concepts/routes` `externaldocs` region
  (`CatalogEntry.Vendor` plain field, `.Supplier` $ref'd field); prose in
  `routes-and-operations.md` + `keywords.md`; fixed stale keywords.md line
  ("tags do not yet carry externalDocs" — now false).

**New capability exemplified — structured examples (1542/1268):** `concepts/examples`
region `complexexample` (`Profile`: JSON-object example on a map field, JSON-array
on a slice) + pane + coercion notice in `examples-and-defaults.md` (JSON literals
parse; a bare comma-list stays a string).

**Prose corrections:**
- 1118 — model title/description heuristic (period→title, no-period→description,
  multiline→split) in `annotations.md` swagger:model Ordering note.
- 1078 — `time.Time` → `{string, date-time}` auto-resolution (model-definitions.md).
- 1512 — field-level `swagger:strfmt int64` numeric override + the intentional
  Go-specific `uint64`/`uint32` formats (annotations.md swagger:strfmt).
- 1279 — inline `Parameters:` block on `swagger:route` (no struct needed)
  (routes-and-operations.md).
- 2701 — embedded-field `in:`/`required:` annotations propagate to promoted
  members; + 1133/1174 unsupported-type warn-and-continue (type-discovery.md).
- 989 — `swagger:route` `Responses:` is a line-based sub-language; a nested
  `description:` continuation is dropped, use same-line `description:` or a
  `swagger:operation` YAML body (sub-languages.md).

Verified OK (no change): 1391, 1398, 1079, 2417, 1063, 1117, and the internal
test-locks (1416, 1115, 1096, 1560, 1587, 1613, 1520). Validation: `go test
./docs/examples/...` green, `golangci-lint --new-from-rev master` 0 issues,
gofmt clean, Hugo 52 pages 0 warn/err, new panes confirmed in rendered HTML.

**Boundary advanced to `0dd122b`** (see memory `project_doc_site_backlog_boundary`).

## F. New Shaping pages (author request, 2026-06-15)

Two annotation-driven shaping techniques had reference-only coverage and no
worked example — added as test-backed Shaping pages (weights 20, 25, grouped
with type-discovery 15):

- **Inline response bodies** (`shaping/inlineresponses` pkg + page): the `body:`
  responses sub-language (`body:string`, `body:[]Pet`, `body:Pet`, trailing
  words → description) declares responses with no `swagger:response` struct.
  Golden = the GET /pets path item. (Reworded prose to avoid a literal
  `swagger:response` token in the comment — it was being classified and
  truncating the operation description.)
- **Forcing a conformant format** (`shaping/formats` pkg + page): field-level
  `swagger:strfmt int64` overrides the Go-derived vendor `uint64` format
  (`{integer, uint64}`) to a precision-safe `{string, int64}`. Golden contrasts
  a default `Raw` field with an overridden `Bounded` field.

Hugo now 54 pages; both packages lint/gofmt clean, golden-verified, panes render.

## G. New tutorial — Polymorphic models (author request, 2026-06-15)

The discriminator/polymorphism use case had only a terse `discriminator` keyword
entry and no tutorial. Added `tutorials/polymorphic-models.md` (weight 15, after
model-definitions) backed by a new `concepts/polymorphism` package:

- base `Pet` with a `discriminator: true` + `required` field → `discriminator:
  "petType"` on the schema; subtypes `Cat`/`Dog` compose via `swagger:allOf`
  (`allOf: [$ref Pet, inline]`). Goldens base/subtype/subtype2.
- Documented the real constraints (verified by probe): the discriminator
  property must be `required`; the discriminator **value** is the subtype's
  definition name — `swagger:discriminatorValue` is **not implemented** (it sits
  in a `TODO` block in the classification fixture). Enriched the `discriminator`
  keyword entry and cross-linked from model-definitions `swagger:allOf`.

Hugo now 55 pages; package lint/gofmt clean, golden-verified, page renders.

## H. New subsection — Decorating a $ref (author request, 2026-06-15)

The `$ref` sibling-override / allOf-lift mechanism was scattered (DescWithRef
shaping page, example/default on defined-type fields, externalDocs lift). Added a
consolidated `## Decorating a $ref` subsection to `model-definitions.md`, backed
by a new `concepts/refoverride` package. Behavior probed and documented exactly:

- a bare `$ref` cannot carry siblings; codescan wraps it as an `allOf` member so
  the property can keep them. Verified landing spots: `description` + `x-*`
  extensions on the property; a value override (`default`/`example`) in a second
  `allOf` member; `required` on the parent. Description-only stays governed by
  `DescWithRef`. Golden `refoverride.json` (Person.home wrapped in single-arm
  allOf, keeping description + x-ui-order). Cross-linked to descriptions-beside-
  a-ref and examples-and-defaults.

## I. Third pass — new commits `0dd122b..b4ce007` (16 commits) — applied

Rebased `doc-site-update` onto the advanced `fix/backlog-lot1` (clean; example
tests stayed green — no golden drift from engine commits 1635/1088/2638). Tracker
re-read (updated 2026-06-15). Then the alignment pass.

**Self-check resolved:** an agent flagged `SkipEnumDescriptions` as "not on
Options" — FALSE; it's `options.go:56/68`, exposed via `type Options =
scanner.Options`. Pass-1 docs correct.

**New capability exemplified — author `x-*` on parameters/headers (1609, 📖):**
`Extensions:` block works on schema fields, parameters, AND response headers
(the fixture's "headers don't honor it" comment is stale — golden proves they
do). New `paramext` region in `shaping/extensions` (param `x-example` + header
`x-units`, scanned with `SkipExtensions: true` to show author x-* survive) +
"Authoring x-* on parameters and headers" section in `vendor-extensions.md`;
added param/header to the keywords.md extensions scope.

**Prose corrections:**
- 1742 — `swagger:parameters`/`swagger:response` structs bind by operation ID
  across all scanned packages (annotations.md).
- 1635 — an anonymously embedded `in: body` struct IS the body ($ref), not
  promoted headers (annotations.md swagger:response; same for parameters).
- 2638 — a multi-name field group (`R, G, B, A uint8`) emits one property per
  name; a `json:` rename can't apply (annotations.md swagger:model).
- 1088 — array element can't be a `$ref` in simple schema: named-primitive
  expands inline, object element dissolves to `items: {}` + warning (validations.md).
- 1713 — per-media-type response `examples:` via the `swagger:operation` YAML
  body (struct `swagger:response` doesn't yet) (examples-and-defaults.md).
- 1725 — `Version:` is static; set `doc.Info.Version` after `Run` (ldflags) or
  overlay via `InputSpec` (document-metadata.md).
- 1734 — same-name cross-package definition collision is silently merged
  (lossy); workaround `swagger:model <DistinctName>` (model-definitions.md warning).

Verified OK (no change): 1711, 1665, 1737 (covered by "Decorating a $ref"), 1708,
1727 (library usage already documented), 1735 (environmental). Validation: full
`go test ./docs/examples/...` green, `golangci-lint --new-from-rev master` clean,
gofmt clean, Hugo 55 pages 0 warn/err, new pane rendered.

**Boundary advanced to `b4ce007`.**

## J. Fourth pass — new commits `b4ce007..270d4b4` (10 commits) — applied

Rebased onto the advanced `fix/backlog-lot1` (clean; tests green — no drift from
1595/F7). Two genuine exemplification gaps + one prose gap; the rest already
covered.

**New exemplified — security (1795, 📖):** declared auth (basic/apiKey) was
reference-only. Added a `SecurityDefinitions:` + `Security:` block to the meta
example (basic_auth + api_key; global basic_auth requirement) and a `## Security`
section to `document-metadata.md`. The meta pane now shows `securityDefinitions`
+ `security`.

**New exemplified — POST request body (1772, 📖):** no `in: body` *request*
parameter example existed (only response bodies). Added a `bodyparam` region to
`concepts/routes` (`CreatePetParams.Body` → request body `$ref`, on
`POST /pets/import` to avoid disturbing the `/pets` golden) + a pane in
`routes-and-operations.md`.

**Prose:** 1865 — one struct may carry several `swagger:parameters` lines whose
operation-ID lists accumulate (annotations.md).

Verified OK (no change): 1595 (block-comment routes — covered in grammar/sub-languages),
1867 (PATCH in method list), 1852 (204 empty description), 1828 (inline response
description), 1881 (array response), 1887 (file type), 1758/1761/1815 (usage/meta/overlay
already covered), 1891 (grouped decls + func-body routes — ALREADY documented in
annotations.md:82-89 from an earlier pass; agent missed it), F7 (transparent
parser fix). Out-of-scope won't-fix: 1860, 1777 (no doc-site action).

Validation: full `go test ./docs/examples/...` green, `golangci-lint
--new-from-rev master` clean, gofmt clean, Hugo 55 pages 0 warn/err, panes rendered.

**Boundary advanced to `270d4b4`.**

## K. Fifth pass — new commits `270d4b4..40d8af1` (4 commits) — applied

Rebased onto the advanced `fix/backlog-lot1` (clean; tests green). Prose-only
pass (no example packages touched).

- 1726 — markdown `*`/`+` bullets are recognised like the `-` dash form and
  normalised to `- ` (matching gofmt), across prose descriptions, flex-lists,
  and enum bodies. Updated sub-languages.md "Markdown semantics" + flex-list
  "Accepted forms".
- 1934 — a `swagger:model`/`swagger:parameters` on a type **local to a function
  body** is discovered. Broadened the func-body note (annotations.md).
- 1955 — the swagger:operation analog of cross-package params: broadened my
  pass-3 "Across packages" note from "route" to "route or operation".
- 1913 — confirms the Polymorphic models tutorial (interface or struct base with
  `discriminator: true`, allOf subtypes). The issue was "sub-types not
  generated" → added a reachability note (subtypes need `ScanModels`/reference;
  no auto-discovery from the base).

Verified OK (no change): 1931 (scoping-the-scan already documents `./...` for a
whole tree), 1925 (`[]map[string]interface{}` body param — resolved, not
contradicted), 1958 (operation-YAML vendor extension preserved — niche).
Out-of-scope won't-fix: 1960 (property order — ordering policy, no doc action).

Validation: Hugo 55 pages 0 warn/err, content rendered, relref resolves. (No Go
changes → no golden regen; example tests unaffected.)

**Boundary advanced to `40d8af1`.**

## L. New tutorial — Security (author request, 2026-06-15)

The brief `## Security` note in document-metadata (pass 4) was not enough. Added
a dedicated `tutorials/security.md` (weight 55) backed by a new test-covered
`concepts/security` package, with three sections (behaviour probed first):

1. **Declare the schemes** — `swagger:meta` `SecurityDefinitions:` (api_key +
   oauth2 w/ scopes) + a `Security:` document-wide default. Golden `schemes.json`.
2. **Require a scheme on a route** — a route's `Security: oauth2: read, write`
   overrides the default; a route with no `Security:` inherits it. Golden
   `route.json`. (Probe finding: an *empty* `Security:` does NOT emit
   `security: []` / public — the operation just inherits the default; so the
   "public override" case is omitted, not misdocumented.)
3. **Keep security out of your code** — the gateway/mesh case: scan app code
   (`concepts/routes`) with NO security annotations + an `InputSpec` base
   carrying the schemes/requirement → merged secured spec. Golden `overlay.json`.

Trimmed the document-metadata `## Security` note to a pointer to the new tutorial.
Hugo now 56 pages; package lint/gofmt clean, golden-verified, page renders + in nav.
(Boundary unchanged at `40d8af1` — this is an author-requested addition, not a
backlog-alignment pass.)

## M. Sixth pass — new commits `40d8af1..d57ac9f` (12 commits) — quirk-fix reconciliation

Rebased onto `fix/backlog-lot1` (clean — git auto-merged my annotations.md edits
with a fixer's `swagger:alias` rewrite, different sections). **All example tests
stayed green — no golden drift** — so my example packages are behaviourally
unaffected by the F-series fixes; only the prose needed reconciling.

The **F-series quirk fixes** landed (fix/quirks-F-series merged). Updated docs to
match (the F3/F5 fixes did NOT update the doc-site; F8 did):

- **F3 — `swagger:type` reconciled to a uniform inline directive.** It now NEVER
  emits a `$ref`; `integer`/`number`/`boolean` resolve (were silent no-ops);
  `[]T` inlines recursively; **new `inline` value**; `array` deprecated→inline
  +warning; `file` rejected with a diagnostic; an unknown token inlines the Go
  type (or a known type name) with a diagnostic; with `swagger:strfmt` the type
  wins and the format is kept only if compatible. Rewrote annotations.md +
  model-definitions.md (which had said "array/file not accepted; unknown
  ignored" — now wrong). Verified against golden `quirk_type_matrix.json`.
- **F5 — `swagger:name` now honoured on struct fields** (was interface-only).
  The maintainers reference already said "field OR method" (correct); updated
  the model-definitions tutorial's interface-only framing.
- **F8 — `swagger:alias` deprecated** (fixer rewrote the annotations.md section;
  heading anchor changed to `#swaggeralias--deprecated`). Reconciled the
  dangling references: annotation-index row ("$ref to target" → deprecated/no
  effect, fixed anchor) and the alias-rendering how-to (dropped the "per-mode
  contract" cross-link to the now-deprecation-only section; added a deprecation
  note — the page itself is the authority for the RefAliases/TransparentAliases
  modes).
- F7 (operation-YAML gofmt), F9 (type-alias no-hang): no doc change (transparent
  / my examples unaffected). F7 resolves the gofmt-operation-yaml retest.

Also: 1992 (🛑📖 "hide composition") — the readOnly answer; added a sentence to
the readOnly section (server-set fields → readOnly, or distinct models).
Verified OK: 2002/2013/2020 (test-locks). Out of scope (CLI/environmental, not
the library doc-site): 1974/2000/2014.

Validation: example tests green (no Go changes this pass), Hugo 56 pages 0
warn/err, alias anchors resolve, F3/F5 notes render.

**Boundary advanced to `d57ac9f` — fully caught up.**

## N. Seventh pass — new commits `d57ac9f..664a347` (2 commits) — steady state

Rebased onto `fix/backlog-lot1` (clean; no doc-site files touched by fixers;
tests green). The two commits are a **tracker-only docs commit** (664a347:
#2027 embedded-struct — covered by type-discovery; #2053 scan-dir — CLI,
out-of-scope) and a **test-lock commit** (2fb9ed2) that LOCKS behaviours the
doc-site already exemplifies — confirming my recent examples are correct:

- #2062 operation security — covered (keywords table lists security in
  route+operation; Security tutorial). Added a 1-line note that a
  `swagger:operation` YAML body also takes `security:` (schemes stay global meta).
- #2064 body-param example/default — covered by the example/default keywords.
- #2106 field Extensions on scalar+array — covered by the 1609 paramext example.
- #2119 SkipExtensions = "skip x-go-name" — covered by vendor-extensions.

No new gaps; the fixers are now mostly adding test-locks that confirm the doc
examples. Hugo 56 pages 0 warn/err.

**Boundary advanced to `664a347` — fully caught up.**

## O. Full ledger 📖 sweep (2026-06-15) — tracker split into archive/backlog-triaged-ledger.md

Cross-checked all **78** `📖 Need-doc` rows in `archive/backlog-triaged-ledger.md` (the
commit-driven passes A–N could miss a 📖 issue triaged without a commit). 63 are
referenced in this doc; the other 15 broke down as:

- **Already covered** (doc-site, not my tracking doc): 559 (the site itself),
  2626 (title/description heuristic = #1118), 2639 (reachability/ScanModels),
  2663 (POST body-param = #1772), 2632/2838 (swagger:parameters op-IDs / discovery).
- **Out of scope** (CLI/environmental/wont-fix, not the library doc-site): 2633
  (M1 `--work-dir`), 2778 (sudo/perms), 2874 (GOROOT), 2777 (repro), 2871
  (dynamic examples — wont-fix), 2924 (`x-go-type` is a *generate*-side extension,
  not produced by code→spec scanning).
- **Borderline → added a note**: 2963 (replace-directive module — 🐞, behaviour
  unconfirmed; left alone).

Two genuine additions made:
- **1026 (x-logo)** — `infoExtensions` was in the reference but never exemplified.
  Added an `InfoExtensions:` block with `x-logo` to the `concepts/meta` example
  (golden regenerated; lands in `info.extensions`) + a note in document-metadata.
- **3134 (per-version specs)** — added a note to scoping-the-scan: run a scan per
  package tree (`./v1/...`, `./v2/...`); no single-run split.

Result: every actionable 📖 in the ledger is now covered or explicitly out of scope.

## P. Eighth pass — `664a347..64ceb83` (incl. #2547) — more quirk fixes + an edge case

Rebased onto `fix/backlog-lot1` (stash/rebase/pop for an in-flight fix). The
fixer touched annotations.md again (swagger:enum F4 rewrite) — git auto-merged
with my pass-1 enum edits cleanly.

**Example drift (1) — caught a self-inflicted bug:** `shaping/formats`
collapsed (`Measurement` → `{string,int64}`). Cause: line 12 of my *type* doc
comment literally began with `swagger:strfmt int64` (prose). Harmless before,
but the **F1 fix** now (correctly) applies a type-level `swagger:strfmt` +
`swagger:model` to the definition. Reworded so no line starts with the token.
→ produced a new doc edge-case note (below).

**Fixed quirks documented** (all probed):
- **F1/F2 — `swagger:strfmt` / `swagger:type` + `swagger:model` → definition +
  `$ref`** (was orphan inline). Updated annotations.md (strfmt, type) +
  model-definitions tutorial (strfmt, enum).
- **F4 — `swagger:enum` + `swagger:model` → definition + `$ref`; bare
  `swagger:enum` accepted.** The fixer rewrote the annotations.md reference;
  I updated the model-definitions tutorial.
- **#2038 — a json-tagged embed nests** as a `$ref` property instead of
  promoting (type-discovery embedding note).
- **#2547 — string `example:`/`default:` surrounding double quotes are stripped**
  (`example: ""` → empty string) (examples-and-defaults).

**Edge case documented:** any comment line *beginning* with a `swagger:<name>`
token is parsed as that annotation, even as prose — added a warning notice to
"How annotations attach" (the footgun that bit the formats example, and earlier
the externalDocs/inline-responses examples).

Verified covered/out-of-scope (triage + test-locks): 2528/2618/2980/2549/2575/
2520 (covered: doc-site, field descriptions, embedding, cross-pkg/example,
in:header params, swagger:type), 2503/2539/2592/2596/3275 (runtime-validation /
maps / overlay / swagger:type / generic — covered or out of scope).

Validation: full `go test ./docs/examples/...` green, `golangci-lint
--new-from-rev master` clean, gofmt clean, Hugo 56 pages 0 warn/err.

**Boundary advanced to `64ceb83`.**

## Q. Ninth/tenth passes — `f10eab7..048d607` — small edge-case notes

Several rapid increments (fixers pushing fast); rebased through each. No golden
drift, no fixer doc-site changes. Test-locks/tracker-refs mostly confirm existing
coverage. Two small notes added:

- **#2384** — the `pattern` keyword keeps **backslash escapes** (`\n`, `\d`)
  verbatim (not interpreted by the scanner); spelled out in keywords.md (it said
  "verbatim" but the confusion was specifically about escapes).
- **#2403** — a `Security:` requirement accepts the YAML **dash-list** form
  (`- auth0: []`) as well as the flat `auth0: []` form (previously the dash form
  mis-parsed); noted in the Security tutorial.

Covered/out of scope: #2371 (array response=#1881), #2379 (cross-pkg $ref), #2383
(func-local structs=#1934), #2353 (body+path params), #2409 (model-level
Extensions — added in pass O+), #2398 (dup-definition warning — covered by the
#1734 same-name-collision note), #2412/#2441 (file=swagger:file), #2419
(swagger:type), #2407 (array example).

## R. Eleventh pass — `048d607..7bbea9a` + draft scaffold

Rebased through more fixer increments. No golden drift. Highlights:

- **#2479 (the "new fix")** — `Security: []` now emits an explicit empty
  requirement (public opt-out), distinct from omitting the keyword (inherit
  default). This is the case I had to OMIT when first building the Security
  tutorial (it didn't work then). Added a `publicReport` route to the
  `concepts/security` example (golden `public.json` = `security: []`) + a tutorial
  pane/note. (Goldens otherwise unchanged.)
- **#2311** — `swagger:ignore` on a struct field: the reference already covered
  it; added the missing one-liner to the model-definitions tutorial (explicit
  doc-action from triage).
- **Draft scaffold (separate, earlier commit):** `resolving-name-conflicts.md`
  (draft: true) for the upcoming feat/name-identity-cyclic-ref — outline only.

Covered/out of scope: #2246/#2251 (example-as-string / non-string map keys —
wont-fix limitations), #2232 (tags with spaces), #2245/#2286 (response schema /
model-as-response — covered), #2299 (determinism — maintainer fact, golden
harness), and the test-locks.

**Boundary advanced to `7bbea9a`.** (Tip already moved to `4de9680` —
`feat(responses): honor name: on response header fields` — for the next pass.)

## S. Twelfth pass — security-as-YAML revisit + enum/name additions (`7bbea9a..e64c662`)

Rebased; no golden drift. Three landed changes, all probed:

- **#2294 — `Security:` now parses as real YAML** (the bespoke line parser is
  gone). Reinspected the whole Security tutorial: rewrote the syntax prose (it's
  a YAML sequence of requirement objects; **multiple keys in one item = AND**,
  **separate items = OR**; flow/block scopes; `[]` opt-out; legacy bare mapping
  still read as OR-per-key). Converted the `concepts/security` example to the
  idiomatic sequence form and **added an `archiveReport` AND route** (golden
  `and.json` = `[{api_key:[], oauth2:[write]}]`). Rewrote the keywords.md
  `security` entry (was "line of shape schemeName: scope1, scope2").
- **#2396 — bracketed enum form** `enum: [a, b, c]` strips the `[ ]` delimiters
  (equivalent to the unbracketed form). keywords.md `enum`.
- **`name:` keyword** (feat f5845ea params + 4de9680 response headers) — a
  field-doc keyword that sets the parameter name / response-header key,
  overriding json-tag/field-name; the field-doc equivalent of `swagger:name`,
  stripped from the description. New keywords.md `### name` entry + notes in
  annotations.md swagger:parameters / swagger:response.

Validation: full `go test ./docs/examples/...` green, lint clean, Hugo 56 pages.

**Boundary advanced to `e64c662`.**

## D. Verified OK (no change) — accurate & golden-consistent

624, 796, 1109, 2860, 3107, 2875, 2907, 2483, 2899, 2961, 2791 (rewritten
example is valid OAS2), 3100, 2837, 2909, 2917, 2799, 2611, 2959.
