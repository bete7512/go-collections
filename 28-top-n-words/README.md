# 28 · Top N frequent words

## Problem

Given a free-form text and an integer `n`, return the `n` most frequent words, ranked by how often they appear. When two words appear the same number of times, the alphabetically smaller word ranks first. You receive raw text, not a pre-split word list — tokenization is part of the problem.

The drill behind it: you cannot sort a Go map. You must count in a `map[string]int`, convert to a slice of word/count pairs, and sort that slice with a two-level comparator. The alphabetical tiebreak is what makes the output deterministic even though map iteration order is randomized — remove it and the same input can produce different results run to run.

## Contract (exactly what the tests enforce)

```go
func TopN(text string, n int) []string
```

- **Tokenization:** words are maximal runs of non-whitespace, i.e. `strings.Fields` semantics. Spaces, tabs, and newlines all separate words.
- **Case folding:** words are lowercased before counting — `"Go"`, `"GO"`, and `"go"` are one word, reported as `"go"`.
- **No punctuation stripping:** `"go,"` and `"go"` are **different** words. (Pinned so the contract stays simple; stripping is a separate drill.)
- **Ranking:** count descending, then word ascending (plain `<` on strings) for equal counts.
- **Result size:** exactly `min(n, distinct words)` entries.
- **`n <= 0`**, empty text, or whitespace-only text → empty result. Tests accept `nil` or `[]string{}` — only the length is asserted.
- **Determinism:** same input → identical output, every call. The tests run tie-heavy inputs in a loop to catch map-order leakage.

## Examples

**1.** `TopN("b b a a a c", 2)` → `["a", "b"]`
Counts: a:3, b:2, c:1. Highest two counts win; no tie to break.

**2.** `TopN("the day is sunny the the the sunny is is", 4)` → `["the", "is", "sunny", "day"]`
Counts: the:4, is:3, sunny:2, day:1. Pure count ordering.

**3.** `TopN("b a a b", 2)` → `["a", "b"]`
Both words have count 2 — the tiebreak decides: `"a" < "b"`.

**4.** `TopN("Go go GO gopher", 1)` → `["go"]`
Case folding merges the three spellings of "go" into count 3, beating gopher:1.

**5.** `TopN("d c b a", 10)` → `["a", "b", "c", "d"]`
All counts equal → fully alphabetical; n larger than the vocabulary returns everything.

## Edge cases the tests cover

- Tie on count → alphabetical order (the core case, run 20× for determinism).
- All words the same frequency → output is fully sorted alphabetically.
- `n` larger than the number of distinct words → all of them, still ranked.
- `n == 0` and negative `n` → empty.
- Empty string and whitespace-only string → empty.
- Mixed whitespace separators (tabs, newlines, runs of spaces).
- Mixed case merging into one word.
- Punctuation kept: `"go go, go,"` counts `go,`:2 and `go`:1.
- Unicode words (`"héllo"`) count and return correctly.
- Numeric tokens (`"42"`) are words like any other.
- Single word repeated; single distinct word with large n.
