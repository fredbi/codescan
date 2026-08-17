# Fix-quirks sequence

Date: 2026-06-03
Branch: `fix/quirks` (off master, post Stream-M merge)
Status: ✅ all phases complete (closed 2026-06-11, PR #32 merged).

Tracking doc for the post-Stream-M quirks-cleanup pass. Source-of-truth
for the individual Q-descriptions is
[`observed-quirks.md`](observed-quirks.md); this doc is the **execution
order**, decision log, and progress board. We update Q-status in
`observed-quirks.md` as each lands; here we track the *flow*.

The companion holding-area is [`deferred-quirks.md`](deferred-quirks.md)
(D-numbers shadow some Q-numbers). Anything we re-defer rather than fix
moves there.

## Why this doc exists

The quirks-cleanup flow has a known failure mode: a single Q-fix turns
into a multi-day sub-investigation, and when we surface the original
ordering and remaining items are lost. This doc is the anchor — every
time we finish a sub-plan or get pulled into a workshop, the next step
is "back to fix-quirks.md, pick the next task."

## Ground rules (carry-overs from M-stream)

- **One commit per Q** (DCO `-s`, `Co-Authored-By: Claude`). Fix + golden
  in the same commit.
- **Witness-then-fix** for every code change (Stream-M discipline).
  Capture pre-change golden where useful, regenerate post-fix, the
  golden *diff* is the audit trail. Especially load-bearing for schema-
  discovery accumulation paths — see
  `feedback_schema_discovery_verify_with_witness.md`.
- **`UPDATE_GOLDEN=1 go test ./...`** to regenerate; lint via
  `golangci-lint run --new-from-rev master` before push.
- Status icons: ⬜ todo, 🟡 in progress, ✅ done, 🔵 paused/blocked,
  ⏸ re-deferred to `deferred-quirks.md`.

## Phase A — Grammar/YAML strict-mode cluster

Localized fixes inside `internal/parsers/grammar`. Low blast radius.
Q28 is breaking and tops the priority list.

### ✅ A1 — Q28: `swagger:meta` `SecurityDefinitions` strict-YAML failure

**Landed 2026-06-03.** Turned out to be two bugs riding the same
fixture, not one — the dedupe layer was needed and useful, but the
*classification* repro's root cause was a lexer indent-stripping bug
that only surfaced once the lib's strict check stopped masking it.
Both fixes shipped together; see Q28 in `observed-quirks.md` for the
full breakdown. Sibling bug discovered during witness construction
(blank-line-before-first-body-line trips RemoveIndent's first-line
anchor) is deferred — no user repro yet.

**Symptom.** `codescan.Run` aborts with `yaml: unmarshal errors: …
mapping key "type" already defined` when scanning
`./fixtures/goparsing/classification` standalone. Triggered by
`doc.go`'s tab+space-mixed `SecurityDefinitions` block under grammar's
strict `go.yaml.in/yaml/v3` decoder.

**Library investigation (2026-06-03).** `go.yaml.in/yaml/v3` has NO
public knob to disable duplicate-key detection. `decoder.uniqueKeys`
is hardcoded `true` in `newDecoder()` (`decode.go:344`). The only
public option on `*Decoder` is `KnownFields(bool)` — a different
strictness mode (unknown-fields-in-struct).

The library does offer a back door: decoding into `*yaml.Node` (the
AST) bypasses `mapping()` entirely (`decode.go:492`), so we can always
parse, walk the AST, dedupe, and re-decode into the caller's target.
`swag/yamlutils.BytesToYAMLDoc` already does the AST-decode shape for
the same reason.

**Decision (2026-06-03).** Silent dedupe — no diagnostic capture in
this fix. Implementation:

1. Parse body to `*yaml.Node` (back-door bypass).
2. Walk mapping nodes recursively; when two children share a key,
   drop the *earlier* occurrence in place (last-wins, matching the
   v1 `gopkg.in/yaml.v2` behaviour).
3. Re-decode the cleaned node into the caller's target.

Diagnostic-emit-on-duplicate-key is **deferred** to the yaml-library
swap tracked in
[`forthcoming-features.md`](../forthcoming-features.md) §3.1
(`goccy/go-yaml` POC). When that library lands — driven by LSP's need
for per-token positions — duplicate-key warnings come along naturally
because we'll already be walking tokens with position metadata. No
point building two parallel diagnostic infras.

**Tasks:**
1. Build the silent-dedupe helper in `internal/parsers/yaml/`.
   Likely shape: a `dedupeMapping(*yaml.Node)` recursive walker +
   wrap the existing `Parse` / `ParseInto` / `TypedExtensions` to
   call it before final decode. Keep the public surface unchanged —
   callers don't opt in, dedupe always happens.
2. Unit-test the dedupe walker — empty body, nested mappings,
   sequence-of-mappings, deeply-nested dups, scalar-only bodies.
3. Re-run `./goparsing/classification` standalone (via a one-off
   `codescan.Run` test or the scrambler oracle path) to confirm
   green.
4. Add integration witness:
   `fixtures/enhancements/meta-securitydefs-duplicate-keys/` —
   minimal meta block with intentional duplicate key in
   `SecurityDefinitions`; golden shows last-wins dedupe applied.
5. Update Q28 status in `observed-quirks.md` to RESOLVED with commit
   ref, plus a forward-reference to the §3.1 future-feature for the
   diagnostic-emit half.

**Files in scope:**
`internal/parsers/yaml/yaml.go`,
`internal/parsers/yaml/operation.go`,
plus tests + the new witness fixture.

### ✅ A2 — Q26: TOS raw-block absorbs adjacent `Schemes:` keyword

**Landed 2026-06-03 as a regression-detector golden (`6e09632`).**
Building the witness fixture revealed Q26 no longer reproduces —
verified directly against the upstream `go-swagger/examples/generated`
source. Some Stream-M commit closed it (not isolated, probably the
M6.5-D terminator audit). The fixture
`fixtures/enhancements/meta-tos-schemes-terminator/` + integration
test land the corrected behaviour as a golden so future regressions
in the terminator rule turn red. Q26 marked RESOLVED.

**Symptom.** Petstore meta produces
`"termsOfService": "http://…/terms/\nSchemes:\nhttp"`. The lexer's
`isSiblingTerminatorFor` is missing the TOS↔Schemes pair (both
asRawBlock since M6.5-D's KwSchemes widening).

**Tasks:**
1. Reproduce with a minimal fixture under
   `fixtures/enhancements/meta-tos-schemes-terminator/`. Meta block
   with `Terms Of Service: …` immediately followed by `Schemes: http`.
   Capture pre-fix golden as a witness (absorption shape locked).
2. Audit the sibling-terminator table in
   `internal/parsers/grammar/lexer.go` (or wherever the rule actually
   lives — `isSiblingTerminatorFor`). Determine whether the rule is
   pair-listed, head-listed, or computed from a keyword attribute.
3. Add the missing terminator relation; consider whether *every* pair
   of asRawBlock siblings in the meta scope needs it (likely yes — one
   table edit covering N pairs is better than N spot-fixes).
4. Regenerate goldens; the witness fixture should now show TOS body
   terminated cleanly before `Schemes:`.
5. Update Q26 status in `observed-quirks.md`.

### ✅ A3 — Q29: `in:` case-sensitive comparison

**Landed 2026-06-03.** Three sites identified
(parameters/doc_signals, responses/doc_signals, routebody/parameters)
— all doing strict-case map lookups. Single canonical helper
`grammar.NormalizeIn(raw, allowFormAlias) (string, bool)` landed in
`internal/parsers/grammar/` (where KwIn's closed vocabulary already
lives). All three consumers route through it; Q27's `form →
formData` v1 affordance stays contained to routebody via
`allowFormAlias=true`. Pre-fix witness golden showed every
mixed-case parameter collapsing to the `query` default; post-fix
each lands at its canonical location. See Q29 in
`observed-quirks.md` for the full implementation breakdown.

## Phase B — Interface / embed-asymmetry cluster

**Re-prioritised 2026-06-03.** Originally these two sat in a
documentation-only Phase C ("intentional behaviour, just document
it"). Promoting them ahead of the alias workshop because:

- Q9 is likely a quick code-fix win (read a struct tag from the
  embedding context — known-shape).
- Q8 may turn out meatier than expected — but is still bounded to
  the schema builder's struct-vs-interface dispatch, *not*
  cross-cutting like the alias cluster.
- Both inform what we know about embed/interface boundary handling
  before the alias workshop opens, so the workshop has a fuller
  surface to reason about (an embedded *named alias* in a struct
  touches both clusters).

### ✅ B1 — Q9: interface-method property naming ignores JSON conventions

**Landed 2026-06-03.** Q9's description in `observed-quirks.md`
turned out to be stale — empirical inspection during the B1
investigation revealed interface methods DO auto-jsonify (via
`swag/mangling.NameMangler.ToJSONName`) and have done so since
somewhere in Stream M development. The status update was missed
then; corrected now.

The real story is a **principled asymmetry**: struct fields mirror
`encoding/json`'s tag-or-verbatim rule (so the spec matches what
`json.Marshal` actually produces), interface methods auto-jsonify
(no runtime serialization to mirror, so codescan invents a default
JSON name). Documented in `internal/builders/schema/README.md`
§method-mangler.

Smaller-than-expected deliverable:
- Witness fixture `fixtures/enhancements/interface-name-verbatim/`
  pinning the `swagger:name X` verbatim contract (PascalCase,
  snake_case, SCREAMING_CASE, hyphenated user inputs all reach the
  spec literally; the mangler is bypassed when JSONName is set).
- Doc note in schema README §method-mangler explaining the
  principled asymmetry and the verbatim guarantee.
- `forthcoming-features.md` §3.3 — one-size-fits-all opt-out
  (`SkipJSONifyInterfaceMethods` global) for v2 when a user asks.
- Q9 status corrected.

**Symptom.** Struct fields snake-case via `json:"xxx"` tags;
interface-method properties stay PascalCase because methods can't
carry struct tags. Only `swagger:name` renames them. Surprising for
an API expected to behave consistently with the struct path.

**Witness:** `fixtures/integration/golden/enhancements_interface_methods.json`
— properties `ID`, `Email`, `Bio`, `Profile`, `Tags` (PascalCase)
plus `fullName` (renamed via `swagger:name`).

**Fix shape (preliminary).** When an interface is being inlined into
the embedding struct, inherit a json-tag-style convention from the
embedder's existing fields if one is consistent — e.g. if the
struct's other fields all snake-case, snake-case the interface
methods. Where this falls down: the interface might be embedded by
multiple structs with conflicting conventions; or by a top-level
swagger:model with no other fields to derive the convention from.

**Tasks:**
1. Read `internal/builders/schema/` interface-walker code; find
   the property-name assignment site.
2. Investigate: what conventions can we derive? (Snake-case-ness
   from sibling fields? A package-level option? An interface-side
   `swagger:name` directive at the method-block level?)
3. Build witness fixture that exercises the desired rename
   behaviour and an edge case (interface embedded by two structs
   with different conventions).
4. Implement; regenerate goldens.
5. Update Q9 status in `observed-quirks.md`.

### ✅ B2 — Q8: re-framed and reclassified into the alias cluster

**Landed 2026-06-03 as a re-classification.** Empirical probe of
five embed shapes (`Base` direct / `BaseAlias` / `*Base` / `Iface`
named interface / anon interface) revealed Q8's original framing
("struct embed → allOf vs interface embed → flat") was wrong. The
actual dispatch (`embedded.go:40-53`) inlines for every named-type
case (struct OR interface); the alias path is the outlier producing
`$ref` + `allOf`.

**Real question:** in Go, `type BaseAlias = Base` makes them
literally indistinguishable types. So why does `struct { BaseAlias }`
emit a different schema shape than `struct { Base }`? This is
exactly the "alias is an unfinished job" framing — alias-cluster
territory.

Q8 reclassified into Phase C with the rest of the alias cluster.
Witness fixture (adding direct-named-struct case to embedded-types)
deferred to C0 so the new golden serves the workshop's decision
rather than locking pre-workshop behaviour. See
`observed-quirks.md` Q8 for the corrected dispatch table.

## Phase C — Alias-theme cluster

**Pause before code. Workshop first.** Q3, Q7, Q8, Q11, Q12, Q13 all
answer the same underlying question: *what does `type X = Y` mean
for a Swagger surface that has no notion of type aliases?* Fixing
Q7 alone may pre-commit a model that wrong-foots the rest.

**Cluster scope grew during pre-merge audit (2026-06-03):**
- Q8 reclassified out of Phase B after a probe showed the original
  "struct vs interface" framing was wrong (real asymmetry is
  alias-embed vs direct-embed).
- Q3 reclassified out of "IMPROVED" after a three-mode probe
  revealed the original date-time-loss bug is still present in
  default Expand mode (only RefAliases was improved). Same root
  cause as Q13 (`buildDeclAlias` Expand branch walks `Underlying`
  without consulting `applyStdlibSpecials`).

**Fred's framing (2026-06-03):** this cluster is essentially an
**unfinished job**. The alias model landed cleanly for the schema
builder during Stream M but not for the allOf path, the parameters
builder, or the responses builder. The workshop's job is to surface
the model the schema builder uses and decide whether to propagate it
to the other three sites, or design a different model that fits all
four uniformly.

After B1 and B2 land, **re-read this section before opening C0** —
what we learn about embed/interface handling in B2 likely informs
the alias decision (embedded named-alias-of-struct touches both).

### ✅ C0 — Alias-handling workshop (closed 2026-06-11, PR #32)

**Goal.** Produce a one-page model in
`.claude/plans/workshops/alias-handling.md` that answers:

1. What is the *intent* of `type X = Y` on each target:
   `swagger:model`, `swagger:parameters`, `swagger:response`,
   field-type, **embed-site (Q8 input)**, unannotated?
2. How do the three modes (default-expand, `RefAliases`,
   `TransparentAliases`) relate to that intent? Is one mode the
   canonical answer and the others a back-compat affordance?
3. What does the schema builder do today (the "landed right"
   side, per Fred)? Mirror it where applicable, or design a
   different model that fits parameters/responses/allOf too?
4. For each of Q3/Q7/Q8/Q11/Q12/Q13, what does the answer dictate?

**Specific Q8 question:** in Go, `type BaseAlias = Base` is a
transparent rename — the two are literally indistinguishable
types. `embedded.go`'s three-arm dispatch makes them produce
*different* embed shapes (`Base` → flat inline,
`BaseAlias` → `allOf` with `$ref`). Should the alias path mirror
the direct-named path (flat), or should the direct-named path
mirror the alias path (allOf), or is the asymmetry intentional?
Companion to Q7 (alias-expand parameter envelope) and Q11
(alias-expand layer duplication) — all three are different
manifestations of "the alias path took a different turn from the
named-direct path during Stream M."

**Out:** model writeup + per-Q action items + the fixture roster we
need before code. For Q8 specifically, the roster needs a
direct-named-struct embed case added to the embedded-types
fixture so the alias-vs-direct comparison is visible in the
golden.

### ✅ C1 — Q7: alias-expand parameters lose envelope (closed 2026-06-11, PR #32, commit `c896cc7`)

**Symptom.** `swagger:parameters` on top-level alias in default mode
emits the alias as a plain `definitions` object instead of merging into
the operation's parameters. `data` loses `in: body`, `search` loses
`in: query`.

**Blocker:** C0 must land first — the fix needs the alias model.

**Tasks (preliminary):**
1. Verify witness fixture `fixtures/enhancements/alias-expand/` still
   triggers per current goldens.
2. A/B against `RefAliases=true` (working path) — golden
   `enhancements_alias_ref.json` is the comparator.
3. Implement per C0 decision.
4. Witness-then-fix: capture pre-fix golden, apply fix, regenerate,
   diff is the audit trail.

### ✅ C2 — Q12: unexported parameter-alias backing struct leaks (closed 2026-06-11, PR #32, commits `c896cc7` + `d84e650`)

**Symptom.** Lowercase-named struct (`exportedParams`) declared only as
the target of an exported alias still ends up in `definitions`.

**Likely couples with C1** — if Q7's fix routes alias parameters
through the parameter builder, the unexported target may stop being
reachable as a schema. Re-check after C1; may collapse to zero work.

### ✅ C3 — Q3 + Q13 (closed 2026-06-11, PR #32, commits `9e8a6e8` + `c9eabd9`)

**Joint task — same root cause.** Both Q3 (`type Timestamp =
time.Time` model emits `{type: object}`, losing date-time format)
and Q13 (`type X = any` model emits `{title, x-go-package}` only)
land in the same `schema.go:168-170` Expand branch — the path walks
`Underlying()` without consulting `applyStdlibSpecials` first. The
patch is one `if applyStdlibSpecials(...) { return nil }` check
before the fallthrough.

The *semantics* — what each stdlib-special should emit in Expand
mode — is the workshop's call (see workshop §3.5/§5.5):

- Q3 candidates: `{type: string, format: date-time}` inline /
  RefAliases-style chain / refuse with diagnostic.
- Q13 candidates: omit the definition / `{type: object}` /
  `{type: object, additionalProperties: true}`.

**Tasks (preliminary):**
1. Implement per C0 §5.5 decision.
2. Add witnesses:
   - Q3: default-mode case for `type Timestamp = time.Time` (the
     existing `ref-alias-chain` fixture only covers RefAliases;
     needs a no-flags companion so the golden visibly carries the
     bug pre-fix).
   - Q13: review the existing `Wildcard` golden against the new
     decision; may also need a default-mode companion.
3. Regenerate goldens; ensure no downstream consumer relied on the
   blank stub (`grep -r "additionalProperties\\|object.*Wildcard"`
   across the fixture tree).
4. Update Q3 and Q13 statuses in `observed-quirks.md` with the
   commit ref.

### ✅ C4 — Q11: alias-expand layer duplication (closed-no-action 2026-06-10; consistent with documented mode behaviour)

**Symptom.** Three-link chain `A = B = Target` produces three
definitions; M-stream made the alias layers use `$ref` chains instead of
full struct copies, but the count of definitions is unchanged.

**Tasks (preliminary):**
1. Per C0: decide whether N alias layers should produce 1 definition
   (collapsed) or N (chain visible). The current "N refs" middle ground
   may be the right answer; the residual quirk is byte-cost, not
   semantic.
2. If collapse: implement + regen; if keep: close as "intentional,
   documented" + update `observed-quirks.md` accordingly.

## Digressions ledger

When we get pulled off this plan into a sub-investigation, log it here
so the trail back is obvious. One line per digression:

- *(none yet)*

## Wrap-up checklist

When all phases land:

- [ ] `observed-quirks.md` has no STILL PRESENT / PARTIALLY RESOLVED
  entries left from the original baseline list (Q26, Q28, Q29 closed
  in A; Q9 closed in B; Q3/Q7/Q8/Q11/Q12/Q13 closed or re-deferred
  in C).
- [ ] `deferred-quirks.md` updated for anything moved.
- [ ] Memory: append a "fix-quirks wave complete" line to
  `project_refactor_golden_payoff.md` velocity ledger.
- [ ] Open PR to master; reference each Q-number in the body.

## Phase D — Final wrap-up pass (post-Phase-C, pre-PR)

**Reminder logged 2026-06-04 (Fred):** before opening the alias-work
PR, do a pass to detach the disclosed artifacts (commits + production
code comments + READMEs) from the private workshop vocabulary (Q-
numbers, R-numbers, "cycle N", "fix-quirks", "W3", workshop ledger,
`.claude/plans/` paths). The `.claude/plans/` artifacts stay private;
the public-facing surface should read clean to a first-time reader
from GitHub.

### ✅ D1 — Inline-comment squashing & promotion to READMEs (closed 2026-06-11, PR #32 commit `599e453`)

Several patches landed during the fix-quirks + alias work carry
narrative WHY-comments inline ("R5 in the W3 alias workshop ledger",
"Q-C resolution", "fix-quirks C3 task", etc.). Walk the codebase:

- Inline comments referencing private vocabulary → either delete,
  reword as standalone rationale, or hoist to the package README's
  long-form section (the schema/parameters/responses READMEs already
  follow the "short godoc on the symbol, long form in README" split
  per CLAUDE.md).
- Comments documenting actual behaviour or design constraints (not
  workshop archaeology) stay — these are the load-bearing ones a
  future reader needs.
- Where a README section explains the new behaviour, drop a
  pointer from the symbol's godoc instead of re-explaining inline.

### ✅ D2 — README sweep for obsolete quirk references (closed 2026-06-11, PR #32 commit `599e453`)

Scan `internal/builders/*/README.md` and any other in-repo docs for:

- References to specific Q-numbers (Q1-Q30) — these will mean
  nothing to an external reader; replace with the behaviour name or
  delete.
- References to "fix-quirks", "Stream M", "M6.x", "W3", "cycle N"
  — internal vocabulary, scrub.
- References to closed quirks as if they were still open — update
  to past tense or remove.
- Pointers into `.claude/plans/` — these paths are gitignored, so a
  GitHub reader can't follow them; replace with whatever rationale
  lives at the destination, inline.

### ✅ D3 — Commit message reword pass (closed 2026-06-11, squashed 18 commits → 7 logical units before PR #32 merge)

Most of the fix-quirks + alias-work commits reference internal
vocabulary in their messages (Q-C / Q-D / R5 / R5-scope / "cycle 2"
/ "W3 alias workshop" / ".claude/plans/" paths). Interactive rebase
to reword them so each commit's message reads as a self-contained
explanation of what changed and why, to a first-time reader from
GitHub.

Suggested rephrasing patterns:
- "R5 in the W3 alias workshop ledger" → "the
  applyStdlibSpecials-bypass family of bugs"
- "Q-C" → "the chain-vs-inline asymmetry for stdlib aliases"
- "Q-D" → "the aliased-embed promoted to allOf without annotation
  bug"
- "fix-quirks C3" → just describe what the patch does

Don't lose the technical detail — the goal is to make the commits
durable as PR-ready history, not strip them. Preserve file paths,
line numbers, and exact behaviour deltas; replace only the
internal-codification vocabulary.

### ✅ D4 — `.claude/plans/` and `.claude/memory/` privacy guard (closed 2026-06-11; `.claude/.gitignore` excludes `plans/`, `skills/`, `commands/`, `agents/`, `hooks/`; only disclosed files are CLAUDE.md and rules/)

Final sanity check: the `.claude/` tree is gitignored in this repo.
Confirm with `git check-ignore .claude/` that nothing under it has
been inadvertently committed during the work. The workshop docs,
ledgers, and memory files are deliberately private.
