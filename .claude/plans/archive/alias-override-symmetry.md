> [!NOTE]
> Last revision: 2026-08-01 (retrenched to `swagger:strfmt` only)

# `swagger:strfmt` dispatch symmetry (Q32)

## Summary

`swagger:strfmt` is honoured on a **named** type declaration and silently dropped on an **alias** one. The register
(`quirks-open.md` §Q32) frames this as an alias bug; reading the code it is an **inconsistency in which dispatch sites
consult a declaration's comments**. The named arms do it everywhere, through five different classifiers; the alias arms do
it in exactly two places, both of them local patches added when a specific context was reported.

Scope is `swagger:strfmt` alone (decided 2026-08-01). The other decl-level annotations each have special cases of their
own and are not part of this story — see §Out of scope.

This plan defines the fixture matrix FIRST: a corpus that exhibits every asymmetric dispatch site, with goldens capturing
today's wrong output. The fix is then an A/B on a table, not a claim. No production change lands until the matrix is agreed
and the "before" goldens are committed.

## Context

### Where `swagger:strfmt` is consumed — seven sites

| # | site | reached for | reads whose comments |
|---|---|---|---|
| 1 | `classifierNamedBasic:333` | named, basic underlying | own decl / `DeclForType` |
| 2 | `classifierNamedStructStrfmt:549` | named, struct underlying | own decl / `DeclForType` |
| 3 | `classifierNamedArrayLike:393` | named, array/slice underlying (`byte`, `bsonobjectid` specials) | own decl / `DeclForType` |
| 4 | `classifierNamedTypeOverride:99` | any, as a rider on `swagger:type` | own decl / `DeclForType` |
| 5 | `inheritedStrfmt:291` | named, chain walk to a redefinition's right | each decl along the chain |
| 6 | `classifierAliasTargetStrfmt:427` | **alias**, `buildNamedAllOf` struct branch only | the alias TARGET's decl |
| 7 | `buildDeclAlias:248` | **alias**, decl site only | own decl |

Sites 1–5 are the named machinery. Sites 6 and 7 are the two local patches: each solves "read the alias decl's comments"
for one call site. A third context reported tomorrow would grow an eighth site. That is the thing to fix.

### The four dispatch sites that can receive an alias

`buildAlias` has exactly four callers. Each has a named counterpart that runs a classifier; `buildAlias` runs none —
it fetches the decl only to test `HasModelAnnotation()`, then dissolves to `tpe.Rhs()` (`schema.go:399-430`). The
`TransparentAliases` branch returns at `:411` before the lookup even happens.

| caller | reached from | named counterpart | classifier on the named side | alias side |
|---|---|---|---|---|
| `schema.go:351` (`buildFromType`) | field, pointer, slice elem, array elem, map value, body param, simple param, response header, response body | `buildNamedType:438` | 1, 2, 3, 4 | **none** |
| `allof.go:185` | `swagger:allOf` walk, alias member | `buildNamedAllOf:201` | 6 (`classifierAliasTargetStrfmt`) | **none** |
| `embedded.go:155` (`processEmbeddedType`) | interface `swagger:allOf` member | `processEmbeddedType` Named branch | stdlib specials only | **none** |
| `special_types.go:28` | stdlib recognizer routing | — | — | n/a |

Plus the decl entry, which is the one place the alias side is mostly fine: `buildFromDecl:180` covers the `swagger:type`
rider and `buildDeclAlias:248` covers bare strfmt.

And one more asymmetry with no `buildAlias` involved: `buildEmbedded:58` unaliases and recurses, so an aliased **struct**
embed dispatches as its target — correct — but the alias's own `swagger:strfmt` is lost on the way, because
`buildNamedEmbedded` never consults comments at all.

## Trajectory

1. ✅ **Agree the matrix** (this document) — retrenched to `swagger:strfmt` 2026-08-01
2. ✅ **F1 core** fixture + goldens + ledger — the `buildFromType` row, which is most of the user-visible damage
3. ✅ **F2 composition** fixture + goldens — the two allOf/embed rows
4. ✅ **F3 simpleschema** fixture + goldens — the `simpleSchema` half of the `buildFromType` row
5. ✅ **F4 stdlib** fixture + goldens — recognizer-vs-override precedence (a real design question, not a missing call)
6. 🔍 **Then** design the fix against the ledger (out of scope here)

## The matrix

### Axes

| # | axis | values |
|---|---|---|
| A1 | declaration kind — **the symmetry pair** | `type T X` (named) · `type T = X` (alias) |
| A2 | RHS / underlying kind — selects which classifier the named side runs | basic · struct (own) · slice · array · named (chain) · stdlib struct (`time.Time`) |
| A3 | dispatch site | the four rows of the table above, plus decl |
| A4 | `swagger:model` on the decl | present · absent |
| A5 | alias mode (**test-side**) | default (Expand) · `RefAliases` · `TransparentAliases` |

Every cell is a **pair**: same RHS, same annotation, same use site, differing only in `=`. The named half is the control —
it defines what "correct" means for that cell, so the ledger needs no hand-written expectation and stays honest if we ever
change what named types do.

### A2 — why each RHS kind earns a row

Not arbitrary: each selects a different classifier on the named side, so each is a distinct thing the alias side must learn.

| RHS kind | named-side classifier | note |
|---|---|---|
| basic (`string`) | `classifierNamedBasic` | the canonical case |
| struct (own) | `classifierNamedStructStrfmt` | strfmt on a struct is a full type replacement |
| slice (`[]byte`) | `classifierNamedArrayLike` | `swagger:strfmt byte` is a whole-schema special, not an items one |
| array (`[16]byte`) | `classifierNamedArrayLike` | `bsonobjectid` special fires for arrays only |
| named (chain) | `inheritedStrfmt` | strfmt one declaration to the right |
| stdlib struct (`time.Time`) | recognizer vs. classifier | precedence question — F4 |

### Fixture packages

**F1 `strfmt-symmetry-core`** — the `buildFromType` row, non-simpleSchema.
A2 {basic, struct-own, slice, array, named-chain} × A3 {decl, field, pointer, slice elem, map value} × A1 pair, A4 absent.
A second envelope repeats {basic, struct-own} × {decl, field} with A4 **present**, so the `swagger:model` gate is visible
without doubling the package.

**F2 `strfmt-symmetry-composition`** — the two allOf/embed rows, split out because `buildEmbedded` /
`processEmbeddedType` / `buildNamedAllOf` are separate dispatch and would drown F1's goldens.
A3 {struct embed, interface `allOf` member, `swagger:allOf` struct branch} × A2 {basic, struct-own} × A1 pair.

**F3 `strfmt-symmetry-simpleschema`** — the `simpleSchema` half, where the `refModel` gate flips and `$ref` is illegal.
A3 {body param, query param, response header, response body} × A2 {basic, slice} × A1 pair.

**F4 `strfmt-symmetry-stdlib`** — precedence, kept apart so its goldens can move independently if we decide the recognizer
should win.
A2 {`time.Time`, `strfmt.UUID`} × A3 {decl, field} × A1 pair.

Four packages × 3 alias modes = 12 goldens.

### Naming convention

Greppable cell IDs so a golden diff names its own cell:

```text
Strfmt<RHSKind><DeclKind>   →  StrfmtBasicNamed / StrfmtBasicAlias
                               StrfmtSliceNamed / StrfmtSliceAlias
```

One `Envelope` per package, one field per cell, `json` name = lower-camel cell ID.

### The ledger test

Per cell, compare the alias half's emitted schema against the named half's and render a table. Cells we agree are
legitimately asymmetric go in an explicit skip-list with a reason, so the table is a contract rather than a TODO list.
Failures are the fix's worklist; after the fix the skip-list is what remains of Q32.

## Out of scope

| annotation | why not here |
|---|---|
| `swagger:type` | has its own decl-entry override path; different story |
| `swagger:enum` | unfixable on alias-to-basic — the type-checker erases `type Unsigned = uint64`, so `const Zero Unsigned = 0` is indistinguishable from any `uint64` constant. Deserves a diagnostic, tracked separately |
| `swagger:default` | `classifierNamedBasic:349` returns `handled=true` without writing to the target — possibly a bug of its own, see Appendix A1 |
| `swagger:additionalProperties`, `swagger:patternProperties` | decl-only by construction (`schema.go:104-105`, definitions path only); no use-site counterpart to be asymmetric with |
| `swagger:alias` | deprecated inert sink (F8) |
| generic instantiations | `buildNamedType:491` dissolves them before any classifier; orthogonal |

## Actions

1. ✅ Review + agree this matrix (Fred)
2. ✅ F1 `strfmt-symmetry-core` fixture + 3 goldens + ledger test — awaiting shape approval before F2–F4
3. ✅ F2 `strfmt-symmetry-composition` fixture + 3 goldens
4. ✅ F3 `strfmt-symmetry-simpleschema` fixture + 3 goldens
5. ✅ F4 `strfmt-symmetry-stdlib` fixture + 3 goldens
6. ✅ Decide the F4 precedence question — **author always wins** (Fred, 2026-08-01)
7. ✅ Plain-embed shared gap (F2) — **out of scope, stays with Q33** (Fred, 2026-08-01): same mechanism, but it changes
   output for named types too and collides with Q33's undecided question about what embedding a formatted type should
   emit. The F2 embed cells stay as unasserted witnesses so Q33 inherits them
8. 📚 Update `quirks-open.md` §Q32 to point at the ledger instead of restating examples
9. ✅ Design + land the fix against the ledger — `212e1f4`; all 69 cells symmetric, every `knownBroken` map empty

## Achievements

### F1 `strfmt-symmetry-core` ⭐⭐

Branch `fix/strfmt-dispatch-symmetry`. 13 cell pairs × 3 alias modes; 3 goldens; ledger test
`internal/integration/strfmt_symmetry_core_test.go`. Full suite green, no drift in any pre-existing golden.

**Result: 32 of 39 cells asymmetric.** The ledger (identical in `default` and `refaliases`):

```text
CELL                     NAMED (control)              ALIAS
fieldBasic               string/isbn                  string/                      BROKEN
fieldStruct              string/duration              object{left,right}           BROKEN
fieldSlice               string/byte                  array<integer/uint8>         BROKEN
fieldArray               string/bsonobjectid          array<integer/uint8>         BROKEN
fieldChain               string/ssn                   string/ssn                   OK
pointerBasic             string/isbn                  string/                      BROKEN
pointerStruct            string/duration              object{left,right}           BROKEN
sliceElemBasic           array<string/isbn>           array<string/>               BROKEN
sliceElemStruct          array<string/duration>       array<object{left,right}>    BROKEN
mapValueBasic            map<string/isbn>             map<string/>                 BROKEN
mapValueStruct           map<string/duration>         map<object{left,right}>      BROKEN
modeledBasic             string/isbn                  string/isbn                  OK
modeledStruct            string/duration              string/duration              OK
```

`transparentaliases` is the same except `modeledBasic` / `modeledStruct` also break.

**Three things the matrix taught us that the register did not say:**

1. **It is not a dropped keyword — it is a different type.** Only the basic cell degrades to `{string}` with the format
   missing. Struct emits `object{left,right}`, slice and array emit `array<integer/uint8>`. A consumer generating code off
   the alias half gets a structurally different type, not a weakly-typed one.
2. **`fieldChain` is symmetric, and that localises the bug precisely.** Dissolving `type StrfmtChainAlias = BaseFormatted`
   lands on a *named* annotated type, so the named machinery picks the format up on the way through. The defect is
   specifically "the alias's OWN comments are never read", not "aliases lose formats". Any fix must preserve this cell.
3. **`swagger:model` is an accidental workaround.** An annotated alias gets its own definition, where `buildDeclAlias:248`
   applies the format, and use sites `$ref` it — so `modeled*` passes in two modes out of three. Under
   `TransparentAliases` the early return at `schema.go:411` precedes the decl lookup, so even that escape hatch fails.
   This is worth documenting for users regardless of when the fix lands.

**Ledger design ⭐⭐** — comparing each alias half against its named half (rather than a hand-written expectation) means the
test states the contract, not the current output; `knownBroken` fails in BOTH directions, so it cannot rot into a stale
TODO list, and after the fix it should be empty.

### F2 `strfmt-symmetry-composition` ⭐⭐

4 cell pairs × 3 modes. **6 of 12 cells asymmetric** — and the other 6 produced the more interesting result.

```text
CELL                     NAMED (control)                    ALIAS
EmbedBasic               object{label}                      object{label}                      SHARED GAP
EmbedStruct              object{label,left,right}           object{label,left,right}           SHARED GAP
AllOfBasic               allOf[string/isbn+object{note}]    allOf[string/+object{note}]        BROKEN
AllOfStruct              allOf[string/duration+object{note}] allOf[object{left,right}+object{note}] BROKEN
```

- **allOf is the money row.** `buildNamedAllOf` runs `classifierAliasTargetStrfmt` at `allof.go:205`; the alias arm at
  `allof.go:185` calls `buildAlias` directly and bypasses it. This is the site one of the two existing local patches was
  written for — and it still only covers the named half.
- **Plain embed is a SHARED gap, not an asymmetry.** Both halves drop the format: `buildNamedEmbedded` switches on the
  member's underlying shape and never consults its comments, so a basic member vanishes entirely and a struct member
  promotes `left`/`right`. Same shape as Q33's TextMarshaler embed. **This is the finding that justified adding the
  control check** — a purely pairwise ledger would have printed a comfortable "OK" here. Left unasserted (`wantNamed`
  empty, a `note` instead) because what an embed of a formatted type *should* produce is an open design question, not
  something this matrix gets to decide.

**One dispatch site deliberately carries no cell.** `processEmbeddedType` (`embedded.go:155`), the fourth caller of
`buildAlias`, is the interface-side allOf walk. Go interfaces can only embed interfaces, so the only alias reachable there
is an alias-to-interface — and no classifier consumes a format on an interface underlying on either side. Documented in
the fixture's godoc as out of scope rather than silently untested.

### F3 `strfmt-symmetry-simpleschema` ⭐⭐

4 cell pairs × 3 modes, **all 12 asymmetric**, identically in every mode.

```text
queryBasic               string/isbn      string/
querySlice               string/byte      array<integer/uint8>
headerBasic              string/isbn      string/
headerSlice              string/byte      array<integer/uint8>
```

Confirms the `simpleSchema` gate changes nothing about the defect: non-body parameters and response headers reach the pair
through `buildFromType` like everything else, so the dissolve happens before any classifier runs. Body parameters and
response bodies are deliberately absent — both are full-schema locations already covered by F1's dispatch.

### F4 `strfmt-symmetry-stdlib` ⭐⭐⭐

2 cell pairs × 3 modes, all 6 asymmetric — **and qualitatively different from F1–F3**.

```text
fieldTime                string/date      string/date-time     ← recognizer overrules the author
fieldRaw                 string/byte      <empty>
```

Everywhere else the alias half loses the annotation and falls back to the bare Go type. Here it loses the annotation and
gets a **confidently wrong different answer**: `swagger:strfmt date` on `type StampAlias = time.Time` emits `date-time`.
The asymmetry has a real cause: `applyStdlibSpecials` is keyed on the declaration's own identity, so `StampNamed` (not
`time.Time`) never reaches a recognizer and its classifier wins — while the alias dissolves onto `time.Time` itself and
the recognizer fires first.

**DECIDED 2026-08-01 (Fred): the author always wins.** `swagger:strfmt` is the escape hatch for exactly the case the
library cannot infer — `time.Time` may go on the wire as `date`, or as some custom format we have no way to guess. A
recognizer is a default for un-annotated code, never an override of an explicit annotation.

So F4's 6 cells are ordinary bugs, not a design question, and they stay in `knownBroken`. The rule the fix must implement:

> Wherever an annotation and a recognizer both have an answer, the annotation wins — on the named side and the alias side
> alike.

**Where the recognizer actually wins — and the placement trap.** Not at `buildAlias:406`: that `applyStdlibSpecials`
call is keyed on the alias's own TypeName, and `StampAlias` is not `time.Time`, so it never fires. The overruling
recognizer is the downstream one at `buildNamedType:446`, reached only after the dissolve. Trace, default mode:

```text
buildFromType:351   field type is *types.Alias        → buildAlias
buildAlias:406      applyStdlibSpecials(StampAlias)   → does NOT fire
buildAlias:415      GetModel("StampAlias")            → decl (and its comments) first available here
buildAlias:426      no swagger:model                  → dissolve: buildFromType(time.Time)
buildNamedType:446  applyStdlibSpecials(time.Time)    → recognizeTime → {string, date-time}
```

So the fix is "read the alias decl's comments before the dissolve". The trap is **placement**: the natural spot is right
after `GetModel` at `:415`, but the `TransparentAliases` early return sits at `:411`, ABOVE it — that mode dissolves
without ever looking the declaration up. A check placed after the lookup fixes default and `RefAliases` and silently
leaves `TransparentAliases` broken, which is exactly the shape of F1's `transparentaliases/modeledBasic` and
`modeledStruct` — the only cells in the matrix that break in one mode alone.

> **Constraint.** The declaration lookup and the annotation check must BOTH move above the `TransparentAliases` early
> return at `schema.go:411`, so the alias's own comments are read before the dissolve in all three modes.

A downstream fix instead (teaching `buildNamedType` to remember the originating alias) would land after
`applyStdlibSpecials:446`, which returns before any classifier in that function, and would need that call reordered.
Upstream is the better route — it also subsumes the two local patches in Appendix A2.

`fieldRaw` is the sharper of the two cells. `fieldTime` at least yields *a* format (`date-time` — wrong, but plausible);
`fieldRaw` yields `{}`:

```json
"fieldRawNamed": { "type": "string", "format": "byte" },
"fieldRawAlias": { }
```

`recognizeRawMessage` emits the open "any JSON" schema, which is the right default for an *un-annotated*
`json.RawMessage` and is precisely wrong as an override of `swagger:strfmt byte` — the property ends up with no type at
all. This is the strongest case for the rule, not an exception to it.

### Totals

| slice | cells | asymmetric | goldens |
|---|---|---|---|
| F1 core | 39 | 32 | 3 |
| F2 composition | 12 | 6 (+2 shared-gap pairs) | 3 |
| F3 simpleschema | 12 | 12 | 3 |
| F4 stdlib | 6 | 6 | 3 |
| **total** | **69** | **56** | **12** |

Branch `fix/strfmt-dispatch-symmetry`. Full `go test work ./...` green, `golangci-lint run --new-from-rev master` clean,
no pre-existing golden moved.

### The fix ⭐⭐⭐ — `212e1f4`

All 69 cells symmetric; every `knownBroken` map is empty. **No pre-existing golden moved** — only the 12 matrix goldens
changed, which is the evidence that the change is confined to aliases carrying a format.

`common.Builder.ClassifierAliasStrfmt` is the missing entry, dispatching on the alias's underlying kind so the format
lands where the equivalent named declaration puts it (whole-schema for basic/struct, the `byte` / `bsonobjectid`
whole-schema specials vs items for slice/array — shared with `classifierNamedArrayLike` via `ApplyArrayLikeStrfmt` so the
two cannot drift). Applied at three sites, all of them use sites: `schema.buildAlias`, `parameters.buildFieldAlias`,
`responses.buildFieldAlias`. Declaration sites are untouched.

**A fourth site was tried and reverted before landing.** Routing `buildDeclAlias` through the same classifier was
described at the time as "incidentally fixing a latent bug"; it had no witness, and probing it showed the opposite —
`type X = []string` with a non-special format emitted `items.format` on the alias while the NAMED control emitted
nothing, i.e. it created the mirror-image asymmetry rather than removing one. Reverting was golden-clean. What it did
surface is a genuine named-side defect, now logged as **Q34**: the decl-entry switch has no slice/array arm at all.

**The matrix earned its keep here.** After fixing `schema.buildAlias`, F1/F2/F4 went green and **F3 did not move at all** —
its goldens did not even change. Parameters and responses have their **own** `buildFieldAlias` with the same defect and
the same lookup-below-the-dissolve shape. Without the SimpleSchema slice the fix would have shipped looking complete,
with query parameters and response headers still dropping the format.

Both ordering constraints held exactly as predicted: the lookup had to move above the `TransparentAliases` return in all
three builders, and the classifier had to run before the right-hand side is reached so the author beats the recognizer.

Bonus: `fieldStructAlias` used to emit a bare `$ref` that also swallowed the field's description; it now carries prose,
type and format.

**Still open under Q32** (out of this branch's scope): `swagger:type`, `swagger:enum` (needs a diagnostic, not a fix),
`swagger:default` (Appendix A1), and the plain-embed shared gap now tracked under Q33.

## Appendix

### A1 — `swagger:default` returns handled without writing — CONFIRMED, now tracked as Q35

Probed 2026-08-01. Worse than this appendix guessed: no form of the annotation ever emits a default,
and `swagger:default <value>` on a **named basic type declaration** leaves the definition and every
property referencing it **typeless**, with no diagnostic. The bare form the doc-site demonstrates is
rejected by the grammar. Full write-up and the decision to take: `quirks-open.md` §Q35.

Original note follows.

`classifierNamedBasic:349`:

```go
if _, ok := s.findAnnotationArg(cg, grammar.AnnDefaultName); ok {
    return true
}
```

`handled=true` tells the caller "target written, terminal", but nothing was written. Either the default is applied
elsewhere and this is a correct short-circuit, or `swagger:default` on a named basic decl silently produces an empty
schema. Out of scope for Q32; log separately if it is a bug.

### A2 — the two existing local solutions

`classifierAliasTargetStrfmt` (allOf struct branch) and the inline check at `buildDeclAlias:248` both already do what a
general fix would do, each for one call site. The fix should subsume both — otherwise the next reported context grows a
third.
