# go-swagger backlog — WONT-FIX (as framed)

_Triaged issues **wont-fix as framed** (out of scope, downstream/UI/codegen, or a
feature we reject), plus duplicates of wont-fix ones. Refinement — which of these
are worth promoting to a forthcoming feature — will **refresh from the GitHub
issues directly**. Example snippets stripped; verdicts retained._

| # | Title | Summary | Example? | Status | Need doc |
|---:|-------|---------|:--------:|:------:|:------:|
| [618](https://github.com/go-swagger/go-swagger/issues/618) | Using slices instead of structs for parameter annotation | Wants to annotate parameters using slices instead of typed structs. | [▶](#a-618) | 🛑 | 📖 |
| [1064](https://github.com/go-swagger/go-swagger/issues/1064) | Two routes using same operation | Two routes share one operation; wants to avoid duplicating swagger:operation for each. | [▶](#a-1064) | 🛑 | 📖 |
| [1777](https://github.com/go-swagger/go-swagger/issues/1777) | Change order to API request | Wants to control endpoint ordering (e.g. POST before list) instead of alphabetical. | [▶](#a-1777) | 🛑 | 📖 |
| [1960](https://github.com/go-swagger/go-swagger/issues/1960) | Generated spec does not preserve property order in structs | Generated spec doesn't preserve struct property order; suggests using tags to sort. | [▶](#a-1960) | 🛑 | 📖 |
| [2202](https://github.com/go-swagger/go-swagger/issues/2202) | x-order not working in generate spec | Duplicate of #1960 — x-order extension doesn't reorder keys in the generated YAML. | [▶](#a-1960) | ♻️ | 📖 |

## Verdicts

_Per-issue triage verdicts (example snippets removed; see GitHub for the original issue body)._

<a id="a-618"></a>

### #618 — Using slices instead of structs for parameter annotation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/618)

**🛑 Wont-fix (as framed) + 📖 (verified 2026-06-14).** The `[]string` is a plain
package var, not an annotation target — a name-only slice carries no type / `in`
/ `required` information, so it cannot drive parameter generation. codescan's
parameter model is struct-based (`swagger:parameters` over a typed struct).
Declined as framed. 📖 Doc-site: document the struct-based parameter model and
why a bare name slice is unsupported.


<a id="a-1064"></a>

### #1064 — Two routes using same operation

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1064)

**🛑 Wont-fix as framed + 📖 (verified 2026-06-16).** Two routes cannot share a
single operation: in OpenAPI 2.0 each path+method is a distinct operation and
`operationId` must be UNIQUE — the same id on two paths produces an invalid
duplicate-operationId spec (both paths emit, but the ids collide). Share the
`swagger:parameters` / `swagger:response` structs to avoid repeating those; the
operation entries themselves are necessarily per-path. **Doc-site action:**
document that each path needs its own operation (unique id); the
duplicate-operationId diagnostic is the #3134 uniqueness idea.


<a id="a-1777"></a>

### #1777 — Change order to API request

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1777)

**🛑 Wont-fix as framed + 📖 (verified 2026-06-15).** OpenAPI 2.0 `paths` is an
unordered JSON object (modelled as a Go map); endpoint declaration/display order
is **not meaningful in the spec** and cannot be controlled by codescan. Ordering
is a UI/rendering concern (e.g. the doc UI can sort by tag/method). **Doc-site
action:** state that path order in the generated spec is not significant and is
not configurable.


<a id="a-1960"></a>

### #1960 — Generated spec does not preserve property order in structs

[↑ table](#go-swagger--spec-related-github-issues) · [issue on GitHub](https://github.com/go-swagger/go-swagger/issues/1960)

**🛑 Wont-fix as framed + 📖 (verified 2026-06-15).** The generated spec does not
preserve struct property order — definitions and their `properties` are Go maps,
which are unordered, and property order is not a meaningful OAS concept. The
reporter's `x-order` tag suggestion would require an order-preserving JSON/YAML
serialization layer, which is below the planning horizon (see the ordering
boundary; same class as #1777/#1960). **Doc-site action:** state that property
and path order in the generated spec is not significant and not configurable.


_(#3211 moved to [`backlog-triaged-feature.md`](backlog-triaged-feature.md) — the
blocking work landed; see the reframe note there.)_
