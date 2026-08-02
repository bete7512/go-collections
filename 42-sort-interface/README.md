# 42 · Implement sort.Interface manually

## Problem

Before `sort.Slice` (Go 1.8) and `slices.SortFunc` (1.21), *every* Go sort was written the way you're about to: define a named slice type, give it the three methods of `sort.Interface`, and hand it to `sort.Sort`. You'll still meet this pattern all over older codebases and inside the stdlib itself — and writing it once demystifies what the convenient forms generate for you. The bonus lesson is subtle and worth the whole drill: `Swap` mutates through a **value** receiver, and understanding why teaches you what a slice header really is.

## Contract (what the tests enforce)

```go
type Person struct {
    Name string
    Age  int
}

type ByAge []Person

func (a ByAge) Len() int
func (a ByAge) Less(i, j int) bool   // strictly: a[i].Age < a[j].Age
func (a ByAge) Swap(i, j int)

// usage: sort.Sort(ByAge(people))
```

- **The three methods on the named slice type `ByAge`**, satisfying `sort.Interface`. The tests include the compile-time proof `var _ sort.Interface = ByAge(nil)`.
- **`Len`** returns the element count. **`Less(i, j)`** reports strictly whether element i must sort before element j — age ascending, `<` never `<=`. **`Swap(i, j)`** exchanges the two elements (the one-line parallel assignment).
- **Value receivers throughout — and Swap still mutates the caller's data.** This is not an exception to #38's copy rule; it's the rule applied precisely. Copying a slice copies the *header* (pointer, length, capacity) — both copies point at the **same backing array**. `Swap` writes through that shared pointer. Contrast with #38, where the struct's data itself was copied. Put this in a comment; it's the insight.
- **`ByAge(people)` is a conversion, not a copy.** Same backing array, new type. Consequence the tests verify: after `sort.Sort(ByAge(people))`, the original `people` variable is sorted.
- **Same ordering contract as #40:** ages ascending; equal-age order unspecified (`sort.Sort` is also unstable — `sort.Stable` is the stable variant). The tests reuse #40's split: exact sequences for distinct ages, age-sequence + permutation for ties.
- The three methods must be honest independently: the tests call `Len`/`Less`/`Swap` directly, not only through `sort.Sort`.

## Worked examples

**Example 1 — the full ritual:**

```go
people := []Person{{"bob", 30}, {"alice", 25}, {"carol", 20}}
sort.Sort(ByAge(people))
// people is now [{carol 20} {alice 25} {bob 30}] — the ORIGINAL slice
```

**Example 2 — the methods are just methods:**

```go
a := ByAge{{"x", 10}, {"y", 20}}
a.Len()        // 2
a.Less(0, 1)   // true  (10 < 20)
a.Less(1, 0)   // false
a.Swap(0, 1)   // a is now [{y 20} {x 10}] — through a value receiver
```

**Example 3 — what sort.Slice was doing all along:**

`sort.Slice(people, less)` wraps your closure and a reflect-based swap into exactly this interface and calls the same internal sort. You've now written the layer it generates.

## Edge cases the tests cover

- Compile-time `sort.Interface` satisfaction.
- `Len` on empty, nil, and populated slices.
- `Less` strictness: `Less(i, j)` and `Less(j, i)` both false for equal ages; asymmetric for distinct ones.
- `Swap` visible through the original slice variable (the shared-backing-array proof), including a double-swap restoring the original.
- Sorting via `sort.Sort`: empty, single, sorted, reverse, shuffled — exact sequences (distinct ages).
- Ties: age sequence + permutation only, order among equals free.
- The original-variable-is-sorted assertion after converting and sorting.
- 10,000 elements through `sort.Sort` — ascending + permutation.
