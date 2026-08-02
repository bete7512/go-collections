# 34 · Rune counts, sorted output

## Problem

Count how many times each rune appears in a string and return the counts as formatted lines in a deterministic order. This is a composition drill: the counting is #21's map-as-accumulator applied to runes, and the deterministic output is #31's collect-sort-range pattern. Two drills you own, snapped together — that composition instinct is the skill.

## Contract (what the tests enforce)

```go
func RuneCounts(s string) []string
```

- **One entry per distinct rune**, formatted exactly as `<rune>:<count>` — the character itself (not its code point number), a colon, the decimal count. `"banana"` → `"a:3"`, `"b:1"`, `"n:2"`.
- **Sorted ascending by rune** (Unicode code point order). That gives a total, deterministic order with consequences you should know going in:
  - Digits sort before uppercase, uppercase before lowercase: `'1'` (49) < `'A'` (65) < `'a'` (97).
  - The space character (32) sorts before all of them.
  - Accented and non-Latin characters sort after ASCII: `'z'` (122) < `'é'` (233) < `'世'` (19990).
- **Every rune counts.** Spaces, punctuation, digits, emoji — no filtering, no case folding. `'a'` and `'A'` are two different runes with two separate entries.
- **Runes, not bytes.** `"éé"` is one distinct rune with count 2, not two mystery bytes. Ranging over a string already yields runes; indexing into it yields bytes — this drill only works with the former.
- Empty string → empty result (the tests accept nil or empty via `slices.Equal`).
- Same input → byte-identical output, every run (the tests loop 50 times).

## Worked examples

**Example 1 — basic:**

```
input:  "banana"
output: ["a:3", "b:1", "n:2"]
```

Counts: a×3, b×1, n×2. Sorted by rune: `a` (97) < `b` (98) < `n` (110).

**Example 2 — mixed categories expose the ordering:**

```
input:  "Go 1!"
output: [" :1", "!:1", "1:1", "G:1", "o:1"]
```

Space (32) first, then `!` (33), then `1` (49), then `G` (71), then `o` (111). If you expected alphabetical-looking output, this example is the correction.

**Example 3 — unicode:**

```
input:  "zéz世é"
output: ["z:2", "é:2", "世:1"]
```

`z` (122) < `é` (233) < `世` (19990). A byte-based implementation would fall apart here — `é` is 2 bytes, `世` is 3.

## Edge cases the tests cover

- Empty string.
- Single rune; single rune repeated many times.
- Case sensitivity: `"aA"` produces `"A:1"` and `"a:1"` as separate entries, uppercase first.
- Spaces and punctuation counted and sorted by code point.
- Digits mixed with letters (ordering pinned above).
- Multi-byte runes: accented Latin, CJK, and an emoji — counted as single runes, ordered by code point.
- A longer mixed sentence with everything at once.
- The 50-run determinism loop — direct map iteration anywhere in the output path fails here.
