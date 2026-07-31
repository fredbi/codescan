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

Last verified: **2026-07-30** · repo side synced by PR #68.

---

## 1. Open — needs a decision

### Q31 — `json:"-"` on a re-declared field evicts a promoted property Go still marshals

**Status:** STILL PRESENT · **decision open** (Fred, 2026-07-30: "I don't really
know how to handle this yet").

`internal/builders/schema/fields.go` deletes a promoted property when an outer
field re-declares it with `json:"-"`. `encoding/json` does not: a `-` field is
ignored **entirely**, never enters the name set, never shadows the promoted
field — so Go keeps marshalling the embedded one and the emitted schema
**understates the wire**.

```text
type OverridingOneIgnore struct { SimpleOne; Age int32 `json:"-"` }

json.Marshal → {"id":1,"name":"n","age":42}   // age present, from SimpleOne
codescan     → {id, name}                     // age deleted
```

Locked by `TestOverridingOneIgnore` (`schema/schema_test.go:512`), so a fix must
rewrite that test too. Full write-up, the four options, and why none is obvious:
**`archive/observed-quirks.md` §Q31** and `swagger-omit.md` §7.

**Blocks:** repointing `fixtures/bugs/1992` at the mechanism go-swagger#1992 is
actually about (it currently witnesses `readOnly`, which belongs to #1063).

**Mitigated by:** `swagger:omit` (PR #67) does the job honestly; the Hint
`scan.shadowed-embed-field` now fires on the shape and points at it.

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
| D1, D2, D5 | "Not attempted" | RESOLVED, PR #32 (`c896cc7`, `e2ec828`) — per `archive/observed-quirks.md`'s own closing tally |
| D6 | "Not attempted" | REFRAMED, PR #32 (`c9eabd9`) |
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
