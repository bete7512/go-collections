# 2 · Count vowels

- **Build:** `CountVowels(s string) int` (a e i o u, case-insensitive; no `strings.Count`).
- **Expected:** `"Go is fun"` → 3; `"rhythm"` → 0; `"AEIOU"` → 5.
- **Edge cases:** empty string; mixed case; non-letter characters.
- **Test:** table test covering the three examples plus an empty string.
- **Done when:** passes without any `strings` package call in the counting loop.
