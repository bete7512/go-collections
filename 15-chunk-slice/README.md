# 15 · Chunk a slice

**Goal:** Break a slice into consecutive sub-slices of size n (last one may be shorter). Drills sub-slicing bounds and forces a decision about sharing vs copying backing arrays.

**Signature:**
```go
func Chunk(s []int, n int) ([][]int, error)   // error on n <= 0
```

**Requirements:**
- Every chunk except possibly the last has exactly n elements; concatenating all chunks reproduces the input in order.
- `n <= 0` is invalid input — return an error (or panic; document).
- Decide: chunks share the input's backing array (`s[i:j]` — cheap, but writes to a chunk show through to the input) or are independent copies. Put your choice in a doc comment.

**Examples:** `[1..7]`, n=3 → `[[1,2,3],[4,5,6],[7]]`; `[1..6]`, n=3 → `[[1,2,3],[4,5,6]]`; `[1,2]`, n=5 → `[[1,2]]`; `[]`, n=3 → `[]` (empty, no chunks); n=0 → error.

**Edge cases:** even split; ragged last chunk; n=1 (each element its own chunk); n >= len; empty input; n<=0.

**Test plan:** table test of the cases above with deep equality; one test asserting the flatten-back property; if you chose sharing, a test that mutates a chunk and documents the effect on the input.

**Done when:** all pass and the sharing-vs-copy decision is written down and tested, not accidental.

