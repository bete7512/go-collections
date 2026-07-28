# 18 · Map with a transform

**Goal:** Build a new slice by applying a transform func to every element. Companion drill to #17; generics upgrade comes in #61.

**Signature:**
```go
func MapSlice(s []int, f func(int) int) []int
```

**Requirements:**
- Result has exactly `len(s)` elements: `result[i] == f(s[i])`.
- Input untouched; result is fresh memory.
- Pre-size the result (`make([]int, 0, len(s))` + append, or `make([]int, len(s))` + index assign) — no growth reallocations.

**Examples:** `MapSlice([1,2,3], double)` → `[2,4,6]`; `MapSlice([1,2,3], square)` → `[1,4,9]`; `MapSlice([], f)` → `[]`.

**Edge cases:** empty input; nil input; transform that returns the same value (identity — result must still be a *different* backing array).

**Test plan:** two transforms; input-unmodified assertion; identity-transform test asserting `&result[0] != &s[0]` style independence (mutate result, input unchanged).

**Done when:** all pass with a pre-sized result.

