# Quirks F-series — fix pass (single branch)

Branch: **`fix/quirks-F-series`** (worktree `.worktrees/fix/quirks-F-series`, base
`30b11ff`, ~5 backlog test-locks behind lot1 — irrelevant to these fixes).
Source of quirks: [`doc-site-quirks.md`](doc-site-quirks.md).

Ground rules (from doc-site-quirks.md): **one commit per F** (DCO `-s`,
`Co-Authored-By: Claude`, fix + golden in the same commit); witness-then-fix
(capture pre-change golden, regenerate, the diff is the audit trail);
`UPDATE_GOLDEN=1 go test ./...` to regenerate; lint clean before push.
Status icons: ⬜ todo · 🟡 in progress · ✅ done · 🔵 blocked · ⏸ deferred.

## Out of scope / already done

- **F6** (`deprecated:` on model field) — ✅ already fixed by #3138 (`x-deprecated`
  on fields, `MarkDeprecated`), present in this base. Documented separately.
- **F7** (gofmt-canonical meta YAML) — ✅ fixed on its own branch
  `fix/scanner-gofmt-meta-yaml` (commit `0545527`); under separate review.
- **F8-quotes** (#2899, string `example`/`default` retain quotes) — ⏸ NOT this
  branch. Its witness is the `#2899` lock test on `fix/backlog-lot1`; the fix
  (`validations/coerce.go` unquote step + flip that assertion) belongs there.
  Flagged for Fred.

## Decisions (settled with Fred, 2026-06-14)

1. **F1/F2/F4 unified** — `+swagger:model` on a named type carrying a
   `swagger:strfmt` / `swagger:type` / `swagger:enum` override should produce a
   **first-class definition + `$ref`** (not inline + orphan), applied uniformly.
2. **F3** — emit a diagnostic on unsupported `swagger:type` values; additionally
   **support arrays of simple types** via the inlined-response-body `[]type`
   syntax (mirror `routebody.stripArrayPrefixes`). `file` keeps the diagnostic +
   ignore (deferred — file types tracked elsewhere).
3. **F4 (collection)** — fix bare `swagger:enum` (no name arg) const collection,
   or document the name as required (verify first). Folds into the F1/F2/F4 work.
4. **F5** — **support `swagger:name` on struct fields** (override the
   json-tag/field-name derivation), matching the documented "field OR method".
5. **F8-alias** — **deprecate `swagger:alias`**: warn + treat as no-op/dissolve,
   correct `annotations.md` (its real behavior was a primitive force-inline, never
   the documented `$ref`-to-target), steer users to `swagger:model` / the
   `RefAliases`/`TransparentAliases` Options.
6. **F9 (HIGH)** — fix the infinite loop on `swagger:model`-annotated Go aliases;
   `swagger:model`-on-alias then yields a definition + `$ref` (consistent with #1).

## Per-quirk plan

### F5 — `swagger:name` on struct fields ✅ (commit `5a05231`)
- **Approach.** Today `swagger:name X` is honoured on interface methods only. Wire
  the field-name derivation in the schema builder to consult a field-level
  `swagger:name` override before falling back to json-tag/Go-name.
- **Files.** schema property-naming path (find where field props get their name);
  classifier `findAnnotationArg(cg, grammar.AnnName)`.
- **Witness.** `swagger:model` struct with a tag-less field carrying
  `// swagger:name balance`; assert the property key is `balance`.
- **Risk.** Low. Verify precedence vs an explicit json tag (override wins).

### F3 — `swagger:type` validation + array-of-simple ✅ (witness `cc59e88`; reconciliation `80300b6`)
- **Witness landed (`1d31701`).** `fixtures/quirks/swagger-type-matrix` +
  `quirk_type_matrix.json` + `TestQuirk_TypeMatrix` capture today's contradictory
  rendering as the documented-current baseline. Per Fred's call, the actual fix
  is split into a separate, reviewed follow-up.
- **Key finding (two disjoint vocabularies).** swagger:type is validated/resolved
  by TWO layers with near-disjoint accepted sets (overlap only `{string,object}`):
  - **grammar** `typeRefVocabulary` (`disambiguate.go`): OAS2 names
    `string/integer/number/boolean/array/object/file/null`; else
    `parse.invalid-type-ref` **(SeverityError)** at `parser.go:581`.
  - **builder** `SwaggerSchemaForType` (`resolvers.go:32`): Go-builtin names +
    `string`+`object`; else silent fallthrough.
  Consequences (pinned by the witness, TODO-F3 markers): `integer/number/boolean/
  file` are grammar-valid but builder-dropped (look valid, no-op); `int64` & other
  Go builtins render but raise a grammar ERROR ("errors but works"); `array` works
  only via fallthrough to a Go slice (Mode-2 idiom, 4 fixtures rely on it);
  `[]string`/`badValue`/scanned-names fall through; strfmt-vs-type precedence is
  site-dependent and lossy.
- **Reconciliation — FINAL design (locked with Fred 2026-06-15).**

  **Core principle:** `swagger:type` is an explicit *inline-this-type* directive —
  it NEVER emits a `$ref`. The default `$ref`-for-definitions remains the
  no-annotation behavior (already coerced to inline for `in:{query|path|header}`
  params + response headers; swagger:type-inline is consistent, no conflict).

  **Arg resolution** (`swagger:type X`; strip N leading `[]` → N-nested array,
  items inlined):
  - **Keyword set — lowercase, case-sensitive:**
    - scalars: `string`/`integer`/`number`/`boolean`/`object`
    - Go builtins: `int64`→{integer,int64}, `uint32`, `float64`→{number,double},
      `bool`, `byte`, `rune`, `error`, … (existing SwaggerSchemaForType map) —
      **no longer a grammar error**
    - `inline` (NEW): expand the field's OWN Go type in place (no $ref); slice →
      array with inlined items
    - `array` (DEPRECATED): same as `inline` for slices + `SeverityWarning`
      "'array' in swagger:type is deprecated, prefer 'inline'"
    - `file`: diagnostic "file override not supported — use swagger:file instead";
      no schema effect (falls through to Go type)
  - **Type-name reference** (anything NOT in the keyword set, case-sensitive
    lookup among known definitions): inline that type's schema in place (no $ref).
    `Custom` → inlined Custom; `Inline` (capital) ≠ keyword `inline` → definition
    "Inline". Unknown → diagnostic "unknown type X" + fall through.
  - **`[]T`:** array, items = resolved T (recursive: `[]string`, `[]Custom`,
    `[]object`, `[][]int64`). `[]inline`/`[]array`/`[]file` → diagnostic.
  - keyword-vs-type-ref disambiguation = membership in the fixed lowercase keyword
    set (a lowercase non-keyword like `badValue` is NOT a type-ref → "unknown
    type").

  **strfmt + type precedence (uniform across schema + simple-schema sites):**
  - `swagger:type` wins (determines + inlines the type).
  - `swagger:strfmt {fmt}` applies as a supplementary format IFF compatible with
    the resolved type: `string` → any; `integer` → `int{n}`/`uint{n}`; `number` →
    `int{n}`/`uint{n}`/`float32`/`float64`; else incompatible → ignore format +
    `SeverityWarning` "format {fmt} incompatible with type {t}; ignored".
  - Compatibility = a SHARED utility next to the simple-schema validation utils
    (`internal/builders/validations/`), reusable by simple-schema validation.

  **Layers to change:**
  - grammar: relax `argTypeRef`/`typeRefVocabulary`/`parse.invalid-type-ref` so
    the lexer accepts any plausible type-ref token (ident, `[]`-prefixed, keyword)
    WITHOUT erroring on unknown names — semantic validation moves to the builder
    (only it knows scanned definitions). Keep an error only for structurally
    malformed tokens.
  - builder: new unified resolver (keyword | `[]T` | type-ref-inline |
    file/unknown diagnostics) with simple-schema-vs-schema awareness + definition
    lookup; reroute the 5 swagger:type sites; add the format-compat util; apply
    strfmt precedence.
  - witness `quirk_type_matrix.json` flips (TODO-F3 assertions); the 4 `array`
    fixtures gain a deprecation warning (spec output unchanged).

  **Implementation order:** (1) format-compat util (isolated, testable) →
  (2) unified builder resolver → (3) grammar relaxation → (4) reroute sites +
  strfmt precedence → (5) flip witness + array-fixture warnings. Lands as the F3
  reconciliation commit (fix + golden).

  **lot1 alignment (rebased 2026-06-15 onto lot1 tip `7cba93a`; clean, suite +
  witness golden green — no drift).** Three overlapping lot1 fixes inform (don't
  conflict with) this design:
  - **#1512** — `swagger:strfmt` ALONE stays `{type:string, format:X}` (locked).
    F3 only changes the strfmt+type-together case; the format-compat util must
    treat `type:string` as accepting any format (consistent with #1512).
  - **#1088** — an existing simple-schema safety net
    (`schema.validateSimpleSchemaOutcome` + `SimpleSchemaProbe` / `ItemsTypable`)
    already dissolves illegal `$ref`/object shapes in simple-schema (incl. array
    items): named primitive → inline, object element → `{}` +
    `CodeUnsupportedInSimpleSchema`. The always-inline swagger:type output (esp.
    `[]Custom` in a simple schema) flows THROUGH this net — do not reimplement it.
  - **#1133/#1174** — unrelated (unsupported Go func-type scanning robustness).
- **Files (fix).** grammar `disambiguate.go` (typeRefVocabulary + `[]` prefix),
  `parser.go` (AnnType arm), `lexer.go` (argTypeRef); builder
  `resolvers/resolvers.go` (SwaggerSchemaForType) + the swagger:type sites in
  `schema/walker_classifiers.go` & `schema/fields.go`.
- **Risk.** HIGH — two-layer change on a fragile, habit-laden surface; the witness
  golden is the guard.

### F1/F2/F4 — strfmt/type/enum + `swagger:model` → definition + `$ref` ✅ (`8e20d2f` A+B, `716d8c1` C)

> **Branch split (2026-06-15).** F5/F3/F9/F8 + the F1/F2/F4 *witness* merged into
> `fix/backlog-lot1` (merge `b6ad383`). The F1/F2/F4 **fix** (HIGH-risk,
> load-bearing) is isolated on **`fix/quirks-model-override`** (worktree
> `.worktrees/fix/quirks-model-override`, off lot1). The witness
> (`coverage_quirk_model_override_test.go` + `quirk_model_override_matrix.json`)
> is the live baseline there; the fix flips it.

**Implementation subtlety found:** the F8 `swagger:alias` deprecation diagnostic
lives inside `classifierNamedBasic`. So the override classifiers + their
diagnostics must run ONCE at definition-build (Part B); field-sites for model
types just `$ref` (Part A skips classifiers — no duplicate diagnostics, F8 stays
green). Parts: B (buildFromDecl def-build cascade) → A (buildNamedType field gate
on `decl.HasModelAnnotation()`) → C (bare `swagger:enum` grammar/collection).

- **Decisions (Fred).** (1) `swagger:model` is the opt-in: WITHOUT it an override
  type inlines as today; WITH it the type publishes a first-class definition
  carrying the full override schema and referencing fields `$ref` it (accepts the
  field-site inline→`$ref` change). (2) Support BARE `swagger:enum` on a type decl
  (infer the enum name = the decl's type; collect its consts).
- **Root (verified).** Two halves disagree:
  - *Definition build* — `buildFromDecl` named case (`schema.go:146-148`) builds
    from `s.Decl.Spec.Type` (the bare underlying RHS, e.g. `string`), bypassing the
    override classifiers → orphan `{type:string}` (strfmt format dropped, enum
    values absent). F2's `{type:string}` is already right post-F3.
  - *Field site* — `buildNamedType` fires the override classifiers
    (`classifierNamedStructStrfmt` :419, `classifierNamedBasic` strfmt/enum arms
    :432) which short-circuit-INLINE before the `resolveRefOr` `$ref` pivot.
- **Plan.**
  - **Part B (definition carries override).** In `buildFromDecl` named case, apply
    a per-type override (strfmt → `Typed(string,fmt)`; enum → collect + `WithEnum`;
    type → `resolveTypeOverride`) to the definition target before the bare-underlying
    fallback (mirrors `buildDeclAlias`'s strfmt handling at `schema.go:182`).
  - **Part A (field → `$ref`).** In `buildNamedType`, when the named type is
    `swagger:model` (`decl.HasModelAnnotation()`), skip the inline override
    classifiers and go to `resolveRefOr`. Non-model override types inline as today.
  - **Part C (bare `swagger:enum`).** Relax the grammar `AnnEnum` missing-arg error
    to allow the bare form; builder infers the enum name from the decl's type and
    collects its consts.
- **Witness.** `fixtures/quirks/model-override-matrix` + `quirk_model_override_matrix.json`
  (documented-current, committed `86696b5`) — flips to the fixed expectations
  (definition carries override, fields `$ref`, bare enum collects) at the fix.
- **Risk.** HIGH — load-bearing schema-discovery; witness golden is the A/B guard;
  validate the broader corpus goldens (strfmt/enum fixtures) for inline→`$ref` drift.

### F8-alias — deprecate `swagger:alias` ✅ (`cfd8917`)
- **Done.** Empty-sink deprecation: `classifierNamedBasic` drops the AnnAlias
  force-inline trigger + emits `validate.deprecated`; primitives now `$ref`
  (default handling). Quirk premise was inaccurate (it force-inlined primitives,
  wasn't a no-op). `TestAliasedTypes` updated; witness `alias-deprecated` +
  `TestQuirk_AliasDeprecated`; `annotations.md` rewritten. Below: original plan.
- **Approach (confirmed with Fred).** Make `swagger:alias` an **empty sink**: the
  grammar keeps recognizing it (stays `AnnAlias`, still stripped from prose, never
  leaks into a description), but the builder applies **no schema effect** — the
  current primitive force-inline at `walker_classifiers.go:156` is removed, and the
  type is handled exactly as if the annotation were absent (default Go handling).
  When the builder walks over the annotation it emits **one structured diagnostic**:
  `Code = validate.deprecated-annotation`, `SeverityWarning`, `Pos` at the
  annotation, message naming `swagger:alias` + the migration path (swagger:model /
  RefAliases / TransparentAliases). Structured-only via `RecordDiagnostic` — there
  is no default stderr printing; delivery to a human is the CLI/frontend's job (the
  TUI already consumes diagnostics; go-swagger does not yet). Correct the
  `annotations.md` section (drop the false `$ref`-to-target claim).
- **Files.** `internal/parsers/grammar/diagnostic.go` (new `CodeDeprecatedAnnotation`),
  `walker_classifiers.go:156` (the `AnnAlias` arm → diagnostic + fall through),
  `docs/doc-site/maintainers/annotations.md:395`.
- **Witness.** A type carrying `swagger:alias` → assert (a) the
  `validate.deprecated-annotation` diagnostic is emitted via an `OnDiagnostic`
  sink, and (b) the emitted schema is identical to the un-annotated form.
- **Risk.** Low. Removing the primitive force-inline must not change a named
  primitive's default emission — verify the fall-through path matches.

### F9 — `swagger:model` on a Go alias infinite loop ✅ RESOLVED + locked (`5a56679`)
- **Outcome.** Does not reproduce on the current base (exact doc-site trigger
  completes in all three alias modes). Dissolved by intervening lot1
  builder/discovery/grammar work. Locked by `fixtures/quirks/alias-model` +
  `TestQuirk_AliasModelNoHang` (regression → CI `-timeout`). Clears the doc-site
  first-class-alias how-to. Original investigation notes below.
- **Approach.** Reproduce against `alias-calibration-embed` `BaseAliasModeled`
  with a **timeout-guarded** test (never an unbounded hang). Locate the loop in
  the first-class-alias build path (expand / ref structural copy;
  `astutil.PathEnclosingInterval` / `go/ast.File.End` per the doc) and add the
  termination guard. Confirm `swagger:model`-on-alias then emits a definition +
  `$ref` (consistent with the unified decision).
- **Files.** `internal/builders/schema/schema.go` (`buildAlias`/`buildDeclAlias`),
  `internal/builders/spec/spec.go` (discovery dedup).
- **Witness.** The hanging fixture, run with a bounded timeout; assert completion
  + the emitted definition.
- **Risk.** HIGH — investigative; guard must not change non-looping alias modes
  (TransparentAliases already completes). Narrow against the non-hanging
  `BaseAliasModeled` calibration.

## Suggested sequence

1. **F5** (small, settled — establishes rhythm + golden flow)
2. **F3** (localized; sets up the array-syntax helper)
3. **F1/F2/F4** (the big unified change; A/B goldens with witnesses)
4. **F8-alias** (deprecation + doc correction)
5. **F9** (crash — investigative, timeout-guarded; highest severity, last only
   because it is the most exploratory — can be pulled forward if preferred)

Each lands as one commit; natural pause points for review between Fs.
