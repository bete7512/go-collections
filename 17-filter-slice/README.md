# 17 · Filter with a predicate

**Goal:** Keep only the elements a caller-supplied function approves. First taste of higher-order functions in Go — the logic lives in the argument, not in Filter.

**Signature:**
```go
func Filter(s []int, keep func(int) bool) []int
```

**Requirements:**
- Call `keep` exactly once per element, in order; keep those returning true, preserve their relative order.
- Filter itself must contain zero domain logic — no evenness checks, no comparisons.
- Input slice unmodified. Decide nil-vs-empty for the "nothing matched" result and document it (`slices.Equal` treats them equal, `reflect.DeepEqual` does not — this bites people).

**Examples:** `Filter([1,2,3,4,5,6], even)` → `[2,4,6]`; `Filter(same, >10)` → `[]`; `Filter(same, always-true)` → the whole slice.

**Edge cases:** empty input; nothing matches; everything matches; predicate with captured state (a closure counting calls).

**Test plan:** run the same input through at least two different predicates; assert input unmodified; use a counting closure to assert `keep` was called exactly `len(s)` times.

**Done when:** two-predicate test passes and the nil/empty decision is deliberate.

