# Cross-ref linkage — source ↔ spec navigation (genspec-tui)

> 🟢 **Status: design settled — ready to build.** All open questions resolved
> (§8); grounded in a recon of the builder pipeline (§9). No code yet. Successor
> to the `naiveLinker` placeholder; the rebase-unlocked half of the genspec-tui
> roadmap in [`wasm-playground.md`](wasm-playground.md) (the other half — the
> diagnostics pane — landed in `b2574fe`). First buildable milestone: `LX-spec-0`
> (spec-side index, not rebase-gated).

Related: [[project_genspec_tui]] · [[project_lsp_diagnostics_target]] ·
[`wasm-playground.md`](wasm-playground.md) §"Position-backed cross-ref linker" ·
[`anonymizer-repro-tool.md`](anonymizer-repro-tool.md) §12.6 ·
[`forthcoming-features.md`](forthcoming-features.md) §3.1.

---

## TL;DR

The cross-ref linker lets the user stand on a Go annotation and jump to the spec
node it produced — and back. It has **two independent halves that meet at a JSON
pointer**:

- **Source side (codescan, greenfield):** record `(jsonPointer → token.Position)`
  at each **anchor** node — those born from a code detail (type decl, struct
  field, method, const/var, the `swagger:meta` block) — with finer nodes
  resolving to their nearest anchored ancestor (it is *not* a bijection). This is
  the half the grammar2 rebase unlocked: positions already ride on every
  `grammar.Block`/`Property` and every `EntityDecl`; today they are simply dropped
  for valid nodes. Where the resulting cross-ref **index lives** is the open
  question (§3.6): caller-owned via a callback now, vs a library-owned derived
  model later. The emission instrumentation is identical either way.
- **Spec side (TUI, NOT rebase-gated):** an index from *rendered-spec line →
  JSON pointer*, built while rendering the spec pane. Already half-needed for
  remark anchoring ([`anonymizer-repro-tool.md`](anonymizer-repro-tool.md)
  §12.6), so building it buys down two features at once.

The linker joins them: **spec line → pointer → source position** (jump editor to
that line) and **source line → enclosing pointer → spec line** (scroll + flash
the spec pane). Bidirectional, position-backed, replacing the name-match
`naiveLinker`.

---

## 1. What we're replacing

`cmd/genspec-tui/internal/ux/linkage.go` already defines the seam — keep the
shape, swap the implementation:

```go
type SourceLinker interface { Targets(sel Selection) []SpecTarget }
type SpecTarget struct { Anchor string; Line int } // Line 0 = locate by Anchor
```

The current `naiveLinker` parses a `.go` file, lists its exported type names, and
hands them back as text anchors located by first occurrence in the rendered
spec. Its own doc comment names the problems: **name matching is ambiguous** (a
type and a property can share a name), **format-dependent** (anchor strings
shift with JSON vs YAML), and **one-directional** (source → spec only, by name).
The position-backed linker fixes all three by keying on pointers + positions
instead of names + substring search.

---

## 2. The pivot: JSON pointer as the join key

Both halves speak RFC 6901 JSON pointers (`/` → `~1`, `~` → `~0`):

| Spec node | Pointer form | Example |
|-----------|--------------|---------|
| Definition | `/definitions/{name}` | `/definitions/User` |
| Property | `/definitions/{def}/properties/{prop}` | `/definitions/User/properties/email` |
| Response (top-level) | `/responses/{name}` | `/responses/ErrorResponse` |
| Path + method | `/paths/{path~esc}/{method}` | `/paths/~1pets~1{id}/get` |
| Parameter | `/paths/{path~esc}/{method}/parameters/{i}` | `…/get/parameters/0` |
| Enum value | `…/enum/{i}` (on the enclosing schema/property) | `/definitions/Grade/enum/2` |

The pointer is the contract between the two halves: the source side maps
`pointer → position`, the spec side maps `line ↔ pointer`. Neither half needs to
know the other's internals.

---

## 3. Source side — the provenance seam (codescan)

### 3.1 API shape — `OnProvenance`, mirroring `OnDiagnostic`

`OnDiagnostic` is the template (`internal/scanner/options.go:57`,
`scan_context.go:152`, `common/builder.go:94`). Add a sibling:

```go
// Provenance ties a spec node (by RFC 6901 JSON pointer) to the source
// position of the Go construct that produced it.
type Provenance struct {
    Pointer string         // e.g. "/definitions/User/properties/email"
    Pos     token.Position // file:line:col of the producing decl/field/annotation
}

// On Options (= scanner.Options, already public via codescan.Options):
//
//   OnProvenance, when non-nil, is invoked once per emitted spec node, in build
//   order. Experimental, like OnDiagnostic — the surface will firm up as LSP
//   integration matures.
OnProvenance func(Provenance)
```

Plumbing: `ScanCtx.OnProvenance()` accessor + `common.Builder.RecordProvenance`
that nil-guards and fires the callback. Every per-decl builder embeds
`*common.Builder`, so the emit verb is available everywhere a node is inserted.

> **Decision point — where does `Provenance` live?** `grammar.Diagnostic` lives
> in `grammar` because the parser raises it. Provenance is a *builder/spec*
> concept (pointer + stdlib `token.Position`), not a grammar one. Candidates:
> (a) a small neutral package (e.g. `codescan` root, re-exported like `Options`);
> (b) `internal/scanner` alongside `Options`; (c) `grammar` for symmetry with
> `Diagnostic`. Leaning (a)/(b) — keeping it out of `grammar` avoids implying the
> parser owns spec pointers.

> **FRED**:
>
> I agree that the mapping between spec and source does pertain to the grammar or diagnostic.
>
> Enabling this mapping should be an option at the top level (so scanner.Option) and the TUI
> would always enable this (but probably not a CLI like go-swagger).
>
> The right place is indeed the builder. Example of a possible future feature that would validate
> this choice: suppose we want to annotate the spec with "x-go-origin: fixtures/codescan.go#L25" extensions.
>
> It should be reminded that the linkage is not a bijection:
>
> * several spec entries may refer to the same go decl (e.g. with $ref)
> * all keys in the spec do not necessarily have a direct decl, but they necessarily have a parent that has one.
>
> Hence OnProvenance is not necessarily called on _all_ nodes, but on nodes initiated by code details (decl, fields, methods, const values, etc).
>
> Example: when parsing the "swagger:meta" tag at the package level, many json nodes are produces
>
> However I am not really sure about this callback mechanism in the scanner. Perhaps it is good for now, perhaps we can do better as a target design.
>
> What would do the callback?
> => maintain an external map of (jsonpointer node) => (code ref)
>
> So we may do that at this stage as being the _less intrusive way_ of maintaining this map (completely left to
> the calling application to do whatever they like with a node reference.
>
> I am thinking of leveraging another stream of evolution: the "internal model" (insulates go-openapi/spec  dependency).
>
> A very simple such "internal model" could be something like:
>
> internal/model/spec.go:
> ```go
>
> type Spec struct {
>   *oaispec.Spec
>   provenances map[string]string
> }
> ```
>
> And the map maintenance would be directly wired by the "Typable" implementations. To be challenged, but worth considering IMHO

> **↳ reply (synthesis).** Three of your points collapse into one reframe worth
> pulling to the front:
>
> 1. **Builder is the home; enablement is a top-level `scanner.Options` flag**
>    (TUI always on, go-swagger CLI off). Settled. (I read your first line as
>    "does *not* pertain to grammar/diagnostic" — agreed; the parser shouldn't
>    own spec pointers.)
> 2. **It's sparse, not per-node.** We record only at *anchor* nodes — those born
>    from a code detail (type decl, struct field, method, const/var value, the
>    `swagger:meta` block, the `swagger:route`/`swagger:operation` annotation
>    block). Everything else (`info.title`, an `items` sub-schema,
>    a `$ref` site) resolves to its **nearest anchored ancestor**. This one rule
>    is your "closest parent that has one," and it dissolves the §3.2 cost worry
>    *and* the not-a-bijection worry simultaneously. The `swagger:meta` example
>    is the proof: one annotation → one anchor (the meta block) → many spec nodes
>    that all resolve upward to it.
> 3. **Instrumentation is durable; the sink is swappable.** Whatever sink we pick
>    — callback / internal-model map / `x-go-origin` — the *call sites* asserting
>    "this node is anchored at this pos" are identical (one `RecordOrigin(ptr,
>    pos)` in `common.Builder`). So the sink choice is **not blocking**:
>    instrument once, start with the thinnest sink, migrate later without
>    re-touching builders. Your "not sure the callback is the target design" is
>    fine — it doesn't have to be. See the new §3.6.
>
> On `x-go-origin`: noted as your **future feature for other CLIs** (e.g.
> go-swagger), *not* the TUI path and not for now. Its value here is as an
> argument that the index deserves a more permanent home than "lives in the TUI"
> — folded into the real open question (§3.6: *where does the index live?*), not
> treated as an immediate sink.

### 3.2 Cost — moot once emission is anchor-only

The original framing here ("`OnProvenance` fires for every node, thousands of
times") was **wrong** — corrected by the anchors-only reframe (§3.1). We emit at
decls/fields/methods/const-values/meta-blocks only: dozens to low-hundreds, not
thousands. There is no hot per-node path left to guard.

What remains is trivial: a single centralized check in `RecordOrigin` (one
nil/enabled test, not scattered across sites). Where a site must build a pointer
string before recording, gate on `b.OriginEnabled()` or pass a `func() string`
thunk so the string is assembled only when recording is on — your closure point.

### 3.3 Emission sites (from recon §9)

| Node | Site | Pointer + position in scope? |
|------|------|------------------------------|
| Definition | `spec.go:71` (`s.definitions[s.Name] = schema`) | ✅ `decl.Names()` + `decl.Spec.Pos()` |
| Property | `schema/fields.go:111` (`target.Properties[c.name] = ps`) | ✅ `s.Name` + `c.name` + `c.afld.Pos()` / `block.Pos()` |
| Response | `responses/responses.go:66` | ✅ `r.Decl.ResponseNames()` + `r.Decl.Spec.Pos()` |
| Path/op | `routes.go:62` / `operations.go:57` | ⚠️ path/method known; position is the grammar block (synthetic) |

### 3.4 Data model — sparse anchors + nearest-ancestor resolution

> **FRED** not all nodes are worth a pointer, the rendering may look for the closest parent that has one.
>
> Since this is essentially a hierarchical model, perhaps the index [pointer key] => [code reference]
> is not optimal and that we should build something with a "synthetic" pointer with only elements stored and the references
> stored along the path (e.g. akin to a trie structure).

> **↳ reply.** Agreed — and this retires the old "pointer-stack through the
> schema recursion" cost I'd flagged as the hard part. Since only **anchors**
> emit, and anchors are decls/fields (not deep sub-schemas), the deep
> `allOf`/`items` nesting mostly *doesn't* anchor — it resolves upward. No need
> to thread a pointer prefix through the recursion at all for the MVP.

The index is **not** a dense `pointer → pos` map. It is a sparse set of anchored
pointers; any query pointer resolves to the **deepest anchored ancestor** on its
path. Two equivalent shapes:

- **Flat map + ancestor walk (MVP):** `map[pointer]ref`; lookup trims pointer
  segments right-to-left until a hit. Trivial to build from the emission stream;
  fine at TUI scale.
- **Trie (target):** segment-keyed tree with `ref` stored sparsely on nodes;
  lookup descends the query pointer and returns the last `ref` seen along the
  path. This is your "synthetic pointer, references stored along the path" — the
  natural structure, and the one the internal-model / `x-go-origin` sinks
  (§3.6) carry anyway.

Both answer the same query. Start flat; move to the trie when the sink moves to
the internal model.

**The reverse (source → spec) — two distinct queries, don't conflate them** (this
corrects an earlier "$ref is free" claim — it isn't):

- **"this decl → its own spec node"** is direct and one-to-one. The field/decl is
  itself an anchor, so its pointer is recorded outright. A `$ref`-valued property
  (`User.Manager *Person`) is anchored at the **field**, pointing at the field's
  position — *not* at `Person`'s decl. No ref-following involved.
- **"this type → everywhere it is referenced"** (LSP-style find-references) is the
  one-to-many case, and it is **not free**. A referencing site is anchored to its
  own field, not to the target decl, so answering it means **resolving `$ref`s at
  render time** — which pulls in JSON-Schema `$ref` quirks (`$ref` inside
  `allOf`, ref-to-ref chains, ignored sibling keywords, external/file refs). A
  `$ref` is a node-level decision; treating it as "the same code node as the ref
  target" only holds if we declare ref-resolution a TUI render-time step and
  accept those caveats. **Scope it as a separate, later feature** — the core
  bidirectional loop does not need it.

> **FRED** yes references could be multiple. Good catch. New naviagation key binding to go to the next/previous
> location (e.g. F3 / Shift-F3)

### 3.5 Position granularity caveats

- **Decl/field nodes** → exact (`token.Position` from the AST). 
- **Routes/operations** → the grammar block's `Pos()` reflects the *reassembled*
  synthetic comment group, not the precise annotation column. Good enough to land
  on the right line; document the coarseness.
- **Inside-YAML positions** (operation bodies, `extensions:`) → coarse (block
  start) until the token-level YAML position work lands
  ([`forthcoming-features.md`](forthcoming-features.md) §3.1, goccy/go-yaml).
  The provenance seam should emit the *best available* position and improve
  automatically when that swap happens.

> Mmh yes position may be inaccurate now. For our TUI, we just need the line.
> This remains a bit flaky for YAML blocks but the "start line of the YAML block" is good enough.
>
> Accurate position at the token level is only useful for future LSP.

> **↳ reply.** Agreed — line-granularity is the contract for now; the YAML-block
> start line is the documented floor. `Pos` stays `token.Position` so column
> precision arrives for free when the goccy swap lands, but nothing depends on it.

### 3.6 Where does the index live? (the question I skipped)

The emission/instrumentation is settled (§3.1–3.4): at each anchor, *something*
learns `(pointer → pos)`. The open question — the one I omitted — is **who owns
and holds the resulting cross-ref index** (the compact trie/prefix structure over
pointer-path parts). Because the call sites are identical either way, this is not
blocking and not a one-way door.

**(1) Caller-owned, fed by a callback — least change, ship the TUI on it.**
`OnProvenance` streams the pairs; the **TUI** builds and holds the trie (your
"completely left to the calling application"). Pros: zero new types in the
returned spec; the library stays a plain `*spec.Swagger` producer; LSP and other
consumers each shape the index to taste. Cons: **ephemeral** (rebuilt every
session) and every consumer re-implements the trie + nearest-ancestor logic.

**(2) Library-owned, in a derived model proxying `spec.Spec` — your proposal.**
A derived structure (`internal/model`: `Spec{ *oaispec.Spec; index }`) carries
the index home, maintained directly by the `Typable` implementations. Pros:
**permanent** — any consumer gets the index without rebuilding; the trie +
ancestor logic live once, in the library; rides the **go-openapi/spec-insulation
/ internal-model** stream you're already contemplating. Cons: needs that
internal-model stream to exist first; widens the return surface (opt-in, so normal
users still get a plain spec); **more change than the TUI feature alone
justifies** — the deterrent you named.

**The far horizon (separate axis, not a near choice): persist provenance into the
spec document** as `x-go-origin` extensions — your future feature for *other CLIs*
(go-swagger). This is the most permanent form (provenance survives serialization,
no side channel) and the strongest argument that the index shouldn't only live in
the TUI. But it's its own feature with its own UX cost (it's visible noise in a
shipped spec) and is **out of scope here** — noted only because it pulls the
ownership question toward (2) rather than (1).

**DECIDED (2026-06-03).** Ship on **(1) caller-owned**: the `OnProvenance`
callback **stands and remains**; the TUI builds and holds the index. It composes
forward — when the internal-model stream lands, the same callback populates that
model and the index structure lifts out of the TUI essentially unchanged, so (1)
is not throwaway. The middle-path "library-provided builder helper" is **not**
taken: the structure is small enough (§3.7) that we **build our own, tailored to
our needs, with no external dependency** rather than add an exported codescan
surface or swallow a trie lib. (2) arrives *with* the internal-model work, not
ahead of it.

### 3.7 Recording mechanics — the build is not always forward-moving

Fred's key constraint: spec nodes are **not** written once and left alone. They
get rewritten and reset mid-build. Grounded in the code:

- `schema.go:102` — `defer s.annotateSchema(schema)()` runs at **function exit**,
  *after* the whole drilldown. The schema is in its settled shape at that point.
- `schema/walker.go:140` — `*ps = oaispec.Schema{…}` rewrites the pointed-to
  schema wholesale; `schema/fields.go:89` — `ps = oaispec.Schema{}` resets a
  property.
- `schema/simpleschema.go` — on an OAS-v2 violation the target is **reset to
  empty `{}`** (`CodeUnsupportedInSimpleSchema`), *after* it was built.

Two consequences for `RecordOrigin`:

1. **Fire from the settled point, not at raw insert.** Co-locate the definition
   record with the `annotateSchema` deferred hook (post-drilldown), and the
   property record *after* the field carrier is applied (final JSON name known).
   This sidesteps the moving-target problem: embedded-promotion and renames have
   already resolved, so the **pointer is final** when we record. (Pointer
   identity is what churns; a reset only blanks the node's *content*.)
2. **The index is upsert / last-wins, and resets do NOT clear origin.** A node
   reset to `{}` still came from its field — navigation should still land there.
   So the code origin is **stable across content rewrites**; re-fires for the
   same pointer are idempotent (last-wins on `pos`, which rarely changes). The
   index never deletes an origin in response to a content reset.

Net: record once, late (settled pointer); tolerate re-fires; never evict on
reset. This keeps the structure trivial (see below) despite the churn.

**Structure — DECIDED: flat map now, (radix) trie deferred to `LX-model`.**
A JSON-pointer hierarchy *is* a prefix tree and nearest-ancestor *is* the
canonical trie query — so a trie is the textbook-correct and right *target*
structure. But it optimizes the wrong variable for our scale. The workload:
`N` = anchors = dozens–low hundreds (a few thousand in the docker-API stress
case); depth ≤ ~6–8; queries = upsert + exact + nearest-ancestor. **At this scale
Big-O is moot** (a walk is microseconds either way), so choose for simplicity,
memory, and fit.

- **Now: flat `map[pointer]ref` + right-to-left segment-trim walk.**
  ```go
  for p := query; p != ""; p = p[:strings.LastIndexByte(p, '/')] {
      if r, ok := m[p]; ok { return r } // deepest ancestor first
  }
  ```
  ~15 lines, no new types; upsert is `m[ptr]=ref` (last-wins — matches the §3.7
  churn requirement); the trims are **zero-alloc** (`p[:i]` shares backing);
  memory is tens of KB even storing full pointer strings.

- **A trie earns its keep only when** (a) **key redundancy bites** — the full
  pointer strings repeat their shared prefixes (`/definitions/User/properties/…`
  over and over), and on a big tree that wasted memory exceeds a trie's per-node
  overhead (below that crossover a trie node costs *more* than the bytes it
  saves); (b) we need **subtree queries** ("every spec line under
  `/definitions/User`") — out of scope for the jump; or (c) the index moves into
  the **library** (`LX-model`) and is queried at scale by many consumers. At that
  point use a **radix/PATRICIA** variant (collapses single-child chains), not a
  plain trie.
- **Cheaper intermediate before a full trie**, if memory bites first: **intern
  the pointer segments** (a shared string pool / store pointers as `[]segmentID`)
  — kills the prefix duplication without the structural rewrite. Reach for it only
  if a profile shows the map's keys dominating; otherwise it's premature.

**The reverse index is a separate structure** — not the pointer hierarchy: a
slice sorted by `(file, line)` + binary search. We store a *point*
`token.Position`, not a span, so "nearest enclosing" = *the greatest anchor
start-line ≤ cursor line, in the same file* (the most recent decl/field above the
cursor). Good enough to land the jump; the UX shouldn't imply exact span
containment.

### 3.8 Corner case — the background spec (`InputSpec` overlay)

`codescan.Options.InputSpec` seeds the builders with a starting spec they add to
(not yet a TUI CLI flag). Those pre-existing nodes are **born from the spec, not
from code** — they have **no origin**, and may have **no anchored ancestor**
either (a whole definition can come from the overlay).

So the linker must treat "no origin found" as a **first-class outcome**, not a
bug: nearest-ancestor resolution can legitimately return *nothing*. The TUI shows
such a node as non-navigable with an honest label (e.g. *"from input spec — no
source"*) rather than mis-jumping. This also means the spec→source direction is
**partial by design** once an overlay is in play — worth stating up front so the
UX doesn't promise a jump for every line.

---

## 4. Spec side — rendered-line → pointer index (TUI, not rebase-gated)

The spec pane renders `*spec.Swagger` via `json.MarshalIndent` (ordered keys) /
YAML. To know "what node is the cursor on," build a **line → pointer** map *over
the rendered bytes*. Re-parsing into `map[string]any` loses key order, so the
robust approach is a **forward scan that tracks a pointer stack as it walks the
rendered text**:

- **JSON — `jsontext.Decoder.StackPointer` (Fred's call, adopted).** Decode the
  rendered bytes with a `jsontext.Decoder`; at each value read `StackPointer()`
  (the RFC 6901 pointer, with escaping handled for us) + `InputOffset()` (byte
  position), then derive the line by counting `\n` up to that offset. This beats a
  hand-rolled tokenizer: correct pointer escaping (`/`, `~`, unicode, escaped
  quotes) is the stdlib's job, not ours.

> **FRED** how about using https://pkg.go.dev/encoding/json/jsontext@go1.26.4#Decoder.StackPointer ?
>
> However, Decoder.InputOffset tells us about the position in the JSON but not the line (needed for proper positioning on the panel).
> Lines may however be infered by counting '\n' in the buffer up to that offset.

> **↳ reply.** Adopted — and it removes the one bug-prone part of the scan. Two
> mechanics to pin: (1) decode **the same bytes we render** (don't re-marshal —
> `spec.Swagger`'s ordered-key `MarshalJSON` stays the source of truth, the
> decoder just reads what's on screen); (2) `InputOffset → line` via a one-time
> line-start table (binary-search offset→line), O(n) per render, negligible.
> **Caveat + decision (2026-06-03).** `encoding/json/jsontext` is still
> `GOEXPERIMENT=jsonv2`-gated even in go1.26.3 (not default-on stdlib). **Decided
> (Fred):** ship it anyway — official binary / WASM builds are produced from
> go1.26 with `GOEXPERIMENT=jsonv2` (we control the release pipeline); users who
> build the TUI themselves are **warned** they need that toolchain+flag. The
> library module is untouched (its 2-version window stands); only `cmd/genspec-tui`
> carries the requirement. Pure Go → WASM-safe.

- **YAML:** `yaml.Node` carries `.Line`; walk the node tree to the same map.

This index is **independent of the source side** and can land first. It is the
same structure remark-anchoring needs, so it pays for two features
([`anonymizer-repro-tool.md`](anonymizer-repro-tool.md) §12.6). Rebuild it on
every render (cheap; the pane already re-renders on format toggle / rescan).

---

## 5. Joining the halves — the bidirectional linker

Replace `naiveLinker` with a `positionLinker` holding:

- `pos   map[string]token.Position` — pointer → source position (from `OnProvenance`).
- `srcIx []pointerAtPos` — same data sorted by `(file, line)` for range lookup.
- `line2ptr map[int]string` / `ptr2line map[string]int` — spec-pane index (§4).

Flows:

- **spec → source** (cursor in spec pane, line `L`): `line2ptr[L]` → `pos[ptr]`
  → open `Pos.Filename` in the file viewer at `Pos.Line`.
- **source → spec** (cursor in editor at `file:L`): find the pointer whose
  position is the **nearest enclosing** `(file, line)` in `srcIx` → `ptr2line[ptr]`
  → scroll the spec pane and flash that line.

Ambiguity (a position covering several nodes, or a pointer rendered at multiple
lines) reuses the existing `[]SpecTarget` cycle semantics — collect candidates,
let the user step through.

---

## 6. Interaction model (TUI)

Assumes both indexes are live in the TUI: forward (spec line → pointer → origin)
and reverse (source `(file,line)` → nearest-ancestor pointer → spec line).

### 6.1 The model — one symmetric "link-nav mode", driver + follower

Fred's proposal, generalized: a **link-navigation mode** toggled with **`f`
(follow)** in the focused pane. The focused pane is the **driver** (a moving
line-cursor); the opposite pane is a **live follower** that re-centers and
highlights the linked target on every cursor move. `ESC` exits back to the pane's
normal behavior (scroll for the spec viewport, editing for the source pane).

- The follower gets **visual** focus only (center + highlight) — keyboard input
  stays in the driver. "Move focus" = scroll the target line to **vertical
  center** (best-effort; clamp at top/bottom edges) and highlight it.
- Tracking is **live** (updates as you scroll the driver), **debounced**, and
  only re-opens a source file when the resolved path actually changes.

### 6.2 Spec → source (driver = spec pane)  — Fred's §1

- **enter:** `f` activates the line-cursor in the spec viewport (current line
  highlighted; up/down move it).
- on each move: resolve cursor line → pointer → **nearest-ancestor** origin; the
  left pane opens that file and centers+highlights the originating line.
- **exit:** `ESC` — the spec viewport returns to plain scrolling; the left pane no
  longer tracks.

### 6.3 Source → spec (driver = source pane)  — Fred's §2

- operates from the **read-only file view** (not mid-edit — see 6.6); `f`
  highlights the current code line.
- on each move: resolve `(file,line)` → **first matching parent** anchor (the
  nearest enclosing decl/field/value that produced a node) → center+highlight its
  spec node in the right pane.
- **exit:** `ESC` returns the source pane to normal editing.

### 6.4 Edge cases (honest by design)

- **No source** (background-spec / `InputSpec` node, §3.8): the driver line maps
  to a node with no origin. Don't jump or clear the follower — show a status note
  (*"no source — from input spec"*) and hold. Navigation is *partial* whenever an
  overlay is in play; the UX must not imply every line jumps.
- **Multi-candidate:** the direct reverse (decl → its own node) is 1:1, so the
  primary jump lands on the single target. When a resolution yields *several*
  spec locations — the find-references case (a type `$ref`'d from many sites,
  §3.4 / `LX-refs`) — **`F3` / `Shift-F3`** step to the next/previous location.
- **Dirty buffer:** reverse positions are valid **as of the last scan**. With
  unsaved edits the line cursor can drift from the recorded `token.Position`
  (highlight off by N lines) until save → rescan refreshes the index. State this;
  consider suppressing reverse-nav while dirty, or showing a "stale" badge.
- **Format toggle (JSON↔YAML):** the spec-side index is per-render. On toggle in
  nav mode, preserve the *pointer* under the cursor and re-locate its line in the
  new format, rather than keeping the raw line number.

### 6.5 Visual language

- **Driver line** vs **follower target** get *distinct* highlights (e.g. driver =
  strong reverse-video bar; follower = secondary tint) so it's clear which pane
  leads.
- **Status line** shows the mode badge (`SPEC-NAV` / `CODE-NAV`) and the resolved
  target (`→ models/user.go:42` or `→ no source`).
- Optional **gutter dots** on lines that *have* a link, for discoverability (which
  lines are navigable) — cheap once the indexes exist.

### 6.6 Decisions (2026-06-03)

1. **Mode key = `f` (follow).** `n`/`N` stay as global search next/prev-match
   (`model.go:274/279`); `f` is unbound and mnemonic. `f` toggles link-nav in the
   focused pane; `ESC` exits.
2. **Enter from the viewer.** Link-nav operates from the **read-only file view**;
   from active `textarea` editing the user `ESC`s to the viewer first (no mid-edit
   chord). Bare-key commands are unambiguous there.
3. **Auto-follow (live), debounced.** Every cursor move re-navigates the follower;
   debounce + reopen-only-on-path-change keep it from thrashing on fast scroll.

4. **Find-references nav = `F3` / `Shift-F3`** (next/prev location) when a
   resolution has multiple targets (§6.4); distinct from `f` link-nav.

These close §6. New bindings to add to the `key` enum + dispatch: `f` (toggle
link-nav), `F3`/`Shift-F3` (cycle multi-ref); a `navMode`/driver-pane bit + a
candidate-cursor on the model; follower center+highlight helpers.

---

## 7. Milestones

| ID | Scope | Side | Gated? |
|----|-------|------|--------|
| **LX-spec-0** | rendered-line ↔ pointer index (JSON, then YAML) | TUI | no — can land first |
| **LX-prov-0** | anchor emission via `RecordOrigin`, sink = caller callback; anchor set per §8 (resolved); callback-collection test (mirror the diagnostics test) | codescan | no |
| **LX-join-0** | `positionLinker` replacing `naiveLinker`; spec → source jump (nearest-ancestor resolve) | TUI | needs prov-0 + spec-0 |
| **LX-join-1** | source → spec direct reverse (decl → its own node, §3.4) | TUI | needs join-0 |
| **LX-model** | move the index into a library-owned derived model (§3.6 option 2) | codescan | follow-up; rides the internal-model stream |
| **LX-refs** | find-references (render-time `$ref` resolution, §3.4) + `F3`/`Shift-F3` cycling | TUI | in scope; after the core loop |
| **LX-prov-2** | precise inside-YAML positions | codescan | gated on goccy swap (§3.5) |

Suggested order: **LX-spec-0 → LX-prov-0 → LX-join-0 → LX-join-1**. spec-0 and
prov-0 are independent and could be built in parallel. `LX-model` arrives *with*
the internal-model stream rather than dragging it forward. Two milestones from
the first draft are **gone**: the pointer-stack/deep-nesting one (retired by
anchors-only, §3.4) and the `x-go-origin` sink (reframed as a parked future
feature, §3.6).

---

## 8. Open questions (for Fred)

**Resolved** (folded into §3):
- ~~`Provenance` type home~~ → builder/scanner, **not** grammar (§3.1.1).
- ~~Emit policy / per-node cost~~ → **anchors only**, nearest-ancestor (§3.1.2,
  §3.2, §3.4). "Fires for every node" retracted.
- ~~Pointer-stack through recursion~~ → unnecessary (anchors resolve upward, §3.4).
- ~~Where the index lives~~ → **caller-owned callback, TUI builds/holds a
  build-our-own structure, no external dep**; library-owned derived model arrives
  with the internal-model stream (§3.6 DECIDED).
- ~~Recording timing under rewrites/resets~~ → record late from the settled
  point; upsert/last-wins; never evict on reset (§3.7).
- ~~The anchor set~~ → type decl · struct field (properties + parameters +
  response fields) · method · const/var value · **enum value** (`…/enum/{i}`,
  anchored to its const — we scan values anyway) · `swagger:meta` block ·
  `swagger:route`/`swagger:operation` block. `RecordOrigin` call sites pinned.
- ~~Index structure~~ → **flat `map[pointer]ref` + segment-trim walk now**;
  (radix) trie deferred to `LX-model` / subtree queries (§3.7). Reverse index is
  a separate `(file,line)`-sorted slice.
- ~~Find-references scope~~ → **in scope** (`LX-refs`): references can be multiple
  and must be navigable with **`F3` / `Shift-F3`** (next/prev location); needs
  render-time `$ref` resolution (§3.4 caveats).
- ~~Spec-side index~~ → **forward-scan via `jsontext.Decoder.StackPointer`** (§4),
  not a custom marshaller.
- ~~Reuse for remarks~~ → **yes** — remarks are anchored to nodes; the spec-side
  index is the same structure remark-anchoring consumes.

**All design questions are resolved.** Remaining work is implementation
(milestones §7) and the `LX-model` sink migration when the internal-model stream
lands.

**Parked (future, separate features):** `x-go-origin` materialization for other
CLIs (§3.6 far horizon); precise token-level positions for LSP (§3.5);
background-spec (`InputSpec`) nodes are origin-less by design — navigation is
partial when an overlay is in play (§3.8).

---

## 9. Evidence / grounding (recon, file:line)

- **OnDiagnostic pattern to mirror:** `internal/scanner/options.go:57` (field) →
  `internal/scanner/scan_context.go:152` (`OnDiagnostic()` accessor) →
  `internal/builders/common/builder.go:94` (`RecordDiagnostic` nil-guards + fires).
  Every builder embeds `*common.Builder`, so the verb reaches every emit site.
- **Spec assembly / insertion sites:** definitions `spec/spec.go:71`; responses
  `responses/responses.go:66`; paths `routes.go:62` / `operations.go:57`.
- **Property emission:** `schema/fields.go:111` (`target.Properties[c.name]=ps`),
  with `s.Name` (enclosing def, set `schema.go:63`) + `c.name` + `c.afld` in scope.
- **Position bridges:** `grammar.Block.Pos()` / `grammar.Property.Pos`
  (`grammar/ast.go:32,120`); `EntityDecl{Spec,Ident,File,Pkg}`
  (`scanner/declaration.go:15`); `ScanCtx.FileSet()` / `ScanCtx.PosOf(token.Pos)`
  (`scan_context.go:119–139`).
- **Greenfield confirmed:** no existing jsonpointer usage, source-map, or
  node→pos storage in the builders. `common.Builder.MakeRef` *constructs* `$ref`
  pointers but stores no provenance. Nothing to retrofit.

---

## Change history

- 2026-06-03 — initial draft after the diagnostics pane landed (`b2574fe`) and
  the Q28 fix unblocked the oracle; grounded in builder-pipeline recon.
- 2026-06-03 (rev 2) — folded Fred's review. Reframe: **anchors-only emission +
  nearest-ancestor** (retires per-node cost and the pointer-stack); reverse map
  is one-to-many.
- 2026-06-03 (rev 3) — Fred's second pass. Corrected three overreaches: (a)
  `x-go-origin` demoted from "target sink" to a parked future feature for other
  CLIs; (b) "$ref is free" retracted — find-references needs render-time `$ref`
  resolution with its quirks, scoped as a later feature, distinct from the direct
  reverse; (c) §3.6 refocused on **the question I'd skipped — where the index
  lives**.
- 2026-06-03 (rev 4) — `RecordOrigin` design hashed out & **decided**:
  caller-owned callback stands; TUI holds a build-our-own structure (flat map
  MVP, optional trie), no external dep; record late from the settled point with
  upsert/last-wins, never evict on reset (§3.7); background-spec overlay nodes are
  origin-less by design (§3.8).
- 2026-06-03 (rev 5) — anchor set **closed**: enum values are anchored
  individually (`…/enum/{i}`) since we scan the consts anyway.
- 2026-06-03 (rev 6) — index structure **decided**: flat `map[pointer]ref` +
  zero-alloc segment-trim nearest-ancestor walk now; (radix) trie deferred to
  `LX-model` / subtree-query needs (§3.7, with the scale rationale). Reverse
  index noted as a separate `(file,line)`-sorted slice. `RecordOrigin` /
  source-side design fully settled.
- 2026-06-03 (rev 7) — **§6 UX drafted** from Fred's proposal: symmetric
  link-nav mode (driver pane + live follower), spec→source and source→spec flows,
  vertical-center "move focus", edge cases (no-source / multi-candidate / dirty
  buffer / format toggle), visual language. Flagged: the `n` key **collides with
  search next/prev** (`model.go:274/279`); editor-modality gate; auto-follow.
- 2026-06-03 (rev 8) — §6.6 UX calls **decided**: mode key **`f` (follow)** (n/N
  stay search); link-nav **enters from the read-only viewer** (ESC out of editing
  first); **auto-follow live + debounced**. §6 closed.
- 2026-06-03 (rev 9) — last open questions **resolved**, design settled:
  spec-side JSON index uses **`jsontext.Decoder.StackPointer`** (TUI bumps to
  go1.26; §4); **find-references is in scope** (`LX-refs`) with **`F3`/`Shift-F3`**
  next/prev navigation (§6.4, §8); spec-side index **is** the remark-anchoring
  structure (confirmed). Status → 🟢 ready to build; first milestone `LX-spec-0`.
