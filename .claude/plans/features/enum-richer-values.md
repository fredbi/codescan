---
title: Enum on var and richer Go expressions
stream: —
origin: i
status: open
release: v0.37
issues: []
prev: "§1"
---

# Enum on `var` and richer Go expressions

**Status:** ⬜ open · **approach undecided (TODO)** — shaping and priority not yet
settled · spec refreshed 2026-08-02, see [Landscape update](#landscape-update-2026-08-02).

**Origin.** W2 §4.1 (workshop 2026-04-21); raised by Fred as a real use case —
enum values that aren't const-representable (`[]byte`, typed structs, typed
maps). This is a core-product enhancement with no dedicated stream; it rides
whichever stream next touches the scanner / enum seam.

A coherent "richer enum values" feature spanning three sub-concerns:

- **`var` declarations and richer value expressions.** ⚠️ **Half of this shipped under
  go-swagger#3412 — see the landscape update below before planning.** What remains:
  - Lift the hard `token.CONST` gate (`internal/scanner/scan_context.go:909`) — still present, so
    `var (...)` is still invisible. **This is the whole of the remaining immediate need.**
  - Composite literals (`Status{Label: "foo", Code: 1}`) and call-result initializers. These are
    **not const-representable in Go**, so they can only ever arrive via `var`, and the type-checker
    has no `constant.Value` for them. They need a separate value mechanism, not merely a wider gate.
- **Inline YAML-block enum values.** A surface where the `enum:` keyword has a
  block body that carries non-scalar values:

  ```go
  // enum:
  //   - complex: {object: value}
  //   - another: [array, element]
  ```

  When `enum:` has a block body (captured as `Property.Body`), hand it to
  `internal/parsers/yaml/` and fold the result into the returned `[]any`.
- **Non-scalar emission parity audit.** v1's JSON path accepts objects, arrays,
  and null, but no fixture demonstrates emission. Before any release claims
  "OpenAPI-complete enum support", add object-valued and array-valued enum
  fixtures and confirm the full pipeline (parser → sub-parser → schema builder
  → spec) round-trips them.

**Already done — v1 enum quirks (§1.2a).** Leading-whitespace in comma lists and
the silent drop of `swagger:enum TypeName` with no matching consts were
**migration-commit obligations fixed in the P5.1 schema-builder migration**. The
`swagger:enum TypeName`-without-consts case now raises
`parse.invalid-enum-option` ("no matching const values found; enum semantics
dropped"), verified 2026-07-30.

**§1.2b — residual: an inline override discards the type's per-value docs,
silently.** The original §1.2a entry claimed the `x-go-enum-desc` quirk was
closed with "nothing remains to do". Half of it was: the *stale* desc no longer
survives to contradict a narrowed enum — it is stripped. What remains is that the
strip is **silent and lossy**. Probed 2026-07-30 against
`fixtures/enhancements/enum-overrides` (case E):

```go
// swagger:enum PriorityE     → consts low / medium / high, each with a doc comment
type PriorityE string

type NotificationE struct {
    // enum: urgent, normal   ← the field narrows to values the type never declared
    Priority PriorityE `json:"priority"`
}
```

emits `{"type":"string","enum":["urgent","normal"]}` — no `x-go-enum-desc`, and
**no diagnostic** (only case D, the no-consts one, warns). So the per-value docs
`PriorityE` contributed vanish without a trace.

Whether that is a bug depends on intent, which is why it sits with this feature
rather than in the quirk register: if the field's override is a genuine narrowing
to *different* values, dropping the type's docs is correct and only the silence is
wrong (→ emit a Hint). If overrides are meant to *subset* the type's values, the
matching docs should be carried through (→ filter `x-go-enum-desc` rather than
strip it). Decide when this feature is picked up; it is the smallest useful slice
of it.

**Update 2026-08-02 — the silence half is settled by precedent.** See
[Landscape update §3](#3-12b-now-has-a-precedent-to-follow): `47710cf` established that when an enum
member cannot be honoured it is dropped *and named in a warning*, because an enum is a closed set and
narrowing it changes the author's contract. Discarding the per-value docs a type contributed is the
same act. So the remaining decision is only **filter vs. strip-and-warn** — staying silent is off the
table.

## Landscape update 2026-08-02

Three things moved under this feature since it was written. None completes it; two change how the
remaining work should be shaped, and one supplies a decision it was waiting on.

### 1. The value-reading mechanism changed (go-swagger#3412) — sub-concern 1 is half done

The spec above was written against a scanner that read member values from **literal syntax**, which
is why it framed the work as "extend `findEnumValue` to accept value expressions beyond
`*ast.BasicLit`". That framing is obsolete. `enumMemberValue` now takes values from the
**type-checker** (`TypesInfo.Defs[name].(*types.Const).Val()`), with the AST literal reader surviving
only as a degraded-load fallback. So every *const-representable* expression already resolves —
`iota`, `1 << 3`, references to earlier members, rune literals, every integer base, above-`MaxInt64`.
"Richer Go expressions" is, for consts, **done**.

That sharpens what is left into two genuinely different problems that the old wording ran together:

| remaining | why it is not just "a wider gate" |
|---|---|
| `var (...)` with constant-foldable initialisers | The type-checker *can* supply a value (`TypesInfo.Types[expr].Value`); the only obstacle is the `token.CONST` gate. **Small.** |
| composite literals, call results, typed maps / `[]byte` | Not const-representable, so no `constant.Value` exists at all. Needs its own value mechanism — AST reading, or a decision to exclude. **This is the real feature.** |

Fred's original use case (`[]byte`, typed structs, typed maps) lands squarely in the second row, so
lifting the gate alone will not serve it.

### 2. Two enum defects fixed on the KEYWORD side — and a boundary confirmed

Fixed on `fix/strfmt-dispatch-symmetry`, both about the `enum:` keyword rather than this feature's
`swagger:enum` annotation:

- **Q36 / `94845da`** — `enum: 1,2,3` on a named int emitted `["1","2","3"]` at declaration sites: an
  enum of strings on an integer schema, which no validator accepts. Decl-site keywords were coerced
  before the Go type was resolved.
- **`47710cf`** — an uncoercible member (`enum: 1, two, 3` on an int) produced the mixed array
  `[1,"two",3]` at *every* site. Now the bad member is dropped and named in a warning.

Neither touches the annotation path, but both were in the enum pipeline this feature audits, and the
second sets a **precedent for §1.2b** (below).

**Boundary confirmed (Q32):** `swagger:enum` on an *alias to a basic type* is unfixable — the
type-checker erases `type Unsigned = uint64`, so `const Zero Unsigned = 0` is indistinguishable from
any other `uint64` constant. Whatever "richer values" grows to cover, it cannot cover that. An alias
to a *named* enum type already works.

**Naming hazard now documented.** `swagger:enum` (annotation, members collected from a `const` block,
typed from the declared Go type) and `enum:` (keyword, members written literally, typed from the
schema it sits on) produce the same spec keyword from opposite directions. A comparison table now
sits in the enumerations tutorial, cross-linked from both reference pages. Worth reading before
scoping this feature — several of the "enum quirks" that accumulated were keyword-side, not
annotation-side.

### 3. §1.2b now has a precedent to follow

§1.2b asks whether the silent, lossy strip of `x-go-enum-desc` should become a Hint, or whether
matching docs should be filtered through. The first half — *should it be silent?* — is now answered
by house precedent: when an enum member cannot be honoured, `47710cf` **drops it and names it in a
warning**, on the grounds that an enum is a closed set and narrowing it changes the author's
contract. The same reasoning applies verbatim to discarding per-value docs the type contributed.

So §1.2b reduces to the substantive half: *filter* the descs to the surviving members, or *strip and
warn*. Silence is no longer one of the options.

### 4. The non-scalar audit gap is unchanged — and verified

Re-checked 2026-08-02: **zero** non-scalar enum members (object, array, null) appear anywhere in the
golden corpus. The audit in sub-concern 3 is still entirely outstanding, and still blocks any
"spec-complete enum" claim.

**When to revisit.** A scanner feature, logically independent of the
parser-migration — pick up when non-const enum values are demanded by real
users. The non-scalar audit blocks any "spec-complete enum" marketing claim, but
not a release inheriting v1's scalar-only corpus.

If it is picked up, the cheapest useful order is now clear: §1.2b's decision (smallest, and precedent
already set) → the `var` gate for const-foldable initialisers (small, mechanism exists) → the
non-scalar audit (needed before any completeness claim) → non-const values (the actual feature).
