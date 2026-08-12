# 68 · Buffered channels and when a send blocks

## Problem

Unbuffered and buffered channels look identical at the call site — `ch <- v` either way — but they mean different things. An **unbuffered** channel is a rendezvous: the send blocks until a receiver is ready, so the two goroutines synchronize at that instant. A **buffered** channel is a queue with a fixed size: the send blocks only when the buffer is full, and the sender may run far ahead of the receiver.

Choosing wrong is a real bug class. Unbuffered where you needed a queue means the producer stalls on every item; buffered where you needed synchronization means "the value was sent" no longer tells you "the other side got it". This drill makes the rules concrete — and does it **without a single `time.Sleep`**, using the non-blocking `select`/`default` idiom, which is the only honest way to test blocking behavior.

## Contract (what the tests enforce)

```go
func TrySend(ch chan int, v int) bool     // non-blocking send: false if it would block
func TryReceive(ch chan int) (int, bool)  // non-blocking receive: false if it would block
func FillBuffer(ch chan int, vals []int) int  // sends until full; returns how many landed
func Stats(ch chan int) (length, capacity int)
```

- **`TrySend`** attempts a send in a `select` with a `default` branch. Returns true if the value went in, false if the send would have blocked. It must **never** block, for any channel state.
- **`TryReceive`** is the mirror: returns `(v, true)` if a value was available, `(0, false)` if a receive would block. Note the subtlety pinned by the tests: on a **closed** channel a receive never blocks, so `TryReceive` returns `(0, true)` — it did receive, it received the zero value from a closed channel. Distinguishing that from a real zero requires the three-value form, which `TryReceive` deliberately doesn't expose; document the limitation.
- **`FillBuffer`** repeatedly `TrySend`s the values in order and returns the count accepted before the channel refused. Never blocks.
- **`Stats`** returns `len(ch)` (values currently buffered) and `cap(ch)` (buffer size). Note in a comment that both are inherently racy under concurrency — fine for observation and tests, never for logic.

**The rules being drilled, which the tests assert directly:**

| | send blocks when | receive blocks when |
|---|---|---|
| **unbuffered** (`make(chan int)`) | no receiver is ready — **always**, absent a partner | no sender is ready |
| **buffered** (`make(chan int, n)`) | the buffer is full | the buffer is empty |
| **closed** | always panics, never blocks | never — returns zero immediately |
| **nil** (`var ch chan int`) | forever | forever |

- **The nil channel row is not trivia.** A nil channel blocks both directions permanently, which sounds useless until #70, where setting a channel variable to nil is precisely how you disable a `select` case. `TrySend`/`TryReceive` on a nil channel must return false — the `default` branch wins, no panic.
- **Send on a closed channel panics**, and the tests verify that with `recover`. `TrySend` cannot protect you: `select` picks the send case (it's "ready") and the panic happens. There is no non-blocking way to safely send on a possibly-closed channel — which is why close-ownership discipline matters more than defensive checks.

## Worked examples

**Example 1 — a capacity-2 buffer filling and refusing:**

```go
ch := make(chan int, 2)
TrySend(ch, 1)   // true    len 1, cap 2
TrySend(ch, 2)   // true    len 2, cap 2 — full
TrySend(ch, 3)   // FALSE   would block; 3 was not sent
TryReceive(ch)   // (1, true)   len 1 — a slot freed
TrySend(ch, 3)   // true    len 2 again
```

No sleeps, no goroutines, fully deterministic.

**Example 2 — unbuffered has no room, ever:**

```go
ch := make(chan int)      // cap 0
TrySend(ch, 1)            // FALSE — nobody is receiving right now
TryReceive(ch)            // (0, false) — nobody is sending
```

An unbuffered channel is never "ready" on its own; it needs a partner at the same moment.

**Example 3 — the closed-channel asymmetry:**

```go
ch := make(chan int, 1)
close(ch)
TryReceive(ch)   // (0, true)  — receives never block on a closed channel
TrySend(ch, 1)   // PANIC: send on closed channel
```

## Edge cases the tests cover

- Capacity 1, 2, and 5 buffers: fill to capacity, refuse the next, drain one, accept one more.
- `len`/`cap` after every operation in a scripted sequence.
- Unbuffered channel: `TrySend` and `TryReceive` both false with no partner.
- Unbuffered channel **with** a waiting receiver: `TrySend` succeeds (a goroutine parked on receive makes the send ready) — synchronized with a ready-signal channel, not a sleep.
- Nil channel: both operations return false, no panic.
- Closed channel: `TryReceive` returns `(0, true)`; `TrySend` panics (recovered).
- `FillBuffer` with more values than capacity (returns cap), fewer (returns len(vals)), exactly capacity, empty slice, and on a nil channel (returns 0).
- Draining a full buffer completely, then confirming FIFO order was preserved.
- A 1,000-element buffer filled and drained, verifying order and counts.
- Everything runs without sleeping; the suite is deterministic and `-race` clean.
