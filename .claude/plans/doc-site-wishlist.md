# Doc-site feature wishlist

Forthcoming improvements for the **codescan doc-site** (the Hugo site under
`docs/doc-site/` + `hack/doc-site/hugo/`, with test-covered examples in
`docs/examples/`). Companions: `archive/doc-site-quirks.md` (scanner bugs the docs
revealed), and the build plan in the doc-site worktree's
`doc-site-reference.md`. Cross-refs to product directions:
`project_wasm_playground`, `project_genspec_tui`, `project_v2_vision`.

Status as of 2026-06-13: reference + tutorials + shaping-the-output complete and
merging. This doc is the **next-wave backlog**, not committed work. Refined with
Fred's review (2026-06-13).

Priority key: 🟢 quick win · 🟡 medium · 🔴 bigger bet. Effort is rough.

---

## A. Interactive rendering

### W1 — OpenAPI UI widget (`{{< openapi >}}`) 🟢→🟡 — ✅ DONE (2026-06-26, doc-features `06ce50f`)
Render the generated spec as live Swagger-UI on full-spec pages — the capstone
(`putting-it-together`) and a dedicated "see it rendered" showcase. Closes the
loop: annotated Go → JSON → **the API docs the annotations produce**.

**Landed:** "Rendered" tab on the capstone, feeding the petstore golden
(`examples/basic/testdata/swagger.json`) to the Relearn `openapi` shortcode via
an always-visible "Seeing it rendered" section (NOT a tab — Swagger UI can't
paint inside a `display:none` tab panel). Swagger UI is theme-bundled
(offline-clean); the widget reads the same UPDATE_GOLDEN golden as the JSON block
so it can't drift. Note: openapi `src` resolves against the global assets root,
so it needs the `examples/` prefix (unlike the `code` shortcode). The dedicated
showcase spec is still open if a second, larger curated spec is wanted later.

**Follow-up:** a forward-looking **tabbed-example** pattern (Source / Spec /
Rendered / future WASM Playground) is designed in `tabbed-examples.md` — it
revisits the tab presentation with a re-render shim so the Rendered widget works
in a tab, and reserves a Playground slot for `W11`.

- Relearn provides the shortcode; we provide the spec as a fetchable asset
  (mount a full golden into `static/`, or feed the shortcode a path).
- **Constraint:** only valid for *whole* specs. Our tutorial goldens are
  fragments (one definition / one path) and won't render standalone → widget
  on the capstone + one curated showcase spec, NOT on every tutorial.
- Heavy JS → lazy-load.
- First step: wire the petstore (`basic/testdata/swagger.json`) onto the
  capstone behind a "Rendered" tab.

### W11 — WASM live playground 🔴
A "Try it" editor: type annotated Go in the browser, see the spec live
(+ the W1 widget). codescan compiled to WASM; the doc-site is the host. This is
the documented `project_wasm_playground` direction — phased behind the
genspec-tui work. Biggest adoption lever; largest effort.

### W10 — Annotation → spec cross-highlight 🔴 (unlocked by W11)
Hover/click a Go line, highlight the spec node it produced (and vice versa).
**Already a feature of the TUI** (`project_genspec_tui`); the path is: backport
the TUI cross-ref to the WASM build (W11), then the doc-site surfaces it through
the playground widget. Not a standalone doc-site effort — it rides on W11.

---

## B. Content expansion

### W6 — Advanced modeling tutorial 🟡
Discriminator / polymorphism, nested `allOf`, maps / `additionalProperties`,
recursive types, embedded interfaces. The current Model definitions page is
deliberately introductory; this is the deep end. (Fred: discriminator is the
clear gap to add.)

### W7 — Document & security: prefer overlay 🟡
Cover the full `swagger:meta` + security surface (`securityDefinitions`,
security requirements, `externalDocs`, `tos`, `infoExtensions`, top-level
`extensions`) — but **lead with the recommendation to overlay a hand-authored
base spec (`InputSpec`) rather than annotate this in Go**. Document/security
metadata is almost always simpler to maintain as a base document the scan merges
onto (see Shaping → Overlaying a spec) than imposed on the source. The page
should make that the default advice and treat the in-code annotations as the
fallback.

### W8 — CLI / TUI getting-started siblings 🟡 (blocked on tooling)
Getting started anticipated siblings to "Usage as a library". Add them as the
tools ship:
- "Usage from the CLI" — when a codescan CLI lands.
- "Usage with the TUI" — `genspec-tui` (`project_genspec_tui`, status WIP).
- "Usage with `go:generate`".

### W16 — "About / why codescan" explainer 🟢 (source ready)
The one Diátaxis *explanation*-tier page we lack. Positions codescan and orients
go-swagger users.

**Framing** (anchored on go-swagger's `docs/about.md`, the design-first vs
code-first primer):
- Two approaches to API dev: **design-first** (contract-first; go-swagger
  generates server/client from a spec) and **code-first** (annotate Go, scan to
  a spec). codescan is the **code-first engine**.
- Relationship: codescan was extracted from go-swagger; it is the scanner
  **behind `swagger generate spec ./...`**. go-swagger remains the main CLI
  consumer. This doc-site documents the *library/scanner*; it sits **upstream of**
  go-swagger's "generate spec" section (which predates this site and will link
  down to it).
- Why scan-from-source: keep the spec in sync with the code, fast code-first
  iteration, document an already-deployed server. When design-first fits better,
  point to go-swagger.
- For go-swagger users: same annotations; you can call the library directly or
  keep using `swagger generate spec`.

**Source to adapt:** `/home/fred/src/github.com/go-swagger/go-swagger/docs/about.md`
(community-toolkit philosophy, the two-approaches section, `swagger generate
spec` example). Keep it short and link out to go-swagger rather than duplicating
the toolkit story.

Bounded and source-ready → good early build.

### W3 — Known limitations page 🟢 (after quirks triage)
A public, honest "known limitations" section once `archive/doc-site-quirks.md` F1–F9 are
triaged/fixed — sets expectations (alias edge cases, OAS2-only surface, etc.).
Gate on the fix branch so we don't advertise bugs that are about to vanish.

### W17 — Syntax pitfalls page 🟢 (source ready)
A focused "gotchas" page on the brittle edges of the annotation syntax — the
places where a spec silently comes out wrong rather than erroring. These bit us
repeatedly while building the examples, so the material is real and test-backed.

Cover at least:
- **Indentation rules in `swagger:operation` YAML bodies.** The body after the
  `---` is YAML, so indentation is significant — and **`gofmt` re-indents doc
  comments with tabs**, which the scanner's `yaml.RemoveIndent` then has to
  expand before stripping (the bug fixed in `project_gofmt_yaml_tab_indent`).
  Show the safe idiom and what a mangled-nesting failure looks like.
- **The prose-token footgun.** A comment line that *starts* with a `swagger:<name>`
  token, or a keyword like `name:` / `example:` / `patternProperties:`, is parsed
  as that annotation/keyword **even in prose** — it silently truncates the
  description or misfires the directive. Rule: keep such tokens mid-line or
  backtick-wrapped. (Bit us in `maps.go` and the `shaping/formats` example.)
- **List separators & bracket forms.** Comma-separated `enum`, the bracketed
  `[a,b,c]` enum form (brackets stripped), security AND (several keys in one
  requirement) vs OR (separate list items), and the YAML dash-list form for
  `Security:`. Easy to get the wrong semantics from the wrong separator.
- **Title vs description split.** A single-line comment ending in punctuation
  becomes `title`/`summary`, not `description` — and the
  `SingleLineCommentAsDescription` knob that opts out (already documented in
  Shaping → Single-line comments; link to it).

Sources ready: `archive/doc-site-quirks.md`, the boundary memory's footgun notes, and
the existing test-covered witnesses (`shaping/singleline`, `concepts/maps`,
`concepts/security`). Bounded, high-value for authors → good early build.

### (dropped) generic recipes / cookbook
Considered and **dropped**: for a surface this small the concept tutorials
already carry real end-to-end examples, so a cookbook would mostly duplicate
them. The only defensible addition is *composition* (several annotations → one
realistic resource), which overlaps the `putting-it-together` capstone. If a gap
appears, add **one or two extra capstone siblings** for genuinely multi-concept
scenarios — not a standalone cookbook section.

---

## C. Tooling & maintainability

### W15 — Render the grammar visually 🟡
Make `maintainers/grammar` legible at a glance. Two tiers:
- **Prettified EBNF (cheap):** the EBNF already lives in fenced blocks; Chroma
  ships an `ebnf` lexer, so highlighted/prettified EBNF is nearly free — tag the
  fences (and confirm the lexer renders our ISO-14977 dialect acceptably).
- **Railroad / syntax diagrams (nicer, more effort):** generate railroad SVGs
  from the productions at build time and embed them. NB mermaid has **no** native
  railroad support (flowcharts only), so this needs a railroad-diagram generator,
  not mermaid. **⏸ Parked (2026-06-23)** — low ROI vs the EBNF tier; only the
  prettified-EBNF tier (W15a) is in the active next wave.

### W4 — Pipeline mermaid diagram 🟢
A mermaid diagram on the Maintainers landing: scanner → parsers/grammar →
builders → `*spec.Swagger`. Relearn supports mermaid out of the box; turns the
prose architecture into a picture. (Mermaid fits here — a flowchart — even
though it can't do the W15 railroad diagrams.)

### W2 — Copy-to-clipboard on panes 🟢
Verify Relearn's built-in clipboard affordance covers our `example`/`compare`
panes; add it if the custom shortcodes bypass it.

### W14 — Example-coverage check 🟢
A test/CI step asserting every annotation in the grammar has at least one
example package + a tutorial anchor — so new annotations can't ship
undocumented. Natural extension of the golden harness.

### W9 — Generate the reference tables from the grammar 🔴 (⏸ parked 2026-06-23)
**⏸ Parked as unrealistic** — large lexer/parser-interaction digression; W14
(coverage gate) covers the cheap part of the same goal. Kept here for the record,
not in the active Stream 7 next wave.

The keyword / annotation reference is hand-maintained and **already drifted once**
(the `swagger:enum` section was wrong, caught only by a real example). The dream
is to emit those tables from the grammar's keyword tables so the reference can't
lie. **But** the lexer/parser-interaction complexity makes this a large
digression (cf. the highly-complex source-generated parts of the
go-openapi/testify doc-site). Park it as the icing on the cake, well after the
stream above; W14 (coverage gate) covers the cheap part of the same goal.

---

## D. Lifecycle

### W13 — OpenAPI 3.x coverage 🔴 (tracks v2 vision)
When OAS 3.x output lands (`project_v2_vision`), the doc-site covers both
dialects and the differences (SimpleSchema disappears, nullable becomes native,
`example` vs `examples`, etc.). Large, downstream of the product work.

### (out of scope) Versioned docs
Per-release versioned docs is an **org-wide** project with CI & release-process
ramifications — tracked as its own stream, not part of this doc-site wishlist.

---

## Suggested first wave (when doc-site work resumes)
1. **W16** About / why-codescan explainer (bounded, source ready) + **W1**
   OpenAPI widget on the capstone (highest visible payoff).
2. **W15 (prettified EBNF tier)** + **W4** + **W2** — cheap, high-legibility polish.
3. **W6** discriminator / advanced modeling, **W7** meta+security (overlay-first),
   **W17** syntax pitfalls (bounded, source-ready author-facing gotchas).
4. **W14** example-coverage gate.
Defer **W3** until the F1–F9 fix lands. Marquee bets **W11 → W10** ride the
genspec-tui/WASM stream. **W9** and the W15 railroad tier are later icing.
