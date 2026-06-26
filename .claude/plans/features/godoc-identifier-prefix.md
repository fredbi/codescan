---
title: Godoc-identifier prefix on swagger:operation
stream: 10
origin: i
status: open
release: v0.39
issues: []
prev: "§3.6"
---

# Godoc-identifier prefix on `swagger:operation`

**Status:** ⬜ open · the leading-identifier step is small and standalone; full
operation≈route convergence tracks C9 (Stream 10).

**Origin.** P1.6 decision — only `swagger:route` accepts a leading godoc
identifier (`DoFoo swagger:route …`); every other annotation must start the
comment line. So far that is the **only** exception to "ignore annotations buried
in prose".

**Decision (groomed 2026-06-23).** The general objective is to bring
`swagger:operation` **progressively closer to `swagger:route`** — the eventual
goal is to make them **perfect synonyms**. As the first step:

- **Extend the leading-identifier exception to `swagger:operation`**, so
  `DoFoo swagger:operation …` works like `DoFoo swagger:route …`.

**Documentation note.** The `{ident} {annotation}` form is a **historical
exception, not the norm** — document it as **discouraged but supported**, and
steer authors toward the common form:

```
{godoc title}

{annotation}
```

**Interactions.** Relates to the godoc-subset entry point of
[comment-source-filtering](comment-source-filtering.md) and the C9
pluggable-styles / `openapi:` prefix work.

**When to revisit.** The leading-identifier step can land anytime; the broader
synonymy with `swagger:route` tracks C9.
