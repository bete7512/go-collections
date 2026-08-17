# 69 · Close a channel and range until it drains

## Problem

`close` is the only way a sender can say **"no more values are coming"**. It's not a cleanup call — channels are garbage collected like anything else, so an unclosed channel with no references leaks nothing. Closing exists purely as a *signal*, and that signal is what lets `for v := range ch` terminate on its own.

The rules around it are short, absolute, and each one is a panic or a hang if you get it wrong. This drill turns every one of them into a passing test, including the three panics — captured with `recover`, because encoding a rule as an executable assertion beats memorizing it.

## Contract (what the tests enforce)

```go
func DrainAll(ch <-chan int) []int                  // range until closed
func DrainWithOK(ch <-chan int) ([]int, int)        // manual receive; returns values + how many receives happened
func SendAndClose(ch chan<- int, vals []int)        // sends all, then closes
func CountAfterClose(ch <-chan int, n int) []bool   // n receives past drain; each element is the ok flag
func SafeClose(ch chan int) (closed bool)           // closes, recovering from a double-close panic
```

- **`DrainAll`** ranges until the channel closes and returns the values in arrival order. Buffered values queued *before* the close are still delivered — closing does not discard them. Empty result is non-nil.
- **`DrainWithOK`** does the same with an explicit `for { v, ok := <-ch; if !ok { break } }` loop, returning both the values and the **total number of receive operations** — which is `len(values) + 1`, since the final receive is the one that reports `ok == false`. That off-by-one is the shape of the two-value receive, and pinning it makes it stick.
- **`SendAndClose`** takes a **send-only** `chan<- int`, sends everything, then closes. Note that a send-only channel *can* be closed — the direction restricts sends and receives, not closes, which is consistent with "only senders close".
- **`CountAfterClose(ch, n)`** performs `n` receives on an already-drained closed channel and returns the `ok` flag from each. All false, all immediate — receiving from a closed channel is idempotent and never blocks.
- **`SafeClose`** closes the channel inside a `defer`/`recover`, returning false if the close panicked (already closed). Document what it is: a **code smell**, not a utility. Needing it means channel ownership is unclear. The clean patterns are a single designated closer, or `sync.Once` — both appear in #73 and #100.

**The rules, each with a test:**

| rule | consequence |
|---|---|
| Only the sender closes | a receiver closing races the sender into a panic |
| Closing twice | **panic:** close of closed channel |
| Sending on a closed channel | **panic:** send on closed channel |
| Closing a nil channel | **panic:** close of nil channel |
| Receiving from a closed channel | never blocks; yields the zero value with `ok == false` |
| Closing is optional | only required if someone `range`s or otherwise waits for termination |

- **The zero-value ambiguity is the reason `ok` exists.** A closed channel and a genuinely-sent `0` both give you `0`; only the second return value distinguishes them. The tests send a real `0` before closing and require both cases to be told apart.

## Worked examples

**Example 1 — buffered values survive the close:**

```go
ch := make(chan int, 3)
ch <- 1; ch <- 2; ch <- 3
close(ch)                  // the sender is done — but the values are still queued

DrainAll(ch)               // → [1 2 3]   all three delivered, then the range ends
```

**Example 2 — the receive count off-by-one:**

```go
// channel carries 5 values, then closes
vals, receives := DrainWithOK(ch)
len(vals)   // 5
receives    // 6 — the last one is the receive that reported ok == false
```

**Example 3 — a real zero versus a closed channel:**

```go
ch := make(chan int, 1)
ch <- 0
close(ch)

v, ok := <-ch   // (0, true)   ← a genuine zero was sent
v, ok = <-ch    // (0, false)  ← now it's the closed-channel zero
```

Identical values, opposite meanings. `ok` is the whole difference.

## Edge cases the tests cover

- Draining a channel closed with buffered values still queued (3, 1, and 0 values).
- `DrainAll` on an already-closed empty channel → empty non-nil slice.
- The `len(values) + 1` receive count for several lengths.
- A genuinely-sent `0` distinguished from the post-close zero.
- `CountAfterClose` for n = 1, 5, 100 — every flag false, completing instantly (timeout-guarded).
- Double close → panic, recovered; `SafeClose` returning true then false on the same channel.
- Send on a closed channel → panic, recovered.
- Close of a nil channel → panic, recovered.
- `SendAndClose` with values, with an empty slice, and with nil, always followed by a successful drain.
- A 10,000-value produce-close-drain round trip, order preserved.
- All tests timeout-guarded; suite runs under `-race`.
