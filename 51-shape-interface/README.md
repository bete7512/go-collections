# 51 · Shape interface

## Problem

Tier 4 opens with Go's defining feature: **implicit interface satisfaction**. You define an interface describing behavior, and any type that happens to have those methods satisfies it — no `implements` keyword, no registration, no import of the interface by the implementing type. `Circle` and `Square` know nothing about `Shape`; `Shape` knows nothing about them. That decoupling is why Go interfaces are usually defined by the *consumer* (the package that needs the behavior), not the producer — the inverse of the Java/C# habit.

Small interfaces are the norm here, not a simplification for teaching: `error`, `fmt.Stringer`, `io.Reader`, `io.Writer` are all one method. "The bigger the interface, the weaker the abstraction" is a Go proverb worth internalizing at this exact moment.

## Contract (what the tests enforce)

```go
type Shape interface {
    Area() float64
}

type Circle struct{ R float64 }
type Square struct{ Side float64 }

func (c Circle) Area() float64   // π·R²
func (s Square) Area() float64   // Side²
```

- **`Circle.Area()` is `math.Pi * R * R`**; `Square.Area()` is `Side * Side`. The tests compare floats with a relative tolerance (1e-12), so any algebraically equivalent form is fine.
- **Value receivers on both.** This is load-bearing and pinned: with a value receiver, *both* `Circle` and `*Circle` satisfy `Shape` (a pointer's method set includes its value-receiver methods). With a pointer receiver, a plain `Circle` value would **not** satisfy `Shape` — the tests assign both values and pointers into `Shape` variables, so a pointer receiver fails to compile. Worth deliberately breaking once after the tests pass: switch to `func (c *Circle) Area()`, read the "does not implement" error, and note that it names the method set as the reason.
- **Compile-time assertions are part of the deliverable** — put these in your file:
  ```go
  var _ Shape = Circle{}
  var _ Shape = Square{}
  ```
  This idiom turns "I forgot a method" into an error at the type definition rather than a confusing failure at some distant call site. It costs nothing at runtime (the variable is discarded) and appears throughout the stdlib.
- **Zero values are legal shapes.** `Circle{}` has area 0; `Square{}` has area 0. No validation, no errors — same reasoning as #37: geometry functions compute, constructors validate.
- Negative dimensions compute the raw formula (`Circle{-2}` → 4π, since R² is positive; `Square{-3}` → 9). Documented so the tests can be exact.
- A nil `Shape` variable is a real thing: `var s Shape` is nil, and calling `s.Area()` panics. The tests check the nil-ness, not the panic.

## Worked examples

**Example 1 — satisfaction is implicit:**

```go
var s Shape = Circle{R: 1}
s.Area()             // → 3.141592653589793

s = Square{Side: 2}  // the SAME variable now holds a different concrete type
s.Area()             // → 4
```

`Circle` never mentions `Shape`. The compiler checks the method set at the assignment.

**Example 2 — a heterogeneous slice:**

```go
shapes := []Shape{Circle{1}, Square{2}, Circle{3}}
for _, s := range shapes {
    s.Area()   // dispatches to the right concrete method
}
```

Different types, one slice, one loop. This is what #52 builds on.

**Example 3 — pointers work too (because of the value receiver):**

```go
c := Circle{R: 2}
var s Shape = &c     // compiles: *Circle's method set includes Area()
s.Area()             // → 12.566370614359172
```

## Edge cases the tests cover

- Compile-time `var _ Shape = Circle{}` / `Square{}` assertions.
- Unit circle (π), radius 2 (4π), zero radius, negative radius.
- Unit square, side 2, zero side, negative side, fractional side.
- Both concrete types assigned to the same `Shape` variable in sequence.
- Pointers (`&Circle{}`, `&Square{}`) assigned to `Shape` — the method-set rule.
- A `[]Shape` holding a mix of both types, each dispatching correctly.
- A nil `Shape` variable comparing equal to nil.
- A third type defined **inside the test file** satisfying `Shape` without touching your code — the open-set property of implicit interfaces.
