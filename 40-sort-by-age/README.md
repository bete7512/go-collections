# 40 · Sort []Person by age with sort.Slice

## Problem

Sort a slice of structs by one field. `sort.Slice` is the everyday Go sort — you hand it a slice and a *less function*, and it does the rest. The drill has two lessons: the shape of a correct less function, and a fact most people learn the hard way — **`sort.Slice` is not stable**, so this drill's tests must be (and are) written to not depend on the order of equal elements. Reading how they do that is part of the exercise.

## Contract (what the tests enforce)

```go
type Person struct {
    Name string
    Age  int
}

func SortByAge(people []Person)
```

- **Sorts in place, ascending by age.** No return value; the caller's slice is reordered.
- **The less function must be strict:** `people[i].Age < people[j].Age` — strictly less-than, never `<=`. A non-strict comparator violates `sort.Slice`'s contract (it can panic or misorder — the sort assumes `less(a,b)` and `less(b,a)` can't both be true).
- **Equal-age order is unspecified, on purpose.** `sort.Slice` may reorder people who share an age (it's an unstable quicksort variant). The contract therefore promises only: ages ascending, and the result is a permutation of the input (nobody added, dropped, or duplicated). The tests assert exactly that and nothing more for tie cases; exact sequences are only asserted where all ages are distinct. If you want ties preserved in input order, that's `sort.SliceStable` — and deterministic tie-breaking by another field is exactly #41.
- Empty and nil slices: no panic, no effect.
- Ages of zero are ordinary values.

## Worked examples

**Example 1 — distinct ages, fully determined:**

```
input:  [{bob 30} {alice 25} {carol 20}]
after:  [{carol 20} {alice 25} {bob 30}]
```

**Example 2 — a tie, deliberately underdetermined:**

```
input:  [{bob 25} {alice 25} {carol 20}]
after:  [{carol 20} {alice 25} {bob 25}]   ← valid
   or:  [{carol 20} {bob 25} {alice 25}]   ← equally valid
```

Both outcomes satisfy the contract. The tests for this case check the age sequence `[20 25 25]` and that alice, bob, carol are all still present exactly once — they do **not** check who of the two 25-year-olds comes first. That's what "unstable sort" means operationally.

**Example 3 — the canonical call:**

```go
sort.Slice(people, func(i, j int) bool {
    return people[i].Age < people[j].Age
})
```

The closure captures `people`; `i` and `j` are indices into it. This shape — slice, closure, strict `<` — is the thing to drill into muscle memory.

## Edge cases the tests cover

- Empty slice and nil slice (no panic).
- Single element.
- Already sorted, reverse sorted, and shuffled inputs with distinct ages — exact order asserted.
- All ages equal — only the permutation property is asserted.
- Some ages tied — age sequence + permutation, tie order left free.
- Age zero sorting before everything.
- In-place proof: the tests assert on the very slice they passed in.
- 10,000 shuffled people — sortedness and permutation verified at scale.
