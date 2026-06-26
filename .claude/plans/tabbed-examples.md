# Tabbed examples — a forward-looking example presentation pattern

How the tutorials present an example, designed to grow a **WASM Playground**
later. Refined with Fred (2026-06-26).

Status: **✅ FEATURE COMPLETE for now (`a770dbd`, `f745536`, `c9e440a`).** Fred
2026-06-26: "good for the SwaggerUI feature." Design settled, golden form = B.
The `example` shortcode carries an opt-in `full=` (right pane → `Spec | SwaggerUI`
tabbed card); the render is multi-widget safe (unique container ids, own
CSS-isolated iframe, theme-matched). Live tab on all route-bearing tutorials
(routes, security, decorators, validations, examples). P3-tail (model-only
tutorials) **declined**; P4 (WASM Playground box-swap) **deferred to W11**.

Priority key: 🟢 quick · 🟡 medium · 🔴 bigger bet.

---

## Scope

- **"Putting it together" capstone stays as-is** — no tabs. Its value is the
  top-down, step-by-step flow (annotate → scan → spec → rendered), so the inline
  W1 widget (`2ebcd31`) stays. It is **not** the general pattern.
- **The pattern applies to the other tutorials' `{{< example >}}` panes** — the
  side-by-side Go-and-JSON examples.

## The shape (agreed)

A **two-card, side-by-side** example. The left card is unchanged; the right card
gains tabs:

```
┌─ left card ────────────┐  ┌─ right card ───────────────────┐
│ Source (Go)            │  │ [ Spec* ] [ SwaggerUI ]        │
│ — always on            │  │  Spec  = the JSON golden        │
│                        │  │  SwaggerUI = live rendered view │
└────────────────────────┘  └────────────────────────────────┘
        (* Spec is the default right tab)
```

- **Left card: Source (Go), always visible.** No tab.
- **Right card: tabbed — `Spec` (default) | `SwaggerUI`.**
- **Default right tab: Spec.**

### Tomorrow — the Playground (W11)

The **entire two-card box becomes collapsible and is replaced by the Playground**
— *not* a third right-card tab. The WASM playground app itself shows source
vs. spec (exactly like the genspec-TUI does today) and injects Swagger-UI
directly, so it subsumes both cards. The shortcode should make the whole box a
single unit that a later flag swaps for the Playground mount. Cross-refs:
[[project_wasm_playground]], [[project_genspec_tui]] (cross-highlight `W10` rides
this).

## The blocker: fragment goldens vs whole-spec Swagger-UI

Swagger-UI needs a **whole** Swagger 2.0 document (`{swagger, info, paths,
definitions}`). But the tutorials' `{{< example json= >}}` goldens are **focused
fragments** — a single definition / schema object (`concepts/models/model.json`
is `{type, properties, …}`, `allof.json` is `{allOf, …}`). Only `concepts/meta`
is already a whole spec. So the `SwaggerUI` tab cannot reuse the fragment the
`Spec` tab shows.

### Decision needed — how to feed the SwaggerUI tab a whole spec

- **B (recommended) — the witness emits a whole-spec golden too.** Each tabbed
  example's test dumps the full `*spec.Swagger` from `codescan.Run` as
  `<name>_full.json`, beside the curated fragment. The `Spec` tab keeps showing
  the readable fragment; the `SwaggerUI` tab renders `<name>_full.json`.
  *Faithful* (it's the real scanner output, UPDATE_GOLDEN-regenerated), no
  synthesis, cheap (the full doc is already in hand — the witness currently just
  extracts a slice of it). Cost: +1 golden per tabbed example.
- A — wrap the fragment in a minimal spec shell *in the shortcode* at build time.
  No new goldens, but fragile (must detect definition-vs-path-vs-items, invent a
  name) and the rendered spec isn't a genuine codescan output (less honest).
- C — SwaggerUI tab only where the scan already yields a whole spec
  (meta/routes); model-only examples keep a Spec-only right card. Honest, but
  most tutorials wouldn't get the tab — contradicts the "Spec | SwaggerUI"
  right card.

**Nuance for B:** a model-only fixture scans to a spec with `paths: {}`, so
Swagger-UI shows a "No operations defined" banner above the rich Models/Schemas
view. Acceptable (it's honest), but it reads best on **route-bearing** examples
(routes-and-operations, security, meta) where operations render. → **pilot on a
route-bearing example first.**

## The Swagger-UI-in-a-tab shim (approved)

Swagger-UI can't paint inside a `display:none` tab panel (iframe needs a laid-out
ancestor → 0 height; `switchTab` re-inits Mermaid but not openapi; init/restore
race). Fred OK'd the shim.

- A small custom script (injected via our already-overridden
  `layouts/partials/custom-header.html`, no theme fork) re-renders an
  `.sc-openapi-container` when its tab is first revealed.
- The theme's `initOpenapi`/`switchTab` are module-scoped (no global hook), so the
  shim re-implements the minimal iframe + `SwaggerUIBundle` build (~30 lines from
  `theme.js`), keyed off the global `SwaggerUIBundle` + the `data-openapi-spec`
  attribute already on the container.
- `Spec` is the default tab, so `SwaggerUI` is *always* revealed lazily through
  the single shim path (also sidesteps the default-tab race).
- **Maintenance:** couples to theme internals; pin the theme version, re-test on
  bump. [[project_doc_site_build_env]].

## Rollout

- ✅ **P1 — Pilot** (`a770dbd`): the `swagger:route` pane of
  `routes-and-operations` rebuilt as left Source + right `Spec | SwaggerUI`;
  `concepts/routes` witness emits `full.json` (B); `examplelive` shortcode +
  custom-header tab CSS/JS (lazy render via an `afterprint` re-trigger — the
  shim shortcut). Fred-approved live ("LGTM"; empty `info` is fine).
- ✅ **P2 — Shared `{{< example >}}` upgrade** (`f745536`): `full=` opt-in on
  `example` turns the right pane into the `Spec | SwaggerUI` tabbed card;
  `examplelive` removed. Render rewritten to be multi-widget safe — each pane
  emits its own unique-id `.el-openapi`, loads the spec by URL, renders into its
  own iframe (links swagger-ui + swagger + format-html CSS to match the theme;
  asset load via the `hasOpenApi` store flag). Two independent live examples on
  the routes page confirm it. `{{< compare >}}` (A/B) untouched.
- ✅ **P3 — Migrate route-bearing tutorials** (`c9e440a`): live tab on one pane
  each of security, other-type-decorators, validations, examples-and-defaults
  (routes already had it); each witness emits `full.json`. sharedparams skipped
  (code/compare, not `example` panes).
  - ❎ **P3-tail — model-only tutorials** (model-definitions, polymorphic-models,
    maps): **declined** (Fred, 2026-06-26) — the "no operations" banner above the
    Schemas view isn't worth it. Route-bearing is the boundary.
- ⬜ **P4 — Playground box-swap** 🔴: make the whole two-card box collapsible and
  swappable for the WASM Playground mount when `W11` lands.

## Settled

- Capstone: inline, no tabs (top-down flow). ✅
- Left card Source always-on; right card tabbed. ✅
- Default right tab: **Spec**. ✅
- Shim appetite: **OK**. ✅
- Playground = whole-box replacement, not a tab. ✅

## Open

- **O1 — golden form for the SwaggerUI tab:** confirm **B** (witness emits a
  `_full` whole-spec golden) over A (shortcode-wraps) / C (only-where-whole-spec).
