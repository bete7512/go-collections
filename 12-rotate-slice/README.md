# 12 · Rotate a slice left by k

**Goal:** Shift every element k positions left, with the first k elements wrapping around to the end. Drills index arithmetic and the modulo-normalization habit.

**Signature (pick one, document your choice):**
```go
func RotateLeft(s []int, k int) []int   // returns a new slice, input untouched
// or
func RotateLeft(s []int, k int)         // rotates in place
```

**Requirements:**
- `k` may be any non-negative int — larger than `len(s)` must work (`k % len(s)`).
- Rotating by 0 or by exactly `len(s)` is a no-op.
- Empty slice must not panic (beware: `k % 0` panics — guard before the modulo).
- If you return a new slice, the input must be unmodified afterward; if in place, no allocation.

**Examples:** `[1,2,3,4,5]`, k=2 → `[3,4,5,1,2]`; k=0 → `[1,2,3,4,5]`; k=5 → `[1,2,3,4,5]`; k=7 → `[3,4,5,1,2]` (same as k=2); `[]`, k=3 → `[]`.

**Edge cases:** k=0; k == len; k > len; k many multiples of len; empty slice; single element with any k.

**Test plan:** table test covering every case above; if you chose the returning version, assert the input slice is byte-for-byte unchanged after the call.

**Done when:** `k=7` on a 5-element slice equals `k=2` and the empty-slice test passes. Bonus (worth it): reimplement in place using the triple-reverse trick (reverse first k, reverse rest, reverse all) and make the same tests pass.

