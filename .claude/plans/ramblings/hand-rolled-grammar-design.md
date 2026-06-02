# Hand-Rolled Grammar for Comment Annotation Parsing

Date: 2026-03-25

## The Problem the Grammar Solves

The current parsing architecture pushes an enormous burden onto the builders. Every builder
(schema, parameter, response, route, operation, meta) must:

1. **Construct a tagger list** — 15-20 `NewSingleLineTagParser` / `NewMultiLineTagParser` calls,
   each wrapping a `Set*` type with a `ValidationBuilder` adapter. The schema builder alone
   creates 16 taggers per field, plus 13 more per array nesting level.

2. **Wire up callbacks** — `WithSetTitle(func(lines []string) { op.Summary = JoinDropLast(lines) })`
   and `WithSetDescription(func(lines []string) { ... })` repeated in every builder.

3. **Handle items nesting** — `parseArrayTypes` recursively constructs `items0Maximum`,
   `items1Maximum`, ... tagger names with `WithItemsPrefixLevel`. Each nesting level adds 13
   taggers. A `[][][]int` creates 39 extra taggers.

4. **Know which taggers apply where** — a `$ref` schema gets only `required`; an inline schema
   gets all 16; a header gets the validation set but not `required`/`readOnly`/`discriminator`;
   a parameter adds `in`/`required`/`collectionFormat`. Every builder re-derives this knowledge.

The grammar replaces all of this with: **parse the comment into a typed AST, hand it to the
builder, let the builder read what it needs.**

## What the Builders Would Look Like After

### Before (current — response builder)

```go
taggers, err := setupResponseHeaderTaggers(&ps, name, afld)
if err != nil {
    return err
}

sp := parsers.NewSectionedParser(
    parsers.WithSetDescription(func(lines []string) { ps.Description = parsers.JoinDropLast(lines) }),
    parsers.WithTaggers(taggers...),
)
if err := sp.Parse(afld.Doc); err != nil {
    return err
}
```

Where `setupResponseHeaderTaggers` is 28 lines of tagger construction.

### After (with grammar)

```go
block, err := parsers.ParseComment(afld.Doc)
if err != nil {
    return err
}

ps.Description = block.Description()
block.ApplyValidations(&ps)   // sets maximum, minimum, pattern, etc.
```

The builder doesn't know about taggers, regexps, or parsing modes. It receives a `CommentBlock`
with typed accessors.

## Grammar Design

### Lexer

The lexer operates on a single comment line (already stripped of `//` / `*` prefixes by a
preprocessing step). It produces one of these token types:

```
ANNOTATION    swagger:model Foo       → (kind="model", args=["Foo"])
KEYWORD_VALUE maximum: < 100          → (keyword="maximum", op="<", value="100")
KEYWORD_ONLY  consumes:               → (keyword="consumes")
ITEMS_PREFIX  items.items.maximum: 5  → (prefix=["items","items"], keyword="maximum", value="5")
YAML_FENCE    ---                     → signals start/end of YAML block
TEXT          any line not matching above
BLANK         empty / whitespace-only line
```

Keyword recognition is a **case-insensitive map lookup**, not a regexp. The map is:

```go
var keywords = map[string]KeywordKind{
    "maximum":           KindMaximum,
    "max":               KindMaximum,    // alias
    "minimum":           KindMinimum,
    "min":               KindMinimum,    // alias
    "multipleof":        KindMultipleOf,
    "multiple of":       KindMultipleOf, // alias with space
    "pattern":           KindPattern,
    "minlength":         KindMinLength,
    "min length":        KindMinLength,
    "maxlength":         KindMaxLength,
    "max length":        KindMaxLength,
    "minitems":          KindMinItems,
    "min items":         KindMinItems,
    "maxitems":          KindMaxItems,
    "max items":         KindMaxItems,
    "unique":            KindUnique,
    "enum":              KindEnum,
    "default":           KindDefault,
    "example":           KindExample,
    "required":          KindRequired,
    "readonly":          KindReadOnly,
    "read only":         KindReadOnly,
    "discriminator":     KindDiscriminator,
    "collectionformat":  KindCollectionFormat,
    "collection format": KindCollectionFormat,
    "in":                KindIn,
    "deprecated":        KindDeprecated,
    "schemes":           KindSchemes,
    "consumes":          KindConsumes,
    "produces":          KindProduces,
    "security":          KindSecurity,
    "responses":         KindResponses,
    "parameters":        KindParameters,
    "extensions":        KindExtensions,
    // ... v2 additions for OAI 3.x
}
```

The lexer algorithm for a single line:

1. Strip the `items.` prefix chain, recording the nesting depth.
2. Find the first `:` that's not inside a pattern or value.
3. Everything before `:` is the keyword candidate — normalize (lowercase, collapse whitespace).
4. Look up in the keyword map.
5. If found → produce `KEYWORD_VALUE` or `KEYWORD_ONLY` (if nothing after colon).
6. If not found but line starts with `swagger:` / `openapi:` / `@` → produce `ANNOTATION`.
7. If `---` → produce `YAML_FENCE`.
8. Otherwise → produce `TEXT`.

**This replaces all 62 regexps with one map lookup + simple string splitting.**

### Items prefix handling

The `items.` prefix is handled by the lexer, not the parser or the taggers:

```go
func (l *Lexer) stripItemsPrefix(line string) (depth int, rest string) {
    rest = line
    for {
        // case-insensitive "items" followed by "." or whitespace
        lower := strings.ToLower(rest)
        if !strings.HasPrefix(lower, "items") {
            return depth, rest
        }
        after := rest[5:]
        if len(after) == 0 {
            return depth, rest
        }
        if after[0] == '.' || after[0] == ' ' {
            depth++
            rest = strings.TrimLeft(after[1:], " ")
            continue
        }
        return depth, rest
    }
}
```

The parser receives `Token{Keyword: KindMaximum, ItemsDepth: 2, Value: "100"}` and knows this
is `items.items.maximum: 100` without any regexp.

### Parser

The parser consumes the token stream and produces a `CommentBlock` AST:

```go
type CommentBlock struct {
    Annotation  *Annotation       // swagger:model Foo, swagger:route GET /pets ...
    Title       string            // first paragraph (before blank line)
    Description string            // remaining text paragraphs
    Properties  []Property        // keyword: value lines
    Blocks      []MultiLineBlock  // consumes:, produces:, security:, extensions:, YAML
}

type Annotation struct {
    Kind string   // "model", "route", "operation", "response", "parameters", ...
    Args []string // ["Foo"] or ["GET", "/pets", "pets", "listPets"]
}

type Property struct {
    Keyword    KeywordKind
    ItemsDepth int      // 0 = direct, 1 = items., 2 = items.items., ...
    Operator   string   // "<", ">", "=" (for exclusive min/max), "" otherwise
    RawValue   string   // the unparsed value string
}

type MultiLineBlock struct {
    Keyword KeywordKind
    Lines   []string // body lines (not including the header)
}
```

The parser algorithm (recursive descent, though the grammar is flat enough that it's really
just a state machine):

```
parse_comment_block:
    if peek() == ANNOTATION:
        block.Annotation = parse_annotation()

    // collect header lines (text before any keyword)
    while peek() == TEXT or peek() == BLANK:
        append to header_lines

    split header_lines into title + description

    // collect properties and blocks
    while peek() != EOF and peek() != ANNOTATION:
        if peek() == KEYWORD_VALUE:
            block.Properties = append(block.Properties, parse_property())
        elif peek() == KEYWORD_ONLY:
            block.Blocks = append(block.Blocks, parse_multiline_block())
        elif peek() == YAML_FENCE:
            block.Blocks = append(block.Blocks, parse_yaml_block())
        elif peek() == TEXT or peek() == BLANK:
            // text after keywords started — belongs to current multiline block
            // or is ignored
            consume()
```

### What the builders receive

The `CommentBlock` provides typed accessors that replace the entire tagger machinery:

```go
// Validations returns all validation properties, optionally filtered by items depth.
func (b *CommentBlock) Validations(itemsDepth int) []Property

// Get returns the first property with the given keyword at depth 0, or nil.
func (b *CommentBlock) Get(keyword KeywordKind) *Property

// GetFloat parses the property value as float64, handling the operator for exclusive min/max.
func (p *Property) GetFloat() (value float64, operator string, err error)

// GetInt parses the property value as int64.
func (p *Property) GetInt() (int64, error)

// GetBool parses the property value as bool.
func (p *Property) GetBool() (bool, error)

// ApplyValidations writes all validation properties at the given items depth
// into a ValidationBuilder. This single method replaces 13 Set* types.
func (b *CommentBlock) ApplyValidations(target ValidationBuilder, itemsDepth int) error
```

The schema builder goes from:

```go
// BEFORE: 16 taggers + 13 per items level + SectionedParser + callbacks
taggers := schemaTaggers(schema, ps, nm)
if fld != nil {
    if ftped, ok := fld.Type.(*ast.ArrayType); ok {
        taggers, err = parseArrayTypes(taggers, ftped.Elt, ps.Items, 0)
        // ...
    }
}
sp := parsers.NewSectionedParser(
    parsers.WithSetDescription(func(lines []string) { ps.Description = parsers.JoinDropLast(lines) }),
    parsers.WithTaggers(taggers...),
)
sp.Parse(afld.Doc)
```

To:

```go
// AFTER: parse once, read what you need
block, err := parsers.ParseComment(afld.Doc)
if err != nil { return err }

ps.Description = block.Description
block.ApplyValidations(schemaValidations{ps}, 0)

// items nesting — the block already has the depth encoded
for depth := range arrayDepth(fld) {
    block.ApplyValidations(schemaValidations{itemsAt(ps, depth)}, depth+1)
}
```

The response header builder goes from 28 lines of tagger setup to:

```go
block, err := parsers.ParseComment(afld.Doc)
if err != nil { return err }

ps.Description = block.Description
block.ApplyValidations(headerValidations{&ps}, 0)
```

## Pluggable Annotation Styles

The lexer's annotation recognition is controlled by a `StyleRecognizer` interface:

```go
type StyleRecognizer interface {
    // MatchAnnotation checks if a line starts an annotation in this style.
    // Returns the annotation kind and arguments, or false if not recognized.
    MatchAnnotation(line string) (kind string, args []string, ok bool)
}
```

Built-in styles:

```go
// SwaggerStyle recognizes "swagger:model Foo", "swagger:route GET /pets tags listPets"
type SwaggerStyle struct{}

// OpenAPIStyle recognizes "openapi:model Foo" — same syntax, different prefix
type OpenAPIStyle struct{}

// AtStyle recognizes "@model Foo", "@route GET /pets" — javadoc-like
type AtStyle struct{}
```

The lexer accepts multiple recognizers and tries them in order:

```go
lexer := NewLexer(
    WithStyles(SwaggerStyle{}, OpenAPIStyle{}),     // accept both
    WithDeprecatedStyle(SwaggerStyle{}, "openapi"), // warn on swagger:, suggest openapi:
)
```

This means:
- **v1 compatibility**: accept `swagger:` prefix.
- **v2 canonical**: `openapi:` prefix.
- **User preference**: `@` prefix for those who prefer javadoc-style brevity.
- **Migration tooling**: the lexer can report which style was used, enabling a `codescan migrate`
  that rewrites annotations.

Keyword recognition is **independent of annotation style**. `maximum: 10` works the same
regardless of whether the block started with `swagger:model` or `@model`.

## Rigorous Multi-Line Comment Processing

### The current pain

Today, multi-line handling is scattered:

- `cleanupScannerLines` strips comment prefixes (`// `, `* `, etc.) — called in multiple places.
- `collectScannerTitleDescription` splits title from description — duplicated in SectionedParser
  and YAMLSpecScanner.
- Multi-line taggers (consumes, produces, security, extensions) each implement their own body
  collection.
- YAML blocks have a separate code path with their own indent removal (`removeIndent`,
  `removeYamlIndent`).
- The `SkipCleanUp` flag exists because YAML taggers need raw indentation — a sign that the
  cleanup is happening at the wrong level.

Each builder has to know:
- "Use `JoinDropLast` to collapse lines into a string."
- "Append the enum description extension to the description."
- "The title is the `Summary` field, not `Title`."

### The grammar solution

Comment prefix stripping happens **once**, in the lexer, before anything else. The lexer
receives `*ast.CommentGroup` and produces clean tokens:

```go
func NewLexer(doc *ast.CommentGroup, opts ...LexerOption) *Lexer {
    // 1. Flatten all comments into lines.
    // 2. Strip comment prefixes: "// ", "/* ", " * ", "// |" (whitespace-preserving).
    // 3. Track original positions for error reporting.
    // 4. Tokenize.
}
```

After this, no downstream code ever sees `// ` prefixes. The `cleanupScannerLines` function
and `rxUncommentHeaders` regexp disappear entirely.

Title/description splitting happens **once**, in the parser, with a clear rule:

```
Title:       Everything before the first blank line (or the first line ending in
             punctuation when there's no blank line).
Description: Everything after the title, up to the first keyword or annotation.
```

This is codified in the grammar, not reimplemented by each builder.

Multi-line block collection happens **in the parser**, uniformly:

```
multiline_block:
    KEYWORD_ONLY           // the header line (e.g. "consumes:")
    { TEXT | BLANK }       // body lines until next keyword or annotation
```

The parser collects body lines into `MultiLineBlock.Lines`. The builder receives them
already collected — no need for the tagger to track "am I currently inside a multiline
block?" state.

YAML blocks are recognized by the `---` fence:

```
yaml_block:
    YAML_FENCE             // "---"
    { TEXT | BLANK }       // YAML content
    [ YAML_FENCE ]         // optional closing "---"
```

The parser hands the raw YAML lines to a YAML-specific handler. The indentation is preserved
because the lexer stripped only the comment prefix, not content indentation.

### What disappears from the builders

| Current builder concern | With grammar |
|------------------------|--------------|
| Construct 15-20 taggers per field | Gone — `ApplyValidations` does it all |
| Wire `WithSetTitle` / `WithSetDescription` callbacks | Gone — read `block.Title`, `block.Description` |
| Handle `items.` prefix nesting with `PrefixRxOption` | Gone — lexer encodes `ItemsDepth` in token |
| Call `JoinDropLast` on description lines | Gone — `block.Description` is already joined |
| Choose between `NewSingleLineTagParser` and `NewMultiLineTagParser` | Gone — parser handles line collection |
| Set `SkipCleanUp` for YAML taggers | Gone — cleanup is in the lexer, not per-tagger |
| `parseArrayTypes` recursive tagger construction | Gone — `ApplyValidations(target, depth)` |
| Know which taggers apply to `$ref` vs inline schemas | Simpler — `if isRef { skip validations }` |
| `setupRefParamTaggers` vs `baseInlineParamTaggers` vs `setupInlineParamTaggers` | One call: `block.ApplyValidations(target, 0)` |

The schema builder's `createParser` (43 lines) + `schemaTaggers` (29 lines) + `itemsTaggers`
(21 lines) + `parseArrayTypes` (15 lines) = **108 lines** that reduce to ~10 lines.

Multiply across parameter, response, and route builders — the total savings is **~250 lines**
of tagger boilerplate removed from the builders.

## Grammar as Documentation

The EBNF can be included in user-facing docs and generated into syntax diagrams:

```ebnf
(* Top-level structure of a swagger/openapi comment block *)
comment_block = [ annotation NL ]
                header
                { property | multiline_block | yaml_block } .

(* Annotation: the swagger:xxx or openapi:xxx or @xxx line *)
annotation    = annotation_prefix annotation_kind { SP arg } .
annotation_prefix = "swagger:" | "openapi:" | "@" .
annotation_kind   = "model" | "response" | "parameters" | "route" | "operation"
                  | "allOf" | "strfmt" | "enum" | "name" | "type"
                  | "default" | "alias" | "file" | "ignore" | "meta" .

(* Header: free-form text split into title + description *)
header        = { text_line } .
text_line     = (* any line that is not a keyword line, annotation, or fence *) .

(* Property: keyword: value on a single line *)
property      = [ items_prefix ] keyword ":" [ operator ] value NL .
items_prefix  = { "items" ( "." | SP ) } .
keyword       = "maximum" | "minimum" | "multiple of" | "pattern"
              | "min length" | "max length" | "min items" | "max items"
              | "unique" | "enum" | "default" | "example"
              | "required" | "read only" | "discriminator"
              | "collection format" | "in" | "deprecated" | "schemes" .
operator      = "<" | ">" | "<=" | ">=" | "=" .
value         = (* rest of line after colon and optional operator *) .

(* Multi-line block: header keyword followed by indented body lines *)
multiline_block = block_keyword ":" NL { body_line } .
block_keyword   = "consumes" | "produces" | "security" | "responses"
                | "parameters" | "extensions" .
body_line       = (* any non-keyword line while inside the block *) .

(* YAML block: fenced YAML content for complex structured data *)
yaml_block    = "---" NL { yaml_line } [ "---" NL ] .
yaml_line     = (* any line until next fence or annotation *) .

(* Shared terminals *)
SP            = (* one or more whitespace characters *) .
NL            = (* newline *) .
arg           = (* whitespace-delimited token *) .
```

This EBNF is **both the formal specification and the implementation guide**. The hand-rolled
parser has one function per non-terminal. When someone asks "what annotations does the tool
support?", point them at the grammar.

## Implementation Roadmap

### Phase 1: Lexer (standalone, testable)

- `Lexer` type: `*ast.CommentGroup` → `[]Token`.
- Keyword map.
- Items prefix stripping.
- Comment prefix stripping (replaces `cleanupScannerLines` and `rxUncommentHeaders`).
- Style recognizers.
- Comprehensive tests: one test per token type, edge cases for case sensitivity, whitespace
  variants, mixed styles.

Can be built and tested independently of the parser. Doesn't touch existing code.

### Phase 2: Parser (standalone, testable)

- `Parser` type: `[]Token` → `CommentBlock`.
- Title/description splitting (replaces `collectScannerTitleDescription`).
- Property collection.
- Multi-line block collection.
- YAML fence handling.
- Error recovery: skip to next line on parse error, collect diagnostics.

Can be tested with constructed token sequences — no need for real Go source files.

### Phase 3: CommentBlock accessors

- `ApplyValidations(target, depth)` — replaces all `Set*` types and tagger lists.
- `GetFloat`, `GetInt`, `GetBool` — replaces `strconv.Parse*` scattered through `Set*.Parse`.
- Typed accessors for annotation args (route method/path/tags/id, model name, etc.).

### Phase 4: Integration

- Replace `SectionedParser` + `TagParser` calls in builders with `ParseComment` + accessors.
- Run old and new parsers in parallel, assert identical spec output on the full fixture suite.
- Remove old parsing code once parity is confirmed.

### Phase 5: LSP hooks (future)

- Expose `Lexer` token stream for completion (available keywords at cursor position).
- Expose `Parser` diagnostics for editor squiggles.
- Expose grammar's keyword list for documentation generation.

## Risks and Mitigations

**Risk**: Behavioral differences between old regexps and new lexer.
**Mitigation**: Golden-file testing. Run both parsers on every fixture, diff the spec output.
The existing 96.5% test coverage of the parsers package gives confidence.

**Risk**: Performance regression — lexer + parser overhead vs compiled regexps.
**Mitigation**: The current approach runs up to 75 regexp matches per field. The lexer does
one map lookup. This should be faster, not slower. Benchmark before/after.

**Risk**: The grammar becomes a maintenance burden.
**Mitigation**: The grammar is simpler than 62 regexps. The EBNF above is 30 lines. The
hand-rolled parser is ~300 lines (comparable to the current `sectioned_parser.go` alone). Net
code reduction across the codebase.

**Risk**: Users rely on undocumented regexp quirks.
**Mitigation**: The test suite's cartesian-product approach already exercises the full space
of accepted inputs. Any quirk that's in the tests is in the grammar. Quirks not in the tests
are bugs that should be fixed, not preserved.
