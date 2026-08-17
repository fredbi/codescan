---
title: Godoc-syntax filtering & idiom recomposition — design & plan
stream: 8
origin: i
status: done
release: v0.36
issues: []
supersedes-stub: godoc-filter.md
---

> **✅ P1–P4 complete** on `feat/feature-v0.36` (`ad0e332` P1 · `2845b26` P2 ·
> `1e67b77` P3 · P4 READMEs), awaiting review/merge. Sections below are the
> design + build log.

# Godoc-syntax filtering & idiom recomposition — design & plan

Opt-in knob `Options.CleanGoDoc` (default `false`). When a title/description is
taken **from godoc**, clean the godoc-specific syntax that renders as noise in a
Swagger spec, **and** recompose `[Ident]` doc-links (and the leading
godoc-convention self-name) to the name the referenced object is actually
**exposed under** in the spec.

> **Name decided (Fred):** `CleanGoDoc` — the scope is broader than just links
> (it also recomposes the leading self-name and humanizes unresolved idents), so
> the name should not narrow to "links".

Third member of the clean-godoc cluster, after
[swagger-description-override](swagger-description-override.md) (landed) and
[comment-source-filtering](comment-source-filtering.md) (landed). Backlog stub:
[godoc-filter.md](godoc-filter.md).

**Scope decided (Fred, 2026-06-25):** **A** (link-noise filtering) **+ B** (idiom
recomposition). **C** (inner-markdown normalization) is **deferred** — it is a
separate feature that will leverage the inner-comments machinery; markdown is too
invasive / prone to misinterpretation at this stage. **Opt-in knob**, default
off ⇒ existing output **byte-identical**.

> **Read §3 first** — the consumption-seam-vs-post-reduce timing split is the one
> genuinely load-bearing architectural decision, and §3.1 records the simpler
> alternative I rejected (and why) for your review.

---

## 0. The shapes — godoc doc-link syntax (Go 1.19+)

```
[Name]               // same-package identifier (type/func/var/const/field/method)
[Name.Field]         // dotted member chain
[pkg.Name]           // package-qualified
[pkg.Name.Field]     // qualified + member
[*Name]              // leading star tolerated by go doc
[Text]: https://…    // reference-style link DEFINITION (its own line)
```

Rendered by `go doc` / pkgsite as hyperlinks; carried verbatim into a Swagger
`description`/`title` they read as literal bracket noise:

```go
// Widget is owned by a [Person] and tracked in [inventory.Ledger].
//
// See [the spec]: https://example.com/spec for the wire format.
type Widget struct { … }
```

→ today's `description`: `Widget is owned by a [Person] and tracked in
[inventory.Ledger].\n\nSee [the spec]: https://example.com/spec …`

→ with the knob on (and `Person`→`person`, `inventory.Ledger`→`ledger` exposed):
`Widget is owned by a person and tracked in ledger.\n\nSee the spec …`

### 0.1 Leading self-name recomposition + sentence-initial titleizing (decided, Fred)

Go convention: a doc comment **opens with the symbol's own name**
(`// Widget is owned by…`). When the decl is exposed under a *different* stem,
that leading word is now inconsistent with the emitted schema name. So:

- **Recompose the leading self-name** (the unbracketed leading word, when it
  equals the decl's Go name) to the decl's **exposed** name.
- **Sentence-initial titleizing:** the first identifier of a title/description
  is restored to sentence case (**first letter upper**, regular English),
  whatever the exposed name's actual case. Mid-sentence substitutions keep the
  exposed name verbatim.

```go
// Widget is owned by a [Person] and tracked in [inventory.Ledger].
//
// See [the spec]: https://example.com/spec for the wire format.
//
// swagger:model gizmo
```
→ `description: "Gizmo is owned by a person and tracked in ledger.\n\nSee the spec …"`

(`Widget`→`gizmo` exposed → sentence-initial → `Gizmo`; `[Person]`→`person`
mid-sentence verbatim; the `[the spec]: URL` definition line dropped.)

This is exactly the offset-0 `redactIdentifiers` gate in Fred's tool (§1 ref) —
match the symbol name at byte 0 — but rewriting to the exposed name instead of
blanking.

---

## 1. A — link-noise filtering

A is a pure, local string→string transform on godoc prose. **Regexes adopted from
Fred's battle-tested tool** (`github.com/fredbi/go-fred-mcp/pkg/doc-filters/godoc-filter`):

- **doc-link span:** `\[\*?\w+(?:\.\w+)+\]|\[\*?[A-Z]\w*\]` — a dotted chain
  (`[pkg.Type]`, `[Order.Field]`) **or** an uppercase-led single (`[Widget]`),
  tolerating a leading `*`.
- **reference-definition line:** `(?m)^\[[^\]]+\]:\s+\S+.*$` — dropped whole.
- **lowercase single** `\[[a-z]\w*\]` — in Fred's tool, only redacted when it
  matches the symbol/import index. For us (schemas-only resolution, §2/§3) a
  lowercase single can only be a *package* link, which never resolves to a
  definition → **v1 leaves it intact** (it stays as prose).

Per-link substitution (the text that replaces a stripped span):
1. **Resolves to a schema (§2)** → the exposed definition/property name.
2. **Unresolved** → **humanize** the leaf identifier via
   `mangling.NameMangler.ToHumanNameLower` (`[CustName]` → `cust name`) — Fred:
   "strip the square brackets and … humanize the ident."
3. **Sentence-initial** (first ident of the title/description) → capitalize the
   first rune of whichever of the above applies (§0.1).

The doc-link regex already excludes the common false positives by construction:
`[]byte` (empty), `[0]`/`[42]` (digit-led), `[see notes]`/`[TODO: x]` (spaces),
and bare-lowercase `[id]` (only the index-gated lowercase rule could touch it,
and that rule is inert in v1).

A is self-contained and is the fallback whenever B cannot resolve a link.

> **Ref:** Fred's `pkg/doc-filters/godoc-filter` and `pkg/godoc` — heavily tested
> at rewriting godoc prose (a few known false positives). Note the key difference:
> that tool **redacts** (length-preserving blanking, for masking); codescan
> **rewrites** (strip → humanize → recompose). We borrow the regex *shapes* and
> the offset-0 leading-identifier gate, not the redaction substitution.

---

## 2. B — idiom recomposition

When a doc-link *resolves* to an object that is actually emitted in the spec,
substitute the **exposed** name rather than merely stripping brackets:

| Doc-link | Resolves to | Substituted with |
|---|---|---|
| `[Widget]` | type emitted as definition `widget` | `widget` |
| `[Widget]` (renamed by collision → `modelsWidget`) | that definition | `modelsWidget` |
| `[Order.CustName]` | field exposed (json/NameFromTags) as `customer_name` | `order.customer_name` |
| `[inventory.Ledger]` | cross-package type → `ledger` | `ledger` |
| `[NotAModel]` | unexported / not emitted / unresolved | **falls back to A** → `NotAModel` |

**Per-segment mapping.** The dotted chain is mapped segment-by-segment to exposed
names (`order.customer_name`, not just the leaf), preserving the reference's
shape while making it true against the emitted spec. *(Open: leaf-only `customer_name`
is the alternative — see §9.)*

**Schemas only (decided, Fred).** Resolution targets **definitions** (schemas)
only — the model index. A doc-link to a func that happens to be registered as an
*operation* will **not** resolve (operations aren't in `#/definitions`); that is
expected and falls through to the humanize fallback. This keeps the resolver a
single, well-understood lookup.

**Resolution policy (bounded, v1):**
- same-package `[Name]` / `[Name.Field]` — `pkg.Scope().Lookup(name)` on the
  decl's `types.Package`.
- qualified `[pkg.Name…]` — resolve `pkg` via the enclosing `ast.File`'s imports →
  package path → the loaded `types.Package` (all packages live on `ScanCtx`).
- the **type** segment defers to post-reduce (final name unknown mid-build); the
  **field** segment is resolved immediately (`ParseFieldTag`/`NameFromTags`
  property names are final at build time, never reduced).
- the **leading self-name** (§0.1) is the trivial resolution case: it *is* this
  decl, so its FQ key is known immediately; only its final name defers.
- anything unresolvable → **humanize** (strip brackets, `ToHumanNameLower` the
  leaf), per §1. **No diagnostic** (decided, Fred — a clean godoc must not be made
  noisy by recomposition).

---

## 3. Timing — the consumption-seam vs post-reduce split (load-bearing)

The two halves pull in opposite directions:

- **A needs provenance.** Stripping must apply only to **godoc-derived** prose,
  never to author-written `swagger:description`/`swagger:title` overrides (those
  are deliberate). Provenance exists **only at the consumption seam** — the godoc
  prose enters via `block.PreambleTitle/PreambleDescription/Prose()`, while
  overrides come via the separate `HarvestOverrides` path (§5). After the build,
  every `.Description` is just a string; provenance is gone.
- **B needs final names.** A referenced type's exposed name is fixed only by
  `reduceDefinitionNames()` (`spec.go:130`), which runs **last** and returns
  `renames map[string]string` (FQ key → final name).

**Resolution — resolve early, substitute late, via a marker.** At the consumption
seam (where provenance lives) we:
1. apply A unconditionally to the prose, and
2. for each link that resolves to a **schema** (or the leading self-name),
   replace it with a NUL-delimited **marker** carrying
   `(fqDefKey, fallbackText, titleize)` — e.g.
   `\x00gl\x1f<fqkey>\x1f<humanized-fallback>\x1f<0|1>\x00`. **Unresolved** links
   are humanized in place *now* (no marker — no final name to wait for).

The marker's `titleize` bit records sentence-initial position (§0.1); its
`fallbackText` is the humanized leaf, used if the key vanishes post-prune.

After `reduceDefinitionNames()`, a single spec-wide pass (`resolveGodocLinkMarkers`)
walks every title/description, and for each marker substitutes
`renames[fqkey]` — or the FQ key's default short name if present as a definition —
or the **fallbackText** if the key is not an emitted definition (pruned /
unexported). If `titleize`, the first rune of the result is upper-cased. The pass
**guarantees no marker leaks**: any unmatched marker collapses to its fallback.

`\x00`/`\x1f` cannot occur in Go source comments, so markers never collide with
real prose. With the knob **off**, no marker is ever produced and A never runs ⇒
byte-identical output.

This is the **exact shape of the existing `defOrigins → FlushDefOrigins(finalName)`
pattern** (`scan_context.go`, fired at `spec.go:137`): buffer keyed by FQ key
during build, re-point to final names after reduce. We are not inventing a
mechanism, we are mirroring a proven one.

**Resolution targets schemas only** (§2): a doc-link to an operation-registered
func won't be in the model index → humanize fallback. Expected, not a bug.

### 3.1 Rejected alternative — resolve to the short name at consumption

We *could* skip markers and substitute the type's default short name
(`DefKey()`'s last segment) directly at consumption, dropping the post-reduce
pass entirely. **Rejected:** the default short name is wrong precisely when a
collision forces a rename (`modelsWidget`) — and collision-renaming is an active,
correctness-sensitive concern (name-identity track). A recomposed link is only
prose (not a validated `$ref`), so the blast radius is small, but emitting a name
that does not match the actual definition undercuts the whole point of B
("replace with the object **actually** exposed"). The marker pass costs ~one
traversal we already perform shape-of in reduce. **Recommend markers.** (Flagged
for your call.)

---

## 4. Consumption sites (apply A; emit B markers)

All nine godoc-prose reads, confirmed by audit:

| # | Site | Feeds |
|---|---|---|
| 1 | `spec/walker.go:27-28` | `Info.Title` / `Info.Description` (`swagger:meta`) |
| 2 | `routes/walker.go:83-84` | route op `Summary` / `Description` |
| 3 | `operations/walker.go:24-25` | inline op `Summary` / `Description` |
| 4 | `responses/walker.go:74` | response `Description` |
| 5 | `responses/walker.go:144` | response header `Description` |
| 6 | `parameters/walker.go:78` | parameter `Description` |
| 7 | `schema/walker.go:88-89` | model `Title` / `Description` |
| 8 | `schema/walker.go:132` | field `Description` (non-`$ref`) |
| 9 | `schema/walker.go:208` | field `Description` (`$ref` override / allOf sibling) |

Wiring: a shared helper in `internal/builders/godoclink/`, e.g.
`Clean(text string, opts CleanOpts) string`, called at each site, gated by the
knob. `CleanOpts` carries: the `resolver` closure (identifier → `(fqkey, ok)`,
schemas only — §2), the shared `*mangling.NameMangler` (humanize), the
**self-name** (this decl's Go name, for the §0.1 offset-0 recomposition), and the
self FQ key. A `nil` resolver ⇒ humanize-only.

Meta/Info (#1): resolution is moot (info text rarely references models); strip +
humanize still apply, no self-name. Routes/ops (#2,#3): resolver scoped to the
handler decl's package; self-name = the handler func name.

---

## 5. Provenance: godoc vs override — already separable (confirmed)

Override text is harvested by `common/builder.go:160-172`
`HarvestOverrides(cg) (title, desc OverrideValue)` and **replaces** godoc prose
when `.Present` (e.g. `schema/walker.go:92-98`; responses via
`overriddenDescription`). Because we filter **only the nine godoc-prose reads**,
author-written overrides are never touched — the override remains the clean escape
hatch. No new provenance plumbing needed.

---

## 6. Option

`Options.CleanGoDoc bool` (default `false`), re-exported via `codescan.Options`,
threaded to builders through `ScanCtx` like the other knobs
(`SkipEnumDescriptions`, `NameFromTags`, …). Off ⇒ helper is a no-op, no markers,
byte-identical golden.

---

## 7. Phasing — one commit per phase on `feat/feature-v0.36`

- **P1 — option + A (strip / humanize / def-line / titleize). ✅ DONE.**
  `internal/builders/godoclink/` package: doc-link recognizer + reference-
  definition-line drop + `ToHumanNameLower` substitution + sentence-initial
  capitalize, with table unit tests; `Options.CleanGoDoc` threaded to `ScanCtx`
  (shared `mangling.NameMangler` on `ScanCtx`, `CleanGoDoc` helper on
  `common.Builder` + a sibling on `spec.Builder` for the meta site); wired at all
  nine sites. **No resolution yet** — every link humanizes; leading self-name
  untouched (deferred to P2/P3). Off/on goldens (`enhancements_godoc_links*.json`)
  prove strip+humanize+def-line+false-positive guards; OFF is byte-identical.
  Suite green (19 pkgs), lint clean.
- **P2 — marker contract. ✅ DONE.** *(Rebalanced: the real go/types resolver
  moved to P3 — without the substitution pass a live resolver would only leak raw
  markers into output. P2 delivers the marker machinery so the contract is fully
  tested before the resolver is wired.)* `godoclink.Options{Mangler, Resolver,
  Self}`, `Resolution{DefKey, Suffix}`, marker encode + `SubstituteMarkers` decode
  (NUL/Unit-Separator delimited, fallback + sentence-initial titleize), and
  leading-self-name span handling. Round-trip unit-tested with a fake resolver
  (emit → substitute, incl. pruned-key fallback and rename). Live `CleanGoDoc`
  still passes a nil resolver ⇒ output unchanged, integration goldens untouched.
- **P3 — real resolver + substitution. ✅ DONE.** Resolver in
  `common/godoc.go` (`godocResolver`/`godocSelf`): reuses `ScanCtx.GetModel`
  (so `swagger:model` overrides are honored), same-package + imported (`pkg.Type`
  via the file's imports), per-segment field → exposed property name
  (`resolvers.ParseFieldTag` + `NameFromTags`); wired into `CleanGoDoc` /
  `CleanGoDocSelf` (the latter on the model title/description site). Post-reduce
  `substituteGodocMarkers` pass in `spec/godoc_markers.go` (a `walkSpecProse`
  traversal mirroring `rewriteAllRefs`), inserted right after
  `reduceDefinitionNames()`; `finalName` = `renames[key]` else the leaf checked
  against the final `Definitions` (so pruned/unresolved keys fall back). Fixture
  upgraded to show it: `Widget`→`Gizmo` (swagger:model + self-name, titleized),
  `[Gadget]`→`Gadget`, `[Order.CustName]`→`Order.customer_name` (json field),
  `[inventory.Ledger]`→`Ledger` (cross-package), `[Sprocket]`→humanized fallback,
  `[the spec]: url` dropped, `[0]`/`[see notes]`/`[id]` intact. Suite green, lint
  clean.
  *(Deferred to a follow-up if wanted: field-level leading self-name, nested
  member chains `[Type.A.B]`, dot-imports.)*
- **P4 — docs + stub. ✅ DONE (READMEs only; doc-site is separate).**
  New `internal/builders/godoclink/README.md` (the two-phase marker contract +
  full mechanics); `internal/scanner/README.md` §clean-godoc (option-level,
  cross-referencing it, with TOC entry); [godoc-filter.md](godoc-filter.md) stub
  flipped to done.

> Goldens churn between P1 (everything humanized) and P3 (resolvable links → exposed
> names) — inherent to the layering and acceptable; each phase pins its own golden.

No merge before review (standing rule).

---

## 8. Fixtures (`fixtures/enhancements/godoc-links/`)

Scanned **off** and **on**:
1. model title + description with `[Person]`, `[inventory.Ledger]` doc-links.
2. **leading self-name**: `// Widget …` + `swagger:model gizmo` → `Gizmo …`
   (sentence-initial titleize).
3. field description with `[Order.CustName]` → exposed `order.customer_name`.
4. `swagger:model` override-named target → recomposed to the override name.
5. collision-renamed target → recomposed to the renamed (`modelsWidget`) form.
6. cross-package `[pkg.Type]`.
7. unresolvable `[FooBar]` → humanized `foo bar` (and `FooBar` → `Foo bar` if
   sentence-initial).
8. doc-link to an operation-registered func → unresolved → humanized (§3 note).
9. `[Text]: URL` reference-definition line dropped.
10. false-positive guards: `[]byte`, `[0]`, `[see notes]`, bare `[id]` left **intact**.
11. **off ⇒ byte-identical** golden (control), proving the knob is inert.

Coverage test `internal/integration/coverage_godoc_links_test.go` (off vs on) +
golden(s), per the cluster's established harness.

---

## 9. Decisions & residual risks

**Decided (Fred, 2026-06-25):**
- **Option name** = `CleanGoDoc` (broader than "links").
- **Leading self-name** recomposition + **sentence-initial titleizing** in scope (§0.1).
- **Schemas only** for resolution; operation-funcs fall to humanize (§2/§3).
- **Unresolved** → humanize via `ToHumanNameLower` (no diagnostic) (§1/§2).
- **Markers** for the consumption-seam↔post-reduce bridge (§3, §3.1 rejected alt).
- **Field form** = dotted exposed chain `order.customer_name` (recommended; stands
  unless you prefer leaf-only).

**Residual (flag if you disagree):**
1. **Cross-package depth** — v1 best-effort via file imports; dot-imports /
   re-exports may not resolve → humanize fallback. Acceptable for v1.
2. **Bare-lowercase brackets** `[id]` — left intact (the index-gated lowercase rule
   is inert in v1 since schemas are exported/uppercase). A lowercase same-package
   var/const doc-link would not be recomposed. Acceptable narrowing vs. clobbering
   prose.
3. **P1↔P3 golden churn** — humanized in P1, exposed-name in P3 (noted in §7).
