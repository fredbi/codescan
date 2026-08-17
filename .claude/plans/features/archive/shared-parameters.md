---
title: Shared swagger:parameters (and swagger:response) across operations
stream: 9
origin: iii
status: open
release: v0.36
issues: [go-swagger#2632]
prev: "§11"
---

# Shared `swagger:parameters` (and `swagger:response`) across operations

**Status:** ⬜ open.

**Origin.** go-swagger#2632 (backlog verification, 2026-06-13). The reporter used
`swagger:parameters *` to apply a shared header-parameter struct to all
operations; `*` is treated as an operation id, matches nothing, and the
parameters are silently applied nowhere.

**Decision (groomed 2026-06-23).** Today a `swagger:parameters` struct either
refers to an operation or is defined within an operation/route context. Add two
non-operation targeting forms:

- **`swagger:parameters *`** → register a **shared parameter at the spec top
  level** (`#/parameters/…`), referenceable across operations.
- **`swagger:parameters /path`** (e.g. `swagger:parameters /pet`) → register
  parameters at the **path-item level** (apply to every operation under that
  path). The **leading slash is mandatory** so a path is distinguishable from an
  operation id. (No tag-level targeting.)

**Merge.** Shared / path-item parameters combine with per-operation ones;
per-operation **overrides on `(name, in)`** collision (OAS2 path-vs-operation
semantics).

**Symmetry with `swagger:response` is a first-class requirement.** Mirror the
same rules: `swagger:response *` registers a **shared response at the spec top
level** (`#/responses/…`). (OAS2 path-items carry no responses, so the `/path`
form is parameter-specific; the `*`/top-level form is the response analog.)
Always keep parameters and responses following the same rules.

**When to revisit.** When shared-parameter demand recurs.
