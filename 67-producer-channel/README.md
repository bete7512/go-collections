# 67 · Producer goroutine over a channel

## Problem

#66 used a WaitGroup to know *when* goroutines finished. Channels do something better: they carry the values out, and they synchronize as a side effect of that transfer. "Don't communicate by sharing memory; share memory by communicating" is the Go proverb, and this is the smallest program that demonstrates it.

Two rules carry most of the weight, and both are about **who closes**. A channel close is a broadcast from sender to receiver meaning "no more values are coming" — so only a sender may close it, and a `range` over a channel that nobody closes waits forever. Getting this wrong is how you meet `fatal error: all goroutines are asleep - deadlock!`, which the README asks you to trigger deliberately once.

## Contract (what the tests enforce)

```go
func Produce(n int) <-chan int                 // sends 1..n, then closes
func Collect(ch <-chan int) []int              // ranges until the channel closes
func ProduceSquares(nums []int) <-chan int     // sends each input squared, then closes
```

- **`Produce(n)`** creates a channel, launches **one** goroutine that sends `1, 2, …, n`, closes the channel when the loop ends, and returns the channel **immediately** — before the sends happen. The function must not block.
  - `n <= 0` → a channel that is closed with nothing sent. `Collect` on it returns an empty (non-nil) slice.
  - The sender closes. Never the receiver.
- **Return type is `<-chan int`** — receive-only. This is not decoration: the type system now prevents a caller from sending or closing, which is exactly the constraint that makes the ownership rule enforceable rather than a convention. A caller trying `close(ch)` gets a compile error.
- **Order is guaranteed here.** One sender and one receiver over a single channel preserves order — each send pairs with exactly one receive, sequentially. (This stops being true with multiple senders; #74/#75 revisit it.)
- **`Collect(ch)`** ranges over the channel until it closes and returns the values in arrival order. `for v := range ch` handles the close automatically — no `ok` check needed, no counting.
  - A channel that closes with no values → empty non-nil slice.
- **`ProduceSquares(nums)`** is the same shape with a transformation: one goroutine, sends `nums[i] * nums[i]` in order, closes when done. Empty/nil input → immediately closed channel.
- **The channel is unbuffered** (`make(chan int)`) unless you have a reason otherwise — the send blocks until the receiver is ready, which is the synchronization. Buffering is an optimization, not a default.
- **The deliberate failure (not tested — it kills the process):** write a variant that never closes, range over it, and read `fatal error: all goroutines are asleep - deadlock!`. Then notice the limitation that matters: the runtime detects this only when **every** goroutine is blocked. A leaked goroutine in a program with other work running is completely silent. That asymmetry is why goroutine leaks are hard, and it motivates #76's done channel.

## Worked examples

**Example 1 — the basic pipeline:**

```go
ch := Produce(5)        // returns immediately; the goroutine is still sending
Collect(ch)             // → [1 2 3 4 5]   in order, then the range ends on close
```

**Example 2 — receiving manually, and what close looks like:**

```go
ch := Produce(2)
v, ok := <-ch    // (1, true)
v, ok = <-ch     // (2, true)
v, ok = <-ch     // (0, false)  ← drained and closed: zero value, ok == false
v, ok = <-ch     // (0, false)  ← still, forever. Receiving from a closed channel never blocks.
```

The two-value receive is how you distinguish "a real zero was sent" from "the channel is closed".

**Example 3 — the empty case:**

```go
ch := Produce(0)   // channel closed with nothing sent
Collect(ch)        // → []   — the range body never executes
```

## Edge cases the tests cover

- `Produce` for n = 1, 5, 100 — exact order asserted.
- `n = 0` and negative n → closed channel, empty non-nil collection.
- `Produce` returning before the values are consumed (the call must not block).
- The channel actually closing: a manual drain loop reaching `ok == false`, and repeated receives after close still returning `(0, false)`.
- Receiving from a closed channel never blocking (checked with `select`/`default`).
- `Collect` on an already-closed empty channel.
- `ProduceSquares` on ordinary values, negatives (squares positive), zeros, empty and nil input.
- A 10,000-value production consumed in full, in order.
- **Every test wrapped in a timeout guard** — a missing `close` would otherwise hang the suite forever instead of failing. That harness (a goroutine plus `select` on `time.After`) is itself a technique worth having.
- The whole suite under `-race`.
