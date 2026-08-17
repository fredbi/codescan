# go-openapi/validate finding-location fixtures

Minimal Swagger 2.0 documents backing the two reports in the parent directory:

- `../validate-pointer-gaps-report.md` — pointers that address nothing, and findings
  with no address at all
- `../validate-map-order-report.md` — the same document reported differently from one
  run to the next

Each document is as small as it can be while still producing the finding of interest.
Names like `zzz`, `X-Count` and `ghostvar` are deliberately distinctive, so that when
a pointer is built from the wrong token it is obvious which token it took.

## Layout

```
pointers/       50 documents, one finding shape each
determinism/     2 documents, each with the same fault in two definitions
harness/battery  prints every finding's pointer, and whether it resolves
harness/maporder runs one document 50 times and counts distinct outcomes
expected/        captured output, released v0.26.2 vs released v0.26.3
```

`harness/` pins **released `v0.26.3`**, where every case below passes. To
reproduce the original behaviour, downgrade to `v0.26.2`; to test a work in
progress, add a replace:

```sh
cd harness/battery
go mod edit -replace github.com/go-openapi/validate=/path/to/validate
go mod tidy && go run . ../../pointers
```

## Reading the output

```
41-default-param-named-zzz ok E /paths/~1a/get/parameters/0         default value for zzz in query does not validate…
41-default-param-named-zzz !! E /zzz                                zzz in query must be of type integer: "string"
```

`ok` / `!!` is the fixture's own verdict: every pointer is walked against the
document it came from (RFC 6901, `~1`/`~0` unescaped, array indices by position).
`!!` means the pointer addressed no node — which `Located.Pointer` documents as
impossible, being "relative to the validated document". `«root»` is the empty
pointer, which addresses the whole document and therefore always resolves.

**`!!` is the assertion.** A fixture corpus like this is worth having precisely
because it needs no expected-value table to be useful: a pointer either addresses
something or it does not.

## Baselines

| | unresolvable pointers | distinct outcomes over 50 runs |
|---|---|---|
| `expected/pointers-v0.26.2-released.txt` | **7** | — |
| `expected/pointers-v0.26.3-released.txt` | **0** | — |
| `expected/determinism-v0.26.2-released.txt` | — | **2** for each document |
| `expected/determinism-v0.26.3-released.txt` | — | **1** for each document |

v0.26.3 closes both reports: every pointer addresses a node, the two documents
settle to one outcome, and `ContinueOnErrors: true` on the circular pair now
reports **both** definitions instead of whichever the map reached first.

## The 7 that failed on v0.26.2, and where they land on v0.26.3

| Fixture | v0.26.2 returned | v0.26.3 returns (the node specified) |
|---|---|---|
| `21-fault-inside-refd-response` | `/paths/~1a/get/responses/200/schema/default/n` | `/definitions/A/default/n` (the `$ref`'d response holds only `$ref`) |
| `37-template-param-not-declared` | `/paths/{id}` | `/paths/~1a~1{id}` |
| `40-default-fails-schema` | `/q` | `/paths/~1a/get/parameters/0/default` |
| `41-default-param-named-zzz` | `/zzz` | `/paths/~1a/get/parameters/0/default` |
| `43-template-var-named-ghostvar` | `/paths/{ghostvar}` | `/paths/~1deep~1nested~1{ghostvar}` |
| `44-header-default-fails` | `/X-Count` | `…/responses/200/headers/X-Count/default` |
| `45-body-schema-default-fails` | `…/parameters/0/properties/n/default` | `…/parameters/0/schema/properties/n/default` |

Two more report at the root although a definite site exists — they resolve, so the
harness marks them `ok`, and they are worth fixtures all the same:

| Fixture | v0.26.2 returned | v0.26.3 returns |
|---|---|---|
| `28-duplicate-operationid` | `«root»` | `/paths/~1a/get/operationId` |
| `35-invalid-ref-target` | `«root»` | `/paths/~1a/get/responses/200/schema` |

## The nested-`required` check, added in v0.26.3

`29-nested-property-required-undefined` was silent on v0.26.2 and is now reported
at `/definitions/A/properties/inner/required/0` — the entry itself, one level below
the definition.

Fixtures `46`–`49` are the negative cases that guard it, and it passes all of them:
a `required` name arriving through `allOf` or a `$ref` is **not** flagged, and a
name that is plainly declared is not either.

`48-nested-required-with-addlprops` **is** flagged — a `required` name covered only
by `additionalProperties`. That is consistent rather than a false positive: v0.26.2
already flagged the same shape at a definition's top level
(`50-toplevel-required-with-addlprops`), so the new check extends the existing rule
faithfully. Whether the rule itself should excuse `additionalProperties` is a
question about the old check, not the new one.

## Still producing no finding

- `27-security-undefined-scheme` — an operation requires a scheme declared nowhere,
  with no `securityDefinitions` in the document at all. Possibly by design.

## Provenance

Written while consuming `Located.Pointer` from a TUI that navigates to each
finding, so every case here is one a consumer actually hits. The corpus these were
distilled from is codescan's own output — 601 generated specs, 774 findings — which
produces none of the seven above, which is why they went unnoticed on both sides.
