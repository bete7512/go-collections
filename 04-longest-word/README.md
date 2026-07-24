# 4 · Longest word

- **Build:** `LongestWord(s string) string`
- **Expected:** `"the quick brown fox"` → `"quick"` (5-way tie with `"brown"` — first wins; state your tie rule in a comment).
- **Edge cases:** empty string → `""`; multiple spaces between words; leading/trailing spaces.
- **Test:** normal sentence, tie case, empty string, messy whitespace (`strings.Fields` handles it).
- **Done when:** tie behavior is deliberate and tested, not accidental.
