---
title: Single-line comment as description-only (opt-in)
stream: 9
origin: iii
status: done
release: v0.35
issues: [go-swagger#2626]
prev: "§13"
---

# Single-line comment as description-only (opt-in)

**Status:** ✅ done (2026-06-17, merged into `fix/backlog-lot1` `b67c2e0`).

`Options.SingleLineCommentAsDescription`: when enabled, a single-line doc comment
routes to `description` (never `title` / `summary`), regardless of trailing
punctuation; multi-line comments keep the existing title/description split.
Plumbed via the grammar parser option `WithSingleLineCommentAsDescription`,
applied in `finaliseBase`'s `demoteSingleLineTitle` (full + preamble pairs).
Default-off.

**Origin.** go-swagger#2626 (backlog verification, 2026-06-14). Today a
single-line doc comment ending in a period is parsed as `title:`; the same line
without a trailing period becomes `description:`. The reporter argues a
single-line comment should never be a title. Declined as a default change (it
would silently reshape every existing single-line-titled model), but worth an
opt-in — cheap to land as an isolated scanner option.
