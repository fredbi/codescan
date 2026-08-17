# Open quirks — the live register

**This is the single source of truth for known-open scanner quirks.** If an item
is not here, it is not open: the historical registers (`archive/observed-quirks.md`
Q-series, `archive/deferred-quirks.md` D-series, `archive/quirks-F-series-fix.md`,
`archive/doc-site-quirks.md`) are **provenance only** and their per-entry status
lines are stale — see §4.

Numbering continues the Q-series so the references already in the code
(`in_normalize.go` cites Q27, `coverage_interface_name_verbatim_test.go` cites Q9)
keep resolving against the archived file.

Rules for this file:

- an item leaves only when it is **fixed** (name the commit) or **reclassified**
  (name where it went);
- a quirk that belongs to a recorded feature is **described in the feature doc**,
  not here — this file keeps a one-line pointer instead (§3);
- claims are verified against goldens or a probe before being listed, not
  inherited from an older document.

Last verified: **2026-08-17** (pre-v0.36.4 sweep, re-checked at `5e4cf04d`). One genuinely new quirk since
2026-08-05: **Q49**, the `swagger:meta` group mix-up, found and fixed the same day in PR #119. Everything
else that shipped in between was already-registered work — Q37 closed with PR #94, and PR #99 / #89 were
fixes, not findings. Two things found in that window are
deliberately *not* here, because this file is the **scanner** register and neither is a scanner defect:

- `-color=always` cannot colour a run whose stdout is redirected — `archive/genspec-cli.md` §Open 4. A CLI defect,
  with the one-line fix and the reason it wants a deliberate decision written up there.
- Two extractor defects in the benchmark corpus unpacker (a guard with a platform-dependent verdict, and a
  measurement gate that resolved before it validated) — `benchmarks-cleanup.md` §6, both fixed in PR #118.

**Commit hashes rewritten 2026-08-17.** 28 of the 37 hashes this file cited were **pre-rebase and resolved
to nothing** — the branches they were written on (`fix/strfmt-dispatch-symmetry`, `fix/coverage-quirks`, …)
were rebased on merge, so every "FIXED (`abc1234`)" named a commit no ref contained. All 28 were matched to
their master equivalents by commit subject and replaced (52 citations). One, `8e20d2f`, is not an object in
this repository at all and was left alone — it predates the rewrite or belongs to another repo.

**Rule for whoever closes the next quirk:** cite the hash **after** the branch merges, not the one you
committed on. A register whose evidence does not resolve is worse than one that names a PR — if the fix is
still on a branch, name the branch and the PR, and come back for the hash.

Previous verification: **2026-08-05** on branch `fix/coverage-quirks`.

- FIXED: Q31, Q32, Q34, Q35, Q36, Q37, Q38, Q39, Q40, Q41, Q42, Q43, Q44, Q46, Q47, **Q49**.
- 🟦 Documented, kept deliberately: Q48 (walk arms Swagger 2.0 cannot reach).
- Q46–Q48 came out of the coverage sweep (`b06eae8e`) — see each entry's provenance note.
- ⛔ WON'T DO: Q33 (promoted marshaller on an embed — read its Resolution before reopening).
- ✅ Q37 UNPARKED 2026-08-06, NARROWED then **RESOLVED 2026-08-08**, merged as PR #94 (`9376e2e1`): the
  declaration-contract half landed in `master` (PR #90); the `responses.go:327` residual turned out to
  be **unreachable code**, not a soft gate, masking a wider hole one level up (a `swagger:response` on
  an alias to any struct-underneath stdlib type emitted no schema at all) — closed by hoisting
  `ApplyStdlibSpecials` into `responses.buildNamedType`. The "nine hand-rolled subsets" are eight
  legitimate refusal guards; see the entry's correction note.
- 🔍 Latent, deferred super-niche: Q45 (Q33's composition premise vs. a plain inlining embed).
- §2–§5 last verified 2026-07-30 · repo side synced by PR #68.

---

## 1. Open — needs a decision

### Q31 — ✅ FIXED (`53747caa`) — `json:"-"` handling now matches encoding/json

**Status:** FIXED on `fix/strfmt-dispatch-symmetry`, awaiting review. The "decision open" that held
this since 2026-07-30 dissolved once `swagger:omit` shipped (PR #67): being faithful to Go no longer
leaves the author without a way to express the intent.

**Two defects, opposite directions**, both verified against Go itself:

| shape | Go | codescan before |
|---|---|---|
| `json:"-"` re-declaring a promoted field | `{id,name,age}` — age survives from the embed | `{id,name}` — deleted |
| `json:"-,"` / `json:"-,omitempty"` | `{"-": …}` — named literally `-` | `{}` — dropped |
| *(controls, already correct)* real-name re-declaration · plain `-` · embed tagged `-` | — | ✅ |

The second was **not in the original write-up** — found by probing all five shapes rather than the
reported one. encoding/json compares the WHOLE tag to `-`; splitting on the comma first conflates
"ignore" with "name it `-`". Fixed in `resolvers.jsonTagIgnores`, with `-` accepted as a name from
the `json` tag only.

**Why it became decidable.** The old blocker was that faithfulness surprises an author who wrote
`json:"-"` meaning "drop it". `swagger:omit` now does that honestly, and
`scan.shadowed-embed-field` already fired on the exact shape naming it. So: emit what the wire
carries, and keep pointing at the tool that expresses the other intent.

**The witness is differential** — this corpus has an ORACLE, which is why it was the right one to
take: the answer is not a design choice but whatever encoding/json produces. The fixture module
marshals its own types into `wire.golden.json` (the only test in that module — it earns the
exception because only that package can marshal those types, and the library must not gain a
dependency on its own fixtures); `TestJSONTagFidelity` asserts the emitted property sets equal it.
Neither side hard-codes an expectation.

**Three more fixtures were asserting the defect** (instances 5–7 of that pattern, see
[[feedback_test_input_space_vs_contract]]): `TestOverridingOneIgnore` asserted 2 properties where
the wire carries 3; `nomodel.go`'s `IgnoredOther` was named for an assumption, not a behaviour; and
`default-allof-embeds-override`'s package doc credited Go's depth rule for shadowing a `-`
re-declaration, which it does not do. All corrected.

**Unblocked:** repointing `fixtures/bugs/1992` at the mechanism go-swagger#1992 is actually about
(it currently witnesses `readOnly`, which belongs to #1063). Not done here.

**Blast radius:** 3 goldens, all restoring a property that had been evicted, plus prose.

### Q32 — ✅ CLOSED (`1b0e7b2f`, `5ad1df18`, `ed9897cb`) — classifier annotations on an ALIAS declaration

**Status:** CLOSED on `fix/strfmt-dispatch-symmetry`, awaiting review. Every classifier annotation on
an alias declaration was dropped silently; each has now been fixed or, where unfixable, made loud.

| annotation | outcome |
|---|---|
| `swagger:strfmt` | **fixed** `1b0e7b2f` — honoured at every use site, in all three builders |
| `swagger:type` | **fixed** `5ad1df18` — every form; brought a SimpleSchema legality gate with it |
| `swagger:enum` | **diagnosed** `ed9897cb` — unfixable by construction, so it now says so |
| `swagger:default` | left this entry entirely: it never worked anywhere → [Q35](#q35), retired |

> **The live account is `alias-override-symmetry.md` + its ledgers.** Two witness families —
> `strfmt-symmetry-*` (4 packages, 12 goldens) and `type-override-symmetry` — compare each alias half
> against its named half. The strfmt matrix was 56 of 69 cells asymmetric; it and the type matrix are
> now 0, but for one deferred cell ([Q40](#q40)). Run
> `go test ./internal/integration/ -run 'TestStrfmtSymmetry|TestTypeOverride|TestAliasEnum' -v`.

**Why `swagger:enum` could not be fixed.** Members are collected by finding the constants declared
WITH the annotated type. A type alias is erased by the type-checker, so in `type Unsigned = uint64`
with `const Zero Unsigned = 0`, `Zero` is a `uint64` constant indistinguishable from every other one.
Nothing to collect, no plumbing that changes it. The diagnostic names the declaration and suggests
the named form, and is deliberately **silent for an alias to a NAMED enum type** — that works,
because the named type survives the alias, and warning would tell authors their working annotation
is broken.

**What the fixes established beyond closing the quirk:**

- **The author always beats a recognizer** (decided 2026-08-01) — the classifier must run before the
  dissolve reaches a stdlib type, not after.
- **Parameters and responses have their own alias paths**, and needed every fix separately — three
  times over. This is the evidence behind [Q39](#q39).
- **A SimpleSchema legality gate**, since `swagger:type` can name a type a location cannot carry
  while `swagger:strfmt` cannot. An override that will not fit is refused with a diagnostic and the
  Go type stands, because `type` is mandatory there.
- **`swagger:type file` is now a synonym for `swagger:file` and the preferred spelling** (Fred,
  2026-08-02), sharing one location gate rather than growing a second.
- Two adjacent defects fell out and were fixed on the way: [Q41](#q41) (`in:` warned beside a
  classifier) and [Q34](#q34)/[Q36](#q36) from the strfmt round.

### Q33 — ⛔ WON'T DO (decided 2026-08-03) — a struct embedding a PROMOTED MARSHALLER marshals as a bare scalar

**Status:** CLOSED, won't fix. See **Resolution** at the end of this entry for the decision and the
reasoning — read that FIRST; everything above it is the investigation that led there and describes
behaviour we have deliberately chosen to keep. · **SPLIT 2026-08-03**: the half that is a plain bug
with no design question moved to [Q44](#q44) and is FIXED.
(Surfaced 2026-08-01 while designing go1.27 stdlib-uuid detection; probed on go1.26, current master;
re-probed against an `encoding/json` oracle 2026-08-03.)

Embedding a type that implements `encoding.TextMarshaler` promotes `MarshalText`
to the outer struct. With no `MarshalJSON` in the way, `encoding/json` renders
**the whole struct as a string** — the sibling fields never reach the wire.
codescan emits an object and, separately, drops the embed entirely.

```text
type Token [16]byte
func (t Token) MarshalText() ([]byte, error) { return []byte("tok"), nil }
type Embedder struct { Token; Name string `json:"name"` }

json.Marshal → "tok"                        // bare string; "name" absent
codescan     → {object, properties:{name}}  // object; embed contributes nothing
```

Wrong in both directions, and **not silent**: `buildNamedEmbedded` switches on the
embedded type's *underlying* shape, and a TextMarshaler scalar (array / basic /
struct-with-MarshalText) that is not a struct or interface falls to the `default`
arm, which raises a `CodeUnsupportedGoType` Warning and skips the embed. Only the
**interface** arm consults `applyStdlibSpecials` — an asymmetry `schema/README.md`
§embedded documents as intentional.

Uuid-agnostic: any `strfmt.UUID` / `time.Time` embed has the same shape, which is
why a fix is behaviour-changing for existing users and needs its own decision.

**Broader than TextMarshaler — `buildNamedEmbedded` consults no comments at all.** The strfmt-symmetry matrix
(`alias-override-symmetry.md`, F2) witnesses the same gap with an explicit annotation rather than an interface:
`swagger:strfmt` on an embedded type is dropped for the NAMED and the alias half alike, so it is a shared gap, not the
alias asymmetry Q32 tracks.

```text
type FmtStructNamed PlainTarget   // swagger:strfmt duration
type EmbedStructNamed struct { FmtStructNamed; Label string }

codescan → {object, properties:{label,left,right}}   // format gone, properties promoted instead
```

A basic-underlying member is worse — it vanishes entirely (`{object, properties:{label}}`), via the same
`CodeUnsupportedGoType` default arm described above. Witnessed by `TestStrfmtSymmetryComposition`'s `EmbedBasic` /
`EmbedStruct` cells, deliberately left unasserted (a `note`, no `wantNamed`) because what an embed of a formatted type
SHOULD emit is precisely the question Q33 has to answer. Whichever way it goes, those cells and
`strfmt_symmetry_composition_*.json` move with it.

Scoped out of the Q32 fix by decision (Fred, 2026-08-01): same mechanism, but fixing it changes output for named types
too, and it needs this entry's design answer first.

---

#### Update 2026-08-03 — measured against an `encoding/json` oracle

The fixtures module can marshal its own types (the trick [Q31](#q31) used), so this stopped being a
matter of opinion. Four embeds, four Go answers:

| embed | Go wire | codescan |
|---|---|---|
| `Token` (has `MarshalText`) | `"tok"` — **whole struct is a string, `name` gone** | `{object, properties:{name}}` |
| `time.Time` (has `MarshalJSON`) | `"0001-01-01T00:00:00Z"` — same | `{object, properties:{label}}` |
| `FmtStruct` (`swagger:strfmt`, NO marshaller) | `{"left":"","label":"l"}` — object, fields promoted | **matches** ✅ |
| `FmtBasic` (`swagger:strfmt`, NO marshaller) | `{"FmtBasic":0,"label":"l"}` — a NAMED property | property missing, warn+skip → [Q44](#q44) |

**This corrects the entry above.** It treated "annotated embed" and "TextMarshaler embed" as one
story. They are not: **`swagger:strfmt` alone does not change the wire.** Rows 3 and 4 marshal as
objects, so honouring the annotation there would be UNfaithful. The annotation is not the trigger —
**a promoted marshaller is.** Row 3 already matches Go and needs nothing; row 4 is a different defect
and is now [Q44](#q44).

So the scope of Q33 is exactly: *an embed whose method set promotes `MarshalJSON` or `MarshalText`
into the embedding type.*

#### Update 2026-08-03 — the json tag is IRRELEVANT (Fred's hypothesis, probed)

Every tag spelling on the embedded field produces the same bare scalar:

```text
Token                      → "tok"
Token `json:"-"`           → "tok"
Token `json:"tok"`         → "tok"
Token `json:""`            → "tok"
Token `json:",omitempty"`  → "tok"
```

`MarshalText` / `MarshalJSON` is promoted into the **outer type's method set**, and `json.Marshal`
consults that before it looks at struct fields at all — so the field tags are never reached. Note
`json:"-"` does NOT suppress it: a tag cannot remove a method from a method set. Good news for the
design: there is no tag interaction to model.

#### Update 2026-08-03 — what DOES bound the rule (probed)

Three modifiers, and the last one is the problem:

```text
two marshallers at the same depth  → {"Token":"tok","Other":"oth","name":"n"}   ambiguous, neither promoted
outer type defines MarshalJSON     → {"mine":true}                              outer wins
promoted through two levels        → "tok"                                      still promotes
POINTER-receiver marshaller:
  json.Marshal(v)                  → {"PtrTok":[0,0,…],"name":"n"}              not in the value method set
  json.Marshal(&v)                 → "ptr"                                      is in the pointer method set
```

**The pointer case is the hard one for any fix.** The same type marshals two different ways
depending on whether the value or its address is handed to the encoder — and codescan reads a type
DECLARATION, not a use site, so it cannot know which. A faithful answer is not always determinable
from what the scanner can see. That argues for a conservative default plus an explicit escape hatch
rather than silent inference, and it is the question to settle before writing any code.

Ambiguity and an outer `MarshalJSON` are both detectable from `go/types` (`types.NewMethodSet` /
`LookupFieldOrMethod` report exactly the promotion Go applies), so those two are cheap to honour.

**Blast radius, when it is decided:** every `time.Time` embed in every scanned codebase turns from an
object into `{string, date-time}`. That is faithful and it is a visible break.

**Related:** `archive/go127-uuid.md` §Appendix (where it surfaced). Note the go1.27 work
does **not** move this: adding identity-based `recognizeStdUUID` to the canonical
safe set changes nothing here, because the embed never reaches a recognizer at all
(verified against `go127_uuid_spec.json` — `Embedder` still emits
`{object, properties:{name}}`). Witnessed by `TestStdlibUUID`'s embed sub-test, so
whichever way Q33 is decided, that sub-test and the golden move together.

---

#### Resolution 2026-08-03 (Fred) — ⛔ WON'T DO: we do not detect a marshaller on an embed

**The decision: a promoted `MarshalText` / `MarshalJSON` does not change what codescan emits for an
embed.** The whole entry above measures codescan against `json.Marshal` run on the *declared* type
with the *default* marshaller. That is the wrong oracle for the round-trip codescan participates in.

**Why.** In the convention codescan describes, an embed means **composition**, and a composed model
does not use the default marshaller. Look at go-swagger's own generated output —
`go-swagger/fixtures/codegen/tmp/models/with_all_of.go`: `WithAllOf` embeds each of its `allOf`
members *and* generates a hand-written `MarshalJSON`/`UnmarshalJSON` that flattens them. Nothing in
the round trip ever relies on Go squashing an embed. So a promoted marshaller sitting in hand-written
source is not evidence about the wire, and modelling it would be reading intent out of a coincidence.

**What this buys us — the undecidable part disappears.** The pointer-receiver bound recorded above
(`json.Marshal(v)` → object, `json.Marshal(&v)` → `"ptr"`, and a declaration cannot say which) was
the blocker. It only exists if we detect marshallers. We don't, so it never arises. Q33 closes by
decision rather than by solving it, which is the honest outcome — there was no faithful answer to
find.

**The author's escape hatch is unchanged and already works:** `swagger:strfmt` or `swagger:type` on
the *embedded type's own declaration*. Someone who really does want the scalar says so there.

**Recorded so it is not relitigated.** This quirk was raised as an unexplained difference between
code paths, and that shape of finding will recur — the next reader of `buildNamedEmbedded` will
notice again that a promoted marshaller goes unread and assume it is an oversight. It is not: it is
this decision. `embedPromotes`' doc comment in `internal/builders/schema/allof.go` and
`schema/README.md` [§embed-marshaller](../../internal/builders/schema/README.md#embed-marshaller)
both carry the reason at the code.

**Residual, accepted:** for a promoted-marshaller embed codescan now emits
`{"Token": …, "label": …}` (via the [Q44](#q44) fix) where the default marshaller would emit `"tok"`.
Deliberate. Witnessed rather than hidden: `fixtures/enhancements/embed-basic-underlying/wire.golden.json`
records the raw `encoding/json` document — which is why that oracle stores raw documents and not key
sets, since this subject is not an object at all — and `TestEmbedBasicUnderlying` asserts the
divergence explicitly instead of skipping the subject.

**Spun off:** [Q45](#q45) — the decision rests on the composition reading of an embed, and not every
embed becomes an `allOf` member.

### Q45 — 🔍 latent — the Q33 decision assumes composition, but not every embed composes

**Status:** LATENT, deliberately deferred as **super-niche** (Fred, 2026-08-03). No witness, no
reproduction pressure. Recorded only so the reasoning behind [Q33](#q33) is falsifiable rather than
folklore.

[Q33](#q33) closes on the premise that *an embed means composition, and a composed model
round-trips through a hand-written marshaller*. That premise is exactly true for an embed that
becomes an `allOf` member — `swagger:allOf`, or any embed under `DefaultAllOfForEmbeds`. It is an
approximation for a **plain inlining embed**, which is a promotion, not a composition, and whose
round trip may well be the default marshaller after all.

So the honest scope of the Q33 rule may be *allOf member* rather than *every embed*, with the plain
inlining embed judged on its own terms.

**Why it is deferred rather than fixed:** splitting the rule means the same Go declaration produces
different member semantics depending on an annotation elsewhere and on an Option
(`DefaultAllOfForEmbeds`) — the shape most likely to produce the very cross-path divergence this
whole stream has been closing ([Q39](#q39)). It would touch the embed walk, the allOf walk and the
conformance suite together. That is a large change to serve a case nobody has reported, so the
cheaper move is to keep one rule and record why it might be wrong.

**What would reopen it:** a real report where a plain (non-`allOf`) embed of a marshaller type
produces a spec its own generated client cannot round-trip.

### Q34 — ✅ FIXED (`9d92874d`) — decl arm dropped a sequence's format; the items-vs-whole rule was an allowlist

**Status:** FIXED on `fix/strfmt-dispatch-symmetry`, awaiting review. Kept here until the branch
merges, since the entry is what the fix is judged against.

Two defects, one witness (`fixtures/enhancements/strfmt-decl-arraylike`, `TestStrfmtDeclArrayLike`):

1. **The named decl arm dropped the format.** `buildFromDecl`'s underlying-kind switch had arms for
   struct and basic only, so a `swagger:model` array/slice with a `swagger:strfmt` published a
   definition without it — and nothing downstream compensates, because `buildNamedType`'s `refModel`
   gate skips the inline classifiers on the assumption the declaration already applied the override.
   A violation of the same "the author always overrides" principle Q32 settled, by the *named* arm.
   Fixed by adding the missing array and slice arms; `buildDeclAlias` applies the same rule so both
   spellings publish the same definition.

2. **The items-vs-whole decision was a two-name allowlist** (`byte`, plus `bsonobjectid` for arrays
   only) — which is why 1. went unnoticed: every existing fixture used one of those two names, where
   the placement is the same either way. Both names are formats for a byte sequence, so the list was
   standing in for a question about the ELEMENT type. Now asked directly, in
   `common.IsStringLikeSequence`: byte and rune sequences are string-like and take the format on the
   schema; anything else takes it on the items.

```go
type ID     [16]byte  // swagger:strfmt uuid  → {string, format: uuid}
type Emails []string  // swagger:strfmt email → {array, items: {string, format: email}}
```

**Still open, deliberately:** whether a format on a sequence of some OTHER element type describes the
sequence or its members is genuinely ambiguous — it cannot be read off the Go type. It stays on the
items, as it always has. Fred, 2026-08-01: *"whether we should apply the swagger:strfmt to the parent
type or to the item is a different problem, and I have to admit that one is hard to solve."*

**Blast radius:** exactly one pre-existing golden moved — `Signature [64]byte` annotated `password`
emitted an array of 64 password strings and now emits a password string.

### Q35 — ✅ FIXED (`4de3a6e3`) — `swagger:default` retired as a deprecated no-op

**Status:** FIXED on `fix/strfmt-dispatch-symmetry`, awaiting review. Decision taken 2026-08-02
(Fred): deprecate, do not implement.

The annotation never emitted a `default` in any placement or form, and on a named basic type it
returned `handled=true` on a target it had not written — publishing a TYPELESS definition and a
typeless property for every field referencing it, silently. It is now an empty sink following the
`swagger:alias`/F8 precedent: parsed, reported `validate.deprecated`, otherwise ignored.

**Why retired rather than implemented.** Every place OpenAPI 2.0 admits a `default` is already
served, so there was no meaning left to give it:

| OAS 2.0 object carrying `default` | mechanism |
|---|---|
| Schema (model field, type decl) | `default:` keyword |
| Parameter (non-body) | `default:` keyword |
| Items | `default:` keyword |
| Header | `default:` keyword |
| Responses (the default response) | `default` code head in a route's `Responses:` body |

The keyword's registered context set (`CtxParam, CtxHeader, CtxSchema, CtxItems`) is exactly the list
of OAS 2.0 objects that carry a `default`; the response-code head closes the remainder. Verified in
code: `keywords.go:289` and `routes/walker.go:482`.

**No usage to preserve.** One occurrence in the whole corpus —
`fixtures/enhancements/named-basic` — written for branch coverage per
`archive/coverage-gap-analysis.md:75`, and its own comment described the typeless schema as the
expected result. Fixture and golden updated; the test now asserts the deprecation fires and that
`Grade` emits as a plain named int.

**Also changed:** the value argument is now optional. It used to be mandatory, which hard-errored on
the bare form the doc-site documented for years.

**Doc site:** reference page, tutorial section, annotation index and context matrix rewritten; the
published example (bare form on a `var` — a placement no builder reads) removed rather than fixed.

### Q36 — ✅ FIXED (`b72a43d2`) — decl-site values were coerced against an empty type

**Status:** FIXED on `fix/strfmt-dispatch-symmetry`, awaiting review.

A declaration's comment block is dispatched before its Go type is resolved onto the schema, so
`default:`, `example:` AND `enum:` ran through `ParseDefault` / `ParseEnumValues` with
`SchemaTypeOf(ps) == ""` and fell back to the author's raw text. Every other site was always
correct. Measured before the fix:

| decl | schema type | emitted |
|---|---|---|
| `default: 8080` | integer | `"8080"` string |
| `default: 1.5` | number | `"1.5"` string |
| `default: false` | boolean | `"false"` string |
| `default: [1,2,3]` | array | `"[1,2,3]"` — a string holding JSON source |
| `enum: 1,2,3` | integer | `["1","2","3"]` — **a spec no validator can satisfy** |
| `default: auto` | string | `"auto"` ✅ fallback coincides with the correct answer |

**Scope grew during the measurement** (the re-scope checkpoint the plan called for): the report was
about `default:`, and `example:` was predicted to match by symmetry — Fred, 2026-08-02: *"default and
example are very symmetric so any design choice we need to support for default is likely to apply to
example just the same"* — which held exactly. `enum:` was the discovery, and the worst cell.

**Fix.** `handlers.RecoerceDeclValues`, called from `Build()` beside `RecheckSchemaShape` — the same
"the type is known now" seam, which exists for the sibling problem on shape gating. Sound because the
fallback preserved the raw text verbatim: a value still stored as a string was never typed.
String-typed schemas are skipped.

**Blast radius:** one test, and it was a latent defect of its own — `TestIssue2540` asserted a
model-level object literal as an escaped string while asserting the field-level one as real JSON, in
the same expectation. Now consistent.

**Also pinned:** the default RESPONSE vs a default VALUE (Fred's constraint). `KwDefault`'s context
set omits `CtxResponse`, so a `default:` keyword cannot appear in a response block; the default
response comes only from the code head in a route's `Responses:` body. The fixture holds both senses
side by side.

**Docs:** `swagger:enum` (annotation, members from a Go `const` block, typed from the declared Go
type) vs `enum:` (keyword, members written literally, typed from the schema it sits on) — a
side-by-side table in the enumerations tutorial, cross-linked from both reference pages.

**Follow-up landed in `23a9d1eb`** (Fred: *"what if RecoerceDeclValues encounters an error? Shouldn't
we emit a diagnostic warning?"*). Probing that question found the same silence in three shapes:

| site | uncoercible value | was | now |
|---|---|---|---|
| decl | `default: notanumber` on int | kept as `"notanumber"` — invalid document | dropped + warned |
| field | same | dropped **silently** | dropped + warned |
| both | `enum: 1, two, 3` on int | `[1,"two",3]` — invalid document | bad member dropped + warned by name |

Drop rather than keep: wrong-typed output fails validation, missing output is merely incomplete.
Enum drops per member and names it, because narrowing a closed set changes the author's contract.
Coercion TOWARDS string never fails (every literal has a string form), so string schemas are skipped
and cannot warn — which makes the `typ == "string"` early return a correctness statement, not an
optimisation. Parameters and headers already reported via their error sink; only the schema paths
were silent.

### Q37 — ✅ RESOLVED 2026-08-08 (`9376e2e1`, PR #94) — identity recognizers ran *after* the declaration lookup that could fail without them

**Status:** PARTLY FIXED 2026-08-02 (`7ff038a5`, see [Q39](#q39) tier 3) · **UNPARKED 2026-08-06** ·
**NARROWED 2026-08-08**. **Read the 2026-08-06 update at the end of this entry before acting on
anything above it**, then this correction to it:

> The 2026-08-06 update's items 1–3 are settled. The witness graph is not merely buildable — the
> declaration-contract fix landed in `master` (PR #90): the strict lookup sites ask
> `ScanCtx.SourcelessPackage()` and render from the type with a `scan.sourceless-type` Warning
> instead of failing, so the three failures in that update's table are **no longer failures**. Item
> 3's never-run test now exists and passes — `TestSourcelessType_DegradesInsteadOfFailing`, subtest
> *"a type the recognizers answer for warns about nothing"*, which is exactly the `time.Time`-in-a-
> syntax-less-package case.
>
> **What is left is the two residuals below, both re-verified against `master` on 2026-08-08:**
> `responses.go:327` still calls `IsStdTime` on a declaration fetched at `:315` (soft gate ⇒ a
> `swagger:response` on a named `time.Time` can lose its `date-time` **silently**), and the
> hand-rolled recognizer subsets are still **nine** call sites across parameters/responses. Both are
> now witnessable without `StubStdlib` and neither can fail a scan.
>
> 🛠 **The `responses` residual is being addressed in parallel (2026-08-08) — do not pick it up here.**
> It is not part of the `CompiledDependencies` derisking, which needs only that the failure mode be
> non-fatal, and that already landed.

#### ✅ Correction 2026-08-08 (branch `more-auto-detect`) — both residuals resolved, and the first was misdiagnosed

**`responses.go:327` was not a soft gate. It was unreachable.** `IsStdTime(decl.Obj())` asks whether
the response type IS `time.Time`, and `time.Time` is a **struct** underneath — so it always took the
`case *types.Struct` arm and could never arrive at the `default` arm where the check sat. Coverage
over the whole suite confirms it: that block's max hit count is **0**, while its neighbours (`:315`,
`:333`, `:340`, `:345`) are all exercised. Nothing was silently losing `date-time` through it, because
nothing ever entered it; `type Stamp time.Time` was — and still is — caught by the written-RHS
redirect above.

**The real hole was one level up, and wider.** `responses.buildNamedType` never consulted the
canonical recognizers at all: it refused `any`/`error`, then dispatched on the underlying shape. So
every recognized type that is a **struct underneath** was read as a response struct whose fields
become headers. Witness: `type Stamp = time.Time` under `swagger:response` emitted
`{"description": ""}` — no schema whatsoever — while the defined-type spelling rendered
`{string, date-time}`. Same for `= big.Int` / `= big.Rat`. `io.Reader` and `json.RawMessage` escaped
only because their underlyings are an interface and a slice, so they reached the delegating arm.

**Fix:** hoist `ApplyStdlibSpecials` to the top of `responses.buildNamedType`, after the `any`/`error`
refusal and before the written-RHS redirect and the shape dispatch — the schema builder's order, for
the schema builder's reason. The dead `IsStdTime` branch goes with it. Witnessed by
`TestResponseSpecials` over `fixtures/enhancements/response-specials/`, A/B'd against the unfixed
builder.

**The "nine hand-rolled subsets" count is retired: eight are legitimate guards, the ninth was the
dead line.** They are refusals, not recognizers, and they must keep running *ahead* of the canonical
set rather than being folded into it:

| site | check | why it is not a recognizer |
|---|---|---|
| `parameters.buildNamedType` / `buildAlias` | `IsAny \|\| IsStdError` | hard error — a parameter *set* cannot be `any`/`error` |
| `responses.buildNamedType` / `buildAlias` | `IsAny \|\| IsStdError` | hard error — same, for a response |
| `parameters.buildNamedField` / `buildFieldAlias` | `IsStdErrorType` | skip the field with a diagnostic; the schema builder's `{type: string}` would be a lie about what a client sends |
| `parameters.buildFieldAlias` / `responses.buildFieldAlias` | `IsAny` | empty schema — same effect as `recognizeAny`, harmless pre-delegation short-circuit |

`buildNamedField` in both builders already calls `ApplyStdlibSpecials`. `parameters.buildNamedType`
needs no canonical set: a `swagger:parameters` declaration must be a struct, so its `default` arm is
correctly an error.

**Q37 can close** once this lands.

`resolvers.IsStdTime` answers from `(package name, type name)` alone — it never reads the
declaration:

```go
func IsStdTime(o *types.TypeName) bool { return o.Pkg().Name() == "time" && o.Name() == "Time" }
```

Both `buildNamedField` implementations nonetheless demand the declaration first, and abort if it is
absent:

```go
// parameters/parameters.go:395              // responses/responses.go:345
decl, found := p.Ctx.DeclForType(o.Type())   decl, found := r.Ctx.DeclForType(ftpe.Obj().Type())
if !found { return fmt.Errorf("unable to find package and source file for: %s", ...) }
if resolvers.IsStdTime(o) { ... }            d := decl.Obj(); if resolvers.IsStdTime(d) { ... }
```

So a `time.Time` parameter or response field is only recognised because `time` happens to be in the
scanned set. The recognizer's whole point is that it does not need to be.

**Why it is parked.** The defect is invisible with a full package graph, so nothing on this branch can
witness a fix. `Options.StubStdlib` — which deliberately omits a package from the graph — is what
makes it reachable, and it lives on `wasi-build`. Landing the fix here would mean shipping a change
no test exercises, on reasoning alone; that is the shape of mistake
[[feedback_witness_before_claiming_a_fix]] records. **Land it with `StubStdlib`, or after it.**

---

#### Audit findings 2026-08-02 — for whoever picks this up

**Blast radius: zero, measured.** Hoisting the recognizer above the lookup in both sites: full
workspace suite green, **no golden moved**. Expected for a latent quirk — with a full graph the
lookup always succeeds, so ordering is unobservable.

**`decl.Obj()` → `ftpe.Obj()` is safe.** `responses.buildNamedField` reads the object off the decl,
which is what forces the lookup. `DeclForType` on a `*types.Named` resolves via
`FindDecl(pkgPath, name)`, and a package cannot declare a name twice — so the two are the same
`*types.TypeName`. The responses site can use `ftpe.Obj()` and hoist freely.

**The `resolveRefOrErr` sibling is a FALSE ALARM — do not "fix" it.** The earlier note flagged
`schema/ref.go` as having the same shape (`GetModel` → `missingSource`, no specials pre-check). It
has five callers, all inside `buildNamedType` / `buildNamedArrayLike`, both downstream of
`applyStdlibSpecials` at `schema.go:480`. The specials have already run. **The schema builder is
consistent throughout**; the gap is confined to parameters and responses.

**There is a THIRD site.** `responses.buildNamedType`'s default arm (`responses.go:294`) has the same
shape — `DeclForType` then `IsStdTime(decl.Obj())`. There the lookup is a soft gate (`if found`), so
it degrades rather than erroring, but the recognizer still cannot fire without the decl. Fix all
three, not the two that surfaced under truncation.

**The wider asymmetry: each site runs a DIFFERENT subset of the recognizers.** Schema applies one
canonical set (`applyStdlibSpecials` = any / time.Time / error / json.RawMessage / uuid) uniformly.
Parameters and responses hand-roll per function:

| site | IsAny | IsStdError | IsStdTime |
|---|---|---|---|
| `parameters.buildNamedType` | ✅ | ✅ | — |
| `parameters.buildAlias` | ✅ | ✅ | — |
| `parameters.buildNamedField` | ✅ | ✅ | ✅ *(after lookup)* |
| `parameters.buildFieldAlias` | ✅ | ✅ | — |
| `responses.buildNamedType` | ✅ | ✅ | ✅ *(after lookup, soft gate)* |
| `responses.buildAlias` | ✅ | ✅ | — |
| `responses.buildNamedField` | — | — | ✅ *(after lookup)* |
| `responses.buildFieldAlias` | ✅ | — | — |

**This costs nothing with a full graph** — probed 2026-08-02: `time.Time` and `json.RawMessage` are
correct in all four positions (query parameter, body schema, response header, response body), because
these builders **delegate to the schema sub-builder**, which supplies the canonical set. The
hand-rolled checks are only pre-delegation short-circuits. Under truncation the delegation is exactly
what stops working, which is why the subsets matter there and nowhere else.

So the real fix is larger than a reordering: **give parameters and responses the canonical set**
rather than eight hand-rolled subsets — the same "one shared rule, not N copies" move that
`common.Builder.ClassifierAliasStrfmt` made for [Q32](#q32). Sequence it with `StubStdlib` so each
step has a witness.

#### Update 2026-08-03 — partially addressed, for one more class of types

[Q38](#q38)'s fix adds `recognizeOpaqueStream` to the canonical `ApplyStdlibSpecials` set, so the
stream types listed there now answer from the object alone at every site, ahead of any declaration
lookup. For those types specifically, the degraded-graph failure this entry describes is gone: a
package graph without `io` no longer produces
`unable to find package and source file for: io.Reader` — it degrades to `{string, byte}` instead.

That is a widening of the canonical set, **not** the structural fix. The eight hand-rolled subsets
in parameters and responses are still there, and the `time.Time` / `json.RawMessage` half still has
no witness without `StubStdlib`. Q37 stays parked; it is simply parked over a smaller surface.

#### Update 2026-08-06 — unparked, and the new evidence is NOT what it first looked like

Written on `on-demand-scanner`, against `feat/source-loader`. Three things changed; only the first
is about Q37.

**1. The parking reason is gone.** This entry is parked because the defect is invisible with a full
package graph, and the truncated graph that exposes it required `Options.StubStdlib` from another
branch. The loader now produces exactly that graph — **types complete, syntax absent** — in two
*shipping* configurations, `Options.CompiledDependencies` and `Options.ExportData`. No stubbing, no
synthetic degradation. `internal/integration/loader_agreement_test.go` runs the corpus under both.
**The witness this entry has waited for is buildable today.**

**2. But the three failures that configuration surfaces are NOT this quirk.** They look like it and
they are not, and the difference decides who fixes them:

| target | type | why it fails |
|---|---|---|
| `bugs/2248` | `time.Duration` | only `time.Time` is recognised — `Duration` has no recognizer |
| `enhancements/opaque-streams` | `io.Writer` | **deliberately** excluded from `opaqueStreamTypes` |
| `goparsing/go123` | `reflect.Type` | no recognizer |

In all three there is **nothing to hoist**: ordering is irrelevant when no recognizer exists. With a
full graph they fall through to a structural walk that reads source; with export data there is no
source and the walk has no fallback, so the scan **fails** rather than degrading. That is the
declaration contract — which this entry's own audit predicted when it said "the real fix is larger
than a reordering". Note `io.Writer`'s exclusion is a considered decision, not an oversight: "a sink
the caller writes into is not something that travels on the wire."

**3. Q37's own case is now testable, and is expected to PASS.** At `parameters.go:359` and
`responses.go:399` the canonical `ApplyStdlibSpecials` now runs *before* `DeclForType` — the
2026-08-02 fix, verified in place 2026-08-06. So a `time.Time` field in a syntax-less package should
recognise without ever asking for a declaration. **That test has never been run.** Running it either
confirms the fix and closes this half, or finds a residual site — and it costs a fixture now, not a
branch.

**What genuinely remains open**, re-verified 2026-08-06:

- **The third site still recognises after the lookup.** `responses.go:325` is still
  `d := decl.Obj(); if resolvers.IsStdTime(d)`. Soft gate, so it degrades rather than erroring — but
  under export data a `swagger:response` on a named `time.Time` loses its `date-time` **silently**.
  That is the worst failure shape available and it now has a reachable path.
- **The hand-rolled subsets are still eight-ish.** `IsAny` / `IsStdError` / `IsStdTime` /
  `IsStdErrorType` appear at nine call sites across parameters and responses, while schema uses the
  canonical set throughout. These builders normally delegate to the schema sub-builder, which
  supplies the canonical set — and under a truncated graph **the delegation is exactly what stops
  working**. So "give parameters and responses the canonical set" is now testable too.

**The WASI side cannot answer this yet** (2026-08-06). A vendored dockerctl scans in 3.9 s there, and
near-instantly on rescan — but that corpus cannot reach family 1 at all: `time.Duration` appears in
106 files under `client/` and **none** under `models/`, and its generated models are JSON-shaped
scalars carrying no stdlib named types. So the green run is not evidence either way. The deciding
question — whether stdlib arrives as source or as export data, which is what decides whether
types-without-syntax ever occurs there — is waiting on better diagnostics UX on `wasi-build` before
it can be checked. Do not read the playground's success as clearing this entry.

**A hazard the fix must not create.** `on-demand-scanner` moved three builder sites onto
`EntityDecl.WrittenRHS`, and `schema.go:231` falls back to `tpe.Underlying()` when it returns false.
Today that branch is unreachable — every `EntityDecl` is built from an AST walk, so the scan fails
loudly first. The moment the declaration contract lets `FindDecl` return a declaration whose syntax
half is absent, these loud failures become a **silent peel**: `type Stamp time.Time` in a syntax-less
package renders as a struct instead of `format: date-time`, because `Underlying()` discards exactly
the named layer the recognizer keys on. Hoisting recognizers above the lookup is part of what
prevents that — a recognised type never asks for a declaration at all.

### Q38 — ✅ FIXED (`1d5ce79`) — stdlib IO interfaces had no recognizer, so they were drilled and leaked into the spec

**Status:** FIXED 2026-08-03. See **Resolution** at the end of this entry. Everything above it is the
probe that motivated the fix and describes the OLD behaviour.

**Original status:** STILL PRESENT on master · probed 2026-08-02 on `wasi-build` (full graph — this one is
**not** latent). Surfaced while auditing what a truncated graph cannot resolve.

`resolvers` recognises `time.Time`, `error`, `json.RawMessage`, `any` and (go1.27) `uuid` by
identity. There is no recognizer for the stdlib **IO interfaces**, so `io.Reader` and friends fall
through to ordinary structural drilling — and an interface is exactly the shape drilling handles
worst. Three positions, three different wrong answers.

**(a) As a parameter — an untyped parameter.** `fixtures/enhancements/in-case-insensitive/api.go:64`,
`Upload io.Reader` under `in: FORMDATA`:

```json
{"x-go-name": "Upload", "name": "upload", "in": "formData"}
```

No `type` at all. SimpleSchema requires one on every non-body parameter, and this is the canonical
file-upload shape, so `type: file` is what the author wanted.

**(b) As a model field — stdlib interfaces become definitions.** Probe module, `Body io.Reader` and
`Stream io.ReadCloser`:

```json
"body":   {"$ref": "#/definitions/Reader"},
"stream": {"$ref": "#/definitions/ReadCloser"},

"Reader":     {"type": "object", "title": "Reader is the interface that wraps the basic Read method.",
               "description": "Read reads up to len(p) bytes into p. …", "x-go-package": "io"},
"ReadCloser": {"allOf": [{"type": "object"},
                {"type": "object", "properties": {
                   "close": {"type": "string", "x-go-name": "Close", "x-go-type": "error"}}}]}
```

Two stdlib types are published as definitions carrying io's godoc as title/description, and
`ReadCloser` invents a **`close` property of `type: string`** out of the `Close() error` method
(interface-method promotion applied to a method that is not an accessor).

**(c) Under a graph that omits `io`** — a hard scan failure rather than a degradation:
`unable to find package and source file for: io.Reader`. This is the reachable half of
[Q37](#q37): unlike `time.Time`, there is no identity recognizer to fall back on, so the ordering fix
alone cannot save it. (Q37 is parked pending `StubStdlib`; this half of Q38 is parked with it, while
the full-graph halves (a) and (b) are not.)

**Fix shape.** An identity recognizer for the IO interfaces, ahead of the drilling path — the same
place `applyStdlibSpecials` already sits for `time.Time` and `json.RawMessage`. Open questions for
whoever takes it: which set (`io.Reader`, `io.ReadCloser`, `io.Writer`, `io.ReadWriteCloser`?), and
what each position maps to — `type: file` is only legal for a formData parameter in OAS2, so a body
or a model field needs a different answer (`string`/`binary` is the usual choice).

**Relation to `swagger:file`.** That annotation is the author's explicit escape hatch and already
works; a recognizer would make the canonical case right by default instead of silently wrong.

Positions verified: formData parameter, model field. Response body **not** probed.

---

#### Resolution 2026-08-03 (Fred) — an "opaque stream" class of special types

Framed as **a new class of special type**, not as an extension of the io-interface problem: the unit
is a **named type** recognized by identity, so the table can hold a struct (`io.LimitedReader`) and a
third-party type (`runtime.NamedReadCloser`) alongside the interfaces. Explicitly **not** structural
— "anything with a `Read` method" would swallow any user interface that happens to expose one, which
is the over-reach that makes inference dangerous here.

**Scope (Fred's list):** `io.Reader`, `io.ReadCloser`, `io.ReadSeeker`, `io.ReadSeekCloser`,
`io.ReadWriter`, `io.ReaderAt`, `io.ReaderFrom`, `io.LimitedReader`, `io.ByteReader`,
`io.ByteScanner`, plus `mime/multipart.File` and `github.com/go-openapi/runtime.NamedReadCloser`.

**`io.Writer` is deliberately excluded.** An API type (param, schema or response) is not expected to
contain a sink the caller writes into; if one does, that is unknown territory for the scanner and is
left to the author to override with whatever intent they had.

**The two answers:**

- **(a) formData parameter → `type: file`.** Not an inference at all — it repairs output that was
  *invalid*: the parameter carried no `type`, and SimpleSchema requires one.
- **(b) everywhere else → `{type: string, format: byte}`.** `byte` rather than `binary`: a JSON body
  cannot carry raw octets, and `byte` is the base64-encoded string OAS 2.0 defines for exactly that.

**On Fred's "two-edged sword" reservation** — *you cannot guess what the Reader is sending over the
wire* — recorded because it is right and the fix is shaped by it. The point is that the choice was
never *guess vs. don't guess*: codescan already guessed, and guessed far worse, publishing io's
interfaces as definitions with io's godoc and inventing a `close` property of type `string` from
`Close() error`. `{string, byte}` **states less**, and it is the standard way of saying "opaque
bytes, framing unstated". The author's `swagger:file` / `swagger:type` / `swagger:strfmt` still wins.

**Blast radius:** ONE pre-existing golden line — `enhancements_in_case_insensitive.json` gains
`"type": "file"` on the `upload` formData parameter, i.e. the invalid parameter becomes valid.

**Witness:** `fixtures/enhancements/opaque-streams/` + `TestOpaqueStreams`, covering model field,
formData parameter, body parameter, non-body parameter, response body and response header, plus the
override control and a `WriterModel` pinning `io.Writer`'s exclusion.

**Half (c) is now partly addressed, and [Q37](#q37) with it** — see the Q37 note.

### Q39 — ✅ CLOSED (`4779aa09`, `3407a141`, `0fd15d4d`, `3848dac9`, `7ff038a5`, `361e2f29`) — the three builders kept private copies of shared rules, and they drifted

**Status:** CLOSED 2026-08-02 · all three tiers done · recorded from three instances verified in
one session, closed after six. This is a **generator of quirks**, not a single defect: each instance below was found
and fixed separately, but the mechanism that produced them is untouched.

`schema`, `parameters` and `responses` each resolve Go types to spec constructs, and each
re-implements rules the others also need. Where the schema builder has one centralised mechanism, the
other two carry hand-written approximations of it. Nothing forces them to agree, and nothing detects
it when they stop.

**Three instances, all verified this session:**

| rule | schema | parameters / responses | how it was caught |
|---|---|---|---|
| honour a classifier before the alias dissolve | `buildAlias` | own `buildFieldAlias`, each with the same defect | only by the SimpleSchema fixture slice — the schema-level fix looked complete ([Q32](#q32), `1b0e7b2f`) |
| stdlib identity recognizers | one canonical `applyStdlibSpecials` set | **eight** hand-rolled subsets, differing per function | audit ([Q37](#q37), table there) |
| diagnose an uncoercible `default:`/`example:` | nothing — silent | already had an `errSink` | probing Fred's question ([Q36](#q36) follow-up, `23a9d1eb`) |

Note the third row runs the other way: schema was the one missing the behaviour. This is not "the
other two are behind"; it is three implementations with no shared contract.

**Root cause (Fred, 2026-08-02): the SimpleSchema / full-Schema split.** The builders did not
diverge carelessly. A non-body parameter and a response header are OAS-2 **SimpleSchema** locations
with a genuinely different legality surface from a full Schema: `type` is restricted to
{string, number, integer, boolean, array, file}, `file` is parameter-only and requires
`in: formData`, `object` is not representable at all, `$ref` is forbidden, and `collectionFormat`
exists only there (see [[reference_simpleschema_oas2]]). Because parameters and responses had to
enforce rules the schema builder does not, they grew their own classification paths — and once
separate, they drifted on the rules that ARE common.

That reframes the fix: the goal is not "make the three identical", it is **separate the shared rule
from the location-specific legality gate**, so the common part can be shared and the differences
become explicit rather than emergent.

**Why it deserves an entry of its own.** Every instance so far was found by accident of coverage —
a fixture slice that happened to exercise the right position, an audit done for another reason, a
reviewer's question. The cost is not the individual bugs but that **a fix verified on one builder
reads as complete**, which is precisely how [Q32](#q32) nearly shipped half-done.

**Precedent for the fix.** `common.Builder.ClassifierAliasStrfmt` (`1b0e7b2f`) put one such rule on
the shared builder because all three needed identical behaviour before their own dissolves. That is
the shape: the rule moves to `common`, the three call it.

---

#### Update 2026-08-02 — detector built (`4779aa09`), first factorization done (`3407a141`)

Fred's hypothesis held: *"now that all the expected behavior is well locked, Q39 should fall more
easily."* The conformance suite took about twenty minutes to build, because every shape in it already
had a witness saying what it should produce — so it carries NO expectations of its own. Each subject
is compared against the model field; the suite only asks whether the three builders still agree.

**It found a real divergence on its first run**, in a fix landed an hour earlier and believed
complete: `swagger:type` on an alias reached the non-body branch of both alias field handlers and not
the body one. `swagger:strfmt` on the same alias was fine, because its classifier sits above the
branch — two annotations fixed together, diverging one branch apart.

**Then it made the factorization safe.** `3407a141` replaced both hand-written `buildFieldAlias`
walks with one shared `schema.BuildFieldAlias`: **158 lines of duplicated control flow became 13.**
Most of each walk was re-implementing what it delegated to — the schema builder already applies the
classifiers, honours `TransparentAliases`, and dissolves. Handing it the ALIAS rather than the
right-hand side was the whole of what the callers needed. One genuine exception survives: a body
field naming a `swagger:model` alias keeps its `$ref` identity.

The suite caught two mistakes mid-refactor that would otherwise have shipped: a shared version that
dissolved above the strfmt classifier (losing formats under `TransparentAliases`), and then a real
divergence where `swagger:type` was honoured for a model field but dropped for a query parameter and
a response header under that mode.

**Two design choices in the suite are load-bearing** — a re-derivation would likely get both wrong:

- **Subjects are the field's type DIRECTLY, not nested in a struct.** A first attempt nested them,
  which delegates everything to the schema sub-builder — all green, no information. The divergences
  live in the short-circuits that fire when a parameter or response field is ITSELF a named or alias
  type (`buildNamedField`, `buildFieldAlias`).
- **SimpleSchema positions are excluded.** They have a genuinely different legality surface, which is
  the root cause above. Comparing them needs a declared projection rather than equality; mixing them
  in would bury real drift under expected difference. A second suite could cover them that way.

**Tiers 1 and 2 are done** (`0fd15d4d` coverage, `3848dac9` refactor). Nine hand-written copies of
"construct a schema sub-builder, build, drain its post-declaration queue" became two shared spellings
— `schema.Delegate` (caller's declaration context) and `schema.DelegateAs` (a resolved one, with
`InferNames`). `buildFromFieldStruct` and `buildFromFieldInterface` disappeared entirely: both were a
delegation with no shape-specific content. Tier 2 (the map arm) dissolved into the same primitive —
its differences were in the surrounding shape, not the sub-build. No golden and no provenance anchor
moved.

The drain is the part that mattered. A model reached solely through one field arrives in the spec via
that queue and nowhere else, and the loop had already gone missing once from the parameters map arm
during an earlier factor-out — a fixture comment in the code records it.

**Coverage had to be extended first, and both extensions earned it** — the conformance suite's
subjects were all named/alias types, reaching only two arms of the dispatch, and
`responses.descendBody("items")` affects cross-ref pointers ONLY, so neither goldens nor a
schema-comparing suite could see it. Verified by removing it: the anchor collapses from
`/responses/{n}/schema/items/properties/code` to `…/schema/properties/code`. The witness needs an
INLINE element; a named one anchors in its own definition, where the threading is unobservable.

**Deliberately not shared: the field dispatch switch.** Sharing it needs callbacks for the named arm,
the alias arm and the items descend — three hooks to remove two lines of difference.

**What remains: the recognizer subsets (tier 3).** Still awkward for the reason the [Q37](#q37) audit found —
with a full package graph they are invisible, because delegation supplies the canonical set anyway.
Safe to unify, unwitnessable here. That half waits on `StubStdlib` with Q37.

---

#### Update 2026-08-02 — tier 3 DONE (`7ff038a5`, `361e2f29`). ✅ Q39 CLOSED

**The block was wrong, and the way it was wrong is the lesson.** Tier 3 was parked on `StubStdlib`
because the subsets looked invisible under a full graph — true for `time.Time` and
`json.RawMessage`, which have declarations in any graph that scans them. It is **false for `error`**:
predeclared, `Pkg()` is nil, so no declaring source exists to find *in any graph at all*. The
truncation `StubStdlib` was going to simulate is permanent for one member of the canonical set, and
that member was sitting in the corpus the whole time.

So the generalisation ("recognizer subsets need a truncated graph to observe") was drawn from two of
the five recognizers and quietly assumed of the third. Cost: one parked quirk that did not need to be.

**Three defects, all found by one coverage extension.** The conformance suite reached the stdlib
types only through ALIASES (`StampAlias = time.Time`), which lands in the alias arm — leaving
`buildNamedField`, the arm that actually carried the subsets, untested in all three positions.
Adding the named spellings plus `error` and `any` surfaced:

| # | defect | severity |
|---|---|---|
| 1 | `responses.buildNamedField` had NO recognizer and went straight to `DeclForType`, which reads `Obj().Pkg().Path()` → **nil dereference** | whole scan down |
| 2 | the parameters refusal keyed on the object, so `error` was dropped and `type Wrapped = error` was emitted | same type, two spellings, two answers |
| 3 | `responseTypable.AddExtension` wrote onto the RESPONSE, never the body schema | `x-go-type` a sibling of `description` |

Defect 3 is the purest specimen of Q39's thesis: `paramTypable.AddExtension` has routed
body→schema all along. One rule, two private copies, one of them never updated.

**A fourth thing, deliberately narrow.** `ScanCtx.DeclForType` now returns `(nil, false)` for a
package-less object instead of dereferencing nil, with its own unit test — the crash site, guarded at
the site. `PkgForType` one function below has the identical shape and is **left alone**: no
production path is known to reach it with such an object, and a guard whose necessity nothing
demonstrates is the hunk [[feedback_witness_before_claiming_a_fix]] warns about. Noted, not fixed.

**Every hunk verified by reverting it** — 4 conformance failures for the hoist, 4 for the extension
routing, 5 for the spelling fix, a panicking unit test for the `DeclForType` guard. Provenance
captured before and after the hoist: identical, so short-circuiting ahead of the delegation drops no
cross-ref anchor (the schema builder already short-circuits before descending, so no field-level
anchor existed to lose).

**Left in place: the top-level rejection gates.** `buildNamedType` / `buildAlias` in both builders
still test `IsAny || IsStdError` to refuse outright. Those are legality gates on a *declaration*
("a `swagger:parameters` set cannot be `any`"), not recognizers resolving a type, and they are
symmetric across the four sites. Both are in fact unreachable as written — a `*types.Named` with a
nil package can only be `error`, which cannot carry an annotation — but they document intent and
removing them buys nothing.

**What `StubStdlib` still owes**: [Q37](#q37) proper — `time.Time` and `json.RawMessage` recognised
under a graph that omits their package. The ordering is now correct at all three field sites, so that
work is a witness, not a fix. [Q38](#q38)'s third half (a graph omitting `io`) stays parked with it.

**Options when this is picked up:**

1. **Opportunistic** — keep hoisting rules to `common.Builder` as instances surface. Cheapest, but
   leaves the generator running.
2. **Audit** — enumerate every rule implemented more than once across the three builders and unify
   deliberately. The recognizer-subset table in [Q37](#q37) is the first row of that inventory, and
   is blocked on `StubStdlib` for its witness.
3. **Cross-builder conformance suite** — the structural antidote. Take one Go shape and reach it as a
   model field, a body-parameter field, and a response-body field; assert the three emitted schemas
   agree. Divergence then fails a test instead of waiting for a fixture to happen to cover the right
   position. Legitimate differences (SimpleSchema forbids `$ref`; a header is not a schema) go in an
   explicit exception list with reasons — the same shape as the strfmt symmetry ledger, which is what
   made [Q32](#q32) tractable.

3 is the one that changes the trajectory; 1 and 2 are catch-up. They compose: the conformance suite
would tell the audit where to look.

### Q40 — ✅ FIXED (`2c350a56`) — the allOf member path honoured no classifier, so `swagger:type` there yielded an empty member

**Status:** FIXED 2026-08-03 (`2c350a56`). Probed 2026-08-02 while measuring Q32's `swagger:type`
half; deferred from that fix because it is a SHARED gap, not an alias asymmetry — it breaks the named half too, and
the named half worse.

`buildNamedAllOf` consults `classifierAliasTargetStrfmt` and `applyStdlibSpecials`, but never
`classifierNamedTypeOverride`. A `swagger:type` on an allOf member is therefore dropped, and when the
member's underlying is a basic type the arm falls through to its default warn-and-skip:

```go
// swagger:type string
type ScalarNamed int

// swagger:model AllOfScalarNamed
type AllOfScalarNamed struct {
	// swagger:allOf
	ScalarNamed
	Note string `json:"note"`
}

codescan → allOf[ {} , {properties:{note}} ]
         + warning: buildNamedAllOf: unsupported Go type *types.Basic (int); skipping
```

An **empty allOf member** — schema-valid but meaningless, and it silently widens the type.

**Updated 2026-08-02 after `5ad1df18`:** the alias half is now CORRECT (`allOf[{string} + …]`), because
`buildAllOf`'s alias arm routes through `buildAlias`, which gained the classifier. So this is now
purely a NAMED-side defect, and the named/alias pair is asymmetric in the unusual direction — the
alias is right and the named declaration is wrong.

**Family.** Same shape as [Q34](#q34) — a dispatch arm that never received the classifier its
siblings run — and a cousin of [Q33](#q33), which is the embed arm of the same story. Three arms of
the composition dispatch (`buildNamedEmbedded` struct arm, `buildNamedAllOf`, `processEmbeddedType`)
each consult a different subset of the classifiers; see [Q39](#q39).

**Witness** exists already: `fixtures/enhancements/type-override-symmetry` carries the
`AllOfScalarNamed` / `AllOfScalarAlias` pair, currently recorded as a known difference.

---

#### Update 2026-08-02 — the allOf position joined the conformance matrix (`728b8f67`)

It is **three classifiers missing, not one.** `TestBuilderConformance` now compares an allOf member
alongside the model field / body parameter / response body, and pinned three cells against the one on
record:

| subject | model field | allOf member | |
|---|---|---|---|
| `swagger:type` on a named int | `string/` | `<empty>` | the recorded defect |
| `swagger:enum` on a named uint64 | `integer/uint64` | `<empty>` | **new** — same shape: no classifier, basic underlying, warn-and-skip default |
| `swagger:strfmt email` on `[]string` | `array<string/email>` | `string/email` | **new** — and this one is WRONG rather than empty |

The third is the worst of them: the composed schema asserts the value **is** an email address where
the Go type is a list of them. The arm's `classifierAliasTargetStrfmt` predates the element-driven
rule ([Q34](#q34)) and writes the format on the whole member.

#### Resolution 2026-08-03 (`2c350a56`)

One root cause for all three: `classifierAliasTargetStrfmt` is **shape-blind**, writing
`Typed("string", format)` whatever the underlying is. The shape-AWARE classifiers already existed,
scattered through `buildNamedType`'s switch; they are now `applyNamedShapeClassifier`, which both
arms call, so neither can answer one of these differently again. Four goldens moved, all intended;
no other spec in the corpus shifted, which is what says the extraction was behaviour-preserving for
the field path.

**Wholesale delegation was tried and reverted.** Routing `buildNamedAllOf` straight into
`buildNamedType` looked like the obvious unification and is wrong: the field arm publishes a
definition for a non-model struct, so composed members turned from inline into `$ref` and put types
in the spec their author never marked as models — `default-allof-embeds` documents
inline-for-non-model as deliberate, in the fixture's own doc comment. **The classifiers are the
shared part; the `$ref`-or-inline policy is not.** Measuring before committing to the refactor is
the only reason that was caught.

So the arm answers correctly only for the classifiers whose implementation happens to sit BELOW it
(the alias arm it delegates to, the stdlib specials it calls directly), and gets every classifier
living in the field dispatch wrong. The fix is not three patches: it is routing this arm through the
cascade the field dispatch already shares.

#### Update 2026-08-02 — what an annotation on the EMBED FIELD does (probed, Fred's question)

**Ignored — not polluting.** With `swagger:strfmt` or `swagger:type` written beside `swagger:allOf`
in the embedded field's own comment, the output is byte-identical to the same embed with nothing
extra. No leak to the parent schema, none to a sibling property, none into the embedded type's own
definition. The Q40 emptiness is not rescued by a field-level `swagger:type` either.

**But the comment IS read.** An UNKNOWN annotation in that position (`swagger:bogusthing`) fails the
whole scan — `classifier: unknown swagger annotation`. So the scanner parses and validates the
annotation name, and the builder then drops the known ones without a word. The author gets
validation feedback implying the annotation is meaningful, and nothing saying it did nothing.

**The asymmetry is exact.** Same annotation, same syntactic position — a field's doc comment —
honoured on a regular field, dropped on an embedded one:

```text
// swagger:strfmt uuid          on `Fmt Scalar`   → {string, format: uuid}   ✅
// swagger:type string          on `Typ Scalar`   → {string}                 ✅
// swagger:allOf + either       on an embed       → the annotation vanishes  ❌
// either alone (plain embed)   on an embed       → the annotation vanishes  ❌ ([Q33](#q33))
```

Precisely: the allOf arm honours the embedded TYPE's `swagger:strfmt` (and nothing else of the
type's), and never the FIELD's. The plain-embed arm honours neither. Whichever way [Q33](#q33) and
this entry are decided, the field-comment case needs a pinned answer too — honour it, or diagnose
it — because today it is accepted, validated and discarded.

### Q41 — ✅ FIXED (`9de2313f`) — `in:` beside a classifier annotation warned about itself

**Status:** FIXED on `fix/strfmt-dispatch-symmetry`, awaiting review. Found 2026-08-02 while probing
`swagger:type file`; Fred: *"that is a bug"*.

`parseClassifierBlock` treats a classifier body as prose-only and warned on every keyword token in
it. But `in:` and `name:` are **field directives**, not schema-body keywords — they say where a field
goes and what it is called, and the parameters / responses builders read them straight from the doc
text, not from the block. A parameter field may legitimately carry both a classifier and `in:`;
indeed `in:` is mandatory there. So the canonical file-upload idiom warned about its own required
directive:

```go
// in: formData
//
// swagger:file
MyFormFile *bytes.Buffer

→ warning: keyword "in" not valid under swagger:file [parse.context-invalid]
```

Fired for `swagger:strfmt`, `swagger:type` and `swagger:file` alike. The same field *without* the
annotation never warned, so the check was inconsistent as well as wrong — it was not doing location
checking at all, merely "classifier bodies are prose-only".

**Every keyword context-invalid diagnostic in the whole fixture corpus was an instance of this.**
After the fix there are zero, which is why `noparams.go` gains a `maximum:` under the same
annotation — a schema keyword there is genuinely invalid, and the TUI's column-translation test
needed a subject that lands exactly on a keyword. Its previous subject was the spurious `in:`.

**No golden moved:** the keywords were dropped either way; only the noise went.

**Left alone, noted here:** the `collectionFormat` value diagnostic points at the separator space
before the value rather than at the value itself (`" pipe"` instead of `"pipe"`). A small
position off-by-one on VALUE diagnostics — keyword positions are exact. Not investigated.

### Q42 — ✅ FIXED (`1a353da6`, `7b3f3c14`) — a `swagger:response` on a named non-struct type emitted no schema

**Status:** probed 2026-08-02 while removing the sibling short-circuits under [Q39](#q39) tier 3.

`responses.buildNamedType`'s default arm — the one reached when a `swagger:response` is declared on a
named type whose underlying is not a struct — short-circuits on `IsStdTime` and on the local
`strfmtFromDoc` helper. Both write into a LOCAL schema variable and `return nil` without the
`resp.WithSchema(&sch)` that only the delegate path performs, so the schema is built and then
dropped on the floor:

```go
// swagger:response stampResp
type Stamp time.Time

codescan → {"description": "..."}     // no schema whatsoever
```

**Two defects, and the first one alone is not enough.** Removing the short-circuits attaches a schema
but a wrong one, because the same arm passes `tpe.Underlying()` to the sub-build rather than
`decl.ObjType()`. The named type is discarded, so the stdlib recognizer never sees `time.Time` and
the declaration's own classifiers never see its `swagger:strfmt`:

| | after removing the shortcuts only | wanted |
|---|---|---|
| `type Stamp time.Time` | still no schema | `{string, date-time}` |
| `type Emails []string` + `swagger:strfmt email` | `{array, items:{string}}` | `{array, items:{string, email}}` |

Its sibling `buildNamedField` two functions down passes `decl.ObjType()` and is correct, which is
what makes this an inconsistency rather than a considered choice.

**The five tests were RIGHT, and the second "defect" was not one.** Passing `decl.ObjType()` sends
the type through the `$ref` machinery and publishes the response type as a DEFINITION — and a
`swagger:response` declares a response, not a model. `TestCoverage_ResponseTopLevelExample` and
`TestCoverage_ResponseEdges` exist to prevent exactly that. So `tpe.Underlying()` is deliberate, and
the fix is the other shape: keep it, and make the short-circuits correct and attaching.

**Fixed:** the two branches now call `resp.WithSchema`, and the format branch goes through
`common.ApplyArrayLikeStrfmt` so the element-driven rule applies. Witness
`fixtures/enhancements/response-named-nonstruct` + `TestResponseNamedNonStruct` pairs each subject
with the same type reached as a model field, and asserts the response types never become definitions.

| subject | before | now |
|---|---|---|
| `type Emails []string` + `swagger:strfmt email` | no schema | `array<string/email>` ✅ |
| `type Code string` + `swagger:strfmt isbn` | no schema | `string/isbn` ✅ |
| `type Count int64` (control) | `integer/int64` | unchanged ✅ |
| `type Stamp time.Time` | no schema | `$ref` → `string/date-time` ✅ (`7b3f3c14`) |

**The `Stamp` cell was a second, distinct defect, fixed in `7b3f3c14`.** `type Stamp time.Time` is not
`time.Time`: the recognizer keys on identity and correctly declines. The model side resolves it
anyway because `buildFromDecl` builds from the declaration's WRITTEN right-hand side (`Spec.Type` =
`time.Time`, a named type), where the recognizer fires one level in. `responses.buildNamedType` used
`o.Type().Underlying()`, which peels every named layer at once and handed the arm time.Time's struct
— read as a response struct whose fields become headers, of which time.Time has none.

It now follows the written RHS, but **only when that RHS is itself a NAMED type**. A struct literal,
a slice or a basic type is already the shape this arm should build, and routing those through the
sub-builder would publish the response type as a definition. That distinction is what keeps
`type ScalarResp string` and `type ArrayResp []string` on their existing path.

The witness's pinned list is now empty: a response body and a model field are both full-schema
positions, so every difference found between them has been a defect.

### Q43 — ✅ FIXED (`6ce3bd4c`, `e6660371`) — a single-character tag or operationId silently voided the whole `swagger:route`

**Status:** FIXED 2026-08-02 · found while probing for [Q39](#q39) (a throwaway fixture used `e` as
its tag and lost both its routes).

```go
// swagger:route GET /aaa e opA     →  no path, no diagnostic, nothing
// swagger:route GET /aaa ee opA    →  /aaa
```

Two regexes each demand a letter followed by **one or more** further characters, so neither can match
a one-character name:

```
rxOpTags = "(\p{L}[\p{L}\p{N}\p{Pd}\.\p{Pc}\p{Zs}]+)"
rxOpID   = "((?:\p{L}[\p{L}\p{N}\p{Pd}\p{Pc}]+)+)"
```

The tags group is optional, so the parse does not fail there — it **backtracks and tries to match
with no tags at all**, which leaves `rxOpID` to swallow `e opA`. Its class has no `\p{Zs}`, so that
fails too and the line matches nothing. An unmatched `swagger:route` is not a malformed route; it is
not a route, so there is nothing to raise a diagnostic about. Same defect in `rxOperation`.

Neither restriction has any basis in OAS 2.0: a tag and an operationId are free-form strings, and
one character is a perfectly ordinary tag name.

**Both halves fixed.**

`6ce3bd4c` — `+` → `*` in both patterns; the leading `\p{L}` already carries the "starts with a
letter" rule. Witnessed by `fixtures/enhancements/route-name-shapes` (five accepted shapes: short
tag, short id, both, short id with no tags, short tag beside a long one) plus parser-level cases for
`swagger:operation`, which shares the patterns.

`e6660371` — the silence, which was the half that mattered. A failed path annotation now raises
`scan.unparsed-path-annotation` with its position and text. Note the ordering property: **the
diagnostic alone would have made Q43 loud** — reverting the regex fix while keeping it turns the
five short-name routes into five warnings instead of five silent disappearances. That is the
general guard; the regex fix is the specific one.

**The false-positive lesson.** The first attempt keyed the diagnostic on the KEYWORD alone and hit
three of this repo's own fixtures — doc comments whose prose opens a line with `swagger:route`,
which is inevitable in files that document annotations. Since annotations must start the comment
line, "starts with the keyword" cannot distinguish intent. The check now matches the annotation
**head** — keyword, method, `/`-rooted path — which a real annotation always has and prose after
the keyword never reaches. Corpus-wide re-scan after the tightening: zero hits outside the
deliberate one.

**Still open, related:** nothing here addresses a path annotation that parses but is wrong in some
other way. The head-matching approach generalises (it is a "looks like X but is not X" detector) if
another such class turns up.

### Q44 — ✅ FIXED (`13e2918`) — a basic-underlying embed was dropped instead of becoming a named property

**Status:** FIXED 2026-08-03. Resolution at the end of the entry; the body describes the OLD
behaviour. · Split out of [Q33](#q33) on 2026-08-03 (Fred), because unlike its parent it had one
faithful answer and no behaviour-change dilemma.

Embedding a named type whose underlying is a basic (or an array) promotes nothing — there are no
fields to promote — so Go emits it as an ordinary property **keyed by the type name**:

```go
type FmtBasic int          // swagger:strfmt duration
type Host struct {
    FmtBasic
    Label string `json:"label"`
}

json.Marshal → {"FmtBasic":0,"label":"l"}          // a property named FmtBasic
codescan     → {object, properties:{label}}        // the property is GONE
             + warning: buildNamedEmbedded: unsupported Go type *types.Basic (int); skipping
```

`buildNamedEmbedded` switches on the embedded type's underlying and has arms only for struct and
interface; everything else hits a `CodeUnsupportedGoType` warn-and-skip. So the property vanishes —
silently as far as the spec is concerned, since a warning about an "unsupported Go type" reads like a
type codescan cannot model rather than one it simply drops.

**The faithful answer is unambiguous**: emit a property named after the embedded type, built from
that type — which means the classifiers apply to it as they would to any named-type property, so the
`swagger:strfmt duration` above lands on it. Note this is the ONE case where the embed's own json tag
matters again (it names an ordinary property), unlike the promoted-marshaller case in
[Q33](#q33).

Same arm as [Q33](#q33) and reachable by the same fixtures, so they are worth doing in one pass even
though only Q33 needs a decision first. Related: the composition arm's version of this gap was
[Q40](#q40), fixed in `2c350a56`.

---

#### Resolution 2026-08-03 — `embedPromotes` routes it to the named-property path

The fix is not in `buildNamedEmbedded` at all. `buildPlainEmbed` now asks **whether the embed
promotes anything** (`embedPromotes`: peel pointer, unalias, is the underlying a struct or an
interface?) and, when it does not, gives it the **same path a json-named embed already took** — a
single named property built from the embedded type, keyed by the Go field name. It is the same
thing, so it is the same code: classifiers reach it, the embed's own json tag renames or drops it,
`x-go-name` behaves as on any field. `buildNamedEmbedded`'s `default` arm survives only as a
defensive guard; nothing on the struct path reaches it now.

**The target shape was not invented.** The fixture carries a `Control` model declaring every
embedded type as an ORDINARY field, so what "built from the embedded type" means is *calibrated
against the builder's existing behaviour* rather than asserted by the test: `Count` → `$ref`,
`FmtBasic` (`swagger:strfmt duration`) → inline `{string, duration}`, `Token` (TextMarshaler) →
`{string, x-go-type}`.

**Blast radius:** 4 goldens. `go123_aliased_spec.json` (`EmbeddedWithAlias` embeds `type UUID =
int64`, which now correctly emits a `UUID` property) and the three
`strfmt_symmetry_composition_*.json`, whose `EmbedBasic` cell moves from "member vanishes" to a
correct member carrying the format. That ledger cell became an `exceptions` entry: a promotes-nothing
embed is keyed by the embedded IDENTIFIER, and a named/alias pair is two different identifiers by
construction, so the halves legitimately differ in the property NAME while agreeing in shape.

**Adjacent case NOT fixed, and left alone deliberately:** an embed of an EMPTY interface
(`type Anything = any` in the same `go123/aliased` fixture) still contributes nothing, because the
interface arm returns early for `utpe.Empty()`. Go would emit `{"Anything": null}`. It is out of the
`default`-arm scope this entry describes, and genuinely niche — noted, not fixed.

**Also fixed alongside (`2886ae7`), a pre-existing false positive it exposed:** the
`scan.ineffective-annotation` warning fired for EVERY annotated embed, including a json-named one —
where the classifier is in fact honoured. So an author naming an embed was told their annotation had
been dropped while it was being applied. The warning now sits on the two arms that really discard it
(the allOf member and the promoting plain embed). Witnessed by `EffectiveOnNamedEmbed` in the
annotation-noise fixture.

---

### Q46 — ✅ FIXED (`6d5a9448`) — a description-only `$ref`'d field dropped its description silently

**Status:** FIXED 2026-08-05 on `fix/coverage-quirks`. Surfaced by the coverage sweep (`b06eae8e`),
which reached the `$ref`-sibling collector for the first time.

A field whose Go type is a model becomes a `$ref`, and draft-4 gives a `$ref` no siblings. When the
field's ONLY decoration is prose, the legacy default emits a bare `{$ref}` and the description is
gone:

```go
// DescOnly is a $ref'd field carrying only a description.
DescOnly Ref `json:"descOnly"`

codescan → {"descOnly": {"$ref": "#/definitions/Ref"}}      // description gone, no x-go-name
```

Add ANY other decoration (`readOnly: true`) and it survives on the compound. The drop itself is
**deliberate and documented** — `README.md#ref-override`, `DescWithRef` (deprecated, default false)
governs exactly this case, and the defaults reproduce the legacy behaviour byte-for-byte. It is not
a bug and was not changed.

**What was wrong was the silence.** The same README states the principle for `SkipAllOfCompounding`:
each drop raises one `CodeDroppedRefSibling` "so the loss is never silent". The description-only
default lost the same class of content and said nothing, and nothing pointed at `EmitRefSiblings`,
which keeps it.

#### Resolution (`6d5a9448`)

`CodeDroppedRefSibling` as a **Hint** (not a Warning — nothing is wrong, the default simply cannot
carry prose beside a `$ref`), naming the field and the option. The code's doc now describes both of
its causes, told apart by severity. **Zero golden drift across the corpus** — only the diagnostic
stream grew, which was the whole claim. Volume measured before settling on Hint: 0 on the petstore,
7 of 31 diagnostics on the classification corpus.

**Provenance note, recorded because the first report was wrong:** this was originally written up as
"a bare `x-foo:` line silently drops the whole doc comment". It does not. A bare `x-` line is not
the extension grammar (`Extensions:` + `---` is) — it is **prose**, and on a plain field it lands in
the description exactly as prose should. On a `$ref`'d field that prose was simply the only content,
which is this quirk. There is no extension-specific defect.

---

### Q47 — ✅ FIXED (`80eeed0d`) — `walkPathItemProse` guarded against an operation that cannot be nil

**Status:** FIXED 2026-08-05. Dead code, found by the same coverage sweep — the branch never
executed because it cannot.

`walkPathItemProse` looped `for _, op := range operationsByMethod(pi)` and skipped `op == nil`. But
`operationsByMethod` already filters: it walks the seven method slots and `continue`s on a nil one
before yielding. The guard was unreachable. Removed, with a comment saying why none is needed.

---

### Q48 — 🟦 DOCUMENTED (`80eeed0d`) — `walkSchemaProse` carries arms a Swagger 2.0 document cannot reach

**Status:** documented, deliberately kept. Not a defect.

`walkSchemaProse` recurses through `AnyOf`, `OneOf` and the tuple form of `Items` (`Items.Schemas`).
None can fire while the emitter targets Swagger 2.0, which has no such constructs — so they read as
untested paths in any coverage report and always will.

They stay: the walk has to remain total over `spec.Schema`, and they become live the day the emitter
targets OAS 3.x. A comment on the function now says so, so the next coverage reader does not chase
them.

### Q49 — ✅ FIXED (`8625ca15`, PR #119) — a `swagger:meta` outside the package doc read a different comment

**Status:** FIXED and merged 2026-08-17. Registered here after the fact — the fix landed alongside the
TUI escape-sequence work, and this register would otherwise never have heard about it.

Detection read every comment group in a file, so a `swagger:meta` was found wherever it sat. The block
was then taken from the file's **package doc** regardless of where it had been found, so what you got
depended only on whether the file happened to have one:

| file shape | before |
|---|---|
| has a package doc, `swagger:meta` elsewhere | an **unrelated sentence** was parsed as the meta block — an ordinary line about the package could set the spec's title and version, while the authored block was dropped |
| no package doc at all | nothing to parse; the nil comment group reached the origin recorder for the info node — **a nil dereference that aborted the whole scan** |

The group carrying the annotation is now the meta block, which is what `detectNodes` already implied and
what the classifier's own documentation already claimed. The documented placement — a block *in* the
package doc comment, which is what every fixture uses — is unaffected: there, the group carrying the
annotation is the package doc.

**Two things worth keeping from how this was found.** The crash arm was reachable **only for a caller
asking for provenance**, which is why `genspec-tui` hit it and the library never did — a reminder that
`OnProvenance` is a second execution path through the scanner, not a passive observer. And every fixture
used the documented placement, so the whole corpus agreed with a broken implementation: a golden suite
proves what the fixtures cover, and here they all covered the one case that worked.

## 2. Open — documented sharp edges (no fix planned, authors need to know)

### Q23-edge — a bare `---` in route prose absorbs everything after it

A `---` written as a markdown horizontal rule in a `swagger:route` description
opens a YAML fence: the rest of the prose is swallowed as a YAML body. Witnessed
by `enhancements_routes_description_yaml_fence_absorb.json`; the behaviour is the
intended one (the fence is how the operation body is introduced), the trap is
that authors do not expect it.

**Action:** doc-site only — a note in the routes tutorial. No code change wanted.

### Accepted behaviours, no action

Each is documented at the source and needs nothing from this file; listed so a
future sweep does not re-discover them as "bugs":

| behaviour | documented in |
|---|---|
| `interface{}` literals render as an empty schema | `schema/README.md#quirks-open` 🟦 |
| generic *declarations* emit nothing; only instantiations do | `schema/README.md#quirks-open` 🟦 |
| `FindModel` is a deprecated alias still on the API surface | `scanner/README.md#quirks-open` |
| `detectNodes` recognises annotation tokens it does not act on | `scanner/README.md#quirks-open` |
| `shouldAcceptTag` precedence when include+exclude are both set | `scanner/README.md#quirks-open` |
| `form` accepted as an alias for `formData` | `routebody/README.md#quirks-open` |
| `collectionFormat:` accepted laxly | `routebody/README.md#quirks-open` |

## 3. Tracked in a feature doc, not here

| item | feature doc |
|---|---|
| routebody does not track per-line **columns** — the precision blocker for LSP diagnostics | `features/column-precision-unicode.md` |
| a field-level `enum:` override **silently discards** the type's `x-go-enum-desc` per-value docs (no diagnostic) | `features/enum-richer-values.md` §1.2b |

## 4. Verified-stale records (do not trust the archived status lines)

Checked 2026-07-30 against current goldens; every one of these is recorded as
open in an archived file but is **fixed**:

| record | archived claim | verified reality |
|---|---|---|
| D1, D2, D5 | "Not attempted" | RESOLVED, PR #32 (`40978ced`, `a758dca8`) — per `archive/observed-quirks.md`'s own closing tally |
| D6 | "Not attempted" | REFRAMED, PR #32 (`5fab72f6`) |
| D4 | "Not attempted" | CLOSED-NO-ACTION 2026-06-10 |
| D3 / `schema/README.md#quirks-open` 🟡 "named-strfmt + `swagger:model`" | open, "reverted, deferred" | **fixed** — golden `enhancements_named_struct_tags-ref.json` shows `PhoneNumber` = `{type: string, format: phone}` and `Contact.phone` = `$ref`. Superseded by F-series **F1** (`8e20d2f`) |
| `schema/README.md#quirks-open` 🟡 "cross-package name collisions silently overwrite" | open, "needs three pieces" | **fixed** by name-identity — goldens show `AWidget`/`BWidget`, `XItem`/`YItem` deconfliction |
| F1–F9 | — | all ✅, incl. the F9 alias infinite loop |
| go-swagger backlog | — | 236/236 triaged, **0 open 🛠** |

**Repo-side follow-up: ✅ done** — merged as PR #68 (`a274219`, commits `9ec25d7`
+ `0926413`). The two stale 🟡 entries moved to
`internal/builders/schema/README.md#quirks-resolved`, the enum entry condensed and
handed to its feature doc, each package's `#quirks-open` now points here, the five
references to the archived `quirks-F-series-fix.md` were repointed, and a stale F9
"currently hangs the scanner" warning was removed from the alias witness (which
also got the alias how-to rewritten around real per-mode output).

## 5. Why the registers went stale

Three registers with no cross-links — the Q-series, the D-series, and the
per-package `README#quirks-open` sections. An item was logged in one and then
fixed by a *different* stream (F-series, name-identity, P5.1), and nothing
updated the original. Hence this file, and hence the rule at the top: one live
register, everything else is provenance.
