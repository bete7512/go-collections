# 16 · Flatten [][]int

**Goal:** Concatenate the inner slices of a `[][]int` into one flat `[]int`, in order. The inverse of #15.

**Signature:**
```go
func Flatten(s [][]int) []int
```

**Requirements:**
- Result order: all of `s[0]`, then all of `s[1]`, etc.
- nil inner slices and empty inner slices contribute nothing and must not panic.
- Result must be a fresh slice — mutating it must never affect the input.
- Efficiency habit: first pass to sum lengths, `make` once with that capacity, then fill — no repeated re-allocation.

**Examples:** `[[1,2],[3],[],[4,5]]` → `[1,2,3,4,5]`; `[]` → `[]`; `[[],[],[]]` → `[]`; `[nil,[1]]` → `[1]`.

**Edge cases:** empty outer; all-empty inners; nil inners mixed in; single inner.

**Test plan:** table test of the above; property check: `len(result)` equals the sum of inner lengths; round-trip with #15: `Flatten(Chunk(x, n))` equals `x`.

**Done when:** the round-trip property test passes and the result is pre-sized (one allocation).

