---
title: Emit x-go-type vendor extension (code→spec)
stream: 9
origin: iii
status: done
release: v0.35
issues: [go-swagger#2924]
prev: "§20"
---

# Emit `x-go-type` vendor extension (code→spec)

**Status:** ✅ done (2026-06-17, merged into `fix/backlog-lot1` `de306ca`).

`Options.EmitXGoType`: stamps `x-go-type` (`<package path>.<type name>`) on each
emitted definition in `annotateSchema`, under the `SkipExtensions` umbrella and
presence-guarded so it never clobbers the special-type recognizers' deliberate
`x-go-type` (`error`, the generic `PkgForType` fallback). Default-off. Cheap — no
new discovery, just one more vendor extension on the existing
`x-go-name` / `x-go-package` emit seam.

**Origin.** go-swagger#2924 (backlog poison qualification, 2026-06-17). codescan
already stamps `x-go-name` and `x-go-package` for origin traceability; #2924 asks
for `x-go-type` too. Conventionally `x-go-type` is a spec→code import directive,
but emitting it on code→spec output records the originating Go type — useful for
round-tripping and for tools that re-import the generated spec.
