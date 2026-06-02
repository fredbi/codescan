# Grammar-Based Parsing vs Regexps for a v2 Annotation Parser

Date: 2026-03-25

## Context

The codescan tool currently uses **62 compiled regexps** across `regexprs.go`, `validations.go`,
`parsed_path_content.go`, and `meta.go` to parse 14 swagger annotation types and ~20 validation
properties. Each annotation and each validation property has its own regexp. Some have parameterized
variants (the `rxItemsPrefixFmt` family generates regexps dynamically per nesting level).

The SectionedParser runs each line against every registered tagger's regexp until one matches. For a
typical schema field comment with 5 lines and 15 taggers, that's up to 75 regexp match attempts per
field — and a model with 30 fields runs ~2,250 matches just for validation tags.

The question: for a future v2, should we replace these regexps with a grammar-based parser?

## The Case Against Regexps (Status Quo Pain)

### 1. The regexps are unreadable

Consider `rxMaximumFmt`:
```
%s[Mm]ax(?:imum)?\p{Zs}*:\p{Zs}*([\<=])?\p{Zs}*([+-]?(?:\p{N}+\.)?\p{N}+)(?:\.)?$
```

This encodes: an optional items prefix, case-insensitive "max" or "maximum", optional whitespace,
a colon, optional whitespace, an optional exclusivity operator (< or =), optional whitespace, a
signed decimal number, and an optional trailing period. **No one can read this without the test
suite as a Rosetta stone.** And there are 13 variants of this pattern for different validation
properties.

### 2. Case-insensitivity is handled per-regexp, not per-parser

Every regexp independently handles `[Mm]`, `[Cc]ollection`, `[Ff]ormat` etc. A grammar would
declare case-insensitivity once at the keyword level.

### 3. The `items.` prefix system is a hack

The `rxItemsPrefixFmt` generates regexps like `(?:[Ii]tems[.\p{Zs}]*){1}` to match nested
`items.items.minimum: 5`. This is string-formatted regexp composition — fragile, hard to extend
to deeper nesting, and the resulting regexps are compiled at runtime (no `MustCompile` safety net
at init time).

### 4. Multiple regexps per line, no shared parse tree

When a line says `maximum: < 100`, three things happen:
- The tagger loop tries every regexp until `rxMaximum` matches.
- `SetMaximum.Parse` runs the *same* regexp again to extract groups.
- The groups are then parsed by `strconv.ParseFloat`.

A grammar would match, extract, and type-convert in one pass.

### 5. Error messages are poor

When a regexp doesn't match, you get nothing — no position, no "expected ':' after keyword", no
"invalid number format". The parser silently ignores malformed annotations. A grammar gives you
parse errors with positions.

### 6. Documentation gap

There is no formal specification of what annotations look like. The regexps *are* the spec, but
they're write-only. A grammar IS a readable specification that could be included in user docs.

## The Case For Keeping Regexps

### 1. They work and they're proven

14 years of production use. The regexp suite, despite being hard to read, is exhaustively tested
(the regexprs_test.go cartesian-product approach tests thousands of combinations). Rewriting
introduces risk.

### 2. Go's regexp engine is safe (no ReDoS)

Go uses RE2 semantics — guaranteed linear time, no catastrophic backtracking. A PEG parser or
hand-rolled recursive descent doesn't have this guarantee automatically (PEG's ordered choice
can backtrack, though in practice the annotation grammar is simple enough that it wouldn't).

### 3. The annotation syntax is genuinely simple

The annotations are all `keyword: value` or `swagger:annotation args`. There are no nested
expressions, no operator precedence, no ambiguity. This is below the complexity threshold where
a grammar pays for itself.

### 4. Performance may not improve

The current regexps are compiled once at init. Go's `regexp` package is not the fastest, but for
the volumes involved (thousands of match attempts, not millions), it's fine. A PEG parser has
overhead too — packrat memoization tables, parser combinator dispatch. For 5-line comment blocks,
the constant factor matters more than algorithmic complexity.

### 5. No new dependencies

Regexps are stdlib. A PEG parser adds a dependency (pigeon, participle, etc.) or requires a code
generator step in the build.

## Grammar Options for Go

### PEG (pigeon)

**`github.com/mna/pigeon`** — generates Go code from a PEG grammar file.

Pros:
- Grammar file is self-documenting.
- Generated code, so no runtime dependency.
- Supports semantic actions (Go code embedded in the grammar).

Cons:
- Requires a `go generate` step.
- Generated code is verbose and hard to debug.
- PEG's ordered choice (`/`) can cause subtle priority bugs.
- Fred has practical experience with it — so the learning curve is known.

### Parser combinators (participle)

**`github.com/alecthomas/participle/v2`** — defines grammar via Go struct tags.

Pros:
- No code generation. Grammar is in Go code.
- Struct tags are readable: `@Ident ":" @Number`.
- Produces typed AST nodes directly — no post-parse extraction.
- Active maintenance, good documentation.

Cons:
- Runtime dependency.
- Struct-tag grammar has limits (no left recursion, limited lookahead).
- For very simple grammars, it might be over-engineered.

### Hand-rolled recursive descent

No library. A `scanner` type that walks the line character by character with methods like
`expectKeyword()`, `expectColon()`, `expectNumber()`.

Pros:
- Zero dependencies.
- Full control over error messages.
- Easiest to optimize (no abstraction overhead).
- Can be written to match the exact current behavior, including all the case-insensitivity
  quirks and the trailing-period tolerance.

Cons:
- More code to maintain.
- No formal grammar artifact (the grammar is implicit in the code).
- Easy to introduce bugs in character-level parsing.

### Hybrid: structured tokenizer + regexp for values

Keep regexps for value extraction (numbers, booleans, patterns) but replace the tagger
dispatch loop with a tokenizer that classifies lines by keyword first:

```go
type Token struct {
    Keyword string   // "maximum", "minimum", "swagger:route", etc.
    Op      string   // "<", ">", "=", "" (for min/max exclusivity)
    Value   string   // the raw value string
    Prefix  []string // ["items", "items"] for nested items
}
```

A single-pass tokenizer strips comment prefixes, identifies the keyword (case-insensitive),
splits on `:`, and returns a Token. The SectionedParser then dispatches on `token.Keyword`
via a map lookup instead of running N regexps.

Pros:
- Minimal change to the architecture (SectionedParser stays, taggers stay).
- Eliminates the O(N) regexp scan per line.
- Keywords become a documented enum, not hidden in regexps.
- Value parsing can still use regexps where they're genuinely useful (number formats).

Cons:
- Still not a "real" grammar — no formal specification artifact.
- The keyword tokenizer itself needs careful handling of the `items.` prefix nesting.

## The v2 Context Changes Everything

The initial analysis above assumed Swagger 2.0 — a frozen spec with a bounded annotation surface.
But v2 targets **OpenAPI 3.x**, which changes the calculus:

### OpenAPI 3.x is richer

OAI 3.x introduces `oneOf`/`anyOf`/`allOf` composition, `discriminator` with mapping, `callbacks`,
`links`, `requestBody` (separate from parameters), `content` with media-type keys, `components`
with `$ref` scoping, `servers` (replacing `host`+`basePath`+`schemes`), and `security` with
OpenID Connect. The annotation surface roughly **doubles**.

With 62 regexps for Swagger 2.0, an OAI 3.x regexp approach would need 100+. That's not
maintainable. The "annotations are simple" argument no longer holds when you need to express:

```
// openapi:requestBody
// content: application/json
//   schema: $ref:#/components/schemas/Pet
// content: multipart/form-data
//   schema:
//     properties:
//       file:
//         type: string
//         format: binary
```

This has nested structure (content → media-type → schema → properties). Regexps don't nest.

### More complex annotations for better specs

Beyond the OAI 3.x baseline, the goal is to support richer annotations that produce more
complete specs — e.g. examples, links between operations, callback definitions, server
variables. The annotation language needs to grow, and every growth step with regexps means
more regexps, more tagger registrations, more dispatch overhead.

### Smart pattern detection (future)

A separate but related v2 goal: the scanner should be smarter at detecting API patterns from
code structure, not just annotations. For example, recognizing that a `func(w http.ResponseWriter,
r *http.Request)` with `r.URL.Query().Get("limit")` implies a query parameter, without requiring
an explicit annotation.

This doesn't directly affect the annotation parser, but it means the annotation parser's job
shifts: instead of being the sole source of spec information, annotations become **overrides
and supplements** to what the scanner infers from code. The grammar needs to be expressive
enough for the override use case ("this parameter is actually a header, not a query param")
without requiring users to annotate everything.

### LSP integration

Editor tooling (VS Code, Neovim) for inserting, completing, and validating annotations is a
v2 goal. This requires:

- A parser that can operate on **incomplete** input (cursor is mid-annotation).
- **Position tracking** — the parser must know where each token starts/ends in the source.
- **Error recovery** — don't stop at the first error, collect all diagnostics.
- A **grammar specification** that the LSP can use to offer completions (available keywords,
  expected value types after each keyword).

Regexps cannot do any of this. A regexp either matches or it doesn't — there's no partial
match, no cursor position, no "what could come next?" query.

This is the strongest argument for a real grammar. Not just for cleanliness, but because
**the grammar becomes a product artifact** shared between the scanner and the editor tooling.

## Revised Assessment

Given that v2 targets OAI 3.x with richer annotations and LSP integration:

| Approach | OAI 3.x scaling | LSP support | Error messages | Effort |
|----------|-----------------|-------------|----------------|--------|
| More regexps | Poor (100+ regexps) | None | None | Low per-feature, high total |
| Hybrid tokenizer | Moderate | Minimal | Basic | Medium |
| PEG (pigeon) | Good | Partial (no error recovery) | Good | Medium-high |
| Participle | Good | Partial | Good | Medium |
| Hand-rolled recursive descent | Excellent | Full control | Excellent | High initial, low ongoing |

The hybrid tokenizer (which was the recommendation for a frozen Swagger 2.0 spec) is now
**insufficient**. The nesting required by OAI 3.x content types and the LSP requirement
both push toward a real parser.

## Revised Recommendation

### The grammar approach: hand-rolled recursive descent

For this specific project, hand-rolled recursive descent wins over PEG/participle because:

1. **No dependencies** — critical for a library in the go-openapi ecosystem.

2. **Full control over error recovery** — PEG parsers abort on first mismatch. Participle has
   limited error recovery. A hand-rolled parser can skip to the next line and continue, which
   is essential for LSP diagnostics ("line 3: expected value after 'maximum:', got end of line;
   line 5: unknown keyword 'mxaimum', did you mean 'maximum'?").

3. **Position tracking is trivial** — the scanner always knows its byte offset. PEG/participle
   parsers provide positions too, but mapping them back to Go AST comment positions requires
   an adapter layer.

4. **Incremental parsing** — for LSP, we need to re-parse a single comment block when it
   changes, not the whole file. A hand-rolled parser is easy to invoke on a substring.

5. **The grammar is simple enough** — even with OAI 3.x, the annotation language is
   `keyword: value` with optional nesting (indentation-based, like YAML-lite). This is well
   within what recursive descent handles naturally. We don't need parser combinators or
   packrat memoization.

### Architecture sketch

```
                  ┌──────────────┐
  comment lines → │   Lexer      │ → token stream
                  │  (per-line)  │    (keyword, colon, value, indent, annotation, eof)
                  └──────┬───────┘
                         │
                  ┌──────▼───────┐
                  │   Parser     │ → typed AST
                  │  (recursive  │    (AnnotationNode, PropertyNode, BlockNode)
                  │   descent)   │
                  └──────┬───────┘
                         │
                  ┌──────▼───────┐
                  │   Emitter    │ → spec mutations
                  │  (visitor)   │    (writes to *spec.Operation, *spec.Schema, etc.)
                  └──────────────┘
```

The **lexer** replaces all 62 regexps. It strips comment prefixes, classifies keywords
(case-insensitive lookup in a map), and tokenizes values. One pass per line.

The **parser** replaces the SectionedParser + TagParser machinery. It understands the
annotation structure: header section (title + description), then keyword:value pairs
with optional indentation-based nesting.

The **emitter** replaces the individual `Set*` types. It walks the parsed AST and writes
values into the target spec objects. This is where type conversion happens (strings to
numbers, booleans, refs).

The LSP integration hooks into the **lexer** (for completion — "what keywords are valid
here?") and the **parser** (for diagnostics — "this block is malformed because...").

### What the grammar might look like (EBNF sketch)

```ebnf
comment_block  = [ annotation ] header { property | block } .
annotation     = "swagger:" annotation_name { arg } .
annotation_name = "route" | "operation" | "model" | "response" | "parameters" | ... .

header         = { text_line } .
text_line      = (* any line not matching keyword ":" or annotation *) .

property       = [ items_prefix ] keyword ":" value .
items_prefix   = { "items" "." } .
keyword        = "maximum" | "minimum" | "pattern" | "enum" | "required" | ... .
value          = number | boolean | string | json_array .

block          = block_keyword ":" newline indent { property | text_line } dedent .
block_keyword  = "consumes" | "produces" | "security" | "extensions" | "content" | ... .
```

This is maybe 20 productions. A hand-rolled recursive descent parser for this is ~300 lines
of Go — comparable to the current `sectioned_parser.go` + `tag_parsers.go` + `validations.go`
combined, but **readable as a specification**.

## Migration Path

1. **v1.x (now)**: Keep regexps, keep SectionedParser. The refactoring into `internal/parsers`
   already isolates them — the rest of the codebase won't need to change when the parser is
   replaced.

2. **v2 alpha**: Implement the lexer + parser alongside the existing regexps. Run both in
   parallel, assert identical output (golden-file testing against the existing fixtures).

3. **v2 beta**: Remove the regexps. The grammar is now the specification.

4. **v2 LSP**: Build the editor integration on top of the parser's token stream and diagnostics.

The `internal/parsers` package boundary created by the current refactoring is exactly the
seam where v2 plugs in. The builders don't care whether the parser uses regexps or recursive
descent — they receive the same typed values.

## Design Decisions (from discussion)

### 1. Comments strategy: godoc + private comments, not embedded YAML

The current `swagger:operation` embeds a full YAML block inside a godoc comment. This was
a pragmatic choice but causes friction:

- YAML indentation is fragile inside Go comments (tabs vs spaces, `//` prefix alignment).
- godoc renders the YAML literally — it's noise in generated documentation.
- Tooling (editors, `go doc`, pkgsite) doesn't understand it.

**v2 approach**: Use two kinds of comments:

- **godoc comments** (`// FunctionName ...` directly above a declaration) carry the title,
  description, and simple `keyword: value` annotations. These render cleanly in godoc.

- **Private comments** (`//` blocks NOT attached to a declaration, or `/* */` blocks) carry
  richer spec content that doesn't belong in godoc — full request/response schemas, callback
  definitions, complex security requirements. The scanner reads these; godoc ignores them.

This means the grammar has two entry points: a "godoc-compatible" subset (flat key:value,
no nesting deeper than `items.`) and a "private spec" mode (full nesting, possibly still
using a YAML-like indentation syntax or a custom structured format). The private comments
can afford to be more verbose and structured since they don't pollute API documentation.

Limited backward compatibility for embedded YAML in godoc comments can be maintained via
a legacy parsing mode, but new features wouldn't use that pattern.

### 2. Struct tags: read from validation frameworks, don't impose our own

Struct tags are too intrusive — adding `openapi:"minimum=0"` alongside `json:"name"` and
`validate:"required"` creates tag soup. But validation frameworks like `go-playground/validator`
already use tags like `validate:"min=0,max=100"`, and ORM frameworks use `gorm:"type:varchar(255)"`.

**v2 approach**: Read and interpret existing struct tags from popular frameworks rather than
defining our own tag format:

- `validate:"required,min=0,max=100"` → infer `minimum`, `maximum`, `required` in schema.
- `json:"name,omitempty"` → already used for field naming (v1 does this). `omitempty` could
  inform `nullable` or `required` heuristics.
- `binding:"required"` (gin) → same as `validate:"required"`.
- `db:"column_name"` / `gorm:"..."` → potentially useful for naming or type hints, though
  this is a stretch.

The scanner provides a **tag reader registry**: an interface that maps framework-specific
tag syntax to OpenAPI schema properties. Users can register readers for the frameworks they
use. The scanner ships built-in readers for the most common ones (`validate`, `json`, `xml`).

This keeps the annotation comment grammar focused on API-level concerns (routes, operations,
responses) while struct-level validation comes from the code itself.

### 3. Smart detection: infer operations and routes from code structure

Currently the scanner only analyzes type definitions (structs → schemas, interfaces → nothing).
Operations and routes require explicit `swagger:route` / `swagger:operation` annotations.

**v2 approach**: The scanner should infer operations from code patterns, with framework-specific
plugins:

**HTTP handler detection** — Recognize common handler signatures:
- `func(w http.ResponseWriter, r *http.Request)` — stdlib
- `func(c *gin.Context)` — Gin
- `func(c echo.Context) error` — Echo
- `func(ctx *fiber.Ctx) error` — Fiber
- `http.HandlerFunc`, `http.Handler` adapters

**Route detection** — Analyze framework-specific routing registration:
- `mux.HandleFunc("/pets/{id}", getPet)` → `GET /pets/{id}`, operationId `getPet`
- `r.GET("/pets/:id", getPet)` → same (Gin syntax)
- `e.POST("/pets", createPet)` → `POST /pets`

**Parameter inference** — Analyze handler bodies:
- `r.URL.Query().Get("limit")` → query parameter `limit`
- `mux.Vars(r)["id"]` → path parameter `id`
- `json.NewDecoder(r.Body).Decode(&pet)` → request body of type `Pet`
- `c.QueryParam("limit")` (Echo) → query parameter

**Response inference** — Analyze return paths:
- `json.NewEncoder(w).Encode(pets)` → 200 response with type `[]Pet`
- `http.Error(w, "not found", 404)` → 404 response
- `c.JSON(200, pets)` (Gin/Echo) → 200 with type from argument

**The role of annotations shifts**: Instead of being the primary source of spec information,
annotations become **corrections and enrichments**:
- Override an inferred parameter location: `// openapi:in header` on a field
- Add descriptions that can't be inferred: `// openapi:description The maximum number of items`
- Declare responses that the code doesn't make obvious (e.g. middleware-added 401s)
- Supply examples, deprecation notices, external docs links

This means the annotation grammar can be **simpler** than the current one — it doesn't need
to express everything, just the delta between what the scanner infers and what the spec needs.

**Framework plugins**: The detection logic is framework-specific. A plugin interface:

```go
type FrameworkPlugin interface {
    // DetectRoutes analyzes route registration calls in the given package.
    DetectRoutes(pkg *packages.Package) []DetectedRoute

    // DetectHandler returns true if the function signature is a handler for this framework.
    DetectHandler(sig *types.Signature) bool

    // InferParameters analyzes a handler body for parameter access patterns.
    InferParameters(fn *ast.FuncDecl, pkg *packages.Package) []InferredParam
}
```

Ship built-in plugins for `net/http` + `gorilla/mux`, `gin`, `echo`, `chi`, `fiber`.
Community can add more.

## Open Questions

- **Backward compatibility**: Should v2 accept `swagger:` prefixed annotations with a
  deprecation warning, or require migration to `openapi:`? A `codescan migrate` CLI command
  could automate the rewriting. Given the user base, a deprecation period with both prefixes
  accepted seems pragmatic.

- **Private comment syntax**: What format for the richer private comments? Options:
  (a) YAML in `/* */` blocks (familiar but has the indentation issues we're moving away from),
  (b) a custom structured format parsed by the grammar,
  (c) JSON (precise but verbose and ugly in source code).
  Leaning toward (b) — a minimal structured format that's a subset of YAML without the
  indentation sensitivity (e.g. explicit `{`/`}` for nesting, or simply one-key-per-line
  with dotted paths like `response.200.schema: Pet`).

- **Detection confidence and overriding**: When the scanner infers something from code, how
  does the user know what was inferred? A `codescan inspect` command that shows the inferred
  spec before annotations are applied would help users understand what they need to override.
  Also: what happens when inference and annotation conflict? Annotation wins, but should
  there be a warning?

- **Framework plugin discovery**: Built-in vs external plugins. Built-in plugins for the top 5
  frameworks make sense. But the long tail (Buffalo, Revel, custom routers) needs a plugin
  loading mechanism. Go plugin system? Code generation? Or just "implement the interface and
  register it in your main.go"?

- **Scope of code analysis**: Handler body analysis (inferring parameters from `r.URL.Query()`)
  requires analyzing function bodies, not just signatures and types. This is significantly more
  complex than the current AST-level scanning. It may require SSA analysis (`golang.org/x/tools/
  go/ssa`) for accurate dataflow tracking through variables and helper functions. This is a
  large engineering effort — worth scoping as a separate design phase.
