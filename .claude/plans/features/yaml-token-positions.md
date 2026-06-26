---
title: goccy/go-yaml token-level YAML positions
stream: 11
origin: i
status: open
release: v0.38
issues: []
prev: "§3.1"
---

# `goccy/go-yaml` token-level YAML positions

**Status:** ⬜ open · LSP prerequisite (Stream 11).

**Origin.** Architecture §5.1 / A.4; flagged 2026-04-21 during P2.5.

**Scope.** Today `internal/parsers/yaml/` wraps `go.yaml.in/yaml/v3`, which gives
coarse line/column in errors only. If LSP wants to highlight the *specific* key
inside a 20-line embedded YAML block, the incumbent library is insufficient.
`goccy/go-yaml` exposes a token-level lexer. A POC comparing both libraries is
the first step.

**Companion follow-on — duplicate-key diagnostic emission.** Shares the same
underlying capability: walking the YAML AST with per-token positions. Q28
(resolved in the fix-quirks wave, 2026-06) silently dedupes duplicate mapping
keys last-wins; that fix has no warning surface because `go.yaml.in/yaml/v3`
exposes no public `UniqueKeys(bool)` knob and building a parallel diagnostic
infra around the AST-mutate workaround was judged too hacky. When the
position-tracking library lands, wire it: each duplicate key drops with a
`CodeDuplicateMappingKey` diagnostic carrying the file:line:col of both
occurrences. The dedupe semantics (last-wins) stays — only the warning becomes
visible. Re-check the Q28 witness fixture
(`fixtures/enhancements/meta-securitydefs-duplicate-keys/`) at the same time.

**When to revisit.** When the LSP server actually needs per-token YAML
positions. Pairs with [column-precision-unicode](column-precision-unicode.md) as
a Stream 11 prerequisite.
