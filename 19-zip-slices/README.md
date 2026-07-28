# 19 · Zip two slices

**Goal:** Pair up two slices element-by-element, stopping at the shorter one. Introduces defining a small struct as a return shape.

**Signature:**
```go
type Pair struct{ A, B int }

func Zip(a, b []int) []Pair
```

**Requirements:**
- `result[i] == Pair{a[i], b[i]}`.
- Length of result is exactly `min(len(a), len(b))` — extra elements of the longer slice are ignored.
- Neither input modified; empty/nil inputs give empty result, no panic.

**Examples:** `Zip([1,2,3], [10,20])` → `[{1,10},{2,20}]`; `Zip([1], [10,20,30])` → `[{1,10}]`; `Zip([1,2], [10,20])` → `[{1,10},{2,20}]`; `Zip([], [1,2])` → `[]`.

**Edge cases:** first shorter, second shorter, equal lengths, one empty, both empty, nil inputs.

**Test plan:** table test of all six shapes above with deep equality on the Pair slice.

**Done when:** the length property `len(Zip(a,b)) == min(len(a),len(b))` holds in every test.

