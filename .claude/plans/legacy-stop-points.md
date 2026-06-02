# Legacy "implied stop" points — v1 parser catalog

Date: 2026-04-21
Status: reference document
Produced by: P1.9 (tasks plan).
Consumed by: P5 bridge-tagger implementers.

---

## Why this document exists

The v2 grammar parser at `internal/parsers/grammar/` is **greedy by
design**: it decodes every recognised line in a comment group into a
typed `Block` with a flat `Properties()` iterator, ordered as they
appeared in source. It never stops early.

The v1 regex-based parser at `internal/parsers/` was the opposite:
it stopped decoding at any of several **implicit markers** (blank
lines, annotation boundaries, explicit `swagger:ignore`, mode
switches, …). A great deal of downstream behavior was quietly
encoded in these stops — lines after a stop were simply never seen
by the tagger machinery.

**P5 migration rule:** a bridge-tagger consuming
`block.Properties()` must decide for itself when it has "enough" and
stop iterating. The parser will not stop for it. Forgetting this
produces false positives (properties from outside the tagger's
logical scope bleed in) and parity-harness failures.

This document catalogs the seven implicit stop points identified in
the v1 parser so every bridge-tagger knows which one(s) apply to it.

---

## Catalog

All code references point into `internal/parsers/` (the legacy
package, still present during migration). Every entry names the
concrete trigger, the affected builder/tagger path, and what the
bridge-tagger must do to preserve parity.

### S1 — Unrecognized `swagger:<name>` annotation terminates the block

- **Trigger:** any comment line containing `swagger:<identifier>`
  where `<identifier>` is not one of the registered annotations the
  current scanner path handles.
- **Code reference:** `sectioned_parser.go:217-218`
  (`if st.annotation == nil || !st.annotation.Matches(line) { return true }`).
  The regex that detects the family is
  `rxSwaggerAnnotation` (`(?:^|[\s/])swagger:([\p{L}\p{N}\p{Pd}\p{Pc}]+)`).
- **Scope:** every `SectionedParser` — all builders that wrap one.
- **Effect in v1:** the block's decoding halts at that line; all
  subsequent content is discarded.
- **P5 bridge behavior:** when iterating `block.Properties()`, stop
  consuming at the point the property position advances past an
  inner annotation boundary. Since v2 `Block` objects are per-comment-group
  (one annotation per block), this translates to: **never cross an
  `AnnotationKind` boundary** — if `block.AnnotationKind()` is not
  one your tagger handles, bail out without iterating.

### S2 — Explicit `swagger:ignore` terminates the block

- **Trigger:** `swagger:ignore [Name]` appearing anywhere in the
  block.
- **Code reference:** `sectioned_parser.go:213-216`; regex
  `rxIgnoreOverride`.
- **Scope:** every builder.
- **Effect in v1:** `st.ignored = true`; block is dropped entirely
  from the spec.
- **P5 bridge behavior:** check
  `block.AnnotationKind() == AnnIgnore` (or the presence of an
  `swagger:ignore` annotation token) **before** consuming any
  property. If ignored, produce no spec output. This is a global
  short-circuit, not a mid-iteration stop.

### S3 — YAML spec delimiter `---` switches mode

- **Trigger:** a comment line whose trimmed content is exactly `---`.
- **Code reference:** `yaml_spec_parser.go:59` (`rxBeginYAMLSpec`).
- **Scope:** `YAMLSpecScanner` — used by `swagger:operation` and
  `swagger:meta` bodies.
- **Effect in v1:** the header (title/description/top-level keyword)
  collection stops; subsequent lines accumulate as raw YAML until a
  closing `---` or EOF.
- **P5 bridge behavior:** **no action required.** The v2 grammar
  parser already isolates YAML bodies into `Block.YAMLBlocks()`,
  and `Block.Properties()` contains only non-YAML properties. The
  body/property separation is automatic.

### S4 — Annotation-family token terminates a YAML spec

- **Trigger:** `rxSwaggerAnnotation` match anywhere in a line while
  the parser is inside a YAML spec body.
- **Code reference:** `yaml_spec_parser.go:54-56`
  (`HasAnnotation(line)` → `break COMMENTS`).
- **Scope:** `YAMLSpecScanner` only.
- **Effect in v1:** closes the YAML body (treating EOF as closing
  delimiter); everything after the swagger-annotation line is
  discarded.
- **P5 bridge behavior:** the v2 parser keeps YAML bodies strictly
  between matched `---` fences (`collectYAMLBody` in
  `grammar/parser.go`). **No action** unless P2.1 later accepts
  unclosed fences the way v1 did — in which case the bridge must
  look for `swagger:` inside `RawYAML.Text` and truncate. Track the
  discrepancy through the P4 parity harness.

### S5 — First blank line splits title from description

- **Trigger:** first empty line after a non-empty header line.
- **Code reference:** `parsers_helpers.go:16-20`.
- **Scope:** every `SectionedParser` with a title/description
  callback.
- **Effect in v1:** lines before the blank become the block title;
  lines after become the description (until S1/S2 or another stop).
- **P5 bridge behavior:** **no action.** `Block.Title()` and
  `Block.Description()` are already populated by the grammar
  parser's `parseTitleDesc` using the same first-blank rule; see
  `parser.go` production tests in `productions_test.go`.

### S6 — Multi-line tagger replaced by another tagger

- **Trigger:** while a multi-line tagger (e.g., `consumes:`,
  `produces:`, `security:`, `responses:`, `parameters:`) is actively
  collecting body lines, a different tagger's opener matches.
- **Code reference:** `sectioned_parser.go:230-237`.
- **Scope:** multi-line taggers.
- **Effect in v1:** the previous tagger's body-line collection
  stops at the new opener; the new tagger takes over from the next
  line.
- **P5 bridge behavior:** when the v2 parser lands P2.3 (multi-line
  block-body capture), consecutive block-head properties will
  naturally split their bodies on the next keyword/block-head. For
  now, block heads are emitted as value-less properties
  (`KeywordBlockHead`); the bridge-tagger iterates
  `block.Properties()` and transitions from body-collection mode to
  property-mode whenever a non-body keyword is seen. Until P2.3
  lands, **body lines are dropped** — this will surface as a parity
  gap the harness reports.

### S7 — Single-line tagger auto-resets after one line

- **Trigger:** any single-line tagger (validation keywords like
  `maximum:`, `pattern:`, meta fields like `version:`, `host:`,
  `basePath:`).
- **Code reference:** `sectioned_parser.go:266-268`.
- **Scope:** every single-line keyword tagger.
- **Effect in v1:** after the line is consumed, the active tagger
  is cleared; the next line must independently match a tagger or it
  becomes header content.
- **P5 bridge behavior:** **no action.** `Block.Properties()` is
  already a flat list — each property is self-contained. The
  "auto-reset" is intrinsic to the data structure, not a behavior
  the bridge must replicate.

---

## Summary for P5 bridge-taggers

| Stop | v1 → v2 status | Bridge action required? |
|------|-----------------|-------------------------|
| S1 — unrecognized annotation | AnnotationKind boundary = block boundary | **Check `AnnotationKind` before iterating.** |
| S2 — `swagger:ignore` | `AnnIgnore` kind on Block | **Short-circuit: emit nothing for `AnnIgnore`.** |
| S3 — `---` YAML delimiter | Isolated automatically | No |
| S4 — annotation inside YAML | Preserved via `---` pairing | No (unless unclosed fences come back) |
| S5 — first blank = title/desc split | `Title()` / `Description()` populated | No |
| S6 — multi-line tagger switch | **Not yet implemented in v2** | **Handle body collection state explicitly in P5 until P2.3.** |
| S7 — single-line tagger reset | Intrinsic to flat property list | No |

**Net P5 obligations** — each bridge-tagger author should, before
touching code, confirm they handle:

1. **S1**: `if block.AnnotationKind() != AnnMyKind { return }` (or
   equivalent early exit).
2. **S2**: short-circuit when `AnnotationKind() == AnnIgnore`.
3. **S6**: if the tagger owns a multi-line block (consumes, security,
   responses, parameters), its iterator body-collection loop stops
   at the next non-body property token. Without this, v2 will
   continue decoding past what v1 would have stopped at.

S3, S4, S5, S7 are handled by the parser and require no bridge work.

---

## Change history

| Date       | Change |
|------------|--------|
| 2026-04-21 | Initial catalog (P1.9). |
