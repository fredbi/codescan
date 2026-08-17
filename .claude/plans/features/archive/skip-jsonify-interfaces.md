---
title: Skip-jsonify-interfaces opt-out
stream: —
origin: ii
status: done
release: v0.36
issues: []
prev: "§3.3"
---

# Skip-jsonify-interfaces — opt out of the interface-method mangler

**Status:** ✅ done on `feat/feature-v0.36` (commit `54cf1fd`), awaiting
review/merge. Shipped as `Options.SkipJSONifyInterfaceMethods` (opt-out,
default false) per the **Shape** below — the global option, not the
`swagger:no-mangle` annotation. Single chokepoint `fields.go:methodCarrier`
gates the `interfaceJSONName` call on `s.Ctx.SkipJSONifyInterfaceMethods()`;
`swagger:name` still wins verbatim. New collision-free fixture
`fixtures/enhancements/interface-no-mangle/` + on/off integration test
(`coverage_skip_jsonify_interface_test.go`) pin the contract; default-off
output is unchanged. Docs: schema/README.md §interface-naming + root CLAUDE.md
options list.

**Origin.** Q9 close-out (2026-06-03, fix-quirks B1). A core-product enhancement
with no dedicated stream; it rides whichever stream next touches the schema
builder's Options surface.

**Scope.** Today `internal/builders/schema/` runs the `swag/mangling`
`ToJSONName` transform on every interface-method property name when the author
didn't provide a `swagger:name` override — a one-size-fits-all convention
(`CreatedAt → createdAt`, `ID → id`); see schema/README.md §method-mangler for
the rationale and the struct-vs-interface asymmetry.

It will not always be what the author wanted. Examples:

- An interface already named for its JSON shape (`Get_user` or similar
  non-Go-idiomatic spellings).
- A codebase with its own canonical-name discipline that wants codescan to stay
  out of the renaming business entirely.

**Shape.** A global option (`Options.SkipJSONifyInterfaceMethods bool` or
similar), defaulting to `false` so existing behaviour is preserved. When `true`,
`fields.go:methodCarrier` skips the `s.interfaceJSONName(fld.Name())` call and
falls back to the Go method name verbatim. `swagger:name` continues to win
regardless of the option.

A per-decl annotation (`swagger:no-mangle` on the interface or package) would
also work and is more granular, but adds a new annotation surface — the global
option is simpler.

**Witness fixture (locks current behaviour).**
`fixtures/enhancements/interface-name-verbatim/` — the `swagger:name` verbatim
contract that any opt-out implementation must still respect.

**When to revisit.** First time a user requests it, or when the Options surface
gets its next pass.
