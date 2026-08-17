> [!NOTE]
> **✅ ALMOST COMPLETE** — archived 2026-08-17. Merged (PR #77). One item outlived it and is a calendar trigger rather than a decision: retire the `//go:build go1.27` tags at go1.28 GA — [`backlog.md`](../backlog.md) §4.2.
>
> _Original document follows unchanged._

> [!NOTE]
> Last revision: 2026-08-01

# go1.27 stdlib `uuid` — identity-based detection

**Status: ✅ MERGED** (PR #77, 2026-08-01). Two items outlived the branch — see Actions §4:
the build-tag removal when min supported Go ≥ 1.27, and `forcing-a-format.md`, deferred to the Q33 fix.

## Summary

go1.27 (≈ one month out) ships a stdlib `uuid` package: `type UUID [16]byte` at package path `uuid`, with a
value-receiver `MarshalText`. Today codescan recognises "a UUID" by **name only** (case-insensitive `uuid`),
caller-gated to `buildFromTextMarshal`. That heuristic already lands `{string, format: uuid}` on the stdlib type
in every common position — so this is a **precision & robustness** job, not a missing feature.

We add an identity recognizer (`Pkg().Path() == "uuid"` + `Name() == "UUID"`, exactly the shape of `IsStdTime`),
always on, no Option, promoted into the safe recognizer set so it fires at **every** call site — including the
embedded / allOf arms that the fuzzy match never reaches, and independent of the `IsTextMarshaler` gate. The
name-based heuristic stays as the fallback for `github.com/google/uuid` and friends.

## Context

Empirical findings, all measured against `~/go/bin/go1.27rc1` on 2026-08-01 (probes were throwaway; results below).

**Current behaviour is already mostly right.** Scanning a go1.27 module with codescan as-is yields
`{type: string, format: uuid}` for plain fields, pointers, slices, named aliases (`type AliasUUID uuid.UUID`),
path & query parameters, and response headers. `x/tools` v0.48.0 reads go1.27 export data without complaint.
Map keys (`map[uuid.UUID]string`) resolve correctly through `IsJSONMapKey` → `IsTextMarshaler`.

**Gap 1 — the embedded arm.** `struct { uuid.UUID; Name string }` emits `{object, properties:{name}}`. The
fuzzy recognizer is caller-gated to `buildFromTextMarshal`, so `applyStdlibSpecials`' call sites in
`embedded.go` / `allof.go` never see it. Note the emitted object is wrong for a *second*, uuid-agnostic reason —
see Q32 below and §Appendix.

**Gap 2 — a guess gated behind a guess.** The name match is a heuristic, and it sits downstream of the
`IsTextMarshaler` synthetic-interface machinery (`.claude/plans/textmarshaler-toolchain-independence.md`). If
that gate ever misses, `uuid.UUID` degrades to `{array, items:{integer, uint8}}` — a badly wrong spec for a
type we can identify with certainty.

**Why the production recognizer must NOT carry `//go:build go1.27`.** codescan never imports `uuid`; the check
compares a `*types.TypeName` harvested from **scanned** code. A build tag constrains the toolchain compiling
*codescan*, not the code it scans — and codescan ships as a compiled binary embedded in downstream tools
(go-swagger), run in a foreign environment. Gating the recognizer would give exactly the inverted matrix: a
codescan built on go1.26 would silently fail to recognize a user's go1.27 uuid. Verified: a module with
`go 1.25.0` built under toolchain 1.27 imports `uuid` fine (build *and* vet clean), and `packages.Load` hands
back `Pkg().Path() == "uuid"`.

**Build-tag semantics, verified (this is the subtle bit).** Release tags follow the **toolchain**, not the
module's `go` directive. A `//go:build go1.27` file inside the `go 1.25.0` fixtures module compiles under
1.27rc1 and is silently excluded under 1.26 — no "package uuid is not in std", `go vet` clean. Consequences:
`fixtures/go.mod` needs **no bump**, and the tagged tests start running by themselves the day CI's `stable`
becomes go1.27. No bespoke rc workflow.

Decisions locked with Fred (2026-08-01):

- always on, no Option — same class of fix as `time.Time`;
- identity on `Pkg().Path()` + `Name()` only; **no** underlying `[16]byte` shape assertion (not worth it,
  mirrors `IsStdTime`);
- keep the name heuristic + `buildFromTextMarshal` fallback — people will use google/uuid and others for a while;
- build tag is for **testing only**; CI picks it up when `stable` becomes go1.27.

**Where things live.** `.claude/plans/` is gitignored (`.claude/.gitignore`), so plan docs are per-worktree
untracked files. This document and `quirks-open.md` stay in the **root worktree**
(`/home/fred/src/github.com/go-openapi/codescan`, master); implementation happens on branch `stdlib-uuid` in
`.worktrees/feat/stdlib-uuid`, which can be torn down without losing either.

## Trajectory

1. ✅ Production recognizer (untagged, always on)
   * ✅ `resolvers.IsStdUUID` beside `IsStdTime`
   * ✅ `recognizeStdUUID` in the canonical safe set, identity-before-fuzzy precedence
2. ✅ Tests 🏁
   * ✅ Toolchain-independent unit test on the predicate (runs on every supported Go)
   * ✅ `//go:build go1.27` fixture + integration golden (runs when stable is 1.27)
3. ✅ Documentation 📚
   * ✅ `internal/builders/schema/README.md` §special-types — fuzzy vs identity
   * ✅ doc-site `tutorials/model-definitions.md` — identity vs name recognition
   * ⛔ doc-site `forcing-a-format.md` — deferred to the Q33 fix
4. ⏳ Housekeeping 😇
   * ✅ Register Q33 (TextMarshaler embed) — out of scope, its own decision
   * 📝 Schedule the tag removal for when min supported Go ≥ 1.27

## Actions

### 1. Production recognizer

1. 📝 **`IsStdUUID`** — `internal/builders/resolvers/assertions.go`, next to `IsStdTime`:
   `o.Pkg() != nil && o.Pkg().Path() == "uuid" && o.Name() == "UUID"`. Path, not `Pkg().Name()` — `IsStdTime`
   uses the looser `Name()`; don't touch it here, but don't copy it either.

2. 📝 **`recognizeStdUUID`** — `internal/builders/schema/special_types.go`:
   * new `recognizeType` constant, identity ⇒ safe;
   * add it to `applyStdlibSpecials`' canonical set, so it reaches all `applyStdlibSpecials` call sites
     (`schema.go` ×6, `embedded.go` ×2, `allof.go` ×1) including the embedded / allOf arms;
   * in `buildFromTextMarshal`'s variadic, order it **before** `recognizeUUID` so the certain match wins;
   * `recognizeUUID` (fuzzy) stays exactly where it is, caller-gated, as the fallback.
   * Explicit `swagger:strfmt` / `swagger:type` keep winning over both — the step-4-before-5 precedence in
     §textmarshal-order is untouched.

3. 📝 **Golden-neutral by construction** — no current fixture can import stdlib `uuid`, so the integration
   goldens must not move. Any drift is a bug in this change; run the integration suite as the check.

### 2. Tests 🏁

1. 📝 **Predicate unit test (untagged, primary)** — `internal/builders/resolvers/resolvers_test.go`: synthesize
   `types.NewPackage("uuid", "uuid")` + a `types.NewTypeName`, assert `IsStdUUID` true; assert near-misses
   false (`github.com/google/uuid.UUID`, a local `uuid` package at a module path, `uuid.Nil`, nil `Pkg()`).
   This runs on **every** supported toolchain, so the predicate is never untested — the tagged fixture below
   only proves end-to-end wiring.

2. 📝 **Fixture** — `fixtures/goparsing/go127/uuid/` (follows the `go118` / `go119` / `go123` naming
   precedent):
   * `model.go` with `//go:build go1.27` — field, pointer, slice, map value, named alias, **embed**, plus a
     `swagger:strfmt date` override to witness classifier-beats-recognizer;
   * an **untagged sibling** `doc.go` (package clause only) so the package is never file-less under an older
     toolchain — otherwise `build constraints exclude all Go files`.
   * `fixtures/go.mod` stays at `go 1.25.0`.

3. 📝 **Integration test** — `internal/integration/schema_uuid_go127_test.go` with `//go:build go1.27`, golden
   `fixtures/integration/golden/go127_uuid_spec.json`. Under 1.25 / 1.26 it isn't built — which also means
   `UPDATE_GOLDEN=1` on an older toolchain cannot clobber the golden.

4. 📝 **Local verification** — `GOTOOLCHAIN=local ~/go/bin/go1.27rc1 test ./...` before landing; and the same
   with the system Go, to confirm the tagged files vanish cleanly.

5. ⛔ **Bespoke rc CI job** — not doing it. The shared workflow's `stable` will start running these tests at
   go1.27 GA on its own; an rc-pinned job would only buy ~1 month of signal for a permanent maintenance cost.

### 3. Documentation 📚

1. 📝 `internal/builders/schema/README.md` §special-types — the section currently describes `recognizeUUID` as
   "the fuzzy one, opt-in via the variadic". Restate as a pair: identity `recognizeStdUUID` in the safe set,
   fuzzy `recognizeUUID` still caller-gated as fallback. Also fix the stale "seven call sites" count while there.
2. ✅ doc-site `tutorials/model-definitions.md` §swagger:model — the two recognition rules stated beside the
   existing `time.Time` sentence: *by type identity* (stdlib go1.27 `uuid.UUID`) vs *by type name* (anything
   else named UUID that marshals as text), with `swagger:strfmt` overruling either. Prose only, deliberately
   no `{{< example >}}`: a live witness would need a go1.27 example package, and `docs/examples` is built by
   the same CI Go as everything else.
3. ⛔ doc-site `shaping-the-output/field-types-and-formats/forcing-a-format.md` — Fred, 2026-08-01: update it
   when Q33 is addressed, not now.

### 4. Housekeeping 😇

1. ✅ **Q33 in `quirks-open.md`** — a struct embedding a `TextMarshaler` with no `MarshalJSON` marshals as a
   **bare string**; codescan emits an object and drops the embed. Reproduced on go1.26 with a local
   `type Token [16]byte` — uuid-agnostic, needs no go1.27. Out of scope here: it is behaviour-changing and
   deserves its own decision. See §Appendix.
2. 📝 **Scheduled cleanup** — repo policy is the 2 latest stable minors. At go1.28 GA the supported set becomes
   1.27 + 1.28, and the `//go:build go1.27` tags can come out, folding the fixture into the plain `go12x` style.

## Achievements

✅ **MERGED 2026-08-01** via PR #77 (`dd9a25b`), commits `4abea75` (feature) + `cb823cb` (doc-site), rebased
onto `c37d503`. CI green — with the caveat that stable/oldstable (1.26/1.25) excluded both go1.27-tagged
files, so CI proved golden-neutrality and the untagged predicate; the stdlib-uuid path itself rests on the
local go1.27rc1 runs until `stable` flips.

### 1. Production recognizer ⭐⭐

- ✅ `resolvers.IsStdUUID` — `assertions.go`, path+name identity, untagged, with the rationale for *not*
  tagging it recorded in the godoc.
- ✅ `recognizeStdUUID` into `applyStdlibSpecials`' canonical set; `buildFromTextMarshal` orders it before the
  fuzzy `recognizeUUID`. Fuzzy path untouched, so google/uuid & friends keep working exactly as before.
- ✅ **Golden-neutrality proven, not assumed**: full suite green on go1.26 *and* go1.27rc1, `git status` shows
  no existing golden moved — only the new `go127_uuid_spec.json`. `-race` green across all builders +
  integration; `golangci-lint run --new-from-rev master` → 0 issues.

### 2. Tests 🏁 ⭐⭐

- ✅ `TestIsStdUUID` (untagged, `internal/builders/resolvers/assertions_test.go`) — synthesizes the
  `*types.TypeName`, so the predicate is covered on **every** supported toolchain, including the ones where
  `import "uuid"` cannot compile. Covers the true case plus five near-misses (google/gofrs/strfmt, a user
  package merely *named* uuid, another type in the stdlib uuid package, case sensitivity, nil `Pkg()`).
- ✅ `fixtures/goparsing/go127/uuid/` — tagged `model.go` + untagged `doc.go`. Verified the sibling is load
  bearing: `go build ./...` and `go vet` in the fixtures module are clean on go1.26 *and* go1.27rc1, and
  `fixtures/go.mod` stayed at `go 1.25.0`.
- ✅ `TestStdlibUUID` (tagged, integration) + `go127_uuid_spec.json`. Field / pointer / slice / map value /
  map **key** / named type / alias / `swagger:strfmt`-beats-recognizer, plus the embed witness below.

### 3. Documentation 📚 ⭐⭐

- ✅ `schema/README.md` §special-types rewritten around the **certain/guessed pair** — why `recognizeStdUUID`
  is in the safe set, why it carries no build tag, why `recognizeUUID` stays caller-gated, and the embed
  exception. §textmarshal-order step 5 and the TOC line updated.
- ✅ doc-site `tutorials/model-definitions.md` §swagger:model — the user-facing half: identity vs name, and
  `swagger:strfmt` overruling both. The same-page `[…](#swaggerstrfmt)` anchor matches the site's existing
  convention (`routes-and-operations.md` does the same); the MD051 the markdown linter reports on it is
  pre-existing noise on colon-bearing headings, not new breakage — now tracked as
  `doc-site-wishlist.md` **W18**.
- ⛔ doc-site `forcing-a-format.md` — deferred to the Q33 fix, per Fred.

### Correction to the design ❌→✅

The plan predicted that promoting the recognizer into the safe set would make the **embedded** arm fire. It
does not, and the golden proves it: `Embedder` still emits `{object, properties:{name}}`. Reason found while
verifying — `buildNamedEmbedded` switches on the embedded type's *underlying* shape, and `uuid.UUID`'s array
underlying falls to the `default` arm (`CodeUnsupportedGoType` Warning, embed skipped); only the *interface*
arm consults `applyStdlibSpecials`, an asymmetry README §embedded calls intentional. So Q33 is strictly
unchanged by this work. `quirks-open.md` Q33 and the README were corrected to say so, and `TestStdlibUUID`
pins the current behaviour *including* the warning, so the day Q33 is decided the sub-test and golden move
together.

## Appendix — Q32, the TextMarshaler embed

Verified on go1.26, current master, no go1.27 involved:

```text
type Token [16]byte
func (t Token) MarshalText() ([]byte, error) { return []byte("tok"), nil }
type Embedder struct { Token; Name string `json:"name"` }

json.Marshal → "tok"                                   // bare string; `name` never emitted
codescan     → {object, properties:{name}}             // object; embed dropped
```

Wrong in both directions: the promoted `MarshalText` makes the *whole struct* render as a string, so `name` is
not on the wire, and the embed contributes no property. Under go1.27 this is what `struct { uuid.UUID; … }`
will hit, which is how it surfaced — but the mechanism has nothing to do with uuid. Any fix is behaviour-changing
for every existing `strfmt.UUID` / `time.Time` embed, hence: register, decide separately.

**Corrected after implementation:** this work does not move the needle here at all. `buildNamedEmbedded`
switches on the *underlying* shape and sends `uuid.UUID`'s array underlying to the `default` arm — a
`CodeUnsupportedGoType` Warning and a skipped embed — so the embed never reaches a recognizer. Only the
interface arm consults `applyStdlibSpecials`. Verified against the new golden.
