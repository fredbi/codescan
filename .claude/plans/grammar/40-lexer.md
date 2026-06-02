# Lexer specification — quirks absorption + token production

**Status:** scaffold for iteration

The lexer sits between **Preprocess** (comment-marker stripping;
produces `[]Line`) and the **grammar parser** (consumes `[]Token`).
The grammar treats the lexer's output as terminals; the EBNF then
describes the legal token sequences without dragging in
context-sensitive line-shape rules.

This file is the contract for the lexer. It specifies what the
lexer absorbs (quirks, disambiguation, multi-line bodies) and what
shape its output takes. The current implementation lives in
`internal/parsers/grammar/lexer.go` — most of section 5 is already
implemented; sections 4, 6, 7 are the deltas.

## 1. Scope & boundary

**Input scope: one Go `*ast.CommentGroup` per lexer invocation.**
The scanner attaches comment groups to AST nodes (`Decl.Doc`,
`Field.Doc`, etc.) and feeds each group to the lexer separately.
Real Go-source blank lines — the kind that *split* one comment
group from another — never appear in the lexer's input; they are
group boundaries handled by the caller. If a swagger annotation is
split across two source-level comment groups, the lexer treats them
as two independent parses (v1 behaviour: the second group is
silently dropped because no annotation header opens it).

Within a single comment group, lines that strip to empty (i.e. `//`
on a line by itself) surface as `TokenBlank`. These are **not**
group boundaries — they are decorative spacers, common in v1
fixtures (e.g. `Consumes:` and `Produces:` siblings separated by a
blank `//` line in `todo_operation_body.go`). They do **not**
terminate any body — see §7.

The lexer owns:

- **Quirks normalisation** — line endings, trailing dots, case
  variations, decorative fences, godoc prefixes (see §5).
- **Line classification** — annotation, keyword, prose, blank,
  fence (see §3).
- **Multi-line body accumulation** — `OpaqueYamlBody`, `RawBlockBody`,
  `RawValueBody` emitted as single tokens with body content captured
  as a string field (see §4, §7).
- **Disambiguation** — value-shape dispatch rules (`JsonValue` vs
  `RawValue`, `EnumWithName` vs `EnumValuesOnly`, etc.) applied
  before token emission so the grammar productions stay
  context-free (see §6).
- **Position tracking** — every token carries the source position of
  its first meaningful payload character.

The lexer does **not** own:

- Semantics — type compatibility, namespace resolution, etc. live in
  the analyzer.
- Annotation routing — the dispatcher in the grammar layer reads
  `TokenAnnotation.Text` and decides which family sub-grammar parses
  the rest.
- AST / Go-type inspection — the scanner provides those; the lexer
  is text-only.

## 2. Pipeline

```
   *ast.CommentGroup
        ↓
    Preprocess          ← strip "// " / "/*…*/" comment markers; split into lines
        ↓
      []Line            ← (Text, Raw, Pos) per line
        ↓
      Lex               ← state-machine line classifier + body accumulator
        ↓
     []Token            ← grammar terminals
        ↓
     Parse              ← grammar productions, dispatch by annotation family
        ↓
      Block             ← AST consumed by the builders
```

Two stages, **separate functions**:

1. **Line classifier** (`lexLine`) — pure function on a single line
   plus minimal carry state (in-fence flag). Emits per-line classifier
   tokens. Today's `lexer.go:lexLine` covers this.
2. **Body accumulator** (new) — state machine over the line-classifier
   token stream, folding multi-line bodies into single
   `Opaque*Body` / `Raw*Body` tokens.

Two stages keep each piece testable in isolation. Future work may
fuse them; the spec doesn't preclude that.

## 3. Token inventory

### Single-line tokens (line-classifier output)

| Kind                  | Carries                                                                  | Example source line                          |
|-----------------------|--------------------------------------------------------------------------|----------------------------------------------|
| `EOF`                 | (sentinel)                                                               | —                                            |
| `Blank`               | nothing                                                                  | (empty after trim)                           |
| `Text`                | `Text` (cleaned), `Raw` (original)                                       | `// Lists the pets known to the store.`      |
| `Annotation`          | `Text` (annotation name), `Args` ([]string)                              | `// swagger:route GET /pets pets listPets`   |
| `Keyword`             | `Text` (canonical kw), `Value` (rest after `:`), `Keyword` (table entry), `ItemsDepth`, `SourceName` | `// maximum: 10` / `// items.maxLength: 5`   |
| `KeywordHead`         | same as `Keyword`, `Value == ""`                                         | `// consumes:`                               |
| `YAMLFence`           | nothing                                                                  | `// ---`                                     |

`Keyword` and `KeywordHead` differ only in whether `Value` is empty.
Open question A5: collapse the two into a single `Keyword` with
empty `Value` indicating "block follows". Today they are distinct.

### Multi-line tokens (body-accumulator output)

| Kind                  | Carries                                                                  | Notes                                          |
|-----------------------|--------------------------------------------------------------------------|------------------------------------------------|
| `OpaqueYamlBody`      | `Body` (string with `\n`), `Raw` (verbatim source)                       | The bytes between matching `---` fences        |
| `RawBlockBody`        | `Body`, `Raw`, `Keyword` (parent kw)                                     | Body of a structural keyword (consumes/produces/security/responses/parameters/extensions/externalDocs/tos) |
| `RawValueBody`        | `Body`, `Raw`, `Keyword` (parent kw), `Typed` (disambiguated value)      | Body of a value-bearing keyword (default/example/enum); single-line scalar OR multi-line |

The body-accumulator emits these in place of the line stream that
opened them — the `KeywordHead` token and any subsequent
`Text` / `Blank` / `RawLine` tokens that belong to the body are
**replaced** by a single body token in the output stream.

A2 (open): body content shape. Today, `grammar.Property.Body` is
`[]string` (one element per source line). The proposal here is to
expose a single string with embedded `\n` (more useful for YAML
unmarshal and string scanning) plus `Raw` (verbatim including
indentation) for downstream parsers that need indentation fidelity.

## 4. State machine modes

```
                    ┌───────────────┐
                    │   Default     │
                    └───────┬───────┘
                            │
       ┌────────────────────┼─────────────────────┐
       │ open YAML fence    │ keyword-head opens  │ inline annotation
       │  (---)             │ a multi-line body   │
       ▼                    ▼                     ▼
┌───────────────┐    ┌───────────────┐     (annotation token,
│ InYAMLFence   │    │ InRawBlock<k> │      stay in Default)
│               │    │ InRawValue<k> │
└───────┬───────┘    └───────┬───────┘
        │ close fence (---)   │ next sibling structural item /
        │ or EOF              │ EOF
        ▼                    ▼
   emit OpaqueYamlBody    emit Raw{Block,Value}Body
        │                    │
        └─────────┬──────────┘
                  ▼
              Default
```

Mode transitions are deterministic and driven by the line-classifier
token kind plus the open keyword's body shape.

**Default mode** emits the line-classifier tokens verbatim, except:
- `YAMLFence` opens `InYAMLFence` and is **consumed** (the fence
  token is not emitted as part of the output stream — see §5).
- `KeywordHead` for a body-bearing keyword opens `InRawBlock<k>`
  or `InRawValue<k>` per the keyword's table entry, and is itself
  the trigger (the head **is** emitted as the first part of the
  body token's metadata).
- `Keyword` (with non-empty `Value`) for a body-bearing keyword
  emits a `RawValueBody` immediately, with the inline value as the
  whole body — single-line case is the trivial path of the
  multi-line accumulator.

**InYAMLFence** mode collects every line until the next `YAMLFence`
(or EOF). Lines are joined with `\n` and emitted as a single
`OpaqueYamlBody`. Decorative `---` fences inside the body are
dropped (see §5).

**InRawBlock\<k\>** mode collects body lines until the next sibling
structural item — see §7 for the per-keyword termination rule.
Emits `RawBlockBody` with `Keyword = k`.

**InRawValue\<k\>** mode collects body lines until the next sibling
structural item. Emits `RawValueBody` with `Keyword = k` and a
`Typed` field carrying the disambiguated value (§6).

## 5. Quirks absorbed by the lexer

| Quirk                                       | Rule                                                                                                                                                                              |
|---------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **CR vs CRLF**                              | Preprocess normalises `\r\n` → `\n` and lone `\r` → `\n` before line splitting. Lexer never sees `\r`.                                                                            |
| **Trailing-dot elision on keyword/annotation lines** | `EOL = [ "." ] , LF`. After extracting `Value` (Keyword) or `Args` (Annotation), strip a single trailing `.` if present. Source preservation lives in `Raw`.                  |
| **Multi-line block bodies**                 | Body-accumulator (§4) folds the head + body lines into a single `OpaqueYamlBody` / `RawBlockBody` / `RawValueBody`.                                                               |
| **Decorative `---` after an inlined block** | Inside `InYAMLFence` *and* inside an `InRawBlock<extensions>`, a `---` line is **dropped silently**. Empirical confirmation: baseline-worktree `fenced_ext_probe_test.go` produced byte-identical output with and without the fence. |
| **`swagger:route` godoc prefix**            | A line of the form `<GoIdent><WS>swagger:route <args>` has the leading `<GoIdent><WS>` stripped before annotation lexing. Only `swagger:route` is allowed this prefix; other annotation names disable the heuristic. Implementation: `matchGodocRoutePrefix` (already present). |
| **Case-insensitive first character on keywords** | `Consumes:` and `consumes:` both lex as the `consumes` keyword. The first-character lowercase is applied before `Lookup(name)`. The rest of the keyword name is matched verbatim (so `Consumes` → `consumes`, but `CONSUMES` does **not** match — only the first character has the case insensitivity, matching v1's `[Cc]onsumes` regex idiom). |
| **`items.` prefix runs**                    | Repeated `items.` segments before a keyword (e.g. `items.items.maxLength:`) are stripped and counted; `ItemsDepth` records the depth. Bare `items:` (no separator) is not stripped. Implementation: `stripItemsPrefix` (already present). |
| **Indentation normalisation**               | The line classifier left-trims for keyword detection; `Raw` keeps the original line including indentation. Downstream parsers that need indentation fidelity (e.g. `internal/parsers/routebody/route_params.go`) read `Raw`. The body accumulator never modifies indentation — it concatenates `Raw` lines verbatim. |

The list is closed for round 1 unless a fixture surfaces a quirk
that doesn't fit. Each rule is grammar-invisible: the EBNF references
the lexer's output token, not the source line.

## 6. Disambiguation moved into the lexer

These dispatch rules belong to the lexer so the grammar productions
stay context-free. The lexer emits **already-disambiguated** typed
tokens; the grammar references them as terminals.

| Site                                                | Rule (full text in source file)                                            | Lexer emits                                          |
|-----------------------------------------------------|-----------------------------------------------------------------------------|------------------------------------------------------|
| `swagger:default <arg>` value                       | `JsonValue` first, fall back to `RawValue` (see `23-classifier-grammar.md` §"Value argument") | `Annotation` with `Args[0]` typed as JSON or raw     |
| `default:` / `example:` / `enum:` body              | `RawValueBody`'s `Typed` field per `helpers.ParseValueFromSchema` against the inferred Go target type (or `ParseEnum` for `enum:`). Where typing requires Go-type context unavailable to the lexer, the typing is deferred to the analyzer; the lexer emits `Typed.Type = ValueDeferred`. | `RawValueBody` with `Typed.Type` set                 |
| `swagger:enum` argument                             | Four-step dispatch (see `23-classifier-grammar.md` §"Disambiguation rule for `EnumArgs`"): `[`-led → bracketed; ident + WS + rest → name + values; lone ident → name only; otherwise → plain list. | `Annotation` with `Args` already split into `(Name?, ValueList)` |
| `EnumValueList` form (plain vs bracketed)           | `[`-led → `EnumBracketedList`; otherwise → `EnumPlainList`. Inside bracketed list: per-item, `JsonValue` first, `EnumPlainItem` fallback. | Pre-tokenised list items                             |
| `enum:` body within `RawValueBody`                  | Same dispatch — re-applied at body re-parse time.                           | Reuses the same logic                                 |

A central `disambiguate.go` module is the natural home for these
rules so the lexer and analyzer share one implementation. (Today,
the rules are split across `internal/parsers/helpers/` and grammar
keyword tables.)

## 7. Termination rules per body kind — **OPEN**

The body-accumulator must decide when each multi-line body ends. The
rule differs per body kind. **This section is the largest open
design call** — the rules below are *proposed* round-1 defaults;
each needs validation against fixtures before adoption.

| Body kind             | Proposed termination rule                                                                                                  | Risk                                                                                          |
|-----------------------|-----------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------|
| `OpaqueYamlBody`      | Closing `---` fence on its own line (or EOF).                                                                              | Mid-body decorative fence is the main trap — handled by §5 (drop inside the active body).      |
| `RawBlockBody` for `consumes:` / `produces:` / `security:` / `tos:` | Next non-blank line that is an `Annotation`, a `Keyword` / `KeywordHead` *not* matching the current body's allowed sub-keys, or EOF. Blank lines do **not** terminate. | "Allowed sub-keys" is itself per-keyword — security has nested name/scope syntax.              |
| `RawBlockBody` for `responses:` (route)   | Next sibling top-level keyword or annotation, or EOF. Inside-body lines are accumulated regardless of indentation (per `routebody/responses.go`'s own scanner). | Similar.                                                                                     |
| `RawBlockBody` for `parameters:` (route)  | Next sibling top-level keyword or annotation, or EOF. The `+ name:` continuation list is recognised by `routebody/route_params.go` after the lexer hands over the body string. | Similar.                                                                                     |
| `RawBlockBody` for `extensions:`          | Next sibling top-level keyword or annotation, or EOF. Decorative `---` fences inside the body are dropped (§5).            | Same.                                                                                          |
| `RawBlockBody` for `externalDocs:`        | Next sibling top-level keyword or annotation, or EOF.                                                                      | —                                                                                              |
| `RawValueBody` for `default:` / `example:` | Single-line case: the inline value after `:`, no body accumulation. Multi-line case (head with empty value): next sibling structural item or EOF. | Block-body absorption for object/array values is a known v1 quirk — Round-1 known omissions.   |
| `RawValueBody` for `enum:`                | Same as default/example. Inside the body, `EnumValueList` dispatch applies (§6).                                            | Multi-line bracketed list spanning `\n` is a fixture-worthy edge case.                         |

"Sibling structural item" is concretised as: a `KeywordHead` /
`Keyword` whose **canonical name** is in the same parent block's
keyword set, or any `Annotation`. Indented lines that look like
keywords but whose canonical name is **not** in the parent's set
remain part of the body (this matches v1's permissive behaviour for
YAML-shaped bodies under e.g. `security:`).

**Blank lines within the input do not terminate a body.** A `//`
line that strips to empty surfaces as `TokenBlank` and is absorbed
into the active body's content (preserved verbatim in `Raw`). v1
fixtures use blank lines liberally as visual separators inside list
bodies, security maps, and parameter continuation lists; the body
extends across them until a real sibling structural item appears.

Real Go-source blank lines (which split comment groups) never
appear in the lexer's input — see §1.

The rule needs fixture validation. Workshop suggestion: collect ten
"most representative" bodies per kind from `fixtures/goparsing/...`
and assert the proposed rule yields the same line groupings as v1
parses.

## 8. Title / Description prose classification — **LEXER**

Prose classification belongs in the lexer. The split rules currently
implemented in `helpers.CollectScannerTitleDescription` are **pure
line-shape heuristics** with no annotation-semantic input:

1. If a `TokenBlank` appears inside the prose run, split there:
   lines before → title, lines after → description.
2. Otherwise, if the first prose line ends with Unicode punctuation
   (`\p{Po}$`) → first line is title, rest is description.
3. Otherwise, if the first prose line matches a markdown-heading
   prefix (`^#+\s+`) → strip the heading marker, first line is
   title, rest is description.
4. Otherwise → entire prose is description, title is empty.

The only context-dependent decision is **whether the split is
applied at all**. The lexer already has it: when an annotation
header opens the input the split applies; for an `UnboundBlock`
(member-level docstring, no annotation header) the entire prose is
description and no title is emitted. The body-accumulator state
machine carries this binary flag — no per-family modes are needed.

### Token shape

Replace generic `Text` tokens with three typed kinds:

| Kind          | Semantics                                                   |
|---------------|-------------------------------------------------------------|
| `TitleLine`   | A prose line classified as part of the title block. Multiple lines are emitted as a contiguous run. |
| `DescLine`    | A prose line classified as part of the description block. Same — contiguous run.                    |
| `Text`        | Reserved for prose lines that are neither title nor description — currently unused; kept for future raw-text needs. |

The `Pre-cleanup` step (`rxUncommentHeaders` — strip leading
whitespace, slashes, asterisks, dashes, optional `|`) is folded
into the line classifier; downstream consumers receive cleaned
text.

### Grammar consequences

Family productions gain explicit `Title` / `Description` slots
where v1 distinguishes them:

```ebnf
SchemaBlock      = Prelude , SchemaAnnotation , [ Title ] , [ Description ] , SchemaAnnotationBody ;
RouteBlock       = Prelude , RouteAnnotation  , [ Title ] , [ Description ] , RouteBody ;
InlineOpBlock    = Prelude , InlineOpAnnotation , [ Title ] , [ Description ] , InlineOperationBody ;
MetaBlock        = Prelude , MetaAnnotation  , [ Title ] , [ Description ] , MetaBody ;
UnboundBlock     = Prelude , [ Description ] , UnboundBlockBody ;          (* no title under UnboundBlock *)

Title            = TitleLine , { TitleLine } ;
Description      = DescLine  , { DescLine  } ;
```

This makes the grammar honest about a feature the helper layer was
silently providing. `helpers.CollectScannerTitleDescription` and
`helpers.CleanupScannerLines` (and `JoinDropLast` once its only
remaining call site is dead) become obsolete once the lexer
implementation lands.

### Implementation note

The classifier walks the prose-line subset of the token stream once
per parse. It is conceptually a third pipeline stage:

```
Preprocess → Lex (line classifier) → BodyAccumulator → ProseClassifier → Parse
```

Or, since prose classification is purely a re-typing of `Text`
tokens that survived body accumulation, it can be inlined into the
body accumulator's emission path. Either is fine — the test surface
is the same.

The four heuristics above are the round-1 contract. New rules
(e.g. ATX-style markdown subtitle, godoc-style "Deprecated:"
trailers) are v2 enhancements.

## 9. Diagnostics

The lexer produces tokens, never errors. Lines that fail
classification (e.g. `swagger:` with no name) fall back to `Text`.
The grammar parser emits diagnostics during dispatch with positions
the lexer attached. This matches today's behaviour and keeps the
lexer total.

Body accumulator failure modes (e.g. unclosed `---` fence) are
handled by **eager close-on-EOF** with a diagnostic flag on the
emitted body token: `Truncated bool` is set when the fence wasn't
closed. The grammar treats truncated bodies as legal token shapes;
the parser emits the diagnostic.

## 10. Examples

(filled out after §7 settles — each open termination rule produces
a different example trace)

## 11. Open questions

| Tag  | Question                                                                                                                  | Stance / next step                                                                                                  |
|------|----------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------|
| A1   | Two stages (line classifier + body accumulator) vs fused single-pass state machine.                                        | Two stages — recommended. Revisit if the inter-stage boundary becomes a bottleneck.                                  |
| A2   | Body content shape: single string (with `\n`) vs `[]string`.                                                                | Single string in `Body` + verbatim `Raw`. Today's `[]string` survives via `strings.Split(body, "\n")` at consumer.    |
| A3   | Termination rule per body kind.                                                                                             | Per-kind proposals in §7; validate against fixtures.                                                                  |
| A4   | Decorative-fence semantics inside `ExtensionYAMLBody` and `OpaqueYamlBody`.                                                 | Drop silently when inside an already-active body. §5.                                                                 |
| A5   | Collapse `Keyword` + `KeywordHead` into a single token with empty `Value` indicating block-head?                            | Keep distinct — disambiguation is cleaner with separate kinds. Revisit if a fixture forces ambiguity.                 |
| A6   | Synthesise `RawValueBody` for inline-only single-line `default: 3`?                                                         | Yes — the body accumulator emits `RawValueBody` for both, single-line is the trivial path. Uniform downstream.       |
| A7   | Title / description in the lexer or downstream?                                                                             | **Lexer.** §8 settled — typed `TitleLine` / `DescLine` tokens; `helpers.CollectScannerTitleDescription` becomes obsolete. |
| A8   | First-character case insensitivity vs whole-word: `CONSUMES` matches?                                                       | First character only. Matches v1's `[Cc]onsumes` regex.                                                                |
| A9   | Where the disambiguation module lives: `internal/parsers/grammar/disambiguate.go`, or split between lexer + helpers?         | Single module under `grammar/`. Lexer + analyzer both call into it.                                                   |
| A10  | Position tracking on body tokens: start position, end position, both?                                                       | Start position only (matches today's per-token convention). End is implied by the next token's start.                 |

Items A3 and A9 are the only ones that block writing the
implementation. The rest can be deferred.
