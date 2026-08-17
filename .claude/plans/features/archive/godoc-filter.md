---
title: Godoc-syntax filtering & idiom recomposition
stream: 8
origin: i
status: done
release: v0.36
issues: []
prev: "§21 (split)"
---

# Godoc-syntax filtering & idiom recomposition

**Status:** ✅ done (P1–P4) on `feat/feature-v0.36`, awaiting review/merge ·
🔷 v0.36. Shipped as `Options.CleanGoDoc` (opt-in). Reviewable design + build
log: [godoc-filter-design.md](godoc-filter-design.md) (scope **A + B**, opt-in
knob; C deferred to ride the inner-comments feature). Maintainer docs:
`internal/builders/godoclink/README.md` + `internal/scanner/README.md`
§clean-godoc.

**Origin.** Named on the published roadmap (v0.36 "go doc filter") but had no
backing spec; split out of [comment-source-filtering](comment-source-filtering.md)
during grooming (2026-06-23) — that feature is about *where* annotations are
read; this one is about *cleaning the godoc text* extracted into the spec.

**Scope.** When a title/description is extracted **from godoc**, handle
godoc-specific syntax that doesn't belong in generated spec documentation:

- **Filter godoc doc-link syntax** — e.g. `[ident]` doc links — which would
  render as noise in spec documentation.
- **Recompose godoc idioms** — e.g. a sentence like "``[Ident]`` does this…"
  where `Ident` is actually exposed under an **overridden / jsonified** name:
  rewrite the reference to the exposed name (or strip the bracket linkification)
  so the spec description reads correctly against the emitted schema.
- *(Adjacent)* inner-markdown handling — preserving/normalising markdown carried
  from godoc into `description` fields.

**Open decision.** Whether this is default-on or behind a knob. Filtering bare
`[ident]` link noise is plausibly safe by default; idiom *recomposition* (which
needs the exposed-name mapping) is more invasive and may warrant opt-in.

**Interactions.** Pairs with [comment-source-filtering](comment-source-filtering.md)
and [swagger-description-override](swagger-description-override.md) (clean-godoc
cluster).

**When to revisit.** v0.36, with the clean-godoc cluster.
