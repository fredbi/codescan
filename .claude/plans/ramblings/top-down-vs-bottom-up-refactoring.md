# Top-Down vs Bottom-Up Package Extraction

Date: 2026-03-25

## The Problem

Splitting a tightly-coupled single Go package (~3000 LOC, ~14 files) into ~12 internal
subpackages. The codebase has deep entanglement: parser types used by builders, scanner
types used by everything, interfaces implemented across package boundaries, unexported
fields accessed via positional struct literals.

## Failed Attempts (5 total: 2-3 human, 2 Claude)

### Bottom-up approach (Claude's attempts)

**Strategy**: Move leaf packages first (ifaces → parsers → scanner → builders → spec),
creating type aliases in the root package as a bridge. Each step should be independently
committable and green.

**What happened**: Each leaf move was "clean" in isolation but created a cascade of
`undefined:` errors in the consuming packages. Fixing those errors revealed more
unexported symbols, which revealed more consumers, which revealed more symbols. The
error count didn't decrease — it shifted. After 30+ minutes of incremental sed/gopls
fixes, the change was reverted.

**Why it fails**:
- You're **blind to actual consumer needs** until you try to compile them. Moving
  parsers first means you don't know which of the 100+ symbols in parsers are actually
  referenced from builders until you try to build the builders.
- The **bridge file grows unboundedly**. Every symbol that crosses the old→new boundary
  needs an alias or var bridge. For parsers alone, this was 100+ declarations.
- **No restructuring insight**. Bottom-up preserves the existing structure — you're just
  shuffling files. You never ask "does this consumer really need this symbol, or should
  the code be organized differently?"
- **Preparation is speculative**. We did extensive preparation (export fields, convert
  positional literals) based on what we *thought* would be needed. Some of it was
  unnecessary; some of it was insufficient.

### Top-down approach (human's approach)

**Strategy**: Start from the public API (`api.go`) and work inward. Move `api.go` first
(it only needs `scanner.NewScanCtx` and `spec.NewBuilder`). Then move `spec` builder
(it needs scanner + all builders). Then move each builder, driven by what the spec builder
actually imports.

**What happened**: At each step, the question is "what does THIS consumer need?" — not
"what should I export from this producer?". This revealed:
- Some symbols didn't need exporting at all — they could be restructured
- New packages emerged naturally (`internal/builders/resolvers`, `internal/builders/items`,
  `internal/logger`) that weren't in the original plan
- The regex variables and parser constructors could be wrapped in higher-level functions
  rather than exported individually

**Why it works better**:
- **Demand-driven**. You only deal with symbols that are actually needed by the current
  consumer. No speculative bulk exports.
- **Restructuring is natural**. When you see "schema builder needs SwaggerSchemaForType
  and IsAliasParam from parsers", you ask: "do these belong in parsers?" The answer is
  no — they're type resolution logic. So `internal/builders/resolvers` is born.
- **Each step teaches you about the next**. Moving the spec builder reveals what the
  schema builder's public API should look like. Moving the schema builder reveals what
  parsers should export.
- **Errors are localized**. At any point, only the package you're currently moving has
  errors. Everything above it (already moved) compiles. Everything below it (not yet
  moved) is untouched.

## The Tradeoff

| Aspect | Bottom-up | Top-down |
|--------|-----------|----------|
| Each step compiles? | Yes (leaf packages are self-contained) | No (current package has errors until finished) |
| Error visibility | Deferred — errors appear when consumers are touched | Immediate — you see exactly what's missing |
| Restructuring insight | None — preserves existing structure | High — reveals better abstractions |
| Mechanical? | Mostly (copy, rename, alias) | No — requires judgment at each step |
| Bridge/alias burden | Heavy (100+ bridges for parsers alone) | Light (only what consumers actually need) |
| Risk of cascading mess | High | Low |

## The Meta-Lesson

**When a refactoring fails multiple times with the same approach, the approach is wrong —
not the execution.**

Both Claude and the human failed with bottom-up. The failures weren't due to bugs in sed,
wrong gopls positions, or missing preparation steps. They were due to a fundamental
mismatch between the approach (supply-driven: "what should I export?") and the problem
(demand-driven: "what does each consumer need?").

The signal was there from the start: the dependency analysis showed that parsers has 100+
symbols but builders only use ~30 of them. Bottom-up forces you to handle all 100+.
Top-down lets you handle only the 30 that matter.

## When to Use Each Approach

**Bottom-up works when**:
- Packages are loosely coupled (few cross-boundary references)
- The existing structure already reflects the target structure
- The move is mostly mechanical (no restructuring needed)

**Top-down works when**:
- Packages are tightly coupled (many cross-boundary references)
- The target structure is uncertain and may evolve during the refactoring
- You suspect the current code organization hides better abstractions
- Previous attempts have failed — you need fresh insight into the dependencies

## Applied to This Codebase

The top-down approach revealed packages we didn't anticipate:
- `internal/logger` — trivial, but isolates the debug logging concern
- `internal/builders/resolvers` — SwaggerSchemaForType, IsAliasParam, and type assertion
  helpers. These were in parser.go but are really about resolving Go types to Swagger
  schemas. They depend on ifaces but nothing else.
- `internal/builders/items` — items typable/validations/taggers. These were mixed into
  parameters.go but are used by both parameter AND response builders. Extracting them
  avoids a parameter→response dependency.

None of these emerged from the dependency analysis or the bottom-up attempts. They emerged
from asking "what does schema builder actually need?" and finding that the answer wasn't
"all of parsers" but "a few specific capabilities that deserve their own home."
