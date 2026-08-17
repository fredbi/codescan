# Grammar parser — architecture (meta-plan)

Date: 2026-04-20 (revised 2026-04-21)
Status: **§§1–5 all settled 2026-04-21.** §3.2.1 (enum), §3.5
(tagger-distinction), §4 (iterator / typed `Block` / cutover), and
§5 (Option E adopted, `keywords.go` as typed slice, parser-as-interface
for property-based builder tests) finalized in this pass. Ready for
task-level implementation plan.

This document frames the architectural choices for replacing the regexp-based
annotation parser with a grammar-based one. It is the meta-plan; a task-level
plan comes later.

Related:
- `.claude/plans/ramblings/vision.md` — 3-step v2 roadmap (this is step 1)
- `.claude/plans/ramblings/grammar-vs-regexp-for-v2.md` — why grammar, not regexp
- `.claude/plans/ramblings/hand-rolled-grammar-design.md` — earlier concrete design
  (some positions below differ; this doc supersedes where they conflict)

---

## 1. Objectives — what the parser must do

### 1.1 Capability table

| # | Capability | Source of need | Scope for v2.0? |
|---|-----------|----------------|------------------|
| C1 | Strip comment noise once (`//`, `/* */`, ` * `, leading table pipes) | Current code does it in ~3 places | Yes |
| C2 | Recognize annotation kind + positional args (e.g., `swagger:route GET /path tags opid`) | 14+ annotation kinds today, more in OAI 3.x | Yes |
| C3 | Parse `keyword: value` with ~35 keyword aliases, case-insensitive | validations + response/param/route properties | Yes |
| C4 | Handle `items.N.X` property nesting (depth-parameterised) | Array validations | Yes |
| C5 | Collect multi-line keyword blocks (`consumes:`, `security:`, `responses:`, `extensions:`) | Meta, route, model, response builders | Yes |
| C6 | Isolate embedded YAML fenced by `---` (swagger:operation body + extensions inline YAML) | Legacy `swagger:operation` + some extensions | Yes (isolate) — parse details is TBD (§3) |
| C7 | Split title + description paragraphs from body | Every annotation target | Yes |
| C8 | Report diagnostics with file/line/column positions, accumulate, don't abort on first error | LSP + better errors today | Yes for infra; LSP consumption later |
| C9 | Pluggable annotation-style prefixes (`swagger:`, `openapi:`, `@`) | v2 migration, user preference | **Deferred to v2.x** — natural evolution if Option E's keyword table is well-factored; don't add ceremony for it now |
| C10 | Parse once per comment group, reusable typed AST | Replace 15–20 taggers per field | Yes |
| C11 | Format-neutral output (no `spec.*` types in parser API) | Enables step 3 IR work | Yes |
| C12 | Type-convert primitive values (numbers, booleans) | Today scattered in `Set*.Parse` | Yes |
| C13 | Position tracking preserved through AST nodes | Error messages + LSP + `codescan explain` | Yes — needed throughout |
| C14 | Support OAI 3.x additions (`requestBody`, `content`, `servers`, `callbacks`, `links`, `cookie`, OIDC…) | Sizes the parser's expressive-power bar | **Deferred to v2.x** — OAI 3.x keywords land as follow-on work. v2.0 stays at v1 keyword parity (grammar replaces regexps, same language). Notably JSONSchema draft evolution (draft4 → 7+) is the larger churn, not the OAI 2→3 shape change. |
| C15 | Partial/incremental parsing (re-parse a single comment block) | LSP + lightweight-first scan (step 2 AnnotationScan) | **Yes — Day 1.** Essential complexity: the lightweight-first scan needs to emit signals from comments without type info. The same parser invoked later with full AST produces full results. |
| C16 | Context-aware keyword validity (e.g., `in:` legal under parameter/header, not schema) | LSP completion + analyzer diagnostics | Yes — via a context table keyed by annotation kind (see §2.2). Parser stays context-free; the table is adjacent data. |
| C17 | Round-trip preservation (comment → AST → regenerate comment) | Migration CLI (`swagger:` → `openapi:`) | **Out of scope** — defer to migrate tool, which can do source rewrites |
| C18 | Smart detection (handler signature → inferred params) | v2 differentiator | **Out of parser scope** — separate subsystem operating on `go/types`, not comments |

**Consolidated notes carried from review:**

- **C1** — reuse `(*ast.CommentGroup).Text()` from the Go stdlib rather than
  reimplementing comment-prefix stripping. It already drops `// ` / `* ` /
  whitespace. Our preprocessor builds *on top* of it (tracking positions,
  handling markdown table pipes, preserving indentation inside fenced blocks).
- **C4** — `items.N.X` support is kept for backward compatibility only.
  A developer-friendlier syntax for nested-array validations is worth a
  dedicated workshop (see §1.6).
- **C5** — multi-line blocks currently mix YAML-ish and markdown-ish flavors
  inconsistently. v2 cleans this up, partly by allowing such blocks to live
  in *private* (non-godoc) comments where richer structure won't pollute
  rendered API docs.
- **C6** — enum and example annotations currently have multiple representations
  (comma-list, YAML list, godoc-style). Unifying their surface is a workshop
  (see §1.6) and out of this architecture's scope.
- **C11** — "format-neutral" is the handoff point where builders become
  visitors over the AST.
- **C18** — confirmed out of scope *for the parser*. The scanner layer is
  where code-signal sources live. Longer-term, the scanner should treat
  functions as first-class operation handler candidates rather than
  shoehorning them into type-centric models.

### 1.2 Non-goals for v2.0 (parser layer)

- **Round-trip emission.** The parser understands comments; it does not write
  them back. Migration tooling can do source rewrites separately.

  **YES totally agree**

- **Macro expansion / includes.** Annotations reference types by Go name, not
  by textual include. No need for inclusion semantics. Bulk merge of external
  content is already handled by `InputSpec` (pre-filled base spec overlay),
  and philosophically: annotated source should not become another maze of
  includes. **Carve-out:** `externalDocs` (a first-class OpenAPI feature
  already supported in v2) is in scope — it's a spec property, not a macro.

- **Handler body analysis.** That's `go/types` + SSA territory, not comment
  parsing. The parser's role is understanding *comment* intent.

  **YES we study a comment parser** At some point we might need a more capable _code scanner_, but that's a different layer in the architecture.

- **Spec-level validation.** "This `in: body` param conflicts with this
  `consumes: application/x-www-form-urlencoded`" is analyzer concern.

  **Indeed. There are many ways to botch comments into an invalid spec**. This also advocate for a PLS diagnostic tool that covers a larger
  scope than mere parsing.

### 1.3 Contract with step 2 (demand-driven loader) — signals & dispatcher

The parser is a **signal source**. So is the scanner (via code analysis).
Both emit observations about source; a **dispatcher** routes them to the
interested builders.

```
                          ┌─────────────────┐
  comments ──► Parser ───►│                 │
                          │    Dispatcher   │───► Builder(s)
  AST/types ──► Scanner ─►│                 │     (schema, operations, …)
                          └─────────────────┘
```

**Signal examples:**

- *From parser:* "file X pos P has `swagger:model Foo`" · "field bar has
  `in: query` with validations {...}" · "function Y has a `swagger:route`
  block with method=GET path=/pets".
- *From scanner (later):* "function Z's signature matches an
  `http.HandlerFunc` — could be an operation" · "const group K looks like
  an enum" · "this `http.HandleFunc("GET /api", myFunc)` call looks like a
  route registration" · "this type implements `TextMarshaler` — strfmt
  candidate".

The dispatcher is a **distinct architectural element** — probably a peer of
`DemandLoader` and `SchemaCache` in step 2. Its design is out of scope for
the parser plan, but the parser API must be shaped so its output plugs in
cleanly (structured signals, positions, typed AST, no spec-coupled side
effects).

**Two parser invocation modes:**

1. **Early, cheap, comment-only** — during AnnotationScan root-discovery.
   No type info required. Goal: emit root signals fast. The parser works
   on comment text alone.
2. **Late, fully loaded** — during schema/operation building, when types
   are resolved. Same parser; same AST shape. Analyzers correlate AST
   contents against the resolved Go types.

**Structured-text sub-languages stay out of the core parse path.** The
parser isolates signal + arguments (e.g., "here's a YAML block starting at
pos P with content X"), and dedicated sub-parsers — YAML, markdown, a
future richer enum/example format — handle the rest. This keeps the core
parser small and stdlib-only. See §3 for the split.

**Implication:** the parser must be usable *without* `packages.Package` or
`*types.Named`. It works on text (via `*ast.CommentGroup` or raw comment
strings). Type resolution is an analyzer concern, happening after parsing.

> NOTES(fred): existing codebase
>
> Pure parser(s) live in internal/parsers
> They are contextualized in the Taggers build by each internal/builder/*
>
> When we make the first move of parsers, the Taggers should be replaced.
>
> Currently, the dispatching functionality is part of the TypeIndex type (processFile method).
> It may stay there temporarily, while we are building the parsing layer.
>
> All signals are currently identified in method processFile() and its dependencies, most notably detectNode().
> All these are candidates for fast-parsing a comment.
>
> In the existing codebase, "signals" are stashed in the type index and walked in a second pass
> by the spec builder. This may 2-phase mechanism may stay when building the parser and
> will go when the "on-demand" loop will be implemented.
>
> The comment parser first-use case is to report the signals (e.g. equivalent to "ExtractAnnotation").

### 1.4 Contract with step 3 (IR + renderer)

The parser's output is a typed AST. Builders translate AST → IR. The parser
must not import `spec.*` (it already doesn't) and must not assume a
particular output format. The AST uses plain Go: strings, numbers, booleans,
enums for keyword kinds.

**Technology note:** bridging AST-like structures to typed trees is a
tractable problem here — the same shape of code already exists as a
by-product of the `go-openapi/testify` codegen work.

**Is AST = IR?** No — they answer different questions. AST = "what did the
comment say, syntactically?". IR = "what API does this describe,
semantically?". The IR is defined by the builder stage (a later plan), not
by the parser, and has its own version-independent validation rules
("can I annotate a function as a model?"). A simplified view of the IR:
a merge of OAI v2 + v3 concepts, renderable to any of the supported output
formats. The ambitious view: renderable to gRPC, JSON-RPC, OpenAPI vX,
etc. (GraphQL is an acknowledged different beast, out of scope.)

The parser's job is to not foreclose any of this. Shape the AST for
clarity about *what the comment said*; let the builder layer decide what
to do with it.

> NOTES(fred): existing codebase
>
> In each builder, the taggers currently weave the spec when consuming source code
> declarations from the TypeIndex.
>
> Currently the whole source code is parsed with full types.
>
> In our target, source code searched for signals of interest (AST only), then parsed for types
> on demand when the builders start walking the signaled node.
>
> When we implement the on-demand parsing, these will consume the "interesting bits" signaled and
> their walking may demand more code scanning.

### 1.5 Scope decisions (previously "debate openers", now settled)

| Capability | Decision | Rationale |
|-----------|----------|-----------|
| **C15** — incremental parsing | **Day 1** | Essential complexity. The lightweight-first AnnotationScan and the later full parse are the *same* parser invoked differently. Retrofitting would require untangling hidden globals. |
| **C9** — pluggable annotation styles (`swagger:`, `openapi:`, `@`) | **v2.x** | Don't pre-engineer. A well-factored keyword table lets the prefix be parameterized later without re-architecting. |
| **C14** — OAI 3.x keywords | **v2.x** | v2.0 stays at v1 keyword parity (grammar replaces regexps, same language; deviations only where they're bug fixes from the Q-pass). OAI 3.x is used only to *size* the parser's expressive power in v2.0 design. |

### 1.6 Workshops pinned for later

The architecture must **not foreclose** these; it must leave room for the
decisions. Each is a small focused design session, not a blocker on the
parser work:

- **W1** — `items.N.X` nested-validations: friendlier alternative syntax
  while keeping the current form supported for backward compatibility.
- **W2** — enum annotation format: reconciling comma-list / YAML / godoc
  variants. Interacts with the discovery of enum types from Go constants
  (scanner signal "this const group looks like an enum").
- **W3** — example annotation format: similar to W2, plus type-aware
  examples (the value's Go type determines the serialization).
- **W4** — private-comment support: annotations allowed outside godoc,
  in `/* */` blocks or unattached `//` blocks, where richer structure
  won't pollute rendered Go API documentation.
- **W5** — `externalDocs` annotation syntax (in scope per §1.2 but
  specific syntax is a workshop).
- **W6** — OAI 3.x keyword onboarding (v2.x scope per C14).
- **W7** — LSP context-validity tables (per C16).

---

## 2. Grammar class — what formalism do we use?

### 2.1 What the annotation language actually looks like

A comment block has this structure (composed from the go-swagger docs and
the 62 regexps' behavior):

```
<annotation-line>?                     e.g. "swagger:model User"
<title-paragraph>                      free text, until blank line
<blank>
<description-paragraphs>               free text, until a keyword line
<property-line>*                       "keyword: value" with optional items.N. prefix
<multiline-block>*                     "consumes:" followed by body lines
<yaml-block>?                          "---" ... "---" (for swagger:operation)
```

Four structural layers:

1. **Envelope** (annotation → title → desc → body) — context-sensitive,
   indentation-aware, small (~6 productions).
2. **Keyword-value grammar** — big flat table of ~35 keywords with aliases,
   each with a value type (number, bool, string, comma-list, regex, path).
3. **Enum mini-language** — today a hybrid surface: `swagger:enum` on a
   type declaration, paired with comment-driven value extraction and
   multiple value-list syntaxes (comma-list / YAML / godoc-style). Lives
   between the keyword-value layer (when an `enum:` line appears) and
   sub-languages (when values come as a YAML block). Explicitly called out
   because it's the layer most likely to change shape per W2.
4. **Sub-languages** — YAML block content, extensions' nested structures,
   route positional args, richer private-comment bodies. Each is its own
   mini-grammar; the core parser isolates them and hands them off.

**Key observation:** the language is not uniform. One formalism per layer
probably fits better than a single mechanism trying to cover all four.
This shapes §2.2.

### 2.2 Formalism options

| Option | Envelope | Keyword table | Sub-languages |
|--------|---------|---------------|---------------|
| **A — pure hand-rolled** | Recursive descent by hand | Go map | Hand-written mini-parsers |
| **B — EBNF + codegen (custom tool)** | Generated from EBNF | Generated from table | Hand-written |
| **C — EBNF + external tool (participle/pigeon)** | Generated | Generated | Generated |
| **D — parser combinators (participle runtime)** | Struct tags | Struct tags | Struct tags |
| **E — hybrid: hand-rolled envelope + data-driven keyword table** | Hand-rolled | YAML/TOML table → generated Go map + docs | Hand-written |

**Preferences captured:** parser combinators as a concept are appealing, but
struct-tag combinators (participle) are not — an external grammar definition
is preferred. EBNF (external file) is acceptable. Pigeon maintenance status
is questionable as of 2026. Other EBNF-driven options exist (e.g.
`golang.org/x/exp/ebnf`, `alecthomas/participle/v2/ebnf`), and other
combinator libraries exist beyond participle.

### 2.2.1 The context-handling question

> "After `swagger:parameters`, we recognize `in:`. After `swagger:model`, we
> recognize `maximum` but not `in:`. How does the grammar express that?"

Three options:

- **(i)** Parser is context-free; tokenization and parsing succeed for any
  recognized keyword regardless of enclosing annotation. A **context table**
  (AnnotationKind → allowed keywords) lives adjacent to the keyword data;
  analyzers/LSP consult it to reject invalid usage or offer completions.
- **(ii)** Parser carries a mode determined by the annotation line; rejects
  unknown-in-context keywords at parse time.
- **(iii)** Multiple grammars, one per annotation kind; parser dispatches
  on kind and switches grammar.

**Proposed position: (i).** The parser stays context-free. Why:

- The parser should not know the builder taxonomy. Coupling "which keywords
  are legal for parameters" into the parse path re-entangles recognition
  and semantics.
- LSP completion needs exactly this: "given the current annotation kind,
  which keywords are valid?" — a pure table query, not a parser feature.
- Error recovery is simpler: parse the line, then diagnose "keyword `in:`
  not valid under `swagger:model`" as a separate pass.
- The context table is just more data, generated from the same keyword
  source file as the main table. One source of truth.

(ii) couples concerns. (iii) multiplies them. (i) keeps them orthogonal.

This answers C16 from §1.1.

### 2.3 Position-by-position

**A — pure hand-rolled.** Fastest to ship. No tooling. Drift between
"grammar as prose spec" and "parser as code" is inevitable. LSP can't
introspect the grammar. Docs hand-maintained.
**Cost:** low upfront, medium ongoing (drift policing).
**Fits C14?** Poorly — each OAI 3.x / JSONSchema draft keyword is
hand-added in two places. Note: JSONSchema's draft evolution (draft4 →
draft7+) is the larger churn here, not the OAI 2→3 structural change.

**B — EBNF + custom codegen.** We own a small generator (~500 lines) that
eats an EBNF grammar file and emits Go code. Grammar is the source of
truth; docs and code derive from it.
**Cost:** high upfront (generator + grammar), low ongoing.
**Fits C14?** Well — add keywords by editing the grammar.
**Risk:** we maintain a parser generator. These tend to grow features
over time. Scope creep.

**C — EBNF + external codegen.** Same benefit as B but we don't own the
generator. Candidates: `golang.org/x/exp/ebnf` (small, stable, experimental
but simple), `alecthomas/participle/v2/ebnf` (integrated with the
combinator runtime), `mna/pigeon` (PEG, maintenance status uncertain in
2026).
**Cost:** medium upfront. External-tool dependency at build time.
**Fits C14?** Well for grammar; less well for error recovery (pigeon and
participle both have weak recovery).
**Risk:** generator idioms don't match our error-message quality needs —
fighting the tool instead of iterating on it.

**D — participle combinators.** Grammar in Go struct tags, runtime parsing.
Other combinator libraries exist (`tombenke/parc`, etc.); participle is
representative.
**Cost:** low upfront. Runtime dep.
**Fits C14?** Grammar scales but struct-tag syntax is not readable at that
size. Error recovery is limited. LSP introspection via struct reflection
works but is awkward.

**E — hybrid.** The structural envelope is small enough (~6 productions)
that hand-rolling is fine and leaves full control over error recovery and
position tracking. The keyword table is data (YAML/TOML), consumed by
`go generate` to produce both the Go lookup map *and* the context table
(§2.2.1) *and* a markdown doc page. Sub-languages are each hand-written
where needed — YAML and markdown in particular.
**Cost:** low-medium upfront. Tiny generator (~50–200 lines, depending on
how many artifacts it emits).
**Fits C14?** Well — new keywords are additions to the data file.
**LSP introspection:** trivial — the keyword + context tables are queryable.
**Risk:** two sources of truth (code for envelope, data for keywords). But
they answer different questions; the boundary is clean.

### 2.4 Proposed position

**Option E.** The annotation language has two cleanly separable parts
(envelope + keyword table), plus isolated sub-languages. Using one
mechanism per layer is simpler than forcing one mechanism to cover all.

- **EBNF is a hammer, our envelope is a nail.** A custom EBNF-driven
  generator (B) pays its cost on a grammar with dozens of productions and
  recursive structure. Our envelope is flat.
- **Keyword lookup is a table, not a grammar.** Putting a 35-entry
  keyword table through EBNF is Using A Generator When A Map Would Do.

**On the "more advanced grammar could improve UX" question.** The two
credible candidates are: (a) replacing `items.items.items.X` with a nested
syntax (W1), and (b) replacing opaque YAML blocks with a richer structured
annotation grammar (W2/W3/W4). **Neither changes the formalism choice.**
Both add new *keywords* (or new *sub-languages*) to the existing layers.
W1 extends the keyword table with a new recognizer; W2–W4 grow the
sub-language set. Option E accommodates both without re-architecting —
which is the point.

If a future workshop produces a proposal that genuinely needs recursive
grammar power in the envelope itself (rather than in a sub-language),
**that's** when we revisit E vs B. Until then, the simpler layering wins.

### 2.5 Debate openers

- Is "two sources of truth" worse than "one, more complex, source"? E
  bets no; B bets yes.
- If we go E, what format for the keyword table? TOML, YAML, custom
  DSL? Data favors YAML (nested alias lists); predictability favors TOML.
- Do we generate docs from the keyword table, or author them separately?
  Generate = single source of truth for keyword docs. Author = richer prose.
- What about `swagger:operation`'s embedded YAML? It's not part of the
  envelope or the keyword table — it's a separate sub-language. §3.

> NOTES(fred):
>
> * I am fine with delegating sub-languages to dedicated parsers.
> * This makes the main grammar very simple and yes, we may adopt something simpler than EBNF
>   (although we may _generate_ an EBNF for documentation).
> * So we have the main language parsing hand-rolled with a simple lookup table.
>   This structure should however capture the semantics well (keyword arguments)
> * We need an external definition, but we don't require it to be YAML or TOML: a map in go source
>   is suitable. We may hook a doc generation program on top of that table just the same.
> * We need to analyze further (workshop) how "enum" come as their own sub-language (kind of at least, perhaps I am overstating this)
> * We must find a way to express the syntax with "non-annotation keywords" such as "in:" and validations ("maximum:", ...).
>   This is part of the annotated language and currently only captured by how the different Taggers are crafted.

---

## 3. Parser / analyzer split — where does the line go?

### 3.1 The principle

**Parser concern:** syntactic structure. What shape is this comment?
**Analyzer concern:** semantic meaning. What does this shape mean for the
spec?

Parser output is deterministic from input text alone. Analyzer output
depends on context (the Go type being described, the enclosing operation,
the spec version being rendered).

### 3.2 Case-by-case

| Element | Parser does | Analyzer does |
|---------|-------------|---------------|
| Annotation kind (`swagger:model`) | Recognize, emit enum | Decide: "this struct becomes a definition" |
| Annotation args (`GET /pets tags listPets`) | Tokenize positional args; route has 4, operation has 4, model has 1 (optional) | Validate method is a valid HTTP verb; validate path syntax |
| `keyword: value` | Look up keyword, tokenize value-as-string, type-convert *primitives* (number, bool) | Apply value to target: "this is minimum on this field" |
| `items.N.x` | Extract depth; emit property with `ItemsDepth: N` | Check depth matches array nesting in Go type |
| Multi-line block body (`consumes:`, `produces:`) | Collect body lines as `[]string`, strip comment prefix | Parse each line as a MIME type; validate |
| `enum: a,b,c,"hello world"` | Tokenize comma-separated list, respect quoted strings | Type-check against field's Go type |
| `default: <value>` | Capture raw value string | Type-convert based on field type (analyzer knows the Go type) |
| `example: <value>` | Same as default | Same as default |
| `pattern: \w+` | Capture value string verbatim (regex is opaque) | Optionally validate regex compiles |
| Extensions block (`extensions: ... x-foo: bar`) | Isolate block; per-line capture raw `x-name: value`; detect inline `---` YAML | Parse inline YAML if present; emit `ir.Extensions{}` |
| YAML body in `swagger:operation` | Isolate `---` fence, capture body as raw text + position | Parse YAML via library; map to operation structure |
| Title vs description | Split by first blank line | — |
| Comment prefix stripping | Single pass in preprocessor | — |

> NOTES(fred):
>
> `swagger:meta` is also a big YAML provider.
> `enum:` is if I remember correctly more complex than that and there are several spots dedicated to this processing
>  peppered in different places (in IndexType, in parsers).

### 3.2.1 Enum: a cross-cutting case

Enum handling is the most entangled annotation today — touch-points span
`parsers/enum.go`, `TypeIndex` classification (enum-type discovery from
Go constants), and schema-builder code that stitches values back to their
declaring type. Values themselves come in at least three surface forms
(comma-list on the `enum:` keyword, YAML list under a block header,
godoc-style bullet list on a type declaration). Value types are
field-dependent (enums may be any JSON primitive), not keyword-dependent,
so they don't fit the §3.4 table-driven model cleanly.

This document does **not** resolve that — W2 is the workshop that will.
What it commits to is:

- The parser recognizes the three `enum:` surface forms and exposes them
  as distinct AST shapes (comma-list vs fenced block vs linked constants),
  preserving positions. It does not unify them.
- Values are captured as raw strings at the parser layer. Type-conversion
  happens in the analyzer using the field's Go type (same carve-out as
  `default:` and `example:` in §3.4).
- Scanner-side signal — "this Go const group looks like an enum of type
  T" — stays a scanner concern (§1.3), not a parser concern. The parser
  handles the comment-side; the scanner handles the code-side; the
  analyzer correlates them.

W2 may later collapse or extend these shapes; the parser's job is to
expose what the comment said, not to pre-decide the unified form.

### 3.3 The YAML question

Today, `swagger:operation`'s body is full YAML embedded in a godoc comment,
and `swagger:meta` relies heavily on YAML-ish structure for nested info
fields (license, contact, security schemes). Per vision.md, this was a
v1 mistake — godoc renders it as noise. v2 wants two styles: godoc-visible
(simple k:v) and private (`/* */`, richer).

**Main (grammar) parser's job:** recognize the `---` fence, capture the
body as a raw string with its position, hand it off. Do NOT parse YAML
syntax in the grammar parser.

**Analyzer's job (decision):** decide *when* to parse a `RawYAML`. If
the body came from `swagger:operation`, invoke the YAML sub-parser and
map to operation fields. If it came from a private comment with custom
structure, invoke the relevant sub-parser. The analyzer owns the
*when*, not the *how*.

**Sub-parser's job (implementation):** a thin wrapper around
`go.yaml.in/yaml/v3` lives in its own sibling subpackage at
`internal/parsers/yaml/`, separate from the main grammar parser. The
grammar parser never imports it; the analyzer does.

**The pattern generalizes.** Any future sub-language
(enum/example per W2/W3, richer extensions, private-comment bodies per
W4, …) gets its own `internal/parsers/<name>/` sibling subpackage.
This keeps the main grammar parser stdlib-only and its namespace
free of sub-language types.

> NOTES(fred):
>
> * For YAML parsing, I found that https://github.com/goccy/go-yaml might be interesting than
>   gopkg.yaml because it exposes a lexer to iterate YAML tokens. Just something to keep in mind for now.

### 3.4 The value-typing question

`default: 5` — is `5` the integer `5` or the string `"5"`? Today, `Set*.Parse`
methods decide based on the property (`SetMaximum` parses as float,
`SetEnum` parses as comma-list of strings). The keyword determines the type.

**Proposal:** the parser's keyword table declares a value type per keyword:

```yaml
maximum:     {type: number,  operator: [<, =]}
minimum:     {type: number,  operator: [>, =]}
multipleOf:  {type: number}
minLength:   {type: integer}
pattern:     {type: string-verbatim}
enum:        {type: comma-list}
required:    {type: boolean}
collectionFormat: {type: string-enum, values: [csv, ssv, tsv, pipes, multi]}
in:          {type: string-enum, values: [query, path, header, body, formData, cookie]}
...
```

Parser uses the table to tokenize values at parse time. Analyzer doesn't
re-parse — it reads `property.AsFloat()` or `property.AsList()` etc.

> NOTES(fred):
>
> I understand from this part that the "grammatical" aspects of when we may have a "in:" or when we have a "maximum:"
> are not handled by the parser, but by the analyzer (our Builder types).
>
> How is this different from the current "Tagger" behavior?

For `default`/`example`, where the type depends on the field being
described (not on the keyword), the parser emits `RawValue: string` and the
analyzer type-converts using the field's Go type. A small number of
exceptions, not a violation of the model.

> NOTES(fred):
>
> enums may also be of any type.

### 3.5 How this differs from today's taggers

Fred's §3.4 note: *"How is this different from the current Tagger
behavior?"* The proposal re-arranges the same responsibilities; it does
not just rename them. Today's taggers conflate five concerns in one
place:

| Concern | Today (tagger) | Proposed split |
|---------|----------------|----------------|
| **Recognition** — "is this line a `maximum:` annotation?" | Regex match inside the tagger's `Matches/Parse` | Parser, driven by keyword table |
| **Type-conversion of primitives** — "parse `5.5` as float" | `SetMaximum.Parse` on the builder target | Parser, driven by per-keyword `type` in the table (§3.4) |
| **Context-validity** — "is `maximum:` legal under this annotation kind?" | Implicit: only the tagger for the right target is constructed | Data: context table keyed by annotation kind (§2.2.1); LSP queries the same table |
| **Target-writing** — "call `SetMaximum(x)` on this schema" | The tagger itself writes to target | Analyzer (builder), now a visitor over AST nodes |
| **Semantic validation** — "does `maximum: 5` conflict with `type: string`?" | Partly in taggers, partly in builders, partly absent | Analyzer — separate pass over the AST; diagnostic rather than fatal |

Three consequences:

- **Single parse, many readers.** A comment block is parsed once into a
  typed AST. Multiple analyzers (spec builder, LSP, `codescan explain`,
  migration tool) read it without re-scanning text. Today each tagger
  re-recognizes its own keywords.
- **Recognition is context-free.** Today the set of recognizable lines
  depends on which tagger tree was constructed. Tomorrow, every
  recognized keyword is always recognized; whether it's *legal* in
  context is a separate table lookup. This is the LSP shape.
- **Errors accumulate.** Today an unrecognized line is either silently
  dropped or fails the build depending on the path. Tomorrow, diagnostics
  are collected and the AST is always produced; the analyzer decides
  policy.

The remaining overlap with taggers is real: the analyzer that walks an
AST and writes into `spec.*` is functionally what a tagger does today,
minus recognition and type-conversion. That's the "visitor" reference
in §4.2. Calling it a tagger or a visitor matters less than recognizing
that *its job has shrunk*: it no longer lexes, no longer type-converts
primitives, no longer matches regexes, no longer guards context-validity.
It picks AST nodes it cares about and maps them to its target.

### 3.6 Debate openers

- Is the table-driven value-typing worth it, or should the analyzer always
  type-convert from raw string? Table: uniformity + early errors. Raw:
  simpler parser.

**Whenever typing is possible, we should get early warnings (think: LSP diagnostic). So table-driven value-typing is good**

- Should route/operation positional args be parsed into semantic parts
  (method, path, tags, id) by the parser, or left as `[]string` for the
  analyzer? Positional = parser. Semantic = analyzer. I'd say
  parser, since the positional count is fixed.

**I am inclined to think parser, otherwise the parser would boil down to just a lexer, which it currently is**

- Should the parser validate extension names (`x-*`) are well-formed? It
  can — the regex for extensions is `^[Xx]-`. Cheap to do there.

**Yes: think LSP diagnostic - errors are not fatal anyway**

---

## 4. Interface and architecture

### 4.1 Stage diagram

```
┌──────────────────────────┐
│  *ast.CommentGroup       │
│  (or raw text for LSP)   │
└────────────┬─────────────┘
             │
┌────────────▼─────────────┐
│  Preprocessor            │  strip // /* */ *  table-pipe; preserve indentation
│  (stdlib strings)        │  output: []Line{Text, Pos}
└────────────┬─────────────┘
             │
┌────────────▼─────────────┐
│  Lexer                   │  line → token stream
│  - keyword table lookup  │  tokens: ANNOTATION, KEYWORD_VALUE,
│  - items. prefix         │          KEYWORD_BLOCK, YAML_FENCE,
│  - style plugins         │          TEXT, BLANK, EOF
└────────────┬─────────────┘
             │
┌────────────▼─────────────┐
│  Parser                  │  token stream → CommentBlock AST
│  - recursive descent     │  error recovery: skip to next line
│  - envelope grammar      │
│  - diagnostic collector  │
└────────────┬─────────────┘
             │
┌────────────▼─────────────┐
│  CommentBlock AST        │  typed nodes with positions
└────────────┬─────────────┘
             │
         (analyzer / builder reads this)
```

### 4.2 Visitor debate

User framed analyzers as "visitors replacing taggers". Options:

**V1 — AST as plain data, analyzers read.** `block.Properties`, `block.Title`,
`block.Annotation.Kind`. Analyzer picks what it needs.

**V2 — Typed accessors on the AST.** `block.ApplyValidations(target)`,
`block.GetFloat(KwMaximum)`, `block.Description()`. Encapsulates common
patterns.

**V3 — Iterator-based traversal.** `for p := range block.Properties()`,
`for yb := range block.YAMLBlocks()`, etc. Uses Go 1.22+ `iter.Seq` so
callers own the loop (natural `break`/`continue`/early-return,
composable with filters). Classical Accept/Visit callbacks are the
rejected alternative — our AST is shallow enough that visitor
compile-time exhaustiveness isn't a meaningful safety win, and iterator
composition reads cleaner at every call site.

> NOTES(fred):
>
> Question: would the the visitor pattern appear as an iterator (i.e. "range Nodes()") or as callback-like call (e.g. "Visit(caller)") ?
>
> **Settled: iterator form (`iter.Seq`).**

**V4 — Full IR pipeline.** Parser → transformer → IR; analyzer reads IR, not AST.
IR = §3 from step 3 (post-step-3).

**Proposed for step 1:** V1 + V2 + V3 from day 1. V3 ships early because
today's taggers become thin iterator-wrappers — see note below. V4 is
step 3, not step 1.

> NOTES(fred):
>
> I agree by and large with V1+V2.
>
> However, we could imagine that the current Taggers remain (until further transforms are needed) to
> precisely wrap the new visitor pattern (this would already reduce the size and complexity of the current Tagger). In this case V3
> may ship earlier.
>
> **Settled: V3 ships day 1. Taggers shrink to thin wrappers (~10–20
> lines) that iterate the AST and route to their existing target
> interface. No builder-side churn during migration.**

**Why this isn't just "taggers renamed":** today's taggers conflate
recognition (what line is this?) with action (write to target). The
proposal separates them: the AST captures recognition (parser's job);
the builder decides action (analyzer's job). In V3 form, a tagger's
remaining body is an iterator loop plus per-property dispatch; the
regex-matching, line-stripping, and type-conversion go away entirely.

### 4.3 Error model

**Proposal:** diagnostics are accumulated, not thrown.

```go
type Diagnostic struct {
    Pos      token.Position
    Severity Severity // Error, Warning, Hint
    Code     string   // "parse.invalid-number", "parse.unknown-keyword"
    Message  string
}

type CommentBlock struct {
    // ... AST fields ...
    Diagnostics []Diagnostic
}
```

Parser always returns a `CommentBlock` (possibly partial). Diagnostics are
collected. Analyzer decides policy (fail loud, warn, ignore).

This is the LSP-compatible shape from day 1.

### 4.4 Package layout

Proposed shape inside `internal/parsers/` (step 1 is internal):

```
internal/parsers/
├── grammar/                 ← main annotation parser (stdlib only, YAML-free)
│   ├── keywords.go          ← typed []Keyword slice, authored (source of truth)
│   ├── gen/                 ← docs generator (table → markdown)
│   │   └── main.go
│   ├── preprocess.go        ← comment prefix stripping
│   ├── lexer.go             ← tokenizer
│   ├── parser.go            ← recursive descent; Parser interface for test injection
│   ├── ast.go               ← Block family (typed) + sub-types
│   ├── diagnostic.go        ← error types
│   └── style.go             ← StyleRecognizer plugin interface
├── yaml/                    ← YAML sub-parser (thin wrapper on go.yaml.in/yaml/v3)
│   └── yaml.go              ← imported by analyzer only, when it decides to parse RawYAML
├── <future sub-parsers>/    ← e.g. enum/, example/ — same pattern, one per sub-language
├── grammar_test/            ← golden tests + fuzz
│   └── ...
├── sectioned_parser.go      ← legacy: kept during migration; removed at cutover
├── tag_parsers.go           ← legacy: kept during migration
├── regexprs.go              ← legacy: kept during migration
└── ...
```

Old parsing code stays operational during the migration. The grammar
subsystem is opt-in (builders switch to it incrementally, one annotation
kind at a time). **At cutover**, the legacy `internal/parsers/*.go`
files are deleted and `internal/parsers/grammar/*.go` is promoted up to
`internal/parsers/` (the `grammar/` subdirectory removed). Sibling
sub-parser subpackages — `internal/parsers/yaml/`, and any future
sub-language package — remain as children. The main parser gets the
clean namespace; sub-parsers are explicit imports only where the
analyzer decides to use them.

### 4.5 Migration path (high-level, not task-level)

1. Build lexer + parser + AST standalone, no builder integration.
2. Golden-parity test harness: for every fixture, run both old and new,
   diff the resulting spec. (Similar to Q-pass, but at parser output level
   instead of spec output level.)
3. Flip builders one-by-one to the new parser. `Options.UseGrammarParser`
   flag at first, maybe.
4. When all builders use new parser AND fixtures match AND coverage is
   restored, delete old parsing code.

**Key principle:** the public API does not change during step 1. `Run()`
still produces `*spec.Swagger`. Only the internals shift. Step 3 (IR +
renderer) is where the public shape evolves.

**Release cadence:** migration steps are recorded as distinct commits
(one per builder flipped, parity tests green at each) but ship as a
single public release — a merge commit with the whole lot migrated.
If parity fails late, we revert cleanly.

**Bridging strategy:** builder-side churn is minimized by doing the
AST-to-target wiring inside the (now-thin) tagger wrappers (see §4.2).
The builders themselves don't change in step 1; they keep consuming
tagger output.

> NOTES(fred):
>
> The progressive migration, with regression tests at each new builder migrated is good.
> The migration steps will be recorded as separate commits.
> However, we'll only ship once, with the whole lot migrated (with a merge commit).
>
> The migration strategy should minimize transformations to the builders in this phase, possibly
> constructing the needed bridging inside the Taggers.

### 4.6 Settled positions (previously debate openers)

- **Package position** — `internal/parsers/grammar/` during migration;
  renamed to `internal/parsers/` at cutover when the old code is deleted.
  (See §4.4.)

- **`Block` is a typed family, not a flat type.** Typed kinds
  (`ModelBlock`, `RouteBlock`, `OperationBlock`, `ParametersBlock`,
  `ResponseBlock`, `MetaBlock`, …) dispatched on the annotation line;
  a generic `UnboundBlock` for comments with no annotation (e.g., struct
  field docstrings carrying validations). A common `Block` interface
  carries `Pos`, `Title`, `Description`, `Diagnostics`.

  First-token dispatch is still context-free — the parser isn't reading
  external context, just selecting an output constructor from the
  annotation line it already recognizes.

  **Scanner contribution to typing (future).** Today the scanner only
  emits signals-of-interest for nodes *with* annotations (plus
  model-dependency discovery). As the scanner grows to signal
  annotation-less nodes (handler signatures, enum-shaped const groups,
  strfmt candidates — see §1.3), it can pass a kind hint to the parser
  so the emitted block is typed even without an annotation line. Out
  of scope for step 1; the `UnboundBlock` shape leaves room.

  **LSP aside:** a `ParseAs(kind, text)` entry point will be useful for
  completion scenarios where the annotation line is missing or being
  typed. Secondary API, doesn't affect the main path.

- **Token stream exposure** — internal only in v2.0. Easy to expose
  later when the LSP server is scaffolded.

---

## 5. Stack / tools / libraries

### 5.1 Runtime dependencies

**Required:**
- Go stdlib only for the grammar subsystem. `strings`, `strconv`, `unicode`,
  `go/token` (positions).

**Split: analyzer decides, sub-parser subpackage implements.**
- YAML parsing for `swagger:operation` body, `swagger:meta` body, and
  inline YAML extensions. The *decision* to invoke YAML parsing lives
  in the analyzer (or the bridging taggers during migration). The
  *implementation* lives in its own sibling subpackage
  `internal/parsers/yaml/`, separate from the main grammar parser.
  The grammar parser never imports it. See §3.3.
- **Library for v2.0:** `go.yaml.in/yaml/v3` (recently corrected from
  the deprecated `gopkg.in/yaml.v3`), imported only by
  `internal/parsers/yaml/`. Per-token YAML position tracking is not
  needed before LSP; when it is, `goccy/go-yaml` (exposes a lexer with
  token positions) is the POC candidate at that point.
- **Pattern generalizes** to any future sub-language — each gets its
  own sibling subpackage under `internal/parsers/` (§3.3).

**Explicitly not wanted:**
- `regexp` in the new parser. (Can remain in legacy code during migration.)
- Parser combinator runtimes (participle, etc.).

> NOTES(fred):
>
> Yes confirmed.

### 5.2 Build-time tooling

**Option E is adopted** (§2.4 tentatively; confirmed here via §2.5 and
§5.2 notes):

- **Hand-rolled annotation parser** producing a typed `Block` family
  (§4.6).
- **Lazy sub-language parsing.** The parser recognizes block identifiers
  (e.g., the `---` YAML fence; future enum/example shapes) and delegates
  to a dedicated sub-parser. YAML is the confirmed case; W2/W3/W4
  determine whether enum/example/extensions need their own sub-grammars.
  This generalizes what the current parser package already does for
  YAML, Meta, and Security blocks.
- **No parsing library.** No combinator runtime, no parser-generator
  runtime — not foreseeably.
- **Keyword table as a typed Go slice** (`keywords.go`, authored — see
  §5.5). A small internal generator (<200 lines,
  `internal/parsers/grammar/gen/main.go`) consumes the slice at
  `go generate ./...` time to emit `docs/annotation-keywords.md`. No
  code is generated; the table itself compiles.

Option B (EBNF + custom codegen) is explicitly not taken.

### 5.3 Test infrastructure

- **Fuzz tests** on the lexer and parser (`go test -fuzz=...`) — catches
  edge cases in line preprocessing and keyword recognition. Stdlib fuzzing
  since Go 1.18.
- **Golden fixtures** reusing the existing `scantest.CompareOrDumpJSON`
  harness — unchanged.
- **Parity tests** during migration: for each fixture, parse with old and
  new, assert the produced `Block` matches a normalized form.
- **Unit tests for each grammar production** — the envelope is small
  enough that each production gets a direct test, not just fuzzing.
- **Parser exposed as an interface to builders.** Builders take a
  `Parser` (or `BlockSource`) abstraction, not a concrete parser
  function. Tests inject synthesized `Block` values and drive the
  builder without running the parser. This is the unlock for the next
  item.
- **Property-based builder tests.** Building on the interface above,
  generate thousands of plausible `Block` shapes (validations,
  parameters, responses in varied combinations) and assert invariants
  on the produced `spec.*` / IR. Pattern already pioneered at
  `go-openapi/testify`; strictly superior to our ~dozens of
  hand-crafted integration scenarios for "plausible but hard-to-trigger"
  coverage.

### 5.4 Dev-side tools considered and rejected

- `golang.org/x/exp/ebnf` — parser for EBNF itself. Useful only if we
  go full codegen. Small, stable; acceptable if B wins.
- `participle/v2` — runtime struct-tag combinators. Rejected (§2).
- `pigeon` — PEG generator. Rejected (§2).
- `alecthomas/participle/v2/ebnf` — their EBNF-to-combinator adapter.
  Rejected (same reasons as participle).

> NOTES(fred):
>
> OK we don't need these.

### 5.5 Settled positions (previously debate openers)

- **Keyword table format** — `keywords.go` as a typed `[]Keyword{...}`
  slice, authored in Go. Editor support and compile-time validation
  outweigh "YAML is machine-readable". Docs are generated from the same
  slice (§5.2).
- **Standalone `codescan grammar check <file>` CLI** — deferred to the
  PLS/IDE integration stream. Not needed for step 1; the parser only
  needs to be ready to plug into that tool when it arrives.

---

## 6. Summary — what this document commits us to

**Positions taken (for the debate):**
- §1: scope v2.0 parser as "infrastructure for OAI 3.x + LSP, keyword set
  stays at parity with v1"; concrete OAI 3.x keywords land in v2.1.
- §2: option E (hybrid — hand-rolled envelope + data-driven keyword table).
- §3: main (grammar) parser handles structure, YAML-fence isolation,
  primitive value typing; analyzer decides *when* to parse
  sub-languages; sub-parser implementations live in sibling subpackages
  under `internal/parsers/` (`yaml/` first, more follow per W2/W3/W4)
  so they don't pollute the main parser's namespace.
- §3.5: taggers are not just renamed — parser owns recognition +
  primitive type-conversion, context-validity moves to a data table,
  target-writing and semantic validation stay in a shrunken analyzer.
- §3.2.1 / §3.3: enum is cross-cutting and defers surface unification to
  W2; `swagger:meta` joins `swagger:operation` as a YAML-heavy annotation
  the parser *isolates* but does not *parse*.
- §4: stages = preprocessor → lexer → parser → AST. AST = **typed
  `Block` family** (one per annotation kind + `UnboundBlock`) with
  helper methods; traversal via **iterator** (`range block.Properties()`),
  not Accept/Visit. Diagnostics accumulated. Parallel-run migration with
  taggers shrinking to thin iterator-wrappers (V3 ships day 1). Single
  release at cutover; `internal/parsers/grammar/` renamed to
  `internal/parsers/`.
- §5: stdlib-only main-parser runtime. **Option E fully adopted.**
  `keywords.go` authored as a typed slice (docs-only generator).
  Test infrastructure adds **parser-as-interface for builder mocking**
  + **property-based builder tests** on top of fuzz/golden/parity.
  YAML library `go.yaml.in/yaml/v3` imported only by
  `internal/parsers/yaml/` sub-parser subpackage.

**Deliberately unresolved (v2.x territory):**
- Incremental parsing (C15) — designed-for, not proven in v2.0.
- OAI 3.x keyword set — infrastructure in v2.0, keywords in v2.1.
- Pluggable styles — infrastructure in v2.0, `openapi:` + `@` in v2.1.

**Next plan document** (once these settle): task-level implementation plan
with milestones, comparable to how `archive/observed-quirks.md` drove the Q-pass.

---

## Appendix — pointers for context continuity

Session-independent pointers so a fresh conversation can pick up this plan
without rebuilding the surrounding context.

### A.1 Current-state snapshot (as of 2026-04-20)

- **Regexps today:** `internal/parsers/regexprs.go` — ~40 compiled regexps
  plus 13 `*Fmt` templates runtime-compiled per items-nesting depth.
  This is the surface being replaced.
- **Parser package:** `internal/parsers/` — 24 `.go` files. Largest:
  `validations.go` (14 KB), `sectioned_parser.go` (9 KB), `regexprs.go`
  (8 KB).
- **Builder seam:** `ifaces.SwaggerTypable` + `ValidationBuilder` are the
  existing interfaces parsers write through. Memory notes these as
  v2-sunset (`project_taggers_to_sunset.md`).
- **Q-pass outcome:** 8/13 quirks landed (Q1–Q6, Q9, Q10); 5 deferred to
  v2 in `.claude/plans/deferred-quirks.md` (D1–D6, mostly alias-theme).
  Post-Q-pass master is the clean baseline this plan builds on.

### A.2 External references used in this plan

- **Annotation reference docs:** `/home/fred/src/github.com/go-swagger/go-swagger/docs/reference/annotations/`
  (`_index.md`, `model.md`, `operation.md`, `params.md`, `response.md`,
  `route.md`, `meta.md`, `allOf.md`, `strfmt.md`, `discriminated.md`,
  `type.md`, `ignore.md`). Historical, not always up to date with current
  codescan behavior.
- **Fixtures surveyed:** `fixtures/enhancements/*` (20 sub-dirs covering
  alias modes, allOf edges, defaults/examples, embedded types, enum docs,
  interface methods, named basics, pointers, ref-alias chains, strfmt
  arrays, swagger-type arrays, text-marshal, top-level kinds, malformed,
  etc.); `fixtures/goparsing/petstore/`.

### A.3 Positions taken but not yet pushed back on

All major positions are settled as of 2026-04-21 (see header and A.4).
Items that were added in the final review pass and may warrant a
fresh-eyes read later:

- §3.5 tagger-distinction write-up (added 2026-04-21).
- §3.2.1 enum-as-cross-cutting and §3.3 `swagger:meta` YAML note
  (added 2026-04-21).
- §4.6 future scanner kind-hints contribution to parser typing (a
  forward-looking note, not a v2.0 commitment).

None of these are blockers for starting the task-level plan.

### A.4 Decisions Fred locked in during review

- **C15** (incremental parsing) → Day 1, essential complexity.
- **C9** (pluggable styles) → v2.x, natural evolution from keyword table.
- **C14** (OAI 3.x keywords) → v2.x; v2.0 keeps v1 keyword parity. OAI 3.x
  used only to size the parser's expressive power.
- **Macros** → philosophically out; `externalDocs` carve-out in scope.
- **Grammar formalism** — tentatively Option E (§2 pending full review);
  external grammar definition preferred over participle-style struct tags;
  pigeon maintenance status uncertain.
- **Signal/dispatcher model** (§1.3) — parser and scanner emit signals to
  a dispatcher (distinct architectural element, designed alongside step 2).
- **§3.6 debate openers** — table-driven value-typing = yes (early LSP
  diagnostics); route/operation positional args parsed into semantic
  parts by the parser (not left as `[]string`); extension names
  validated in the parser (non-fatal diagnostic).
- **§2.5 grammar definition format** — a Go `map` (typed) is the
  external definition of choice, not YAML/TOML. Doc-gen and an
  EBNF-for-docs ride on top. Sub-languages (YAML, enum variants,
  extensions) are delegated to dedicated parsers.
- **§4.2 traversal model** — iterator form (`iter.Seq`), not
  Accept/Visit callbacks. V1 + V2 + V3 all ship day 1; taggers become
  thin iterator-wrappers so builders don't churn during migration.
- **§4.4 package position** — `internal/parsers/grammar/` during
  migration; renamed to `internal/parsers/` at cutover (old code
  deleted).
- **§4.5 release cadence** — granular migration commits (one per
  builder flipped, parity-green at each); single merge commit / single
  release at cutover.
- **§4.6 typed `Block` family** — typed AST kinds per annotation line,
  `UnboundBlock` for annotation-less comments. First-token dispatch
  stays context-free. Scanner may contribute kind hints later for
  annotation-less signals-of-interest.
- **§5.1 YAML library + sub-parser location** —
  `go.yaml.in/yaml/v3` (recently corrected from the deprecated
  `gopkg.in/yaml.v3`), imported only by a dedicated sub-parser
  subpackage `internal/parsers/yaml/`. The main grammar parser stays
  YAML-free. The same pattern generalizes to future sub-languages
  (each gets its own `internal/parsers/<name>/` subpackage).
  `goccy/go-yaml` is a POC candidate only when LSP needs token-level
  YAML positions.
- **§5.2 Option E fully adopted** — hand-rolled annotation parser,
  lazy sub-language delegation (YAML confirmed; enum/example/extensions
  depend on workshops), no parsing library, docs-only internal
  generator.
- **§5.3 parser-as-interface** — builders take a `Parser` abstraction
  so tests inject synthesized `Block` values. Unlocks property-based
  builder tests (pattern already used at `go-openapi/testify`).
- **§5.5 `keywords.go`** — authored typed Go slice, not external
  YAML/TOML. `codescan grammar check` CLI deferred to PLS/IDE stream.

### A.5 What a fresh session needs to do

1. Read this file end-to-end. All five sections are settled (see
   header); A.4 is the condensed index of decisions.
2. The next document is the **task-level implementation plan** at
   `.claude/plans/grammar-parser-tasks.md`, with milestones comparable
   to how `archive/observed-quirks.md` drove the Q-pass.
3. Keep this file as the "why" reference. Re-litigate only if
   implementation surfaces a constraint the architecture didn't see.
