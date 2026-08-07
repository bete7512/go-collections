# 52 · Total area over []Shape

## Problem

Write a function that sums the areas of *any* shapes — including shapes that don't exist yet. `TotalArea` is four lines, and every one of them mentions only the interface. That constraint is the entire exercise: the moment a function's body needs to know which concrete type it's holding, the abstraction has failed and you should have designed a different interface.

This is the payoff half of #51. There, you learned that satisfaction is implicit; here you see what that buys — code written today that correctly handles types written next year by someone who never read your package.

## Contract (what the tests enforce)

```go
type Shape interface {
    Area() float64
}

type Circle struct{ R float64 }
type Square struct{ Side float64 }

func TotalArea(shapes []Shape) float64
```

- **`TotalArea` sums `Area()` over every element**, left to right.
- **The body must mention only `Shape`.** No type switches, no type assertions, no `Circle`/`Square` by name, no `reflect`. The tests can't inspect your source — but they *can* pass a type your file has never seen, which fails immediately if you branched on concrete types.
- **Empty slice → 0. Nil slice → 0.** Both are ordinary, neither is an error.
- **nil elements are skipped, not fatal.** A `[]Shape{Circle{1}, nil, Square{2}}` is representable — an interface value can be nil — and calling `Area()` on it would panic. `TotalArea` must defensively skip nil elements and sum the rest. This is pinned because it's a real API-boundary decision: you're summing data you didn't construct.
- The input slice is not modified.
- Float comparison in tests uses a relative tolerance (1e-12); any summation order or algebraic form is fine.
- `Circle`/`Square` carry forward from #51 unchanged (value receivers, π·R², Side²) — own package, so rebuild them; that's the intended re-drill.

## Worked examples

**Example 1 — mixed types, one loop:**

```go
shapes := []Shape{Circle{R: 1}, Square{Side: 2}}
TotalArea(shapes)   // → π + 4 ≈ 7.141592653589793
```

The loop body has no idea one of these is round.

**Example 2 — degenerate inputs:**

```go
TotalArea(nil)          // → 0
TotalArea([]Shape{})    // → 0
TotalArea([]Shape{Circle{}, Square{}})   // → 0  (zero-size shapes)
```

**Example 3 — the future-proofing property:**

```go
// defined somewhere TotalArea has never heard of:
type Hexagon struct{ Side float64 }
func (h Hexagon) Area() float64 { ... }

TotalArea([]Shape{Circle{1}, Hexagon{2}})   // just works — zero changes to TotalArea
```

If adding a shape type required editing `TotalArea`, the interface would be decoration rather than abstraction.

## Edge cases the tests cover

- Empty slice, nil slice → 0.
- Single element (each concrete type).
- Mixed `Circle`/`Square` slices, including many elements.
- All-zero-area shapes summing to 0.
- Negative dimensions (areas still positive — the squares).
- **nil elements interleaved** with real shapes — skipped, remainder summed correctly.
- A slice of *only* nil elements → 0.
- Pointers (`&Circle{}`) mixed with values in the same slice.
- **Two foreign types declared in the test file** (`hexagon`, `unitShape`) summed alongside yours — the open-set proof.
- A 1,000-element slice of alternating types with an arithmetically known total.
- Input-unmodified check.
