# 8 · Join without `strings.Join`

- **Build:** `Join(parts []string, sep string) string`
- **Expected:** `Join([]string{"a","b","c"}, "-")` → `"a-b-c"`.
- **Edge cases:** empty slice → `""`; single element → no separator; empty separator.
- **Test:** compare against `strings.Join` as the oracle.
- **Done when:** it passes using `strings.Builder` (not `+=` in a loop — know why: quadratic copying).
