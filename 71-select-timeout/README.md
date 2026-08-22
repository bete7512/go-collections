# 71 · Add a timeout to select

## Problem

Waiting forever is a bug. Any blocking receive in production code needs an escape hatch, and `select` gives you one for free: add a case that becomes ready after a deadline, and whichever fires first wins.

`time.After(d)` is the simplest form — it returns a `<-chan Time` that receives once, after `d`. It's also the one with a caveat worth carrying: **inside a loop, `time.After` allocates a fresh timer on every iteration**, and each one stays alive until it fires. In a hot loop with a long timeout that's real garbage and real memory. The fix is `time.NewTimer` + `Reset`, or a context deadline. This drill has you build both forms so the difference is muscle memory rather than a blog post you half-remember.

## Contract (what the tests enforce)

```go
func ReadWithTimeout(ch <-chan int, d time.Duration) (int, bool)
func ReadWithTimer(ch <-chan int, d time.Duration) (int, bool)     // reusable timer, no per-call allocation churn
func DrainWithIdleTimeout(ch <-chan int, idle time.Duration) []int // collect until the channel goes quiet
```

- **`ReadWithTimeout(ch, d)`** blocks in a `select` on `ch` and `time.After(d)`.
  - A value arrives in time → `(v, true)`.
  - Nothing arrives within `d` → `(0, false)`, returning at roughly `d`, not later.
  - **A closed channel is not a timeout.** Its receive fires immediately with `ok == false`; return `(0, false)` for that too — but the tests assert it returns *fast*, well under `d`, so you can't conflate the two by just letting the timer expire.
  - `d <= 0` → `time.After` fires essentially immediately; the tests only require a prompt `(0, false)` when no value is waiting, and accept either outcome when a value is already buffered (both cases are ready, so `select` picks at random — the tests do not pin it).
  - A nil channel never fires, so the timeout always wins.
- **`ReadWithTimer(ch, d)`** does the same job with `timer := time.NewTimer(d)` and `defer timer.Stop()`. Same observable behavior; the point is the shape.
  - Note in a comment why `Stop` matters: an unstopped timer holds a runtime entry until it fires. (Since Go 1.23 unreferenced timers are collectible, so this is no longer a hard leak — but `Stop` remains correct and cheap, and older code depends on it.)
  - The tests run this one in a 500-iteration loop, which is where the allocation difference lives.
- **`DrainWithIdleTimeout(ch, idle)`** collects values until either the channel closes **or** no value arrives for `idle`. Returns everything received so far, empty non-nil if nothing did.
  - This is the realistic shape: a per-value deadline reset each time something arrives, so a steady stream never times out but a stall does.
  - Implement it with a reusable timer (`Reset` after each value) — it's a loop, which is exactly where `time.After` misbehaves.
  - Note the `Reset` rule in a comment: `Reset` is only safe on a timer that has been stopped or has already fired and been drained. The tests exercise both paths.
- No `time.Sleep` in your implementations. Durations in the tests are tens of milliseconds so the suite stays fast.
- **Timeouts don't stop the producer.** A goroutine still sending after you gave up keeps running — you stopped *waiting*, you didn't cancel anything. That distinction is what #77 and #78 fix with context; write it in a comment here.

## Worked examples

**Example 1 — value versus deadline:**

```go
ch := make(chan int, 1)
ch <- 42
ReadWithTimeout(ch, 500*time.Millisecond)   // (42, true) — immediately

empty := make(chan int)
ReadWithTimeout(empty, 50*time.Millisecond) // (0, false) — after ~50ms
```

**Example 2 — closed is not slow:**

```go
closed := make(chan int)
close(closed)
start := time.Now()
ReadWithTimeout(closed, 5*time.Second)      // (0, false) after microseconds, not 5s
```

Both return `false`, but for opposite reasons — one because there's nothing left, one because time ran out.

**Example 3 — idle timeout on a stream:**

```go
// producer sends 1,2,3 quickly, then stalls
DrainWithIdleTimeout(ch, 50*time.Millisecond)   // → [1 2 3], returns ~50ms after the third value
```

The deadline resets on every arrival, so a fast stream of a thousand values never trips it.

## Edge cases the tests cover

- Value already buffered → returned immediately, well under the timeout.
- Value arriving after a short delay but before the deadline.
- Nothing arriving → `false`, with elapsed time in a sane window (≥ most of `d`, and not wildly beyond it).
- Closed channel → `false` **fast**, not after the timeout.
- Nil channel → timeout wins.
- `d <= 0` on an empty channel → prompt `false`.
- Both `ReadWithTimeout` and `ReadWithTimer` run through the same table.
- `ReadWithTimer` in a 500-iteration loop, all succeeding.
- `DrainWithIdleTimeout`: a fast stream of 1,000 values with a short idle window (deadline must reset per value); a producer that stalls mid-stream (returns what arrived); a channel that closes normally (returns everything, promptly); an immediately-closed channel (empty non-nil); a channel that never sends (empty after ~idle).
- All timing assertions use generous margins; suite runs under `-race`.
