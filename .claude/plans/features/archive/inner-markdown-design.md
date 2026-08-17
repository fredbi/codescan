---
title: Inner markdown — design & build log
stream: 8
origin: iii
status: design
release: v0.36
issues: [go-swagger#3211]
---

# Inner markdown — `swagger:description |` verbatim block scalar

Overview / status: [inner-markdown.md](inner-markdown.md).

## 1. Problem

go-swagger#3211 wants markdown that survives into a `description`: tables
(leading `|`), nested-list indentation, and significant blank lines. Two
distinct failure modes exist today, on two different code paths:

- **Preamble path** (a decl's godoc → `description`, via `classifyProse` /
  `PreambleDescription`). Blank lines already survive, but `trimContentPrefix`
  (`grammar/preprocess.go:161`, `TrimPrefix(s, "|")`) strips the leading pipe
  of markdown table rows → `| a | b |` becomes `a | b |`. This is the literal
  #3211 repro (`fixtures/bugs/3211`, on branch `feat/go-swagger-3211`).
- **Annotation path** (`swagger:description` override → `collectDescriptionBody`,
  `grammar/lexer.go:693`, "Option B"). Two limits:
  - `strings.TrimSpace(in[j].Text)` discards indentation, and `.Text` already
    lost the pipe and normalised bullets;
  - the body loop stops at the first `TokenBlank`, so a blank line truncates
    the description.

We do **not** try to recover markdown from uncontrolled godoc prose (the
wont-fix-as-filed stance holds). We make markdown work where the author asks
for it **explicitly**: the `swagger:description` override.

## 2. Decision — `swagger:description |`

A YAML literal block-scalar marker on the annotation line switches the body
from blank-terminated (Option B) to **verbatim literal block**:

- trigger: a lone `|` (optionally `|-` / `|+`, see §chomping) as the
  annotation's inline argument;
- capture: every following line **verbatim** (`Line.Raw` — indentation, pipes,
  bullets, blank lines all preserved);
- terminator: the next `swagger:*` annotation, a sibling structural keyword, or
  end of comment (EOF). No closing delimiter.
- default unchanged: without `|`, `swagger:description` stays exactly as today.

### Why this is builder-agnostic

The marker is handled at the **lexer** level, inside `collectDescriptionBody`,
which produces the annotation's single `Args` token (`AnnotationArg()`). Every
consumer of `swagger:description` — schema model title/description, field
description, and any future site — reads that arg, so they all get the verbatim
body with no per-builder change.

## 3. Mechanism

The capture machinery already exists; this is mostly routing.

- `Line.Raw` (`preprocess.go:27`) keeps every char after the comment marker —
  indentation, the table pipe, un-normalised bullets. Every `tokenText` token
  already carries it (`lexer.go:99`).
- `collectRawBlock` (`lexer.go:772`, used by `examples` / `extensions` /
  `consumes` / …) already accumulates a body that (a) does **not** terminate on
  blank lines (it buffers `pendingBlanks` and flushes them), and (b) for
  YAML-bodied keywords reads `.Raw` (`TrimRight` only) so indentation survives.
  It terminates at the next sibling structural item or EOF.

### Why this must intervene at lex **stage 1** (empirically confirmed)

The first instinct — route to a raw accumulator in `collectDescriptionBody`
(stage 2, `accumulateBodies`) — does **not** work, because the global `---`
fence toggling happens earlier, in `classifyLines` (stage 1). A probe confirmed:
for a `|` body containing a lone `---` followed by another annotation, e.g.

```
// swagger:description |
// Overview
// ---
// after the dash
// swagger:model Foo
```

stage 1 turns the `---` into a `tokenYAMLFence`, flips `inFence`, and — because
`lexLine`'s `inFence` short-circuit (lexer.go:69) runs *before* the annotation
check — re-classifies the following `swagger:model Foo` as a `tokenRawLine`. It
then becomes one `OPAQUE_YAML` token that **swallows the model annotation
entirely**. An odd `---` in the body destroys following annotations. This is the
very collision that ruled out a `---` fence (§rejected); it bites the `|` design
too, but only at the lexer layer, and only if handled too late.

### Resolution — a stage-1 "literal description" mode

`classifyLines` gains an `inLiteralDesc` flag alongside `inFence`. When it emits
a `swagger:description` annotation whose sole inline arg is `|`
(`isLiteralDescMarker`), it enters literal mode. While in literal mode, each
subsequent line is classified with `lexLine(line, false)`:

- if it lexes as a `TokenAnnotation` → **terminator**: leave literal mode, emit
  it normally (so a following `swagger:model` survives);
- otherwise → emit a `tokenRawLine` carrying `Line.Raw` **verbatim** — no fence
  toggling, no blank-line termination, no keyword sensitivity. `---`, code
  fences, blank lines, and keyword-looking lines are all body.

Then stage 2 dispatches the `|`-marked description to `collectDescriptionLiteral`
(instead of `collectDescriptionBody`), which folds the contiguous `tokenRawLine`
run into the annotation's single `TokenRawValue` arg (join `"\n"`, drop trailing
blank lines = bare-`|` clip), **dropping the `|` marker itself** so it never
reaches the description text.

```go
// classifyLines (stage 1)
if inLiteralDesc {
    tok := lexLine(line, false)
    if tok.Kind != TokenAnnotation {        // body: verbatim, no fence/blank/keyword logic
        out = append(out, rawLine(line)); continue
    }
    inLiteralDesc = false                    // annotation terminates the block
}
... normal lexLine(line, inFence) ...
if !inFence && isLiteralDescMarker(tok) { inLiteralDesc = true }

// accumulateBodies (stage 2)
case TokenAnnotation:
    if t.Name == labelDescription {
        if isLiteralDescMarker(t) { i = collectDescriptionLiteral(in, i, &out) }
        else                      { i = collectDescriptionBody(in, i, &out) }   // Option B, unchanged
    }
```

Terminator is therefore **next annotation or EOF only** — no keyword check (per
review decision 3: keyword detection is indentation-sensitive and unsafe inside
freeform markdown). Plain `swagger:description` (no `|`) is completely untouched.

Constraint (documented): a body line that itself lexes as a `swagger:*`
annotation ends the block — don't start a markdown body line with `swagger:`.

## 4. Scope

- Applies to `swagger:description` wherever it is consumed. Today that is the
  **schema family** (model description + field description) per the
  description-override feature; the lexer-level capture means it extends to any
  site automatically.
- `swagger:title` stays single-line (titles are not block content).
- Pairs with **AfterDeclComments**: the intended idiom is the override authored
  as an inner / private comment so the godoc surface stays clean, but the
  marker works in a regular doc comment too.

## 5. Chomping — DECIDED: bare `|` only

Support bare `|` only; always clip (drop trailing blank lines, keep interior
blanks). No `|-` / `|+`. Keeps the surface minimal; easy to add later if asked.

## 6. Terminator — DECIDED: next annotation, otherwise EOF

- Terminators: the next `swagger:*` annotation, or EOF. EOF is the graceful,
  expected end (the inner-comment idiom ends at EOF).
- **No keyword check.** A keyword-looking line inside the markdown (e.g. prose
  starting `Schemes:`) is body, not a terminator — keyword detection is
  indentation-sensitive and would mis-fire on freeform markdown. This is the
  key divergence from `collectRawBlock` (which is sibling-keyword-aware).
- No diagnostics in scope for now (bare `|`, capture-or-empty). An empty-block
  Hint can be added later if it proves useful.

## 7. #3211 disposition

- Close #3211 as **reframed / by-design**: the supported path is an authored
  `swagger:description |` block, not markdown recovery from godoc.
- The `feat/go-swagger-3211` branch's `fixtures/bugs/3211` repro (preamble path,
  pipe-strip) is *not* the fix target. Either:
  - (a) leave the preamble path as-is (pipe-strip stays; document the
    recommended override path), or
  - (b) also stop stripping the leading pipe on the preamble path.
  **DECIDED: (a)** — leave the preamble path as-is. Markdown belongs in an
  explicit `swagger:description |` override, not ambient godoc.

## 8. Phasing (one DCO-signed commit per phase, pause for review)

- **P1 — lexer capture.** ✅ DONE (`cf6d918`). `isLiteralDescMarker` + stage-1
  `inLiteralDesc` mode + `collectDescriptionLiteral`; 7 lexer unit tests
  (indentation, blank lines, table pipes, the `---`-keeps-following-annotation
  regression, keyword-line-is-body, trailing-blank clip, empty body, plain
  unchanged). Default Option B path untouched.
- **P2 — DISSOLVED.** Chomping decided bare-`|`-only and no diagnostics
  (review §5/§6), so there is no separate P2; folded into P1.
- **P3 — end-to-end + fixture + #3211.** ✅ DONE (`7765d44`).
  `fixtures/enhancements/inner-markdown/` (model + field markdown override:
  table, nested ordered list, blank-line paragraph) + integration golden prove
  the path. Added the single-godoc-convention-space strip (found during
  integration — `// text` decoration is not body). go-swagger#3211 closed as
  reframed. Old preamble-path repro left on branch `feat/go-swagger-3211` (not
  ported — preamble path unchanged per §7).
- **P4 — docs.** ✅ DONE (`c199fb3`). `grammar/README.md` §literal-description
  (marker, two-stage capture, convention-space strip, terminator contract) +
  root `CLAUDE.md` mention + two terminator tests. Doc-site how-to deferred to
  the doc-site cadence.

## 9. Review decisions (locked 2026-06-26)

1. **Chomping:** bare `|` only, always clip. (§5)
2. **#3211 preamble pipe-strip:** leave as-is — markdown is an explicit
   override only. (§7)
3. **Terminator:** next `swagger:*` annotation, otherwise EOF. **No** keyword
   check (keyword detection is indentation-sensitive and unsafe in freeform
   markdown). (§6, §mechanism)
4. **Naming:** keep `swagger:description |`. No new annotation surface.
