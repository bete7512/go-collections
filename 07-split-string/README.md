# 7 · Split without `strings.Split`

- **Build:** `Split(s, sep string) []string`
- **Expected:** `Split("a,b,,c", ",")` → `["a" "b" "" "c"]`; `Split("abc", ",")` → `["abc"]`.
- **Edge cases:** empty input → `[""]`; separator at start/end (`",a,"` → `["" "a" ""]`); multi-char separator; separator not found.
- **Test:** compare your output against `strings.Split` for ~6 inputs — the stdlib is your oracle.
- **Done when:** it matches `strings.Split` on every case you throw at it (single-rune `sep` is enough; you can skip `sep == ""`).
