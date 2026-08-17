---
title: Shared parameters / responses — fixture harness & expected specs
stream: 9
status: approved
release: v0.36
issues: [go-swagger#2632]
---

# Shared parameters / responses — fixture harness (for validation)

This document is the **validation harness** for the shared-parameters feature.
It locks the grammar I intend to implement, then gives — per fixture — the spec I
expect codescan to produce. Fred validates the grammar + expectations + coverage;
once approved these become real goldens (`UPDATE_GOLDEN=1`) and the safe harness
for development.

> **Read the grammar section first.** Everything below depends on it; if my reading
> is off, the expected specs are easy to re-derive.

---

## 0. What already exists today (grounding)

- **`#/responses` + `$ref: #/responses/{name}` already exist.** Named
  `swagger:response` structs already land at the spec top level and operations
  that name them in a route `Responses:` block already emit a `$ref`. So the
  response side is mostly *force-registration + conflict-checking + the `*`
  marker*, not a new wiring mechanism.
- **`#/parameters` is never emitted, and no `$ref: #/parameters/` exists.** The
  entire parameter-sharing surface is new — this is the heart of #2632.

---

## 0b. Review round 2 — resolutions & added scope

Folding in Fred's annotations. The three open questions are resolved (confirm if
any is wrong), and four new contract items are added with fixture coverage.

### Resolved open questions

- **Q1 — reference-marker anchor.** A standalone reference marker may live on **any
  declaration** (the scanner reads every comment group); the **target is always
  explicit** (no opID inference). Rationale: keeps the marker free of its
  placement (you can co-locate path-item refs on a `func _()`, op refs on the
  route func or elsewhere) and avoids parser coupling to the route func. Op-id
  *omission/inference* is rejected for now.
- **Q2 — conflict determinism.** Survivor chosen by **package import path, then
  declaration position** (file order, then line). Stable, reproducible golden.
- **Q3 — dangling `$ref`.** **Warn + drop**, *not* keep. This matches the existing
  precedent (`fixtures/bugs/1109/responseref`: an unresolved `#/responses/` ref is
  dropped with a warning rather than emitting an invalid reference). Consistency
  wins; a dangling ref must never reach the output. Fixture 4 updated.

### Added contract items (from Fred's caveats)

- **C1 — dedup target lists** (§1a). `swagger:parameters * opID opID …` dedups the
  op-id list; a duplicate raises `scan.duplicate-target` (warning) and is dropped.
- **C2 — dedup reference lists** (§1b). `swagger:parameters target name name …`
  dedups the name list; a duplicate raises `scan.duplicate-ref` (warning).
- **C3 — name override is the key/reference that matters** (§1c, §1g below). The
  shared key is the **final, resolved** parameter name — after the `name:` keyword
  / `swagger:name` override and after `NameFromTags` tag selection — *not* the Go
  field name. References must use that final name. New §1g + Fixture 5.
- **C4 — prune extension** (§1a, §6b below). `PruneUnusedModels` gains a pass that
  prunes shared parameters/responses referenced by no operation/path-item, then
  re-runs definition reachability (a definition kept alive only by a now-pruned
  shared object becomes prunable). New §6b + Fixture 6.

### Disambiguation is non-trivial (§1c note)

Reference resolution must handle: **missing** name (→ drop + diagnostic, C2/Q3
family) and **ambiguous** name (post-conflict, the ref lands on the §2 survivor,
which may be the wrong shape — the conflict warning is the only available signal;
we cannot do better with short-name-only refs). No new mechanism, but the
resolver must not assume a clean 1:1.

---

## 1. Grammar (please validate)

`swagger:parameters` plays **two roles**, disambiguated by whether the marker is
attached to a **struct type** (a *definition*) or stands **alone** on a non-type
declaration (a *reference*).

### 1a. Definition markers — attached to a `type X struct {…}`

The struct's fields are the parameters. The **first token is the target**:

| Marker | Effect |
|--------|--------|
| `swagger:parameters opID [opID…]` | **inline** the fields into each named operation *(existing behavior, unchanged)* |
| `swagger:parameters /path` | **inline** the fields into the path-item `parameters` of the **exact** path *(new)* |
| `swagger:parameters *` | **register** each field at `#/parameters/{key}` *(new)*; not applied anywhere on its own |
| `swagger:parameters * opID [opID…]` | register at `#/parameters/{key}` **and** `$ref` each into the listed operations *(new — the small-spec convenience)* |

> **FRED**: I agree

> **FRED**: side note - ensure the opID [opID...] list is deduplicated. Drop duplicates. Emit warning if dupes found.

> **FRED**: side note - impact on "prune unused"
>
> pruning current only prunes unused definitions.
>
> It should be extended (with another pass) to prune unused shared parameters and shared responses (i.e. not referred by any operation).

### 1b. Reference markers — standalone (not on a struct)

No fields of its own; it wires an **existing** shared parameter into a target.
**First token = target, remaining tokens = shared parameter names**:

| Marker | Effect |
|--------|--------|
| `swagger:parameters opID name [name…]` | add `$ref: #/parameters/{name}` to the operation *(new — the scaling channel for big specs)* |
| `swagger:parameters /path name [name…]` | add `$ref: #/parameters/{name}` to the path-item *(new)* |


Each `{name}` is checked against the shared namespace; **missing → diagnostic,
ignore that argument**.

> **FRED**: I agree

> **FRED**: side note - ensure the name [name...] list is deduplicated. Drop duplicates. Emit warning if dupes found.

### 1c. Disambiguation rule

- Marker on a `type … struct` → **definition** (tokens are *targets*).
- Marker not on a struct (on a func / var / free comment) → **reference**
  (first token target, rest are *shared names*).

This cleanly separates the two genuinely-ambiguous cases:
`swagger:parameters listPets createPet` (two op targets, inline — needs a struct)
vs `swagger:parameters listPets X-Request-ID` (reference — no struct).

> **FRED**: I agree. And disambiguation is not trivial: can be missing, can be ambiguous short-name $ref

> **FRED**: attention. Parameter names may be overriden so the override is the reference that matters.

### 1d. Responses

Responses keep their existing naming channel — the route `Responses:` block, which
carries the status code the parameter side has no analog for. So:

| Marker / site | Effect |
|---------------|--------|
| `swagger:response *` | **force-register** the response at `#/responses/{StructName}` (1 struct = 1 response, so the key is the struct/`swagger:response`-name) |
| route `Responses:` → `200: ErrorResponse` / `default: ErrorResponse` | reference it → `$ref: #/responses/ErrorResponse` *(already works)* |

There is **no** `swagger:response * opID` form (a bare op list can't say *which
status*) and **no** path-item form (OAS2 path-items hold no responses).

> **FRED**: I agree

### 1e. Top-level key naming

- **Parameters:** key = the parameter **name** (`X-Request-ID`), because one
  struct yields N parameters. `#/parameters/X-Request-ID`.
- **Responses:** key = the struct / `swagger:response` **name**, because one
  struct = one response. `#/responses/ErrorResponse`.

### 1f. Override semantics (path-item vs operation)

Per OAS2: a parameter is unique by `(name, in)`. Path-item params and operation
params **both appear**; the operation-level one wins at resolution (it cannot be
*removed* at the operation level, only overridden). So an "override" is expressed
by **co-presence**, not subtraction — see Fixture 2.

> **FRED**: I agree. Good catch

### 1g. Name override is the reference that matters (C3)

A parameter's spec name is not necessarily its Go field name. It is the **final,
resolved** name, after (in precedence order): the `name:` keyword /
`swagger:name` override → the tag selected by `NameFromTags` (see the shipped
`NameFromTags` feature) → the Go field name. **The shared key and every reference
use that final name.**

Example: a field `RequestID string` tagged `json:"X-Request-ID"` but carrying
`// name: X-Correlation-ID` registers at `#/parameters/X-Correlation-ID`, and a
reference must say `X-Correlation-ID` (not `X-Request-ID`, not `RequestID`).
Witnessed by Fixture 5.

### ⚠ Open questions — resolved

The three open questions raised in round 1 (reference-marker anchor, conflict
determinism, dangling-ref policy) are resolved in **§0b**. No open grammar
questions remain; the harness below reflects the resolutions.

---

## 2. Cross-package conflicts & namespaces (challenging your belief)

**I agree with you — do not attempt cross-package conflict *resolution*.** But one
nuance vs the `#/definitions` machinery:

- `#/definitions` resolves collisions by **renaming** (`scan.renamed-definition`).
- Shared params/responses **must not be renamed** — they are referred to *only* by
  short name, so a rename would silently break every `$ref` that points at them.
  Therefore the policy here is **keep-first + drop-duplicate + warn**, never rename.

> **FRED**: I agree.

So: on a duplicate shared **short name**, keep the first registration, drop the
later one, emit a `scan.shared-*-conflict` **warning**, and continue. The dropped
struct's `$ref`s still point at the surviving definition (correct iff they're
shape-compatible; if not, that's the author's bug and the warning is the signal).

**Namespaces are independent.** `#/definitions/X`, `#/parameters/X`,
`#/responses/X` do not collide with each other — only same-namespace short names
conflict. Fixture 3 witnesses both: a real cross-package conflict *and* a
benign same-short-name-different-namespace coexistence.

**Determinism question (for you):** "keep-first" needs a deterministic order or the
golden is unstable. I propose ordering by **package import path, then declaration
position** — stable and reproducible. Confirm?

---

## 3. swagger:operation wholesale YAML

Parameters/responses defined via the post-`swagger:operation` YAML body must go
through the **same checks**:

- a `$ref: '#/parameters/{name}'` / `'#/responses/{name}'` in the YAML is
  validated against the shared namespace → missing ⇒ diagnostic;
- the YAML cannot *define* new shared (top-level) objects — it is operation-scoped
  — so there is nothing to register from here, only references to validate.

Fixture 4 covers a resolving ref and a dangling ref.

---

## 4. Fixtures & expected specs

Four fixtures under `fixtures/enhancements/`. Expected specs are written as YAML
for readability; **codescan-derived incidentals** (body-param synthetic names,
the error-response body definition name, parameter array ordering) are marked
`# codescan-derived` and will be pinned by the generated golden — focus review on
the **new structural elements** (`$ref`s, top-level `parameters`/`responses`,
co-presence overrides, diagnostics).

---

### Fixture 1 — `shared-parameters/` (single package)

Spec-level `*` namespace + **both** parameter reference channels + shared response.

**Source (sketch):**

- `CommonHeaders` — `swagger:parameters *`, field `RequestID json:"X-Request-ID"` `in: header`
  → registers `#/parameters/X-Request-ID` (referenced by nobody directly).
- `AuthHeader` — `swagger:parameters * createPet`, field `APIKey json:"X-API-Key"` `in: header` `required`
  → registers `#/parameters/X-API-Key` **and** `$ref`s into `createPet`.
- `ListPetsParams` — `swagger:parameters listPets`, field `Limit json:"limit"` `in: query`
  → inlined into `listPets`.
- standalone on the `listPets` route func: `swagger:parameters listPets X-Request-ID`
  → `$ref`s `X-Request-ID` into `listPets`.
- `CreatePetParams` — `swagger:parameters createPet`, field `Body Pet` `in: body` `required`.
- `ErrorResponse` — `swagger:response *` → registers `#/responses/ErrorResponse`.
- `Pet` — `swagger:model`.
- routes `listPets GET /pets`, `createPet POST /pets`, both `Responses: default: ErrorResponse`.

**Expected spec:**

```yaml
paths:
  /pets:
    get:                       # listPets
      operationId: listPets
      parameters:
        - {name: limit, in: query, type: integer, format: int64}  # inline
        - {$ref: '#/parameters/X-Request-ID'}                      # reference channel
      responses:
        default: {$ref: '#/responses/ErrorResponse'}
    post:                      # createPet
      operationId: createPet
      parameters:
        - {$ref: '#/parameters/X-API-Key'}                         # from "* createPet"
        - name: petBody        # codescan-derived body name
          in: body
          required: true
          schema: {$ref: '#/definitions/Pet'}
      responses:
        default: {$ref: '#/responses/ErrorResponse'}
parameters:
  X-Request-ID: {name: X-Request-ID, in: header, type: string}
  X-API-Key:    {name: X-API-Key, in: header, required: true, type: string}
responses:
  ErrorResponse:
    description: ErrorResponse describes ...           # codescan-derived
    schema: {$ref: '#/definitions/errorResponseBody'} # codescan-derived name
definitions:
  Pet: {type: object, properties: {name: {type: string}}}
  # + error response body definition
```

Witnesses: register-only `*`; `* opID` register+ref; standalone op reference
channel; inline + `$ref` co-existing on one operation; shared response `$ref`.

---

### Fixture 2 — `shared-parameters-pathitem/` (single package)

Path-item inline + path-item reference + `(name,in)` override + exact-path
(no hierarchy).

**Source (sketch):**

- `CommonHeaders` — `swagger:parameters *`, `RequestID json:"X-Request-ID"` `in: header`
  → `#/parameters/X-Request-ID`.
- `PetPathParams` — `swagger:parameters /pets`, `APIKey json:"X-API-Key"` `in: header` `required: true`
  → inlined into `paths./pets.parameters`.
- standalone `func _()`: `swagger:parameters /pets X-Request-ID`
  → `$ref`s `X-Request-ID` into `paths./pets.parameters`.
- `ListPetsParams` — `swagger:parameters listPets`, `APIKey json:"X-API-Key"` `in: header` `required: false`
  → operation-level **override** of the path-item `X-API-Key`.
- routes: `listPets GET /pets`, `createPet POST /pets`, `getPet GET /pets/{id}`
  (separate exact path — must **not** inherit `/pets` path-item params).

**Expected spec:**

```yaml
paths:
  /pets:
    parameters:                                   # path-item level (inherited, no $ref for inline)
      - {name: X-API-Key, in: header, required: true, type: string}   # inline from /pets struct
      - {$ref: '#/parameters/X-Request-ID'}                            # path-item reference
    get:                                          # listPets
      operationId: listPets
      parameters:
        - {name: X-API-Key, in: header, required: false, type: string} # OVERRIDE (co-present; op wins)
    post:                                         # createPet — inherits path-item as-is, no own params
      operationId: createPet
  /pets/{id}:
    get:                                          # getPet — exact path, NO /pets inheritance
      operationId: getPet
      parameters:
        - {name: id, in: path, required: true, type: string}
parameters:
  X-Request-ID: {name: X-Request-ID, in: header, type: string}
```

Witnesses: path-item inline (no `$ref`); path-item `$ref` reference; `(name,in)`
override by co-presence; no path hierarchy (`/pets/{id}` clean).

---

### Fixture 3 — `shared-parameters-conflict/` (multi-package)

Duplicate shared short names across packages → keep-first + drop + warn; plus a
cross-namespace non-conflict witness.

**Source (sketch):**

- `pkga/api.go`:
  - `TokenA` — `swagger:parameters *`, `Token json:"X-Token"` `in: header`
    → `#/parameters/X-Token` (the survivor).
  - `ErrorResponse` — `swagger:response *`, body `{code int}` → `#/responses/ErrorResponse` (survivor).
  - `Status` — `swagger:model` → `#/definitions/Status`.
  - `StatusParam` — `swagger:parameters *`, `Status json:"Status"` `in: header`
    → `#/parameters/Status` (coexists with `#/definitions/Status` — **no** conflict).
- `pkgb/api.go`:
  - `TokenB` — `swagger:parameters *`, `Token json:"X-Token"` `in: query` (**different `in`**, same key)
    → conflict on `X-Token` → dropped + warning.
  - `ErrorResponse` — `swagger:response *`, body `{message string}` (different shape, same key)
    → conflict on `ErrorResponse` → dropped + warning.
- one route per package so the scan has operations.

**Expected spec (structural):**

```yaml
parameters:
  X-Token: {name: X-Token, in: header, type: string}   # pkga wins (import-path order)
  Status:  {name: Status,  in: header, type: string}    # coexists with definitions/Status
responses:
  ErrorResponse: {description: ..., schema: {$ref: '#/definitions/...'}}  # pkga wins
definitions:
  Status: {type: object, ...}                           # independent namespace, no clash
```

**Expected diagnostics:**

- `scan.shared-parameter-conflict` (warning): duplicate shared parameter short
  name `X-Token`; kept `pkga`, dropped `pkgb`.
- `scan.shared-response-conflict` (warning): duplicate shared response short name
  `ErrorResponse`; kept `pkga`, dropped `pkgb`.

Witnesses: keep-first/drop/warn (no rename); deterministic survivor;
cross-namespace coexistence (`#/definitions/Status` vs `#/parameters/Status`).

> Depends on the **determinism** decision in §2.

---

### Fixture 4 — `shared-parameters-yaml/` (single package)

`swagger:operation` wholesale YAML referencing the shared namespace, plus a
dangling reference.

**Source (sketch):**

- `CommonHeaders` — `swagger:parameters *` → `#/parameters/X-Request-ID`.
- `ErrorResponse` — `swagger:response *` → `#/responses/ErrorResponse`.
- `swagger:operation GET /a opA` YAML:
  ```yaml
  parameters:
    - {$ref: '#/parameters/X-Request-ID'}
  responses:
    default: {$ref: '#/responses/ErrorResponse'}
  ```
  → both refs resolve, kept verbatim.
- `swagger:operation GET /b opB` YAML referencing `#/parameters/DoesNotExist`
  and `#/responses/Missing` → both dangling.

**Expected spec:**

```yaml
paths:
  /a:
    get:
      operationId: opA
      parameters: [{$ref: '#/parameters/X-Request-ID'}]
      responses: {default: {$ref: '#/responses/ErrorResponse'}}
  /b:
    get:
      operationId: opB
      # both dangling refs DROPPED (Q3: warn + drop, per 1109 precedent)
      responses: {}
parameters:
  X-Request-ID: {name: X-Request-ID, in: header, type: string}
responses:
  ErrorResponse: {...}
```

**Expected diagnostics:** `scan.dangling-parameter-ref` / `scan.dangling-response-ref`
(warnings) for `DoesNotExist` / `Missing`.

> Resolved (Q3 / §0b): **warn + drop**. A dangling ref must never reach output;
> matches `fixtures/bugs/1109` (unresolved `#/responses/` ref dropped with a
> warning). `opB` is left with empty (dropped) refs.

---

### Fixture 5 — `shared-parameters-overrides/` (single package)

Name override as the shared key/reference (C3) + target/reference dedup (C1, C2).

**Source (sketch):**

- `CommonHeaders` — `swagger:parameters *`, field `RequestID json:"X-Request-ID"`
  with `// name: X-Correlation-ID` `in: header`
  → registers `#/parameters/X-Correlation-ID` (the **overridden** name, not
  `X-Request-ID`).
- `AuthHeader` — `swagger:parameters * createThing createThing` (**dup op id**),
  field `APIKey json:"X-API-Key"` `in: header`
  → registers `#/parameters/X-API-Key`, `$ref` into `createThing` **once**
  (dedup) + `scan.duplicate-target` warning.
- standalone on the `listThings` route func:
  `swagger:parameters listThings X-Correlation-ID X-Correlation-ID` (**dup name**)
  → `$ref` `X-Correlation-ID` into `listThings` **once** (dedup) +
  `scan.duplicate-ref` warning. Also proves a reference resolves by the
  *overridden* name.
- routes `listThings GET /things`, `createThing POST /things`.

**Expected spec:**

```yaml
paths:
  /things:
    get:                       # listThings
      operationId: listThings
      parameters: [{$ref: '#/parameters/X-Correlation-ID'}]   # deduped; overridden name
    post:                      # createThing
      operationId: createThing
      parameters: [{$ref: '#/parameters/X-API-Key'}]          # deduped op-id target
parameters:
  X-Correlation-ID: {name: X-Correlation-ID, in: header, type: string}
  X-API-Key:        {name: X-API-Key, in: header, type: string}
```

**Expected diagnostics:** `scan.duplicate-target` (createThing), `scan.duplicate-ref`
(X-Correlation-ID).

Witnesses: shared key = final/overridden name (C3); op-id list dedup (C1);
reference-name list dedup (C2); reference resolves by overridden name.

---

### Fixture 6 — `shared-parameters-prune/` (single package, scanned twice)

Prune extension (C4): shared params/responses referenced by no operation are
pruned under `PruneUnusedModels`, retained without it.

**Source (sketch):**

- `UsedHeader` — `swagger:parameters *`, `Used json:"X-Used"` `in: header`
  → `#/parameters/X-Used`; **referenced** by `listP` → survives.
- `UnusedHeader` — `swagger:parameters *`, `Unused json:"X-Unused"` `in: header`
  → `#/parameters/X-Unused`; **referenced by nobody** → pruned under prune.
- `UsedResponse` — `swagger:response *` → referenced by `listP`'s block → survives.
- `UnusedResponse` — `swagger:response *` → referenced by nobody → pruned.
- standalone on `listP`: `swagger:parameters listP X-Used`;
  route `listP GET /p`, `Responses: default: UsedResponse`.

**Expected — `ScanModels` only (no prune):** all four present
(`#/parameters/{X-Used,X-Unused}`, `#/responses/{UsedResponse,UnusedResponse}`).

**Expected — `ScanModels` + `PruneUnusedModels`:**

```yaml
parameters:
  X-Used: {name: X-Used, in: header, type: string}    # X-Unused pruned
responses:
  UsedResponse: {...}                                   # UnusedResponse pruned
```

**Expected diagnostics:** `scan.pruned-unused` (Hint) for `X-Unused` and
`UnusedResponse`.

Witnesses: the new prune pass over `#/parameters` + `#/responses` (C4); twice-scan
toggle mirrors `coverage_prune_unused_test.go`.

> See §6b for the pass ordering this depends on.

---

## 6b. Prune extension design (C4)

Today (`internal/scanner/README.md#prune`) shared `responses`/`parameters` are
**reachability roots** — they keep their downstream definitions alive but are
never themselves pruned. The extension adds, under `PruneUnusedModels`:

1. **Prune unused shared objects first.** Drop `#/parameters/*` and
   `#/responses/*` entries referenced by no operation **and** no path-item
   (`scan.pruned-unused` Hint each). `InputSpec`-supplied entries are pinned
   (never pruned), mirroring the definitions rule.
2. **Then re-run definition reachability** with the *surviving* shared objects as
   roots — so a definition kept alive only by a now-pruned shared object becomes
   prunable in the existing definitions pass. Order matters: shared-object prune
   precedes the definition prune, which still precedes name reduction.

This is additive to the existing prune; the definitions behavior is unchanged when
no shared objects are pruned.

---

## 5. Coverage matrix

| Capability | F1 | F2 | F3 | F4 | F5 | F6 |
|------------|----|----|----|----|----|----|
| `swagger:parameters *` register-only | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `swagger:parameters * opID` register+ref | ✓ | | | | ✓ | |
| op reference channel `swagger:parameters opID name` | ✓ | | | | ✓ | ✓ |
| path-item inline `swagger:parameters /path` | | ✓ | | | | |
| path-item reference `swagger:parameters /path name` | | ✓ | | | | |
| `(name,in)` override (co-presence) | | ✓ | | | | |
| exact-path / no hierarchy | | ✓ | | | | |
| `swagger:response *` + route-block `$ref` | ✓ | | ✓ | | | ✓ |
| cross-package `$ref` | | | ✓ | | | |
| duplicate short-name conflict (param + resp) | | | ✓ | | | |
| cross-namespace non-conflict | | | ✓ | | | |
| `swagger:operation` YAML ref (resolving) | | | | ✓ | | |
| `swagger:operation` YAML ref (dangling → drop+diag) | | | | ✓ | | |
| name override = shared key/ref (C3) | | | | | ✓ | |
| op-id list dedup (C1) | | | | | ✓ | |
| reference-name list dedup (C2) | | | | | ✓ | |
| prune unused shared param/resp (C4) | | | | | | ✓ |

---

## 6c. Implementation phasing (approved 2026-06-24, build started)

Hooks confirmed by a flow trace (`swagger.Parameters` + `PathItem.Parameters` are
never written today — routes.go:28 says "shared parameters - unsupported for now";
the responses top-level path `spec.Builder.buildResponses` is the template).

- **P1 — grammar front-end (scanner).** Recognise the new `swagger:parameters`
  targets/forms and the standalone reference markers; dedup op-id (C1) and name
  (C2) lists. Files: `internal/parsers/regexprs.go` (let `rxParametersOverride`
  admit `*`, `/path`, and reference forms), `internal/parsers/matchers.go`,
  `internal/scanner/declaration.go` (classify target + extract names; new methods
  beside `OperationIDs()`), `internal/scanner/index.go` (discover standalone
  reference markers on non-struct decls). Unit-tested at the parser level.
- **P2 — top-level `#/parameters` + conflict (B, C3).** `spec.go`: init
  `input.Parameters`; `buildSharedParameters()` mirroring `buildResponses()`;
  key by final/overridden name; keep-first/drop/warn conflict (import-path then
  pos). Anchors via `JSONPointer("parameters", name)`.
- **P3 — reference → `$ref` into operations (C).** `* opID` and standalone
  `opID name` emit a `$ref` parameter on the operation; missing → drop+warn;
  reuse `mergeBoundParameters` for `(name,in)` co-presence.
- **P4 — path-item parameters (C).** `/path` inline → `PathItem.Parameters`;
  `/path name` → `PathItem.Parameters` `$ref`; exact-path; co-presence override.
  Lift the routes.go:28 "unsupported" stub.
- **P5 — `swagger:response *` force-register + conflict (D).** Symmetric; extends
  existing top-level response path.
- **P6 — `swagger:operation` YAML ref validation (E).** Validate `$ref`s to the
  shared namespace; drop+warn dangling (Q3). Hook near `operations/walker.go`.
- **P7 — prune extension (C4).** New pass per §6b; `spec/prune.go`.
- **P8 — tests + goldens (G).** Integration coverage tests for F1–F6 mirroring
  `coverage_*_test.go`; generate goldens; verify against the expected specs above.

Each phase is an independently buildable/testable step; pause + reassess at phase
boundaries.

**Build log.**
- ✅ **P1 — grammar-owned targeting parse** (2026-06-24). *Layering correction:*
  the first P1a attempt parsed targeting in the scanner-facing `parsers` regex
  layer + rewired `EntityDecl.OperationIDs()`. That was the wrong layer — the
  grammar already parses `swagger:parameters` args (`ParametersBlock`), and the
  builder was the only thing still reading the scanner regex (a leftover from the
  grammar migration). Redone correctly:
  - **Grammar** owns the parse: new `TokenWildcard` (`*`); `classifyParametersArgs`
    (lexer) emits `*`/`/path`/ident tokens; `parseParametersArgs` (parser) →
    enriched `ParametersBlock{Target, Path, Args, Dups}` (+ `OperationIDs()`
    accessor), with dedup. `*`/`/path` valid only in first position.
  - **Builder** (`parameters.Build`) reads the grammar `ParametersBlock`(s) via
    `ParseBlocks(decl.Comments)` — accumulating across multiple
    `swagger:parameters` lines — instead of `EntityDecl.OperationIDs()`.
  - **Scanner**: `EntityDecl.OperationIDs()` deleted; `rxParametersOverride`
    reduced to a *permissive* presence gate — `swagger:parameters` + any
    non-empty arg (content unused) — so malformed args pass through to the
    grammar for diagnosis instead of being silently dropped (no arg-shape
    re-validation in the scanner).
  - Reverted the misplaced `parsers/shared_params.go(+test)`. Grammar lexer/parser
    unit tests added; full suite green, lint clean, **no golden drift**.
- ✅ **P1b — scanner discovery** (2026-06-24). New `scanner.ParameterRef`
  {Comments, File, Pkg} + `TypeIndex.ParameterRefs` + `ScanCtx.ParameterRefs()`
  iterator. `processFileDecls` now collects a `swagger:parameters` marker on a
  **func doc** as a reference (`collectParameterRef`) — per §1c a func-hosted
  marker is always a reference, never a definition (those stay on struct
  TypeSpecs). Args still parsed by the grammar when a builder consumes the ref.
  Tests: `parameter_ref_test.go` (discovery via grammar Block on Fixtures 1 & 2 +
  negative on the classification corpus) and a fixtures-scan-clean smoke test over
  all 6 fixtures. *Caught & fixed:* Fixtures 1 & 2 package docs enumerated raw
  `swagger:*` forms in prose, which the loose classifier flagged (Fixture 1: a
  spurious parameters/response struct-conflict). Slimmed both package docs — forms
  are documented on the declarations that exercise them. Boundary: only func docs
  are scanned for refs (covers the fixtures); var / free-comment hosts are not (no
  current need). Full suite green, lint clean, no golden drift.
- ✅ **P2 — top-level `#/parameters` + conflict** (2026-06-24). `spec.Builder`
  initialises `input.Parameters` and carries the `parameters` map (mirrors
  `responses`; empty map is omitted, no golden drift). `parameters.Builder.Build`
  dispatches per `ParametersBlock.Target`: operations → `buildIntoOperations`
  (unchanged inline); shared (`*`) → `buildShared` harvests fields keyed by
  resolved name (C3 free) → `SharedParameters()`. `spec.registerSharedParameters`
  merges with **keep-first** conflict (new `CodeSharedParameterConflict` warning,
  never rename); resolution sorted by (pkg path, pos, name) so it is
  load-order-independent. Origins skipped for shared (empty opID). Tests:
  `coverage_shared_parameters_test.go` — top-level registration (Fixture 1) +
  conflict/keep-first + independent namespaces (Fixture 3), survivor determinism
  asserted. Path-item (`/path`) target stubbed for P4. Full suite green, lint
  clean, no golden drift. parameters/README §builder updated.
- ✅ **P3 — reference → `$ref` into operations** (2026-06-24). Both channels:
  `* opID` (parameters.Builder exposes `SharedRefOperations()`; intents collected
  in buildParameters) and standalone func refs (`ScanCtx.ParameterRefs`, parsed by
  the grammar). New `spec.applyParameterRefs` runs after the shared map is complete
  and before buildRoutes: resolves each `{op, name}` to a `#/parameters/{name}`
  $ref on the operation (get-or-create), deduped, applied in deterministic
  (op, name) order. Dangling ref → drop + `scan.dangling-parameter-ref` warning
  (Q3). C1 (`scan.duplicate-target`, emitted in parameters.Builder) and C2
  (`scan.duplicate-ref`, emitted in collectStandaloneParameterRefs) from grammar
  Dups. New codes: CodeDanglingParameterRef / CodeDuplicateTarget / CodeDuplicateRef.
  Tests: Refs (Fixture 1, both channels + inline coexists) and OverridesAndDedup
  (Fixture 5: C3 overridden-name ref, C1, C2, dangling drop — added a danglingRef
  witness func to Fixture 5). Full suite green, lint clean, no golden drift.
- ✅ **P4 — path-item parameters** (2026-06-24). `parameters.buildPathItem`
  harvests `/path` struct fields into `PathItemParameters()` (keyed by exact path).
  New `spec.applyPathItemParameters` runs **after** buildRoutes/buildOperations
  (paths must exist): appends inline params (sorted pkg/pos) then `/path name`
  $refs (sorted name) to `PathItem.Parameters`. Exact-path (unknown path →
  drop+warn via CodeInvalidAnnotation); co-presence override (op params untouched —
  both appear, op wins per OAS2). Standalone path refs from `ParameterRefs`
  (Target==path) via `collectPathItemRefs`; dangling→drop+warn, C2 dups. Test:
  PathItem (Fixture 2 — inline X-API-Key required + $ref X-Request-ID on /pets
  path-item; listPets op-level X-API-Key required:false co-present; /pets/{id} does
  NOT inherit). Full suite green, lint clean, no golden drift.
- ✅ **P5 — `swagger:response *` synonym + conflict + grammar migration**
  (2026-06-24). *Scope grew on investigation:* responses ALREADY register
  top-level + are referenceable (pervasive in goldens — we never called them
  "shared"); `swagger:response *` was silently dropped (`*` failed the name
  regex — the same grammar leftover P1 fixed for parameters). Per Fred's call:
  `*` is a **pure synonym** for bare `swagger:response` (no op-ids — too hard to
  document a default-injection; no `/path`). Done:
  - **Grammar**: lexer accepts `*` for AnnResponse (TokenWildcard) → ResponseBlock
    Name="" (= bare). `swagger:response` and `swagger:response *` are synonyms.
  - **Regex**: `rxResponseOverride` accepts `*` in the name alternation while still
    rejecting a malformed `utils.Error` (bug-874 preserved); classification gate
    only.
  - **Killed scanner method**: `EntityDecl.ResponseNames()` removed; name resolved
    from the grammar via new `responses.Builder.ResponseName()` (mirrors P1).
  - **Conflict**: `spec.buildResponses` orders decls (pkg path, pos) and applies
    keep-first with new `CodeSharedResponseConflict`; InputSpec overlay merge
    preserved (only scanned-vs-scanned collide). Safe — keep-first vs last-wins
    only differ on collision, absent from clean fixtures → no golden drift.
  - Tests: SharedResponses (Fixture 1 — `*` registers ErrorResponse, route
    `default:` now resolves to `$ref`), Conflict (Fixture 3 — response keep-first +
    warning), grammar `swagger:response *` synonym. Full suite green, lint clean,
    no golden drift. responses/README §builder updated.
  - **no-arg asymmetry confirmed JUSTIFIED**: `swagger:parameters` (no arg) is an
    error (N params/struct, no single name); `swagger:response`/`*` infer the type
    name (1 response/struct). Documented for the record.
- ✅ **P6 — `swagger:operation` YAML ref validation** (2026-06-24). New
  `spec.validateSharedRefs` runs after applyPathItemParameters (paths + namespace
  complete): walks every path-item/operation `parameters` + operation `responses`,
  drops dangling `#/parameters/{name}` (CodeDanglingParameterRef) and
  `#/responses/{name}` (new CodeDanglingResponseRef) refs — never emits an invalid
  reference (Q3). Catches swagger:operation wholesale-YAML refs (unmarshaled
  verbatim) and is a uniform safety net (valid refs + #/definitions refs
  untouched). Test: YAMLRefs (Fixture 4 — opA resolving refs kept, opB dangling
  refs dropped + both warnings). Full suite green, lint clean, no golden drift.
- ✅ **P7 — prune extension** (2026-06-25). New `spec.pruneUnusedSharedObjects`
  runs inside `pruneUnusedModels` BEFORE the definition reachability walk: drops
  every `#/parameters/*` / `#/responses/*` referenced by no operation and no
  path-item (`collectSharedRefs` is the read-only "is referenced" mirror), each
  with a located `scan.pruned-unused` Hint. `InputSpec`-supplied shared objects
  are pinned via `pinnedParams`/`pinnedResponses` (snapshot at NewBuilder), never
  pruned. Surviving shared objects then seed the existing definition walk, so a
  def kept alive only by a now-pruned shared object becomes prunable. Gated by
  ScanModels like the rest of the prune (no-op contract unchanged). Provenance:
  a pruned shared response would dangle its `/responses/{name}/...` anchors
  (fired inline from the responses builder), so under PruneUnusedModels those
  anchors are buffered (new `ScanCtx.BeginDeferredOrigins`/`Drop`/`Flush`,
  verbatim-flush sibling of the def-origins window), dropped on prune, and
  flushed after — non-prune anchor stream byte-identical. Tests: Prune_Off +
  Prune_On (Fixture 6) + golden `enhancements_shared_parameters_prune.json`;
  provenance non-leak verified. README §prune updated. Full suite green, no
  golden drift, lint clean (`--new-from-rev master`).
  - ⬜ Remaining P8: backfill goldens for F1/F3/F4/F5 (F2 + F6 goldens exist;
    F1/F3/F4/F5 are covered by behavioral assertions today).

## 6. After validation

1. Resolve the open questions (§1d, §2, §4).
2. Generate goldens (`UPDATE_GOLDEN=1 go test ./...`) once the feature is built —
   until then these YAML expectations are the contract.
3. Wire integration coverage tests mirroring the existing
   `internal/integration/coverage_*_test.go` pattern.
```
