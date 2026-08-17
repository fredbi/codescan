> [!NOTE]
> **✅ SHIPPED RECORD** — split out of `../doc-site-wishlist.md` on 2026-08-17 so the live wishlist holds
> only open work. Nothing here needs doing. The live file is `../doc-site-wishlist.md`.

# Doc-site wishlist — what shipped

The W-series ran from 2026-06-13. These landed; the reasoning behind each is in git history and in the
pages themselves, not repeated here.

| # | Item | Landed |
|---|------|--------|
| W1 | OpenAPI UI widget (`{{< openapi >}}`) on the capstone | 2026-06-26 (`06ce50f`) |
| W6 | Advanced modeling — polymorphic types tutorial, complete and illustrated | 2026-06-15/17 |
| W10 | Annotation ↔ spec cross-highlight | 2026-08-08 (PR #79) |
| W11 | WASM live playground | 2026-08-08 (PR #79) |
| W15 | Render the grammar visually — **both** tiers, incl. the railroad diagrams that had been parked | 2026-08-14 |
| W16 | "About / why codescan" explainer (`about.md`, 97 lines) | 2026-06-13 (#41) |
| W8a | Getting started — "Usage as a headless CLI" | 2026-08-16 (PR #115) |
| W8b | Getting started — "Usage as a terminal UI" | 2026-08-08 (`2da57cad`) |
| — | `maintainers/performance.md` — never a W-item; came out of the benchmarks work | 2026-08-17 (PR #118) |

## Two notes worth keeping

**W15 was unparked.** The railroad tier was parked 2026-06-23 as low-ROI and shipped anyway, as a
`railroad` shortcode used 14× in `maintainers/grammar.md`. It was hardened during the security pass
(`3681d380`: `securityLevel` off `'loose'`, error paths built with `textContent`) — start there for any
future diagram shortcode. Parked in the live file means "nobody is carrying it", not "decided against".

**W7 shipped its surface but not its framing** — see `../doc-site-wishlist.md`, which keeps the residue
as an open question rather than claiming it done.
