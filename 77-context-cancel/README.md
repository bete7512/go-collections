# 77 · Replace done with context.WithCancel

## The idea in one paragraph

Challenge #76 built cancellation by hand: a `chan struct{}`, a `close`, a `sync.Once` to make stopping idempotent. This challenge is the **same observable behavior** rewritten in the vocabulary every Go library actually speaks: `context.Context`. Under the hood a context's `Done()` **is** a done channel that gets closed on cancel — nothing magic. What you're buying by switching is everything around that channel:

1. **A cancellation tree** — contexts derive from parents (`context.WithCancel(parent)`), and cancelling a parent cancels every context derived from it, transitively. One `cancel()` at the top tears down a whole request's worth of goroutines.
2. **Deadlines and timeouts** — `WithTimeout`/`WithDeadline` (#78) are just cancellation wired to a clock.
3. **Request-scoped values** — `context.WithValue` for trace IDs and the like.
4. **One standard type** — every stdlib and third-party API (`database/sql`, `net/http`, gRPC, NATS…) accepts a `ctx`. Your hand-rolled done channel composes with none of them.

Be able to list those four from memory — the "Done when" line asks for exactly that.

## Vocabulary

| term | meaning |
|---|---|
| `context.Context` | interface carrying cancellation signal, deadline, and values |
| `ctx.Done()` | returns a `<-chan struct{}` that is **closed** when the context is cancelled — receive on it in a `select`, exactly like #76's done channel |
| `context.CancelFunc` | the `cancel` returned by `WithCancel`; calling it cancels the context (and its whole subtree) |
| `ctx.Err()` | `nil` while the context is live; `context.Canceled` after `cancel()`; `context.DeadlineExceeded` after a timeout (#78) |
| derived / child context | `ctx2, cancel2 := context.WithCancel(ctx)` — `ctx2` is cancelled when *either* `cancel2()` is called or `ctx` is cancelled |

## Conventions the tests and `go vet` hold you to

- **`ctx` is the first parameter, named `ctx`**, and is never stored in a struct field. A type that owns a lifecycle stores the **`cancel` func** (and a WaitGroup), not the context.
- **`defer cancel()` always**, even when you also call `cancel()` explicitly somewhere. `cancel` is idempotent, and skipping the deferred call leaks the context's internal resources; `go vet` (lostcancel) flags the missing case.
- **Never send on or close `Done()` yourself** — you can't; the type system gives you a receive-only channel. Cancellation happens only through `cancel()`.

---

## The functions

| function | what it teaches |
|---|---|
| `CountUntilCancel` | #76's `CountUntilDone`, with `<-ctx.Done()` as the stop case |
| `RunUntilCancel` | the spec's pinned signature — one `cancel()` stops N workers |
| `ProcessUntilCancel` | cancellable send + **reporting why you stopped** via `ctx.Err()` |
| `NewCancelRunner` | #76's `Stopper` rebuilt on `WithCancel` — and the `sync.Once` disappears |

---

## Function 1 · `CountUntilCancel(ctx context.Context) int`

Port of #76's `CountUntilDone`.

1. Loop, incrementing a counter each iteration.
2. Each iteration, check `ctx.Done()` without blocking (`select` with a `default` branch).
3. When the Done case fires, return the count.

**Returns:** however many iterations ran. Timing-dependent; the tests assert only that it returns promptly after cancel and that the count is positive when it was given time to spin.

**Traced example:**

```
caller: ctx, cancel := context.WithCancel(context.Background())
        defer cancel()                       ← always, even though we call it below
        go func() { result <- CountUntilCancel(ctx) }()
        ctx.Err() == nil                     ← still live
        ...let it spin...
        cancel()                             ← closes ctx.Done() under the hood

worker: iteration N: select → <-ctx.Done() fires → return N

caller: errors.Is(ctx.Err(), context.Canceled) == true
```

---

## Function 2 · `RunUntilCancel(ctx context.Context, workers int) int`

The spec's exact signature. Start `workers` goroutines, each looping like `CountUntilCancel` with its own count, all watching the **same** `ctx.Done()`. Wait for all of them, then return the **total** work done (sum of the per-worker counts).

**Pinned contract:**

- `workers <= 0` is treated as `1`.
- One `cancel()` releases every worker — same broadcast-by-close as #76, now hidden inside the context.
- Returns the sum; timing-dependent, so tests assert it is positive after a spin-up period and that the call returns within the timeout guard.
- **The tree:** the tests pass in a *child* context (`context.WithCancel(parent)`) and cancel the **parent**. Your function doesn't know or care — `ctx.Done()` fires either way. That's the point.

**Traced example (the tree):**

```
parent, cancelParent := context.WithCancel(context.Background())
child,  cancelChild  := context.WithCancel(parent)     // derived
go RunUntilCancel(child, 5)                            // 5 workers watch child.Done()

cancelParent()          ← cancel the PARENT

child.Done() fires too  → all 5 workers return
errors.Is(child.Err(), context.Canceled) == true
```

---

## Function 3 · `ProcessUntilCancel(ctx context.Context, jobs <-chan int, results chan<- int) (int, error)`

Port of #76's `ProcessUntilDone`, plus one upgrade: the function **reports why it stopped**.

1. Loop with a **blocking** `select` (no `default`): the `<-ctx.Done()` case, the `j, ok := <-jobs` receive, and the send of the processed result into `results` — the send must itself be a select case racing `ctx.Done()`, for exactly the reason #76 taught: a goroutine is only as cancellable as its most blocking line.
2. Process a job by doubling it (`j * 2`), same as #76.

**Pinned contract:**

- Returns `(count, nil)` when the **jobs channel closes** — normal drain, no error.
- Returns `(count, ctx.Err())` when **cancelled** — the tests assert `errors.Is(err, context.Canceled)`.
- `count` is the number of jobs fully processed: received **and** delivered into `results`.
- Already-cancelled context with no jobs available → returns `(0, context.Canceled)` promptly.

**Traced example:**

```
jobs: 1 2 3 4 ... (never closes)    results: unbuffered, nobody reading

  select: <-jobs fires → value 2
  select: results <- 2 ... blocks, but <-ctx.Done() is in the same select
  cancel()
  select: <-ctx.Done() fires → return (count, ctx.Err())

  caller: errors.Is(err, context.Canceled) == true
```

---

## Function 4 · `NewCancelRunner(workers int) *CancelRunner`

```go
type CancelRunner struct { /* cancel func, waitgroup, counts — NOT the ctx */ }

func NewCancelRunner(workers int) *CancelRunner
func (r *CancelRunner) Stop()          // idempotent: safe to call many times, concurrently
func (r *CancelRunner) Wait() []int    // blocks until all workers exited; per-worker counts
```

#76's `Stopper`, rebuilt on `context.WithCancel`.

1. `NewCancelRunner` creates a context internally, starts `workers` goroutines looping on its `Done()` (passing `ctx` as an argument — never storing it in the struct), and returns immediately. `workers <= 0` → 1.
2. `Stop()` calls `cancel()`. **Notice what's gone:** #76 needed `sync.Once` because closing a channel twice panics; `cancel()` is already idempotent, so the guard vanishes. Calling `Stop()` five times, or from twenty goroutines at once, must be harmless — the tests do both.
3. `Wait()` blocks on the WaitGroup, then returns the per-worker counts (one entry per worker).

**`Wait()` before `Stop()` blocks forever** — same ordering requirement as #76; the tests always call `Stop()` first.

---

## Rules that apply to all four

- **Workers must exit promptly** after cancellation — every test is timeout-guarded.
- **No goroutine outlives the cancel.** Goroutine counts are checked before and after.
- **`workers <= 0`** → treat as 1.
- **No `time.Sleep` in your implementations.**
- **`defer cancel()`** on every context you create — `go vet ./77-context-cancel/` must be clean.
- **`go test -race ./77-context-cancel/`** must be clean.

## Edge cases the tests hit

- Cancel called twice (safe — idempotent, no `sync.Once` anywhere in this challenge).
- Cancel before any worker starts (already-cancelled context → prompt return).
- A child context derived from a parent: **cancelling the parent stops workers watching the child** (the tree).
- `ctx.Err()` is `nil` before cancellation, `context.Canceled` after — asserted with `errors.Is`.
- The uncancellable-send leak: `results` with no reader, rescued only by a select-case send.
- Zero and negative worker counts.

## What the tests cover

**`CountUntilCancel`:** returns promptly after `cancel()`; positive count when given time to spin; already-cancelled context returns immediately; `ctx.Err()` nil-then-Canceled around the cancel; double-cancel harmless; 50 repeated start/cancel cycles.

**`RunUntilCancel`:** one cancel releases 1, 5, and 50 workers; positive total after a spin; zero/negative workers; already-cancelled context; **parent-cancels-child tree test**; goroutine-leak check.

**`ProcessUntilCancel`:** the no-reader leak test (must return after cancel, with `context.Canceled`); jobs-channel close → `nil` error and exact count; correct results delivered when a reader exists; already-cancelled → `(0, context.Canceled)`.

**`CancelRunner`:** Stop-then-Wait lifecycle with per-worker counts; `Stop()` five times and from 20 goroutines concurrently; zero workers treated as one; goroutine-leak check.
