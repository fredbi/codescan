# Flaky Async Tests: Timing Ratios on CI Shared Runners

**Extracted:** 2026-02-08
**Context:** Fixing flaky time-based tests that pass locally but fail intermittently on CI shared runners

## Problem

Tests that assert timeout behavior using `time.Sleep` vs `context.WithTimeout` (or similar timer pairs) fail intermittently on CI shared runners. The root cause is goroutine scheduling jitter: on heavily loaded hosts, both the "slow condition" and the "timeout" timers fire within the same scheduling window, making the outcome non-deterministic.

A 5:1 ratio (e.g., 5ms sleep vs 1ms timeout) is NOT enough — CI runners can easily distort timings by 5-10x.

## Solution

Use a **100x ratio** between the slow operation and the timeout. This makes it statistically impossible for scheduler jitter to close the gap, even on the slowest CI runners.

```go
// BAD: 5:1 ratio — flaky on CI
condition := func() bool {
    time.Sleep(5 * time.Millisecond)  // only 5x the timeout
    return true
}
Eventually(mock, condition, time.Millisecond, time.Microsecond)

// GOOD: 100:1 ratio — deterministic even on slow runners
condition := func() bool {
    time.Sleep(100 * time.Millisecond)  // 100x the timeout
    return true
}
Eventually(mock, condition, time.Millisecond, time.Microsecond)
```

### Future: testing/synctest (Go 1.25+)

For projects targeting Go 1.25+, `testing/synctest` provides fake time inside a bubble, eliminating timing flakiness entirely. All `time.Sleep`, `time.Ticker`, and `context.WithTimeout` calls use deterministic fake time. This is the preferred long-term solution for async assertion testing.

## When to Use

- Any test that pairs a `time.Sleep` against a `context.WithTimeout` or `time.After` to assert timeout behavior
- Tests using `Eventually`, `Never`, or similar polling patterns with deliberate timeouts
- Any time-based test that passes locally but fails intermittently on CI
- When the test logic is "this should take longer than the deadline" — the gap must be wide enough to survive scheduling distortion
