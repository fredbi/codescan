# W3 — Alias workshop judgment ledger

Date: 2026-06-03
Status: ⬜ open (cycle 1 — calibration)
Companion: `alias-matrix.md` (axes + gap inventory),
`alias-handling.md` (the questionnaire, will be back-filled from this
ledger).

This is the running record of empirical judgments made during the
TUI-driven corpus walk. Each row is one **(fixture × cell × mode)**
combination. Filling rows out of order is fine — we'll reorder
once patterns stabilize.

Judgment legend:
- ✓ — current shape looks right
- ✗ — current shape looks wrong; "Desired shape" column says what
- ? — undecided; come back to it
- ~ — depends on something else still being decided

---

## Cycle 1 — Calibration (fixture: `alias-calibration-primitive`)

Source: `fixtures/enhancements/alias-calibration-primitive/types.go`.

Six cells × three modes = 18 judgments. The agreed shape of these
calibrates vocabulary for every later fixture.

Open both TUIs pointed at the fixture; toggle the alias modes inside
each. Fill **Current shape (TUI₁)** from what you see in the spec
panel; fill **Judgment** after looking at it; if ✗, fill **Desired
shape** to record what you wish you saw.

### 1.1 Decl-as-model × `UserID = int64`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `{type: integer, format: int64, title, x-go-package}` | ✓ | | annotation forces definition; user-intended |
| RefAliases=true | `{type: integer, format: int64, title, x-go-package}` — identical to Expand | ✓ | | annotation override; same as Expand |
| TransparentAliases=true | `{type: integer, format: int64, title, x-go-package}` — **identical to Expand and Ref** | ? | `{$ref}` ↔ inline ambiguity at field sites makes the def **dangling** here; see open question Q-A below | swagger:model annotation forces decl-level registration even under Transparent |

### 1.2 Decl-as-model × `Name = string`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `{type: string, title, x-go-package}` | ✓ | | |
| RefAliases=true | `{type: string, title, x-go-package}` — identical to Expand | ✓ | | |
| TransparentAliases=true | `{type: string, title, x-go-package}` — identical | ? | per Q-A | dangling under Transparent — see Q-A |

### 1.3 Decl-as-model × `Active = bool`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `{type: boolean, title, x-go-package}` | ✓ | | |
| RefAliases=true | `{type: boolean, title, x-go-package}` — identical to Expand | ✓ | | |
| TransparentAliases=true | `{type: boolean, title, x-go-package}` — identical | ? | per Q-A | dangling under Transparent — see Q-A |

### 1.4 Field-in-Envelope × `User UserID`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `{$ref: #/definitions/UserID}` — field godoc lost | ✓ | | annotation → $ref intended; field-godoc loss is the DescWithRef trade |
| RefAliases=true | `{$ref: #/definitions/UserID}` — identical to Expand | ✓ | | |
| TransparentAliases=true | `{description: "User identifier — …", type: integer, format: int64, x-go-name: User}` — **inline, godoc preserved** | ? | per Q-A | annotation says $ref; mode says inline — which wins? |

### 1.5 Field-in-Envelope × `Nick Name`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `{$ref: #/definitions/Name}` — field godoc lost | ✓ | | |
| RefAliases=true | `{$ref: #/definitions/Name}` — identical to Expand | ✓ | | |
| TransparentAliases=true | `{description: "Nick — …", type: string, x-go-name: Nick}` — inline | ? | per Q-A | |

### 1.6 Field-in-Envelope × `On Active`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `{$ref: #/definitions/Active}` — field godoc lost | ✓ | | |
| RefAliases=true | `{$ref: #/definitions/Active}` — identical to Expand | ✓ | | |
| TransparentAliases=true | `{description: "On — …", type: boolean, x-go-name: On}` — inline | ? | per Q-A | |

### 1.7 Decl (auto-discovered) × `LegacyID = int64`

Unannotated alias. Per Fred's rule it should NOT produce a dangling
definition.

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | | | | not yet captured for extended fixture |
| RefAliases=true | **PRESENT** in `definitions` as `{description, type: integer, format: int64, x-go-package}` | ? | per Q-B below | NOT dangling (Envelope.legacy `$ref`s it); but the *unannotated* alias gets a definition emitted by buildAlias's MakeRef path |
| TransparentAliases=true | absent from `definitions` | ✓ | | rule confirmed — no def under Transparent |

### 1.8 Decl (auto-discovered) × `LegacyName = string`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | | | | not yet captured |
| RefAliases=true | PRESENT as `{description, type: string, x-go-package}` | ? | per Q-B | NOT dangling — Envelope.legacyNick refs it |
| TransparentAliases=true | absent from `definitions` | ✓ | | rule confirmed |

### 1.9 Field-in-Envelope × `Legacy LegacyID` (unannotated RHS)

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | | | | not yet captured |
| RefAliases=true | `{$ref: #/definitions/LegacyID}` — identical shape to annotated `user` field | ? | per Q-B | the field-level path doesn't distinguish annotated from unannotated under non-Transparent modes |
| TransparentAliases=true | `{description: "Legacy — …", type: integer, format: int64, x-go-name: Legacy}` — inline | ✓ | | |

### 1.10 Field-in-Envelope × `LegacyNick LegacyName` (unannotated RHS)

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | | | | not yet captured |
| RefAliases=true | `{$ref: #/definitions/LegacyName}` — identical shape to annotated `nick` field | ? | per Q-B | same as 1.9 |
| TransparentAliases=true | `{description: "LegacyNick — …", type: string, x-go-name: LegacyNick}` — inline | ✓ | | |

### Open question Q-A — annotation × Transparent at field sites

In Transparent mode with `swagger:model UserID = int64` and `User
UserID` field, the current behaviour is:

- `UserID` definition exists in `definitions` (annotation forces it).
- Field `Envelope.user` is inlined as `{type:integer, format:int64,
  description, x-go-name}`.

Result: **`UserID` is dangling** — nothing in the spec `$ref`s it.

By Fred's rule ("`swagger:model` forces a `$ref`; auto-discovered
should not produce a dangling definition"), this current shape
violates the rule for annotated aliases. Two ways to resolve:

- **Option A — annotation wins everywhere.** Even in Transparent
  mode, an annotated alias produces `$ref` at field sites. Transparent
  affects only the *unannotated* path. Result for our fixture: the
  Transparent capture would be byte-identical to Expand. Modes
  effectively only matter when there's no annotation.
- **Option B — Transparent wins for fields; the dangling definition
  is the user's choice.** The annotation guarantees the definition
  exists if the user wants the named entity available (downstream
  generators can `$ref` it explicitly elsewhere), but Transparent
  inlines at every use site regardless. The dangling is opt-in.
- **Option C — Transparent dissolves *everything* annotated too.**
  Annotation has no effect under Transparent; no `UserID` definition,
  no `$ref`, inline everywhere. Symmetric with auto-discovered case.

Fred to pick before judging the Transparent rows.

### Open question Q-B — strict-vs-strong reading of "no dangling definition" rule

Fred's cycle 1 rule: *"swagger:model forces a $ref; auto-discovered
should not produce a dangling definition."* Captures expose two
readings:

- **Strict (dangling = unreferenced):** "no dangling" means "no
  definitions that are not `$ref`'d from anywhere." Under this, the
  current RefAliases output is fine — `LegacyID` exists but is
  properly referenced by `Envelope.legacy`. Rule passes.
- **Strong (unannotated = no definition):** "no independent
  definition for unannotated aliases, regardless of whether
  references exist." Under this, RefAliases violates the rule —
  `LegacyID` is a definition even though the user never asked for it.

The Transparent capture is consistent with both readings (no
`LegacyID` either way). RefAliases distinguishes them — RefAliases
is fine under strict, violates under strong.

**Why this matters downstream.** Under strict (status quo), a code
generator reading the RefAliases spec generates a Go typedef for
`LegacyID` because it's a named definition. Under strong (rule
change), the generator would see only `int64` and the named-entity
identity is lost at the spec level.

This is the central question for what RefAliases / Expand mode *means*
for unannotated aliases. Fred to pick before judging 1.7-1.10
RefAliases rows.

### 1.7 Cycle 1 summary

After filling the 18 cells above, write 2-3 sentences here:

- **What's the pattern in the ✓ cells?** What rule, if any, unifies the
  current behaviour we agree with?
- **What's the pattern in the ✗ cells?** What rule would have produced
  the desired shape instead?
- **First-order vocabulary calibrated:** what we now mean by *"the
  alias should produce a $ref"* / *"the alias should dissolve"* /
  *"the alias should inline"* in this corpus.

This summary feeds straight into `alias-handling.md` §5 once we have
a few cycles' worth.

---

## Cycle 2 — Stdlib-special RHS (fixture: `alias-calibration-stdlib`)

Source: `fixtures/enhancements/alias-calibration-stdlib/types.go`.

Same structural pattern as cycle 1 (annotated decls + Envelope using
them + one unannotated stdlib case for R4 confirmation), but with
RHS types that `applyStdlibSpecials` recognises. Eight cells × three
modes = 24 judgments.

**Hypotheses going in (from cycle 1's R1-R4):**

- **R1 partially breaks.** Stdlib-special types are `*types.Named`,
  so the `$ref`-branch switch in `schema.go:174` has a matching
  case (not the `default:` fallthrough that collapsed Default and
  Ref for primitives). Default and Ref may diverge here.
- **R4 should hold.** Unannotated `SilentTime` should follow the
  same shape as cycle 1's `LegacyID`: present-and-reffed under
  Default/Ref (if strict), absent under Transparent.
- **Q3 visible.** Default mode's `Timestamp` definition is the bug
  we already empirically confirmed — should emit `{type: object}`
  rather than `{type: string, format: date-time}`.

### 2.1 Decl-as-model × `Timestamp = time.Time`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | pre-patch: `{type: object, title, description}` ✗ Q3; **post-patch (`36b99a2`)**: `{type: string, format: date-time, title, description, x-go-package}` ✓ inline correct | ✓ post-patch | | R5-scope patch resolves Q3 family in Expand mode — recognizer consulted before Underlying walk |
| RefAliases=true | `{$ref: #/definitions/Time, title, description}` — 2-hop chain through `Time` def which carries the canonical shape | ✓ on shape correctness, ? on chain-vs-inline | (Q-C below) | Recognizer IS consulted on this path (via Time's own buildNamedType run). End result correct after 2 hops. |
| TransparentAliases=true | `{type: string, format: date-time, title, description}` — **INLINE CORRECT** | ✓ | | Recognizer applied directly on the alias's RHS. No chain. No `Time` def emitted. The cleanest of the three modes for this RHS. |

### 2.2 Decl-as-model × `Err = error`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | pre-patch: `{type: object, description}` ✗; **post-patch (`36b99a2`)**: `{type: string, x-go-type: error, description, x-go-package}` ✓ inline correct | ✓ post-patch | | R5-scope patch fix |
| RefAliases=true | `{type: string, description}` inline (extensions skipped in TUI) | ✓ | | Now correct post-patch (`4c76644`). Predeclared types route through the new nil-pkg guard → `applyStdlibSpecials.recognizeError` fires. |
| TransparentAliases=true | `{type: string, description}` — INLINE (same as post-patch Ref) | ✓ | | Same recognizer hit via the Transparent dissolve path. |

### 2.3 Decl-as-model × `Raw = json.RawMessage`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | pre-patch: `{type: array, items: {integer, uint8}, title, description}` ✗; **post-patch (`36b99a2`)**: `{title, description, x-go-package}` open envelope ✓ recognizer-canonical (Q13 family bare-envelope idiom) | ✓ post-patch | | R5-scope patch — Raw now matches recognizer's "any JSON" output. Q13 family ambiguity (bare envelope) remains but is the canonical idiom. |
| RefAliases=true | `{$ref: #/definitions/RawMessage, title, description}` — 2-hop chain through `RawMessage` def which has `{title, description}` only (open shape via `recognizeRawMessage`) | ✓ on shape correctness via chain, ? on chain-vs-inline (Q-C); ? on bare-envelope shape (Q13 family — the `{title, description}` wrapper around nothing is ambiguous "any JSON" idiom) | (Q-C and Q13 family) | Same chain pattern as Timestamp. End result is the recognizer's open shape, two hops deep. |
| TransparentAliases=true | `{title, description}` — INLINE, but bare envelope (no type, no format) | ✓ on dissolve-and-recognize, ? on bare-envelope (Q13 family) | | The bare-envelope shape is the recognizer's "open" idiom (target.Schema() with nothing else). Same Q13 ambiguity as `any` and `LegacyID`-but-stronger. |

### 2.4 Decl (auto-discovered) × `SilentTime = time.Time`

R4 test for stdlib RHS.

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | pre-patch: `{type: object, description}` ✗; **post-patch (`36b99a2`)**: `{type: string, format: date-time, description, x-go-package}` ✓ inline correct, present per R4-strict | ✓ post-patch on both shape and presence | | R5-scope patch fixes Q3; R4-strict (unannotated alias surfaces under Default) preserved |
| RefAliases=true | `{$ref: #/definitions/Time, description}` — same chain as annotated Timestamp | ✓ on chain target, ? on chain-vs-inline | (Q-C) | R4 holds across RHS kinds: unannotated alias still gets def + $ref under Ref. Shape via the recognised Time chain. |
| TransparentAliases=true | **ABSENT** from `definitions` | ✓ | | R4 confirmed across stdlib RHS — unannotated stdlib alias dissolves under Transparent same as primitive `LegacyID` did in cycle 1. No dangling. |

### 2.5 Field × `At Timestamp`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `{$ref: #/definitions/Timestamp}` | ✓ on $ref shape, ✗ on what it points to | $ref is right per R2; the dereferenced target is wrong per 2.1 | field-reach behavior consistent with primitive cycle 1 (annotated → $ref); badness is in the target |
| RefAliases=true | `{$ref: #/definitions/Timestamp}` — same as Default; field-reach not mode-sensitive when annotated | ✓ on $ref, ✓ on dereferenced target (Timestamp now chains to Time which is correct) | | |
| TransparentAliases=true | `{description, type: string, format: date-time}` — INLINE, godoc preserved | ✓ | | Field-level godoc preserved (R3 trade) + canonical recognizer shape. The cleanest of the three modes for this field. |

### 2.6 Field × `Failure Err`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `{$ref: #/definitions/Err}` | ✓ on shape | | depends on 2.2 resolution |
| RefAliases=true | `{$ref: #/definitions/Err}` | ✓ on $ref, ✓ on target (Err inline string per the patch) | | |
| TransparentAliases=true | `{description, type: string}` — INLINE, godoc preserved (no x-go-type per skipExt) | ✓ | | Same recognizer outcome as 2.5. |

### 2.7 Field × `Payload Raw`

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `{$ref: #/definitions/Raw}` | ✓ on $ref shape, ✗ on target | | target shape per 2.3 |
| RefAliases=true | `{$ref: #/definitions/Raw}` | ✓ on $ref, ✓-ish on target (Raw chains to RawMessage which is open shape; Q13 ambiguity) | | |
| TransparentAliases=true | `{description}` ONLY — no type, no format. Just the field-level godoc and nothing else | ✓ on dissolve, ? on bare-payload (Q13 family — RawMessage's open shape with field godoc reduces to bare description) | (Q13) | The bare `{description}` shape is the recognizer's open-schema idiom (target.Schema() empty) plus the field godoc. Functionally `{}` per OAS2 ("any value") with documentation. Valid spec but visually unusual. |

### 2.8 Field × `Silent SilentTime` (unannotated stdlib RHS)

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `{$ref: #/definitions/SilentTime}` — reffed from Envelope to the auto-discovered def | ✓ on $ref-vs-inline (strict R4), ✗ on target | | confirms R4 strict reading; target shape per 2.4 |
| RefAliases=true | `{$ref: #/definitions/SilentTime}` — same as Default; R4 holds across modes | ✓ on $ref, ✓ on target (SilentTime chains to correctly-shaped Time def) | | |
| TransparentAliases=true | `{description, type: string, format: date-time}` — INLINE, identical to annotated 2.5 | ✓ | | The mode dissolves the alias regardless of annotation; the canonical recognizer shape lands at the field site. R4 holds: no SilentTime def appears. |

### 2.9 Cycle 2 summary

All 24 cells captured (modulo the post-patch Ref re-run). The
findings, in order of how load-bearing they are:

**1. R1 breaks for stdlib RHS, as predicted — and the patch landed
mid-cycle confirmed why.** `*types.Named` RHS doesn't take the
`default:` fallthrough that collapsed Default and Ref for
primitives; it has a matching case in the `$ref`-branch switch.
That case originally lacked nil-guards and special-recognizer
checks, producing the panic on `error` and the chain shape for
time.Time / json.RawMessage. Patch `4c76644` added the nil-guard;
the recognizer-bypass on the Expand path is still open (R5-scope).

**2. R4 holds across RHS kinds (primitive, stdlib).** `SilentTime`
(unannotated `time.Time` alias) behaves identically to `LegacyID`
(unannotated `int64` alias) under each mode. Under Transparent both
are absent; under Default/Ref both are present-and-reffed. The
annotation gates definitions; mode controls field-reach behaviour.

**3. R3 holds across RHS kinds.** `$ref`-ifying a field loses the
field-level godoc; inlining preserves it. The cycle 1 trade ("named
entity OR field doc, not both without DescWithRef") applies
uniformly to stdlib aliases too.

**4. Q3 generalises to the full applyStdlibSpecials family — R5.**
Default mode's wrongness for Timestamp, Err, Raw, SilentTime all
trace to the same `buildDeclAlias` Expand branch missing an
`applyStdlibSpecials` call. The one-line patch in `fix-quirks.md`
C3 covers all four manifestations.

**5. Transparent mode is the cleanest of the three for stdlib
aliases.** It applies the recognizer correctly (via the dissolve
path's `buildFromType(rhs, target)`); it dissolves unannotated
aliases per R4; field-level godoc survives at use sites. The only
visible "ugliness" is the bare-envelope shape for `json.RawMessage`
(`{description}` only) — but that's the recognizer's "open shape"
idiom, the Q13 ambiguity, not a mode bug.

**6. New open question — Q-C.** Post-patch, the three modes now
diverge for stdlib aliases:
- Default: still broken (R5-scope)
- Ref: chain (`Timestamp` → `Time`, `Raw` → `RawMessage`)
- Transparent: inline correct

Should Default + Ref converge to Transparent's inline shape (collapse
to one decl, no chain)? Or is the Ref chain genuinely valuable —
preserving a named entity for code generators to typedef? Same Q11
chain question, but specifically for stdlib aliases.

**Q30 parked.** Stdlib godocs leaking into definition descriptions
under Ref's chain (`Time` carries the full time.Time godoc) — Fred:
"correct but noisy, pre-existing." Logged in `observed-quirks.md`
Q30, not a workshop-blocker.

**Vocabulary calibrated for cycle 3:**
- "Inline" = the recognizer's canonical shape sits directly in the
  alias's own definition or field.
- "Chain" = the alias's definition is a `$ref` to a separately-
  built definition that carries the canonical shape (one or more
  hops).
- "Dissolved" = no definition for the alias; the use site emits
  the recognizer's shape directly.
- "Q3 family" = `applyStdlibSpecials` bypassed; structural walk of
  Underlying produces a wrong shape (empty object, byte array,
  etc.).
- "R4 strict" vs "R4 strong" = current Q-B framing for unannotated
  aliases.

---

## Cycle 3 — Embed composition (fixture: `alias-calibration-embed`)

Source: `fixtures/enhancements/alias-calibration-embed/types.go`.

Five outer structs, each embedding the same `Base` (or its alias /
pointer / interface) in a different way. The central question: does
the Q8 asymmetry (alias-embed = allOf+$ref; everything else = flat
inline) still hold post-cycle-1+2 patches, and if so, what should it
be?

**Hypotheses going in (from B2 probe + cycles 1-2):**

- **Q8 unchanged.** Our patches all touched `buildDeclAlias` and the
  `*types.Named` branch of its `$ref` switch. None touched
  `buildEmbedded` / `buildNamedEmbedded`. So the embed dispatch
  should produce the same shapes B2 observed: FLAT for direct named
  embed, allOf+$ref for alias embed.
- **Field-site Base-vs-alias parity.** Cycles 1-2 showed annotated
  alias produces `$ref` at field sites (same as plain annotated
  named type). `Envelope.Direct Base` and `Envelope.ViaAlias
  BaseAlias` should both emit `{$ref: #/definitions/Base}` since
  Base and BaseAlias are the same type.

### 3.1 Decl-as-model × `Base` (the embedded target)

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `{type: object, properties: {id, name}, required: [id]}` | ✓ | | stable named struct, mode-invariant |
| RefAliases=true | identical to Expand | ✓ | | |
| TransparentAliases=true | identical to Expand | ✓ | | |

### 3.2 Decl × `EmbedsDirectStruct` (Base direct embed)

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | **FLAT** `{type: object, properties: {id, name, extra}, required: [id]}` | ✓ | | Base's fields inlined; mode-invariant |
| RefAliases=true | identical FLAT | ✓ | | |
| TransparentAliases=true | identical FLAT | ✓ | | |

### 3.3 Decl × `EmbedsAlias` (BaseAlias embed) — **Q8 confirmed across all modes**

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `allOf: [{$ref: BaseAlias}, {extra inline}]` | ? Q-D | (Q-D) | Q8 OUTLIER. Different shape from EmbedsDirectStruct (3.2) even though Base and BaseAlias are the same Go type. |
| RefAliases=true | identical to Expand | ? Q-D | | |
| TransparentAliases=true | `allOf: [{$ref: Base}, {extra inline}]` — BaseAlias dissolved, but allOf composition PERSISTS | ? Q-D | | **Embed-site asymmetry exists in ALL three modes.** Transparent normalizes the $ref target (Base instead of BaseAlias) but the structural allOf vs flat distinction persists. |

### 3.4 Decl × `EmbedsPointer` (`*Base` embed)

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | FLAT `{id, name, extra}` | ✓ | | pointer peeled → named-direct path → flat |
| RefAliases=true | identical FLAT | ✓ | | |
| TransparentAliases=true | identical FLAT | ✓ | | |

### 3.5 Decl × `EmbedsInterface` (Methods embed)

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | FLAT `{describe, tag}` — method `Describe` promoted as property | ✓ | | named-interface path → flat |
| RefAliases=true | identical FLAT | ✓ | | |
| TransparentAliases=true | identical FLAT | ✓ | | |

### 3.6 Decl × `Envelope` (Base vs BaseAlias field-site comparison)

| Mode | Current shape (TUI) | Judgment | Desired shape | Note |
|---|---|---|---|---|
| Expand (default) | `direct: {$ref: Base}`, `viaAlias: {$ref: BaseAlias}` — **field-site asymmetry leaks** | ? Q-E | (Q-E) | Same Go type, different $ref targets. The aliasing IS visible at the field site under Default. |
| RefAliases=true | identical to Expand | ? Q-E | | |
| TransparentAliases=true | `direct: {$ref: Base}`, `viaAlias: {$ref: Base}` — **field-site asymmetry resolved** | ✓ | | Transparent dissolves BaseAlias at use sites → both fields converge on Base. |

### 3.7 BaseAlias decl (unannotated, reached via field + embed)

| Mode | Current shape (TUI) | Judgment | Note |
|---|---|---|---|
| Expand (default) | `{type: object, properties: {id, name}, required: [id]}` — **FULL COPY of Base** | ? Q-F | Expand walks Underlying → emits structural mirror = entire Base shape duplicated. Q11 chain-duplication applied to a 1-link alias. |
| RefAliases=true | `{$ref: Base, description}` — chain | ✓ | The Q-C fix's pattern: alias chains to its target via $ref, no structural duplication. Cleaner than Default. |
| TransparentAliases=true | **ABSENT** | ✓ | R4 strict holds for user-struct alias: unannotated alias dissolved under Transparent. |

### 3.7 Side observations to watch

- Whether `BaseAlias` appears in `definitions` as its own entry, or
  is dissolved (relevant in Transparent mode, and in EmbedsAlias's
  allOf $ref target under Expand/Ref).
- Whether `Methods` appears in `definitions`.
- Whether `Base` def shape is identical across modes (it's an
  annotated named struct, the "boring" case — should be stable).

### 3.8 Cycle 3 summary

Three findings, three open questions.

**1. Q8 confirmed across ALL three modes — and B2's framing was right.**
EmbedsDirectStruct produces flat, EmbedsAlias produces allOf+$ref, in
every mode. Transparent dissolves the *target* of the allOf $ref
(Base instead of BaseAlias) but the **structural composition itself
persists** — EmbedsAlias remains allOf-shaped while
EmbedsDirectStruct remains flat. The asymmetry is baked into
`buildEmbedded`'s dispatch: `*types.Named` arm calls
`buildNamedEmbedded` which inlines via `buildFromStruct`;
`*types.Alias` arm calls `buildAlias` which emits a $ref that the
outer struct treats as an allOf member.

→ **Q-D — should EmbedsAlias inline (match EmbedsDirectStruct)?**

**2. Field-site asymmetry leaks under Default/Ref, resolved under
Transparent.** `Envelope.direct` (typed `Base`) and
`Envelope.viaAlias` (typed `BaseAlias`) produce **different** $ref
targets under Default/Ref (`Base` vs `BaseAlias`), even though both
fields hold the same Go type. Under Transparent both resolve to
`{$ref: Base}`. The user wrote `BaseAlias` as a deliberate name in
source — the question is whether downstream codegen should see "two
different types" or "one type, two names."

→ **Q-E — should field-site $ref point to the alias's terminal
target, or preserve the user-written alias name?**

**3. Default-mode BaseAlias is a full COPY of Base.** Expand mode
walks Underlying and emits BaseAlias as `{type: object, properties:
{id, name}, required: [id]}` — a complete structural duplicate of
Base. Ref mode produces the chain shape `{$ref: Base}`; Transparent
dissolves entirely. The Expand-mode copy is Q11 chain-duplication
applied to a 1-link alias — the layer doesn't add information, it
just duplicates the target's content.

→ **Q-F — should Default mode collapse BaseAlias to a chain shape
(match Ref) or to inline-target-shape (no separate def)?**

**Where the asymmetries trace to in code:**

| Question | Code site | Fix shape |
|---|---|---|
| Q-D (embed asymmetry) | `embedded.go:46` — `*types.Alias` arm calls `buildAlias` | Replace with structural inline: walk Underlying of the alias and treat as struct embed |
| Q-E (field asymmetry) | `schema.go:289` `buildAlias` calls `MakeRef(alias-decl)` | MakeRef the unaliased target instead — `types.Unalias(rhs).(*types.Named).Obj()` |
| Q-F (Default-mode copy) | `schema.go:168` Expand fallthrough `buildFromType(Underlying)` | For `Alias-to-Named` specifically: emit `$ref` to the Named target instead of walking Underlying |

All three would converge alias-embed/field/decl behaviour with the
direct-named equivalent — eliminating the alias asymmetry entirely.
Whether that's the right call is the workshop's design judgment.

---

## Running pattern table

As ✓/✗ judgments accumulate, summarize here the candidate rules and
which fixtures support / refute each. This is where the alias model
emerges from the corpus.

| Candidate rule | Supports | Refutes | Status |
|---|---|---|---|
| **R1 — Modes collapse for `*types.Basic` RHS at the decl level.** All three modes produce byte-identical definitions for `swagger:model X = primitive`. The Expand and Ref branches converge on `buildFromType(rhs, target)`; Transparent goes there directly. | Cycle 1: 1.1-1.3 across all three modes | (none yet) | observed |
| **R2 — `swagger:model` forces decl-level registration regardless of mode.** Even Transparent mode keeps `UserID` / `Name` / `Active` in `definitions` — the annotation triggers orchestrator-level discovery independently of `buildDeclAlias`'s "Dissolve" intent. Transparent effectively only affects FIELD (and presumably embed/element) reach context, not annotated decls. | Cycle 1: 1.1-1.3 Transparent rows show the definitions present | (hypothesis: removing the annotation would let Transparent dissolve — untested) | observed at decl level; field path remains the actual mode-divergence point |
| **R3 — `$ref` to a primitive-alias definition loses field-level godoc.** When `Expand` / `Ref` mode emits `{$ref: #/definitions/UserID}` for a field, the field's own godoc description and `x-go-name` extension are dropped. Transparent mode (inline) preserves both. This is the `DescWithRef=false` consequence (the JSON Reference spec forbids siblings of `$ref`). | Cycle 1: 1.4-1.6 Expand vs Transparent — the description / x-go-name appear only in Transparent rows | (`DescWithRef=true` would emit allOf wrapping — untested in this cycle) | observed; informs judgment of whether $ref is desirable for primitive aliases |
| **R1+R2+R3 derivation.** For `swagger:model X = primitive` + field usage of X, the user faces an unavoidable trade today: either get a named definition that downstream code-generators can typedef (`UserID` as its own Go type) but lose field-level documentation, OR keep field documentation but lose the named entity (Transparent dissolves it at use sites). The two valuable things — named entity AND field docs — cannot coexist without `DescWithRef=true`. | (derived) | | hypothesis |
| **R4 — Annotation is the gatekeeper of `definitions` entries.** (Fred, cycle 1.) Aliases without `swagger:model` do not surface as independent definitions, even when referenced. Field references to them are inlined at the use site instead. Cleaner formulation candidate: *"annotated → definition + $ref at use sites; unannotated → no definition; use site inlines via Underlying."* | Cycle 1 Transparent capture: `LegacyID` / `LegacyName` absent from `definitions` despite being reachable via `Envelope.legacy` / `.legacyNick` | (need Default and Ref captures to confirm the rule holds across all three modes — the Transparent capture is consistent with the rule but doesn't rule out Default/Ref behaving differently) | observed under Transparent only; pending confirmation |
| **R4-implication (Q-A consequence).** If R4 holds, the role of the modes shrinks: annotation determines whether a definition exists; modes only control field-reach behaviour. Then Transparent's current shape (annotated decl present + field inline = dangling annotated def) is the only mode-shape that violates R4's "no dangling" half. Resolving Q-A means deciding whether the annotation always wins at field sites (kill the dangling), or whether Transparent gets to override annotation (keep current shape, accept dangling as user-requested). | (R4-derived) | | hypothesis |
| **R5 — Special types are not recognised when hidden behind an alias.** (Fred, cycle 2.) `applyStdlibSpecials` is called when the schema builder encounters `time.Time` / `error` / `json.RawMessage` / `any` directly, but NOT when those same types are reached via an alias decl in Expand mode (`buildDeclAlias`, `schema.go:168-170`). The Expand branch walks `Underlying()` and produces a structural shape (empty object, byte-array, etc.) rather than the recognizer's canonical shape (`string+date-time`, `string`, `{}`, etc.). | Cycle 2 Default capture: all four stdlib aliases (Timestamp, Err, Raw, SilentTime) emit wrong shapes; each matches what walking `Underlying()` would produce for that stdlib type | (predicted: Ref and Transparent modes do consult `applyStdlibSpecials` via `buildNamedType` and the `$ref`-branch's `case *types.Named` → `MakeRef`. So R5 is Expand-only.) | observed for Expand; mode scope pending Ref/Transparent captures |
| **R5-fix.** Same one-line patch already proposed for Q3/Q13 in `fix-quirks.md` C3: consult `applyStdlibSpecials` before the `buildFromType(Underlying, target)` fallthrough in `buildDeclAlias`. Now empirically validated as covering the *whole* applyStdlibSpecials family (time / error / RawMessage / any), not just time.Time + any. | (derived from R5) | | known fix |
| **R5-scope (post-patch).** Originally open (Expand-only patch was needed). **CLOSED `36b99a2`.** `buildDeclAlias` Expand branch now consults `applyStdlibSpecials` via `rhsTypeName(rhs)` before walking Underlying. All four stdlib aliases (Timestamp / Err / Raw / SilentTime) produce their canonical recognizer shapes under Default mode, matching what Transparent already emits. | Cycle 2 post-patch captures (Fred 2026-06-04): three modes all return correct shapes | | closed |
| **Q-C — chain vs inline for stdlib aliases under Ref mode.** Fred's 2026-06-04 judgment on the patched state: **"Ref looks good IMHO."** Accepting the within-mode asymmetry — predeclared types inline (Err = `{type: string, x-go-type: error}`), packaged stdlib types chain (Timestamp → Time, Raw → RawMessage, SilentTime → Time). The chain shape preserves the named-entity identity at the cost of Q30 godoc-noise on the chain target. Decision: status quo. | Cycle 2 Ref capture; user judgment | | accepted (current shape stands) |
| **R6 — Annotation gates first-class spec entity at use sites (schema builder).** (Fred + workshop, cycle 3, 2026-06-10.) Generalises R4 from Transparent-only to all modes. An unannotated alias dissolves at every field / element / allOf-member use site, and never produces a `definitions` entry. An annotated alias (`swagger:model`) keeps its identity at use sites and carries its own definition; `TransparentAliases` still supersedes (dissolves at use sites regardless of annotation), and `swagger:model` still forces decl-level registration (R2 holds — annotated decl present even under Transparent). | Q-D / Q-E closeout; cycle-3 bidirectional witnesses (Envelope vs EnvelopeAnnotatedAlias; EmbedsAliasOptIn vs EmbedsAliasModeledOptIn); commits `ba22e06`, `c86578f` | | closed (schema layer) |

---

## Cycle 4 — parameters builder (kickoff capture, 2026-06-10)

Pre-patch state captured by
`TestCoverage_AliasParametersCalibration_{Default,Ref,Transparent}`
against `fixtures/enhancements/alias-parameters-calibration/`.
Goldens written to `enhancements_alias_parameters_calibration_*.json`.

### 4.1 Snapshot per mode

`paths` populates correctly under all three modes (two routes,
two operations); the issue is in `definitions` and in body-field
$ref targets.

| Mode | Body $refs (all body fields, including annotated) | Definitions pollution |
|---|---|---|
| Default | ALL → `$ref: Payload` (annotation has no effect; even `BodyAliasModeled` collapses) | `AliasedTopParams`, `internalParams`, `PayloadAlias`, `PayloadAlias2`, `QueryIDAlias` |
| Ref | `BodyAliasChain` → `$ref: PayloadAlias` (one chain step); rest → `Payload` | same + `QueryID` (now also leaks from the non-body chain) |
| Transparent | ALL → `$ref: Payload` (Transparent behaviour as expected) | only `Payload`, `PayloadAliasModeled` — CLEAN |

### 4.2 First observations (pre-judgment)

1. **Top-level `swagger:parameters` alias leaks under Default and
   Ref.** Both `AliasedTopParams` and the unexported
   `internalParams` surface in `definitions`. R7 clause 1 says
   they should not. Transparent gets this right because the alias
   dispatch dissolves before AppendPostDecl runs.

2. **`swagger:model` on a body-field alias has NO effect under
   Default / Ref.** `BodyAliasModeled` typed `PayloadAliasModeled`
   resolves to `$ref: Payload`, the same as the unannotated
   `BodyAliasPlain`. Worse than schema's pre-R6 state — the schema
   builder at least preserved the alias name. The parameters
   builder unconditionally expands the alias regardless of
   annotation.

3. **The 2-link chain `PayloadAlias2 = PayloadAlias = Payload`
   shows a one-step shape under Ref:** `$ref` lands on the
   *first* chain target (`PayloadAlias`), not the unaliased
   target (`Payload`). The schema-layer R6 patch made this
   irrelevant for the schema builder (chain dissolves entirely),
   but the parameters builder still walks one chain step.

4. **`QueryIDAlias` (non-body, SimpleSchema target) correctly
   inlines** under all three modes — the `In() != body` branch in
   `buildFieldAlias` already does the right thing for the
   non-body field. The only issue is the dangling `QueryIDAlias`
   definition (Default / Ref) which R7 clause 3 would remove.

5. **The Q12 leak is universal across non-Transparent modes.**
   `internalParams` (unexported backing struct) and `QueryID`
   (named primitive, surfacing under Ref only) both leak because
   the parameters builder's `AppendPostDecl(decl)` calls run
   unconditionally — no annotation gate.

### 4.3 Pending Fred-side judgment (TUI side-by-side)

- Confirm R7 clause 1: visually verify that `paths` content under
  Transparent is the desired shape, and that Default / Ref should
  match it modulo the alias-decl semantics.
- Confirm R7 clause 2: visually verify that `BodyAliasModeled`
  should produce `$ref: PayloadAliasModeled` (annotated → preserve
  identity) post-patch under Default / Ref, matching the cycle-3
  field-site result.
- Q-G — the dissolve depth question for the top-level case (does
  `internalParams` exist as a step at all?). Pre-patch captures
  suggest the dissolve walks both layers; confirm acceptable.
- Q-H — chain dissolve depth for body fields (full dissolve like
  R6, or one-step like current Ref?). Hunch is full dissolve,
  matching R6.

### 4.3a Q-F + Q11 closure (post-R6/R7/R8 reflection, 2026-06-10)

After the R6/R7/R8 wave landed, two questions that had been kept
on the docket from cycle 3 turn out to need no patch:

- **Q-F (annotated alias decl shape under Default).** Originally
  framed pre-R6 when ALL aliases — annotated or not — produced
  the structural copy under Default. R6 closed the unannotated
  case entirely (dissolve). For annotated aliases, the
  Default/Ref/Transparent shapes are just the documented mode
  semantics × R2 doing what they should. **Closed, no action.**
- **Q11 (annotated chain duplication).** Same shape of question
  for N-link annotated chains. Same answer: each annotated link is
  a user opt-in; the per-mode decl shape is consistent with the
  documented contract. The Default-mode N× duplication is real
  but is the price of three named entities; the user-visible
  escape valve is `RefAliases=true`. **Closed, no action.**

Both close as `consistent with documented mode behaviour`. Same
disposition as Q-A, Q-B, Q-C — accepted shape, not patched.

### 4.4 R7 candidate (to be confirmed by the walk)

The candidate rule on the canvas, stated in patch terms — these
are the gates I expect to add in `parameters.go`:

- `buildAlias` (top-level): drop both `AppendPostDecl(alias-decl)`
  and `AppendPostDecl(rhs-decl)`. Resolve the field-bearing
  struct via Unalias and process its fields. Neither the alias
  nor any chain link surfaces as a definition.
- `buildFieldAlias` (field-level):
  - Drop the unconditional `AppendPostDecl(alias-decl)`.
  - Gate on `decl.HasModelAnnotation()`: unannotated → dissolve
    via `Unalias` + `buildFromField`; annotated → MakeRef the
    alias.
  - Mode handling unchanged otherwise (Transparent supersedes;
    Ref / Default share the gate at the start).

The schema-side R6 patch was one line; R7 should land in roughly
the same shape, applied in two places.
