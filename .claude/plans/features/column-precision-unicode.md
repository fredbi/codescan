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

**Sibling gap — routebody tracks no columns at all (folded in 2026-07-30).** The
multi-byte question above is precision *within* a tracked column; `routebody` does
not track per-line columns in the first place, so every diagnostic raised from a
`Parameters:` / `Responses:` body resolves to the line's start rather than the
offending token. Documented at `internal/parsers/routebody/README.md#quirks-open`,
previously loose in the quirk registers. Both must be closed before LSP
diagnostics can point at a span rather than a line — do them in one pass, since
they touch the same `Line.Pos` contract.

**When to revisit.** When LSP consumers actually show non-ASCII comments —
probably never for English docs, but an eventual concern for i18n annotations or
URL-embedded chars. Pairs with
[yaml-token-positions](yaml-token-positions.md) as a Stream 11 prerequisite.
