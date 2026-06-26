---
title: Property-based Block generator
stream: 2
origin: i
status: open
release: unscheduled
issues: []
prev: "§5.3"
---

# Property-based `Block` generator

**Status:** ⬜ open.

**Origin.** Architecture §5.3 (go-openapi/testify pattern) / task P7.4. Tracked
under Stream 2 (grammar/parser).

**Scope.** A generator for synthetic `Block` values used in builder tests.
Produces thousands of plausible shapes; an assertion runs over the spec output.

**When to revisit.** P7 hardening — after cutover. Tracked as P7.4 in the tasks
plan; this entry exists so it's visible here alongside its siblings.
