# 32 · Two-sum

## Problem

Given a slice of ints and a target, find two **different positions** whose values add up to the target, and return their indices. This is the archetypal "trade memory for speed" problem: the obvious solution checks every pair in O(n²); one map turns it into a single O(n) pass. The map-based version is the drill — the nested loop teaches nothing.

## Contract (what the tests enforce)

```go
func TwoSum(nums []int, target int) (int, int, bool)
```

- **Return `(i, j, true)` with `i < j`** such that `nums[i] + nums[j] == target`.
- **Two different positions, not two different values.** `[3,3]` with target 6 → `(0, 1, true)`: same value, two indices — valid. `[5]` with target 10 → `false`: you may not use index 0 twice. This is the case that catches implementations that look up the complement before excluding the current position.
- **No answer → `(_, _, false)`.** The two index values are meaningless when the bool is false; the tests only check the bool.
- **When several valid pairs exist, any one of them is accepted.** The tests verify validity (`i < j`, indices in range, values summing to target) rather than pinning one specific pair.
- Negative numbers, zeros, and negative targets are all legitimate.
- The input slice is not modified.
- Expected shape: one pass, one `map[int]int` from value → index, checking for the complement among *already-seen* elements before storing the current one. That ordering is what makes same-index reuse impossible by construction.

## Worked examples

**Example 1 — basic:**

```
nums:   [2, 7, 11, 15], target: 9
answer: (0, 1, true)          // 2 + 7 = 9
```

**Example 2 — same value at two indices:**

```
nums:   [3, 3], target: 6
answer: (0, 1, true)          // two distinct positions, same value — allowed
```

**Example 3 — the same-index trap:**

```
nums:   [5], target: 10
answer: (_, _, false)         // 5 + 5 = 10, but there is only ONE index 0
```

An implementation that inserts `nums[i]` into the map *before* looking up the complement will wrongly answer `(0, 0, true)` here.

**Example 4 — negatives and zero target:**

```
nums:   [-3, 4, 3, 90], target: 0
answer: (0, 2, true)          // -3 + 3 = 0
```

## Edge cases the tests cover

- Empty slice and single-element slice → false.
- The answer being the first two elements, the last two elements, and first+last (index extremes).
- Duplicate values forming the answer (`[3,3]`).
- The half-of-target single occurrence (`[5]`, target 10 → false; `[5,5]` → true).
- Negative numbers, negative target, zero target.
- Values summing with zeros (`[0,0]`, target 0 → true).
- Multiple valid pairs — validity-checked, not pinned.
- No answer among many elements → false.
- A 100,000-element input with the answer split between the two ends — an O(n²) solution will be visibly slow here, an O(n) one instant.
- Input-unmodified snapshot on every case.
