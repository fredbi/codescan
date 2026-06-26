# Coverage gap analysis — baseline worktree (cover2.out)

**Total baseline coverage: 80.4%** (after the first round of enhancement tests).

Source: `mcp__go-fred-mcp-baseline__coverage_details` against
`.worktrees/baseline/cover2.out` — 469 uncovered blocks across 12 files,
spanning ~100 partially covered functions and ~20 fully uncovered ones.

Every uncovered block below has been classified into one of four buckets:

| Bucket | Meaning | Action |
|---|---|---|
| **A — defensive guards** | panics, invariant assertions, "can never happen" returns | **Skip** |
| **B — invalid-input guards** | error returns for malformed annotations / invalid numeric/bool/URL/email/JSON input | **Optionally** — one "malformed annotations" fixture covers most of these cheaply |
| **C — genuine feature gaps** | real behaviour that simply isn't exercised today | **Prioritise** |
| **D — dead or unreachable** | public-looking symbols that no caller reaches, or code behind call-sites that themselves are uncovered | **Mark for deletion in the refactor**, do not test |

## Bucket A — defensive guards (skip)

These are either `panic(errInternal)` assertions or `return err` paths immediately after them.

- `assertions.go:28`, `assertions.go:36` — `mustNotBeABuiltinType`, `mustHaveRightHandSide`. Panic by contract; every caller has already excluded these preconditions by checking `Pkg() != nil` / `Rhs() != nil`.
- `schema.go:1395-1413` (`buildEmbedded`: Union, TypeParam, Chan, Signature, default) — the AST already rejects embedding a chan/signature/union; the `log.Printf` warnings are unreachable without a broken `go/types` dump.
- `schema.go:1453-1464` (`buildNamedEmbedded`: Union/TypeParam/Chan/Signature) — same reasoning.
- `schema.go:1366-1383` (`buildNamedAllOf`: TypeParam/Chan/Signature/default) — same.
- `schema.go:234-245` (`buildFromDecl`: TypeParam/Chan/Signature/default), `472-488` (`buildNamedType`: same) — same.
- `schema.go:1313-1324` (`buildAllOf`: Pointer recursion + TypeParam/Chan/Signature) — the recursion on `*types.Pointer` *is* reachable; the warning branches are not.
- `parameters.go:310-311`, `responses.go:196-197` etc. — default branches of `switch tpe.(type)` that the go/types switch already exhausts for well-formed code.
- Every `if err != nil { return err }` immediately after a `mustHaveRightHandSide`/`mustNotBeABuiltinType` guard: unreachable because the panic fires first on malformed input.
- `specBuilder.Build` and every sub-`buildXxx` error-propagation block (`spec.go:55-56,59-60,63-64,68-69,72-73,76-77,80-81,84-85`, plus each `buildX` wrapper at `104-105`, `120-121`, `130-131`, `176-177`, `191-192`, `206-207`, `224-225`, `161-162`) — these only fire when an inner builder returns an error. The inner errors themselves are bucket B; covering Build's own error plumbing adds no signal.

**Count: ~120 blocks. Skip.**

## Bucket B — invalid-input guards (one fixture covers most)

Uniform pattern: `strconv.ParseFloat` / `ParseBool` / `ParseInt` returning an
error on a malformed annotation value, plus the "empty lines" guard at the
top of every `setXxx.Parse`.

One compact fixture containing an annotated struct with deliberately
malformed validator tags (`maximum: abc`, `minimum:`, `multipleOf: NaN`,
etc.) would cover:

- `setMaximum.Parse`, `setMinimum.Parse`, `setMultipleOf.Parse`, `setMaxLength.Parse`, `setMinLength.Parse`, `setMaxItems.Parse`, `setMinItems.Parse`, `setUnique.Parse`, `setPattern.Parse`, `setCollectionFormat.Parse`, `setEnum.Parse`, `setDefault.Parse`, `setExample.Parse`, `setRequiredParam.Parse`, `setReadOnlySchema.Parse`, `setDeprecatedOp.Parse`, `setDiscriminator.Parse`, `setRequiredSchema.Parse`, `setSchemes.Parse`, `setSecurity.Parse`, `setOpResponses.Parse`, `setOpExtensions.Parse`, `setMetaSingle.Parse`, `setCollectionFormat.Parse` — all with the same (`669-670`, `675-676`)-shaped pattern of "empty line guard + parse error".

Other invalid-input guards that a malformed-YAML fixture could reach:

- `metaVendorExtensibleSetter:60-65`, `infoVendorExtensibleSetter:77-82`, `metaSecurityDefinitionsSetter:48-49`, `buildExtensionObjects:1556-1595` — bad JSON / non-`x-` extension keys.
- `yamlSpecScanner.UnmarshalSpec:392-427` — YAML parse errors and empty-section handling.
- `parseContactInfo:188-189` — `mail.ParseAddress` on a malformed contact.
- `parseValueFromSchema:926-927,938-940` — invalid JSON in `default`/`example` for object/array-typed schemas. **Partially fixed** in the current enhancement tests; the `nil schema` branch (938-940) is still untested.
- `setOpResponses.parseResponseLine:1384-1421` — malformed response descriptions.
- `setInfoContact:171-172,175-176`, `setInfoLicense:203-204`, `setInfoVersion:162-163` — empty-tag early returns.
- `setSwaggerHost:145-146` — the one you flagged: empty `Host:` falls back to `"localhost"`. Genuinely **bucket C** (a feature, not an error), noted here for grouping because it lives alongside other `setInfoXxx` empty-line cases.

**Recommendation:** one tightly-scoped `fixtures/enhancements/malformed-annotations/` fixture that error-tests by asserting `Run` returns a non-nil error — no golden file, just `require.Error(t, err)` per scenario. A separate small meta-empty-lines test exercises the "bucket B adjacent" empty-line defaults (`setSwaggerHost` → "localhost", `setInfoVersion` → empty, etc.).

**Count: ~90 blocks. Cover with 2 compact fixtures.**

## Bucket C — genuine feature gaps (prioritise)

Grouped by investment size. Each line names the function and the *reason*
the path isn't covered today, not just the line numbers.

### C.1 High-leverage gaps (one fixture unlocks many lines)

1. **Interface method extraction** — `processAnonInterfaceMethod` (11 gaps, 707-759), `processInterfaceMethod` (7 gaps, 963-996), `buildFromInterface` (843-874). Today no fixture has an interface with: non-exported methods, methods with parameters or multiple return values, methods with `swagger:ignore`, methods with `swagger:name` overrides, methods whose return type is `*T` (x-nullable), methods annotated with `swagger:strfmt`. **→ one `interface-methods/` fixture with ~8 methods covers all paths.**

2. **Parameter alias handling (non-transparent)** — `parameterBuilder.buildAlias` 242-284 (25% coverage), `buildFieldAlias` 408-470 (34%). Today the `transparentalias` fixture only exercises the transparent branch. The non-transparent (default) path — aliases resolving to named types, aliases-of-aliases, aliased body vs non-body — is dark. **→ extend existing `transparentalias` fixture with a non-transparent variant.** This also picks up the `setupRefParamTaggers` (0% coverage) path, since aliases to body types set `ps.Ref` before the tagger runs.

3. **Response alias handling (non-transparent)** — `responseBuilder.buildAlias` 321-367 (26%), `buildFieldAlias` 416-428. Same shape as the parameter gap above. **→ covered by the same fixture.**

4. **Named-field edge cases in responses** — `responseBuilder.buildNamedType:277-313` (the `isStdTime` / `strfmt` / no-model-annotation-with-fallback branches), `buildFromStruct:433-456` (embedded-in-response), `buildNamedField:373-391`. Today no response fixture has a field of type `time.Time`, or a response whose body is `*T`, or a response with an embedded struct that has its own `swagger:response` annotation. **→ extend `fixtures/enhancements/responses-edges/`.**

5. **`buildNamedBasic` strfmt / defaultName / typeName** — `schema.go:494-539`. The named-basic + `swagger:strfmt` / `swagger:default` / `swagger:type` / `swagger:alias` branches. We have strfmt on arrays now, not on named basic types like `type Email string` + `swagger:strfmt email`. **→ small fixture.**

6. **`buildDeclAlias` ref-aliases deep chain** — `schema.go:618-684`. Covers the `refAliases: true` path where an alias points at another alias pointing at a named type. Existing tests only go one level deep. **→ extend the refaliases integration test with a 2-level alias chain.**

### C.2 Medium-leverage gaps (one fixture each)

7. **`setPathOperation` operation-ID conflict** — `operations.go:117-124`. Two `swagger:operation` blocks with the same ID should collide; never exercised. Small fixture with two duplicate IDs asserting an error.

8. **Required-schema tag variations** — `setRequiredSchema.Parse:1114-1138`. Multiple comma-separated required names, trimming, required-with-empty-list. Small property-level test.

9. **`findEnumValue` description building** — `application.go:475-491`. Needs (a) const blocks with doc comments per value, (b) multi-name ValueSpec (`const A, B Status = "a", "b"`). Small fixture with doc-commented enum.

10. **`buildFromTextMarshal` edge cases** — `schema.go:288-319`. Custom `MarshalText`/`UnmarshalText` types with strfmt tag, with nil package, with builtin underlying.

11. **`setOpExtensions.Parse`** — `parser.go:1695-1717`. YAML extension block at operation level (as opposed to meta/schema level).

12. **`buildExtensionObjects`** — `parser.go:1556-1595`. Nested extension blocks with arrays and maps.

13. **`yamlSpecScanner.UnmarshalSpec` happy paths** — `parser.go:408-427`. Multi-line YAML definitions under `definitions:` / `responses:`. Several existing meta tests skim these, but the `definitions`/`responses` inline-YAML paths are untested.

14. **`detectNodes` annotation detection branches** — `application.go:780-821`. Each branch detects a specific annotation kind (`swagger:route`, `swagger:operation`, `swagger:model`, `swagger:parameters`, `swagger:response`, `swagger:meta`). The ones missed are the less-common combinations (files with `swagger:meta` alongside operations, files with only comments). Low-priority.

### C.3 Low-leverage per-function gaps (investigate case-by-case)

15. **`scanCtx.DeclForType:357-367`, `PkgForType:383-388`, `FindDecl:284-310`** — all the "not found" branches, already triggered by existing tests via indirect paths (`FindModel` sits on top of these). Coverage gap here likely just means these early-return branches don't fire in the current fixture set. Not worth targeted tests.

16. **`buildDeclNamed:251-269`, `buildNamedStruct:547-565`, `buildFromDecl:186-196`** — `strfmt` / `typeName` / `defaultName` branches on named struct declarations (not on named basic). Less common than C.5.

17. **`parseTags:1290-1291,1330-1334,1336`** — `yaml` / `xml` / custom tag parsing. Low priority; we don't use non-JSON tags in fixtures.

18. **`sectionedParser.parseLine:568-578`, `sectionedParser.Parse:556-557`** — edge cases in line-buffer handling (trailing blank lines, mid-section blank lines). Likely already exercised by some path; low priority.

**Count: ~200 blocks. Prioritise C.1 + C.2 (≈150 blocks) for a second round.**

## Bucket D — dead or unreachable (propose deletion)

These are symbols the refactor can drop without loss. Confirm by grepping
call sites in both trees before deleting.

- `paramTypable.In`, `paramTypable.Level`, `paramTypable.SetRef`, `paramTypable.AddExtension` — never called in production.
- `itemsTypable.In`, `itemsTypable.Level`, `itemsTypable.Schema`, `itemsTypable.AddExtension`, `itemsTypable.WithEnum` — same.
- `responseTypable.In`, `responseTypable.SetSchema`, `responseTypable.CollectionOf`, `responseTypable.AddExtension`, `responseTypable.WithEnum` — same.
- `schemaTypable.In` — same.
- `parameterBuilder.makeRef`, `responseBuilder.makeRef` — only reachable from currently-uncovered `buildAlias` branches that we plan to test; if those tests *still* don't reach these helpers, the helpers are dead.
- `testError.Error` — `errInternal` is panicked, never compared via `Error()`. Safe to drop the method receiver.
- `scanCtx.PkgForPath` — already relocated to `export_test.go` in the refactor. Baseline gap is a non-issue.

**Recommendation:** run `gopls find-references` on each before deleting;
stage the deletions in a separate cleanup commit on the refactor branch,
after new tests land.

## Proposed rollout

| Round | Scope | Goldens | Expected coverage gain |
|---|---|---|---|
| R2a | C.1.1 interface methods fixture | 1 | +1.5 pp |
| R2b | C.1.2 + C.1.3 alias non-transparent fixture (params & responses) | 2 | +2.0 pp |
| R2c | C.1.4 response-field edges fixture | 1 | +1.0 pp |
| R2d | C.1.5 named-basic strfmt fixture | 1 | +0.5 pp |
| R2e | C.1.6 ref-alias chain fixture | 1 | +0.3 pp |
| R2f | C.2.9 enum-with-docs fixture | 1 | +0.3 pp |
| R2g | Malformed-annotations bucket-B fixture (no goldens, error-only) | 0 | +1.5 pp |
| R2h | Meta-empty-lines bucket-B (no goldens) | 0 | +0.3 pp |
| R3 | C.2 items 7, 8, 10-14 bundled | ~4 | +1.5 pp |
| R4 | Bucket-D deletions | — | — (refactor only) |

**Estimated final coverage after R2+R3: ~87-88%**, with the remaining
~12-13% being bucket A (defensive) — which is fine to leave uncovered.
