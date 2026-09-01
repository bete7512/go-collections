# 76 · Stop goroutines with a done channel

## The idea in one paragraph

Everything so far ended **naturally**: workers stopped because the jobs channel closed, forwarders stopped because their input drained. Now you need to stop goroutines that would otherwise run forever — a poller, a background loop, a worker waiting on a queue that never empties. The tool is a **done channel**, and the mechanism is one small, beautiful fact:

> **Closing a channel makes every receiver's case fire, immediately and permanently.**

A *send* is received by exactly one goroutine. A *close* is seen by **all** of them. That's why cancellation uses `close(done)` and never `done <- struct{}{}` — one close stops a thousand goroutines with no counting, no per-goroutine bookkeeping, no risk of sending the wrong number of signals.

This is also where **goroutine leaks** become real. A goroutine blocked forever on a channel nobody will ever write to is never garbage collected — it holds its stack and everything it references until the process dies. Go won't warn you. The tests here count goroutines before and after to make leaks visible.

---

## The functions

| function | what it teaches |
|---|---|
| `CountUntilDone` | the basic loop: work until told to stop |
| `WorkersUntilDone` | one close stops N goroutines at once (broadcast) |
| `ProcessUntilDone` | done + real work in one `select`, and why send-outside-select leaks |
| `NewStopper` | wrapping the pattern in a type with an idempotent `Stop()` |

---

## Function 1 · `CountUntilDone(done <-chan struct{}) int`

**What it gets:** a receive-only done channel. Someone else closes it.

**What it must do:**

1. Loop, incrementing a counter each iteration.
2. Each iteration, check the done channel **without blocking** — a `select` with the done case and a `default` branch, so the loop keeps working while done is still open.
3. When the done case fires (the channel was closed), return the count.

**What it must return:** however many iterations ran. The exact number is timing-dependent and the tests never check it — they check that the function **returns at all**, promptly, after the close.

**Why `<-chan struct{}`:** receive-only means the worker cannot close the signal it's supposed to observe — the type system enforcing ownership again. And `struct{}` because the channel carries no data, only the fact of its closure; an empty struct is zero bytes.

**Traced example:**

```
caller: done := make(chan struct{})
        go func() { result <- CountUntilDone(done) }()
        ...let it spin...
        close(done)          ← broadcast

worker: iteration 1: select → default → count=1
        iteration 2: select → default → count=2
        ...
        iteration N: select → <-done fires (channel closed) → return N
```

---

## Function 2 · `WorkersUntilDone(done <-chan struct{}, workers int) []int`

**What it gets:** the same done channel, plus how many workers to run.

**What it must do:**

1. Start N goroutines, **all watching the same done channel**.
2. Each loops like `CountUntilDone`, keeping its own count.
3. `wg.Wait()` for all of them, then return a slice of the per-worker counts (one entry per worker, index = worker ID).

**What it must return:** exactly `workers` counts. Values are timing-dependent; the length is not.

**The point of this function:** one `close(done)` releases all N workers **simultaneously**. Compare the alternative — sending N times — which requires knowing N, sending exactly that many values, and hoping no worker takes two. Close-as-broadcast has none of those problems. Write that comparison in a comment.

**Traced example:**

```
5 workers, all blocked in select on the same done channel

  close(done)  ←── ONE operation

  worker 0 → done fires → returns
  worker 1 → done fires → returns      all five, from one close
  worker 2 → done fires → returns
  worker 3 → done fires → returns
  worker 4 → done fires → returns

  wg.Wait() unblocks  →  [1204, 998, 1301, 1150, 1077]
```

---

## Function 3 · `ProcessUntilDone(done <-chan struct{}, jobs <-chan int, results chan<- int) int`

**What it gets:** a done channel, a jobs channel to read from, and a results channel to write to.

**What it must do:**

1. Loop with a **blocking** `select` (no `default` this time) over **three** cases:
   - `<-done` → stop and return.
   - `j, ok := <-jobs` → if `ok` is false the jobs channel closed, so stop and return; otherwise process the job.
   - the **send** of the processed result into `results` — see the warning below.
2. Return the number of jobs processed.

**The leak this function exists to teach.** The obvious version writes the result *after* the select:

```
select {
case <-done:      return count
case j := <-jobs: value := j * 2
}
results <- value        ← OUTSIDE the select
```

If nobody is receiving from `results`, that send blocks **forever** — and `done` closing does nothing, because this goroutine isn't in a select anymore. You've built a goroutine that cannot be cancelled. The fix is to make the send itself a select case, racing it against done:

```
select {
case results <- value:  count++
case <-done:            return count
}
```

Now every blocking operation the goroutine performs is cancellable. That's the rule worth carrying: **a goroutine is only as cancellable as its most blocking line.**

**What it must return:** the count of jobs fully processed (received *and* delivered).

**Traced example:**

```
jobs: 1 2 3 4 5 ...  (never closes)     results: unbuffered, nobody reading

  select: <-jobs fires → value 2
  select: results <- 2 ... blocks, but <-done is also in this select
  close(done)
  select: <-done fires → return count

  → returns promptly instead of hanging forever
```

---

## Function 4 · `NewStopper(workers int) *Stopper`

```go
type Stopper struct { /* done channel, waitgroup, counts, sync.Once */ }

func NewStopper(workers int) *Stopper
func (s *Stopper) Stop()          // idempotent: safe to call many times
func (s *Stopper) Wait() []int    // blocks until all workers exited; per-worker counts
```

**What it must do:**

1. `NewStopper` creates the done channel, starts `workers` goroutines looping on it, and returns immediately.
2. `Stop()` closes the done channel — but **closing twice panics**, so guard it with `sync.Once`. Calling `Stop()` five times must be harmless.
3. `Wait()` blocks on the WaitGroup, then returns the per-worker counts.

**Why wrap it:** callers shouldn't have to know there's a channel in there, and they certainly shouldn't be able to close it twice by accident. This is the ordinary way lifecycle-owning types are built in Go — you'll reuse the shape in #94's TTL store and #100's broker.

**`Wait()` before `Stop()` would block forever** — document that ordering requirement. The tests always call `Stop()` first.

---

## Rules that apply to all four

- **`close(done)` is the signal; never send on it.** Never close it from inside a worker.
- **Workers must exit promptly** after the close — every test is timeout-guarded.
- **No goroutine outlives the stop.** Goroutine counts are checked before and after.
- **`workers <= 0`** → treat as 1.
- **No `time.Sleep` in your implementations** (the tests use tiny sleeps to let workers spin up; your code shouldn't need any).
- **`go test -race ./76-done-channel/`** must be clean — several workers touch shared state here.

---

## What the tests cover

**`CountUntilDone`:** returns promptly after close; returns immediately when handed an already-closed channel; runs a nonzero number of iterations when given time; and 50 repeated start/stop cycles all terminating.

**`WorkersUntilDone`:** exactly `workers` counts returned for 1, 5, and 50 workers; all workers released by a single close; zero/negative worker counts treated as 1; an already-closed done channel returning immediately.

**`ProcessUntilDone`:** stops on done with an unbuffered `results` channel that nobody reads — **the leak test**, which hangs unless the send is inside the select; stops when the jobs channel closes; processes jobs correctly when a reader is present; returns an accurate count; and an already-closed done channel returning 0.

**`Stopper`:** `Stop()` then `Wait()` returns per-worker counts; **`Stop()` called five times does not panic**; `Stop()` from several goroutines concurrently is safe; goroutine count returns to baseline after `Wait()`.

**Leak checks** on all four, using `runtime.NumGoroutine()` before and after with a settle delay.
