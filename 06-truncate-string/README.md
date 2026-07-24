# 6 · Truncate to N runes with ellipsis

- **Build:** `Truncate(s string, n int) string`
- **Expected:** `Truncate("hello world", 5)` → `"hello..."`; `Truncate("hi", 5)` → `"hi"` (no ellipsis if nothing was cut).
- **Edge cases:** `n == 0`; `n` ≥ rune count; string with emoji where byte length ≠ rune length; negative `n` (decide: panic, clamp, or error — document it).
- **Test:** cut / no-cut / exact-length / emoji cases.
- **Done when:** `Truncate("héé", 2)` returns `"hé..."`... i.e. two *runes*, not two bytes.
