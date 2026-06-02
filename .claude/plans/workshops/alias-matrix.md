# W3 — Alias matrix audit

Date: 2026-06-03
Status: ⬜ draft, awaiting review
Companion: `alias-handling.md` (workshop questions), `alias-ledger.md`
(to be created — judgment ledger during the corpus walk).

Purpose: produce a code-derived map of the alias surface — what
branches the builders actually take, what combinations the fixture
corpus covers today, and where the gaps are. The map drives the
corpus walk (TUI-based judgment cycle) so we know in advance the
fixture inventory we'll need.

This document is **descriptive**, not normative. It tells us what
combinations exist. The corpus walk decides what they *should*
produce.

---

## 1. Axes (derived from the code paths)

The decision branches in the alias-handling code give us the axes
mechanically — no design judgment yet. Each axis below corresponds to
an `if` / `switch` in a builder.

### 1.1 Mode (Options flags)

`schema.go:155-170`, `parameters/parameters.go:112-145`,
`responses/responses.go:215-298`.

| Value | Effect |
|---|---|
| `Expand` (default — both flags false) | LHS gets a structural definition mirroring the underlying. Schema only — parameters/responses have no Expand branch on the alias-decl path. |
| `RefAliases=true` | LHS gets a `$ref` to RHS's target. Body parameters / body responses only; non-body always expands. |
| `TransparentAliases=true` | No LHS definition; build from RHS directly. Wins over RefAliases. |

3 cells. Every fixture should be runnable under all three (the cheap
test-harness change from earlier).

### 1.2 RHS kind (`tpe.Rhs().(type)`)

`schema.go:172-204` switch; the parameters and responses paths
re-derive similar cases.

| Value | Notes |
|---|---|
| `*types.Basic` | Primitive (int64, string, bool, …). Subject to `UnsupportedBuiltinType` filter. |
| `*types.Named` (regular) | The "happy path" — most existing fixtures. |
| `*types.Named` (stdlib-special) | `time.Time`, `error`, `json.RawMessage` — `applyStdlibSpecials` recognizes these. Q3. |
| `*types.Named` (generic instantiation) | `List[int]` etc. — dissolves to Underlying per `dissolve-named`. |
| `*types.Alias` (regular, chained) | Alias-of-alias. Q11. |
| `*types.Alias` (any / interface{}) | `recognizeAny` short-circuit. Q13. |
| `*types.Struct` (anonymous) | `type X = struct{…}` — emit inline. |
| `*types.Interface` (anonymous) | `type X = interface{…}` — emit object with methods. |
| `*types.Interface` (empty) | `type X = any` is the same; `interface{}` literal — degenerate. |
| `*types.Slice` / `*types.Array` | `type X = []T`. |
| `*types.Map` | `type X = map[K]V`. |
| `*types.Pointer` | `type X = *T`. |
| Unsupported builtin | `unsafe.Pointer`, `chan T`, `func()` — `UnsupportedBuiltin` filter; warn-and-skip. |

~12 cells. Most are concrete and produce visibly different schema
shapes.

### 1.3 Reach context

Where in the type graph the alias is encountered.

| Value | Code path |
|---|---|
| **Decl** — top-level `type X = Y` annotated `swagger:model` / `:parameters` / `:response` | `schema.buildDeclAlias` / `parameters.buildAlias` / `responses.buildAlias` |
| **Field** — struct field whose type is an alias | `schema.buildAlias` (the two-mode field/element entry) |
| **Embed** — struct embeds the alias | `schema.buildEmbedded` → `schema.buildAlias` |
| **Element** — slice/array/map/pointer element | `schema.buildFromType` → `buildAlias` (same as Field) |
| **Transitive** — alias not annotated, discovered via reachability | Same entry as the typed reference that surfaced it |

5 cells. Field and Element collapse to the same path in code; keep
them distinct in the matrix because the surrounding spec shape
differs (a field has a property name + description; an array element
doesn't).

### 1.4 Annotation context (top-level only)

Drives the builder choice and the validation rules.

| Value | Builder |
|---|---|
| `swagger:model X = Y` | schema |
| `swagger:parameters X = Y` | parameters |
| `swagger:response X = Y` | responses |
| `swagger:allOf X = Y` | schema (embedded-decl walk) |
| None (transitive discovery) | schema |

5 cells. Parameters and responses each support fewer RHS kinds
(`parameters.buildAlias:104` rejects `any` and `error` outright;
non-body locations reject body-shaped RHS).

### 1.5 Override at the alias's own decl site

`fd` carries field-doc signals; the decl-site classifiers handle
type-level overrides. These apply BEFORE the alias dispatch in many
cases.

| Value | Where it acts |
|---|---|
| None | (the bulk of cases) |
| `swagger:strfmt X` | Alias decl carrying its own strfmt — short-circuits to `{type:string, format:X}`. |
| `swagger:type X` | Decl-site type override — short-circuits to whatever X resolves to. |
| `swagger:default X` | Decl-site default. Empty-schema rendering trigger. |
| `swagger:name X` | Property rename — only meaningful when the alias surfaces as a property. |
| `swagger:ignore` | Skip entirely. |

6 cells, but most cells × other-axes are uninteresting (most
overrides apply at site, not at the alias decl itself). The
**important** ones to exercise: strfmt and type at the alias decl
site.

### 1.6 Tangential options (likely orthogonal)

`SkipExtensions`, `SetXNullableForPointers`, `DescWithRef`,
`ScanModels`. These change orthogonal aspects of the output but don't
drive alias-specific shape decisions. Hold constant for the workshop;
revisit only if a cell behaves unexpectedly.

---

## 2. Realistic combinatorics

Naive Cartesian: 3 × 12 × 5 × 5 × 6 = **5400 cells**. Most are
impossible or equivalent.

### 2.1 Impossibilities (combinations the type system forbids)

- `RHS=Basic` × `Reach=Embed` — Go doesn't allow embedding primitives.
- `RHS=Slice/Map` × `Reach=Embed` — same.
- `RHS=Pointer-to-Alias` × `Annotation=swagger:model` — possible but
  degenerate (pointer-of-alias dissolves to alias).
- `Annotation=swagger:parameters/response` × `Reach=Field/Embed` —
  parameters/response annotations live on enclosing structs, not on
  fields.
- `Annotation=swagger:parameters` × `RHS=any/error` — rejected at
  `parameters/parameters.go:104`.

### 2.2 Equivalence classes (cells producing the same shape)

- `RHS=*types.Interface (empty)` ≡ `RHS=*types.Alias (any)` —
  `recognizeAny` collapses them.
- `RHS=*types.Slice` ≈ `RHS=*types.Array` — same downstream path
  (`buildFromType` calls `buildFromType(Elem(), Items())` either way).
- `Reach=Field` ≈ `Reach=Element (slice/array element)` — same code
  path; only the surrounding spec context differs.
- `Override=None` is the baseline; override cells should only be
  exercised at the alias decl site, not for every (RHS × mode) cell.

### 2.3 Workshop-relevant residue

After collapsing: ~50-60 cells produce genuinely distinct schema
output. The corpus needs to cover those, at three modes each.

The shape of the matrix (sketch — populated in §4):

```
                     ┌─ Mode (3)
                     │
RHS (8 collapsed) ───┼─ Reach (4 collapsed) ─── Annotation (3 useful)
                     │
                     └─ Override (3 useful: none / strfmt / type)
```

Roughly **8 × 4 × 3 × 3 = 288 (RHS × reach × annotation × override)
expanded to × 3 modes = ~860 cell-visits**. Most cells map to the
same fixture under different modes, so the FIXTURE count is
proportionally smaller — order of 50-80 fixtures expected.

---

## 3. Existing fixture inventory

| Fixture (path under `fixtures/`) | RHS kinds covered | Reach | Annotation | Modes tested |
|---|---|---|---|---|
| `goparsing/go123/aliased/schema/order.go` | int64-alias, any-alias, empty-struct-alias, named-struct (`ExtendedID`), anon-struct field types | decl, field | swagger:model | unknown (likely default — needs checking) |
| `goparsing/go123/aliased/schema/embedded.go` | named-interface, alias-of-interface, anon-interface, named-struct, alias-of-struct, alias-of-any | embed (struct + interface) | swagger:model | unknown |
| `goparsing/go123/aliased/schema/extra.go` | empty-struct, any, interface, anon-iface, slice (`[]any`, `[]struct{}`), alias-of-slice, alias-of-anon-iface, anonymous-struct | decl | swagger:model + redef vs alias | unknown |
| `goparsing/transparentalias/parameters.go` | named-struct (TransparentPayload), named-string (QueryValue) | decl-as-parameters, body field, query field | swagger:parameters | TransparentAliases=true |
| `goparsing/transparentalias/responses.go` | named-struct (ResponseEnvelope) | decl-as-response, body field | swagger:response | TransparentAliases=true |
| `enhancements/alias-expand/api.go` | named-struct, alias-of-alias chain, exported-alias-to-unexported | decl-as-parameters, decl-as-response | swagger:parameters, swagger:response | **default + RefAliases** (two integration tests) |
| `enhancements/alias-findmodel-witness/api.go` | named-struct (unannotated PlainTarget reached via alias) | field (body), field (response body) | swagger:parameters, swagger:response | RefAliases=true |
| `enhancements/alias-response/api.go` | (need to inspect) | response | swagger:response | (need to inspect) |
| `enhancements/alias-response-shapes/api.go` | named-struct, alias-of-alias chain | decl-as-response, body field, header field | swagger:response | **default + RefAliases** (two integration tests) |
| `enhancements/ref-alias-chain/types.go` | named-struct, alias-of-alias, `time.Time` alias, `any` alias | decl, field | swagger:model | RefAliases=true only |
| `enhancements/embedded-types/types.go` | alias-of-struct, anon-interface, error, named-interface | decl, embed | swagger:model | default |

Gaps already evident from this row glance — **before** filling out
the cells properly:

- No fixture covers ALL THREE modes uniformly. Most are tested
  under one mode; a few under two.
- No `swagger:strfmt` or `swagger:type` override at the alias decl
  site.
- `time.Time` decl alias only tested under RefAliases (Q3 reveals
  the default-mode bug is invisible to the existing corpus).
- No "direct-named-struct embed alongside alias-embed" (Q8 gap).
- No exhaustive coverage of `*types.Pointer` as RHS or as
  reach-context.
- No coverage of `*types.Map`-as-RHS at decl level.

---

## 4. The matrix sketch (to fill during step 2 of the corpus walk)

Rather than emit the full Cartesian product here, the workshop walk
populates the matrix incrementally. Each row is one fixture; columns
flag which cells it covers and which modes it's been judged under.

Template for the ledger (`alias-ledger.md`):

```
| Fixture | RHS | Reach | Annotation | Override | Mode | Current shape | Judgment | Note |
|---|---|---|---|---|---|---|---|---|
| q3-witness/time-alias-as-model | time.Time | decl | swagger:model | none | Expand | {type:object} | wrong → {type:string, format:date-time} | Q3 |
| q3-witness/time-alias-as-model | time.Time | decl | swagger:model | none | Ref | {$ref:Time} 2-hop | ??? | Q3 |
| q3-witness/time-alias-as-model | time.Time | decl | swagger:model | none | Transparent | {type:string, format:date-time} inline | ✓ | Q3 |
| …
```

The matrix THEN derives from the ledger — group rows by
judgment-pattern to see which axes drive the differences. That's
where the model emerges.

---

## 5. Gap-filling fixture proposals

Bounded set of new fixtures that close the highest-value gaps. Each
is small (one Go file, 20-50 LOC). Total: **8 new fixtures** for the
schema side, **2 each** for parameters and responses (where Fred's
hunch about smaller combinatorics holds — these surfaces are nearly
saturated already).

### Schema-side additions (8)

1. **`alias-of-stdlib-decl`** — `type Timestamp = time.Time`,
   `type Err = error`, `type Raw = json.RawMessage` all annotated
   `swagger:model`. Three modes each. Closes Q3 and the
   stdlib-special × decl gap.
2. **`alias-vs-direct-embed`** — same struct embedded directly AND
   via alias in two sibling test types. Closes the B2 gap. Three
   modes.
3. **`alias-with-decl-override`** — `swagger:strfmt uuid` and
   `swagger:type string` on aliases of base types. Currently
   untested at the alias decl site. Three modes.
4. **`alias-of-primitive`** — `type UserID = int64`, `type Name =
   string`, `type Active = bool`. The calibration baseline. Three
   modes.
5. **`alias-of-pointer-and-pointer-to-alias`** —
   `type PT = *Target` and `type APT = *Alias`. Pointer-RHS gap.
6. **`alias-of-map`** — `type Headers = map[string]string`. Map-RHS
   gap.
7. **`alias-of-unsupported`** — `type C = chan int`,
   `type F = func()`. Witness the warn-and-skip path. Probably
   produces NO definition; locks that expectation.
8. **`alias-of-generic`** — `type Bag = List[int]` (Go 1.18+).
   Confirms the dissolve-via-Underlying path.

### Parameters / responses additions (2 each)

1. **`alias-params-non-transparent`** — Q7's repro under default
   AND RefAliases (currently `transparentalias/parameters.go` only
   covers Transparent). Three modes.
2. **`alias-response-non-transparent`** — same for responses.

### Harness change (separate, one-time)

A small change to the integration test scaffolding so each
alias-flavored fixture is run under all three modes by default,
producing `<name>_default.json` / `<name>_ref.json` /
`<name>_transparent.json` goldens. This is what generalizes the
existing pattern of "two tests per fixture" to "three tests per
fixture" across the corpus.

---

## 6. Calibration starter — what we walk in cycle 1

`alias-of-primitive` (proposal #4). Smallest possible scope, no
stdlib specials, no embed, no overrides. Just:

```go
package alias_calibration

// swagger:model UserID
type UserID = int64

// swagger:model Name
type Name = string

// swagger:model Active
type Active = bool

// swagger:model Envelope
type Envelope struct {
    User UserID `json:"user"`
    Nick Name   `json:"nick"`
    On   Active `json:"on"`
}
```

Six cells per mode = 18 judgments total. The agreed shape of these
sets the vocabulary for everything else.

Expected first question on each cell: *"Should `UserID` produce a
definition under the name `UserID`, or should we just resolve it to
`int64` inline?"* — the answer to that calibrates §3.1 ("intent of
`type X = Y`") for the simplest case.

---

## 7. Process for the walk

1. **Now:** review this audit. Confirm axes, equivalences, gaps;
   correct anything I've miscounted or mis-categorized.
2. **Then:** I produce the calibration starter fixture and the
   ledger skeleton. You build TUI₁ and TUI₂ from current
   master/feature branches.
3. **First cycle:** you point both TUIs at the calibration fixture;
   I record what each produces in the ledger; you mark each cell
   with judgment + note.
4. **Iterate outward:** next cycle picks the next bounded fixture
   (`alias-of-stdlib-decl` is my nominee — Q3 is concrete and
   high-value). Same TUI-judgment-record loop.
5. **Pattern recognition:** every 3-5 fixtures, I summarize the
   pattern of judgments and propose a candidate rule (which becomes
   our model). You vet the rule against what's been judged.
6. **Implementation cycle:** when a rule stabilizes, I produce the
   smallest code change to apply it; you rebuild TUI₂, compare
   against TUI₁, judge the corpus-wide effect.
7. **Iterate until §5 of `alias-handling.md` writes itself** —
   the decisions stop being abstract picks and become a summary of
   the empirical rules we've already converged on.
