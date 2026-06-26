---
title: ExternalDocs on non-meta OAIv2 objects
stream: 9
origin: iii
status: done
release: v0.35
issues: [go-swagger#2872, go-swagger#2655]
prev: "§9"
---

# ExternalDocs on non-meta OAIv2 objects

**Status:** ✅ done (closed 2026-06-17).

All four OAS2 `externalDocs` host objects are now covered: **meta/info** (the
original #2872 fix), **operation**, **schema**, and **tag**.

- **operation** — `swagger:route` (new case in `routes/walker.go`
  `dispatchRouteKeyword`) and `swagger:operation` (already worked via the YAML
  body unmarshal; locked by test).
- **schema** — `swagger:model` and any full Schema (e.g. a body parameter's
  schema), via a new `KwExternalDocs` arm in `handlers.schemaRawHandler`. Treated
  as full-Schema-only: on a SimpleSchema site (non-body param / header / items)
  it is dropped with `CodeUnsupportedInSimpleSchema` (gated in `handlers.Raw`).
  Shared `handlers.ParseExternalDocs` helper; the meta builder keeps its own
  inline copy (spec pkg doesn't import handlers).
- **tag** — a per-tag `externalDocs:` inside a `swagger:meta` `Tags:` list, via
  the `KwTags` → `[]spec.Tag` unmarshal in `spec/walker.go`
  (`dispatchMetaYAMLBlock`). This was actually already implemented by the #2655
  tag-list work, not deferred.

**Locked by** `TestCoverage_Bug2655` (`fixtures/bugs/2655/api.go`), which carries
a `pet` tag with nested `externalDocs` and asserts `Description` + `URL` survive.

**Origin.** go-swagger#2872 (backlog verification, 2026-06-13): the focused fix
wired `ExternalDocs` into the meta/info builder, but OpenAPI 2.0 also allows
`externalDocs` on schema, operation, and tag objects.
