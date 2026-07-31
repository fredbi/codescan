---
title: swagger:description / swagger:title override annotations
stream: —
origin: ii
status: open
release: v0.36
issues: []
prev: "§3.4"
---

# `swagger:description` / `swagger:title` override annotations

**Status:** ✅ done (2026-06-25, `feat/feature-v0.36`, P1–P5). Full design +
fixture harness + build log in
[swagger-description-override-design.md](swagger-description-override-design.md).
Awaiting review before merge.

**Origin.** Q30 close-out (2026-06-04, W3 alias workshop cycle 2). Every type that
surfaces in `definitions` carries its godoc as the `description` — including
stdlib types the user doesn't control (`time.Time`'s 2305-char godoc on monotonic
clocks etc. leaks into any spec that touches it). The behaviour is correct per the
rules; the gap is the **lack of an override affordance**. A core-product
enhancement with no dedicated stream.

**Decision (groomed 2026-06-23).** Annotation-driven, **not** an option. Two new
annotations that **override the lexer's automatic title / description
extraction** for the annotated decl:

- **`swagger:description [raw block]`** — replaces the extracted description. May
  be followed by a multiline comment block (the raw block carries through).
- **`swagger:title [raw block]`** — replaces the extracted title / summary.

(Supersedes the earlier `Options.DescriptionOverrides` map idea — dropped: this
is the annotation form, and there is no separate option-based feature.)

**Residual (out of scope).** Stdlib types reached via embed / field have no
user-controlled decl to annotate, so the annotation can't reach them; not solved
here.

**Interactions.** The annotations can live after the decl via
[comment-source-filtering](comment-source-filtering.md); complements
[godoc-filter](godoc-filter.md). `../archive/observed-quirks.md` Q30 documents the empirical
`time.Time` observation.

**When to revisit.** When spec-noise complaints recur, or alongside the
clean-godoc cluster.
