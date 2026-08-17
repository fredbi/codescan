> [!NOTE]
> Last revision: 2026-08-17 — rewritten. Open work only; the shipped record moved to
> [`archive/doc-site-wishlist-done.md`](archive/doc-site-wishlist-done.md).

# Doc-site wishlist

Open work on the Hugo site (`docs/doc-site/`, theme in `hack/doc-site/hugo/`, test-covered examples in
`docs/examples/`). Not committed work — a list to pick from.

**Rule adopted this revision: "source ready" is not a status.** Several items carried it while nothing
existed beyond notes about what might be possible. An item is 📝 *planned* until a page exists on the
site. Claims below say what was actually checked.

| # | Item | State | Size |
|---|------|-------|------|
| W20 | Maintainers-oriented reference, consolidating the repo's READMEs (absorbs W4) | 📝 planned | 🔴 large |
| W17 | Syntax pitfalls page | 📝 planned — pre-writing only | 🟡 |
| W8c | Getting started — "Usage with `go:generate`" | 📝 planned | 🟢 |
| W13 | OpenAPI 3.x coverage | ⏸ on hold | 🔴 |

⛔ Dropped 2026-08-17: **W19**, **W9**.
❓ Undisposed, need a call: **W7-residue**, **W3**, **W14**, **W2**, **W18** — last section.

---

## W20 — Maintainers-oriented reference 🔴

The project's real architecture documentation lives in `README.md` files through the source tree and
none of it is on the site. **16 package READMEs, ~7,000 lines** (surveyed 2026-08-17, excluding `hack/`,
the vendored theme and `.claude/`):

| lines | file | lines | file |
|---:|---|---:|---|
| 2139 | `internal/builders/schema` | 302 | `internal/builders/validations` |
| 1194 | `internal/parsers/grammar` | 297 | `internal/benchmarks` |
| 1025 | `internal/scanner` | 247 | `cmd/genspec-wasi` |
| 618 | `cmd/genspec-tui` | 226 | `internal/builders/routes` |
| 342 | `internal/builders/responses` | 189 | root `README.md` |
| 320 | `internal/builders/parameters` | 185 | `internal/parsers/yaml` |
| 178 | `internal/builders/handlers` | 175 | `internal/parsers/routebody` |
| 162 | `internal/builders/common` | 157 | `cmd/genspec` |
| 136 | `internal/builders/godoclink` | 99 | `internal/builders/operations` |

These are the **normative contracts**, cited by section anchor from `.claude/CLAUDE.md` and from code
(`internal/scanner/README.md#loader`, `internal/builders/schema/README.md#allof`, `#omit`,
`#ref-override`, …). A contributor must already know the tree to find any of it.

⚠️ **Decide first: mount, copy, or generate.** Copying is the obvious move and the wrong one — that is
the drift that made the keyword reference lie once already (the reason W9 existed), here at 7,000 lines
against documents that change with every builder fix.

- **Mount** (as `docs/examples/` already does via `{{< code >}}`) — no drift, but these are written for
  maintainers reading source: relative links and code-relative anchors break as site pages. The
  link-rewriting is the real work.
- **Generate** a Maintainers section at build time — same no-drift property, more machinery, inherits
  whatever heading structure the READMEs happen to have.
- **Curate** a smaller site-side synthesis, READMEs stay authoritative — no drift risk (the site never
  claims completeness), but it is new writing and the two can still disagree.

Convention already settled: `maintainers/commands.md` is the seam for "how the tools are built rather
than what they do". W20 is the same question at ten times the size — **answer it once, for all of it.**

### W20a — pipeline diagram (was W4)

Folded in 2026-08-17: maintainers-oriented material, so **not** the cheap standalone item it was filed
as. A mermaid graph exists **as a plan and was never published** — `.claude/plans/type-dependencies.mmd`,
dated 2026-03-23 and now five months stale, i.e. it predates the package split it describes. Verified:
the only mermaid on the site is on `performance.md` and `ROADMAP.md`; the Maintainers landing has none.

## W17 — Syntax pitfalls page 🟡

A "gotchas" page on the brittle edges of the annotation syntax — where a spec silently comes out wrong
rather than erroring.

**State: pre-writing only.** Material gathered, no page written (verified: nothing on the site matches
pitfall / gotcha / known-limitation). What is gathered:

- **Indentation in `swagger:operation` YAML bodies** — the body after `---` is YAML, and `gofmt`
  re-indents doc comments with tabs, which `yaml.RemoveIndent` must expand before stripping.
- **The prose-token footgun** — a line *starting* with `swagger:<name>` or a keyword (`name:`,
  `example:`, `patternProperties:`) parses as that annotation even in prose, silently truncating a
  description. Rule: keep such tokens mid-line or backticked.
- **List separators and bracket forms** — comma-separated `enum`, the bracketed `[a,b,c]` form, security
  AND (several keys in one requirement) vs OR (separate items).
- **Title vs description split** — a single-line comment ending in punctuation becomes `title`, and the
  `SingleLineCommentAsDescription` knob that opts out.

Witnesses exist and are test-covered: `shaping/singleline`, `concepts/maps`, `concepts/security`.

## W8c — "Usage with `go:generate`" 🟢

Last of the three Getting-started siblings; the other two shipped. Cheapest item here.

## W13 — OpenAPI 3.x coverage ⏸ on hold

The documentation arm of the OAI v3 stream: both dialects and their differences (SimpleSchema
disappears, nullable becomes native, `example` vs `examples`). **Resumes when that stream does** — gated
on `internal-document-model.md`. Not startable, not lost.

---

## ⛔ Dropped 2026-08-17

- **W19 — "try it" tab on tutorial examples.** Technically too constraining. The feasibility work
  (46 files / 44 packages, imports almost all stdlib, 8.4 MB gzipped artifact on first activation)
  survives in the archive if anyone revisits; the answer is no.
- **W9 — generate the reference tables from the grammar.** Too hard, too complex, probably not feasible.
  Previously parked, now closed. **W14 was its cheap half** — see below.

## ❓ Undisposed — need a call

Five leftovers, one of them a correction. Listed so they get decided rather than forgotten.

1. ⚠️ **W7-residue — correcting an earlier claim of mine.** I marked W7 (document & security, prefer
   overlay) ✅ DONE because the pages mention `InputSpec`/overlay several times. That check was wrong —
   counting mentions is not checking framing. **The surface shipped** across
   `tutorials/document-metadata.md` and `tutorials/security.md`, but W7's actual ask — *lead* with the
   overlay recommendation, treat in-code annotations as the fallback — was not applied: overlay appears
   at `document-metadata.md:59` as "Alternatively". Reframe the two pages, or drop the ask. Both are
   defensible; leaving it recorded as done is not.
2. **W3 — known limitations page.** Its gate (F1–F9) cleared, so it is deferred by choice now. If
   written, draw on `quirks-open.md`'s small deliberate open set — Q45, Q48, Q23-edge, the
   accepted-behaviours list — not the F-series.
3. **W14 — example-coverage gate.** A CI step asserting every grammar annotation has an example package
   and a tutorial anchor. It was W9's cheap half; with W9 dropped, this is either the surviving useful
   piece of that idea or it goes the same way.
4. **W2 — copy-to-clipboard on panes.** Verified open: neither `example.html` nor `compare.html` carries
   a copy affordance, so the custom shortcodes do bypass Relearn's built-in. Small.
5. **W18 — MD051 / relref linter false positives.** The linter flags every same-page link to a
   colon-bearing heading (almost every annotation heading) and every `{{% relref %}}`. The cost is
   masking — a genuinely broken fragment is indistinguishable from the noise. Cheapest fix is disabling
   MD051 for `docs/doc-site/**`; the only fix that keeps the rule useful is teaching the analyzer Hugo's
   anchor derivation, in the go-fred-mcp markdown analyzer, which is ours.

---

Companions: `archive/doc-site-quirks.md` (scanner bugs the docs revealed),
`archive/doc-site-reference.md` (the original build plan),
`archive/doc-site-backlog-alignment.md` (the go-swagger-issue alignment audit, complete).
