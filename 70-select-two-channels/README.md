# 70 · select over two channels

## Problem

`select` is the multiplexer: it waits on several channel operations at once and proceeds with whichever becomes ready first. Every non-trivial Go concurrency pattern — timeouts, cancellation, fan-in, worker loops — is `select` plus something.

It also contains the single most common concurrency bug in Go, and this drill is built around making you hit it and fix it. **A closed channel is permanently ready.** Its receive case fires immediately, forever, yielding zero values — so a naive `select` loop over a closed channel spins at 100% CPU producing garbage. The fix is a genuine idiom, not a workaround: set the channel *variable* to `nil`, because a nil channel blocks forever and a `select` case on it can therefore never be chosen. You disable a case by nilling its channel.

## Contract (what the tests enforce)

```go
func Merge2(a, b <-chan int) []int          // drains both until both are closed
func Tagged(a, b <-chan int) []string       // same, but records which source each value came from
func FirstReady(a, b <-chan int) (int, bool)// one value from whichever is ready first
```

- **`Merge2(a, b)`** loops on a `select` with one receive case per channel, collecting every value from both, and **returns only when both channels are closed and drained**.
  - When a receive reports `ok == false`, that channel is done: **set the local variable to nil** so its case stops being selected. When both are nil, exit the loop.
  - Must not spin: the tests run it under a timeout, and a CPU-burning loop that never exits fails there. A correctness test also checks it terminates when one channel closes long before the other.
  - Both channels already closed at entry → empty non-nil result, immediate return.
  - Either channel may be nil from the start (a nil channel never fires — treat it as already done).
- **`Tagged(a, b)`** does the same but returns strings like `"a:1"` and `"b:20"`, so the tests can verify *which* source each value came from and that **each source's own values arrive in order**. Interleaving between sources is unspecified — the tests never assert it.
- **`FirstReady(a, b)`** performs exactly one `select`, returning the first value received and true, or `(0, false)` if both channels are closed. It blocks until something is available — this is the plain blocking `select`, no `default`.
- **Randomness is real and must not be defeated.** When multiple cases are ready simultaneously, `select` picks one **uniformly at random** — deliberately, to prevent starvation. The tests assert totals and per-source ordering, never a specific interleaving. One test does exercise the randomness directly: with both channels always ready, 1,000 rounds must produce a mix from both sources rather than always the first case.
- No `time.Sleep` in your implementation. The tests use timeout guards; your code should be purely event-driven.

## Worked examples

**Example 1 — the nil-disabling idiom:**

```
a sends 1,2,3 then closes
b sends 10,20 then closes

loop:
  select on a and b
  a reports ok=false  →  a = nil        (its case can never fire again)
  b reports ok=false  →  b = nil
  both nil            →  exit
```

`Merge2` returns 5 values: `{1,2,3,10,20}` in some interleaving.

**Example 2 — what goes wrong without it:**

```go
for {
    select {
    case v := <-a:   // a is CLOSED: this fires instantly, v == 0
        collect(v)   // ...forever. 100% CPU, infinite zeros.
    case v := <-b:
        collect(v)
    }
}
```

Closed ≠ blocked. Closed means *always ready*. This loop never sleeps and never ends.

**Example 3 — per-source order holds, interleaving does not:**

```go
Tagged(a, b)
// → ["a:1" "b:10" "a:2" "a:3" "b:20"]   ← one legal outcome
// → ["b:10" "b:20" "a:1" "a:2" "a:3"]   ← equally legal
// but "a:2" ALWAYS precedes "a:3", and "b:10" always precedes "b:20"
```

## Edge cases the tests cover

- Both channels with values, closing at different times.
- One channel empty-and-closed from the start; the other still producing.
- Both closed and empty at entry → immediate empty result.
- A nil channel passed as `a`, as `b`, and as both.
- Very unequal lengths (1 value vs 1,000) — the short one closing early must not stall or spin.
- Per-source ordering preserved in `Tagged`, verified by extracting each source's subsequence.
- Total counts exact across 50 repeated runs (catches dropped or duplicated values).
- `FirstReady` with only `a` ready, only `b` ready, both ready, and both closed.
- The randomness check: both sources always ready, 1,000 selections, both must appear (a strictly-ordered `select` would fail).
- Timeout guards on everything; suite runs under `-race`.
