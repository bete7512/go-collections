# 11 · Reverse a slice in place

**Goal:** Mutate the caller's slice so its elements end up in reverse order — no new slice allocated. This drills the difference between mutating shared backing memory and returning a copy.

**Signature:**
```go
func ReverseInPlace(s []int)
```

**Requirements:**
- No return value — the caller's slice must be changed after the call (that's the proof it's in place).
- Swap symmetric pairs from the two ends toward the middle; stop when the indices meet.
- Zero allocations: no `make`, no `append`, no second slice.

**Examples:** `[1,2,3,4]` → `[4,3,2,1]` (even length); `[1,2,3]` → `[3,2,1]` (odd — middle element stays); `[7]` → `[7]`; `[]` → `[]`.

**Edge cases:**
- Empty slice and nil slice — must not panic, simply nothing happens.
- Single element — loop body never runs.
- Even vs odd length — off-by-one in the loop bound shows up here.

**Test plan:** table test with even, odd, single, empty inputs; call the function, then assert on the *same variable you passed in* (not a return value). Add one test that keeps a second reference to the same backing array and asserts it also changed — that's the aliasing insight.

**Done when:** all cases pass, and the implementation is a two-index swap loop with no allocation.

