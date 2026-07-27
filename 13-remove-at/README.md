# 13 · Remove element by index, preserving order

**Goal:** Delete position i from a slice, keeping the relative order of everything else. Drills `append` mechanics and teaches slice aliasing the hard way.

**Signature:**
```go
func RemoveAt(s []int, i int) ([]int, error)   // error (or panic — document) on bad index
```

**Requirements:**
- Result has length `len(s)-1`; elements before i unchanged; elements after i shifted left by one.
- Out-of-range i (negative or >= len) must be handled deliberately: return an error or panic — write down which and why.
- Know (and demonstrate in a test) that the idiomatic `append(s[:i], s[i+1:]...)` **mutates the input's backing array** — the original slice variable still sees shifted data.

**Examples:** `[1,2,3,4]`, i=1 → `[1,3,4]`; i=0 → `[2,3,4]`; i=3 → `[1,2,3]`; i=9 → error.

**Edge cases:** first index, last index, only element (result empty), out-of-range, negative index.

**Test plan:** table test for first/middle/last/single/out-of-range. Plus one "aliasing surprise" test: keep the original slice variable, call RemoveAt, and assert what the original now contains — encode the surprising truth, don't hide it.

**Done when:** all pass and you can explain in one sentence why the original slice changed.

