# TUI UX round 2 — real JSON pointers from the validator, plus UX follow-ups

Base camp: worktree `.worktrees/fix/tui-ux-round2`, branch `tui-ux-round2`, off
master `0273dcf` (the round-1 merge, PR #92).

Round 1 shipped the validation tab with a documented caveat: a finding's location
was recovered by converting the validator's **dotted** path (`paths./pets.get.responses.200`)
into a JSON pointer, and resolving it by walking up to the nearest node that
actually exists. This round removes the need for that guesswork.

## Upstream: the fix already exists in go-openapi/validate

Local checkout `/home/fred/src/github.com/go-openapi/validate`, branch
`fix/error-location-as-jsonpointer` — 3 commits, plus uncommitted work in
`ref_locations.go`, `spec.go` and their tests:

| Commit | Subject |
|--------|---------|
| `4c55223` | feat: report the JSON pointer of each validation error |
| `97d6cca` | feat: locate the $ref diagnostics in the document |
| `db6c34d` | fix: do not read $ref declarations out of default values |

The new surface is small and exactly what the TUI needs:

```go
// Located pairs a validation error with the location of the value that caused it.
type Located struct {
    Err     error
    Pointer string // RFC 6901, relative to the validated document; empty when unknown
}

func (r *Result) LocatedErrors() []Located
func (r *Result) LocatedWarnings() []Located
```

It builds against **published** dependencies (it wants `analysis v0.26.0`, which
is released), so the only unreleased piece is validate itself.

## Measured, not assumed

Scanned `fixtures/goparsing/classification` (the corpus in the round-1
screenshots) to a spec, then ran both paths over the same document and resolved
every pointer against it with an independent RFC 6901 resolver:

| | findings | resolve **exactly** | land on an ancestor |
|---|---|---|---|
| master today — released `v0.26.1` + our dotted-path conversion | 25 | 16 (64%) | 9 |
| this work — branch + `Located.Pointer` | 29 | **23 (79%)** | 6 |

The improvement is precisely the caveat we documented: **array indices**.
Master produces `/paths/~1orders/post/parameters/in`, which cannot resolve
because the parameter list is an array; the branch produces
`/paths/~1orders/post/parameters/1/in`.

Every one of the 6 remaining ancestor-only cases is a **required-but-missing**
finding, whose whole complaint is that the node it names is absent
(`in`, `type`, `items`, `schema` in body is required). Those are irreducible: no
pointer can resolve to a node that does not exist. So the walk-up stays, but only
for the case where it is logically necessary rather than as a workaround for the
notation.

Also measured: **no finding had an empty pointer**, and the branch reports 29
where the release reports 25. Worth understanding before landing (finer
granularity? the new `$ref` diagnostics? duplicate reporting?) — see open
questions.

## Outcome — settled against released validate v0.26.2

Re-measured over the same corpus (601 spec files scanned from every
`fixtures/goparsing`, `enhancements` and `bugs` tree plus the integration goldens),
resolving all **754** findings against their own document:

| | resolve exactly | ancestor-only |
|---|---|---|
| released `v0.26.1` + our dotted conversion (before) | 126 | 628 |
| released **`v0.26.2`** + `Located.Pointer` | **752** | 2 |

568 of the 752 are the **root**: `v0.26.2` reports a finding about something the
document lacks entirely against the whole document, which RFC 6901 spells as the
EMPTY pointer. That needed a change on our side — we had read empty as "names
nowhere" and refused to navigate for the commonest finding there is.

The 2 remaining are the limitation upstream documents: a finding inside a response
or parameter written as a `$ref` is addressed through the referring site, and the
authored document has nothing below it. The same finding is reported again against
the shared definition, and that one is exact.

### Still open upstream (pre-existing, not from this work)

`validateRequiredDefinitions` ranges over the definitions **map** and does
`break DEFINITIONS` on the first error, so with the default
`ContinueOnErrors: false` a document with two such faults reports whichever
definition Go's randomised map order reached first — the other appears in roughly
2 runs in 10. The same range and break are on validate's `master`, and `9b2c746`
did not change them. A finding that vanishes half the time is worse than one that
lands imprecisely.

Our tests work around it by keeping one fault per document.

## Workstreams

### A — consume the located API ✅ DONE (`2db8680`)

- `internal/ux/validation`: take `Pointer` from `LocatedErrors()` /
  `LocatedWarnings()` instead of deriving it. `pointerFor` and the dotted-path
  splitting go away; `locationOf` keeps a narrower job (the message still carries
  the human-readable path).
- Keep a fallback for an empty `Pointer` — the upstream contract allows it, so it
  must not read as "the document root".
- Keep the walk-up in `validationTarget`, now justified by the required-missing
  case alone; say so in the comment.
- `TestValidation_PointerResolutionAccuracy` is the round-1 test that measured
  this. Re-point it at the new numbers so it records the improvement rather than
  the old caveat.

### B — dependency ✅ DONE: pinned to released `validate v0.26.2`, no `replace`

The branch is local-only and untagged. Nothing in workstream A can land in a
committable state until we settle how codescan consumes it.

### C — documentation ✅ DONE (folded into `2db8680`)

- `docs/doc-site/getting-started/usage-as-a-tui.md`: the "Where a finding lands"
  notice describes the array-index case as one of two imprecisions. With the fix
  only the required-missing case survives, so the notice shrinks.
- `cmd/genspec-tui/README.md`: same, in the "Validating the generated spec"
  section.
- Both currently say the imprecision is "the validator's notation rather than the
  conversion". That framing stays true and becomes historical — the notation was
  fixed.

### D — other UX enhancements (unscoped; Fred's call)

Carried over as known and deliberate, from round 1's Limitations:

- **Theming (light/dark)** — deferred in round 1 as noise, never built.
- **The editor rewrites whitespace on save** — `bubbles/textarea` expands tabs
  and normalises CRLF, so `Ctrl-S` reformats a tab-indented file. The sharpest
  remaining footgun.
- **Edit mode is unhighlighted** — textarea owns its rendering.
- **Split sizes are session-only** — persisting needs a config file the TUI does
  not have.
- **`$ref` resolution is a site index** — no ref-to-ref chains, no `$ref` nested
  in `allOf`.
- **~574 comment lines broken mid-clause** — reported in round 1, deliberately
  not swept.

## Where this stands

Two commits on `tui-ux-round2`, both green, neither pinned to a validate version:

| | |
|---|---|
| `0ed83f2` | `fix(genspec-tui): show the validator's warnings` — works against **released** validate, so it can land on its own |
| `2db8680` | `feat(genspec-tui): take a finding's location from the validator` — **needs the unreleased branch**; build with the local `replace` |

The `replace` lives in `cmd/genspec-tui/go.mod` in the worktree and is
deliberately uncommitted, along with the two indirect bumps it drags in
(`analysis v0.26.0`, `swag/fileutils v0.28.0`). Both belong in the same commit as
the eventual version pin.

### Upstream bug found on the way: the finding SET is order-dependent

`validateRequiredDefinitions` ranges over the definitions **map** and does
`break DEFINITIONS` on the first error, so with the default
`ContinueOnErrors: false` a document with two such faults reports whichever Go's
randomised map order reached first — and the second only sometimes. Measured: 2 of
8 runs on the same document.

**Pre-existing**, not from this branch: the same range and break are on validate's
`master`. Worth reporting upstream, since it makes the reported set of findings
unstable run to run, which is worse than a location being imprecise.

## Open questions

1. **Why 29 findings against 25?** Not yet understood. The extra rows look like
   finer granularity (`/allowEmptyValue` reported separately) rather than
   duplicates, but that is an impression, not a measurement.
2. **Is the `$ref`-traversal location worth fixing upstream?** `Located.Pointer`
   is documented as relative to the validated document, and for a finding reached
   through a `$ref` that document is the expanded one — so the pointer can name a
   node the authored spec does not have. Documented as a limitation for now.
3. **Which of D, if any?** Deferred: pointers first.
