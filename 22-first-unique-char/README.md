# 22 · First non-repeating character

**Goal:** Find the first rune that appears exactly once. Teaches why you need two passes: map iteration order in Go is deliberately randomized, so the map alone can't tell you "first".

**Signature:**
```go
func FirstUnique(s string) (rune, bool)
```

**Requirements:**
- Pass 1: count occurrences per rune. Pass 2: walk the *string* again in order, return the first rune whose count is 1.
- Return `(0, false)` when every rune repeats or the string is empty.
- Operate on runes, not bytes — multi-byte input must work.

**Examples:** `"swiss"` → `('w', true)`; `"aabb"` → `(0, false)`; `"abc"` → `('a', true)`; `"ééa"` → `('a', true)`; `""` → `(0, false)`.

**Edge cases:** empty string; all repeating; all unique (first char wins); multi-byte runes; case sensitivity (`"aA"` — decide whether these are the same character; document).

**Test plan:** table test with the examples above. Add one test that runs the function 100 times on the same input and asserts an identical answer each time — that catches any accidental reliance on map ordering.

**Done when:** the repeat-100-times test is stable and you can explain why iterating the map instead of the string would be a bug.

