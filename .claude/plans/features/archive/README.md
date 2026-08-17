# Shipped features

The 25 feature documents whose feature has landed, moved here 2026-08-17. **Provenance, not status** —
several carry a stale `status:` line in their own frontmatter (`shared-parameters` and
`swagger-description-override` both still say `open`, and the `-design` companions say `design` /
`in-progress`), which is part of why they were moved: the catalog is the truth, not the file.

The live catalog is [`../../forthcoming-features.md`](../../forthcoming-features.md); the stream view is
[`../../roadmap.md`](../../roadmap.md). Both link here for anything shipped.

| Feature | Release | Issue |
|---|---|---|
| `prune-unused-models` | v0.35 | go-swagger#2639 |
| `name-identity-disambiguation` | v0.35 | #1734 |
| `additionalproperties-control` | v0.35 | #2539 / #3005 |
| `map-additionalproperties-keys` | v0.35 | #2251 |
| `fail-loud-diagnostics` | v0.35 | #2886 / #2874 |
| `externaldocs-non-meta` | v0.35 | #2872 / #2655 |
| `single-line-description` | v0.35 | #2626 |
| `emit-x-go-type` | v0.35 | #2924 |
| `naming-tags` | v0.36 | #2912 / #1391 |
| `response-examples-by-mime` | v0.36 | #2871 |
| `shared-parameters` (+ `-fixtures`) | v0.36 | #2632 |
| `inner-markdown` (+ `-design`) | v0.36 | #3211 |
| `comment-source-filtering` (+ `-design`) | v0.36 | — |
| `godoc-filter` (+ `-design`) | v0.36 | — |
| `skip-jsonify-interfaces` | v0.36 | — |
| `swagger-description-override` (+ `-design`) | v0.36 | — |
| `default-allof-for-embeds` | v0.36 | — |
| `discriminator-subtype-discovery` (+ `-build`) | v0.37 | #1913 |
| `example-values` | v0.36.3 | #1268 / #2246 |

Thirteen features remain open in `../` — the unscheduled and deferred ones.
