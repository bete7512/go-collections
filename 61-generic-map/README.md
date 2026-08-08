# 61 · Generic Map

## Problem

Generics arrive. Back in #18 you wrote `MapSlice(s []int, f func(int) int) []int` — useful, but locked to `int` in *both* positions. It could double numbers; it could never turn users into names or ints into strings. Type parameters remove that ceiling: `Map[T, U any]` reads a slice of one type and produces a slice of another, with the compiler generating a correctly-typed function at each call site.

The subtler skill here is **type inference**. Written correctly, no call site needs explicit type arguments — Go derives `T` from the slice and `U` from the function's return type. If you find yourself writing `Map[int, string](...)`, something forced inference to fail, and knowing which situations do that is worth more than the function itself.

## Contract (what the tests enforce)

```go
func Map[T, U any](s []T, f func(T) U) []U
```

- **`result[i] == f(s[i])`** for every index, order preserved, length equal to the input's.
- **`f` is called exactly once per element**, in index order. The tests verify both count and order with a stateful closure — so no skipping, no double-evaluation, no reordering.
- **Two independent type parameters**, both `any`. `T` and `U` may differ, may be identical, may be structs, pointers, slices, or interfaces.
- **Pre-size the result** with `make([]U, len(s))` (or `make([]U, 0, len(s))` + append). Growing by repeated append is the thing this drill should stop you doing.
- **Empty and nil inputs return an empty non-nil slice** — `make([]U, 0)` when `len(s) == 0`. Pinned: the tests check `got != nil` so you can't return a nil slice for empty input. (`slices.Equal` treats nil and empty as equal, which is exactly why the nil-ness needs its own explicit check.)
- **The input slice is never modified**, and the result is fresh memory — mutating the result must not touch the input, even when `T == U`.
- **`f` is never called for an empty input.**
- **Call sites must not need explicit type arguments.** Every call in the test file is written bare — `Map(nums, strconv.Itoa)` — so if your signature breaks inference, the test file won't compile.

## Worked examples

**Example 1 — the thing #18 could not do:**

```go
Map([]int{1, 2, 3}, strconv.Itoa)
// → []string{"1", "2", "3"}     T = int, U = string, both inferred
```

A method value or an existing stdlib function can be passed directly; no wrapper needed.

**Example 2 — struct → field extraction, the everyday use:**

```go
type User struct{ ID int; Name string }

users := []User{{1, "ada"}, {2, "grace"}}
Map(users, func(u User) string { return u.Name })   // → []string{"ada", "grace"}
Map(users, func(u User) int    { return u.ID   })   // → []int{1, 2}
```

Same input slice, two different `U`s, inferred from each literal's return type.

**Example 3 — T == U is still a fresh slice:**

```go
in  := []int{1, 2, 3}
out := Map(in, func(x int) int { return x * 2 })   // → []int{2, 4, 6}
out[0] = 999
in[0]   // → still 1. Different backing arrays.
```

## Edge cases the tests cover

- `int → string`, `string → int`, `struct → string`, `struct → int`, `int → struct`, `int → bool`.
- `T == U` (int → int) with a result-independence check.
- Empty slice and nil slice → empty **non-nil** result, and `f` never invoked.
- Single element.
- A transform returning a pointer type, and one taking a pointer type.
- Call-count and call-order verification via a closure recording every argument.
- Input-unmodified check on a struct slice.
- A transform with captured state (a running counter) proving left-to-right evaluation.
- 10,000 elements.
- Composition: `Map(Map(x, f), g)` type-checking and matching a direct `Map(x, g∘f)`.
