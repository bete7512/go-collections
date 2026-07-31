# 31 · Iterate a map in sorted key order

## Problem

Go randomizes map iteration order on purpose — two runs over the same map visit keys in different orders. Any output built by ranging a map directly is therefore nondeterministic: log lines shuffle, golden-file tests flake, config dumps diff against themselves. The standard fix is a three-step pattern you will use for the rest of your Go career: **collect the keys, sort them, range the sorted keys and index the map.** This drill makes that pattern automatic.

## Contract (what the tests enforce)

```go
func SortedKeys(m map[string]int) []string
func FormatSorted(m map[string]int) string
```

- **`SortedKeys`** returns all keys in ascending lexicographic (byte-wise) order.
  - Empty or nil map → empty result (the tests accept nil or empty via `slices.Equal`).
  - The map itself is not modified.
- **`FormatSorted`** renders the map as one line per entry, exactly `key=value\n`, entries in sorted-key order:
  ```
  a=1
  b=2
  c=3
  ```
  - Every line ends with `\n`, **including the last**.
  - Empty or nil map → `""` (empty string, no newline).
- **Byte-wise string ordering is the ordering.** That means:
  - All uppercase letters sort before all lowercase (`"B" < "a"`, because `'B'` is byte 66 and `'a'` is 97).
  - Numeric-looking keys sort as strings: `"10" < "9"`.
  - The empty-string key `""` sorts before everything.
- **Determinism is the point:** the tests run both functions 50 times over the same map and require identical output every single run. Any accidental direct map-range in your output path will flake here.

## Worked examples

**Example 1 — basic:**

```
input: {"b":2, "a":1, "c":3}
SortedKeys  → ["a", "b", "c"]
FormatSorted → "a=1\nb=2\nc=3\n"
```

**Example 2 — ASCII ordering surprise:**

```
input: {"apple":1, "Banana":2, "cherry":3}
SortedKeys → ["Banana", "apple", "cherry"]
```

`"Banana"` first — uppercase `B` (66) is smaller than lowercase `a` (97). If you expected alphabetical-ignoring-case, that's a different (unpinned) function.

**Example 3 — numeric-looking keys:**

```
input: {"9":9, "10":10, "1":1}
SortedKeys → ["1", "10", "9"]
```

String comparison is character-by-character: `"10"` beats `"9"` because `'1' < '9'`.

## Edge cases the tests cover

- nil map and empty map for both functions.
- Single entry.
- Keys inserted in reverse order (input order is irrelevant — only sorted order matters).
- Uppercase/lowercase mixed keys (byte-order pinned above).
- Numeric-string keys (`"1"`, `"10"`, `"9"`).
- The empty-string key `""` (sorts first; formats as `=value`).
- Negative and zero values formatting correctly (`x=-5`, `y=0`).
- The 50-run determinism loop over a 10-key map — the test that fails if you range the map anywhere in the output path.

## Worth knowing (not tested)

The modern one-liner for the collect-and-sort step is `slices.Sorted(maps.Keys(m))` (Go 1.23+, `maps.Keys` returns an iterator). The classic form is a `make`+`append` loop plus `sort.Strings`. Know both — you'll read both in real codebases.
