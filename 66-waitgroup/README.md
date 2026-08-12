# 66 · Goroutines and sync.WaitGroup

## Problem

Tier 5 opens. A goroutine is cheap — a few KB of stack, multiplexed onto OS threads by the runtime — so starting a hundred is unremarkable. The hard part is never starting them; it's **knowing when they've finished**. `main` returning kills every goroutine mid-flight, so without a join mechanism, concurrent work is work you can't rely on.

`sync.WaitGroup` is the join primitive: a counter you raise before launching and lower as each finishes, with `Wait` blocking until it hits zero. The four-line skeleton is the muscle memory this drill installs — and the ordering rules inside it exist because getting them wrong produces races that pass a hundred test runs and fail in production.

## Contract (what the tests enforce)

```go
func RunWorkers(n int) []int                    // returns each worker's ID, any order
func SumParallel(nums []int, workers int) int   // splits the work, sums the parts
```

- **`RunWorkers(n)`** launches `n` goroutines, each of which produces its own ID (0…n−1), waits for all of them, and returns the collected IDs. Order is **unspecified** — the tests sort before comparing, because scheduling order is the runtime's business.
  - `n == 0` returns an empty (non-nil) slice; `Wait` on a zero counter returns immediately.
  - Negative `n` is treated as 0.
  - Collect results without a data race: a buffered channel of capacity `n` (each goroutine sends, then you drain after `Wait`), or a pre-sized slice where goroutine *i* writes only `results[i]` — distinct indices need no mutex. Both are fine; a shared `append` from many goroutines is not.
- **`SumParallel(nums, workers)`** splits `nums` into roughly equal contiguous chunks, sums each in its own goroutine, and returns the total. It must equal the sequential sum for any split.
  - `workers <= 0` is treated as 1; `workers > len(nums)` is harmless (some goroutines get nothing).
  - Empty/nil input → 0.
- **The ordering rules, which are the point:**
  - `wg.Add(1)` goes **before** the `go` statement, in the parent. Calling `Add` inside the goroutine races with `Wait` — the goroutine may not have been scheduled yet when `Wait` checks the counter, so `Wait` returns early.
  - `defer wg.Done()` as the goroutine's **first line**, so it fires on every exit path including a panic.
  - Pass the WaitGroup as `*sync.WaitGroup` if it crosses a function boundary. Copying one breaks it — `go vet`'s copylocks check exists for this.
- **Loop variables:** Go 1.22+ gives each iteration its own variable, so capturing `i` in the closure is safe on a modern toolchain. Passing it as an argument is still clearer and version-proof. Know what the pre-1.22 bug looked like (every goroutine seeing the final value) — it's in a comment, not in the tests.
- **The whole suite runs under `-race`.** That's the tool this tier is really about: run `go test -race ./66-waitgroup/`.

## Worked examples

**Example 1 — the skeleton, which should become automatic:**

```go
var wg sync.WaitGroup
for i := 0; i < n; i++ {
    wg.Add(1)              // BEFORE the go statement
    go func(id int) {
        defer wg.Done()    // FIRST line inside
        // work
    }(i)                   // passed explicitly
}
wg.Wait()                  // blocks until the counter reaches zero
```

**Example 2 — collecting results safely:**

```go
RunWorkers(5)   // → some permutation of [0 1 2 3 4]; sorted: [0 1 2 3 4]
```

Five goroutines, five IDs, every time — but `[3 0 4 1 2]` is as correct as `[0 1 2 3 4]`.

**Example 3 — splitting work:**

```go
SumParallel([]int{1,2,3,4,5,6,7,8}, 3)   // → 36, same as summing sequentially
SumParallel(nums, 1)                     // → 36, the degenerate single-worker case
SumParallel(nums, 100)                   // → 36, more workers than elements
```

## Edge cases the tests cover

- `RunWorkers` for n = 0, 1, 5, 100 — count and (sorted) contents exact.
- `n = 0` returning an empty non-nil slice and completing immediately.
- Negative n.
- Repeated runs (50 iterations of `RunWorkers(10)`) — every run must produce the complete set, which is what catches an `Add`-inside-the-goroutine race.
- All goroutines actually finishing before the function returns: a shared atomic counter incremented by each worker must equal n immediately after `Wait`.
- `SumParallel` against a sequential sum for several sizes and worker counts, including workers = 1, workers > len, empty and nil input, negative numbers, and a 100,000-element slice.
- A test that runs `SumParallel` 100 times on the same input asserting an identical total each time (nondeterministic sums mean a lost update).
- Everything under `-race`.
