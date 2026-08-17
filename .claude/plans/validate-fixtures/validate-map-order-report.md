# go-openapi/validate: which definition a finding is reported against depends on Go's map iteration order

Observed on **v0.26.2** (`83d7d4e`). Long-standing: both sites below are unchanged
on `master` before the JSON-pointer work in #279, so this is not a regression from
it.

## Summary

Two spec-level checks iterate the `definitions` **map** and exit early on the first
fault they find. Because Go randomises map iteration order, a document with the
same fault in two definitions reports **a different one on each run** — and in one
of the two cases the findings are mutually exclusive, so a caller is told about one
definition and never hears about the other.

Validating the same bytes twice can produce different output. That is more
disruptive for a consumer than an imprecise location: a location can be worked
around, a finding that is absent half the time cannot.

## Site 1 — `validateRequiredDefinitions` (`spec.go:578`)

```go
DEFINITIONS:
	for d, schema := range s.spec.Spec().Definitions {   // randomised order
		if schema.Required != nil {
			for i, pn := range schema.Required {
				...
				if !isValid && !s.Options.ContinueOnErrors {
					break DEFINITIONS
				}
			}
		}
	}
```

**Reproducer** — one fault each in two definitions: `Pet` lists a required property
it never declares (an error), `Tag` marks one both `required` and `readOnly` (a
warning):

```json
{
  "swagger": "2.0",
  "info": {"title": "t", "version": "1"},
  "paths": {
    "/pets": {"get": {"operationId": "getPets",
      "responses": {"200": {"description": "ok", "schema": {"$ref": "#/definitions/Pet"}}}}}
  },
  "definitions": {
    "Pet": {"type": "object", "required": ["notDeclared"], "properties": {"name": {"type": "string"}}},
    "Tag": {"type": "object", "required": ["readOnlyToo"], "properties": {"readOnlyToo": {"type": "string", "readOnly": true}}}
  }
}
```

50 runs of `NewSpecValidator(doc.Schema(), strfmt.Default).Validate(doc)`, collecting
`LocatedErrors()` + `LocatedWarnings()`:

```
ContinueOnErrors=false (default)      => 2 distinct outcomes
   42 runs  [ERR /definitions/Pet/required/0]
    8 runs  [ERR /definitions/Pet/required/0  WARN /definitions/Tag/required/0]

ContinueOnErrors=true                 => 1 outcome, stable
   50 runs  [ERR /definitions/Pet/required/0  WARN /definitions/Tag  WARN /definitions/Tag/required/0]
```

Whether `Tag` is heard from depends on whether the map reached it before `Pet`.
Setting `ContinueOnErrors` is a workaround here, at the cost of every other check's
early exit.

## Site 2 — `validateDuplicatePropertyNames` (`spec.go:238`), worse

```go
	for k, sch := range s.spec.Spec().Definitions {   // randomised order
		...
		ancs, rec := s.validateCircularAncestry(k, sch, knownanc)
		...
		if len(ancs) > 0 {
			res.addErrorsAt(newPathSegments(swaggerDefinitions, k), circularAncestryDefinitionMsg(k, ancs))
			return res                                 // NOT gated on ContinueOnErrors
		}
```

The `return` is unconditional, so no option restores the missing finding.

**Reproducer** — two self-referential definitions, each its own ancestor:

```json
{
  "swagger": "2.0",
  "info": {"title": "t", "version": "1"},
  "paths": {
    "/a": {"get": {"operationId": "getA",
      "responses": {"200": {"description": "ok", "schema": {"$ref": "#/definitions/A"}}}}},
    "/b": {"get": {"operationId": "getB",
      "responses": {"200": {"description": "ok", "schema": {"$ref": "#/definitions/B"}}}}}
  },
  "definitions": {
    "A": {"type": "object", "allOf": [{"$ref": "#/definitions/A"}]},
    "B": {"type": "object", "allOf": [{"$ref": "#/definitions/B"}]}
  }
}
```

50 runs:

```
ContinueOnErrors=false   => 2 distinct outcomes:  45 runs [ERR /definitions/A]   5 runs [ERR /definitions/B]
ContinueOnErrors=true    => 2 distinct outcomes:  48 runs [ERR /definitions/A]   2 runs [ERR /definitions/B]
```

The two are mutually exclusive: you are told about **one** circular definition,
picked at random, and fixing it reveals the other on the next run.

The split is nowhere near even (roughly 1 in 6, and 1 in 10) because Go's iteration
start is biased for a small map — which makes this the kind of flake that hides for
years and then fires in CI.

## Suggested fix, lowest risk first

`validate` is old and widely depended on, so the ranking matters more than the
ideal.

1. **Iterate in sorted key order.** Two lines per site, and it changes *nothing*
   about how many findings are reported or when iteration stops — only *which*
   fault is "first", which becomes stable and reproducible:

   ```go
   for _, d := range slices.Sorted(maps.Keys(s.spec.Spec().Definitions)) {
       schema := s.spec.Spec().Definitions[d]
   ```

   `spec.go` already imports `slices`, and `maps` is used elsewhere in the package,
   so no new dependency. This alone converts a flake into a defined, documentable
   behaviour ("the first fault in definition-name order").

2. **Gate site 2's `return` on `ContinueOnErrors`**, so a caller that asks for
   everything gets everything. Small semantic fix, consistent with site 1.

3. **Not suggested now:** removing the early exits. It would change the number of
   findings a document reports, which is likely to churn goldens in dependents.

## Worth checking while there

If any existing fixture has two same-fault definitions, its expectations may be
passing by luck today. Sorted iteration would make such a test either reliably pass
or reliably fail — the latter being useful information rather than a new problem.

## How this was measured

Each reproducer was run 50 times in one process, the reported set sorted and
counted by distinct outcome, against v0.26.2 from the module proxy. Found while
consuming `Located.Pointer` in a TUI that navigates to each finding: a test written
over a two-fault document passed, then failed, then passed.
