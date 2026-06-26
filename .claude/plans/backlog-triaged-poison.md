# go-swagger backlog — POISON queue (empty)

_Triaged issues **parked** during the first pass: under-specified / can't build a
repro yet / environmental. The queue is now **empty** — all parked issues have
been qualified and routed to a category ledger._

**2026-06-17 — queue emptied.** The three remaining parked issues (#2633, #2778,
#2874) were a single "docker/env silent-stop → empty/incomplete spec" family.
They are resolved by the fail-loud diagnostics feature §8.2 (`detectDegradedLoad`
after `packages.Load`, commit `d2479a4`), which converts a degraded package load
into a located `scan.degraded-load` diagnostic + an aborting `ErrDegradedLoad`
instead of a silent empty spec. All three moved to
[`backlog-triaged-ledger.md`](backlog-triaged-ledger.md) as ✅ (#2874 ✅♻️, likely
dup of #2633). 📖 doc flags carried over for the empty-spec / degraded-load
troubleshooting pages.

Earlier poison residents were qualified out on 2026-06-17 (see
[`project_backlog_tracker_split` memory] and the ledger / feature files): #1560,
#1670, #2838, #2963, #3134 → fixed; #2924 → feature §20 (since landed); #1712,
#2777 → removed (relabeled out on GitHub).

_No open 🐞 rows remain._
