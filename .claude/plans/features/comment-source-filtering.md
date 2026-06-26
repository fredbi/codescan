---
title: After-declaration annotation comments (clean godoc)
stream: 8
origin: i
status: done
release: v0.36
issues: []
prev: "§21"
---

# After-declaration annotation comments (clean godoc)

**Status:** ✅ done (2026-06-25, `feat/feature-v0.36`). Scanner-only
`Options.AfterDeclComments`: struct inside-body + alias/type & struct-field
trailing comments. const-enum deferred (no builder semantics; would break the
scanner-only constraint). Design + build log:
[comment-source-filtering-design.md](comment-source-filtering-design.md).
Awaiting review before merge.

**Origin.** Initial design vision — godoc is not always aligned with how a spec's
documentation should read. Authors want a clean godoc *and* the swagger
annotation machinery kept out of the published documentation.
(`ramblings/grammar-vs-regexp-for-v2.md` §1.)

**Decision (groomed 2026-06-23).** Add an **opt-in** option (working name
`Options.AfterDeclComments`) that lets the scanner **also consume the comment
group sitting immediately *after* the declaration** (the next comment group
below it). Later / inline comments are ignored.

- From the scanner's perspective this is **just another comment group appended**
  to the decl's annotation source — **same annotation grammar**, no new
  private-spec syntax (this resolves the earlier "YAML-ish vs custom format"
  open question: neither, reuse the grammar).
- The godoc *above* the decl stays clean and human-facing; the swagger
  annotations live *below* the decl, where godoc renderers won't surface them.

```go
// MyStruct has a title.
//
// And it has a description too.
type MyStruct struct{ /* … */ }

// swagger:model myExposedName
//   maxProperties: 10
```

**Interactions.** Complements [swagger-description-override](swagger-description-override.md)
and [godoc-filter](godoc-filter.md) (clean-godoc cluster); the godoc-subset entry
point relates to [godoc-identifier-prefix](godoc-identifier-prefix.md).

**When to revisit.** v0.36 (CLI/TUI + faster-scanner cycle).
