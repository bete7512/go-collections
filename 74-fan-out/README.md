# 74 · Fan-out

## The idea in one paragraph

**Fan-out** means taking one stream of work and spreading it across several workers so they process it in parallel. Here's the thing you're about to discover: **you already built it in #72**. Multiple goroutines receiving from one channel *is* fan-out — the runtime delivers each value to exactly one receiver, so N workers reading one channel automatically split the stream between them. There is nothing else to write.

This challenge makes that concrete, then contrasts it with the two things people confuse it with:

- **Fan-out ≠ broadcast.** Fan-out: each value goes to *one* worker (work splitting). Broadcast: each value goes to *every* subscriber (event distribution). Same channels, opposite semantics. Broadcast is capstone #100 and needs completely different machinery — you must copy the value to each subscriber yourself.
- **Fan-out ≠ partitioning.** The free version gives you *no control* over which worker gets which value. When you need "all events for user 42 go to the same worker" (to preserve per-user ordering), you need an explicit dispatcher with one channel per worker. That's a different, more expensive pattern — and you'll build it here too, so you know the difference from the inside.

---

## The functions

| function | what it teaches |
|---|---|
| `FanOut` | fan-out is free: N receivers on one channel |
| `FanOutDispatch` | explicit per-worker channels, when you need control |
| `PartitionBy` | deterministic routing: same key → same worker, always |

---

## Function 1 · `FanOut(in <-chan int, workers int) []Result`

```go
type Result struct {
    Value    int  // the input value
    WorkerID int  // which worker handled it
}
```

**What it gets:** an input channel someone else feeds and closes, plus a worker count.

**What it must do:**

1. Start N worker goroutines. **All of them range over the same `in` channel** — that single shared receive is the entire fan-out mechanism.
2. Each worker records `Result{Value: v, WorkerID: itsOwnID}` for every value it receives.
3. Wait for all workers (they exit when `in` closes), then return every Result collected.

**What it must return:** exactly one Result per value that came out of `in`. Order is unspecified.

**What to notice while writing it:** you never decide who gets what. There's no `if`, no counter, no round-robin. You start N receivers and the runtime does the distribution. Write a comment saying so — recognizing that a pattern is already solved is worth as much as implementing one.

**Traced example:**

```
in: 1 2 3 4 5 6, then closed        workers: 3

  worker 0 grabs 1, then 4, then 6
  worker 1 grabs 2
  worker 2 grabs 3, then 5

  → [{1,0} {2,1} {3,2} {4,0} {5,2} {6,0}]   (in some order)

  Every value appears EXACTLY ONCE. The split differs every run —
  a worker that finishes fast comes back and takes more.
```

**Collecting safely:** N goroutines appending to one slice is a race. Use a results channel plus the closer pattern from #73, a mutex, or per-worker slices merged at the end.

---

## Function 2 · `FanOutDispatch(in <-chan int, workers int) []Result`

**Same signature, same output contract, completely different mechanism.**

**What it must do:**

1. Create **N separate channels**, one per worker.
2. Start N workers; worker *i* ranges over **its own** channel `chans[i]` only.
3. Start a **dispatcher** goroutine that ranges over `in` and sends each value into one of the N channels — round-robin is fine (`i % workers`).
4. When `in` closes, the dispatcher must **close all N worker channels** so the workers' range loops end.
5. Wait for the workers, return all Results.

**What it must return:** same as `FanOut` — one Result per input value.

**Why build this at all,** given it's more code for the same result? Because now *you* choose the destination, which is what Function 3 needs. The cost is visible: an extra goroutine, N extra channels, N extra closes, and a dispatcher that can become a bottleneck since every value passes through one goroutine. In the free version, values go straight from producer to worker.

**The trap:** step 4. Closing only `in` leaves the workers ranging over their own channels forever. The dispatcher owns those N channels, so the dispatcher closes them — all of them, after its own range over `in` ends.

**Traced example:**

```
in: 1 2 3 4 5 6, closed             workers: 3, round-robin dispatch

  dispatcher: 1→ch0  2→ch1  3→ch2  4→ch0  5→ch1  6→ch2
              then closes ch0, ch1, ch2

  worker 0 gets exactly 1, 4      worker 1 gets 2, 5      worker 2 gets 3, 6

  → assignment is now DETERMINISTIC, unlike FanOut
```

---

## Function 3 · `PartitionBy(in <-chan string, workers int, key func(string) int) []StringResult`

```go
type StringResult struct {
    Value    string
    WorkerID int
}
```

**What it gets:** a stream of strings, a worker count, and a **key function** returning an integer for each value.

**What it must do:**

1. Same structure as `FanOutDispatch` — N channels, N workers, one dispatcher.
2. The dispatcher routes each value to worker `abs(key(v)) % workers` instead of round-robin.
3. Close all worker channels when `in` closes; collect and return.

**What it must return:** one StringResult per input value, and — the property the tests check — **every value with the same key lands on the same worker, in every run.** Two values with equal keys must never be split across workers.

**Why this matters in real systems:** it's how you keep per-entity ordering under parallelism. All events for one user, one order, one partition key go to a single worker, so that worker sees them in sequence while other workers run concurrently. Kafka consumer groups, sharded queues, and NATS partitioned subjects all work this way. The free fan-out cannot do it — that's precisely what you're paying the dispatcher for.

**Handle a negative key** (a hash can be negative): `key(v) % workers` can be negative and index out of range. Take the absolute value or mask the sign; the tests use a key function that returns negatives.

**Traced example:**

```
in: "user1:a" "user2:x" "user1:b" "user3:p" "user1:c"
key(v) = the numeric part of the user id           workers: 2

  key("user1:a")=1 → worker 1%2 = 1
  key("user2:x")=2 → worker 0
  key("user1:b")=1 → worker 1     ← same key, same worker
  key("user3:p")=3 → worker 1
  key("user1:c")=1 → worker 1     ← same key, same worker

  → all three user1 values landed on worker 1. Guaranteed, every run.
```

---

## Rules that apply to all three

- **`workers <= 0`** → treat as 1. **`workers` greater than the number of values** → fine.
- **A closed-and-empty `in`** → empty non-nil result, prompt return, no hang.
- **Every input value produces exactly one Result** — none dropped, none duplicated.
- **`FanOut`'s worker assignment is never asserted** by the tests, only that ≥2 workers participate on a large stream. `FanOutDispatch` and `PartitionBy` *are* asserted, because they're deterministic by design.
- **Whoever creates a channel closes it:** the caller closes `in`; the dispatcher closes the per-worker channels.
- **No `time.Sleep`.** **`go test -race ./74-fan-out/`** must be clean.

---

## What the tests cover

**`FanOut` and `FanOutDispatch` share a table:** 6 values / 3 workers; single value; 1 worker; more workers than values; 0 and negative workers; duplicate values; empty stream; and 1,000 values / 8 workers. Every input accounted for exactly once in all of them.

**`FanOut` specifically:** ≥2 workers participate across 1,000 values (distribution actually happens), and 20 repeated runs each accounting for every value.

**`FanOutDispatch` specifically:** the round-robin assignment is exact and repeatable — value at index *i* always lands on worker `i % workers`.

**`PartitionBy`:** same-key-same-worker verified by grouping results by key and asserting each group has exactly one distinct worker ID; a negative-key function; all values sharing one key (everything on a single worker); more distinct keys than workers (collisions are expected and fine); and worker IDs always in range.

Everything is timeout-guarded — a missing close on the per-worker channels hangs otherwise, and the failure message says so.
