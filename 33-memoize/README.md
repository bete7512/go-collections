# 33 · Naive memoization

## Problem

Wrap a slow, pure function so that repeated calls with the same argument compute once and answer from a cache afterwards. The mechanism is a **closure**: `Memoize` returns a new function that captures a private map; the map lives on between calls because the returned function holds a reference to it. This drill is about really owning that mechanism — closures over shared state are how Go expresses what other languages need classes for.

## Contract (what the tests enforce)

```go
func Memoize(f func(int) int) func(int) int
```

- **Transparent caching.** For any argument, the wrapper returns exactly what `f` would return. Callers can't tell the difference — except by speed.
- **`f` runs exactly once per distinct argument.** Call the wrapper five times with `7` → `f(7)` executed once. The tests count invocations through a captured counter inside `f`.
- **A cached zero is a cache hit.** If `f(x)` returns `0`, calling the wrapper again with `x` must **not** re-invoke `f`. This is the trap the drill exists for: a cache check written as `if m[k] != 0` treats a legitimate zero result as a miss and recomputes forever. You need the comma-ok form: `if v, ok := m[k]; ok`.
- **Independent caches.** Every call to `Memoize` creates a fresh cache. Two wrappers around the same `f` don't share hits; memoizing two different functions never cross-contaminates.
- **Any int argument is valid** — negative and zero arguments cache like any other key.
- **Not safe for concurrent use — and that's part of the deliverable.** Concurrent calls to the wrapper would be concurrent map read/writes, which the Go runtime kills with a fatal error (not even a recoverable panic). Write that limitation as a doc comment on `Memoize`. Making it safe is #79's territory (mutex) — here, naive is the assignment.

## Worked examples

**Example 1 — cache hit:**

```go
calls := 0
slow := func(x int) int { calls++; return x * 2 }
fast := Memoize(slow)

fast(5)   // → 10, calls == 1  (computed)
fast(5)   // → 10, calls == 1  (cached — slow never ran)
fast(6)   // → 12, calls == 2  (new argument, computed)
```

**Example 2 — the zero-value trap:**

```go
calls := 0
zero := Memoize(func(x int) int { calls++; return 0 })

zero(42)  // → 0, calls == 1
zero(42)  // → 0, calls must STILL be 1
```

If your cache lookup is `!= 0`-based, the second call recomputes and `calls` becomes 2 — the tests fail exactly here.

**Example 3 — independent caches:**

```go
a := Memoize(slow)
b := Memoize(slow)
a(1)      // computes
b(1)      // computes AGAIN — b has its own empty cache
```

## Edge cases the tests cover

- Repeated calls with the same argument (once-per-distinct-arg counting).
- A function returning `0` for everything — the comma-ok trap.
- Negative arguments and argument `0` as cache keys.
- Two wrappers over the same function: separate caches, separate counters.
- Wrappers over two *different* functions returning different values for the same argument — no cross-talk.
- Interleaved arguments (`1,2,1,3,2,1`) — count stays at the number of distinct values.
- 100 distinct arguments each called 3 times → exactly 100 underlying invocations, and every returned value still correct on the third round.
