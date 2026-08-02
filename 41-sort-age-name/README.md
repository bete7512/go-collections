# 41 · Sort by age, then by name

## Problem

Same slice as #40, but now ties are resolved: equal ages order by name. This is the two-level comparator — the shape every real-world sort eventually needs (order by priority then created-at, by score then id, by status then due-date). And it changes the testing story completely: with a total tiebreak the output is **fully deterministic**, so unlike #40 the tests here assert exact sequences everywhere and run repeatedly expecting identical results. A deterministic comparator is what makes stable-vs-unstable sorting irrelevant.

## Contract (what the tests enforce)

```go
type Person struct {
    Name string
    Age  int
}

func SortByAgeThenName(people []Person)
```

- **In place; age ascending; among equal ages, name ascending.**
- **Name comparison is byte-wise** (Go's `<` on strings): all uppercase letters sort before all lowercase — `"Bob" < "alice"` because `'B'` (66) < `'a'` (97). Pinned, tested; if you want case-insensitive, that's a different (unpinned) function.
- **The canonical comparator shape** — learn it as a template:
  ```go
  sort.Slice(people, func(i, j int) bool {
      if people[i].Age != people[j].Age {
          return people[i].Age < people[j].Age   // primary key decides
      }
      return people[i].Name < people[j].Name     // tiebreaker
  })
  ```
  Compare the primary key; if it differs, answer from it alone; only on a tie fall through. This extends to any number of levels by stacking more `if !=` blocks. The classic bug is `return a.Age < b.Age || a.Name < b.Name` — which wrongly lets the name override a *losing* age. If you write it that way, several tests fail; look at which.
- **Fully equal elements** (same age *and* name) are genuinely interchangeable — any order is correct by definition, and the permutation check covers them.
- Empty and nil slices: no panic, no effect.
- Determinism: the tests re-shuffle and re-sort the same data 20 times and require identical output every time.

## Worked examples

**Example 1 — the tiebreaker at work:**

```
input:  [{bob 25} {alice 25} {carol 20}]
after:  [{carol 20} {alice 25} {bob 25}]
```

carol's 20 wins outright; the two 25s order by name — alice before bob. Exactly one correct answer now (compare #40's Example 2, where both orders were legal).

**Example 2 — tiebreaker never fires on distinct ages:**

```
input:  [{zoe 18} {adam 30}]
after:  [{zoe 18} {adam 30}]
```

zoe stays first despite `"adam" < "zoe"` — the name is consulted *only* on age ties. A test with names ordered against ages catches comparators that mix the keys.

**Example 3 — the byte-order pin:**

```
input:  [{bob 25} {Alice 25} {ann 25}]
after:  [{Alice 25} {ann 25} {bob 25}]
```

`"Alice"` first: uppercase `A` (65) beats lowercase everything. Then `"ann"` vs `"bob"` byte-wise.

## Edge cases the tests cover

- Empty, nil, single element.
- Ties resolved alphabetically; multiple independent tie groups in one slice.
- All ages equal → pure name sort.
- All names equal → pure age sort.
- Names anti-correlated with ages (tiebreaker must NOT fire — catches the `||` bug).
- Mixed-case names (byte order pinned).
- Fully identical people (age and name) — permutation-checked.
- 20-run determinism loop over a reshuffled copy of the same data.
- 5,000-element case verified against an independently computed expected order.
