# 23 · Anagram check

**Goal:** Decide whether two strings contain the same characters with the same multiplicities. Drills the single-map increment/decrement trick.

**Signature:**
```go
func IsAnagram(a, b string) bool
```

**Requirements:**
- Early exit when rune counts differ (cheap and correct — but compare *rune* counts, not `len()` bytes).
- Use one `map[rune]int`: increment for every rune of `a`, decrement for every rune of `b`; anagram iff all counts end at zero.
- Do **not** sort — sorting is O(n log n) and a different drill.
- Case rule: pick case-sensitive or insensitive, document it, test it.

**Examples:** `("listen","silent")` → true; `("hello","world")` → false; `("","")` → true; `("aab","abb")` → false; `("Listen","silent")` → depends on your case rule — test it.

**Edge cases:** different lengths; empty vs empty; empty vs non-empty; unicode (`"héllo"` vs `"olléh"` → true); same letters different counts.

**Test plan:** table test of all the above; one unicode pair; one pair that differs only in count (`"aab"`/`"abb"`) — that's the case a naive "same set of letters" implementation gets wrong.

**Done when:** the `aab`/`abb` case fails correctly and there's exactly one map and no sort.

