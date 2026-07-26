# 10 · Min, max, sum in one pass

- **Build:** `MinMaxSum(nums []int) (min, max, sum int, err error)` — or `(…, ok bool)`; pick one and justify it.
- **Expected:** `[3,1,4,1,5]` → `1, 5, 14`.
- **Edge cases:** empty slice (this is why the error/ok exists); single element; all-negative values (initializing `max := 0` fails here — classic bug).
- **Test:** normal, single, empty, all-negative.
- **Done when:** exactly one loop over the slice, and the all-negative case passes.
