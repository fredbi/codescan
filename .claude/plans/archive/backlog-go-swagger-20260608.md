# go-swagger — Spec-related GitHub Issues

_Source: `issues-spec-20260608.json` (236 open issues, label-filtered to spec generation). Generated 2026-06-08._

The **Example?** column marks issues whose body embeds an example spec or code snippet. When marked **[▶](#a-NUM)**, the extracted snippet(s) are reproduced in the [Appendix](#appendix--embedded-examples) at the end of this document.

The **Status** column tracks per-issue triage state during the backlog pass (Stream 9 in [roadmap.md](roadmap.md)). Legend:

- ⬜ open / untriaged — default starting state.
- 👀 **verify-only** — likely already fixed by current code; needs TUI verification + screenshot, then close.
- 🐞 **repro-needed** — behaviour unclear or no example; reproduce locally first.
- 🛠 **fix-needed** — confirmed bug, needs a code change in this repo.
- ✅ **fixed** — closed with proof of fix attached.
- 🛑 **wont-fix** — out of scope, deferred to v2/OAI3, or a feature we reject.
- ♻️ **duplicate** — collapsed into another issue in this table.

The **Need doc** column flags issues that are (or turned out to be) primarily a
documentation matter — behaviour works as designed, but the confusion in the
issue points to something the doc site should explain. Mark with 📖; we make a
dedicated pass over all 📖 rows when building the doc site. A row can be both
✅ and 📖 (the code is correct, but the usage needs documenting).

When a row changes status, also annotate the closing PR / TUI screenshot inline if useful.

> **Triage complete (236/236) — 0 open 🛠.** The fixers' worklist is now empty:
> the last two 🛠, the name-identity / cyclic-$ref pair #2637 + #2783, landed on
> `fix/backlog-lot1` (2026-06-17, `feat/name-identity-cyclic-ref` merged at
> `9740c6d`) and moved to the fixed ledger — joining #1499, #2251 and #3005 from
> earlier that day. Every issue now lives in one of the four category files
> (verdicts kept, example snippets stripped — re-pull issue bodies from GitHub
> for deep analysis):
> - [`backlog-triaged-ledger.md`](backlog-triaged-ledger.md) — ✅ fixed /
>   works-as-designed / N-A, the 👀 verify-only row, and ♻️ duplicates of
>   fixed-or-active-fix issues (212).
> - [`backlog-triaged-wontfix.md`](backlog-triaged-wontfix.md) — 🛑 pure
>   wont-fix as framed, no recorded-feature tie (5; incl. 1 ♻️ dup #2202→#1960).
>   (#3211 markdown-indentation moved to the feature ledger 2026-06-26 — the
>   blocking work landed as the inner-markdown `swagger:description |` feature.)
> - [`backlog-triaged-feature.md`](backlog-triaged-feature.md) — 🛑 wont-fix-as-framed
>   but tied to a recorded forthcoming feature §N (9); convert to ✅ as each lands.
>   (§13 single-line-as-description #2626, §20 emit-x-go-type #2924 and §17
>   explicit-additionalProperties #2539 landed 2026-06-17 and moved to the fixed
>   ledger; #3211 added 2026-06-26, ✅ landed as inner-markdown v0.36.)
> - [`backlog-triaged-poison.md`](backlog-triaged-poison.md) — 🐞 poison queue
>   (**now empty** — #2633, #2778, #2874 retired to the fixed ledger 2026-06-17
>   via fail-loud feature §8.2).
>
> With the worklist empty, the four category files can now merge back into one.
> (Corpus is now **226** —
> 6 issues relabeled out on GitHub as not-spec-generation/closed: #1520, #2131,
> #2266, #2000, #2014, #2503; and 4 more closed/relabeled out-of-scope and removed
> from the ledgers 2026-06-17: #1860 (routes from a non-Go .conf file), #2053 (no
> Go files in scan dir — trivial misuse, now better documented), #1712 (historical
> go-version package-load failure, closed), #2777 (go-swagger `swagger serve` UI,
> relabeled go-swagger-specific).)

_No open 🛠 rows remain — the worklist is empty._
