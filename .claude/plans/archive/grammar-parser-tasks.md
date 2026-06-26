# Grammar parser — task-level plan

Date: 2026-04-21
Status: **P0 → P6 complete.** Grammar parser is the single path;
legacy SectionedParser and regex taggers are removed. P7
(fuzz / property-based hardening) is the only remaining phase.
Companion: `.claude/plans/grammar-parser-architecture.md` (the "why")
Modeled on: `.claude/plans/observed-quirks.md` (the Q-pass pattern)
Living catalog of parked items: `.claude/plans/forthcoming-features.md`

This document enumerates the concrete work to replace the regexp-based
annotation parser with the grammar-based one agreed in the architecture
doc. Each task is sized to a single commit; phase groupings are the
review checkpoints.

---

## 0. Overview

Six phases on a long-lived feature branch (`feat/grammar-parser`). Each
phase has an exit criterion; no phase advances until the previous one's
exit criterion is green. Cutover (P6) is a single merge commit to
master (per §4.5 of the architecture doc).

| Phase | Milestone | Status |
|-------|-----------|--------|
| P0 | Skeleton + keyword table + docs-gen | 🟢 |
| P1 | Core parser (preprocess → lex → parse → typed `Block`) | 🟢 |
| P2 | Sub-language isolation (YAML fence + sibling sub-parser) | 🟢 |
| P3 | Parser API (`Parser` interface, `ParseAs` hook) | 🟢 |
| P4 | Test infra (unit + parity harness) | 🟢 |
| P5 | Builder migration — 5 sub-steps | 🟢 |
| P6 | Cutover (delete old, promote package, single release) | 🟢 |
| P6.1 | Meta builder migration (originally bundled into P6) | 🟢 |
| P7 | Hardening — fuzz + property-based (post-cutover) | 🟦 |

Status icons: 🟦 not started · 🟡 in progress · 🟢 done · 🔴 blocked
(project-planning-conventions).

**Parity principle.** P4's parity harness (old parser vs new parser →
same spec on every fixture) is the safety net throughout P5. No
builder flips until parity passes for its annotation kinds.

**Commit cadence.** Tasks are sized to one commit each. Phase commits
land sequentially; P5 migration commits land per-builder. All with
DCO sign-off per `.claude/rules/contributions.md`.


> NOTES(fred):
>
> * we may defer the fuzz part and property based part to a "hardening phase" (either P5+ or P7)
> * P5 is decomposed by builder. I see 5 sub-steps.
>   * schema, operations, routes, parameters, responses
>   * the spec builder is an orchestrator - I don't think it is directly affected by the migration
>
> **Settled: fuzz + property-based moved to new P7 hardening phase
> (post-cutover, non-blocking for v2.0 ship). P5 reduced to 5 sub-steps;
> `items` rides with `schema`, `spec` is the orchestrator and its
> `swagger:meta` wiring is bundled into P6 plumbing.**

---

## Phase 0 — Skeleton and foundations

**Goal:** land the directory structure, the authored keyword table,
and the docs generator. No behavior yet; purely scaffolding.

### Tasks

- [x] **P0.1** 🟢 Create `internal/parsers/grammar/` with placeholder
      files (`preprocess.go`, `lexer.go`, `parser.go`, `ast.go`,
      `diagnostic.go`, `style.go`). Each file carries a package
      comment and `// TODO: P1`.
- [x] **P0.2** 🟢 Author `keywords.go` — a typed `[]Keyword` slice at
      parity with the v1 keyword set. Each entry holds:
      canonical name; aliases; value type (per architecture §3.4:
      `number`/`integer`/`boolean`/`string-verbatim`/`comma-list`/
      `string-enum`); the set of annotation kinds under which it is
      legal (the context table from §2.2.1); short doc string.
      Seed list extracted from `internal/parsers/regexprs.go` and the
      existing taggers (~35 entries).

      **Opportunistic workshop touch-points during this task** (no
      separate stop needed; ~30 minutes each):
      - **W5 — `externalDocs` annotation syntax.** Land the keyword
        entry here. Single row, doesn't warrant a separate workshop.
      - **W7 — LSP context-validity tables.** Lightly review the
        "legal annotation kinds" column for each keyword as the data
        LSP will later consume. Seed from observed v1 behavior; no
        need to design LSP surface yet.
- [x] **P0.3** 🟢 Author `diagnostic.go` — `Diagnostic{Pos, Severity,
      Code, Message}`, `Severity` enum (Error/Warning/Hint), code
      convention (`parse.invalid-number`, `parse.unknown-keyword`,
      `parse.context-invalid`, …).
- [x] **P0.4** 🟢 Write `internal/parsers/grammar/gen/main.go` — reads
      `keywords.go` at `go generate` time and emits
      `docs/annotation-keywords.md` (auto-generated marker, no
      handwritten content).
- [x] **P0.5** 🟢 Wire `//go:generate` directive; commit the generated
      doc; add a CI check that regenerating leaves the file
      unchanged. (CI check = `TestGeneratedDocIsCurrent` in
      `gen/gen_test.go`; runs as part of the existing `go test ./...`
      workflow — no bespoke CI file needed.)

### Exit criteria

- `go build ./...` clean; skeleton compiles.
- `keywords.go` exposes the full v1 keyword set (verified by count
  against a list in `P0.2`'s commit message).
- `go generate ./...` regenerates `docs/annotation-keywords.md`
  byte-identical.

---

## Phase 1 — Core parser pipeline

**Goal:** consume an `*ast.CommentGroup` (or raw text, for LSP/test)
and produce a typed `Block`. No builder integration.

### Tasks

- [x] **P1.1** 🟢 `preprocess.go` — layer on top of
      `(*ast.CommentGroup).Text()`. Track positions line-by-line,
      strip markdown table pipes (`| foo | bar |`), preserve
      indentation inside fenced blocks. Output:
      `[]Line{Text string, Pos token.Position}`.
      (Fence-body indentation preservation deferred to P1.2 lexer
      where fence state is tracked — preprocessor stays stateless.)
- [x] **P1.2** 🟢 `lexer.go` — `[]Line` → token stream.
      Token kinds: `ANNOTATION`, `KEYWORD_VALUE`, `KEYWORD_BLOCK_HEAD`,
      `YAML_FENCE`, `TEXT`, `BLANK`, `EOF`. Keyword lookup against
      `keywords.go` (case-insensitive, alias-aware). `items.N.X`
      prefix expanded into a depth-tagged sub-path on the token.
      (Godoc-ident-prefix form "DoFoo swagger:route …" deferred to
      P1.4 parser — lexer handles start-of-line `swagger:` only.)
- [x] **P1.3** 🟢 `ast.go` — `Block` interface (`Pos`, `Title`,
      `Description`, `Diagnostics`, `AnnotationKind`); typed kinds
      (`ModelBlock`, `RouteBlock`, `OperationBlock`, `ParametersBlock`,
      `ResponseBlock`, `MetaBlock`, `UnboundBlock`); helper iterators
      (`Properties() iter.Seq[Property]`, `YAMLBlocks() iter.Seq[RawYAML]`,
      `Extensions() iter.Seq[Extension]`). `Property{Keyword, Value,
      Typed, Pos, ItemsDepth}` + `TypedValue`/`RawYAML`/`Extension`
      support types. `AnnotationKind` enum with `String()` /
      `AnnotationKindFromName()` round-trip.
- [x] **P1.4** 🟢 `parser.go` — recursive-descent envelope:
      annotation line → title → description → properties → sub-language
      heads. First-token dispatch picks the `Block` constructor
      (architecture §4.6). Error recovery: skip to next line on parse
      failure, collect diagnostic, never abort, never panic.
      (Also landed: preprocessor now preserves `-`, needed for YAML
      fence detection. Noted in the commit.)
- [x] **P1.5** 🟢 Primitive value-typing inside the parser: `number`,
      `integer`, `boolean`, `string-enum` converted at parse time per
      `keywords.go`. `string-verbatim`, `comma-list`, `raw-value`
      captured as strings (analyzer type-converts later — architecture
      §3.4). v1's `maximum: <5` / `minimum: >=0` operator form
      supported via new `TypedValue.Op` field; StringEnum canonicalises
      to the table spelling.
- [x] **P1.6** 🟢 Positional-args parsing for `swagger:route` /
      `swagger:operation` into semantic fields (method, path, tags,
      opID), per Fred's answer in §3.6 ("not just a lexer"). Also
      covers the godoc-ident-prefix form `Func swagger:route …`
      (v1 rxRoutePrefix) in the lexer — applies to `route` only,
      not other annotations.
- [x] **P1.7** 🟢 Context-validity diagnostic: if a recognized keyword
      appears under an annotation kind that doesn't permit it, emit
      `parse.context-invalid` (non-fatal). Implementation = table
      lookup from `keywords.go`. (AnnotationKind → permissive union of
      Kind contexts; UnboundBlock and simple annotations skip the
      check since their target context isn't known at the parser
      layer.)
- [x] **P1.8** 🟢 Unit tests per grammar production — the envelope is
      small enough (~6 productions) that each gets a direct test, not
      just fuzzing. (`productions_test.go` holds one focused test per
      §2.1 production: annotation-only, title-only, multi-paragraph
      description, property-lines, block-head, empty YAML, multiple
      YAML blocks, and a full-envelope composition test.)
- [x] **P1.9** 🟢 Characterize the legacy regex parser's **"implied stop"
      behavior** — document the markers after which the old parser
      stops decoding (typically blank-line-after-description, first
      unrecognized line, or kind-specific terminators). Output is a
      short appendix to this document (§A "Legacy stop points") that
      P5 bridge taggers consult. Research task, not code.
      Landed as `.claude/plans/legacy-stop-points.md` — 7 stop points
      catalogued (S1–S7); P5 obligations reduced to three concrete
      checks (AnnotationKind boundary, ignore short-circuit,
      block-body collection boundary).

> NOTES(fred):
>
> At this point, we need to understand better a behavior implied by the regexp-parser:
> it stops at some marker and doesn't parse the remainder.
>
> Our parser will obviously decode all known tokens. So this "implied stop" logic will have to be
> explicited in the analyzer (our Tagger bridge)
>
> **Settled: grammar parser is greedy by design (decodes all recognized
> tokens). Legacy stop semantics become a bridge-tagger responsibility
> during P5 — see the cross-cutting task template in P5 and the P1.9
> research deliverable.**

- [x] **P1.10** 🟢 Catch-up on flagged sub-issues — do NOT let them
      accumulate into a P4 parity-harness backlog. This task exists
      to enforce the rule *no "we'll check in P4" deferrals*: if a
      flag is raised during P1, it lands in P1.10 before the phase
      closes. Currently known items:

      - **Verbatim YAML body contract.** P1.4 captures YAML bodies
        via `reconstructLine(token)` in parser.go after the lexer
        has already classified interior tokens. Indentation and
        exact punctuation are lost — the body is unfit for re-parsing
        by `internal/parsers/yaml/`. Upgrade: add fence-state tracking
        in the lexer (or a raw-line-preserving path in the
        preprocessor) so bytes between `---` and `---` survive
        verbatim. `reconstructLine` becomes unnecessary and is
        removed. Update `TestParseYAMLFenceBalanced` to assert the
        interior line is present *with its leading indentation*.
      - **Preprocessor `-` stripping divergence from v1.** P1.4
        removed `-` from `trimContentPrefix`'s strip set so `---`
        fences survive. This diverges from v1 (which silently ate
        bullet-list dashes). Decide: restore the v1 strip but
        special-case the fence detection, or confirm v2 keeps the
        dash in description text as the intended contract. Either
        way, add a test that locks in the decision.
      - **Godoc-identifier prefix for `swagger:route`** — if P1.6
        does not cover the "`DoFoo swagger:route GET /path tags opid`"
        form, handle it here.
      - **Any further flags raised during P1.5–P1.9** land here
        before P1's exit criteria are declared green.

### Exit criteria

- Parses every annotation kind from `fixtures/enhancements/` and
  `fixtures/goparsing/petstore/` into the right typed `Block`.
- Diagnostics accumulate; parser never panics on any fixture (including
  `fixtures/goparsing/invalid/`).
- All v1 keywords are recognized with correct value-typing.
- Unit tests green; coverage on `internal/parsers/grammar/` ≥ 90 %.
- **P1.10 backlog is empty** — every flag raised during P1 has been
  resolved (fixed or explicitly declared out of scope with rationale),
  not deferred.

---

## Phase 2 — Sub-language isolation

**Goal:** recognize sub-language boundaries (YAML fences, inline
extensions) and hand the bodies off as raw-with-position. The main
grammar parser never parses sub-languages; implementations live in
sibling sub-parser subpackages under `internal/parsers/`, first of
which is `internal/parsers/yaml/`.

### Tasks

- [x] **P2.1** 🟢 YAML fence (`---`) detection in the lexer: emit
      `YAML_FENCE` open/close tokens; capture the body as
      `RawYAML{Text, Pos}`. Fence appears in `swagger:operation` and
      `swagger:meta`. (Landed across P1.4 `4ac06fd` + P1.10 `71bb504`
      — lexer tracks fence state, emits `TokenRawLine` verbatim
      inside fences; parser's `collectYAMLBody` preserves
      indentation. Seven tests exercise this path.)
- [x] **P2.2** 🟢 Inline extensions block (`extensions:` followed by
      indented `x-*` lines) — collect as `[]Extension{Name, RawValue,
      Pos}`. Value type left raw; analyzer decides whether to parse
      YAML if the value looks structured. Block.Extensions() iterator
      exposes them; Property.Body still carries the raw lines for
      analyzers that want verbatim content.
- [x] **P2.3** 🟢 Multi-line keyword blocks (`consumes:`, `produces:`,
      `security:`, `responses:`): capture as `Property.Body []string`.
      Body lines kept raw; analyzer tokenizes (MIME types, etc.)
      per-kind. Stops at next structured token (legacy S6 semantics);
      trailing blanks trimmed, interior blanks preserved.
- [x] **P2.4** 🟢 Extension-name well-formedness diagnostic
      (`^[Xx]-`) per architecture §3.6 — non-fatal
      `parse.invalid-extension-name`. Invalid names still end up in
      `Block.Extensions()` so analyzers can decide policy; the
      diagnostic is warning-level.
- [x] **P2.5** 🟢 Create `internal/parsers/yaml/` sub-parser subpackage —
      a thin wrapper around `go.yaml.in/yaml/v3` exposing e.g.
      `Parse(raw grammar.RawYAML) (YAMLNode, []grammar.Diagnostic)`.
      Imported only by the analyzer/bridging taggers; *never* by
      `internal/parsers/grammar/`. Unit-tested against representative
      YAML bodies from the fixtures.
      **This establishes the sub-parser pattern.** Future
      sub-languages (enum/example per W2/W3, richer extensions, private
      bodies per W4) get sibling subpackages following this seam.

### Exit criteria

- `go.yaml.in/yaml/v3` import appears only in `internal/parsers/yaml/`;
  `internal/parsers/grammar/` remains stdlib-only.
- Every `swagger:operation` / `swagger:meta` fixture carries its
  YAML body as a `RawYAML` field with byte-accurate positions.
- Every `extensions:` block round-trips `Name`/`RawValue` to the
  Block without loss.
- `internal/parsers/yaml/` compiles and has a small but representative
  unit-test set.

---

## Phase 3 — Parser API

**Goal:** expose the parser as an interface so builders (and tests)
consume it without caring about internals.

### Tasks

- [x] **P3.1** 🟢 `Parser` interface:
      ```go
      type Parser interface {
          Parse(*ast.CommentGroup) Block          // primary path
          ParseText(text string, pos token.Position) Block  // LSP/tests
          ParseAs(kind Kind, text string, pos token.Position) Block
              // LSP kind-hint path (§4.6)
      }
      ```
      Concrete implementation in the same package; builders take the
      interface.
- [x] **P3.2** 🟢 Package-level `NewParser(opts ...Option)` constructor.
      Options: logger, diagnostic sink override. Default is
      zero-configuration. Ships with `WithDiagnosticSink(cb)` as the
      first concrete option — cb receives every diagnostic as it's
      emitted (LSP streaming seam). Logger option deferred until a
      real need surfaces.
- [x] **P3.3** 🟢 Helper accessors on Block (V2 layer per §4.2):
      `GetFloat(name)`, `GetInt(name)`, `GetBool(name)`,
      `GetString(name)`, `GetList(name)`, `Has(name)`. Thin wrappers
      over the typed property list. Returns (zero, false) if
      mismatched. Case-insensitive lookup, alias-aware. GetList
      unifies comma-list values and block-head bodies; returns a
      defensive copy. GetString canonicalizes StringEnum values.

### Exit criteria

- Interface documented; two use sites in tests (mock Parser returning
  crafted Blocks; real Parser used end-to-end).
- Godoc example for each API entry.

---

## Phase 4 — Test infrastructure

**Goal:** the machinery that makes P5 safe. Parity harness is the
critical element. Fuzz + property-based testing are **deferred to
P7 hardening** — not required for migration safety.

### Tasks

- [x] **P4.1** 🟢 Harness scaffolding under
      `internal/parsers/grammar/grammar_test/`: iterates fixture sources,
      parses via v2 `grammar.Parser`, normalizes to a common view,
      diffs against committed JSON snapshots. v1 side (extraction
      from the TypeIndex path) lands alongside P5 bridge-tagger
      migration — pre-P5 the harness is the v2 regression suite.
- [x] **P4.2** 🟢 `NormalizedCommentView` + `ViewFromBlock` in
      `grammartest`: annotation kind/args, title, description,
      properties (with typed values), YAML bodies, extensions,
      sorted diagnostics by code. JSON-serialisable,
      diff-deterministic.
- [ ] **P4.3** Golden fixtures integration: existing
      `scantest.CompareOrDumpJSON` harness unchanged; new grammar
      parser is wired behind a build tag or
      `Options.UseGrammarParser` feature flag so both paths coexist
      during migration. **Deferred to the first P5 bridge-tagger
      commit** — the flag needs a builder consumer to be useful, and
      that consumer is P5.1.

### Exit criteria

- Parity harness runs in CI and passes against every fixture the
  current code passes against.
- Normalized-view diffs are actionable (human-readable) when parity
  breaks during P5.

---

## Workshop gate before P5.1 — W2 (enum) [+ W3 deferred]

**W2 closed 2026-04-21** — see `.claude/plans/workshops/w2-enum.md`.
Decisions feeding into P5.1:

- Keep all three v1 surface forms (comma-list, JSON array, linked
  `swagger:enum TypeName`) — v2 parser migration is non-breaking.
- New sub-parser subpackage `internal/parsers/enum/` created during
  P5.1, mirroring `internal/parsers/yaml/`. Main grammar parser
  stays oblivious to enum value shape.
- `var` / composite / computed enum sources stay out of scope —
  deferred as forthcoming feature.
- YAML-block enum values (never implemented in v1) deferred too.
- Non-scalar enum emission is permitted — the JSON path already
  supports it; fixture audit will follow in a later release.

**W3 (example / examples) deferred 2026-04-21** — see
`.claude/plans/forthcoming-features.md` §2. Rationale: same shape
as W2 (array of any-type values), not supported in v1 beyond a
single scalar raw-value-per-field, so parity = inherit current
behavior. When W3 is later held (likely alongside OAI 3.x
onboarding), it reuses W2's sub-parser pattern directly.

**Gate status: CLEAR.** P5.1 may start.

Why this is the right moment:

- The grammar parser already exposes the three `enum:` surface forms
  as distinct AST shapes (architecture §3.2.1); parser implementation
  is unblocked, but the analyzer's reconciliation target is not yet
  decided.
- The schema builder (P5.1) is the first consumer that has to *do*
  something with those shapes. Without W2/W3 outcomes, P5.1 would
  stall or ship a placeholder we'd have to revisit.
- The parity harness (P4) is in place, so workshop proposals can be
  evaluated against real fixtures.

Output: a short decision doc under `.claude/plans/workshops/`
capturing the unified target shape for enum/example values, the
backward-compat surface (if any), and type-conversion rules per Go
field type. Feeds directly into the P5.1 bridge-tagger implementation.

---

## Phase 5 — Builder migration

**Goal:** flip each tagger tree to consume grammar Blocks. Builders'
external behavior unchanged (parity harness proves this).

**Migration targets — five tagger-carrying builders.** The `spec`
builder is the orchestrator, not directly migrated; its
`swagger:meta` consumption is a small plumbing change bundled into
P6. The `items` builder has no annotations of its own (array-item
validations only) and rides along with `schema` — it shares the
validation surface.

Order is simplest-first (smaller surface, no sub-languages → bigger
surface → YAML-carrying):

1. **P5.1 `schema` builder (+ `items`)** ✅ — landed as P5.1a
   (`25fb351`, items) + P5.1b (`1701ae7`, schema). The second
   step's redesign (bite the full bullet — grammar owns
   description classification, one commit per builder) set the
   template for P5.2–P5.5. Enum handling stayed on v1's
   `ParseEnum` for parity; the switch to
   `internal/parsers/enum.Parse` is a post-migration follow-up
   (forthcoming-features.md §1.2a). The W2 quirks (comma-list
   whitespace trim, no-consts warning, stale `x-go-enum-desc`
   cleanup) landed earlier as Q1/Q2/Q3 so P5.1 could run
   parity-clean.
2. **P5.2 `parameters` builder** ✅ — commit `f5aebc1`. Routes
   `required:` to `param.Required`, adds `collectionFormat`,
   leaves `in:` as upstream-resolved (grammar's lexer already
   classifies it as a KEYWORD_VALUE so it never reaches prose).
3. **P5.3 `responses` builder** ✅ — commit `198511b`. Two call
   sites (decl + headers); headers reuse the schema-family
   dispatcher minus required/readOnly/discriminator.
4. **P5.4 `operations` builder** ✅ — commit `dab83ff`. YAML body
   via `internal/parsers/yaml/` + `parsers.RemoveIndent` for tab
   normalization.
5. **P5.5 `routes` builder** ✅ — commit `1d68cd5`. Rich bodies
   (Consumes/Produces/Schemes/Security/Parameters/Responses/
   Extensions) delegate to existing v1 body-parsers; grammar
   added RawBlock sub-context keyword absorption + indentation-
   preserving extensions to make the dispatch land correctly.

### Task template (applied to each P5.x) — outcome

What actually happened per builder (the plan's "thin iterator
wrappers" simplification broke on P5.1b when the v1 taggers'
dual role emerged — see `p5.1b-schema-walkthrough.md`):

- [x] **Grammar bridge file** per builder (`bridge.go`) with a
      dispatcher that iterates `block.Properties()` /
      `block.Extensions()` / `block.YAMLBlocks()` and routes to the
      per-target write methods. Each bridge includes a small items-
      level walk (`collectItemsLevels` / `collectParamItemsLevels` /
      `collectHeaderItemsLevels`) mirroring the legacy
      `parseArrayTypes` recursion.
- [x] **Legacy regex + SectionedParser taggers** deleted as dead
      code at P6 / P6.1 cutover (couldn't be removed piecemeal
      because v1 taggers also served to claim lines away from the
      description accumulator).
- [x] **Implied-stop semantics** ported into grammar's
      `collectBlockBody` rather than per-bridge. Notable: raw-block
      bodies absorb sub-context keyword-shaped lines (`default:`,
      `in:`, `required:`, `max:`) as body text rather than
      terminating; route/operation/meta-structural keywords
      (`schemes:`, `deprecated:`, other RawBlock heads) still
      terminate. Extension bodies preserve source indentation via a
      Token.Raw field.
- [x] Parity harness (TestParity) stayed green across 24 fixtures
      throughout the migration; deleted at P6 alongside the flag.
- [x] Dual-mode CI harness (`CODESCAN_USE_GRAMMAR=1`) surfaced
      additional gaps on the non-parity test suite — fixed before
      P6 cutover (commit `d55a960`).

### Exit criteria (overall P5) ✅ all met

- [x] All six migratable builders switched (items, schema,
      parameters, responses, operations, routes). Meta followed at
      P6.1.
- [x] `internal/parsers/` legacy files removed at P6/P6.1; what
      remains is the grammar parser itself, the yaml sub-parser,
      and the residual helpers the bridges + scanner
      classification still call.
- [x] `Options.UseGrammarParser` removed at P6.

---

## Phase 6 — Cutover

**Goal:** remove the old parser, rename the package, ship.

### Tasks

- [x] **P6.4** ✅ commit `641cb4a` — Remove
      `Options.UseGrammarParser` feature flag and
      `internal/integration/parity_test.go`, collapse all
      `if ctx.UseGrammarParser() { … } else { … }` branches in the
      six builders to the grammar arm, delete the validation
      tagger types / match-only taggers / `SetDeprecatedOp` /
      `SetEnum` that served only the legacy pipeline. Dead-regex
      prune: `rxDiscriminator`, `rxReadOnly`, `rxDeprecated`. Also
      deletes the tactical `CODESCAN_USE_GRAMMAR` env-var harness
      added in commit `d55a960` — the harness was explicitly
      scoped to this cutover.
- [x] **P6.1** ✅ commit `761c439` — Migrate `swagger:meta`
      through the grammar bridge (post-cutover because SectionedParser
      was kept alive for meta during P6.4). `internal/builders/spec/meta_bridge.go`
      replaces `parsers.NewMetaParser`. Dead SectionedParser +
      tag_parsers.go + meta-specific regexes + `NewSetSchemes` +
      `newSetSecurity` + `multilineDropEmptyParser` all deleted in
      the same commit. `parsers.MetaSection` flattened to
      `*ast.CommentGroup` in TypeIndex.
- [ ] **P6.2** Delete remaining legacy `internal/parsers/*.go`
      survivors. After P6.1 the survivors are the ones actually
      consumed by the bridges (NewConsumesDropEmptyParser,
      NewProducesDropEmptyParser, NewSetSecurityScheme, NewSetParams,
      NewSetResponses, NewSetExtensions, NewYAMLParser, RemoveIndent)
      plus scanner helpers (matchers.go, parsed_path_content.go,
      lines.go, parsers_helpers.go). A further consolidation pass
      could fold them into leaner package shapes; not blocking.
- [ ] **P6.3** Promote `internal/parsers/grammar/*.go` up to
      `internal/parsers/`. Deferred: meaningful only alongside
      P6.2 cleanup + renaming of sibling subpackages. Left as a
      cosmetic follow-up.
- [x] **P6.5** Update `CLAUDE.md` package-layout table —
      **deferred** to a docs pass after P6.2/P6.3 since the parsers
      package shape is still in flux.
- [ ] Executed by Fred: **P6.6** Write release notes / CHANGELOG entry. Single public
      release with the whole lot.
- [ ] Executed by Fred: **P6.7** Squash-or-merge the feature branch into master with a
      single merge commit (per architecture §4.5 release cadence).

### Exit criteria

- [x] `Options.UseGrammarParser` flag and `TestParity` removed.
- [x] All tests pass on master.
- [x] `golangci-lint run --new-from-rev master` clean on the
      migration commits.
- [ ] No `regexp` import in `internal/parsers/` beyond what the
      surviving scanner matchers and route/operation annotation
      regexes require. (Partial — the grammar-side code itself has
      none; matchers + parsed_path_content still use regexp.)

---

## Phase 7 — Hardening (post-cutover)

**Goal:** stress-test the grammar parser and builders beyond the
fixture corpus. Non-blocking for v2.0 ship; schedules in the v2.0.x
window once P0–P6 have landed.

### Tasks

- [ ] **P7.1** Fuzz target on `preprocess.go` — seed corpus from every
      `*_test.go` fixture comment group. Stdlib fuzzing since Go 1.18.
- [ ] **P7.2** Fuzz target on `lexer.go` — seed from P7.1's corpus.
- [ ] **P7.3** Fuzz target on `parser.go` — seed likewise. Every crash
      becomes a regression fixture.
- [ ] **P7.4** Property-based builder tests — generator for synthetic
      `Block` values (valid shapes → near-miss shapes), driven through
      the `Parser` interface mock (P3.1). One generator per builder;
      assert invariants on the produced `spec.*`. Pattern follows
      `go-openapi/testify`.
- [ ] **P7.5** Commit ≥1h of seed runs per fuzz target to the corpus.
- [ ] **P7.6** Triage + fix any P7 findings; regressions turn into
      fixtures in the main corpus.
- [ ] **P7.7** **Extend annotation surface-form documentation.**
      Origin: Q4 design debate (2026-04-22) exposed that
      multi-line block-body syntax (consumes/produces/security/…) is
      under-documented — users and maintainers had different
      implicit models of what the body contract was. Landing work
      for Q4 makes the contract strict (YAML-list via
      `internal/parsers/yaml/`), so the doc needs to:
      - State explicitly that block bodies are parsed as YAML lists
        (strict); scalar bodies emit `parse.invalid-block-body`.
      - Enumerate every keyword whose body follows this contract
        (`consumes`, `produces`, `security`, `securityDefinitions`,
        `responses`, `parameters`, `extensions`, `infoExtensions`,
        `tos`, `externalDocs`, `schemes`).
      - Cover the YAML-list vs bare-form distinction with migration
        guidance for any `swagger:` comments that used bare form.
      - Cross-reference the per-keyword docs generated by
        `internal/parsers/grammar/gen/` (docs/annotation-keywords.md)
        so the surface-forms doc and the keyword table agree.

      Also revisit the enum surface-forms doc in the same pass (W2
      §1 identified three forms; user-facing docs never explained
      the JSON-array and linked-const variants clearly).

### Exit criteria

- Each fuzz target runs ≥1h clean (no crashes, no diagnostics surfacing
  unexpected classes).
- Property-based tests generate ≥1000 valid `Block` shapes per builder
  with zero regressions against the golden corpus.
- Any P7 findings triaged: either fixed, or logged as deferred issues
  with rationale.

---

## Cross-cutting checklist

Applies throughout:

- DCO sign-off on every commit (`git commit -s`).
- Fred is commit author; agents listed as `Co-Authored-By:` per
  `.claude/rules/contributions.md`.
- No new feature flags unless ephemeral and removed at P6.
- Never widen production API to satisfy tests — use `export_test.go`
  or integration tests (memory note `feedback_test_api_surface.md`).
- No silent fallthroughs in `go/types` switches (memory note
  `feedback_go_types_defensive_guards.md`).
- **No "we'll check in P4 parity" deferrals.** Every phase ends with
  a *catch-up* task (P1.10, P2.x, P3.x, …) that resolves flags raised
  during the phase. If a backlog can't be resolved, it's explicitly
  punted with rationale — never implicitly. The parity harness in P4
  is a safety net, not a queue for accumulated known issues.

---

## Risks and mitigations

| Risk | Mitigation |
|------|------------|
| Parity harness is slow on large fixture set | Run per-builder subset; full run nightly in CI |
| Old parser has undocumented behavior captured by fixtures | Parity harness catches it; new parser replicates or declares a pre-migration bug (candidate for deferred-quirks.md entry) |
| Builder migration order picks a builder with hidden dependencies on another | Review `internal/ifaces` and `resolvers` upfront; schema+items are independent, operations depends on parameters/responses — honored in order above |
| Sub-language edge (nested YAML inside an extensions block) | P2.2 captures raw; analyzer-side test coverage in P5.4/P5.5 |
| Legacy "implied stop" behavior diverges silently between old and new parsers | P1.9 research output; bridge-tagger obligation in P5 task template; parity harness catches residual cases |
| Long-lived feature branch drifts from master | Weekly rebase on master; keep migrations small and independent |

---

## Open items (to resolve during P0/P1)

These are small enough to settle in-flight; flagged so they're not
forgotten:

- **Kind enum shape** — is `Kind` an `int` with constants
  (`KindModel`, `KindRoute`, …) or a string alias? Lean toward `int`
  for switch-exhaustiveness helpers.
- **`Block` in `context.Context`?** Probably no; pass explicitly.
- **Diagnostic sink** — per-parse accumulator on `Block`, plus an
  optional global sink (for LSP streaming). Default is local-only.
- **`keywords.go` authoring ergonomics** — struct literal list or
  a builder helper (`kw("maximum", num(), ann(Param, Header))`)?
  Decide during P0.2 based on readability of the first dozen entries.
