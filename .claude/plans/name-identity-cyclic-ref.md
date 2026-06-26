# Name identity & cyclic-$ref — design track (go-swagger #2637 + #2783)

Date: 2026-06-15 (model converged 2026-06-15)
Status: 🟢 **Stage 0 ✅ + Stage 1 ✅ (committed `1ef2f25`) + Stage 2 ✅ DONE** (2026-06-16) — concat rung landed: collisions now get minimal-depth PascalCase names (BTest/CTest, AMongoBook/BMongoBook, …) + collision diagnostic + cyclic-collision witness; full suite + race + lint green; only collision goldens changed. **Next: Stage 3 (hierarchical fail-safe — K2 decision: include or defer)** — pending Fred's review of Stage 2
Owner: Fred (sponsor) + agent
Audience: maintainers (private — `.claude/plans/`)

This is the design track for **name-identity / cyclic-$ref**. §1–§4 state the
problem, the verified current machinery, and the goals/invariants. §5–§8 were
the design-space exploration — **now superseded by §9 (the converged model)**.
§10 preserves the wrinkles surfaced during the debate, §11 the open knobs Fred
will annotate, §12 the phasing sketch.

Related: #2651 (operation/parameters binding) and #3211 (markdown here-doc) are
separate tracks. #2637 and #2783 are the **same family** and ride one branch
(`fix/go-swagger-2783`).

---

## 1. The problem in one sentence

A spec **definition is identified by its short Go type name** (or a
`swagger:model <name>` override) — with no package component — so two distinct
Go types that share a short name collide on one definition key. This short-name
key is a **proxy for true model identity**, and it breaks in two observable ways.

### Witness A — silent merge (#2783)

`fixtures/bugs/2783`: package `b` and package `c` each declare
`// swagger:model Test`. `TestResponseBody` references both (`b.Test`, `c.Test`).

Observed: one `Test` definition, **union of both structs' fields** (`b` + `c`),
**last-package-wins and non-deterministic** between runs. Both `$ref`s point at
the single `#/definitions/Test`. The two genuinely-distinct schemas are conflated.

### Witness B — degenerate self-`$ref` (#2637)

`fixtures/bugs/2637`: `type CreateDomainRequest mongo.CreateDomainRequest`
(a defined type whose RHS is a **same-named** `swagger:model` in another
package). Both want the key `CreateDomainRequest`.

Observed: the definition is emitted as `{"$ref": "#/definitions/CreateDomainRequest"}`
— **a definition whose entire body is a reference to itself**. Invalid OAS;
hangs downstream codegen.

> Note the asymmetry to *legitimate* recursion: `type Node struct { Next *Node }`
> correctly emits `Node.properties.next.$ref = #/definitions/Node`. A back-ref
> from a **property** to an ancestor definition is valid and must keep working.
> The bug is a **top-level** definition body that is a bare self-`$ref`.

---

## 2. Current machinery (verified survey)

### 2.1 Name derivation — always the short name

`internal/builders/schema/parsing_stuff.go:33` `inferNames()`:

```go
goName := s.Decl.Ident.Name      // short type name
s.GoName = goName
s.Name  = goName                 // definition key defaults to short name
model := s.findAnnotation(s.Decl.Comments, grammar.AnnModel)
if override, ok := model.AnnotationArg(); ok {
    s.Name = override            // `swagger:model <override>` — still unqualified
}
```

`internal/scanner/declaration.go:51` `EntityDecl.Names()` mirrors this for every
other caller (e.g. `MakeRef`). **The package path never enters the key.**

### 2.2 The registry — stored by `*ast.Ident`, looked up by (pkg,name)

`internal/scanner/index.go:74` — `TypeIndex.Models` / `ExtraModels` are
`map[*ast.Ident]*EntityDecl`. Distinct types from distinct packages have
distinct `*ast.Ident`, so they **coexist in the registry** (the registry itself
does not lose them). Lookups (`ScanCtx.GetModel(pkgPath, name)`,
`scan_context.go`) are already (pkgPath, name)-aware. So the identity is *not*
lost at storage — it is lost at **naming / emission**.

### 2.3 `$ref` emission

`internal/builders/common/builder.go:195` `MakeRef`:

```go
nm, _ := decl.Names()                              // short name
ref, _ := oaispec.NewRef("#/definitions/" + nm)    // unqualified ref target
prop.SetRef(ref)
s.AppendPostDecl(decl)                             // enqueue (dedup by *ast.Ident)
```

`AppendPostDecl` dedups by `decl.Ident`, so two same-named decls from different
packages **both** enqueue.

> **FRED**
>
> If the user has specified an explicit name for the model, e.g. swagger:model myName,
> that is the one used for the name of the $ref.

### 2.4 The discovery loop — where the merge happens

`internal/builders/spec/spec.go:97` `buildDiscovered()`:

```go
queued := make(map[string]struct{})       // keyed by definition NAME (string)
for _, d := range s.discovered {
    nm, _ := d.Names()
    if _, alreadyDone := s.definitions[nm]; alreadyDone { continue }
    if _, dupInPass := queued[nm];        dupInPass    { continue }  // <-- collision drop
    queued[nm] = struct{}{}
    queue = append(queue, d)
}
```

`s.definitions` is `map[string]oaispec.Schema` keyed by the short name. Two
distinct decls with the same `nm`:

- **#2783**: the second is silently dropped by the `dupInPass` / `alreadyDone`
  guard, but each reference site still wrote `$ref: #/definitions/Test`; whichever
  decl built the single `Test` entry wins → merge, order-dependent.
- **#2637**: the local `CreateDomainRequest` (a defined type) resolves its RHS
  (`mongo.CreateDomainRequest`) through `resolveRefOr` → `MakeRef` → a `$ref`
  whose target key equals **its own** definition key → the body becomes a
  self-`$ref`.

### 2.5 Cycle handling today

There is **no** dedicated cycle detector. Recursion is bounded only by the
discovery-loop's `alreadyDone` / `dupInPass` name guards (`spec.go:107-116`).
That suffices for legitimate recursion *because the name is stable*, but it is
exactly what conflates colliding names. (`embedDepth` exists but only feeds the
ambiguous-embed diagnostic, not cycle safety.)

**There is no real conflict detection anywhere — the short-name key is the
proxy.** The only acknowledgement is in the two seed tests' TODO comments.

---

## 3. Goals & invariants

- **G1 — distinct identities → distinct definitions.** Two different Go types
  must never silently merge.
- **G2 — no degenerate self-`$ref`.** A definition body must never be a bare
  `$ref` to itself.
- **G3 — legitimate recursion preserved.** Property/items/allOf back-refs to an
  ancestor definition keep working (`type Node struct{ Next *Node }`).
- **G4 — determinism.** Output must be stable across runs (today's merge is not).
- **G5 — zero churn in the common case.** When there is no collision, output
  (definition names) must be byte-identical to today. Definition names are part
  of the user's API contract; we do not rename gratuitously.
- **G6 — honor explicit intent.** An explicit `swagger:model <name>` is the
  user's chosen contract name; the design must respect it (and decide what an
  *explicit* collision means — see Q3).

---

## 4. Constraints / forces

- **Definition names are public.** They appear in the emitted OAS and drive
  downstream codegen (go-swagger client/server). Auto-inventing names is a
  user-visible act, not an internal detail.
- **Golden blast radius.** Any change to the *common-case* name would ripple
  through nearly every golden. G5 keeps that at zero; only collision fixtures
  should change. We must prove this.
- **`x-go-package` already exists** on definitions (a vendor extension carrying
  the source package path) — a natural disambiguation source if we want one.
- **Identity is already available** at the registry/decl level (pkgPath+name);
  the fix is about *using* it at naming/emission, not recovering it.

> **FRED**
>
> The memory of $ref's is already maintained by the spec document built by
> the spec.Builder discovery process.
>
> So we don't really need to maintain a separate index of $ref provided our
> construction mechanism doesn't clobber keys (i.e. produced names are unique).
>
> Since we have this possibility of user-injected names (swagger:model myType),
> this can't be however totally be ruled out. In that case, detect the attempt
> to clobber an existing $ref and issue a diagnostic (since this is author-land, we are fine
> with dropping duplicates and warn them of conflicting name overrides).
---

## 5. Design space (open — to debate)

Two semi-independent axes.

### Axis 1 — what to do on a name collision

- **Opt A · Always package-qualify.** Every definition keyed by a qualified name
  (e.g. `b.Test`, or a mangled `BTest`). ✗ violates G5 (massive churn, ugly
  names with no collision), ✗ surprises users. Rejected unless we find no
  alternative.

  > **FRED**. I don't currently see the situation where we need a mangled name. So this has to be clarified.
  >
  > Let's assume we have either: (i) a go type name unique in its fully qualified package name
  > or (ii) a named model from the "swagger:model {name}" annotation.
  >
  > In case (i) we may produce a $ref like:
  >
  > #/definitions/{fully qualified package name}/{name}
  >
  > At this stage, only uniqueness matters and we don't care (yet) about the readability of the full URL name.
  >
  > In case (ii) we produce a $ref like:
  >
  > #/definitions/{name}
  >
  > If we clobber an existing definitions at $ref creation time (e.g. same package, user entered duplicates): first
  > wins and other reuse the existing $ref with a diagnostic.
  >
  > NOTE: we should check that the model name is cleansed from the forbidden URL charset -> urlescape+diagnostic
  > I don't think this may happen for go identifiers, but it could with user-provided names.
  >
  > Given these assumptions, if we hit a collision, this means necessarily that we have detected a cycle in $ref
  > (this is possible in go with fields referencing a parent type, using pointers).
  >
  > In that case, the should not create a $ref but merely reuse it.

- **Opt B · Qualify only on collision.** Short name when unique; on collision,
  disambiguate the colliders (e.g. append package leaf / a counter). ✓ G5.
  Open: tie-break (who, if anyone, keeps the bare name?), the qualified-name
  scheme, ref-site resolution to the right variant, determinism of the scheme.

  > **FRED**. the discovery round (driven by the spec.Builder) can run gracefully
  > without producing duplicate $ref following my previous note.
  >
  > After the whole spec has been gathered, what we want is to reduce the $ref namespace
  > to a minimal, conflict-less tree.
  >
  > This part comes last and mutates the spec before the final rendering.

- **Opt C · Diagnose + hard-stop on collision.** Emit a located diagnostic
  ("two types named Test from packages b and c; disambiguate with
  `swagger:model <unique>`") and refuse to merge — either error out or skip the
  losers with a warning. ✓ safest, ✓ no invented names, ✓ pushes the contract
  decision to the user (whose contract it is). ✗ less "it just works".

  > **FRED**. see above
  >
  > No hard stop: ignore and diagnose, continue.

- **Opt D · Hybrid.** Auto-qualify *discovered/unannotated* collisions (Opt B
  for the implicit case) but treat two *explicit* `swagger:model <same>` as a
  user error (Opt C for the explicit case). Rationale: an explicit name is a
  deliberate contract claim; two of them colliding is genuinely ambiguous.

  > **FRED**. see above (agree).

### Axis 2 — the cyclic-`$ref` invariant (G2/G3)

Largely *falls out* of fixing Axis 1 (give `type X mongo.X` a key distinct from
`mongo.X` and the self-ref disappears). But we should still assert G2 explicitly
— a guard at `MakeRef` / the defined-type-alias path that refuses to emit a
top-level body that is a `$ref` to the decl's own key, converting it to either
the dissolved underlying or a diagnostic. Decide whether that guard is
belt-and-suspenders or load-bearing.

  > **FRED**. See above: if we guarantee uniqueness, detecting a conflict equates
  > finding a cycle.
  >
  > Exception: user-provided names may conflict for other reasons and that's why
  > there is a diagnostic (behavior in that case is like when we hit a cycle in the auto-detected type case (i)).

### Tentative lean (for debate, not decided) — ⚠️ superseded by §9

**Opt D** reads as the best fit for the goals: it preserves G5 (clean names in
the overwhelming common case), it never silently merges (G1/G4), it respects the
user's explicit contract (G6), and it turns the genuinely-ambiguous explicit
case into an actionable diagnostic rather than an invented name. The defined-type
self-ref (#2637) is then a special case of "implicit collision between the LHS
defined type and its same-named RHS" — auto-qualify or dissolve. But the
disambiguation *scheme* for the implicit case (Axis 1 / Opt B mechanics) is the
crux still to nail down.

  > **FRED**. see above (agree).

---

## 6. Open questions — ✅ all resolved by §9/§14

These were the pre-decision questions; the converged model answered every one
(none lingers — they were all downstream of decisions since made):

| Q | Resolution |
|---|------------|
| **Q1** disambiguation scheme + codegen-accept | §9.2 / D-7: PascalCase concat via `mangling.ToGoName`, score-budget (K1), hierarchical fail-safe; dotted rejected (#874). Codegen-accept = W7 (verbatim-today + enum precedent; live check deferred). |
| **Q2** does any colliders keep the bare name? | Dissolved: a node lifts to bare *only if globally unique*, so in a collision group **none** keeps bare — all qualify, symmetric, deterministic. No tie-break (§9.2). |
| **Q3** two explicit `swagger:model Test` | D-2/D-3/D-8: not a hard error, not first-wins — disambiguate + diagnose. |
| **Q4** `type X pkg.X` distinct vs dissolve | §9.1 / D-3/D-5: distinct `{fqpkg}` ⇒ distinct keys ⇒ two defs, no self-ref. (Alias-`=` dissolve is a pre-existing, orthogonal behavior.) |
| **Q5** centralise the namer | D-6 / §9.1: one build-side key-derivation helper (`{fqpkg}/{name}`) + `reduce` as the single namer. |
| **Q6** MakeRef needs the resolved name | Obsoleted: build emits the always-unique qualified key; `reduce` rewrites afterward — so we never "assign final names before emit." |

The only genuinely-open items are the small ones in §14 (budget score formula;
alias→type index mechanics, deferred) — none blocks P1.

---

## 7. Verification approach

- Reuse `fixtures/bugs/2783` (merge) and `fixtures/bugs/2637` (self-ref) as the
  two primary witnesses; flip their TODO assertions when the model lands.
- Add witnesses for: 3-way collision; explicit-vs-implicit collision; the
  legitimate-recursion guard (`Node{Next *Node}` still self-refs in a property);
  a no-collision control (proves G5 — identical output).
- **Determinism**: run the collision fixtures repeatedly (the 2783 test already
  flags non-determinism) — the fix must produce a stable golden.
- **Golden blast radius**: after the model lands, the only golden diffs must be
  in collision fixtures. Any common-case golden drift is a G5 violation to
  investigate (per `feedback_schema_discovery_verify_with_witness`).

---

## 8. Next step — ✅ done; see §9/§14

The debate landed a converged model (§9, decisions log §14, knobs §11). All of
§6's questions are resolved (table above). Next concrete step is **P1** (§12):
the build-side key swap + the `reduce` stage, gated on zero golden churn outside
the two witness fixtures.

---

## 9. Converged design (decision)

The pipeline moves from today's single lazy pass to **explicit phases**:

```
build (canonical identity-keyed)  →  [prune unreachable]  →  reduce (name ladder)  →  emit
```

Correctness lives in **build** (G1/G2/G3/G4); naming policy lives in **reduce**
(G5/G6); the **prune** seam sits between (a forthcoming feature that lands
gracefully here — see §9.4).

### 9.1 Build in canonical identity space

Every definition is keyed at build time by a **tree-addressed, compiler-unique
key**: `#/definitions/{fqpkg}/{name}`, **for every model regardless of origin**
(unified — no case-(i)/(ii) split). The leaf `{name}` is decided **at build
time** and baked into the key:

- `{name}` = the `swagger:model {name}` override when present, else the `{goName}`;
- **same-package duplicate** (a second *different* type in the *same* package
  claims an already-taken `{name}` — detectable because the exact `{fqpkg}/{name}`
  key already exists for a *different* decl): a definite user error → **fall back
  to `{goName}` for this one + diagnostic**.

`reduce` (§9.2) never sees how `{name}` was chosen — it is a pure,
name-source-agnostic minimizer over `{fqpkg}/{name}` paths. (The override
context is intentionally *not* carried forward; it's gone by reduce time.)

The `{fqpkg}` segment makes the key **compiler-unique by construction** → #2783's
silent merge and #2637's self-`$ref` are both impossible (distinct packages →
distinct keys; the local `type X pkg.X` keys `{here}/X` while its RHS keys
`{pkg}/X`). This unifies **cycle detection**: a key already present in
`definitions` for **the same decl** ⟺ a cycle (or a re-scan) → **reuse the
`$ref`, don't recreate**; present for a **different** decl in the same package ⟺
the user-error duplicate above. Legitimate recursion (a *property* `$ref` back to
an ancestor) is just the cycle/reuse case (G3). The single-document,
all-local-`#/definitions/` nature means none of `analysis.Flatten`'s remote-mesh
complexity applies. **The `Spec.definitions` map is the only memory needed** —
build just adds conflict-less `$ref`s into it; `reduce` reads the `$ref` graph
back via `analysis.New` (§9.5).

The only build-side change from today is swapping the key that `MakeRef` /
`EntityDecl.Names()` / the definitions map write — from the bare `goName` (the
current bug) to `{fqpkg}/{name}`. Everything downstream stays additive.

  > **FRED**. Regarding the reuse of the go-openapi/analysis library, I think that
  > the `New() *Analyzed` utility is the most useful. Its main benefit is that it walks and identifies all
  > $ref in a spec document and reports nicely about those.
  >
  > The `Flatten` utility should probably not be used here. It was primarily designed to flatten in a single document
  > references resolved from many remote documents. It is IMHO deeply flawed although
  > the heuristics work not so bad in practice (I am the author of that part and not particularly
  > proud of it, besides solving imperfectly a problem within given API constraints that made
  > it deeply untractable).
  >
  > NOTE: how to build a definition with multiple levels using go-openapi/spec serialization:
  > such as #/definitions/{path part 1}/.../{path part n}/{name}
  >
  > We create a definition:
  >  at level p (1 <= p <= n) the schema inside is defined by:
  >    x-go-package: "{path part 1}/.../{path p}"
  >    additionalProperties: true  # <- that creates a schema of any type (a map[string]any in go)
  >
  >  eventually we add the schema for the leaf node (type itself).
  >
  > NOTE: x-go-package emission is subject to some option being enabled.

### 9.2 Reduce: the candidate-set name ladder

`reduce` projects each surviving identity to the **shortest acceptable** user-facing
name. It consumes an **ordered candidate set per identity** and picks the
*cheapest candidate that is globally unique among the reachable set, is a valid
identifier, and sits within the budget*:

  > **FRED**. The $ref-reduce stage is carried out after the type discovery in the spec.Builder.
  > I don't think we really need to refactor the builder, just add this final stage.
  >
  > we would start by analyzing the spec document and start scanning the $ref's in it.
  >
  > Now we apply a ladder, that rewrites $ref with minimal depth under a uniqueness constraint.
  >
  > * we first simplify the tree: all non-conflicting node are lifted in the tree, until a conflict occurs
  >
  > * then we apply the "concat-with-parents" heuristics, under a budget constraint.
  >   We'll start with a default budget of 2 parts (tunable knob), meaning that
  >   if a conflict may be resolved by concatenating the name with its parent package name, then we pick that solution
  >   and lift the resulting nodes in the $ref tree.
  >
  > NOTE: we may use a slightly more complex score for the budget, accounting for the lengths of the parts in the name. To be discussed.
  > See below the proposed score formula.
  >
  > Caveat: at this point we cannot guarantee any longer that what the user specified in swagger:model will remain unchanged.
  > The user-specified name started with #/definitions/{name} but nothing guarantees us that no conflict may be discovered
  > from another package. To be checked further.

1. **bare `goName`** — when globally unique. Verbatim, byte-identical to today
   (definition names are emitted verbatim today — no mangling; see §2.1). This
   is the G5 common case.

  > **FRED**. Indeed. The above algorithm introduces no mangling, no alteration.

2. **PascalCase concat** of the minimal distinguishing prefix — candidate
   spellings drawn from BOTH the **canonical package leaf(s)** and the
   **author's import alias** (§9.3): `ClassificationBook` / `Go123Book`; deeper
   segments only if leaves also collide (`AMongoBook` / `BMongoBook`). Concat is
   identifier-safe and codegen-proven (go-swagger already uses PascalCase concat
   for enum-value Go names). **Dotted/slashed flat keys are rejected**: a `.`
   breaks the "definition name is a valid identifier" invariant just established
   by go-swagger#874 (`b475830`), and a nested `/` key round-trips through
   `go-openapi/spec` into `ExtraProps` (proven — see §10/W2).

  > **FRED**. ok
  > So here we may reuse the go-openapi/mangling.NameMangler:
  > `ToGoName({parent}+" "+{child})` will just do that: pascal case-concat.
  > _En passant_ we get a safe mangling of typical package separators like "-" or "."

3. **hierarchical/nested fail-safe** — the guaranteed fallback, reached only when
   the flat heuristics can't yield an *acceptable* (within-budget) name.

  > **FRED**. yes, when the concat heuristics is not selected (because it produces too long/too intricate name),
  > then we fallback to the definition hierarchy. However unusual, this rule produces
  > a readable model.
  > We might want to generate an info diagnostic to tell the user they can alter this by forcing swagger:model {name}.

_Definition of "simple"_ (Fred): a collision group is simple when the cheapest
rung that yields a globally-unique valid identifier sits within the budget
(segment count / length — see §11/K1). Crucially, flat concat can never fail on
correctness — concatenating the full path is always a unique valid identifier
(just long). So the hierarchical form is a _readability_ fail-safe, not a
correctness one; "simple" = "the heuristics produce an acceptable flat name
within budget."

### 9.3 Author-driven candidate: import aliases

The concat heuristic's candidate pool is enriched with the **author's import
alias for the target's package, taken in the context of the *referencing* type**
(`mongo-go-driver/mongo` imported as `mongodb` → `MongodbBook`, more meaningful
than `MongoBook`). This is *intent-for-a-spelling*, not an arbitrary
winner-among-equals priority rule — a deliberately different (defensible) kind of
choice.

Mechanically it rides the reachability graph: each `$ref` is an **edge**
(source-context → target-identity); the alias lives on the source side
(`ast.File.Imports`: `importPath → localName`, nil = default leaf). Annotate
each edge with the alias used at its source file (captured at reference-recording
time — `MakeRef` has both the referencing decl's `*ast.File` and the target
identity), then aggregate per target identity into the candidate set. Within a
rung, **prefer the author alias over the canonical leaf** when both fit (intent
first). See §10/W3 for the many-to-one rule.

  > **FRED**. yes. The idea here is to use the package alias authored by the user
  > instead of the canonical package name.
  >
  > This would allow to try the "concat-with-parent" rule with 2 different input
  > prioritizing the user-defined alias.
  >
  > However, mapping package aliases (a pure AST thing) with types is a bit complex
  > so we may defer this improvement a bit (we have an example of this kind of alias index building
  > in our codegen at go-openapi/testify).

**Fallback technique for per-edge context (Fred's hack — if the AST→alias index
proves fiddly).** Instead of a separate edge index, stash the per-reference
context (e.g. the author's import alias) on the referring node as a transient
extension `x-codescan-{runid}-*` at build time; `reduce` reads it back, then
strips it. `Extensions` is a `map[string]any`, so it can carry anything; the
random `{runid}` avoids clobbering a real user `x-codescan-*`.

> ⚠️ **Corrected caveat (probed on spec v0.22.6):** contrary to the original
> "a `$ref`'s siblings don't marshal" assumption, extensions on a `$ref` schema
> **DO serialize** in the pinned `go-openapi/spec` (it preserves `$ref` siblings,
> per modern JSON-Schema). So **cleanup before emit is mandatory, not free** —
> left in, `x-codescan-*` leaks into the output. The natural cleanup point is
> `reduce`'s final `$ref` rewrite pass (same `analysis.New` index); golden tests
> are a backstop (a leak shows as drift). This is a *fallback*; the primary
> design needs no build→reduce context channel at all (reduce works purely on
> `{fqpkg}/{name}` paths).

### 9.4 Prune seam (forthcoming feature, designed-for)

`prune unreachable` slots between build and reduce. Pruning **before** reduce is
not just fewer entries — it yields **better names**: a *dead* definition must not
tax a *live* one with a qualified name. If `b.Test` is unreferenced and only
`c.Test` is used, pruning `b.Test` first lets `c.Test` keep the **bare `Test`**.
So the ladder's "globally unique?" test asks the right denominator — the
**reachable** set. The reachability graph (§9.3 edges) is the same structure the
`analysis.New(spec)` ref-index gives reduce, so prune and reduce share
infrastructure. (Prune's *root set* depends on `ScanModels` — see §10/W5; that's
prune's concern, our design just exposes the seam.)

  > **FRED**. yes. The "prune unreachable" step, when implemented will prune the $ref tree
  > just before the deconflicting.

### 9.5 Pass-2 mechanics & x-go-package

Reduce rewrites the spec with `analysis.New(spec)` as the **ref-index** (the
convenient enumerator of every `$ref`-bearing node — NOT `Flatten`): compute the
identity→name map, then rewrite ref strings + re-key `definitions`. Bounded,
deterministic, single-document. `x-go-package` stays **on the definition** (as
today) carrying provenance for free; namespace-node `x-go-package` is only
relevant inside the hierarchical fail-safe.

  > **FRED**. See above (confirmed)

### 9.6 In-object dedup scope (1/2/3) — keep + diagnose

Orthogonal to model identity, but in scope for this track (every silent drop
gets a diagnostic). These stay **local to a builder's subtree** — they are NOT
the model-identity registry (which is new and lives in the orchestrator). See
§0/§2 baseline. **These are decoupled, standalone tasks — see §13 so they don't
get lost behind the engine work.**

- **(1) parameters** / **(2) response headers** — the **local-duplicate**
  diagnostic (NOT `$ref`/identity related). Keep the existing per-builder `seen`
  dedup (Swagger 2.0 forbids duplicate parameters / header names); emit a
  **diagnostic** whenever a duplicate is found and dropped — today the drop is
  **silent**. (Tracked as §13/ST1.)
- **(3) embedded-promotion fields** — the Go promotion rule appears honoured by
  `diagnoseAmbiguousEmbed` + `nameByJSON`; **prove it with a dedicated golden**
  and add an **informational** diagnostic on the ambiguous-promotion drop.
  (Tracked as §13/ST2.)

  > **FRED**. yes and this is independent from the $ref handling algorithm

---

## 10. Wrinkles to preserve (surfaced in the debate)

- **W1 — re-qualification on a newly-reachable collision (not "instability").**
  Per Fred: the algorithm introduces **no instability** — output is a
  deterministic pure function of the reachable set (W6). The only user-visible
  effect is that adding a *genuinely-referenced* model that collides with an
  existing one re-qualifies the previously-bare name (`Book` →
  `ClassificationBook`) — a real, deterministic consequence of a real new
  conflict, not non-determinism. The one annoyance worth designing against is a
  collision forced by a definition the user *doesn't care about* — which is
  exactly why **prune-unreachable matters** (§9.4): it removes dead colliders
  before reduce so they never tax a live name. Reframed from "accepted
  instability" to "deterministic, prune-mitigated."

- **W2 — hierarchical fail-safe representation caveat.** A nested
  `#/definitions/classification/Book` is a deep JSON pointer, not a flat key.
  Proven via probe against `go-openapi/spec`: it round-trips into `ExtraProps`,
  so `Definitions` ends up keyed `[classification, go123]` (container schemas);
  typed `ResolveRef` **fails**, only `ExpandSpec` resolves it. `additionalProperties: true`
  on the container covers validation leniency. Acceptable **only** because the
  hierarchical form is a rare fail-safe; a definitions-*enumerating* consumer
  (go-swagger codegen generating one model per entry) would see the container
  nodes. (This is a reason the flat rungs are preferred whenever they fit.)

- **W3 — alias many-to-one.** One identity referenced from multiple files with
  different aliases (or default + alias). Rule: **offer the alias candidate only
  when unambiguous across all ref edges** (all agree, or a single ref site); on
  disagreement, drop it and fall back to the canonical leaf. Aliases *enrich* the
  pool when authors are consistent; they never *force* an arbitrary pick.

  > **FRED** yes not so easy, uh? We can make a lot of progress before we reach that point. Defer (but don't forget).

- **W4 — concat can introduce a NEW collision.** A heuristic concat
  (`ClassificationBook`) might collide with a real existing type of that name.
  The ladder must test each candidate for global uniqueness over the reachable
  set and deepen (or fall back) on a hit — never emit a concat that itself
  collides.

  > **FRED** ah right. If it does, then the concat pick is ineligible.

- **W5 — prune root-set depends on `ScanModels`.** `ScanModels` on → annotated
  models are roots (kept even if unreferenced); off → only ref-reachable survive
  (`#2639` territory). Prune defines roots; our design only exposes the seam.

- **W6 — determinism.** Names must be a **pure function of the reachable identity
  set (+ unambiguous aliases)** — never of discovery / map-iteration order
  (today's #2783 non-determinism). The reduce pass runs after the set is complete,
  which guarantees this.
- **W7 — codegen tolerance.** Moot for the flat rungs: definition names are
  emitted verbatim today and consumers already mangle Go identifiers; concat
  stays a valid identifier — and we mint it via `mangling.NameMangler.ToGoName(parent+" "+child)`,
  which also makes package separators (`-`, `.`) safe for free. The `swagger`
  binary is not available in-sandbox, so a live `go-swagger generate model` check
  on a collision spec is deferred to when one is available.
- **W8 — URL-escaping of definition keys.** Likely already handled by
  `go-openapi/spec` under the hood (Fred). Treat as **verify, not build**; user
  names that aren't URL-safe get a diagnostic. (Go identifiers are always safe;
  only `swagger:model` overrides could carry odd chars.)

---

## 11. Knobs (resolved in review round 1)

- **K1 — budget = a tunable score, not a fixed part-count.** ✅ **LANDED**
  (2026-06-17). The score balances **total length, #parts and longest-segment
  length**, each normalized to [0,1] and weighted (0.50 / 0.30 / 0.20), so every
  criterion pushes the score up as it grows; 4+ parts is always 1.0. Picked
  empirically from a side-by-side harness (seed ratio-form vs. linear vs.
  weighted — see `builders/spec/concat_score_test.go`, disabled). Production:
  `concatScore` in `reduce.go`. The threshold is `Options.NameConcatBudget`
  (zero value → built-in default **0.65**); beyond it → hierarchical fail-safe
  (Stage 3). NB: in-session this score/budget pass was called "K2" informally —
  not to be confused with the K2 *knob* below (hierarchical-defer).
- **K2 — hierarchical fail-safe: deferred.** ✅ Deferred — full-path concat is
  always a correct flat option, so the early phases ship bare→concat and avoid
  the caveat-laden nested representation (W2) up front. Hierarchy stays the
  designed fail-safe (buildable via §9.3's `additionalProperties:true` +
  `x-go-package` recipe), built later.
- **K3 — alias source: deferred.** ✅ Build the candidate-set abstraction early;
  wire the import-alias source later (mapping AST aliases→types is fiddly).
  Technical references for the deferred work:
  `go-openapi/testify/codegen/internal/scanner/import.go` (an import-alias index
  built this exact way) and the skill note
  `go-openapi/testify/.claude/skills/.local-skills/ast-types-bridging.md`
  (position-based `go/ast`↔`go/types` bridging — same family as codescan's
  existing `FindASTField`/`FileForPos`). Low risk either way.
- **K4 — within-rung preference.** ✅ Author alias > canonical leaf when both fit
  (intent first).
- **K5 — hierarchical representation (if/when built).** Nested +
  `additionalProperties:true` containers + per-level `x-go-package`
  (option-gated) per §9.3. Defer with the feature.

---

## 12. Execution plan (one merge; each stage gated by goldens)

**Delivery shape.** This is complex but lands as **one merge** of
`feat/name-identity-cyclic-ref` into `fix/backlog-lot1` (not the per-issue
squash cadence). Stages below are **internal checkpoints within the branch**,
each gated by a **golden** checkpoint, so progress is verifiable and reviewable
stage-by-stage; only the final, all-green branch merges.

**Merge trailer set (the whole name-conflict family closes at once).** The
triager grouped six issues in this family — two active fixes plus four marked
♻️ duplicate-of-this-track in the ledger. They are all resolved by this one
engine, so the **merge** must carry a `contributes` trailer for each (closure
happens when go-swagger bumps the dep):

```
* contributes go-swagger/go-swagger#2783   (models mixed across packages)
* contributes go-swagger/go-swagger#2637   (self-cyclic $ref, same-named type)
* contributes go-swagger/go-swagger#2662   (same ref names -> invalid spec)
* contributes go-swagger/go-swagger#1734   (same name service/deps -> nondeterministic)
* contributes go-swagger/go-swagger#2126   (same-named models, examples not mixed)
* contributes go-swagger/go-swagger#2398   (warn on duplicate definitions)
```

Coverage map lives in the corpus test header
(`internal/integration/coverage_name_identity_test.go`). #2762 / #2791 — also on
the original `fix/go-swagger-2783` working branch — are NOT in this family (empty
not-applicable / works-as-designed verdicts) and stay in the backlog stream.

**The golden discipline (applies at every stage).** The collision/witness
fixtures are the moving target; **everything else must stay byte-identical**
(G5). So each stage's golden *diff* is exactly the intended change and nothing
else — `UPDATE_GOLDEN=1 go test ./...` then inspect: any churn outside the
in-scope fixtures is a regression to chase (per
`feedback_schema_discovery_verify_with_witness`). Plus, at each stage: full
`go test ./...` green, determinism (collision fixtures stable across repeated
runs — G4), lint, markdown.

### Stage 0 — prep & verification surface (no engine change) — ✅ DONE (2026-06-15)

- **Impl-prep:** read `analysis.New` (the `$ref` index we'll consume) and *how*
  `analysis.Flatten` re-points referencing nodes when it mutates a `$ref` — the
  rewrite **mechanics only**, ignoring its naming heuristics. That's the
  technique `reduce` needs.
- **Fixture corpus + golden harness.** Branch already carries `bugs/2783` +
  `bugs/2637`. Add: a **3-way** collision; a **legitimate-recursion**
  (`Node{Next *Node}`); an **explicit+implicit** mix (one `swagger:model`
  collision + one auto-detected collision); a **same-package-dup `swagger:model`**
  (D-4); a **no-collision control**; a **pathological-depth** case (same leaf,
  different parent — reserved for Stage 3). Convert the witnesses to
  golden-captured (`CompareOrDumpJSON`), capturing **current (buggy)** output as
  the baseline.
- **Gate:** suite green; goldens lock today's behavior (the bugs are *in* the
  baseline goldens, so Stage 1's diff visibly *is* the fix).

#### Stage 0 — outcome & findings (record)

Corpus landed under `fixtures/enhancements/name-identity-*` (6 members) + one
coverage file `internal/integration/coverage_name_identity_test.go`; full suite
green, lint clean, zero churn to existing files (all additions). Members:
`name-identity-{3way,recursion,mixed,same-pkg-dup,no-collision,deep}`.

- **F-0a — `analysis` rewrite helpers are UNEXPORTED.** `UpdateRef` /
  `RewriteSchemaToRef` live in `analysis/internal/flatten/replace` — not
  callable from codescan. Only the **enumeration** surface is public
  (`analysis.New(spec) *Spec` → `AllRefs()/AllReferences()/AllDefinitionReferences()`,
  in-memory, no loader, single-doc OK). It tracks **JSON-pointer strings**, not
  parent schema pointers. ⟹ **`reduce`'s ref-rewrite + definition re-key must be
  reimplemented locally regardless.** **Stage-1 decision point:** since the
  rewrite is in-house either way and we own the single document, *lean toward a
  small in-house `$ref` walker* (no new `analysis` dependency; walks
  Definitions/Paths/Responses/Parameters yielding **parent `*spec.Ref` pointers**
  for in-place mutation — which `analysis` does *not* give us) rather than pulling
  `analysis` just for string enumeration. Settle at Stage 1 start.
- **F-0b — all 4 collision fixtures are NON-DETERMINISTIC today** (empirically
  verified, repeated runs: which package's `x-go-package`/`required`/fields win
  varies). Confirms G4 is a real bug, and means baseline collisions **cannot be
  golden-captured** — the controls (recursion, no-collision) are golden-locked;
  the collision fixtures pin only order-independent facts (definition count +
  colliding key). They flip to deterministic goldens when the engine lands.
- **F-0c — implicit vs explicit merge differ today.** Explicit
  `swagger:model` collisions **union** fields (`Item`→`[x1 y1]`); implicit
  (referenced, unannotated) collisions take a **single winner** (`Record`→`[rx]`
  only). Both are wrong; both resolve to distinct defs post-fix.

### Stage 1 — engine core: qualified build + lift-to-bare reduce ⟵ the big one

- **Build:** one key-derivation helper → `{fqpkg}/{name}` (name = `swagger:model`
  override else `goName`; same-package-dup → `goName` fallback + diagnostic),
  used by `MakeRef` / `Names()` / the discovery dedup. Cycle =
  key-present-for-same-decl → **reuse the `$ref`** (D-5).
- **Reduce v1** (new final stage in `spec.Builder`, after discovery): build the
  `$ref` index (`analysis.New`); **lift every globally-unique node to its bare
  leaf**; collision groups stay at full qualified depth for now (no concat yet);
  rewrite `$ref`s + re-key `definitions`.
- **Gate (the milestone):** full corpus **zero churn** (common case lifts back to
  today's bare names → G5 proven); witnesses `2783`/`2637`/3-way flip
  buggy→**distinct** (deep-qualified) defs — merge and self-`$ref` both gone;
  recursion fixture **unchanged** (property back-`$ref` preserved → G3);
  determinism stable. Correctness is *done* here; names are just ugly.

### Stage 2 — concat rung (minimal-depth, pretty names)

- **Reduce v2:** on collision, `mangling.NameMangler.ToGoName(parent+" "+leaf)`
  concat under the K1 score budget; W4-safe (a concat that itself collides is
  ineligible → deepen). Emit the collision **diagnostic** (D-8).
- **Cyclic-within-collision coverage (carried over from Stage 1).** Stage 1
  proved legitimate cycles only on the *unique-name → bare* path
  (`name-identity-recursion`). Once reduce v2 actually rewrites collision keys,
  add a `name-identity-recursion-collision` witness — two packages each with a
  self-recursive (and/or mutually-recursive) type that **also collides** on its
  short name — and verify the cycle's back-`$ref` is rewritten **in lockstep**
  with its renamed key (no dangling / no stale deep ref). This is the regression
  guard for "reduce rewrites a ref that points into the very group being
  renamed."
- **Gate:** **only** the collision-fixture goldens change (deep → `BTest`/`CTest`
  …); corpus still zero churn; diagnostics asserted; the cyclic-collision
  witness renders a valid, self-consistent cycle.

#### Stage 2 — outcome (landed 2026-06-16)

All gate criteria met: full suite + `-race` (integration/scanner) + lint + gofmt
clean; **only collision goldens changed** since `1ef2f25` (the `no_collision` /
`recursion` controls are byte-identical → still G5/G3). Code in
`builders/spec/reduce.go`: `computeNameReductions` now does pass-1 (unique→bare,
reserve names) + pass-2 (`resolveCollisionGroup` minimal-depth concat). The
ladder tries leaf+1, leaf+2, … nearest parent segments, accepting the shallowest
depth where the whole group is mutually distinct AND free of the global `taken`
set (W4); full-path + numeric-suffix is the pathological backstop. Names via
`mangling.NewNameMangler().ToGoName(join(segs," "))` (leading-digit segment
`2637` → valid `X2637…`). `diagnoseCollision` emits one positionless
`validate.colliding-model-name` (D-8) per group listing pkg→name.
**No `x-go-name` is added on concat** — setting it to the shared leaf would
re-collide the generated Go types downstream (deliberate).

Corpus results: 2783→`BTest`/`CTest`; 3way→`A/B/CWidget`; mixed→`X/YItem`,
`X/YRecord`; deep→`AMongoBook`/`BMongoBook` (d=2, since `MongoBook` collides at
d=1); 2637→`X2637CreateDomainRequest` + `MongoCreateDomainRequest` (local body
`$ref` → mongo, not self). New witness `name-identity-recursion-collision`
(`PNode`/`QNode`) proves a self-`$ref` into the renamed group is rewritten in
lockstep (`PNode.next`→`#/definitions/PNode`). Deterministic across runs (sorted
group + key iteration).

#### Stage 2.5 — readability score & budget option (K1) — ✅ LANDED (2026-06-17)

The §11/K1 tunable score is now built (it was "deepen until unique" before).

- **`concatScore(concat, parts) float64`** in `reduce.go` — the weighted blend
  (total len / #parts / longest segment; weights 0.50/0.30/0.20; 4+ parts → 1.0;
  clamped to [0,1]). Chosen empirically against a `seed` (ratio form, kept as a
  cautionary baseline — its `/maxWord²` term anti-correlates with intuition) and
  a `linear` baseline in the disabled harness `concat_score_test.go`.
- **`Options.NameConcatBudget float64`** (public, via `scanner.Options`) — the
  caller-tunable threshold; zero value → built-in `defaultNameConcatBudget`
  (**0.65**). Plumbed through `ScanCtx.NameConcatBudget()`.
- **Seam, not yet behavior:** `resolveCollisionGroup` now returns the group's
  worst score; `Builder.noteBudget` debug-logs when a group exceeds the budget
  ("hierarchical fallback (Stage 3) would apply here"). Names are unchanged — the
  budget governs nothing observable until Stage 3 wires the fallback. Goldens
  therefore unchanged.
- **Tests:** `TestConcatScore` + `TestConcatScoreVersusDefaultBudget` pin the
  function's contract and its separation around 0.65; the empirical harness stays
  in-tree but `t.SkipNow()`-disabled for future re-challenge.

### Stage 3 — hierarchical fail-safe (the "K3" of the session shorthand) — ✅ LANDED (2026-06-17)

The decision (deferred at the Stage-2 K2 knob) was taken as **include**. Full-path
concat remains the always-correct flat fallback, so this is purely a readability
upgrade for the over-budget tail, gated behind an opt-in.

**Representation (probe-confirmed against `go-openapi/spec`).** A nested
`#/definitions/<pkg>/<Name>` marshals with the container as a top-level
`Definitions` entry (`additionalProperties:true` + `x-go-package`) and the model
under its `ExtraProps`; the model keeps its own `x-go-package`. Typed `ResolveRef`
fails (only `ExpandSpec` resolves it) — exactly W2 — which is why this is opt-in
and default-off (a definitions-enumerating consumer like go-swagger codegen would
see the container nodes).

**Implementation (`reduce.go`).** The `noteBudget` seam became a branch in
`computeNameReductions`: when `score > budget` AND `EmitHierarchicalNames`,
`resolveHierarchicalGroup` picks the nested path (the same nearest segments,
raw-nested not ToGoName-joined; shallowest depth where paths are distinct and no
root shadows a flat name; `containerRoots` lets two groups share a container).
The slash-valued rename rides the existing ref-rewrite (`repointer` is already
string-based); `rekeyDefinitions` skips slash values; `placeHierarchical` builds
the container chain and nests the model (merging into shared containers).
`diagnoseHierarchical` emits `validate.hierarchical-model-name`.

**Knobs added (public via `scanner.Options`):** `EmitHierarchicalNames bool`
(default false). `NameConcatBudget` (Stage 2.5) governs the trigger.

**Gate met:** new fixture `name-identity-hierarchical` (two long-named packages,
both `swagger:model Config`; flat concat scores 1.0 > 0.65). Two goldens —
`_flat` (default, `RecommendationengineConfig`/`NotificationserviceConfig`) and
`_nested` (opt-in, `#/definitions/recommendationengine/Config`). **Every
pre-existing golden byte-identical** (default-off proven). Unit tests
`TestResolveHierarchicalGroup` (depth/taken/shareable) + the score contract
tests. Full suite + `-race` + lint clean.

**Open refinements (not blockers):** diagnostic severity is currently `Warn`
(could be `Hint`/info — see plan note); per-level `x-go-package` is stamped on the
innermost container only (deeper levels get `additionalProperties:true` but no
provenance) — revisit if a multi-level fixture warrants it.

### Stage 4 — in-object diagnostics (independent; §13)

- **ST1:** param/header local-duplicate diagnostic + the `(name,in)` dedup-key
  check/fix + fixtures. **ST2:** embedded-promotion Go-rule golden + informational
  diagnostic. Each gated by its own golden / diagnostic assertion. Can land any
  time (not engine-coupled).

### Out of this merge

> **MERGED 2026-06-17** into `fix/backlog-lot1` (Stages 0–3 + cross-package leaf
> resolution for type-name keywords + golden-test migration). The deferred
> advanced features below are now logged in
> [forthcoming-features.md §14.1](./forthcoming-features.md) (import-alias
> candidates / W3 many-to-one, ST3 qualified author refs, K5 per-level
> `x-go-package`).

- **Prune-unreachable** — separate forthcoming feature (#2639); `reduce` only
  exposes the seam (§9.4). Lands later, just before `reduce`. Tracked in
  forthcoming-features.md §12.
- **Author-alias candidates** (K3) — deferred; a Stage 5 if time allows,
  otherwise post-merge (the candidate-set abstraction built in Stage 2 leaves the
  door open). See forthcoming-features.md §14.1.

---

## 12.1 Stage 1 implementation design (concrete — derived from the Stage-0 code survey)

Settled from the build-side survey (`spec.go`, `common/builder.go`,
`schema/{schema,parsing_stuff}.go`, `scanner/declaration.go`) + the `analysis`
know-how study. **F-0a decided: in-house walker, no `analysis` dependency**
(rewrite is in-house regardless; an in-house walker also gives us the parent
`*spec.Ref` pointers that `analysis` withholds).

### A. Build in canonical identity space — one identity key, 4 sites

Add **`EntityDecl.DefKey() string`** = `d.Obj().Pkg().Path() + "/" + name` where
`name` is `Names()`'s first return (the `swagger:model` override else the
goName). This is the single source of truth for the build-time definition key —
used by **both** the ref-emitter and the def-store so a ref string can never
disagree with a definitions key. The 4 sites:

1. **`common/builder.go:216` `MakeRef`** — ref target becomes
   `"#/definitions/" + decl.DefKey()`.
2. **`schema/schema.go:75,88` `Build`** — key the definitions map by
   `s.Decl.DefKey()` instead of `s.Name`. **Keep `s.Name` as the leaf** — it
   still drives `annotateSchema`'s `x-go-name` (`s.Name != s.GoName`) and
   `x-go-package`, so **extensions do not churn** (the overload trap avoided).
3. **`spec.go:109-116` `buildDiscovered`** — dedup + already-built check key by
   `d.DefKey()`.
4. (consistency) the schema lookup at `schema.go:75` (allOf/embic accumulation)
   uses the same `DefKey()`.

After build, the whole intermediate spec is fq-keyed and self-consistent. This
**structurally fixes both witnesses**: #2783's three `Widget`s get distinct keys
(no merge); #2637's local `CreateDomainRequest` keys `…/bug2637/CreateDomainRequest`
while its RHS keys `…/mongo/CreateDomainRequest` → the body ref points elsewhere,
**no self-`$ref`** (G2). Legitimate recursion (G3) is unchanged: `Node` refs its
**own** `DefKey` (same decl) → the existing dedup reuses it → property back-ref
preserved. Cycle = key-present-for-same-decl ⟹ reuse (D-5) is exactly today's
dedup, now over `DefKey`.

### B. Author-short-name readers (blast radius — must not regress G5)

Some sites resolve a definition by the **author's short name**, which no longer
matches the fq keys mid-build:

- **`routes/walker.go:510`** `r.definitions[decl.ResponseRef]` (the
  definition-fallback promotion; `routes-responses-definition-fallback` golden
  exercises it). Add a `resolveByLeaf(defs, short) (fqKey, ok, ambiguous)`:
  unique leaf → its fq key (golden unchanged); none → dangling (today's
  diagnostic); **multiple → ambiguous diagnostic + drop** (new, D-8 style).
- **Audit during coding:** the operations builder (swagger:operation inline-YAML
  `$ref: #/definitions/Short` author refs) and any `resolvers` short-name
  lookups. The corpus doesn't use inline-YAML refs; existing goldens are the
  backstop.

### C. Reduce v1 — new final stage (`spec.Builder.reduceDefinitionNames`)

Inserted as the **last** step in `Build()` — **after `buildMeta` (spec.go:88),
before `return s.input` (line 94)** — because `buildRoutes`/`buildOperations`
(lines 78–84) also emit definition refs via `MakeRef`. (The Stage-0 agent's
"after `buildDiscovered`" placement is too early.)

Algorithm (v1 = lift-unique-to-bare only; **collisions stay fq, no concat yet**):

1. **walk** every schema-bearing node, collecting each `$ref` with an in-place
   setter (see D). Group definition keys by **leaf** (substring after last `/`).
2. for each **unique-leaf** group: `renameDefinition(fqKey → leaf)` — re-key the
   `Definitions` map **and** rewrite every ref equal to `#/definitions/<fqKey>` to
   `#/definitions/<leaf>`.
3. **collision** groups (leaf count > 1): **leave the fq slash-key untouched**
   (deep, ugly, but distinct + deterministic — §12 Stage 1). Stage 2 concats.
4. **determinism (G4/W6):** iterate keys **sorted**; the rename is a pure
   function of the key set.

Consequences on the corpus: controls (`no-collision`, `recursion`) lift every
key back to bare ⟹ **byte-identical to today's goldens (G5/G3 proven)**;
collisions (`3way`/`mixed`/`same-pkg-dup`/`deep`, #2783, #2637) become **distinct
deep-keyed defs** ⟹ merge + self-ref gone, deterministic ⟹ flip their facts-only
baseline tests to the fixed counts (and they gain goldens once deterministic).

### D. In-house walker (know-how from `analysis`, mutating by pointer)

`forEachSchemaRef(sw, func(setRef func(spec.Ref), cur spec.Ref))` visits, with
nil-guards: `Definitions[*]`; top-level `Responses[*].Schema` /
`Parameters[*].Schema`; per-path × per-verb `op.Parameters[i].Schema` and
`op.Responses.{StatusCodeResponses[*],Default}.Schema`; recursing each schema's
`Properties`, `AllOf/AnyOf/OneOf`, `Items.{Schema,Schemas[]}` (SchemaOrArray —
both arms), `AdditionalProperties.Schema` (SchemaOrBool nil-guard), `Not`,
nested `Definitions`/`PatternProperties` (completeness). Quirks adopted from
`analyzer.go`/`replace.go`: the SchemaOrArray/SchemaOrBool nil-guards; the
single-vs-tuple items split; we mutate the **pointer in place** (cleaner than
`replace.go`'s value-vs-pointer type-switch). We **skip**: param/response/
pathItem self-`.Ref` (codescan never points those at `#/definitions`); and
**JSON-pointer escaping** — we match on **ref-string equality** and rewrite,
never navigate pointers, so `~0/~1` escaping never arises (Go map keys with `/`
are fine; nothing resolves/marshals mid-build).

### E. Stage 1 gate

Controls byte-identical (G5/G3); every collision fixture + both witnesses flip to
**distinct** defs (deep-keyed) with **merge & self-`$ref` gone**; determinism
stable across repeated runs (G4); full `go test ./...` green; lint clean;
**zero golden churn outside collision/witness fixtures** (the discovery-fallback
golden stays green via `resolveByLeaf`). Correctness is *done* at Stage 1; names
are deliberately ugly until Stage 2.

### F. Open implementation audits (verify while coding, not blockers)

- `oaispec.NewRef("#/definitions/" + fqKey)` must accept slash/dot-laden
  fragments — quick probe at the start of coding.
- Confirm no other short-name definitions reader beyond `walker.go:510`.
- `buildRoutes` does **not** drain `s.discovered` (unlike params/responses) —
  pre-existing; confirm Stage 1 introduces no new dangling ref.

### G. Stage 1 — outcome (landed 2026-06-16)

All gate criteria met: full `go test ./...` green, `-race` clean on every touched
package (incl. the reduce path via integration), lint clean. Code:

- `scanner/declaration.go` — `DefKey()` (= `pkgpath/name`), `SuppressModelOverride()`/
  `ModelOverrideSuppressed()`, `Names()` honours suppression.
- `builders/common/builder.go` `MakeRef` — emits `#/definitions/<DefKey>`.
- `builders/schema/{schema,parsing_stuff}.go` — map keyed by `DefKey`; `s.Name`
  kept as leaf (no `x-go-name` churn); override skipped when suppressed.
- `builders/spec/spec.go` — `resolveSamePackageDuplicates()` pre-pass (D-4,
  deterministic, diagnostic) before build; `reduceDefinitionNames()` as final step.
- `builders/spec/reduce.go` (new) — in-house ref walker + lift-unique-to-bare +
  re-key, no `analysis` dependency (F-0a).
- `builders/routes/walker.go` — `resolveDefinitionByLeaf` for the author-short-name
  definition-fallback (unique → promote, none → dangling, **ambiguous → diagnose+drop**).
- `parsers/grammar/diagnostic.go` — `CodeDuplicateModelName`.

Results: #2783 three `Widget`s → 3 distinct defs (was 1 merged, non-deterministic);
#2637 local→`$ref` to the mongo def, **no self-`$ref`** (G2); 3way/mixed/deep
distinct; D-4 → `Dup`(First)+`Second`(fallback)+`Root`. **G5 proven**: the
true-integration (codescan.Run) goldens are byte-identical — 0 of 29 modified
goldens are integration-consumed; the modified ones are all sub-builder UNIT
goldens now reflecting the fq intermediate contract (Fred's call), plus the
`scantest.ResolveTestKey` retrieval helper. Collision fixtures golden-captured
(deterministic now). Stage 2 will replace the deep keys with PascalCase-concat.

Audit results: `NewRef` accepts slash/dot fragments (probed ✓); `walker.go:510`
was the only short-name definitions reader (operations/`resolveBodySchema` ride
the unique-leaf-bares mechanism); `buildRoutes` non-drain introduced no dangling
ref (suite green).

---

## 13. Independent side-tasks (decoupled from the deconfliction engine)

These need none of the two-phase machinery and could be done/merged on their own
at any time. Recorded as first-class tasks so the "easy, therefore forgotten"
trap is avoided.

### ST1 — local duplicate param/header diagnostic (NOT `$ref`)

**Problem.** Within ONE `swagger:parameters` / `swagger:response`, two fields can
resolve to the same parameter / header and the **second silently overwrites the
first** — no diagnostic. This is a *local* (intra-object) duplicate, entirely
separate from the cross-package `$ref`/model-name deconfliction that is this
track's main subject.

**Where.** `parameters.processParamField` (`seen[name]` overwrite,
`parameters.go:443`) and `responses.processResponseField` (`seen[name]=true`,
`responses.go:569`). Both per-builder, per-subtree `seen` maps (see §2/baseline).

**Task.**

1. On a duplicate, emit a **diagnostic** (located at the losing field) before the
   drop/merge — "duplicate parameter %q (in=%q) — keeping the first / last".
2. **Verify the dedup key.** Swagger 2.0 parameter uniqueness is the **(name,
   `in`)** pair, but `seen` is keyed by **name alone** — confirm whether a
   query `id` and a path `id` wrongly collide today; if so, re-key on (name, in)
   as part of this task (and add a witness). Response headers are unique by name,
   so the header `seen` key is fine.
3. Witness fixtures: a swagger:parameters with two same-(name,in) fields → one
   param + diagnostic; a same-name/different-in pair → two params, no diagnostic.

Low effort, high "don't-forget" value. Independent of P0–P3.

### ST2 — embedded-promotion ambiguity golden + diagnostic

Covered in §9.6(3): prove the Go promotion rule with a dedicated golden and add
an informational diagnostic on the ambiguous-promotion drop
(`diagnoseAmbiguousEmbed`). Rides with P2 but is otherwise standalone.

### ST3 — qualified-ref pinpointing for author short refs (deferred; overlaps K3)

**Problem.** Author-written model refs in route annotations (`200: Test`,
`body:Test`, definition-fallback) resolve through `routes.resolveDefinitionByLeaf`
by **pure leaf** — exact match on the segment after the last `/` of the
fully-qualified key. So when two packages both declare `Test`, the short ref is
**ambiguous → diagnosed + dropped (D-8)** and the author has **no inline way to
pinpoint** which one; the only lever is to rename a type with
`swagger:model <uniqueName>` so its leaf is unique.

**Enhancement.** Let the author write a package-qualified hint and match by
**segment-suffix** instead of pure leaf, narrowing (often resolving) the
ambiguous set. Three coordinated pieces:

1. **Matcher** (`resolveDefinitionByLeaf`) — accept a qualifier and match keys
   ending in `…/<qualifier>/<name>` (exact full-key still wins; pure leaf stays
   the fallback).
2. **`routes.resolveBodySchema`** — today it blindly prepends `#/definitions/`;
   it must mint a ref the reduce stage keeps consistent for a qualified hint.
3. **Author-facing syntax** — decide `pkg.Type` (Go-style, but `.` is otherwise
   an ordinary token) vs `pkg/Type` (mirrors the internal key). Multi-segment?

**Why deferred / where it belongs.** Overlaps the import-alias work (§9.3 / K3):
the same AST import-alias index that lets the concat heuristic prefer
`mongodb.Foo` is what would let an author *hint* `mongodb.Foo` at a ref site.
Out of the Stage 1–4 merge; revisit with K3. Pure-leaf + `swagger:model` remains
the disambiguation contract until then.

---

## 14. Decisions log — review round 1 (resolved)

The crux forks, and how the debate settled them (the inline `> FRED` notes in
§2/§4/§5/§9–§11 are the working record):

- **D-1 Unified key, no case split.** Every definition keys at build time as
  `#/definitions/{fqpkg}/{name}`, regardless of origin. `{name}` = `swagger:model`
  override else `goName`. (Earlier case-(i)/case-(ii) split dropped — it failed
  to fix the witnesses; see D-3.) §9.1.
- **D-2 `swagger:model` is a preference baked at build, not carried to reduce.**
  The override picks the leaf `{name}` at build time; `reduce` is a
  name-source-agnostic minimizer over `{fqpkg}/{name}` and never sees the
  annotation. Consequence (accepted): an explicit name may be re-qualified by
  reduce on a genuine cross-package collision — "preference, not guarantee."
- **D-3 Explicit collisions disambiguate (option a).** Two explicit
  `swagger:model {same}` across packages → distinct keys (distinct `{fqpkg}`) →
  reduce disambiguates + diagnoses. Both witnesses (#2783, #2637) thus resolve to
  **distinct definitions** (matching the seeded tests), not first-wins/merge.
- **D-4 Same-package duplicate `{name}` = user error.** A *different* type in the
  *same* package claiming an already-taken `{name}` → drop the override, fall
  back to `goName`, diagnose. §9.1.
- **D-5 Cycle = key-found-for-same-decl ⟹ reuse.** Uniform cycle detection falls
  out of unique keys; no separate self-`$ref` guard needed (G2/G3). §9.1.
- **D-6 No orchestrator refactor.** One build-side key swap + a new `reduce`
  final stage in `spec.Builder`; the `Spec.definitions` map is the only memory.
  §12.
- **D-7 Tooling.** `analysis.New` for the `$ref` index (NOT `Flatten`);
  `mangling.NameMangler.ToGoName` for the concat (separator-safe); hierarchical
  fail-safe via `additionalProperties:true` + per-level `x-go-package`
  (option-gated, the existing extension — `x-go-package-name` was a typo).
- **D-8 No hard stop, ever.** Every collision/duplicate/odd-name path is
  ignore-fix-deterministically + diagnose + continue.

**Still genuinely open (small, none blocking P1):** the exact budget *score*
formula (K1 — balances #parts/total-length vs leaf length; try several). The
`x-go-package` emission is settled (single, option-gated). The alias→type index
mechanics are deferred (K3 references).
