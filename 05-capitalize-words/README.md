# 5 · Capitalize every word

- **Build:** `TitleCase(s string) string`
- **Expected:** `"hello world"` → `"Hello World"`; `"go IS fun"` → `"Go IS Fun"` (only the first letter changes — don't lowercase the rest).
- **Edge cases:** empty string; single word; word starting with a digit or symbol (leave it alone); multi-byte first letter (`"éclair"` → `"Éclair"`).
- **Test:** the above four. Hint: `unicode.ToUpper` on the first rune of each field.
- **Done when:** the multi-byte case passes.
