# 78 · context.WithTimeout on a slow operation

## The idea in one paragraph

#77 cancelled work because *you* decided to; this challenge cancels work because *the clock* decided to. `context.WithTimeout(parent, d)` (and its sibling `WithDeadline`) is #77's `WithCancel` wired to a timer: when `d` elapses, the context cancels itself, `Done()` closes, and `ctx.Err()` reports **`context.DeadlineExceeded`** instead of `context.Canceled`. Two distinct sentinel errors — which one you get tells you *why* the work stopped, and the tests assert the distinction both ways.

**The real lesson** is in the second function: a timeout that returns early while the underlying work grinds on to completion is *theater* — you've hidden the latency from the caller but reclaimed nothing; the goroutine, the CPU, the connection are all still burning. Work is only abandonable if it's **ctx-aware**: `http.NewRequestWithContext`, `sql.QueryContext`, and most modern APIs check the context at their blocking points. A pure CPU loop checks nothing — if you want it interruptible, *you* must put a check inside the loop. The "Done when" line asks you to name two ctx-aware stdlib calls and one thing a context cannot interrupt — have them ready.

## Vocabulary (on top of #77's)

| term | meaning |
|---|---|
| `context.WithTimeout(parent, d)` | context that self-cancels `d` from now; also returns a `CancelFunc` |
| `context.WithDeadline(parent, t)` | same, but at an absolute time `t`; `WithTimeout` is sugar over it |
| `context.DeadlineExceeded` | `ctx.Err()` after the deadline fired |
| `context.Canceled` | `ctx.Err()` after an explicit `cancel()` — even on a timeout context, if you cancel *before* the deadline |
| `ctx.Deadline()` | `(time.Time, bool)` — the deadline, and whether one is set |

**`defer cancel()` still, always** — even though a timeout context cancels itself. The deferred call releases the timer and the context's resources immediately instead of holding them until the deadline; `go vet` (lostcancel) flags the omission.

---

## The functions

| function | what it teaches |
|---|---|
| `SlowOp` | racing simulated work against `ctx.Done()`; `DeadlineExceeded` vs `Canceled` |
| `SumToN` | a pure CPU loop is uninterruptible until *you* add the ctx check |
| `CallWithTimeout` | the caller side: create the timeout context, `defer cancel()`, call |

---

## Function 1 · `SlowOp(ctx context.Context, d time.Duration) (string, error)`

The spec's pinned signature. Simulates an operation that takes `d` to complete.

1. `select` on `case <-time.After(d)` (the work "finishing") versus `case <-ctx.Done()` (give up).
2. If the work finishes first → return `("done", nil)`.
3. If the context ends first → return `("", ctx.Err())`.

**Pinned contract:**

- The success value is exactly `"done"`; the value on any error is exactly `""`.
- The error is `ctx.Err()` verbatim: `context.DeadlineExceeded` when a deadline fired, `context.Canceled` when someone called `cancel()` — the tests assert each, and assert the *other* one didn't fire.
- A context that is already expired or cancelled on entry returns immediately.
- **The elapsed-time assertion is the proof:** an 80ms timeout against 2s of "work" must return in well under 1s. Returning `DeadlineExceeded` after 2s would pass the error check and fail the clock.

**Traced example:**

```
ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
defer cancel()

SlowOp(ctx, 2*time.Second)

  t=0ms    select waits on: time.After(2s) | ctx.Done()
  t=80ms   deadline fires → ctx.Done() closes
           select: <-ctx.Done() → return ("", ctx.Err())

  → ("", context.DeadlineExceeded) after ~80ms, not 2s
```

---

## Function 2 · `SumToN(ctx context.Context, n int64) (int64, error)`

The anti-theater drill: a CPU-bound loop that must be **made** interruptible.

1. Sum the integers `1..n` in a plain loop.
2. Inside the loop, check `ctx.Done()` without blocking (the #77 `select`/`default` shape). No check → the context is powerless: nothing in `i++; sum += i` ever looks at it.
3. If the context ends before the loop finishes, stop and return `(partial sum so far, ctx.Err())`.

**Pinned contract:**

- `n <= 0` → `(0, nil)`.
- Completed → `(n·(n+1)/2, nil)`; the tests check `SumToN(ctx, 10) == 55`.
- Interrupted → the partial sum and `ctx.Err()`. The partial value is timing-dependent, so the tests assert only the error and the promptness of the return — the contract is that you return *what you had*, not zero.
- The judge calls `SumToN(ctx, math.MaxInt64)` with a ~60ms timeout inside a hang guard: without a check in the loop the call effectively never returns and the guard fails the test.

---

## Function 3 · `CallWithTimeout(d time.Duration, op func(ctx context.Context) (string, error)) (string, error)`

So far the tests built every context; this one makes *you* the caller.

1. Create `ctx, cancel := context.WithTimeout(context.Background(), d)`.
2. `defer cancel()`.
3. Call `op(ctx)` and return its two results **verbatim** — value and error, whatever they are, including `op`'s own non-context errors.

**Pinned contract:**

- The context handed to `op` must be live on entry (for `d > 0`) and must carry a deadline — the tests pass an `op` that inspects `ctx.Err()` and `ctx.Deadline()` and fails if you handed it a bare `Background()`.
- `d <= 0` → `WithTimeout` produces an already-expired context; a ctx-aware `op` then returns immediately with `context.DeadlineExceeded`, and so must you.
- No result rewriting: `op` returning `("ok", nil)` comes back as `("ok", nil)`; `op` returning a custom error comes back as that error.

**Traced example:**

```
CallWithTimeout(60*time.Millisecond, slowCtxAwareOp)   // op needs 2s

  wrapper: ctx deadline = now+60ms, defer cancel()
  op:      select work-timer | <-ctx.Done()
  t=60ms   deadline → op returns ("", context.DeadlineExceeded)
  wrapper: returns ("", context.DeadlineExceeded)      ← verbatim, ~60ms elapsed
```

---

## Rules that apply to all three

- **`defer cancel()`** on every context you create, timeout or not — `go vet ./78-context-timeout/` must be clean.
- **`errors.Is`**, never `==` string comparison, to classify the error.
- **No `time.Sleep` in your implementations** (`time.After` inside `SlowOp`'s select is the simulated work, not a sleep).
- Keep everything prompt: every timing test asserts elapsed time well under the durations it would take without early return.
- **`go test -race ./78-context-timeout/`** must be clean.

## Edge cases the tests hit

- Generous timeout vs fast work → `("done", nil)`.
- Short timeout vs slow work → `("", context.DeadlineExceeded)` **and** elapsed < 1s.
- Explicit `cancel()` before the deadline → `context.Canceled`, *not* `DeadlineExceeded`.
- Already-expired and already-cancelled contexts passed in → immediate return, correct sentinel each.
- A context with no deadline at all (`context.Background()`) → work just completes.
- Zero and negative `d` in `CallWithTimeout` → instantly expired context.
- `SumToN` with `n <= 0`, with a small `n`, and with an effectively-infinite `n` under a tight deadline.

## What the tests cover

**`SlowOp`:** success under a generous timeout and under plain `Background()`; the timeout case with both the `errors.Is(err, context.DeadlineExceeded)` and the elapsed-time assertion; explicit-cancel → `context.Canceled`; pre-expired and pre-cancelled contexts; a repeated-cycles determinism loop.

**`SumToN`:** exact sum for small `n`; zero/negative `n`; the hang-guarded huge-`n` timeout case proving the loop checks its context; pre-cancelled context returning promptly.

**`CallWithTimeout`:** verbatim result passthrough (success and custom error); the deadline actually set and the context live on entry; short-timeout expiry with elapsed-time proof; `d = 0` expiring instantly.
