# 53 · Type switch

## Problem

#52 was about never needing the concrete type. This one is the deliberate opposite: sometimes you genuinely must recover it — classifying errors, decoding untyped JSON, writing a formatter. The tool is the **type switch**, `switch v := s.(type)`, and the point of `v :=` is that inside each case, `v` is already that concrete type, no assertion needed.

Use it knowing the trade-off: a type switch over an interface is a **closed set** of behavior in an **open set** of types. Add a type tomorrow and this function silently falls to `default`, while `TotalArea` from #52 keeps working untouched. That's not a reason never to use one — it's the reason to ask, each time, whether the behavior belongs as a method on the interface instead.

## Contract (what the tests enforce)

```go
type Shape interface{ Area() float64 }

type Circle struct{ R float64 }
type Square struct{ Side float64 }

func Describe(s Shape) string
```

Exact return strings, pinned so the tests can be literal:

| input | returns |
|---|---|
| `Circle{R: 1}` | `circle with radius 1` |
| `Circle{R: 2.5}` | `circle with radius 2.5` |
| `Square{Side: 2}` | `square with side 2` |
| `nil` | `no shape` |
| any other `Shape` | `unknown shape with area <A>` |

- **Number formatting is `%v` on the `float64`** — Go's default float rendering: `1` not `1.0`, `2.5` as `2.5`, `-3` as `-3`. Same rule as #39's Stringer.
- **`unknown shape with area <A>`** formats the result of calling `s.Area()` with `%v` — so the default branch still uses the interface, which is the honest fallback: you don't know the type, but you know it's a Shape.
- **A `case nil:` branch is required.** A nil interface value does *not* match `case Circle:` and would otherwise fall into `default`, where calling `s.Area()` panics. Handle it explicitly and first.
- **`Circle` and `*Circle` are different cases.** A type switch matches the dynamic type exactly; `case Circle:` will not catch a `*Circle`. The tests pass pointers and expect them to land in `default` (formatted via `Area()`, which works because of the value receiver). Pinning it this way makes the rule visible rather than papering over it — if you want pointers handled, that's `case Circle, *Circle:` or a separate branch, and it's a deliberate choice.
- **Use the bound variable.** Read `v.R` and `v.Side` from the switch variable inside the branches — not by re-asserting `s.(Circle)`. That's the whole ergonomic point of the `v :=` form.

## Worked examples

**Example 1 — the concrete branches:**

```go
Describe(Circle{R: 1})     // → "circle with radius 1"
Describe(Square{Side: 2})  // → "square with side 2"
```

Inside `case Circle:`, `v` is a `Circle` — `v.R` compiles. Inside `case Square:`, the same `v` is a `Square` and `v.Side` compiles. One variable, different static type per branch; only a type switch can do that.

**Example 2 — nil comes first:**

```go
var s Shape          // nil interface
Describe(s)          // → "no shape"   (never touches s.Area())
```

**Example 3 — the open set arriving:**

```go
type Triangle struct{ B, H float64 }
func (t Triangle) Area() float64 { return t.B * t.H / 2 }

Describe(Triangle{B: 4, H: 3})   // → "unknown shape with area 6"
```

`Describe` was written before `Triangle` existed and degrades gracefully — but it *did* degrade. Compare with `TotalArea`, which handled the same new type perfectly. That contrast is the lesson.

## Edge cases the tests cover

- Both concrete branches with several values, including fractional and negative dimensions (`circle with radius -2`).
- Whole numbers rendering without a decimal point (`radius 1`, not `radius 1.000000`).
- Zero-value `Circle{}` and `Square{}`.
- Explicit nil `Shape`, and a `var s Shape` never assigned.
- `*Circle` and `*Square` landing in `default` (the exact-dynamic-type rule).
- Two foreign types declared in the test file hitting `default` with their computed areas.
- A `[]Shape` of every category run through `Describe` in one pass.
