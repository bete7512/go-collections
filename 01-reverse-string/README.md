# 1 · Reverse a string (rune-safe)

- **Build:** `Reverse(s string) string`
- **Expected:** `"hello"` → `"olleh"`; `"héllo"` → `"olléh"`; `"go👋"` → `"👋og"`. Reversing raw bytes would corrupt `é` and `👋` — that's the whole point.
- **Edge cases:** empty string; single rune; string that is already a palindrome.
- **Test:** table test with ASCII, accented, and emoji inputs; assert `Reverse(Reverse(s)) == s`.