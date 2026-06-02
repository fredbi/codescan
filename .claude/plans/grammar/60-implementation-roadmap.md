# Implementation roadmap — from spec to code

The grammar + lexer specification is complete. This doc is the
handoff: what's settled, what blocks implementation, and how the
work decomposes into phases.

## Spec status

| File                          | State                                             |
|-------------------------------|----------------------------------------------------|
| `00-overview.md`              | Final                                              |
| `10-shared.md`                | Final                                              |
| `20-schema-grammar.md`        | Final                                              |
| `21-operation-grammar.md`     | Final                                              |
| `22-meta-grammar.md`          | Final                                              |
| `23-classifier-grammar.md`    | Final                                              |
| `30-delegated.md`             | Final                                              |
| `40-lexer.md`                 | Final modulo §7 (termination rules per body kind — see below) |
| `50-full.md`                  | Final — synthesis read against lexer-emitted terminals  |
| `open-questions.md`           | Q1, Q2, Q3, Q5–Q9, Q11–Q18 SETTLED; Q10 deferred to v2 design; Q19, Q20 marked OPEN/v2 realignment |

## Pre-flight blockers

### A3 — lexer body termination rules (`40-lexer.md` §7)

The lexer's body accumulator needs a per-body-kind rule for "this
line ends the body". §7 of `40-lexer.md` proposes round-1 defaults
per body kind. Each rule needs **fixture validation** before
implementation begins:

- Pick 5–10 representative bodies per kind from
  `fixtures/goparsing/...`.
- Walk each by hand against the proposed rule.
- Confirm v1 output and the proposed rule produce the same line
  groupings.

This is a workshop session, not a coding session — the output is
a settled rule set per body kind, written into `40-lexer.md` §7
as the final contract.

Until A3 lands, the body accumulator can't be coded.

### A9 — disambiguation module location (`40-lexer.md` §11)

Stance taken: single module under
`internal/parsers/grammar/disambiguate.go`, called by both lexer
and analyzer. No further blocker; mentioned for completeness.

## Implementation phases

### Phase 1 — Lexer rewrite

**Scope:** Rewrite `internal/parsers/grammar/lexer.go` to emit the
new terminal vocabulary (`50-full.md` §1).

Sub-tasks:
- Token-kind enum extension: add `TokenIdentName`, `TokenJsonValue`,
  `TokenRawValue`, `TokenTypeRef`, `TokenHttpMethod`, `TokenUrlPath`,
  `TokenNumberValue`, `TokenIntValue`, `TokenBoolValue`,
  `TokenStringValue`, `TokenCommaListValue`, `TokenEnumOptionValue`,
  `TokenTitleLine`, `TokenDescLine`, `TokenRawBlockBody`,
  `TokenRawValueBody`, `TokenOpaqueYamlBody`. Preserve positions for
  every token.
- Line classifier: per-keyword/per-annotation typing rules (when the
  lexer sees `maximum:`, expect a `NUMBER_VALUE` for the inline
  value; etc.). Soft typing — emit the most specific lexical shape
  recognised; the grammar enforces pairing.
- Body accumulator: state machine implementing §4 of `40-lexer.md`,
  with the per-body-kind termination rules from A3.
- Prose classifier: emit `TITLE` / `DESC` per the four heuristics
  in §8. Folded into the body-accumulator emission path or as a
  third stage — implementer's call.
- Disambiguation: implement `disambiguate.go` for `JsonValue` vs
  `RawValue`, `EnumWithName` vs `EnumValuesOnly`, `EnumPlainList`
  vs `EnumBracketedList`, `IDENT_TAG` vs `IDENT_OP_ID` (last-ident
  rule for `OperationArgs`).
- Quirk absorption confirmed at this phase: CR/CRLF, trailing-dot
  elision, godoc prefix, decorative fences inside extension
  bodies, `items.` runs, first-character case insensitivity on
  keywords.

**Test surface:** unit tests on the lexer in isolation. Each
fixture from §7 (post-A3) becomes a lexer-test fixture asserting
expected token sequence.

### Phase 2 — Grammar parser refactor

**Scope:** Rewire `internal/parsers/grammar/parser.go` to consume
the new terminals and produce the same `Block` AST shape (or an
extended one) that today's bridges already consume.

Key consideration: the existing AST surface
(`block.Properties()`, `block.YAMLBlocks()`, `block.ProseLines()`,
`block.AnnotationKind()`) is what the bridges call. Either the new
parser produces the same AST (low-risk), or the AST gets refreshed
and bridges adapt.

Sub-tasks:
- Recursive-descent productions per `50-full.md` §2-§7.
- Diagnostics: the lexer is total (every line classifies somehow);
  the parser emits structured diagnostics on production mismatches.
- Backwards-compat wrapper if needed so Phase 3 doesn't gate Phase 2.

**Test surface:** parser unit tests + the existing
`grammar/parser_test.go` corpus. The grammar should accept every
v1-corpus comment block and produce the same `Block` for it.

### Phase 3 — Builder bridge adaptation

**Scope:** The bridges in `internal/builders/{schema,operations,
routes,parameters,responses}/bridge.go` consume the AST from Phase
2. Most bridge code stays — but several helper-package call sites
become obsolete because the lexer now does that work.

Helpers to retire (as the bridges no longer call them):
- `helpers.CollectScannerTitleDescription` — replaced by
  `block.Title()` / `block.Description()` (per the new lexer's
  `TITLE` / `DESC` emission).
- `helpers.CleanupScannerLines` — folded into the lexer.
- Possibly `helpers.JoinDropLast` — depends on whether the new AST
  exposes joined strings or line slices.
- Possibly `helpers.RemoveIndent` — depends on whether the lexer
  normalises YAML body indentation.

Per-keyword dispatch in `dispatch{Numeric,Integer,String,Flag,...}`
keeps the same shape but reads the typed-value attributes from the
new terminals.

### Phase 4 — Builder merge (Q20)

**Scope:** Merge `internal/builders/routes/` and
`internal/builders/operations/` into one builder that handles both
annotations uniformly.

Per the Q20 / earlier discussion: this is the heavy lift, ~2-3
days. Concentrated work, contained blast radius (both write to
the same `*oaispec.Operation` slot via `setPathOperation`).

**Settled before this phase:**
- Precedence rule when a YAML body and inline keyword raw-blocks
  both appear (proposed: YAML body first, inline keywords
  override).
- Builder package layout: merge into
  `internal/builders/operations/`; deprecate `routes/`.

After Phase 4 lands and stabilises, the grammar collapse —
`OperationAnnotation = ANN_ROUTE | ANN_OPERATION` — is a one-line
EBNF edit. No further code changes needed because the builder
already handles both uniformly.

### Phase 5 — Property-based hardening (was P7)

After Phases 1-4 land, return to the originally planned P7:
property-based and fuzz testing. The new grammar's smaller
permissive surface should make this much easier than against the
v1 corpus.

## Sequencing notes

- Phases 1 and 2 can overlap if the lexer publishes its token
  stream early.
- Phase 3 starts only after Phase 2 stabilises; it touches every
  bridge.
- Phase 4 is independent of Phase 3 timing — it can land months
  later as a separate effort.
- Phase 5 starts only after Phase 4 (or after Phase 3 if Phase 4
  is deferred).

## What's not in this roadmap

- v2 enhancements (Q10 value-format unification, `type:` /
  `format:` body keywords, schema-level `required:` / `discriminator:`,
  SimpleSchema vs full Schema split for parameters/response headers).
  All deferred per the family files' "Out of scope — v2 enhancements"
  sections.
- Workshops `w2-enum`, `w3-example`, `w4-private` — pre-existing
  workshop docs in `.claude/plans/workshops/` that may inform Phase
  3 details but are not gating.

## Handoff checklist

Before Phase 1 begins:
- [ ] A3 termination rules settled against fixtures (see §"Pre-flight blockers" above).
- [ ] Disambiguation module skeleton agreed (file location, function signatures).
- [ ] Token-kind enum drafted in code (a stub is enough — populate as Phase 1 proceeds).
- [ ] Builder merge / Phase 4 scheduling decision: now or later?
