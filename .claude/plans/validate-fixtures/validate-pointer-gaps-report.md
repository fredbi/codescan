# go-openapi/validate: pointers that address nothing, and findings with no address at all

Observed on **v0.26.2** (`83d7d4e`). This is a follow-up to the map-order report and
covers a different problem: `Located.Pointer` is documented as an RFC 6901 pointer
*relative to the validated document*, and in four families of finding it is not one.

Every example below is a complete, minimal Swagger 2.0 document. For each I give
what `Located.Pointer` returns and what the document actually addresses. Where the
pointer resolves against its own document I mark it `ok`; where it addresses nothing,
`!!`.

## Why this needs its own report

The `required` work in #279 is right, and at scale it is right in the only way my
corpus could show. Across **601 generated specs / 774 findings**, every finding is
as specific as its own message, and all **568** root pointers are the single shape
`info in body is required` — a document with no `info` at all, where the document
*is* the holder. Nothing regressed and nothing was lost.

That is also why the cases below were invisible to me until I went looking. My
corpus is codescan's output, which never emits a bad `default`, a path template
without a parameter, or a duplicate `operationId` — so a consumer's fixtures and
validate's fixtures can both be green while these stay broken. Two of the four
families are the *most common authoring mistakes there are*.

I should also correct my own earlier measurement: I scored the empty pointer as
"resolves exactly", because the root always resolves. That hid family D entirely.

## Family A — a value-level finding is rooted at the value, not at the document

When a `default` (or a header default) is checked against its schema, the inner
result's pointers escape into the outer result unprefixed. The outer finding is
located correctly; the inner one, which is the specific one, is not.

### A1 — no prefix at all: the pointer is the parameter's own name

```json
{"swagger":"2.0","info":{"title":"t","version":"1"},
 "paths":{"/a":{"get":{"operationId":"a",
   "parameters":[{"name":"zzz","in":"query","type":"integer","default":"nope"}],
   "responses":{"200":{"description":"ok"}}}}}}
```

```
ok  /paths/~1a/get/parameters/0     default value for zzz in query does not validate its schema
!!  /zzz                            zzz in query must be of type integer: "string"
        expected: /paths/~1a/get/parameters/0/default
```

The parameter is deliberately named `zzz` to show the pointer is built from the
parameter name. Same shape for a response header default:

```json
{"swagger":"2.0","info":{"title":"t","version":"1"},
 "paths":{"/a":{"get":{"operationId":"a","responses":{"200":{"description":"ok",
   "headers":{"X-Count":{"type":"integer","default":"nope"}}}}}}}}
```

```
ok  /paths/~1a/get/responses/200/headers/X-Count   default value in header X-Count … does not validate
!!  /X-Count                                       X-Count in header must be of type integer: "string"
        expected: /paths/~1a/get/responses/200/headers/X-Count/default
```

### A2 — prefixed, but the body parameter's `schema` token is missing

`db68410` introduced structural tokens for the response side. The body-parameter
schema does not get one, so the pointer descends into the parameter object as
though `properties` were a member of it:

```json
{"swagger":"2.0","info":{"title":"t","version":"1"},
 "paths":{"/a":{"post":{"operationId":"a",
   "parameters":[{"name":"payload","in":"body",
     "schema":{"type":"object","properties":{"n":{"type":"integer","default":"nope"}}}}],
   "responses":{"200":{"description":"ok"}}}}}}
```

```
ok  /paths/~1a/post/parameters/0                             default value for payload in body does not validate
!!  /paths/~1a/post/parameters/0/properties/n/default         paths./a.post.parameters.payload.n.default in body must be of type integer…
        expected: /paths/~1a/post/parameters/0/schema/properties/n/default
```

## Family B — malformed path pointer: the template variable used as the path key

```json
{"swagger":"2.0","info":{"title":"t","version":"1"},
 "paths":{"/deep/nested/{ghostvar}":{"get":{"operationId":"a",
   "responses":{"200":{"description":"ok"}}}}}}
```

```
!!  /paths/{ghostvar}      path param "{ghostvar}" has no parameter definition
        expected: /paths/~1deep~1nested~1{ghostvar}
```

The variable name is used where the path key belongs, and it is not `~1`-escaped.
The distinctive name shows which token was taken. This one looks like a small fix:
the finding already knows the path it is walking.

## Family C — a definite site exists, but the finding carries no address

Both of these report at the root, where the document is not the subject.

### C1 — duplicate operationId

```json
{"swagger":"2.0","info":{"title":"t","version":"1"},
 "paths":{"/a":{"get":{"operationId":"dup","responses":{"200":{"description":"ok"}}}},
          "/b":{"get":{"operationId":"dup","responses":{"200":{"description":"ok"}}}}}}
```

```
ok(root)  «root»      "dup" is defined 2 times
        expected: one of the offending operations, e.g. /paths/~1b/get/operationId
```

Both sites are known — the count comes from having seen them. A consumer that
navigates to a finding can do nothing with the document root here.

### C2 — unresolvable `$ref`

```json
{"swagger":"2.0","info":{"title":"t","version":"1"},
 "paths":{"/a":{"get":{"operationId":"a","responses":{"200":{"description":"ok",
   "schema":{"$ref":"#/definitions/Missing"}}}}}}}
```

```
ok(root)  «root»   some references could not be resolved in spec. First found: nil value has no key
        expected: /paths/~1a/get/responses/200/schema/$ref
```

"First found" suggests the others are collapsed into one message too.

## Family D — a fault reached through a `$ref` (already known)

Recorded for completeness, since it is the one case your own commit message calls
out as not covered. It is also the mildest, because the same fault is reported a
second time at a resolvable location.

```json
{"swagger":"2.0","info":{"title":"t","version":"1"},
 "paths":{"/a":{"get":{"operationId":"a","responses":{"200":{"$ref":"#/responses/shared"}}}}},
 "responses":{"shared":{"description":"ok","schema":{"$ref":"#/definitions/A"}}},
 "definitions":{"A":{"type":"object","default":{"n":"not-an-object"},
   "properties":{"n":{"type":"object"}}}}}
```

```
ok  /paths/~1a/get/responses/200                    in operation "a", default value in response 200 does not validate
!!  /paths/~1a/get/responses/200/schema/default/n   …default.n in body must be of type object: "string"
ok  /definitions/A/default/n                        definitions.A.default.n in body must be of type object: "string"
```

`/paths/~1a/get/responses/200` is a bare `$ref`, so nothing exists below it.

## Adjacent: two checks that never fire

Not pointer problems — but they are in the `required` family you have just been
working in, so worth knowing while the fixtures are open. Both produce **no finding
at all**.

### The required-vs-declared check does not descend into nested schemas

Reported at the definition's top level (`/definitions/A/required/0`), but not for a
property's own schema:

```json
{"swagger":"2.0","info":{"title":"t","version":"1"},
 "paths":{"/a":{"get":{"operationId":"a","responses":{"200":{"description":"ok",
   "schema":{"$ref":"#/definitions/A"}}}}}},
 "definitions":{"A":{"type":"object","properties":{
   "inner":{"type":"object","required":["ghost"],"properties":{"real":{"type":"string"}}}}}}}
```

→ no findings. Moving `required:["ghost"]` up to `A` reports it at
`/definitions/A/required/0`.

### A security requirement naming an undeclared scheme

```json
{"swagger":"2.0","info":{"title":"t","version":"1"},
 "paths":{"/a":{"get":{"operationId":"a","security":[{"nosuch":[]}],
   "responses":{"200":{"description":"ok"}}}}}}
```

→ no findings, with no `securityDefinitions` in the document at all. If this is
deliberate, ignore it.

## What already works — a conformance matrix worth keeping as fixtures

Every one of these resolves against its own document on v0.26.2, and they are the
cases the #279 work fixed. Offered as a ready-made grid, since it is the part your
fixtures may not pin yet:

| Missing / faulty | Pointer returned |
|---|---|
| top-level `info` | `«root»` |
| top-level `paths` | `«root»` |
| `info.title`, `info.version` | `/info` |
| `info.license.name` | `/info/license` |
| `externalDocs.url` | `/externalDocs` |
| `tags[0].name` | `/tags/0` |
| response `description` | `/paths/~1a/get/responses/200` |
| shared response `description` | `/responses/shared` |
| operation `responses` | `/paths/~1a/get` |
| body param `schema` / non-body param `type` | `/paths/~1a/…/parameters/0` |
| param `name`, `in` | `/paths/~1a/get/parameters/0` |
| shared param `type` | `/parameters/shared` |
| path-item-level param `type` | `/paths/~1a/parameters/0` |
| array param `items` | `/paths/~1a/get/parameters/0` |
| nested array `items` | `/paths/~1a/get/parameters/0/items` |
| response header `type`, `items` | `/paths/~1a/get/responses/200/headers/X` |
| response `schema.items` | `/paths/~1a/get/responses/200/schema` |
| definition property `items` | `/definitions/A/properties/list` |
| `additionalProperties.items` | `/definitions/A/additionalProperties` |
| `allOf` member property `items` | `/definitions/A/allOf/0/properties/l` |
| `securityDefinitions.k.type` / `.name` / oauth2 `flow` | `/securityDefinitions/k` |
| `required` entry naming an undeclared property | `/definitions/A/required/0` |
| duplicate properties via `allOf` | `/definitions/Child` |
| path param not in template | `/paths/~1a` |
| duplicate params in an operation | `/paths/~1a/get/parameters` |
| more than one body param | `/paths/~1a/post` |
| overlapping path templates | `/paths/~1a~1{y}` |

## How this was measured

Each document above was run through
`validate.NewSpecValidator(doc.Schema(), strfmt.Default).Validate(doc)`, taking
`LocatedErrors()` and `LocatedWarnings()`, and every pointer was resolved against
the document with an independent RFC 6901 walker (`~1`/`~0` unescaped, array indices
by position). `ok` means the walk reached a node; `!!` means it did not. Names like
`zzz`, `X-Count` and `ghostvar` are deliberately distinctive so the token a bad
pointer was built from is unambiguous.
