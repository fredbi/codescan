---
title: Inner markdown — verbatim swagger:description block scalar
stream: 8
origin: iii
status: done
release: v0.36
issues: [go-swagger#3211]
prev: "godoc-filter §C (deferred)"
---

# Inner markdown — verbatim `swagger:description` block scalar

**Status:** ✅ done (P1 `cf6d918` · P3 `7765d44` · P4 `c199fb3`) on
`feat/feature-v0.36`, awaiting review/merge. Reframes + closes go-swagger#3211.
Full design + build log: [inner-markdown-design.md](inner-markdown-design.md).

**Origin.** Deferred member of the clean-godoc cluster (godoc-filter scope
decision split off "C — inner markdown" to "a different feature that will
leverage the inner comments feature"). go-swagger#3211 ("[spec/parsing] Add
indentation support") asked for free markdown — tables, indentation, blank
lines — inside descriptions. Original stance: **wont-fix as filed**, because
godoc is allergic to plain markdown and nothing general could be salvaged from
an ad-hoc "make my comment render" request.

**Reframe.** Now that we have (1) `swagger:description` overrides and
(2) AfterDeclComments (annotations inside / below a decl, godoc stays clean),
the request becomes tractable and *non-polluting*: the author opts in by
**explicitly authoring** a `swagger:description` override carrying a verbatim
markdown block, ideally as an inner (private) comment so the godoc surface is
untouched. #3211 is closed as **reframed** — the fix is an authored annotation,
not magic markdown recovery from godoc prose.

**Decision.** A YAML literal **block-scalar marker** on the annotation line:

```go
// swagger:description |
// Overview
// ---
//
// | col1 | col2 |
// |------|------|
// | val1 | val2 |
//
// A paragraph after a *significant* blank line.
Name string `json:"name"`
```

`swagger:description |` ⇒ everything below is captured **verbatim** (blank
lines and indentation preserved) until the next `swagger:*` annotation or end
of comment. Plain `swagger:description` keeps today's blank-terminated
behaviour (Option B) — `|` is the opt-in, so output is unchanged without it.

**Why `|` and not a `---` fence.** `---` carries two markdown meanings —
thematic break *and* setext H2 underline (`text` then `---`) — so a `---` fence
would be silently closed by ordinary markdown content (silent truncation of the
very content it must hold). Putting the marker on the annotation line keeps the
body completely unconstrained. `|` is the YAML literal-block idiom, so it reads
as "keep newlines and indentation, literal text" to anyone who has written
YAML. See the design doc §rejected for the fence analysis.

**When to revisit.** v0.36 (this branch), after the design is reviewed.
