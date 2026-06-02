# Scrambler / Issue-Repro Tool for codescan

Date: 2026-06-02
Status: 🟢 design largely settled — L1 rename dropped (§5); UX/interaction model
& key build decisions locked (§12.5); a few implementation open points (§10).
Next: build L0 (prune) on `feat/genspec-tui`.
Origin: `wasm-playground.md` §11 ("Other ideas"); memory [[project_genspec_tui]].

## TL;DR

- **The support problem.** Users hit a codescan bug, but rarely share their code
  (proprietary descriptions, example values, business logic). Maintainers can't
  reproduce → issues stall, bereft of a runnable repro.
- **The tool.** From source (driven by the TUI, later WASM, or programmatically),
  produce a **minimized, prose-scrubbed Go source tree** that *still reproduces
  the bug*, ready to attach to a GitHub issue — the user's real prose and example
  values never leave their machine.
- **Two passes (rename dropped, §5):**
  - **L0 — Prune.** Reachability-minimize from the annotation roots + strip
    function bodies. Delete-only → *cannot* hide the bug, and must yield a
    **byte-identical spec** (strong oracle).
  - **L2 — Redact.** Replace descriptions / titles / example values with
    **plausible, load-bearing gibberish** (a description stays a real-looking
    description — *not* `*****` blackout), layout-preserving so annotations still
    parse.
- **No renaming.** Identifiers (type/field/package names) stay **real** — so the
  maintainer's fix is directly usable by the reporter, and we never risk a rename
  masking or inventing behavior. (§5 records why, and how L1 could return.)
- **codescan is its own oracle.** We scan both originals and prove fidelity:
  L0 ⇒ identical spec; L2 ⇒ spec differs only in the redacted strings (§2).
- **Lives in the main module, AST-only, dual-use.** `internal/scrambler` (with
  its own tiny annotation-aware scanner) + a thin public `codescan.Scramble(...)`
  entry. Pure `go/ast` / `go/printer` / `archive/tar` → WASM-clean. Reused later
  as an **internal test-fixture minimizer** (Fred).
- **Driven by a TUI "issue-report mode"** (§12): an explicit, transient mode with
  an in-memory buffer, live spec/diagnostic regen as the verification surface, and
  jsonpointer-anchored remarks on the generated spec. This is the canonical UX.
- **Not rebase-gated.** No grammar2 diagnostics/positions needed → buildable
  *now* on `feat/genspec-tui`, in parallel with the library branch (§9).

---

## 1. Motivation

From `wasm-playground.md` §11, verbatim intent:

> So far it has been super painful to help users report issues as they rarely
> want to share their code. As a result, many issues remain pending, bereft of
> sufficient information to reproduce.

The fix is to make sharing *safe and cheap*:

1. **Safe** — the artifact carries no recognizable domain *content*: no business
   logic (bodies stripped), no real descriptions/examples (redacted). What
   remains is the annotated **type shapes** with their real identifiers.
2. **Cheap** — one action in a tool the user is already running to *see* the bug,
   producing a ready-to-attach archive + a prefilled issue body.
3. **Faithful** — and this is the crux — the repro must *actually reproduce* the
   reported behavior, or it wastes everyone's time. We verify it (§2).

This is squarely a **maintainer-support** feature, which is why it ranks high
despite not being on the critical scanner path.

### What we deliberately keep visible (and why)

Dropping the rename pass (§5) means **identifiers stay real**. That's a chosen
trade-off, not an oversight:

- A maintainer's answer ("the bug is in how field `CreatedAt` on `Order` maps")
  is **immediately usable by the reporter** — no secret map to translate back.
- We never risk a rename **masking the bug or inventing a new one** — fidelity
  reasoning stays simple.
- The genuinely sensitive payload is **logic + prose + values**, and those are
  exactly what L0+L2 remove. Type/field *names* are rarely the secret.

Users for whom even names are sensitive can hand-edit before sharing, or we
revisit L1 later behind the verification loop (§5).

---

## 2. codescan as its own oracle

Anonymizing for a bug repro is normally an act of faith. We can check it, because
the transformation's correctness is testable against the very tool the bug is
about. With rename gone, the invariants are *tight*:

```
spec_original = codescan.Run(original)        # private, never shared
pruned        = L0(original)
spec_pruned   = codescan.Run(pruned)
assert  spec_pruned == spec_original           # EXACT — L0 deletes only unread code

redacted      = L2(pruned)
spec_redacted = codescan.Run(redacted)
assert  spec_redacted == redact(spec_original) # differs ONLY in redacted strings
```

- **L0 must be byte-identical.** We removed only decls/files/imports the scanner
  never reads and function bodies it never executes. Any spec change is a *prune
  defect* (we cut something load-bearing) — fail loud, never ship.
- **L2 differs only in the redacted prose/values**, and `redact(spec_original)`
  applies the *same* deterministic word-substitution to the private spec's
  description/title/example fields. Equality there proves redaction touched
  nothing structural.
- **Verdict surfaced to the user:** "✅ verified faithful — safe to share" or a
  diff. If the user's *bug* is itself a spec divergence, that's the thing they're
  reporting — the loop still confirms anonymization didn't add a *second* change.

> Caveat — crashes. If the *original* panics, there's no spec to compare. Fall
> back to asserting the **same panic / error class** on the anonymized tree.

This invariant doubles as the package's own regression oracle: golden fixtures
already pair source → spec, so tests assert L0 preserves the spec exactly and L2
preserves everything but the scrubbed strings.

---

## 3. What codescan observes (the preservation contract)

The scrambler is a **rewrite where "semantics" = what the scanner reads** — far
narrower than "compiles to the same program." Everything *outside* that surface
is free to strip or scrub; everything *inside* it is sacred.

| Source element | Scanner cares about | Anonymizer action |
|---|---|---|
| Annotation keywords (`swagger:model`, `swagger:route`, …) | exact text drives dispatch | **sacred — verbatim** |
| Validation directives (`minimum`, `required`, `enum`, format names), `in:body`, status codes, HTTP verbs | semantic args | **sacred — verbatim** |
| Type *kinds* + the type graph (struct/slice/map/ptr/iface/alias) | structure → schema shape | **sacred — preserve shape** |
| Recognized std/ext identity (`time.Time`, `strfmt.UUID`, `json.RawMessage`) | matched by (import path, name) | **sacred — keep path+name** (`resolvers/assertions.go:96,104`) |
| Method sets (`MarshalText` → `IsTextMarshaler`) | presence of named methods | **L0: keep method *signatures***, strip bodies (`assertions.go:77`) |
| Struct-tag *keys* (`json:`, `validate:`, custom) | which tags exist | **sacred — verbatim** |
| Type / field / package / const **names** | identity only | **kept REAL** (rename dropped, §5) |
| `json` tag *name* / field name → spec property name | the property name | **kept REAL** (so the spec reads naturally for the OP) |
| Descriptions / titles / comment prose | free text → spec strings | **L2: load-bearing gibberish, layout-preserving** |
| String-literal example/default values | some feed `example:`/`default:` | **L2: scramble, keep Go type & shape** |
| Enum const *values* | appear in spec `enum` | **kept REAL for now** (§12.3) |
| Function *bodies* | never read (no execution) | **L0: empty via named-return rewrite** (§12.3) |
| Unreferenced decls / files / imports | never reached | **L0: delete** |

---

## 4. The two transforms

Independently toggleable passes, run L0 → L2. Both work at the AST level; output
is gofmt-clean Go.

### L0 — Prune (minimize)

- **Reachability sweep.** Roots = decls carrying `swagger:` annotations
  (detectable directly from AST doc-comments, or via `internal/scanner`'s
  `TypeIndex` classification — same module, free to consult). Walk the type graph
  (struct fields, elems, map key/value, embeds, named-type targets, and
  receiver/param/return types of recognized interface methods) and **keep only the
  transitive closure.** Then drop empty files and now-unused imports.
- **Empty function bodies via named-return rewrite** (name the results, bare
  `return`; §12.3). **Keep method signatures** — method *presence* drives
  `IsTextMarshaler` and friends; dropping `MarshalText` silently flips a schema
  from string to object.
- **Risk: lowest.** Delete-only; deleting unread code can't change scanner output.
  The §2 oracle demands a byte-identical spec.
- **Value: highest.** 50k-line project → ~200-line repro; also strips the actual
  business logic (the bodies), which is most of the sensitivity.

### L2 — Redact (de-identify prose & values)

- **Descriptions / titles / comment prose** → **plausible gibberish that keeps
  the functional load**: still a real-looking, multi-word description/title — not
  a `*****` blackout. **Layout-preserving**: same line count, blank-line
  positions, leading whitespace, continuation shape (the sectioned parser keys
  off exactly these — `parsers/sectioned_parser.go`,
  [[project_lsp_diagnostics_target]]). Replace *words*, never *structure*.
- **Example / default string values** → same-shaped gibberish (keep type, rough
  length); numbers/bools → benign same-type constants — but never where the value
  is itself a sacred directive arg.
- **Enum values** → **kept real for now** (Fred, §12.3) — not scrambled.
- **Mechanism note (§7):** rewriting comment prose is safest as **AST-informed
  source surgery** (locate comment byte-ranges via parsed positions, substitute
  in place) rather than mutate-AST-and-reprint, which can shuffle comment
  attachment.
- **Risk: higher** (layout sensitivity), and the one pass that benefits from the
  scanner's annotation parse to know prose-vs-directive precisely (§7).

---

## 5. Scope: L1 (rename) dropped — recorded decision

**Decision (Fred, 2026-06-02): drop the rename pass.** Rationale:

1. **Fidelity.** Renaming makes it hard to anticipate bugs / unexpected behavior —
   the rename itself could mask the reported bug or introduce a new one.
2. **OP-usability.** A maintainer's fix expressed against scrambled names is not
   directly exploitable by the original reporter, who'd have to translate it back
   through a secret map — friction that defeats the "help users" purpose.

So v1 is **L0 + (optionally) L2**. The remaining cut to settle:

| Cut | Gives | Costs |
|---|---|---|
| **L0 only** | minimal runnable repro, bodies+unread code gone; spec byte-identical | none beyond the prune walk |
| **L0 + L2** | + prose/value scrubbing (the privacy tier most users need) | layout-preserving redactor + annotation-structure awareness |

**Provisional lean (to debate):** L0 is the safe, high-value first milestone
(builds the harness + the §2 oracle, useful even without redaction). L2 follows
as the pass that makes output actually shareable for prose-sensitive users.

**If L1 ever returns:** only behind the verification loop, as an opt-in "I also
need names hidden" mode that emits the secret reverse-map locally — never the
default, given the OP-usability cost.

---

## 6. Architecture & placement

**Lives in the main `codescan` module, AST-based, dual-use** (Fred): besides the
issue-repro use, it's a **fixture minimizer for our own internal testing** (mint
small golden cases from real code). That argues for a clean *programmatic* entry,
not just a TUI button.

```
internal/scrambler/         # pure go/ast · go/printer · go/types · archive/tar (WASM-clean)
  scan.go     # dedicated tiny scanner: find annotation roots + classify comment
              #   spans (keyword / directive / prose) — NO dep on internal/scanner
  reach.go    # L0: transitive closure from annotation roots over the type graph
  prune.go    # L0: empty bodies (keep method sigs) → drop unreachable → drop unused imports
  redact.go   # L2: load-bearing, layout-preserving prose gibberish (source surgery)
  verify.go   # §2 oracle: re-scan + assert (identical | redacted-only delta)
  emit.go     # tree → tgz · prefilled issue body
  scrambler.go # pipeline + Options (passes on/off, gibberish style)
```

**No public wrapper needed.** The genspec-tui module
(`github.com/go-openapi/codescan/cmd/genspec-tui`) imports `internal/scrambler`
**directly** — verified: the `internal/` rule is *path-prefix* based, not
module-bound, so a separate module *under* `codescan/` may import its internals
(`go.work`/`require` handle resolution; `internal/` packages ship in the module
zip). The public API stays untouched. (A public package would be needed only for
a consumer *outside* the `codescan/` prefix — e.g. go-swagger — which is not a
current goal.)

Naming note: Fred's term is **`internal/scrambler`** (this doc's earlier
"anonymizer" = same thing). The package carries its **own dedicated annotation
scanner** rather than depending on `internal/scanner` — keeps it ast-only,
WASM-clean, and decoupled from the grammar2 rebase.

- **Consumption:** the TUI imports `internal/scrambler` directly (see box above);
  internal tests use it directly too. Public surface stays minimal by adding
  *nothing* ([[feedback_test_api_surface]]).
- **TUI surface** (`cmd/genspec-tui`): the **issue-report mode** (§12) — `Ctrl+I`
  in/out, `m`/`s` minimize/scramble, spec remarks, `Ctrl+X` export sheet. Live
  regen is the verification surface; the §2 oracle backs it at export.
- **WASM surface** (Phase 2): same package compiles; anonymize entirely
  in-browser → strongest privacy claim. Synergy with [[project_wasm_playground]].
- **Standalone `cmd/codescan-anon`?** Possible for CI/scripted fixture-minting;
  defer until the programmatic API settles.

---

## 7. Hard parts / tensions (post-L1-drop)

1. **L2 comment surgery vs reprint.** `go/printer` re-emitting a mutated AST can
   shuffle comment placement. Safer: parse for *positions*, then do targeted
   byte-range substitution on the source text for comment prose (file still
   parses; everything else byte-identical). Prune (L0) can reprint freely; redact
   (L2) should surgically edit the pruned source text.
2. **Knowing prose from directives (L2).** To gibberish a description but not a
   `swagger:` keyword or a `minimum:` arg, L2 needs the annotation's *structure*.
   Two options: **(a)** consult the scanner's parsed annotation spans via a small
   read-only accessor (precise; Fred OK'd adding the accessor) — but couples to
   the parser being rewritten under grammar2; **(b)** AST-only heuristic (treat
   `swagger:`-keyword and `key:`-directive lines as sacred, gibberish the rest)
   — parser-agnostic but fuzzier. Lean (a) where available, (b) as the portable
   fallback. *(This is the only place L2 touches the parser; see §9.)*
3. **Method-set preservation (L0).** Keep method signatures, never strip
   `MarshalText`/`UnmarshalText`; the §2 oracle catches mistakes (schema flips).
4. **Reachability completeness (L0).** The closure must be a *superset* of what
   the schema builder walks (embeds, aliases, array elems, map kv, recognized
   method receivers). Under-pruning is safe (bigger repro); over-pruning is caught
   by the byte-identical-spec assertion. Prefer the independent `go/types` walk
   over a new scanner trace hook — no production-API widening.
5. **Determinism.** L2's substitution must be deterministic so the §2 reference
   redaction matches and re-runs are stable.
6. **Import pruning must stay toolchain-free (WASM).** ⚠️ Pruning unreachable
   decls leaves dangling imports; the obvious fix (`goimports` /
   `golang.org/x/tools/imports`) **shells out to the `go` binary** and can't ship
   to WASM. But we **only ever *remove* imports** — so the two expensive,
   toolchain-coupled behaviors are out of scope: **(1) no reordering** of the
   surviving imports, **(2) no resolving/adding** new entries. What's left is a
   **small self-contained AST pass** (Fred): collect the package idents used in
   the surviving AST → delete the unreferenced `ast.ImportSpec`s (resolve unnamed
   imports via path→assumed-name). Reference for the removal logic only:
   go-swagger `generator/internal/language/format_lite.go`
   (`deleteImportSpec`/`collectTopNames`/`importPathToAssumedName`) — skip its
   `imports.Process` sort and its add-missing path.

---

## 8. Output artifacts

*(The full export flow + handoff is §12.4; this is the artifact inventory.)*

- **`repro.tgz`** — the minimized/scrambled tree (gofmt-clean; `archive/tar` +
  `compress/gzip`, WASM-clean) — or **inline fenced block** when small enough to
  embed in the issue body (the common case after minimize).
- **The resulting spec** — so the maintainer sees actual output and the user can
  show "expected vs actual."
- **The diagnostics** and **the remarks** (jsonpointer-anchored, §12).
- **Prefilled issue body** — codescan version, exact `Options`, summary, the §2
  verdict; posted via clipboard / browser / `gh` (§12.4).
- *(No secret manifest needed — nothing is renamed.)*

---

## 9. Why this is NOT rebase-gated

The scrambler needs only `go/ast` + `go/types` (stdlib, parser-agnostic) and the
scanner's classification (annotation roots) — present on master today. So:

- **L0 is fully buildable now** on `feat/genspec-tui`, in parallel with grammar2.
- **L2** is buildable now via the AST-only heuristic (§7.2b); its *precise* mode
  (§7.2a, scanner annotation spans) is the only part that benefits from — but is
  not blocked by — the grammar2 rebase (cleaner comment-span positions, the
  read-only accessor riding the v2 index contract,
  [[project_grammar_parser_migration]]).

A rare TUI-adjacent feature that *advances the library* (a reusable fixture
minimizer) while the parsing branch cooks.

---

## 10. Open questions

Settled since first draft — see §12.5: UX/interaction model, buffer pin-model,
keybindings, body-emptying (named-return rewrite), scramble scope (comments +
example/default; no enums), import pruning (pure-AST removal, §7.6), export sheet
+ GitHub-post handoff. Still open:

- ⬜ **v1 cut** (§5) — L0 only vs L0+L2. (Lean: L0 milestone-1, L2 fast-follow.)
- ⬜ **Public entry shape.** `codescan.Scramble(*ScrambleOptions)
  (*ScrambleResult, error)` returning in-memory files + verdict, with archiving
  as a separate helper? Confirm the result type (files map? `io.Writer` for the
  tgz?).
- ⬜ **L2 prose/directive discrimination** (§7.2) — accessor (precise, parser-
  coupled) vs heuristic (portable). Possibly: heuristic now, accessor post-rebase.
- ⬜ **Gibberish style** (Fred: "not sure yet") — lorem-ipsum, random
  pronounceable words, or hashed-but-readable. Diff-readability vs leak-resistance
  for descriptions/values.
- ⬜ **Multi-package repros.** Keep package structure (faithful; import paths may
  be scanner-relevant) vs flatten to one package (smaller). Lean: keep structure.
- ⬜ **Reachability rules** — enumerate exactly what the closure must follow to
  guarantee the byte-identical-spec invariant; cross-check against the schema
  builder's walk.
- ⬜ **Verification strictness** — exact JSON equality vs structural. `spec`
  marshals deterministically, so exact is likely fine; confirm.
- ⬜ **Crash-repro path** (§2 caveat) — assert same panic/error class.

---

## 12. Interaction model — TUI "issue-report mode" (Fred's proposal + my conclusions)

🟡 **Debate in progress.** This section is the canonical UX spec once settled.
Below: Fred's proposed model (cleaned up), my resolutions to the TBDs, and the
decisions still to take. Read it, then comment.

### 12.0 What the maintainer needs (the north star)

Four things make a codescan issue analyzable; the tool must make all four cheap
for the reporter to provide:

1. **the code excerpt** (minimized + scrambled),
2. **the generated spec** (if any),
3. **the diagnostics**,
4. **the reporter's remarks** tied to a *specific node* of the generated spec.

Every UX choice below serves "give the reporter maximum ease to provide all four
without leaking their domain."

### 12.1 The model (as proposed, lightly cleaned)

- **Enter/leave:** `Ctrl+I` toggles **issue-report mode**. On enter: **all
  in-scope source is buffered in memory, file watchers OFF, tree frozen.** On
  leave: buffer freed, tree refreshed, watchers back on. (With an exit guard if
  there are unsent remarks/edits — §12.4.6.)
- **Mode is loud** ✅: a **bright blue persistent banner** so the hidden state —
  watchers off, editing a buffer not disk — is never a surprise:
  `● ISSUE REPORT MODE   [m]inimize  [s]cramble   Ctrl-X export   Ctrl-I exit (edits lost)`.
- **Two transforms, on the left-panel "source" header as checkboxes:** ✅
  `[ ] Minimize   [ ] Scramble`. Toggle with **`m` / `s` directly** (*not*
  `Ctrl+M`: `Ctrl+M` is `Enter`/`\r` in a terminal and can't be reliably
  distinguished; direct mnemonics are also more discoverable).
- **Code stays editable** in the text areas; the spec (and diagnostics)
  **regenerate live on every buffer change** — so the reporter can *watch the bug
  persist* through minimize/scramble/edits. This live regen IS the verification
  surface (better than a trust badge).
- **Spec becomes annotatable** (element #4) ✅ — interaction settled:
  - `Ctrl+S` on the current spec line: **create** a remark (opens the textarea
    popup) **or delete** the existing one on that line.
  - `Enter` on a line that already has a remark: **open** its textarea popup.
  - `Esc` / focus-change: **close** the popup.
  - **Empty textarea ⇒ no comment created**; emptying an existing remark **deletes
    it** (so create-then-leave-blank and clear-to-delete are the same gesture).
  - Remarks render in the **left margin as side-notes** (💬). *Limitation today:*
    one remark per line (margin side-notes don't stack). Each is stored as a
    string keyed by the **jsonpointer of the nearest spec node** to the line — so
    it survives the JSON/YAML toggle and reformatting. The anchor can be *slightly*
    off (line ≠ exact node), but in practice fine on YAML and indented JSON where
    one node ≈ one line.
- **Export:** `Ctrl+X` (§12.4).

### 12.2 Buffer state machine — THE knot to settle first

Minimize is *destructive* (empties bodies, drops files) yet drawn as a
*reversible checkbox*, and the reporter may *also* hand-edit. Those collide. My
proposed resolution — the **pin model**:

- **`pristine`** — the buffer captured on mode entry; immutable.
- **`working = scramble(minimize(pristine))`** — re-derived purely whenever a
  toggle flips. Machine-owned.
- **The instant the reporter hand-edits a file, that file is *pinned*** — their
  text wins verbatim, is no longer auto-transformed, and is shared as-is (shown
  with a marker). Toggling minimize/scramble re-derives only the *un-pinned*
  files.

Predictable ("if you touched it, you own it"), reuses the dirty-marker idea, and
avoids the "did my edit survive the toggle?" trap.

✅ **Settled (Fred):** double-buffering with an immutable `pristine` is the model.
On `Ctrl+I` *exit*, the buffers (pristine + working + pins + remarks) are freed and
**the app reloads from disk** — back to normal watch/render. (So the mode is fully
ephemeral: nothing it does persists unless exported.)

### 12.3 The two transforms (TBDs resolved)

**Minimize** (order matters):
1. **Empty function bodies via named-return rewrite** ✅ (Fred). Requires
   **return analysis**: rewrite the signature to name the results and replace the
   body with a bare `return`, which Go zero-initializes — no per-type zero-value
   synthesis:
   ```go
   func (x int, y int) (v1 *Type, v2 error) { return }   // unnamed results → synthesized names
   func (x int) (out Foo)                  { return }    // already-named results → keep names
   func (x int)                            {}            // no results → empty body
   ```
   Synthesized names (`v1, v2…`) must avoid colliding with param names. **Always
   keep method *signatures*** (`MarshalText` etc. — presence drives
   `IsTextMarshaler`; the live spec visibly flips string↔object if mis-stripped;
   named vs unnamed results are the *same type*, so interface satisfaction holds).
2. **Drop unreachable decls** — not transitively referenced from an annotation
   root. *This removes whole files* — surface it (`minimized 42→6 files`; tree
   shows survivors only). ⚠️ Fred: genuinely the complex part — it forces an
   import-rewrite pass for correctness (next point).
3. **Drop now-unused imports**, emptied files, and `_test.go`. Repro must stay
   type-checkable (codescan needs `go/types` to pass).
   - ⚠️ **WASM constraint (Fred): do NOT use `goimports` / `golang.org/x/tools/
     imports` — it shells out to the `go` binary, which can't ship to WASM.**
     We only ever *remove* imports (pruning never adds), so we need just the
     **pure-AST removal half**: collect the package identifiers actually used in
     the surviving AST, delete `ast.ImportSpec`s whose bound name isn't among them
     (handling unnamed imports via path→assumed-name). No toolchain, WASM-clean.
   - 📎 **Reference (go-swagger):** `generator/internal/language/format_lite.go` —
     `fixImports` / `deleteImportSpec` / `collectTopNames` /
     `importPathToAssumedName` (the latter copied from x/tools internals) are
     exactly this pure-AST logic. Borrow the *removal* path; skip its final
     `imports.Process` sort (that's the tooling-coupled part — a trivial lexical
     sort suffices, or let the host gofmt it natively when not WASM).

**Scramble** — prose only, structure preserved:
- Per comment, tokenize → replace each alphabetic word with a **deterministic**
  pseudo-word (seeded by the word → repeats map consistently, live spec stable),
  preserving **length-class, capitalization pattern, punctuation, whitespace,
  line breaks.** Still *reads* like prose ("load-bearing, not `*****`").
- **Must not corrupt annotation syntax.** `swagger:` keywords + directive
  keys/args (`minimum:`, `in: body`, `required:`, type names) are sacred; only
  *prose* is gibberished. Prose lives in two places: plain Go doc-comments, and
  the **title/description/summary** fields *inside* annotation blocks (scramble
  those; keep their sibling directives verbatim).
- Needs the **tiny annotation classifier** (§6 `scan.go`): per comment line →
  keyword / `key:value` directive / prose. Prose → scramble.
- **Safety net:** misclassify → the live spec/diagnostics visibly break → the
  reporter sees it.
- ✅ **Scramble scope settled (Fred):** comments **+ `example`/`default` string
  values**. **Enum values are NOT scrambled for now** (kept real). Revisit only if
  a real case shows enum PII.

### 12.4 Export (`Ctrl+X`) — the other TBD

An **export sheet** (✅ modal) assembling all four elements:

1. **Bundle contents** (all default-on, toggleable): ① code → `repro/` tree as
   `repro.tgz` *or* **inline fenced block** when small; ② the spec (on-screen
   format, or both); ③ the diagnostics; ④ the remarks as a readable list —
   `at #/definitions/Order/properties/createdAt: «remark»`.
2. **Title + summary field** — the issue-level "what's wrong / what I expected."
3. **Output — escalating handoff:**
   - **Always:** write `repro.tgz` (show path) **+** copy a **prefilled Markdown
     issue body** to clipboard (title, summary, environment = codescan version +
     exact `Options`, remarks list, diagnostics, spec in a collapsible block).
   - **`[o]` open** `issues/new?title=…&body=…` prefilled in the browser.
   - **`[p]` post directly** ✅ (Fred — "good idea"): create the issue on
     `go-openapi/codescan` via the GitHub API. Prefer the **`gh` CLI**
     (`gh issue create`) so it reuses the user's existing auth; fall back to a PAT
     (`GITHUB_TOKEN`). In WASM, a `fetch()` POST with the user's token.
   - ⚠️ **Attachment caveat:** the GitHub API can't attach a file to an issue.
     **The minimize pass is what saves us** — a pruned repro is usually small
     enough to **inline as a fenced ```go block in the body**, so the one-click
     POST is fully self-contained. Larger repros: link a **gist** (one extra API
     call) or fall back to "tgz saved locally — drag it into the issue."
4. **Stay in the mode after export** (re-export after tweaks); exit only `Ctrl+I`.
5. Nice touch: render remarks **inline in the YAML** as `# «remark»` at the right
   node (YAML allows comments; JSON doesn't) — maintainer reads them in context.
6. **Exit guard** ✅: leaving the mode with unsent remarks/edits confirms
   ("discard issue report?").


### 12.5 Decisions — now settled

- ✅ **Buffer state machine / pin model** (§12.2) — double-buffer with immutable
  `pristine`; pin on hand-edit; free + reload-from-disk on exit.
- ✅ **Scramble scope** (§12.3) — comments + `example`/`default` strings; **no
  enums** for now.
- ✅ **Keybindings** — `m`/`s` direct toggles (no `Ctrl+M`); pane-scoped `Ctrl+S`
  for spec remarks; `Enter` opens an existing remark; editor edits are
  live-in-buffer (no save in-mode).
- ✅ **Minimize aggressiveness surfaced** — file-count delta + survivors-only
  tree; whole-file pruning is the default.
- ⚠️ **Carried as a real build constraint, not a UX choice:** the import pass must
  be **pure-AST removal** (no `goimports`/`go` binary) to keep the WASM target —
  see §12.3 step 3 + the go-swagger reference, and §7.6.

### 12.6 Free synergies / notes

- **Spec line → jsonpointer index** (needed for remark anchoring) is the
  **spec-side half of the future cross-ref linker** — and it is **not
  rebase-gated** (grammar2 gates only the *source* side). Building remarks buys
  down the linker.
- **The §2 oracle keeps a residual role**, orthogonal to the user's bug: run it
  silently, speak only at export — `⚠️ Minimize altered your spec beyond stripping
  code — review before sending`. Catches an over-aggressive prune.
- **Orphaned remarks:** if a hand-edit deletes a node a remark targets, show it as
  orphaned — never silently drop.
- **Scale:** "buffer all in-scope source" can be heavy on large trees (the docker-
  API stress case). Consider prompting to narrow scope on mode entry when the tree
  is large; minimize then shrinks it anyway. ([[project_genspec_tui]] scale note.)

---

## 13. Evidence / references

- `wasm-playground.md` §11 — origin of the idea.
- `internal/scanner/index.go` — `TypeIndex` classification = the annotation roots.
- `internal/scanner/declaration.go` — `EntityDecl` (decl ↔ file/pkg) for the
  closure walk.
- `internal/builders/resolvers/assertions.go:77,96,104` — name/method-based
  recognizers that constrain the sacred set (§3).
- `internal/parsers/sectioned_parser.go` — blank-line/layout sensitivity (L2).
- Related: [[project_wasm_playground]] (shared WASM-clean core),
  [[project_genspec_tui]] (host front-end), [[project_lsp_diagnostics_target]]
  (position fidelity feeds L2 precision), [[feedback_test_api_surface]]
  (keep public surface minimal).
