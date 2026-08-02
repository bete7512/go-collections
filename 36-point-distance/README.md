# 36 · Point with a Distance method

## Problem

Define a `Point` struct and give it a method computing the Euclidean distance to another point. The math is one line — the drill is about methods: attaching behavior to a type, choosing a receiver kind deliberately, and meeting the stdlib function that does this *better* than the formula you'd write by hand. This opens the structs-and-methods tier.

## Contract (what the tests enforce)

```go
type Point struct{ X, Y float64 }

func (p Point) Distance(q Point) float64
```

- **Euclidean distance:** √((p.X−q.X)² + (p.Y−q.Y)²).
- **Value receiver.** The method reads, never mutates; `Point` is two words — copying is cheaper than sharing. Write the receiver-choice reasoning in a one-line comment; it's the habit being drilled.
- Distance to itself is exactly 0; distance is always ≥ 0; `p.Distance(q) == q.Distance(p)` (the tests check symmetry over several pairs).
- Negative coordinates are ordinary coordinates.
- **Extreme magnitudes must not overflow or underflow.** This is the pin that forces the right tool: for coordinates around 1e200, the naive `math.Sqrt(dx*dx + dy*dy)` computes `dx*dx` ≈ 1e400 — beyond float64's max (~1.8e308) — and returns `+Inf`. For coordinates around 1e-200 it underflows the same way to 0. **`math.Hypot(dx, dy)` exists precisely to avoid this** — it rescales internally. The tests include both extremes; the formula-by-hand version fails them, `Hypot` passes.
- The tests compare with a *relative* tolerance (1e-12 of the expected value), because an absolute epsilon like 1e-9 is meaningless next to 1e200.

## Worked examples

**Example 1 — the 3-4-5 triangle:**

```
Point{0, 0}.Distance(Point{3, 4})   → 5
```

dx=3, dy=4 → √(9+16) = √25 = 5.

**Example 2 — negatives, same triangle shifted:**

```
Point{-1, -1}.Distance(Point{2, 3}) → 5
```

dx=3, dy=4 again — position doesn't matter, only the difference.

**Example 3 — the overflow case Hypot exists for:**

```
Point{1e200, 0}.Distance(Point{0, 1e200}) → ≈1.4142135623730951e200  (√2 × 1e200)
```

Naive: dx² = 1e400 → +Inf → the test fails with `got +Inf`. `math.Hypot(1e200, 1e200)` → the right answer, finite.

## Edge cases the tests cover

- Distance to itself → exactly 0.
- Horizontal-only and vertical-only separations (one axis zero).
- Negative coordinates in all quadrants.
- Irrational result (√2) — tolerance-checked.
- Symmetry `p.Distance(q) == q.Distance(p)` across a table of pairs.
- Huge coordinates (1e200 scale) — overflow trap.
- Tiny coordinates (1e-200 scale) — underflow trap: `{3e-200, 4e-200}` from origin must be 5e-200, not 0.
- Mixed magnitudes (one big, one small coordinate).
