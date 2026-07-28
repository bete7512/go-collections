# 21 · Word frequency

**Goal:** Count how many times each word appears in a paragraph. The canonical "map as accumulator" drill.

**Signature:**
```go
func WordFreq(text string) map[string]int
```

**Requirements:**
- Split on whitespace (`strings.Fields` handles runs of spaces, tabs, newlines).
- Normalize case — lowercase everything so "The" and "the" are one word. Document it.
- Decide on punctuation: either strip leading/trailing punctuation (`strings.Trim` with a cutset, or `unicode.IsPunct`) or don't. Document your rule; the tests must match it.
- Empty/whitespace-only input returns an empty (non-nil) map.
- Go's map increment on a missing key works without initialization (`m[w]++` starts from the zero value) — use that, don't write an `if _, ok` guard.

**Examples:** `"the cat and the hat"` → `{"the":2,"cat":1,"and":1,"hat":1}`; `"The the THE"` → `{"the":3}`; `""` → `{}`.

**Edge cases:** mixed case; multiple spaces / tabs / newlines between words; punctuation-attached words (`"cat,"` vs `"cat"`); empty string; string of only spaces.

**Test plan:** table test comparing whole maps (`maps.Equal` from the stdlib `maps` package, or `reflect.DeepEqual`); include a case that exercises your punctuation rule explicitly.

**Done when:** one pass over the fields, one map, and the case-normalization test passes.

