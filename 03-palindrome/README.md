# 3 · Palindrome (ignore case + punctuation)

- **Build:** `IsPalindrome(s string) bool`
- **Expected:** `"A man, a plan, a canal: Panama"` → true; `"race a car"` → false; `"No 'x' in Nixon"` → true.
- **Edge cases:** empty string → true; single char → true; all-punctuation string → true; digits count as characters.
- **Test:** the three examples + edge cases. Hint: `unicode.IsLetter`, `unicode.IsDigit`, `unicode.ToLower`.
- **Done when:** two-pointer version passes (not "clean the string then compare with its reverse" — do that only as a warm-up).
