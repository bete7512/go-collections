# 30 · Merge two maps, second wins on conflict

## Problem

Combine two maps into one. Every key from either map appears in the result; when both maps carry the same key, the **second map's value wins**. This is the shape of config layering (defaults overridden by user settings), merging label sets, and combining counters snapshots — the override direction is the whole API, so it must be exact.

## Contract (what the tests enforce)

```go
func Merge(a, b map[string]int) map[string]int
```

- **b wins on conflict.** For any key in both maps, the result holds b's value — including when b's value is `0`. A zero value is a real value, not "absent"; an implementation that skips zero values fails.
- **Union of keys.** Keys only in `a` keep a's values; keys only in `b` keep b's values.
- **Fresh result.** The returned map is new memory. Writing into `a` and returning it is the tempting bug — the tests mutate the result afterward and verify both inputs are untouched.
- **Inputs unmodified.** Neither `a` nor `b` may change, ever.
- **nil is a legal input** for either or both arguments. Reading a nil map is safe in Go; writing to one panics — so `Merge(nil, nil)` must return an empty map without panicking, never nil-write.
- The result is always non-nil, even when both inputs are nil or empty.

## Worked examples

**Example 1 — basic merge with a conflict:**

```
a: {"x":1, "y":2}
b: {"y":9, "z":3}
→  {"x":1, "y":9, "z":3}
```

`y` exists in both; b's `9` wins. `x` comes from a alone, `z` from b alone.

**Example 2 — b wins even with a zero value:**

```
a: {"count":5}
b: {"count":0}
→  {"count":0}
```

The override direction doesn't care what the value is. If your implementation checks `if bVal != 0` anywhere, this case catches it.

**Example 3 — nil inputs:**

```
Merge(nil, {"a":1})  → {"a":1}
Merge({"a":1}, nil)  → {"a":1}
Merge(nil, nil)      → {}
```

## Edge cases the tests cover

- Conflict where b's value is zero (b still wins).
- Conflict where the values happen to be equal (result is that value; no double-count).
- All four nil/empty combinations of `a` and `b`.
- `b` empty → result equals `a`, but as a distinct map (mutating the result must not touch `a`).
- Negative values merge like any other.
- Empty-string key `""` is a legitimate key.
- Input-unmodified snapshots of both `a` and `b` after every merge.
- Result independence: adding and changing keys in the result leaves both inputs untouched.
- Complete disjoint merge (no shared keys) — result length is `len(a)+len(b)`.
