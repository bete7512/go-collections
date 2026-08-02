# 37 · Rectangle with value receivers

## Problem

A `Rect` with two methods, `Area` and `Perimeter`. The arithmetic is trivial on purpose — this drill is about being able to *argue* the receiver choice, not just make it. #38 will give you the opposite case (a method that must mutate); together they form the complete receiver decision rule you'll apply on every type you ever define.

## Contract (what the tests enforce)

```go
type Rect struct{ W, H float64 }

func (r Rect) Area() float64      // W × H
func (r Rect) Perimeter() float64 // 2 × (W + H)
```

- **Value receivers on both methods.** The reasoning, which belongs in a comment above the type: the methods only read; the struct is two words, so a copy is as cheap as a pointer; and value semantics keep `Rect` freely copyable and comparable (usable as a map key, comparable with `==`) without anyone worrying about shared mutation.
- **Dimensions are taken as given.** `Area` and `Perimeter` compute the formulas on whatever `W`/`H` hold — including zero and negative values. No validation, no clamping: a `Rect{-3, 4}` reports area −12 and perimeter 2. Validation belongs at construction time (a `NewRect` returning an error), not silently inside arithmetic — that boundary is pinned so the tests can be exact.
- The zero value `Rect{}` works: area 0, perimeter 0, no panic.
- Methods don't mutate the receiver — the tests call them repeatedly and on copies, expecting identical answers.
- Comparisons use exact `==` in the tests: all test inputs are chosen to be exactly representable in float64 (integers and halves), so no tolerance is needed. Notice that choice — designing test data so exactness is safe is itself a technique.

## Worked examples

**Example 1 — basic:**

```
Rect{3, 4}      → Area 12, Perimeter 14
```

**Example 2 — zero value:**

```
Rect{}          → Area 0, Perimeter 0
```

The zero value of a well-designed type is usable without ceremony.

**Example 3 — square and fractional:**

```
Rect{5, 5}      → Area 25, Perimeter 20
Rect{2.5, 4}    → Area 10, Perimeter 13
```

## Edge cases the tests cover

- Zero-value rect.
- Zero in one dimension only (`Rect{0, 7}` → area 0, perimeter 14 — degenerate but well-defined).
- A square (W == H).
- Fractional dimensions (exactly-representable halves).
- Negative dimensions computing the raw formulas (pinned above: area −12 for `{-3, 4}`).
- Both dimensions negative (area positive again — the formulas, not opinions).
- Repeated calls returning identical results (no hidden mutation).
- Calling through a copy: `r2 := r; r2.Area()` equals `r.Area()` — value semantics visible.
