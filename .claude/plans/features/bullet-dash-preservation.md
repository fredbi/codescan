---
title: Bullet-list dash preservation — analyzer-side tolerance
stream: —
origin: iv
status: open
release: unscheduled
issues: []
prev: "§4.2"
---

# Bullet-list dash preservation — analyzer-side tolerance

**Status:** ⬜ open · no consumer yet.

**Origin.** P1.10 decision to keep leading `-` in Text (a v2 divergence from v1's
silent strip). A deferred refinement with no dedicated stream; it rides whichever
builder migration first turns up the relevant fixture diffs.

**Scope.** The Block's Title/Description now contains `- foo - bar` literally. If
any downstream renderer trims leading `-` for cosmetic reasons (v1 did so
silently), we keep parity by default at the parser layer but may need an opt-in
`StripBulletDashes` option on the bridge-tagger. No known consumer yet.

**When to revisit.** First time a builder migration turns up fixture diffs
involving bullet-list description formatting.
