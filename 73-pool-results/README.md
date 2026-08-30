# 73 · Worker pool with a results channel

## The idea in one paragraph

In #72 the results came back through a slice or a mutex-guarded map. Here they come back through a **second channel**: workers read from `jobs` and write to `results`, and the caller ranges over `results` to collect them. That sounds like a small change, and it introduces the most common deadlock in Go.

The deadlock is a genuine chicken-and-egg problem, so it's worth stating up front:

- To `range` over `results`, something must **close** it — otherwise the range never ends.
- You can't close it before the workers finish — they'd panic sending on a closed channel.
- You can't `wg.Wait()` for the workers *before* you start ranging — with an unbuffered `results`, the workers are blocked trying to send, so they never finish, so `Wait` never returns, so you never start receiving. **Everything stops.**

The way out is the point of this challenge, and you'll build it in `PoolStream`.

---

## The functions

| function | what it teaches |
|---|---|
| `PoolCounted` | collect results by **counting** — you know how many to expect |
| `PoolStream` | collect results by **ranging** — the closer-goroutine pattern |
| `PoolResultsWithErrors` | carry **success and failure** through one channel |

---

## Function 1 · `PoolCounted(jobs []int, workers int, process func(int) int) []int`

**What it gets:** jobs, worker count, and a transform to apply.

**What it must do:**

1. Make `jobs` and `results` channels. Start N workers; each ranges `jobs`, calls `process(j)`, and sends the answer into `results`.
2. Feed all jobs, then `close(jobs)` — the workers' range loops end when the jobs run out.
3. In the caller, receive **exactly `len(jobs)` times** from `results`, using a plain counted loop.
4. Return everything you received.

**What it must return:** a slice with `len(jobs)` values, containing exactly the `process` outputs — but in **arrival order**, which is unpredictable. The tests therefore sort before comparing.

**Why this version needs no closer goroutine:** you already know how many results are coming, so you never need `results` to close. You stop because you counted, not because the channel ended. That's simpler — and it's only possible when the count is known in advance.

**The buffering choice:** if `results` is unbuffered, each worker blocks on its send until the caller receives it — which is fine here, because the caller is actively receiving. If you make it `make(chan int, len(jobs))`, workers never block at all and finish immediately. Both work; note in a comment which you chose and why.

**Traced example:**

```
PoolCounted([]int{1,2,3,4}, 2, func(x int) int { return x * 10 })

  jobs:    1 2 3 4  →  close
  worker 0 takes 1 → sends 10
  worker 1 takes 2 → sends 20
  worker 1 takes 4 → sends 40      ← finished before worker 0's next job
  worker 0 takes 3 → sends 30

  caller receives 4 times: 10, 20, 40, 30
  → returns [10 20 40 30]   (arrival order — sorted: [10 20 30 40])
```

---

## Function 2 · `PoolStream(jobs []int, workers int, process func(int) int) []int`

**What it gets:** the same three arguments. **Same output contract** as `PoolCounted` (all results, arrival order).

**What it must do — this is the drill:**

1. Make both channels. Start N workers as before: range `jobs`, send into `results`. Track them with a `sync.WaitGroup`.
2. Feed all jobs, `close(jobs)`.
3. **Start one extra goroutine whose entire job is:** wait for the WaitGroup, then close `results`.
   ```
   go func() { wg.Wait(); close(results) }()
   ```
4. In the caller, `for v := range results { ... }` — collect until the channel closes.
5. Return what you collected.

**Why step 3 must be a goroutine.** This is the whole lesson. Trace the broken version:

```
   feed jobs, close(jobs)
   wg.Wait()                 ← caller blocks here
                                but workers are blocked sending into results,
                                because nobody is receiving yet
   for v := range results    ← never reached
   ────────────────────────────────────────────
   DEADLOCK: caller waits for workers, workers wait for caller
```

Putting `wg.Wait()` in its own goroutine breaks the cycle: the caller goes straight to the `range` and starts draining, the workers unblock and finish, the WaitGroup releases, the closer closes `results`, and the range terminates on its own.

**Get it wrong once on purpose.** Write the inline-`Wait` version, run it, and read `fatal error: all goroutines are asleep - deadlock!`. Then fix it. That five minutes is worth more than the paragraph above.

**Traced example (the working version):**

```
  main:    feeds 1..4, closes jobs, then immediately ranges results
  workers: pull jobs, send results (main is receiving, so sends succeed)
  closer:  wg.Wait() unblocks once both workers return  →  close(results)
  main:    range ends  →  returns [10 40 20 30]
```

---

## Function 3 · `PoolResultsWithErrors(jobs []int, workers int, process func(int) (int, error)) ([]int, []error)`

**What it gets:** a `process` that can **fail**.

**What it must do:**

1. Define a result struct carrying both outcomes, e.g. `type result struct { value int; err error }`, and make `results` a channel of that type.
2. Workers call `process(j)` and send **one** result struct either way — success or failure. A failed job still produces exactly one message; nothing is silently dropped.
3. Use the closer-goroutine pattern from `PoolStream` to range over `results`.
4. Split as you collect: successes into the values slice, failures into the errors slice.
5. Return both slices.

**What it must return:** `len(values) + len(errors) == len(jobs)`, always. Both slices non-nil even when empty.

**Why one channel and not two:** two channels (`results` and `errs`) means two closes, two ranges, and a `select` that has to know when both are done. One channel carrying a sum type keeps the shape identical to `PoolStream` — the struct absorbs the complexity instead of the control flow. This is how `errgroup`-style code and #97's URL fetcher are built.

**Traced example:**

```
PoolResultsWithErrors([]int{1,2,3,4,5}, 2, half)
    where half(x) = (x/2, nil) if x is even, else (0, error)

  job 1 → {0, err}       job 2 → {1, nil}
  job 3 → {0, err}       job 4 → {2, nil}
  job 5 → {0, err}

  → values [1 2]   (sorted; 2 successes)
    errors [3 errors]
    2 + 3 == 5 jobs ✓
```

---

## Rules that apply to all three

- **`workers <= 0`** → treat as 1. **`workers > len(jobs)`** → fine, extras exit immediately.
- **Empty or nil jobs** → empty non-nil results, prompt return, no hang. (You still must close `jobs`.)
- **`process` is called exactly once per job.**
- Results arrive in **unspecified order** — the tests sort before comparing and never assert an interleaving.
- **Only one goroutine closes each channel:** the feeder closes `jobs`; the dedicated closer closes `results`. A worker must never close either.
- **No `time.Sleep`** in your implementations.
- **`go test -race ./73-pool-results/`** must be clean.

---

## What the tests cover

**Both pool variants**, run through the same table: several sizes and worker counts; 1 worker; more workers than jobs; 0 and negative workers; empty and nil jobs; a single job; duplicate job values; and a 1,000-job run. Results are sorted before comparison, and every call is timeout-guarded — a deadlock fails the test in seconds instead of hanging the suite forever.

**`PoolStream` specifically:** a 50-run repetition asserting the same multiset each time, and an unbuffered-`results` case, which is exactly where the inline-`Wait` deadlock manifests.

**`PoolResultsWithErrors`:** all-success, all-failure, and mixed inputs; the `values + errors == len(jobs)` accounting property on every case; the specific errors preserved (not just counted); and empty input yielding two empty non-nil slices.

**Call counting:** `process` invoked exactly once per job, verified through a buffered channel sized to catch duplicates.
