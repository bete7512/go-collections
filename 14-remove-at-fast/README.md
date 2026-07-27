# 14 · Remove by index, swap-and-truncate

**Goal:** Same deletion, but O(1): move the last element into the hole and shrink by one. Order is sacrificed for speed.

**Signature:**
```go
func RemoveAtFast(s []int, i int) ([]int, error)
```

**Requirements:**
- Copy `s[len(s)-1]` into `s[i]`, then truncate the last slot.
- Constant time — no shifting loop, no allocation.
- Same out-of-range policy as #13 (be consistent).

**Examples:** `[1,2,3,4]`, i=1 → `[1,4,3]`; i=3 (last) → `[1,2,3]` (swap with itself); `[9]`, i=0 → `[]`.

**Edge cases:** removing the last index; single-element slice; out-of-range.

**Test plan:** assert the new length, assert the removed value is absent, assert every *other* original value is still present — but do **not** assert order (write the test with a sort-then-compare or a count map).

**Done when:** tests pass, and you can name a real situation where you'd pick this over #13 (e.g. a free-list / worker pool where order is irrelevant) and one where you must not (ordered event log).

