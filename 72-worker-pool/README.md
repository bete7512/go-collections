# 72 · Worker pool

## The idea in one paragraph

A worker pool is **N goroutines all receiving from the same channel**. You put jobs into that channel; the runtime hands each job to whichever worker happens to be free. There is no dispatcher, no round-robin logic, no manual splitting of the slice — the channel *is* the distribution mechanism, because a value sent on a channel is delivered to exactly one receiver.

You will write three functions. They are the same pool three times over, each one asking for a bit more:

| function | what it answers |
|---|---|
| `RunPool` | "did every job get processed?" |
| `PoolWithWorkerIDs` | "**which worker** processed each job?" |
| `ProcessAll` | "give me the **results**, in the original order" |

---

## The five steps every one of them follows

Whatever you're building, the pool skeleton is the same. Learn these five steps as a unit:

1. **Make one jobs channel.** All workers will receive from this single channel.
2. **Start N worker goroutines.** Each one loops with `for j := range jobs { ... }`. That loop blocks when the channel is empty and ends **only when the channel is closed**.
3. **Feed the jobs in.** Loop over your input slice, sending each item into the channel.
4. **Close the jobs channel.** This is the signal "no more jobs are coming". It's what makes every worker's `range` loop finish and the goroutine exit.
5. **Wait.** `wg.Wait()` blocks until all N workers have returned. After it, all work is done.

**If you forget step 4**, every worker sits parked forever waiting for a job that never arrives, `Wait` never returns, and the test hangs. That's the single most common mistake in this pattern, and it's why every test in this challenge is wrapped in a timeout that fails with exactly that diagnosis.

**Never close the jobs channel from inside a worker.** Only the feeder (step 4) closes it. N workers closing the same channel means N−1 panics.

---

## Function 1 · `RunPool(jobs []int, workers int) int`

**What it gets:** a slice of job values, and how many workers to run.

**What it must do:**

1. Run the five-step skeleton above.
2. Each worker, for every job it receives, counts that job as processed. (The "work" itself is irrelevant here — the point is the counting.)
3. After `Wait`, return the **total number of jobs processed across all workers**.

**What it must return:** exactly `len(jobs)`. Not more (a job handled twice), not fewer (a job dropped or a worker still running when you counted).

**The tricky part:** N goroutines all incrementing one counter is a data race. Two options, both acceptable:
- guard the counter with a `sync.Mutex`, or
- give each worker its own local count and combine them afterwards (via a results channel, or a per-worker slot in a slice).

A plain `count++` from multiple goroutines will be caught by `go test -race`.

**Traced example:**

```
RunPool([]int{1,2,3,4,5,6,7,8,9}, 3)

  jobs channel gets 1..9, then closes
  worker 0 happens to take: 1, 4, 5, 9   → 4 jobs
  worker 1 happens to take: 2, 6         → 2 jobs
  worker 2 happens to take: 3, 7, 8      → 3 jobs
                                    total 9

  → returns 9
```

The split between workers changes every run. The **total never does**.

---

## Function 2 · `PoolWithWorkerIDs(jobs []int, workers int) map[int]int`

**What it gets:** the same two arguments.

**What it must do:**

1. Same five-step skeleton, but now **give each worker an ID** when you start it — worker 0, worker 1, … worker N−1. (Pass the loop variable into the goroutine.)
2. When a worker handles job value `j`, record `result[j] = itsOwnID`.
3. Return that map after `Wait`.

**What it must return:** a map with **one entry per job**, mapping the job's value to the ID of the worker that handled it. The test uses distinct job values, so the map size must equal `len(jobs)`.

**The tricky part:** a map written by several goroutines at once is a *fatal* runtime error in Go, not just a race — `concurrent map writes` kills the program. So either:
- protect the map with a mutex, or
- have workers send `{job, workerID}` pairs down a results channel and build the map in one goroutine afterwards.

**Traced example:**

```
PoolWithWorkerIDs([]int{10,20,30,40,50,60,70,80,90}, 3)

  → map[10:0  20:1  30:2  40:0  50:0  60:1  70:2  80:2  90:0]
        │     │     │
        │     │     └── job 30 was handled by worker 2
        │     └──────── job 20 was handled by worker 1
        └────────────── job 10 was handled by worker 0

  9 entries, one per job. WHICH worker got which job differs every run.
```

**What the test checks:** that the map has one entry per job (nothing dropped or double-handled), and — on a 1,000-job run with 4 workers — that **at least two different worker IDs** appear. It never checks a specific split, because the distribution genuinely isn't fair: a worker that finishes fast comes back for more and takes a bigger share. That's a feature.

---

## Function 3 · `ProcessAll(jobs []int, workers int, process func(int) int) []int`

**What it gets:** the jobs, the worker count, and a **function to apply to each job**.

**What it must do:**

1. Same skeleton, but each worker calls `process(job)` and the result has to come back to you.
2. Return all results **in the same order as the input slice** — `results[i]` must be `process(jobs[i])`.

**What it must return:** a slice of length `len(jobs)`, in input order. Empty (but non-nil) for empty input.

**The tricky part — and the real lesson of this function:** jobs finish in unpredictable order. Job 4 may complete before job 1. If you append results as they arrive, the output is scrambled.

The clean solution: **carry the index along with the job**. Send a small struct like `{index, value}` through the jobs channel instead of a bare int. Pre-allocate `results := make([]int, len(jobs))` and have each worker write to `results[itsIndex]`. Because every worker writes to a *different* index, no mutex is needed — separate memory locations, no race. Then return `results`.

**Traced example:**

```
ProcessAll([]int{1,2,3,4}, 3, func(x int) int { return x * 10 })

  jobs sent as {index:0,val:1} {index:1,val:2} {index:2,val:3} {index:3,val:4}

  completion order (unpredictable): index 3 finishes, then 1, then 0, then 2
  each worker writes into its own slot:

      results[3] = 40      ← written first
      results[1] = 20
      results[0] = 10
      results[2] = 30      ← written last

  → returns [10 20 30 40]     ← input order, restored for free
```

One test deliberately makes larger values take longer, so completion order is the exact **reverse** of input order — and the result must still come back in input order.

---

## Rules that apply to all three

- **`workers <= 0`** → treat as 1 worker. Don't error, don't panic.
- **`workers > len(jobs)`** → fine. The extra workers get nothing, hit the closed channel, and exit. Must not hang.
- **Empty or nil job slice** → return 0 / an empty map / an empty non-nil slice. Note that you still have to close the channel, or the workers hang even with no work.
- **Duplicate job values** are counted individually — `[7,7,7,7]` is four jobs, not one.
- **No `time.Sleep` in your implementation.** Synchronization comes from the channel and the WaitGroup.
- **`go test -race ./72-worker-pool/`** must be clean. This challenge is specifically about shared state across goroutines; the detector is the point.

---

## What the tests cover

**RunPool:** 9 jobs / 3 workers; single job; 100 workers with 3 jobs; 1 worker; 0 and negative workers; duplicate values; empty and nil slices; and 50 repeated runs of a 200-job pool asserting the same total every time (a varying total means a dropped or duplicated job).

**PoolWithWorkerIDs:** one map entry per job; at least two workers participating across 1,000 jobs.

**ProcessAll:** input order preserved for several transforms and sizes; an identity function; a negating function; a transform whose duration varies by input so completion order is reversed; 1,000 jobs; and a check that `process` is invoked **exactly once per job** (counted through a buffered channel large enough to catch duplicates).

**Everything** is wrapped in a timeout guard whose failure message names the likely cause: a missing `close(jobs)`.
