# 20 · Binary search

**Goal:** The classic O(log n) search on a sorted slice — the single most drilled algorithm in interviews. Get the loop invariant into your fingers.

**Signature:**
```go
func BinarySearch(s []int, target int) (int, bool)
```

**Requirements:**
- Input is assumed sorted ascending (document that precondition).
- Found → `(index, true)`; not found → `(0 or -1 — pick one and document, false)`.
- Iterative loop (no recursion); midpoint computed as `lo + (hi-lo)/2`.
- With duplicates, returning *any* matching index is acceptable — document it.

**Examples:** `[1,3,5,7,9]`: target 5 → `(2, true)`; 1 → `(0, true)`; 9 → `(4, true)`; 4 → `(_, false)`; 0 → `(_, false)`; 10 → `(_, false)`.

**Edge cases:** empty slice; single element (hit and miss); target below the first element; above the last; first/last positions (off-by-one territory); duplicates.

**Test plan:** table test with every case above. Then an oracle loop: for a fixed sorted slice, for every value in a range (say -1..12), compare your result against `sort.SearchInts`-derived truth. The oracle catches off-by-ones your hand-picked cases miss.

**Done when:** the oracle loop agrees everywhere, and you can state why `lo + (hi-lo)/2` is preferred over `(lo+hi)/2` (overflow in fixed-width languages) even though Go ints make it mostly theoretical.

