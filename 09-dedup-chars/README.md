# 9 · Dedup characters, preserve order

- **Build:** `DedupChars(s string) string`
- **Expected:** `"banana"` → `"ban"`; `"aabbcc"` → `"abc"`.
- **Edge cases:** empty string; already-unique string; multi-byte runes (`"ééa"` → `"éa"`).
- **Test:** the three examples + empty.
- **Done when:** you used `map[rune]bool` (or `struct{}`) for seen-tracking and order is preserved.
