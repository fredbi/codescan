---
title: Map additionalProperties for non-string keys
stream: 9
origin: iii
status: done
release: v0.35
issues: [go-swagger#2251]
prev: "§18"
---

# Map `additionalProperties` for non-string keys

**Status:** ✅ done (2026-06-17, `feat/additional-properties` Phase 1). Plan:
[additional-properties.md](../additional-properties.md).

`buildFromMap` now emits `additionalProperties` for any JSON-legal key — string,
integer, uint, or `TextMarshaler`, via `resolvers.IsJSONMapKey` — and warns
(`CodeUnsupportedType`) on a json-illegal key instead of silently dropping the
property to an empty (typeless) schema.

**Origin.** go-swagger#2251 (backlog verification, 2026-06-16). `buildFromMap`
emitted `additionalProperties` only when the map key's underlying type was
literally `string` or implemented `encoding.TextMarshaler`; every other key type
hit `return nil`. That was too narrow: `encoding/json` marshals integer-kind map
keys (`map[int]V`, `map[int64]V`, named integer types) as JSON string keys, so
they ARE representable as `{type: object, additionalProperties: V}` — but
codescan dropped them, emitting a bare `{x-go-name}` with no type.

**Deferred refinement.** Auto-deriving a tighter `patternProperties` regex from
an integer-keyed map (instead of the loose `additionalProperties`) is tracked in
[pattern-properties inference](pattern-properties-inference.md).
