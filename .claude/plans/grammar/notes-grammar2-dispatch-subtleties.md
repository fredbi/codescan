# Grammar2 annotation-dispatch subtleties

**Origin:** captured during P7 schema migration; relocated from auto-
memory to `.claude/plans/grammar/` as part of P7.1 plan revision so all
project documentation lives in one place.
**Audience:** future builder migrations (P8a parameters/responses,
P8b operations/routes) will hit these same traps.
**Subject:** three non-obvious traps when reaching from a builder
into grammar2's parsed Block — per-line vs whole-block parsing,
single-word arg filter, and the AnnName UnboundBlock fall-through.

When migrating a builder off the regex-based `parsers.*` utilities
(`parsers.ModelOverride`, `StrfmtName`, `TypeName`, `EnumName`,
`NameOverride`, `DefaultName`, `AllOfName`, `AllOfMember`, `Ignored`,
`AliasParam`) to grammar2-driven helpers, three subtleties surfaced
during P7.

**Why:** the v1 regex utilities scan every comment line
independently with anchored end-of-line matching; grammar2's
per-block `Parse(cg)` returns a single typed Block keyed on the
FIRST annotation in source order. The semantics differ in three
concrete ways.

**How to apply:** when porting any `parsers.*` regex utility, copy
the pattern from `internal/builders/schema/comment_lookups.go`
rather than calling `Parse(cg)` directly. Specifically:

1. **Use lexer-tokens, not whole-block Parse.** Some prose lines
   accidentally start with `swagger:<kind>` (e.g.
   `// swagger:type so the scanner emits...` in
   `fixtures/enhancements/named-basic`). Whole-block Parse picks
   that as THE annotation; the real `swagger:type string` later in
   the block is treated as body and lost. Iterate
   `Lex(Preprocess(cg, fset))` and find `TokenAnnotation` entries
   matching the kind.

2. **Filter args by "single-word."** v1's `commentSubMatcher`
   regex captured `\S+` and anchored to `$`, so multi-word arg
   strings ("so the scanner emits...") never matched. grammar2's
   lexer happily captures the rest of the line. Reject args
   containing whitespace via `strings.ContainsAny(arg, " \t")` —
   matches v1's de-facto filter.

3. **Do NOT enforce closed-vocabulary args.** grammar2 emits
   `CodeInvalidTypeRef` when `swagger:type X` has X outside the
   closed vocabulary. v1's regex didn't validate the vocabulary,
   and `fixtures/enhancements/swagger-type-array` deliberately
   uses `swagger:type badvalue` so `SwaggerSchemaForType` fails
   and the schema builder falls through to its $ref-emitting
   branch. Filtering on diagnostics breaks this fixture. The
   single-word filter alone is sufficient — it rejects prose
   lines without rejecting unrecognised-but-well-formed args.

**One more trap, narrower scope:** `swagger:name` (AnnName) was
promoted from `familyClassifier` to `familySchema` in P7/S3 so its
schema-shaped body works. But `parseSchemaBlock`'s `default`
branch returns `*UnboundBlock` for AnnName — losing the IDENT_NAME
arg at the Block level. Using `Parse(cg).(*X)` doesn't yield the
arg. Reading from `TokenAnnotation.Args` directly (the
lexer-tokens approach above) sidesteps this. Logged as a P7.1/A1
cleanup target — a typed `*NameBlock` with an `IDENTName` field
will close the gap and make this trap obsolete.

**Two opposite arg-presence semantics.** v1 had two regex matchers
for "annotation with optional arg":

- `commentSubMatcher` — strict; bare annotation returns
  `("", false)`. Used by `AllOfName`, `StrfmtName`, `TypeName`,
  `EnumName`, `NameOverride`, `DefaultName`.
- `commentBlankSubMatcher` — bare annotation returns
  `("", true)`. Used by **only** `ModelOverride` (and
  `ResponseOverride` / `ParametersOverride` at the scanner layer).

Easy to conflate when porting — initial mapping of `AllOfName` to
the optional variant produced an `x-class: ""` extension on every
bare `swagger:allOf` block (`TestEmbeddedAllOf` caught it). The
schema-package helpers split this:

- `annotationArgRequired(cg, kind)` — strict, single-word.
- `annotationArgOptional(cg, kind)` — bare returns `("", true)`.
  Use for `ModelOverride` only.
- `hasAnnotation(cg, kind)` — bare-presence boolean. Use for
  `Ignored` / `AliasParam` / `AllOfMember`.

After P7.1's Phase A4 (`grammar2.ParseAll`) and the typed
`*NameBlock`, traps (1) and the AnnName fall-through both go away
— the Walker reads typed Block fields directly, no re-lexing
needed. Trap (2) and trap (3) remain relevant for any layer that
still touches lexer tokens directly.
