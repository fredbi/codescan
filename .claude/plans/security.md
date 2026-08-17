> [!NOTE]
> Last revision: 2026-08-17 (rev 3 — item 4 merged as PR #119; `hack/browser` retirement decided)

# Security posture and scanner-finding triage

## Summary

We are evaluating [Aikido](https://www.aikido.dev/) against this repo. Some of what it reports is real and worth fixing;
a good part of it is pattern-matching on sink shapes with no reachable source, escalated to "high" precisely *because*
the tool could not trace the origin. This plan records the triage rule we settled on, the disposition of every finding
seen so far, and the two genuine issues we found ourselves while checking Aikido's claims.

The durable deliverable is not a set of patches — it is a **written threat model** in `SECURITY.md`, so a dismissed
finding is closed by pointing at a sentence rather than re-argued from scratch on every scan. That is the main
outstanding action.

Standing constraint from Fred: **no complexity added to tick boxes on invalid warnings.**

## Context

Aikido runs on a free plan: no wholesale export, so findings arrive pasted one at a time. Seven arrived on 2026-08-14
(four front-end / supply-chain, three Go), and the pattern is now clear enough to predict the rest.

The repo's own gates already cover a lot of this ground: `.golangci.yml` is `default: all`, so **gosec runs and passes**
(verified: 0 issues on the package Aikido flagged), and `npm audit` is clean on both npm packages. What our gates do
*not* see is anything fetched at build time — Mermaid and the Relearn theme are curl'd, not npm dependencies — which is
exactly where the real findings were.

The threat-model question that decides every one of these: **codescan is a local developer tool running as the invoking
user.** argv, the terminal and the working directory are the user's own. But the tool's *purpose* is to read code that
may not be the user's — someone's API repo, a contributor's branch — so **content originating in a scanned repository
is untrusted**, and so are **assets fetched at build time and published to visitors**. Those two are our real boundaries.

## Trajectory

1. 😇 Triage rule — settled, applied to 7 findings in a row
   1. ✅ Name the boundary: who supplies the value, and do they differ from who runs the tool?
   2. ✅ Weigh what the attacker actually *gets*, not the shape of the sink
2. 😇 Write the threat model down (`SECURITY.md`) so dismissals are citable
3. 😇 Fix what is inside the boundary
   1. ✅ The doc-site supply chain (fetched assets published to visitors)
   2. ⚠️ Content from a scanned repository redirecting a write (found by us, not by Aikido) — **still Fred's
      undecided call, and now the ONLY security item left before v0.36.4**
   3. ✅ Control characters from scanned source reaching the terminal — merged, PR #119
4. 📚 Keep the local (non-CI) paths in step with the hardened CI ones
5. 📝 Retire dead surface rather than defend it — `hack/browser` **decided: delete** (2026-08-17)

## Actions

### Outstanding

1. ⚠️ 😇 **`.codescan.yaml` can redirect where `genspec` writes** — the one real Go finding, and Aikido missed it
   - `cliconf.search()` (`internal/cliconf/discover.go:114`) walks up to the **filesystem root**: no stop at the module
     root, the git root or `$HOME`. Running `genspec` inside a repo reads *that repo's* `.codescan.yaml`.
   - `document.output` is a configurable key (`cmd/genspec/configfile.go:33`) and `cmd/genspec/output.go:72` does an
     unconstrained `os.WriteFile`. Config loses to anything typed — but the default output is stdout, so a user who
     passes no `-output` has nothing for it to lose to.
   - Attack: clone a hostile repo, run `genspec ./...`, and the repo picks the path *and* substantially controls the
     content (descriptions come from its own doc comments). Deterministic arbitrary file write, any platform.
   - It matters more than it looks because the toolchain-free loader is sold as never exec'ing anything: users may
     reasonably believe scanning is inert.
   - **Recommended fix** (small, in the idiom already there): make `output` — and probably `input` — settable only from
     a config file the user *named* with `-config`, never from a *discovered* one. Two entries in the existing
     `notConfigurable` table with a reason string. No path validation, no new machinery, discovered config keeps
     working for every other key.
   - Rejected alternative: bounding the upward search at the module root. Reads like a fix, isn't one — the repo root's
     own `.codescan.yaml` is still attacker-controlled.
   - Also defensible: do nothing and document it, the way `go build` treats hostile modules. Forfeits something
     codescan otherwise has. **Fred's call, undecided.**

2. 📝 😇 📚 **Write `SECURITY.md`'s threat model** — 3–5 lines, roughly:
   > codescan is a local developer tool that runs as the invoking user. The command line, the terminal and the process
   > environment are the user's own, and are not treated as untrusted input. What *is* untrusted: content originating
   > in a scanned repository (source, doc comments, a discovered `.codescan.yaml`), and assets fetched at build time
   > and published to visitors of the documentation site.
   - Point being: the two real issues sit *inside* the stated boundary, so they read as consistent rather than as
     exceptions, and the dismissals get a citation.

3. 📝 😇 **Retire `hack/browser` — DECIDED 2026-08-17 (Fred): delete it.** A remnant of earlier prototypes,
   superseded by the playground; its own README already called it "not the playground — the seed of one". Deleting it
   retires a whole finding class rather than defending it, which is this plan's stated preference.
   Two inbound references have to go with it, or the deletion leaves dangling links:
   - `hack/doc-site/genspec-wasi/README.md:13` — links to `../../browser/README.md`
   - `.github/dependabot.yaml:92` — an exclusion comment explaining why `hack/browser` is deliberately unwatched;
     the exclusion becomes moot with the directory
   🔍 One judgement call left: the README carries **measurements** (the round-1 browser-side numbers) that have
   documentation value independent of the code. If they are worth keeping, move them into
   `hack/doc-site/genspec-wasi/README.md` in the same commit — otherwise they are lost with the directory, which is
   fine but should be deliberate rather than incidental.

### Done

4. ✅ 😇 **Control characters from scanned source into the terminal** — **FIXED AND MERGED**, PR #119
   (`ef5c149b`, "show control characters instead of obeying them"). Went from ♥️ someday to merged in
   three days. See Achievements; the original framing is kept below because it is the impact argument:
   - Real class (iTerm2 CVE-2019-9535 was RCE from displaying crafted sequences; OSC 52 can write the clipboard on
     permissive terminals), but impact is bounded and emulator-dependent: on modern defaults `allowWindowOps` is off
     and there is no generic execute-a-command sequence, so the realistic worst case is a garbled or spoofed pane.
   - If ever done: **one choke point** at the render boundary — strip C0/C1 except `\t`/`\n` as scanned strings enter a
     pane — never validation scattered through the model.

5. ✅ 😇 Doc-site supply chain, `SECURITY.md`-worthy boundary, all merged (see Achievements)
6. ✅ 📚 Local Mermaid recipe realigned with CI — **MERGED** 2026-08-15 as `804998c5` (PR #113)

## Achievements

### Triage rule ⭐⭐⭐

Two questions, in order. They have sorted 7 findings in a row without a single coin-flip:

1. **Name the boundary.** CWE-22 and friends require a *trust boundary crossing*. In a local dev tool running as the
   invoking user, **argv is not untrusted input — it is the user's intent**. A finding is real only if the value comes
   from somewhere the user does not control (network, uploaded archive, **a file inside the scanned repo**), or the
   process holds privileges its supplier lacks (setuid, daemon, CI acting on PR-supplied content).
2. **What does the attacker actually get?** This is what separates our two real findings from the dismissed ones under
   the *same* "hostile repo" premise: an arbitrary file write with attacker-chosen content on every platform, versus a
   garbled display on some terminal configurations. Same premise, two orders of magnitude apart in consequence.

Corollary worth remembering: Aikido escalates on *uncertainty* ("unknown origin … so path traversal risk is highest").
Unknown provenance should lower confidence, not raise severity. When a report says it could not trace the source, that
is a confession, not evidence.

### Fixed — doc-site supply chain ⭐⭐

All in master via the `fix/doc-site-mermaid` branch (hashes as rebased on merge):

1. ✅ `3681d380` **sanitize what the railroad shortcode injects** — Aikido's finding pointed at
   `el.innerHTML = out.svg`, but the reason it bit was two lines up: `securityLevel: 'loose'` switches off the
   sanitizer Mermaid otherwise runs over the SVG it returns. Now `'strict'`; both error paths build their `<pre>` with
   `textContent` instead of concatenating a renderer message into markup.
2. ✅ `a08b6007` **vendored Mermaid 11.16.0 → 11.16.1** — five published advisories, *all* fixed in that patch:
   prototype pollution via config APIs (GHSA-c4c3-pg64-4m4v) and architecture diagrams (GHSA-3rrr-jr9j-h3q3), CSS
   injection into sibling elements (GHSA-6x64-9x62-f2gx), two DoS loops (GHSA-2v8p-3f2j-5mp7, GHSA-rhh3-jpg6-66xh).
   Invisible to `npm audit`: Mermaid is curl'd, not a dependency.
3. ✅ `9da6e295` **checksum every asset the doc build downloads** — Mermaid and the Relearn tarball now come down with
   `-f --proto '=https' --tlsv1.2` and are verified against a recorded sha256 before anything is unpacked. The missing
   `-f` was a live bug on its own: curl wrote the body of an HTTP error response *as* `mermaid.min.js` and the build
   carried on.
4. ✅ `91d6f4d2` **browser probe status line built as text** — three `innerHTML` writes gone from `hack/browser/probe.js`.
   *(Superseded: that directory is now slated for deletion — see Outstanding 3. The fix was right at the time; retiring
   the surface is strictly better than keeping it hardened.)*

### Fixed — control characters into the terminal ⭐⭐

Merged 2026-08-17 as PR #119 (`ef5c149b`). The one item on this plan that started as ♥️ *someday* and
was built anyway — worth noting, because it was filed as bounded-impact and that assessment has not
changed. What changed is that it turned out to be cheap.

**The one-choke-point rule held, and the earlier doubt about it was wrong.** The concern raised while
the work was uncommitted was that nine edited render sites could not be one choke point. What shipped
is `cmd/genspec-tui/internal/ux/safetext/` — **one sanitizer with a two-mode API**, `Sanitize` (strict)
and `SanitizeBlock` (keeps layout), called from 14 sites across 7 files. That is the rule honoured, not
broken: the mandate was *one implementation at the render boundary*, never *one call*. Validation
scattered through the model is what it forbade, and there is none.

Covered by `safetext_test.go` plus four `escape_test.go` suites (ux, panels, diagnostics, validation) —
one per pane family, which is the right granularity for a rule about what reaches a pane.

**It deliberately excludes the clipboard**, and that exclusion is the fix agreeing with dismissal ⛔8
below rather than an oversight. `internal/ux/gadgets/clipboard.go` still emits OSC 52 and does not
import `safetext` (verified at `5e4cf04d`). It should not: emitting escape sequences is what a TUI *is*,
this one is safe by construction because the payload is base64 — no input can break out of the sequence
— and sanitizing it would break the feature to satisfy a pattern match. **The distinction the whole item
turns on: escape sequences codescan *writes* are fine; escape sequences it *relays from scanned source*
are not.** That is the line `safetext` draws, and it is the reason the fix and the dismissal are
consistent rather than contradictory.

### Fixed — documentation ⭐

5. ✅ `ed7e0dd0` (branch, unmerged) **local Mermaid recipe** — the note beside the vendored asset still named 11.16.0 and
   handed out a plain `curl -sL`, so the procedure a developer follows installed the version we had just moved off,
   unverified. We had hardened CI and left the human path behind. Caught only because Aikido's own evidence quoted that
   file.

### Verification techniques that earned their keep ⭐⭐

Worth reusing; each one changed a conclusion:

- **Headless Chromium for a JS-behaviour claim.** `chromium --headless=new --virtual-time-budget=30000 --dump-dom`
  against a locally served Hugo build, with the site under its `/codescan/` base path (a root-served build 404s on
  `relURL` assets and every diagram silently reports "failed to load Mermaid"). Proved `strict` costs nothing: 7/7
  diagrams render and each is **byte-identical** to the `loose` baseline.
- **Cross-check a pinned digest against a second origin.** A digest we computed from the CDN we are defending against is
  trust-on-first-use. Verified `18327bef…` against the npm registry: tarball hash matches the registry's own
  `dist.integrity` (`sha512-TQsq6u22…`), and `dist/mermaid.min.js` inside it hashes to exactly what we pinned. The
  package also carries SLSA provenance attestations if we ever want the publisher-signature arm too.
- **Run the workflow step for real.** Extracted the `run:` block with its `env:` into a script and executed it —
  including negative tests: tampered digest exits 1, missing asset exits 22.
- **OSV over the versions our own gates cannot see** — how the five Mermaid advisories surfaced at all.
- **Trace the chain by grep before judging.** For `profile.go`, every hop from `os.Create` back to `fs.String` was
  checked, plus a repo-wide grep proving no other construction site exists.

❌ **Gotcha that nearly produced a wrong answer:** a module worker caches its imports. A hostile-fixture test in the
browser returned the *old* result while the server was serving the new file; `node node-probe.js` bypassed the cache and
gave the truth. Do not trust a browser reload alone when the input is an imported module.

### Dismissed ⛔

6. ⛔ **`cmd/genspec-tui/.../profile.go:255` "path traversal", High** — invalid. Full chain: `os.Create(path)` ←
   `filepath.Join(p.report.Dir, <string literal>)` ← `Profiling{Dir: abs}` ← `filepath.Abs(*c.profileDir)` ← the
   `-profile-dir` flag, only with `-profile`, defaulting to `os.MkdirTemp`. Three claims in the report are factually
   wrong about this code: there is no config (the TUI does not use `cliconf` at all), no UI path (profiling is set at
   launch — `MemProfileRate` must be constant for the process), and all five joined filenames are compile-time
   literals, so there is no attacker-controlled *component* to traverse with. gosec: 0 issues.
   Suggested disposition text: *"Source is the `-profile-dir` CLI flag (`main.go:89`) in a local single-user tool;
   joined filenames are constants; no trust boundary is crossed."*
7. ⛔ **"the TUI does not validate input enough"** — invalid. The input is Go source that already survived the parser
   and type-checker; there is no boundary between the user's code and the user's terminal. Validation with no named
   boundary is the box-ticking we are refusing, in the hot path of a redraw loop.
8. ⛔ **"the TUI issues escape sequences"** — invalid, and the weakest report yet. That is
   `internal/ux/gadgets/clipboard.go:73`, OSC 52 with the tmux passthrough wrapper: emitting escape sequences is what a
   TUI *is*, and this one is safe by construction because the payload is base64 — no input can break out of the
   sequence. Flagging it means matching on `\x1b` rather than reasoning.

## Appendix: how Aikido is doing, so far

Fair assessment after 7 findings: **strong where a boundary genuinely exists and the chain is traceable** — the Mermaid
sink and the unverified download were both correct, well-argued, and the second one's remediation matched what we
shipped almost word for word. **Weak where it cannot trace provenance**, where it escalates on the uncertainty instead
of discounting for it, and where it rates sink shape without asking whether the source is reachable or the page even
ships (`hack/browser` is a throwaway that never leaves localhost).

Net: worth keeping if the triage rule above is applied by a human every time. Not worth wiring into a merge gate — it
would have blocked on three invalid highs, and it missed the one Go issue that is actually real.
