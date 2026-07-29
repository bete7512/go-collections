# 24 · Group strings by first letter

**Goal:** Build a map from first letter to the list of words starting with it. Drills the "map of slices" pattern — appending to a map value that may not exist yet.

**Signature:**
```go
func GroupByFirst(words []string) map[rune][]string
```

**Requirements:**
- `append(m[k], w)` works even when the key is absent (nil slice appends fine) — use that; no pre-initialization needed.
- Within each group, preserve the input's relative order.
- Empty strings in the input: skip them, or bucket under a sentinel rune — document your choice.
- Case: decide whether 'A' and 'a' share a bucket; document.

**Examples:** `["apple","avocado","banana"]` → `{'a':["apple","avocado"], 'b':["banana"]}`; `[]` → `{}`; `["zebra"]` → `{'z':["zebra"]}`.

**Edge cases:** empty input; empty string element; single-rune words; multi-byte first letter (`"élan"` → key `'é'`); mixed case.

**Test plan:** compare the whole map with deep equality; a dedicated test asserting within-group order matches input order; a test for your empty-string rule.

**Done when:** the order-preservation test passes and the empty-string rule is explicit.

