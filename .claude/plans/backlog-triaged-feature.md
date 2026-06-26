# go-swagger backlog — FORTHCOMING-FEATURE-tied

_Triaged **wont-fix-as-framed** issues that each map to a **recorded forthcoming
feature** in [`forthcoming-features.md`](forthcoming-features.md). Split out of
`backlog-triaged-wontfix.md` (2026-06-17) so the pure wont-fix set stays clean.
Each verdict cites its §; example snippets stripped. These convert to ✅ as their
feature lands._

## Feature map

-  #422 → ✅ LANDED v0.36 forthcoming-features **§11** — Group parameters for each path; generate swagger parameters section
- #1268 → forthcoming-features **§2.1** — example annotation only supports strings
- #1391 → ✅ LANDED v0.36 (`NameFromTags`, feat/feature-v0.36) — Generate doc for struct having tag other than json
- #1118 → ✅ LANDED v0.36 forthcoming-features **§13** — Generate spec sometimes sets title, sometimes description for model (deterministic now; §13 = opt-in further control)
- #1992 → forthcoming-features **§19** — Hide parts of composition for a struct (per-operation field views)
- #2246 → forthcoming-features **§2.1** — swagger:model Example field forces JSON to string
- #2539 → forthcoming-features **§17** — Generate spec with additionalProperties
- #2626 → forthcoming-features **§13** — Single line comment should never be parsed as title
- #2632 → ✅ LANDED v0.36 forthcoming-features **§11** — Applying swagger:parameters to all operations
- #2924 → forthcoming-features **§20** — emit `x-go-type` vendor extension (code→spec; cheap opt-in)
- #3275 → forthcoming-features **§16** — question(generate spec) : generating spec from go types, specify required without explicit comment
- #3211 → ✅ LANDED v0.36 (inner markdown — `swagger:description \|`, feat/feature-v0.36) — [spec/parsing] Add indentation support

| # | Title | Summary | Example? | Status | Need doc |
|---:|-------|---------|:--------:|:------:|:------:|
| [422](https://github.com/go-swagger/go-swagger/issues/422) | Group parameters for each path; generate swagger parameters section | Feature request: group/share parameters per path and emit a reusable parameters section instead of duplicating per path. | [▶](#a-422) | ✅ | 📖 |
| [1268](https://github.com/go-swagger/go-swagger/issues/1268) | example annotation only supports strings | example: annotation only supports string examples; wants array/object examples like enum. | [▶](#a-1268) | 🛑 | 📖 |
| [1391](https://github.com/go-swagger/go-swagger/issues/1391) | Generate doc for struct having tag other than json | Parameter names use Go field names; wants to honor a non-json struct tag (e.g. schema:) for naming. | • | ✅ | 📖 |
| [1118](https://github.com/go-swagger/go-swagger/issues/1118) | Generate spec sometimes sets title, sometimes description for model | Inconsistent: some models get title set from the preceding comment, others get description. | [▶](#a-1118) | ✅ | 📖 |
| [1992](https://github.com/go-swagger/go-swagger/issues/1992) | Hide parts of composition for a struct. | Wants to hide composed/embedded fields (e.g. ID, Created) for POST while keeping them for GET, without polluting the model. | [▶](#a-1992) | 🛑 | 📖 |
| [2246](https://github.com/go-swagger/go-swagger/issues/2246) | swagger:model Example field forces JSON to string | swagger:model Example field forces valid JSON into a string when combined with --exclude-deps. | [▶](#a-2246) | 🛑 | 📖 |
| [2632](https://github.com/go-swagger/go-swagger/issues/2632) | Applying swagger:parameters to all operations | Wants to apply a swagger:parameters struct to all operations via a wildcard instead of listing operation ids. | [▶](#a-2632) | ✅ | 📖 |
| [3275](https://github.com/go-swagger/go-swagger/issues/3275) | question(generate spec) : generating spec from go types, specify required without explicit comment | Wants required to be inferred (e.g. from json omitempty / x-nullable) instead of adding // required:true to hundreds of structs. | [▶](#a-3275) | 🛑 | 📖 |
| [3211](https://github.com/go-swagger/go-swagger/issues/3211) | [spec/parsing] Add indentation support | Parsing strips leading whitespace/hyphens/pipes, breaking markdown tables/lists in descriptions; proposes auto-unindent. | [▶](#a-3211) | ✅ | 📖 |

## Verdicts

_Per-issue verdicts (example snippets removed; see GitHub for the original issue body)._

<a id="a-422"></a>

### #422 — Group parameters for each path; generate swagger parameters section

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/422)

**🛑 Wont-fix as framed + 📖 (verified 2026-06-16).** Grouping/sharing parameters
across paths and emitting a reusable top-level `parameters:` section is the
shared-`swagger:parameters` feature tracked as forthcoming-features §11 (#2632).
Today a parameter struct binds per-operation by id (and can be reused by listing
several ids). **Doc-site action:** document the multi-operation `swagger:parameters`
reuse; the reusable-section emission is §11.


<a id="a-1268"></a>

### #1268 — example annotation only supports strings

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1268)

**🛑 Deferred to feature §2 + 📖 (verified 2026-06-14).** An `example:` on a
`[]string` field is kept as the raw string (`"name1,name2"`), not coerced to an
array (`[name1, name2]`). This is the broader example/examples enhancement
(forthcoming-features.md §2.1); fixing this one case alone would not move
response/example handling forward. No fixture. 📖 Doc-site: note that complex
(array/object) examples are not yet coerced.


<a id="a-1391"></a>

### #1391 — Generate doc for struct having tag other than json

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1391)

**✅ LANDED v0.36 (2026-06-23, feat/feature-v0.36).** Shipped as the
`NameFromTags []string` option — an ordered list of struct-tag types the
emitted name is derived from (default `["json"]`; e.g. `["form","json"]` for
gin). Applies to schema properties, parameters and response headers; only the
name is sourced this way (encoding/json directives stay json-sourced). See
`features/naming-tags.md`. Also contributes #2912. **Doc-site action (still
open):** document json-tag naming + the new `NameFromTags` option.


<a id="a-1118"></a>

### #1118 — Generate spec sometimes sets title, sometimes description for model

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1118)

**🛑 Wont-fix as framed → forthcoming feature §13 + 📖 (re-triaged 2026-06-17).**
The 2017 complaint was that title-vs-description seemed to fire with "no clear
pattern." Post-grammar2 the split is **deterministic** and witnessed by
`fixtures/bugs/1118` + `TestCoverage_Bug1118`: a single-line comment **ending in
a period** → `title:`; the same line **without** a period → `description:`; a
multi-line comment → `title:` (first paragraph) + `description:` (remainder after
a blank line). So the "inconsistency" is resolved — it was the heuristic firing
on differently-shaped comments. Same surface as #2626; authors wanting single-line
comments routed to `description:` regardless of punctuation get the **opt-in §13
knob** (`SingleLineCommentAsDescription`). **Doc-site action:** document the
title/description heuristic (period → title, blank line separates title from
description) and, when it ships, the §13 knob.


<a id="a-1992"></a>

### #1992 — Hide parts of composition for a struct.

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1992)

**🛑 Wont-fix as framed → forthcoming feature §19 + 📖 (feature-tied 2026-06-17).**
Per-operation field views — hiding server-assigned fields (Id, Created) on create
while keeping them in responses — is registered as forthcoming-features **§19**
(per-operation field projections). It is partially supported today via
`readOnly: true` (witnessed by `fixtures/bugs/1992` + `TestCoverage_Bug1992`); the
full per-operation projection is the §19 design (worked example deferred).
**Doc-site action:** document `readOnly` for server-assigned fields; note §19 for
full projections.



<a id="a-2246"></a>

### #2246 — swagger:model Example field forces JSON to string

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2246)

**🛑 Wont-fix as framed + 📖 (verified 2026-06-16).** A JSON-object `example:`
on a field is kept as an **escaped string** (`"{\"kind\":...}"`), not parsed
into an object. This is the example-coercion gap registered as
forthcoming-features §2.1 (alongside #1268: an `example:` value is carried as a
raw string, not coerced to its field type). **Doc-site action:** note that
`example:` values are currently emitted verbatim (string), pending §2.1.


<a id="a-2632"></a>

### #2632 — Applying swagger:parameters to all operations

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/2632)

**🛑 Wont-fix (as framed) / 📖 (verified 2026-06-13).** `swagger:parameters *` —
the `*` is treated as an operation id, matches nothing, so the shared header
params are applied nowhere (silently). Binding a param struct to all/glob/tag
operations is a feature, parked as forthcoming-features.md §11. 📖 Doc-site:
list operation ids explicitly (or await §11).



<a id="a-3275"></a>

### #3275 — generating spec from go types, specify required without explicit comment

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3275)

**🛑 Wont-fix as framed → forthcoming feature §16 + 📖 (verified 2026-06-15).**
`required` is set only by an explicit `// required: true`; inferring it from json
`omitempty` / pointer-nullability for hundreds of structs is registered as an
**opt-in** feature in `forthcoming-features.md` §16 (the omitempty↔required
mapping is a convention, and flipping it by default would reshape every existing
spec). **Doc-site action:** document the explicit `// required:` mechanism and
the planned opt-in inference.


<a id="a-3211"></a>

### #3211 — [spec/parsing] Add indentation support

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/3211)

**✅ LANDED v0.36 (reframed) + 📖.** Originally wont-fix-as-framed: godoc
comment stripping removes the leading `|`/whitespace of markdown table/list rows,
and recovering arbitrary markdown from ambient godoc isn't generalisable. The
blocking pieces have since landed (`swagger:description` overrides +
AfterDeclComments), so the request is now served by **explicit authoring**: end
the annotation line with the YAML literal block-scalar marker —
`swagger:description |` — and the body below is captured verbatim (blank lines,
indentation, table pipes preserved) until the next line-leading annotation or
EOF. Shipped on `feat/feature-v0.36` (inner-markdown feature; lexer P1 `cf6d918`,
end-to-end P3 `7765d44`, docs P4 `c199fb3`). The old preamble-path pipe-strip is
unchanged by design — markdown belongs in the explicit override, not ambient
godoc. **Doc-site action:** how-to for the `swagger:description |` markdown body
(rides the doc-site cadence). See `features/inner-markdown.md`.
