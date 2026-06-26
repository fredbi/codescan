---
title: Example value coercion (verification)
stream: —
origin: i
status: open
release: v0.37
issues: [go-swagger#1268, go-swagger#2246]
prev: "§2"
---

# Example value coercion (verification)

**Status:** ⬜ open · **low priority — likely subsumed**.

**Origin.** W3 deferred 2026-04-21 — richer `example:` / `examples:` support.
Groomed 2026-06-23: this mostly **dissolves into a verification task**, it is not
a distinct feature.

**What's actually left.**

- **`example:` (singular) coercion** — go-swagger#1268 (an `example:` on a
  `[]string` kept as the raw string, not coerced to an array) and go-swagger#2246
  (a JSON-object `example:` kept as an escaped string, not parsed) are
  **allegedly already fixed**. **Verify** with fixtures/tests; close if confirmed.
- **`examples:` (plural)** — in **OAS2** this is specific to **responses** and is
  covered by [response-examples-by-mime](response-examples-by-mime.md). In **OAS3**
  (name-keyed map) it is **deferred to OAI v3** (Stream 10). Nothing to decide here.

So the standalone feature is a **candidate for retirement** once the #1268/#2246
verifications confirm the singular path is coerced correctly.

**When to revisit.** After the verification; otherwise retire.
