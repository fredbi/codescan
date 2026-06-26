---
title: Column precision beyond ASCII
stream: 11
origin: i
status: open
release: v0.38
issues: []
prev: "§4.1"
---

# Column precision under `/* */` continuation beyond ASCII

**Status:** ⬜ open · LSP prerequisite (Stream 11).

**Origin.** P1.1 preprocessor — current column math uses byte offsets; works
correctly for ASCII godoc but overstates columns when a continuation line
contains multi-byte runes before the content start.

**Scope.** Switch to rune-based column counting (or UTF-16 for LSP interop) in
`stripLine`. Small change, affects all `Line.Pos.Column` values.

**When to revisit.** When LSP consumers actually show non-ASCII comments —
probably never for English docs, but an eventual concern for i18n annotations or
URL-embedded chars. Pairs with
[yaml-token-positions](yaml-token-positions.md) as a Stream 11 prerequisite.
