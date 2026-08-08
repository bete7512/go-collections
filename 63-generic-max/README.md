# 63 · Generic Max

## Problem

`any` says "I don't care what type this is." But `Max` *does* care — it needs `<`, and most types don't have it. This is the drill where a type parameter gets **constrained**: `cmp.Ordered` is an interface listing every type that supports `<`, `<=`, `>`, `>=` (all integer kinds, all float kinds, and string). Write one function body, get a correct `Max` for `int`, `uint8`, `float64`, `string`, and every named type built on them.

The second lesson is the empty case. There is no safe value to return for an empty slice — `0` is wrong for all-negative ints, `""` is wrong for strings — so the signature must carry a failure path, and you need `var zero T` to produce *something* alongside the error. That idiom (the only way to name a generic type's zero value) shows up constantly once you write generic code.

## Contract (what the tests enforce)

```go
import "cmp"

var ErrEmpty = errors.New("max of empty slice")

func Max[T cmp.Ordered](s []T) (T, error)
```

- **Returns the largest element** and a nil error. One pass, starting from `s[0]`.
- **Empty or nil slice** → `(zero, ErrEmpty)` where `zero` is `T`'s zero value, obtained via `var zero T`. `errors.Is(err, ErrEmpty)` must be true, so return the sentinel (wrapped or bare, your call).
- **`cmp.Ordered`** (stdlib since Go 1.21) is the constraint. On older toolchains the equivalent was `golang.org/x/exp/constraints.Ordered` — know both names, since you'll meet the second in older code.
- **Ties:** with duplicate maxima, the returned value is the same either way — but return the **first** one encountered (use strict `>` when updating, not `>=`). Unobservable for `int`; observable for named types or if you later extend this to return an index, which is why it's pinned.
- **Named types work:** `type Celsius float64` and `type Priority int` satisfy `cmp.Ordered` because their underlying types do. The tests define such types and pass slices of them, with the result's type checked by assignment.
- **String comparison is byte-wise**, so `"Z" < "a"` — `Max([]string{"apple", "Zebra"})` is `"apple"`. Pinned; a case-insensitive max is a different function.
- **NaN caveat (documented, not handled):** every comparison involving NaN is false, so a NaN in a float slice can make the result depend on position. The tests pin what your straightforward implementation does — a NaN at index 0 stays "max" (nothing compares greater), a NaN later is ignored — and require a comment naming the hazard. Don't add NaN handling; know it exists.
- **Compare with the stdlib afterward:** Go 1.21 added the builtin `max(a, b, ...)` for a fixed argument list, and `slices.Max` for a slice — which **panics** on empty rather than returning an error. Note in a comment which API style you'd choose for a library and why.

## Worked examples

**Example 1 — one body, three element types:**

```go
Max([]int{3, 1, 4})                    // (4, nil)
Max([]float64{-2.5, -9.0})             // (-2.5, nil)   ← init-to-zero would return 0 here
Max([]string{"apple", "pear", "fig"})  // ("pear", nil) ← lexicographic
```

**Example 2 — the empty case needs `var zero T`:**

```go
v, err := Max([]string{})
// v   == ""            (T's zero value, produced by `var zero T`)
// err  matches ErrEmpty
```

You cannot write `return 0, err` in a generic function — `0` isn't valid for `string`. `var zero T` is the only way to name it.

**Example 3 — named types ride along for free:**

```go
type Celsius float64
temps := []Celsius{21.5, 30.0, 18.2}
hottest, _ := Max(temps)   // hottest is a Celsius, not a float64
```

The constraint is satisfied through the underlying type, and the return type stays `Celsius`.

## Edge cases the tests cover

- `int`, `int64`, `uint8`, `float64`, `string` — five element types through one function.
- Two named types (`Celsius`, `Priority`) with the returned value assigned back to that named type.
- All-negative slices (catches an implementation that starts the running max at zero).
- Single element; all-equal elements; duplicate maxima.
- Max at the first position, the last position, and the middle.
- Empty and nil slices → `errors.Is(err, ErrEmpty)` and the correct zero value for each of several `T`s.
- String byte-ordering (`"Zebra" < "apple"`), empty strings in the slice, and unicode.
- `uint8` at 255 and `int64` near its maximum (no overflow in the comparison).
- Float slices containing `+Inf` and `-Inf`.
- Documented NaN behavior: NaN first vs NaN later.
- 10,000 elements with the maximum planted at the end.
