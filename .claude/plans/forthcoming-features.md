# Forthcoming features — flat catalog

Synthetic list of every feature-level increment. **Detail lives per feature in
`features/<slug>.md`.** Stream hierarchy + release markers: `roadmap.md`
(Overview). This file is the flat lens; `roadmap.md` is the stream-first lens.

## Where we stand (2026 milestones)

Doubles as the recap source for the published roadmap.

- **v0.32 (Mar)** — repo carved out of go-swagger; relint; library setup
  (not environment-sensitive); go1.25+.
- **v0.33 (Apr)** — reduced public interface; type-array for parameters; new
  layered `internal/` package layout.
- **v0.34 (May)** — grammar-based parser: lexer + grammar replaces the
  regexp engine; many parsing quirks fixed.
- **v0.35 (Jun)** — large bug-bash (~200+ go-swagger issues); doc site; all
  validations; parser diagnostics; name-conflict + circular-`$ref` handling.
  Shipped features: prune-unused-models, name-identity disambiguation,
  additionalProperties (control + non-string keys), fail-loud diagnostics,
  externalDocs on non-meta objects, single-line-as-description, emit `x-go-type`.
- **v0.36 (Jul, in progress)** — CLI & TUI; faster / incremental scanner;
  comment-source filtering (go doc filter, private comments, inner markdown).
  **All discrete features landed** on `feat/feature-v0.36` (awaiting review/merge):
  ✅ `NameFromTags` (#2912/#1391); ✅ response-level `examples` by mime (#2871);
  ✅ shared `swagger:parameters`/`swagger:response` (#2632); ✅ `swagger:title`/
  `swagger:description` overrides; ✅ AfterDeclComments (private/inner comments);
  ✅ CleanGoDoc (go doc filter); ✅ inner markdown — `swagger:description |`
  (#3211); ✅ `SkipJSONifyInterfaceMethods`; ✅ `DefaultAllOfForEmbeds`. The
  comment-source-filtering theme is complete. **Outstanding (non-feature):**
  genspec TUI (Stream 5, separate `feat/genspec-tui` branch) + the CLI half;
  faster / incremental scanner (not started).
- **v0.38 (Sep)** — LSP & IDE (prereqs: token-level YAML positions, column
  precision). **v0.39 (Oct)** — OAI v3 (tentative).

**Settled migration-era items** (no longer tracked as features): parity suite
removed at P6 cutover, P5 per-builder catch-up — done with the grammar
migration; `splitCommaList` quote-respect retired as obsolete (Stream M). The
annotation surface-form doc folded into the doc site (`doc-site-wishlist.md`).

## Origins legend

- **(i)** initial design vision
- **(ii)** idea that surfaced mid-build
- **(iii)** triaged go-swagger issue filed as a forthcoming feature
- **(iv)** deferred refinement left on a shipped fix

## Features

Single list, shipped first (achievements in sight), then open. Each: status ·
title → file · stream · origin · issues. Markers: **low** = deferred low priority ·
**TODO** = approach undecided · **verify** = mostly a check of allegedly-fixed work.

- ✅ [Prune unused models under `-m`](features/prune-unused-models.md) · Stream 9 · (iii) · go-swagger#2639 · v0.35
- ✅ [Definition-name auto-disambiguation](features/name-identity-disambiguation.md) · Stream 9 · (iii) · go-swagger#1734 · v0.35
- ✅ [Explicit additionalProperties control](features/additionalproperties-control.md) · Stream 9 · (iii) · go-swagger#2539/#3005 · v0.35
- ✅ [Map additionalProperties for non-string keys](features/map-additionalproperties-keys.md) · Stream 9 · (iii) · go-swagger#2251 · v0.35
- ✅ [Scanner robustness / fail-loud](features/fail-loud-diagnostics.md) · Stream 9 · (iii) · go-swagger#2886/#2874 · v0.35
- ✅ [ExternalDocs on non-meta objects](features/externaldocs-non-meta.md) · Stream 9 · (iii) · go-swagger#2872/#2655 · v0.35
- ✅ [Single-line comment as description](features/single-line-description.md) · Stream 9 · (iii) · go-swagger#2626 · v0.35
- ✅ [Emit `x-go-type` vendor extension](features/emit-x-go-type.md) · Stream 9 · (iii) · go-swagger#2924 · v0.35
- ✅ [Naming from struct tags (`form:`, `schema:`)](features/naming-tags.md) · Stream 9 · (iii) · go-swagger#2912/#1391 · v0.36
- ✅ [Response-level examples by mime](features/response-examples-by-mime.md) · Stream 9 · (iii) · go-swagger#2871 · v0.36
- ✅ [After-declaration annotation comments](features/comment-source-filtering.md) · Stream 8 · (i) · 🔷 v0.36
- ✅ [Godoc-syntax filtering & idiom recomposition](features/godoc-filter.md) · Stream 8 · (i) · 🔷 v0.36
- ✅ [Inner markdown — `swagger:description \|` block scalar](features/inner-markdown.md) · Stream 8 · (iii) · 🔷 v0.36 · go-swagger#3211
- ✅ [Shared `swagger:parameters` / `swagger:response`](features/shared-parameters.md) · Stream 9 · (iii) · 🔷 v0.36 · go-swagger#2632
- ✅ [Discriminator subtype discovery](features/discriminator-subtype-discovery.md) · Stream 9 · (iii) · go-swagger#1913
- ⬜ [Infer `required` from field shape](features/infer-required-from-shape.md) · Stream 9 · (iii) · go-swagger#3275 · **TODO**
- ⬜ [Per-operation field views (projections)](features/per-operation-projections.md) · Stream 9 · (iii) · go-swagger#1992 · **low**
- ⬜ [Name-identity advanced](features/name-identity-advanced.md) · Stream 9 · (iv) · **low**
- ⬜ [`withPatternProperties` auto-inference](features/pattern-properties-inference.md) · Stream 9 · (iv) · **low**
- ⬜ [Token-level YAML positions](features/yaml-token-positions.md) · Stream 11 · (i) · 🔷 v0.38
- ⬜ [Column precision beyond ASCII](features/column-precision-unicode.md) · Stream 11 · (i) · 🔷 v0.38
- ⬜ [Godoc-identifier prefix on `swagger:operation`](features/godoc-identifier-prefix.md) · Stream 10 · (i) · 🔷 v0.39
- ⬜ [Enum richer values](features/enum-richer-values.md) · Stream — · (i) · **TODO**
- ⬜ [Standard formats only (no extended widths)](features/standard-formats-only.md) · Stream — · (ii) · **low**
- ⬜ [Example value coercion (verification)](features/example-values.md) · Stream — · (i) · go-swagger#1268/#2246 · **verify**
- ⬜ [Property-based Block generator](features/property-based-block-generator.md) · Stream 2 · (i)
- ✅ [Skip-jsonify-interfaces opt-out](features/skip-jsonify-interfaces.md) · Stream — · (ii) · 🔷 v0.36
- ✅ [`swagger:description` / `swagger:title` overrides](features/swagger-description-override.md) · Stream — · (ii) · 🔷 v0.36
- ✅ [`DefaultAllOfForEmbeds`](features/default-allof-for-embeds.md) · Stream — · (ii) · 🔷 v0.36
- ⬜ [`DiscoverAliasesAsTypes`](features/discover-aliases-as-types.md) · Stream — · (ii)
- ⬜ [Bullet-list dash preservation](features/bullet-dash-preservation.md) · Stream — · (iv)

_Stream `—` = cross-cutting core enhancement (no dedicated stream; rides
whichever stream touches its seam)._
