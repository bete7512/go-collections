# 79 · Mutex-protected counter, verified with -race

## The idea in one paragraph

`n++` looks like one operation; the machine runs three — load, add, store. Two goroutines interleaving those three steps silently lose increments: both load 41, both store 42, one increment vanishes. That's a **data race**: two goroutines touching the same memory concurrently, at least one writing, with no synchronization between them. Go's answer here is `sync.Mutex` — wrap the load-add-store in a **critical section** so only one goroutine is ever inside it — and Go's superpower is the **race detector**: `go test -race` instruments every memory access and prints the two conflicting accesses, both stack traces, and where each goroutine was created. This challenge makes you cause a race on purpose, read one full report end to end, then fix it. Every Go engineer should have done that deliberately at least once.

## Vocabulary

| term | meaning |
|---|---|
| data race | concurrent unsynchronized access to the same memory, ≥1 write; undefined behavior in Go, not "just a wrong number" |
| critical section | the code between `Lock()` and `Unlock()` — at most one goroutine inside at a time |
| `sync.Mutex` | mutual exclusion lock; the **zero value is ready to use** — no constructor |
| race detector | `-race` flag; reports races that actually occur during the run (it proves presence, never absence) |
| `sync/atomic` | lock-free primitives (`atomic.Int64`) for single-word operations — the right tool for a plain counter |
| `sync.RWMutex` | many concurrent readers **or** one writer — worth it when reads vastly outnumber writes |

---

## Step 1 — cause the race on purpose (do this first)

Write `UnsafeCounter` (below) with no synchronization at all, then run the deliberately racy demo:

```
RACE_DEMO=1 go test -race -run UnsafeCounterRaceDemonstration ./79-mutex-counter/
```

The demo test is **skipped by default** so the graded suite stays green; the env var opts in. Read the whole report: the `WARNING: DATA RACE` header, the *read/write at 0x…* pair with both stack traces, and the `created at` lines showing which `go` statements spawned the goroutines. Also notice the final count it logs — usually well under 10000, and different every run. Only then build `SafeCounter`.

---

## The types

| type | what it teaches |
|---|---|
| `UnsafeCounter` | the race, kept on display |
| `SafeCounter` | the fix: mutex-guarded writes **and reads** |
| `AtomicCounter` | the lock-free alternative, benchmarked against the mutex |

All methods on **pointer receivers**. A `sync.Mutex` must never be copied — passing a `SafeCounter` by value copies the lock and `go vet` (copylocks) flags it. `go vet ./79-mutex-counter/` must be clean.

---

## Type 1 · `UnsafeCounter`

```go
type UnsafeCounter struct { /* just the int */ }

func (c *UnsafeCounter) Inc()
func (c *UnsafeCounter) Value() int
```

No mutex, no atomic — plain `n++` and plain read. **Pinned contract:** correct under single-goroutine use (`Inc` five times → `Value() == 5`); under concurrent use it loses updates, which is the point. The graded tests only ever touch it from one goroutine; the opt-in demo hammers it from 100.

---

## Type 2 · `SafeCounter`

```go
type SafeCounter struct {
    mu sync.Mutex
    n  int
}

func (c *SafeCounter) Inc()
func (c *SafeCounter) Value() int
```

**Pinned contract:**

- Zero value ready: `var c SafeCounter` works immediately — `Value() == 0`, no constructor exists.
- `Inc` adds exactly 1; 100 goroutines × 100 increments → `Value() == 10000`, **every single run** — the tests run the full hammer three times and re-check.
- **`Value()` takes the lock too.** A read concurrent with a write is a data race even though the reader changes nothing — this is the half people skip, and the concurrent read/write test exists precisely to make `-race` catch an unlocked reader.
- `defer c.mu.Unlock()` immediately after `Lock()` — the unlock survives any future early return or panic.
- In a comment, name the alternatives and when you'd pick them: `atomic.Int64` for a plain counter (faster, no lock), `sync.RWMutex` when reads vastly outnumber writes.

**Traced example — why `n++` loses updates without the lock:**

```
        goroutine A            goroutine B          n
        load  n → 41                                41
                               load  n → 41         41
        add   → 42                                  41
        store 42                                    42
                               add   → 42
                               store 42             42   ← two Incs, one increment
```

---

## Type 3 · `AtomicCounter`

```go
type AtomicCounter struct { /* atomic.Int64 */ }

func (c *AtomicCounter) Inc()
func (c *AtomicCounter) Value() int64
```

**Pinned contract:** same behavior as `SafeCounter` — zero value ready, exact 10000 under the 100×100 hammer — implemented with `sync/atomic` (`atomic.Int64`), no mutex anywhere in the type. Note `Value()` returns `int64`, matching the atomic's word.

The file also ships two benchmarks (`BenchmarkSafeCounterInc`, `BenchmarkAtomicCounterInc`, both via `b.RunParallel`). Run them and look at the gap:

```
go test -bench=. -benchmem ./79-mutex-counter/
```

---

## Edge cases the tests hit

- The unprotected version's nondeterministic result (opt-in demo — observed and logged, never asserted exact).
- Forgetting the lock in `Value()` — the concurrent read/write test makes the detector report it.
- Copying the struct by value — `go vet` copylocks (your code just has to not do it).
- Zero values of all three types being immediately usable — no constructors anywhere.
- Sequential correctness before any concurrency: N increments → exactly N.

## What the tests cover

**`UnsafeCounter`:** single-goroutine correctness; the `RACE_DEMO=1`-gated 100×100 hammer that produces a real race report under `-race` and logs the lost-update count.

**`SafeCounter`:** zero-value readiness; sequential exactness; the 100 × 100 → 10000 assertion repeated three full runs; a simultaneous `Inc`/`Value` hammer (50 writers, 50 readers) asserting every intermediate read is in range and the final value exact — run everything with `-race`.

**`AtomicCounter`:** the same 10000 hammer and zero-value check, plus the two throughput benchmarks.

**Done when:** `go test -race ./79-mutex-counter/` is silent and green, the 10000 assertion passes, and you have read one complete race report end to end via the demo.
