# Plan — Reference & Tutorial doc-site

Branch: `doc-site-reference` · One branch, land all at once.
Builds on the scaffold merged in `f09fa79` (see `[[project_doc_site_branding]]`,
`[[project_doc_site_reference_migration]]`).

Status legend: ✅ done · 🚧 in progress · ⬜ todo · 🟡 needs decision · ❓ open question

---

## 1. Goal

Fold the four heavy reference docs at repo-root `docs/` (`annotations.md`,
`keywords.md`, `sub-languages.md`, `grammar.md`) into the Hugo site, and pair
them with an author-friendly, example-first learning path — so the same facts
are reachable two ways: *exemplified* (Tutorials, by spec concept) and
*lecturing* (Maintainers, the normative compendium), tied together by a
single annotation **quick index**.

Decided (2026-06-12):

- **Maintainers** holds the *complete reference compendium* — all four docs,
  verbatim-migrated and cross-linked. The "lecturing" voice.
- **Tutorials** are the author's primary home, organized **by spec concept**
  (not one-page-per-annotation): *model definitions*, *routes & operations*,
  *validations*, *examples & defaults*, *document metadata*. Each section shows
  **Go annotation + resulting JSON side-by-side** ("annotation in → spec
  concept out").
- **Quick index** = one table recapping every annotation, each row linking to
  *both* its Tutorial (exemplified) and its Maintainers (lecturing) entry.
- **Shaping the output** = how-to section for the rendering knobs
  (`$ref`/alias/`allOf`/`x-nullable`/extensions/overlay), each with a
  **before/after golden pair**.

Everything author-facing stays test-backed: Go lives in the `docs/examples`
module, JSON is golden output a test regenerates — same honesty contract as
`basic-scan` today.

---

## 2. Final information architecture

Promote the current `usage/` umbrella to **top-level sections** (the site is now
reference-heavy; flat sections read better and match the testify/go-openapi
shape). New sidebar order:

```
/ (home)
├── Getting started        (w10)  — existing; install + usage-as-a-library
├── Tutorials              (w20)  — NEW · author home, by spec concept
│   ├── _index             — intro + how to read the side-by-side panes
│   ├── Model definitions  — swagger:model/strfmt/enum/allOf/alias/type/name/ignore/file
│   ├── Routes & operations— swagger:route/operation/parameters/response
│   ├── Validations        — keyword surface on fields (numeric/length/format/enum…)
│   ├── Examples & defaults— example:/default: rendering
│   └── Document metadata   — swagger:meta → info/host/basePath/schemes/consumes…
├── Shaping the output     (w30)  — NEW · how-to, rendering knobs (before/after)
│   ├── _index
│   ├── Alias rendering        — expand vs RefAliases vs TransparentAliases
│   ├── Descriptions beside a $ref — allOf / DescWithRef
│   ├── Nullable pointers      — SetXNullableForPointers / x-nullable
│   ├── Vendor extensions      — SkipExtensions / x-go-*
│   ├── Overlaying a spec      — InputSpec
│   └── Scoping the scan       — Include/Exclude, *Tags, BuildTags, ExcludeDeps
├── Annotation index       (w40)  — NEW · the quick-index table (light landing)
├── Maintainers            (w50)  — NEW · the complete reference compendium
│   ├── _index             — audience note + map of the four docs
│   ├── Annotations        ← docs/annotations.md
│   ├── Keywords           ← docs/keywords.md
│   ├── Sub-languages      ← docs/sub-languages.md
│   └── Grammar            ← docs/grammar.md
└── Project                (w60)  — existing
```

Migration of the existing `usage/*` content:

- `usage/getting-started/*` → top-level `getting-started/` (unchanged content).
- `usage/examples/basic-scan.md` → becomes the capstone of **Tutorials** (or a
  "Putting it together" page); the `usage/examples/` section dissolves into
  Tutorials.
- `usage/reference/_index.md` (the current link-out + TODO) → **deleted**,
  replaced by the Annotation index + Maintainers section.
- Home `_index.md` cards rewritten to point at the new sections.

🟡 Decision point: keep `Annotation index` as its own top-level section, or make
it the `_index.md` (landing) of **Tutorials**. Leaning standalone so it reads as
"the reference entry point" for someone who already knows what they want.

---

## 3. Doc-site mechanics to build

### 3.1 Side-by-side `example` shortcode 🚧
Relearn ships `tabs`/`tab` (tabbed) but no two-column primitive. Build
`hack/doc-site/hugo/layouts/shortcodes/example.html`: a responsive 2-pane grid,
left = Go snippet, right = JSON output, that **delegates to the existing `code`
shortcode** for each pane (so region/lines/Full-source links keep working).
Stacks vertically on narrow screens. Sketch:

```
{{< example >}}
  {{< code file="concepts/models/pet.go" lang="go" region="model" >}}
  {{< code file="concepts/models/testdata/pet.json" lang="json" >}}
{{< /example >}}
```

(or a param form `{{< example go="…" region="…" json="…" >}}` — pick whichever
nests cleanly in Hugo; nested shortcodes need `{{< >}}` inner calls). A tiny CSS
rule lands in `themes/codescan-assets/`.

🟡 Alternative: use Relearn `{{< tabs >}}` (Go | JSON tabs) instead of literal
columns. Fred asked for side-by-side, so columns is the target; tabs is the
fallback if column layout fights the theme.

### 3.2 Concept example packages in `docs/examples` ⬜
Mirror the `basic/` pattern: small, pedagogical, test-covered packages whose
annotations exercise one spec concept each. Each package ships a test that scans
it and writes a golden JSON under `testdata/` (honoring `UPDATE_GOLDEN=1`).

```
docs/examples/concepts/
├── models/      — Pet/strfmt/enum/allOf/alias/type/name/ignore/file  → testdata/*.json
├── routes/      — route + operation + parameters + response          → testdata/*.json
├── validations/ — numeric/length/format/enum/default on fields       → testdata/*.json
├── examples/    — example:/default: surfaces                         → testdata/*.json
└── meta/        — swagger:meta package                               → testdata/meta.json
```

These feed the Tutorials panes. Keep each example MINIMAL — pedagogy over
coverage; the big corpus lives in `fixtures/`.

### 3.3 Rendering-knob golden pairs in `docs/examples` ⬜
For **Shaping the output**, one fixture scanned under different `Options` to
produce paired goldens — exactly the shape of the internal matrix tests
(`TestParamsParser_OptionVariants`, `TestParseResponses_OptionVariants`,
`TestEmbeddedDescriptionAndTags_OptionVariants`).

```
docs/examples/shaping/
├── aliases/     — one package, scanned 3×: default / RefAliases / TransparentAliases
├── descref/     — scanned 2×: DescWithRef off / on
├── pointers/    — scanned 2×: SetXNullableForPointers off / on
├── extensions/  — scanned 2×: SkipExtensions off / on
└── overlay/     — InputSpec merge demo
```

Test writes e.g. `testdata/aliases_default.json`, `aliases_ref.json`,
`aliases_transparent.json`; the how-to page renders the pair(s) with the
`example`/`code` shortcode so the diff is real and CI-guarded.

Reference (do NOT copy wholesale; build minimal analogues):
`internal/builders/{parameters,responses,schema}/*_test.go` matrix tests and
`fixtures/integration/golden/{classification_params*,bugs_3125_schema*,enhancements_alias_field_*}.json`.

### 3.4 Capstone (nice-to-have) ⬜
Use Relearn `{{< openapi >}}` to render the assembled petstore spec as live
Swagger-UI on the Tutorials capstone page. Optional; gated on it not bloating
the build.

---

## 4. Content migration — source → destination

| Source (`docs/`) | Destination | Work |
|---|---|---|
| `annotations.md` | `maintainers/annotations.md` | move; rewrite `./keywords.md`-style links to Hugo `relref`; keep the compatibility matrix; verify fixture paths still exist |
| `keywords.md` | `maintainers/keywords.md` | move; fix cross-links |
| `sub-languages.md` | `maintainers/sub-languages.md` | move; fix cross-links |
| `grammar.md` | `maintainers/grammar.md` | move; fix cross-links |
| (new) | `maintainers/_index.md` | audience note + map of the four |
| (new) | `annotation-index/_index.md` | the quick-index table (12 annotations × {what · attaches · renders · → tutorial · → reference · fixture}) |
| `usage/reference/_index.md` | — | delete (replaced) |

🟡 Decision: do the four docs stay at repo-root `docs/*.md` too (as the
GitHub-readable source of truth) and the site mounts them, or do they MOVE into
`docs/doc-site/maintainers/` and repo-root copies go away? Mounting-in-place
avoids duplication but the current Hugo content mount is `docs/doc-site/` only.
Leaning: **move** them under `docs/doc-site/maintainers/` (single source), drop
the repo-root copies, and point the repo README / CLAUDE.md at the site.

Cross-link rewrites needed inside the migrated docs: every `./annotations.md`,
`./keywords.md`, `./grammar.md`, `./sub-languages.md`, and `#anchor` reference →
Hugo `{{< relref >}}` or site-relative links. Fixture path mentions
(`fixtures/...`) become links to GitHub source.

---

## 5. The quick index (Annotation index) — shape

One row per annotation (12), columns:

| Annotation | Attaches to | Renders as | By example | Full reference |
|---|---|---|---|---|
| `swagger:model` | type decl | a `definitions` entry | → Tutorials/Model definitions#model | → Maintainers/Annotations#swaggermodel |
| … | | | | |

Source of truth for rows: the existing `annotations.md` TOC + its
"How annotations attach" and the compatibility matrix.

---

## 6. Workstreams (build order, one branch)

1. ✅ **Skeleton** — section tree promoted to top-level; `usage/` dissolved;
   home cards rewritten; `usage/reference` deleted. All concept/howto child
   pages stubbed (frontmatter + per-annotation headings so index anchors
   resolve).
2. ✅ **Mechanic** — `example` side-by-side shortcode + shared
   `partials/code-block.html` (refactored out of `code.html`) + `.example-grid`
   CSS in `custom-header.html`. Renders 2 panes, golden JSON in the right pane,
   no marker/nolint leak.
3. ✅ **Maintainers compendium** — four docs moved under `maintainers/`
   (single source; no external refs needed repointing); 34 intra-doc links →
   `{{% relref %}}` percent form with folded anchors; body H1s stripped;
   descriptions added; `_index.md` written.
   → **Checkpoint reached.** Hugo build clean (38 pages, 0 warn/err);
   `go test ./docs/examples/...` green. Navigable site for manual preview.
4. ✅ **Concept examples** — `models`, `routes`, `validations`, `examples`
   (typed coercion verified; `swagger:default` no standalone output),
   `decorators` (readOnly field, deprecated op; field-level `deprecated`
   dropped → quirk F6), `meta` (full info/host/basePath/schemes/consumes/
   license/contact). All golden-checked.
5. ✅ **Tutorials** — `model-definitions`, `routes-and-operations`,
   `validations`, `examples-and-defaults`, `other-type-decorators`,
   `document-metadata`, + `putting-it-together` capstone. All panes
   golden-backed; verify-first throughout (quirks F1–F6 logged, none blocking).

   **Findings from the first slice (validate, don't assume):**
   - codescan emits a named type only when **reachable**; an unreferenced enum
     or alias type is silently absent. The enum example references `Priority`
     from a `Task` model to surface it.
   - `swagger:enum` behaviour vs `maintainers/annotations.md`: the doc says the
     named type gets an `enum` array **as a definition**. Reality: with
     `swagger:enum <Name>` + consts, the enum + `x-go-enum-desc` **inline onto
     the referencing field**, and the type is *not* a standalone definition
     (matches `fixtures/enhancements/enum-overrides` Case A). Bare
     `swagger:enum` (no name) produced a value-less type. → **the migrated
     annotations.md swagger:enum section needs a correction pass.**
   - alias and file deferred out of model-definitions: alias → Shaping
     (`alias-rendering`), file → routes. Annotation-index rows repointed.
6. ✅ **Shaping examples** — `compare` shortcode added; packages `nullable`
   (omitempty suppresses x-nullable), `extensions` (SkipExtensions), `descref`
   (DescWithRef), `discovery` (reachability rule), `buildtags` (BuildTags),
   `overlay` (InputSpec merge), `aliases` (safe dissolve form — first-class
   alias modes hang, see F9). All golden-checked, no hangs.
7. ✅ **Shaping the output** — all 8: scoping-the-scan, type-discovery,
   nullable-pointers, vendor-extensions, descriptions-beside-a-ref,
   overlaying-a-spec, alias-rendering, build-tags. Quirks F8 (swagger:alias
   no-op) and F9 (swagger:model-alias hang) surfaced & logged.
8. ✅/⬜ **Quick index** — table wired to both destinations (links live;
   tutorial targets are stubs until WS5).
9. ⬜ **Capstone** (optional) — `{{< openapi >}}` rendered spec.
10. ⬜ **Verify** — `go test ./docs/examples/...`, `golangci-lint` clean, Hugo
    build 0 errors, no broken in-site links, `code`/`example` panes render, no
    snippet markers/nolint leak. Then squash into logical commits.

---

## 7. Constraints (carry-over)

- DCO `-s`; author = Frederic BIDON; agents `Co-Authored-By` only.
- SPDX Apache-2.0 headers on every Go file; inline comment on every `//nolint`.
- `golangci-lint fmt` for formatting; `gomoddirectives` already disabled for the
  multi-module replace.
- Tests mandatory for the example packages (they ARE the honesty contract).
- No `go.work` (breaks `./...` scanning).
- Commit/push only when Fred asks.

---

## 8. Decisions (resolved 2026-06-12)

1. ✅ **Annotation index** — standalone top-level section (the reference entry
   point), NOT the Tutorials landing.
2. ✅ **Repo-root docs** — **move** the four `docs/*.md` under
   `docs/doc-site/maintainers/`; delete repo-root copies; repoint repo README /
   CLAUDE.md at the site. Single source of truth.
3. ✅ **Side-by-side panes** — literal two-column `example` shortcode is the
   target; Relearn `tabs` is the fallback only if columns fight the theme.
4. ✅ **Promote `usage/`** — dissolve the umbrella; Getting started / Tutorials /
   Shaping the output / Annotation index / Maintainers / Project become
   top-level sections.

## Behavioral findings — decisions needed (2026-06-13)

Empirically verified against `concepts/models` scans + builder code. Several
contradict expectations / the migrated `annotations.md`:

| # | Behavior (verified) | Doc/decision |
|---|---------------------|--------------|
| F1 | `swagger:strfmt`+`swagger:model`: field still **inlines** `{string,format}`; the type emits an orphan `{type:string}` definition with **format dropped**. No `$ref` variant. | The "$ref variant" Fred asked for doesn't exist. Bug, or document inline-only? |
| F2 | `swagger:type`+`swagger:model`: field **inlines** the overridden type; orphan `{type:string}` definition, unreferenced. No `$ref` variant. | Same question as F1. |
| F3 | `swagger:type` has **no OAS2 validation**: accepts `string/integer/number/boolean/object` + Go builtins; `array`/`file` unsupported; any other value **silently falls through** to the underlying Go type, no warning (`walker_classifiers.go:87`, `resolvers.go:32`). | Fred's "verify a check exists" → there is none. Add a warning/check, or just document? |
| F4 | `swagger:enum` is **inline-only** when consts exist; `+swagger:model` gives an orphan value-less `{string}` definition + inline on field. No `$ref`+definition opt-in. | Confirms `annotations.md` is stale (says enum→definition). Update doc to inline reality. Want an enum-as-$ref opt-in (future)? |
| F5 | `swagger:name` works on **interface methods**; **ignored on struct fields** (every corpus use is a method). | Is struct-field support intended (bug) or is the doc wrong (methods only)? `annotations.md` says "field OR method". |

**Done from Fred's review:** allOf example now 3-member (2 `$ref` + inline arm,
with the *why*); `swagger:name` example reworked to a real interface-method
override (`jsonClass` ≠ default `structType`); `swagger:type` accuracy note
added (F3). **Parked pending decision:** strfmt/type `$ref` variants (F1/F2),
enum doc correction (F4), `annotations.md` `swagger:name` field claim (F5).

**Structural additions landed (stubs):** tutorial *Other type decorators*
(`readOnly`/`deprecated`, weight 45); Shaping *When the scanner emits a type*
(reachability + `swagger:model`, weight 15, 2nd); Shaping *Build tags*
(weight 70, last); *Validations* simple-schema reduced-surface notice.

## Build checkpoint

Workstreams 1–3 (skeleton + shortcode + migrate the four docs) land a navigable
site with the full reference content — the natural pause for a manual preview
before the example-heavy workstreams 4–8.
