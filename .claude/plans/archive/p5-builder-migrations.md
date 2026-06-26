# P5 builder migrations — execution plan

Date: 2026-04-21
Status: **✅ COMPLETE** — P5.1 → P5.5 + P6 cutover + P6.1 meta migration all
landed. Legacy SectionedParser / YAMLSpecScanner / regex taggers
removed. Grammar parser is the single path.
Companion: `grammar-parser-tasks.md`, `grammar-parser-architecture.md`,
`workshops/w2-enum.md`, `legacy-stop-points.md`,
`forthcoming-features.md`.

This doc is the migration playbook. The tasks plan said "parity-green
at every step"; this doc says *how*. One section per builder, all
sharing the same ten numbered steps.

## Completion log

| Phase | Commit | Summary |
|-------|--------|---------|
| P5.1a | `25fb351` | items bridge-tagger (step 6.items). |
| Q5 | `d165f50` | finish 09f6748 — annotation terminator must be line-start. |
| P5.1b | `1701ae7` | schema bridge — grammar owns schema-field parsing (redesign after shadow-path surfaced dual-role of v1 taggers). |
| P5.2 | `f5aebc1` | parameters bridge. |
| P5.3 | `198511b` | responses bridge. |
| P5.4 | `dab83ff` | operations bridge + YAML body through `RemoveIndent`. |
| P5.5 | `1d68cd5` | routes bridge + raw-block sub-context keyword absorption + indentation-preserving extensions body. |
| P5 harness | `d55a960` | dual-mode CI via `CODESCAN_USE_GRAMMAR` env var; fixes exposed across the full suite. |
| P6 | `641cb4a` | cutover — remove `UseGrammarParser` flag, legacy validation taggers, `TestParity`. ~2700 lines deleted. |
| P6.1 | `761c439` | meta builder migrated; `SectionedParser` + `TagParser` machinery + remaining meta regexes deleted. |

---

## 1. Dependency order (settled)

Based on the survey (2026-04-21):

```
items → schema → parameters
              → responses → operations → routes
                 meta  (orchestrator; bundled into P6, not P5)
```

Order:

| # | Builder | Why this slot |
|---|---------|---------------|
| P5.1a ✅ | `items` | Leaf — no builder depends on pre-migration items. Smallest validation surface (9 keywords, no type-level path). Perfect first migration. |
| P5.1b ✅ | `schema` | Dependency root for the rest. Two pipelines: field-level taggers + type-level annotations (on type decls). Biggest surface. |
| P5.2 ✅ | `parameters` | Consumes schema; adds `in:`, `required`, `collectionFormat`. First non-root migration. |
| P5.3 ✅ | `responses` | Consumes schema; parallel structure to parameters with headers. |
| P5.4 ✅ | `operations` | First user of YAML body (sub-parser wiring via `internal/parsers/yaml/`). |
| P5.5 ✅ | `routes` | Path + positional + rich bodies (Consumes/Produces/Schemes/Security/Parameters/Responses/Extensions). |
| P6.1 ✅ | `meta` | Last SectionedParser consumer migrated (originally deferred; landed post-cutover). |

`meta` originally was not in this list — the orchestrator's
`swagger:meta` consumption was deferred to P6.1 which landed
after P6 cutover.

Items + schema go together as P5.1 because they share the same
`ValidationBuilder` target and can ride one commit if the scope
stays under control. If during migration the commit feels too
large, split at `items` → `schema` boundary.

---

## 2. The ten numbered steps (per builder)

Pre-migration (shared, one-time before P5.1 starts):

1. ✅ **Survey** the builder(s) — done upfront, per section below.
2. ✅ **Plan doc section** — this file, filled in per builder.
3. ✅ **Parity-harness v1 adapter** — landed as spec-level
   `TestParity` (commit `44c9c81`, simpler than the originally-
   proposed view-level adapter). Deleted at P6 cutover.
4. ✅ **`Options.UseGrammarParser` feature flag** — landed as
   commit `b40c283`, removed at P6 cutover (`641cb4a`).
5. ✅ **`internal/parsers/enum/` sub-parser** — landed as commit
   `5675a66`. Bridges kept routing through v1's ParseEnum for
   parity; the direct `enum.Parse` swap is a forthcoming
   follow-up (see `forthcoming-features.md` §1.2a).

Per-builder migration (redesigned mid-migration — see P5.1b
walkthrough for the pivot):

In practice steps 6–10 collapsed into a single commit per builder
because the shadow-path design from the original plan proved
unworkable (v1 taggers served a dual role: writing AND claiming
lines from the description accumulator; removing them caused
description leaks). Each builder commit landed as:

6'. **Bridge + flag routing** — single commit flipping the flag on
    at each builder's call site(s) and delivering the full keyword
    dispatch. Enum stayed on v1 ParseEnum for parity (forthcoming
    §1.2a). Validation errors for malformed default/example
    propagate up.

Steps 7–10 from the original plan are subsumed into step 6' since
every builder landed in one commit.

Steps 1–5 were one-time shared infrastructure; step 6' repeated
once per builder (P5.1a through P5.5 + P6.1 for meta).

---

## 3. Commit cadence (refining the P5 task template)

Each step from 3 onward is ONE commit. The P5 task template
mandates:

- [ ] Parity harness green for fixtures under this builder's scope.
- [ ] Coverage ≥ pre-migration (sanity check — bridge-taggers are
      simpler; coverage shouldn't drop but the measurement catches
      accidental test deletion).
- [ ] Commit subject `feat(grammar): migrate <builder> <step>`.
- [ ] Commit body lists:
    - Which taggers migrated.
    - Any v1 behavior deliberately diverged (cite fixtures and
      rationale).
    - Which tests (if any) were updated.

If a step surfaces a flag (parser bug, fixture quirk,
scope creep), log it to `forthcoming-features.md` in the same
commit — "no check in P4 parity" deferrals per the cross-cutting
rule.

---

## 4. P5.1 — `items` + `schema` migration plan

### 4.1 Items builder — scope

Taggers (9): `maximum`, `minimum`, `multipleOf`, `minLength`,
`maxLength`, `pattern`, `maxItems`, `minItems`, `unique` (also
`enum`, `default`, `example` at `items.` depth via
`itemsTaggers()`). Target interface: `items.Validations`
(implements `ValidationBuilder`).

**Not** in items: `required`, `readOnly`, `discriminator`, `in:`
(those are field-level / parameter-level), `collectionFormat`
(that's `OperationValidationBuilder`, a parameters/headers concern).

### 4.2 Items — step sequence

- **Step 6.items** — bridge-tagger for validations at any
  `items.N.` depth. Consume `grammar.Block.Properties()` iterator,
  filter by `Property.ItemsDepth > 0`, route to
  `items.Validations` via the existing `ValidationBuilder`
  methods. No enum, no type-level.
- **Step 7.items** — no-op for items alone; nested items depth is
  already captured in step 6 since `ItemsDepth` is part of every
  Property.
- **Step 8.items** — enum handling with the new sub-parser. Inline
  JSON + comma-list parsing at the items level.
- **Step 9.items** — no-op (items has no type-level annotations).
- **Step 10.items** — remove `internal/builders/items/taggers.go`'s
  regex-based path; the new bridge-tagger stays.

Items is small enough that steps 6–10 are likely **one commit**,
not five. Only split if parity reveals unforeseen complexity.

### 4.3 Schema builder — scope

**Field-level taggers** (12): all of items' nine plus `required`,
`readOnly`, `discriminator`. Same `ValidationBuilder` interface.

**Type-level annotations** (not taggers — direct regex in
`schema.go`):
- `swagger:model [Name]` — model declaration (already a
  `ModelBlock` in the grammar parser's AST).
- `swagger:strfmt <name>` — format hint. `UnboundBlock`
  (parser emits this for strfmt since it has no body).
- `swagger:allOf [Base]` — composition marker.
- `swagger:enum TypeName` — linked-const enum (`AnnEnumDecl`).

**YAML-fence handler** (`NewYAMLParser`) — vendor extensions on
models (`x-*` as a YAML block). Integrate via `Block.YAMLBlocks()`
and `internal/parsers/yaml/`.

### 4.4 Schema — step sequence

- **Step 6.schema** — bridge-tagger for field-level validations.
  All 12 keywords. Parity harness run. No enum, no type-level yet.
- **Step 7.schema** — bridge-tagger for items nesting via
  `ItemsDepth` on properties inside a schema's struct fields.
  Recycles step 6's keyword dispatch.
- **Step 8.schema** — enum: inline forms + const linking. **W2 §2.6
  obligations land here**:
  - Trim per-value whitespace in comma-list.
  - Emit `parse.context-invalid` for `swagger:enum TypeName` with
    zero matching consts.
  - Drop stale `x-go-enum-desc` when inline values override.
  - Inline-wins override rule preserved.
- **Step 9.schema** — type-level annotations. Four sub-cases:
  `swagger:model`, `swagger:strfmt`, `swagger:allOf`,
  `swagger:enum TypeName`. Each dispatches on
  `Block.AnnotationKind()` rather than re-matching a regex.
- **Step 10.schema** — remove `internal/builders/schema/taggers.go`'s
  regex-based paths and the direct regex matches in `schema.go`.
  Keep the target interfaces (`schemaValidations`, `schema.Typable`)
  intact — no builder-target change.

Schema is larger. **Minimum 3 commits** (6, 8+9, 10). Likely
closer to 5 — one per step.

---

## 5. Parity harness — spec-level compare (step 3 deliverable)

**Revised 2026-04-21:** the v1 view-adapter described in the first
draft is dropped. Replaced by a simpler and stricter approach —
run each fixture twice (flag off, flag on), assert the resulting
`*spec.Swagger` values are JSON-equal.

```go
// internal/integration/parity_test.go
func TestParity(t *testing.T) {
    for _, tc := range parityFixtures {
        t.Run(tc.Name, func(t *testing.T) {
            docV1 := mustRun(t, tc.Opts, false)
            docV2 := mustRun(t, tc.Opts, true)
            assertSpecsEqual(t, docV1, docV2)
        })
    }
}
```

### 5.1 Why this over a view-level adapter

- Measures the user-observable contract (the spec) rather than an
  internal-state proxy (per-comment-group views).
- Reuses every existing `TestCoverage_*` fixture — zero
  reconstruction code.
- No lossy reverse-engineering. The v1 adapter would have had to
  map post-build spec fields back to source comment groups —
  intrinsically fragile (multiple comments touch one field; some
  fields are synthesised).
- Failure messages show the exact diverging JSON path —
  `.definitions.Foo.properties.bar.maximum` — which is what's
  useful for debugging, not "the view for `pkg::Foo::bar`
  differs".

### 5.2 What it doesn't catch

If v1 and v2 produce identical specs through different internal
paths (v1 drops a comment but fills the field from elsewhere; v2
extracts from the comment but that field would have been filled
anyway), spec-level compare misses the internal divergence.

**Accepted trade-off:** the spec is the contract; internal paths
aren't user-facing. If we hit a case where this matters (unlikely
during a same-builders migration), we add a targeted view-level
assertion for the specific site, not a whole adapter.

### 5.3 Removal at P6 cutover

The parity suite is **tactical, not permanent**. Once the flag
is removed at P6 and grammar-parser is the only path, TestParity
has nothing to parity-check against — both runs would execute
identical code. At that point the suite becomes a maintenance
burden (extra CI time, false-positive flakes, future-migration
confusion) with zero signal.

Removal is tracked as an explicit task in
`grammar-parser-tasks.md` P6 cutover list, alongside
`Options.UseGrammarParser` removal. The two go together: without
the flag, the test can't do anything useful.

### 5.4 Feature flag — step 4 deliverable

```go
// in internal/scanner/options.go
type Options struct {
    // ...
    UseGrammarParser bool
}
```

When true, swap the scanner's comment-group dispatch to call
`grammar.Parser.Parse()` and route to bridge-taggers. When false,
preserve the legacy pipeline bit-for-bit. No shared state; one
branch or the other runs per comment group.

Flag removed at P6 cutover (task P6.4).

---

## 6. `internal/parsers/enum/` sub-parser (step 5 deliverable)

Per W2 §3, produced as a standalone subpackage before the
schema-enum bridge-tagger consumes it in step 8.

### 6.1 API

```go
package enum

// Parse converts a raw enum-value string into a []any.
// fieldType may be nil (callers without type info get untyped values).
func Parse(raw string, fieldType types.Type) ([]any, []Diagnostic)
```

Shape detection:
- Prefix `[` (trim space first) → JSON array path
  (`encoding/json` unmarshal, surface error on malformed).
- Otherwise → comma-list path, each value trimmed.

### 6.2 Tests

- Every W2 §2.6 case reproduced with both paths.
- Non-scalar values through the JSON path (W2 §1.3 — a fixture
  audit follow-up, but the sub-parser handles them).
- Malformed inputs → `parse.invalid-enum-value` diagnostic.

---

## 7. Risks + flags

- **Schema's type-level dispatch is outside the tagger pipeline.**
  Current `schema.go` calls `parsers.EnumName(cmt)` /
  `parsers.StrfmtName(cmt)` / etc. — direct regex on the type's
  `*ast.CommentGroup`. Step 9.schema refactors these to
  `grammar.Parse(cmt)` + `switch on AnnotationKind`. Lower risk
  than field-level because no tagger interface is crossed, but
  there's more `schema.go` surface to change.
- **Parity-harness v1 adapter complexity.** Reconstructing
  per-comment-group views from post-build spec trees is the
  hardest single piece of P5.1 prep. If it proves fragile,
  consider a narrower approach: only cover fixtures where v1 and
  v2 are known-equivalent by construction (small hand-picked set)
  and run the full-fixture parity only at P6.
- **YAML body extensions.** Schema's `NewYAMLParser` handles
  model-level vendor extensions via YAML. Step 10.schema folds
  this into the bridge-tagger path via `Block.YAMLBlocks()` +
  `internal/parsers/yaml/`. First real consumer of the sub-parser.

---

## 8. Change history

| Date | Change |
|------|--------|
| 2026-04-21 | Initial plan. Dependency order locked (items → schema → params/resp/ops/routes). Ten-step template. P5.1a+b split documented. |
