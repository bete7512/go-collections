# 64 · Generic Set

## Problem

#26's `Set` worked on strings only. Generalizing it introduces the third constraint you need in practice — **`comparable`** — and the reason it exists is concrete: a set is a map underneath, map keys must support `==`, and not every Go type does. Slices, maps, and functions are **not** comparable, so `Set[[]int]` must not compile. Reading that compile error is part of this exercise; it's the constraint system telling you something true about the runtime.

The other new thing is a **method set on a named map type**. `type Set[T comparable] map[T]struct{}` gets methods directly — no wrapper struct — which means a `Set[T]` value behaves like a reference: copies share the same underlying map. That's convenient and occasionally surprising, so the tests pin it.

## Contract (what the tests enforce)

```go
type Set[T comparable] map[T]struct{}

func NewSet[T comparable](items ...T) Set[T]
func (s Set[T]) Add(v T)
func (s Set[T]) Has(v T) bool
func (s Set[T]) Remove(v T)
func (s Set[T]) Len() int
func (s Set[T]) Items() []T   // unordered; for tests and iteration
```

- **`map[T]struct{}`**, not `map[T]bool` — the empty struct occupies zero bytes, so the map stores only keys. Be able to say that out loud; it's a standard Go interview question and a real memory difference at scale.
- **`NewSet` is variadic** and deduplicates: `NewSet(1, 2, 1)` has `Len() == 2`.
- **`Add` of an existing element is a no-op** (length unchanged); **`Remove` of an absent element is a no-op**, never a panic (Go's `delete` is already safe on missing keys).
- **`Has` on an empty or nil set returns false** without panicking — reading a nil map is legal in Go.
- **Value receivers, and mutation still works.** A `Set[T]` is a map value; copying it copies the map header, and both copies index the same buckets. The tests verify that adding through a copy is visible in the original — the same shared-backing-store lesson as #42's slice `Swap`, in a different costume.
- **The nil-set asymmetry is pinned:** `var s Set[int]` (nil map) supports `Has`, `Len`, `Remove`, and `Items` safely, but `Add` on it **panics** (assignment to entry in nil map). The tests assert both halves — the safe reads *and* the panic, captured with `recover`. Document this: it's why `NewSet` exists, and it's a real trade-off of the named-map-type design versus a struct wrapper that could lazily initialize.
- **`Items` returns the elements in unspecified order.** The tests always sort before comparing — never assert an order from a map.
- **Element types the tests use:** `string`, `int`, a struct (`Point`), a pointer, an array (`[2]int`), and a named type. Structs and arrays are comparable when all their fields/elements are; that's why they work as keys.
- **`Set[[]int]` must not compile.** Keep a commented-out declaration in your file with the compiler's error text next to it.

## Worked examples

**Example 1 — one type, many instantiations:**

```go
NewSet("a", "b", "a").Len()          // 2
NewSet(1, 2, 3).Has(2)               // true
NewSet(Point{1, 2}, Point{1, 2})     // Len 1 — structs compare field by field
```

**Example 2 — copies share the map:**

```go
a := NewSet(1, 2)
b := a          // copies the map header, not the contents
b.Add(3)
a.Has(3)        // TRUE — same underlying map
a.Len()         // 3
```

If you want an independent copy, you must build a new set explicitly. (Compare `slices.Clone`.)

**Example 3 — the nil set:**

```go
var s Set[int]   // nil map
s.Has(1)         // false — safe
s.Len()          // 0 — safe
s.Remove(1)      // safe, no-op
s.Add(1)         // PANIC: assignment to entry in nil map
```

**Example 4 — the constraint doing its job:**

```go
// type _ = Set[[]int]
// invalid map key type []int  /  []int does not satisfy comparable
```

`comparable` is precisely "usable as a map key". The error is the constraint system preventing a runtime impossibility at compile time.

## Edge cases the tests cover

- Full lifecycle (`Add` → `Has` → `Remove` → `Has`) for `string`, `int`, `Point`, `[2]int`, a named type, and a pointer type.
- Double `Add`, `Remove` of an absent element, `Remove` down to empty, then re-`Add`.
- `NewSet` with no arguments, with duplicates, and with a single element.
- Zero values as elements (`0`, `""`, `Point{}`) — legitimate members, not sentinels.
- Nil set: `Has`/`Len`/`Remove`/`Items` safe; `Add` panics (recovered and asserted).
- Copy-shares-the-map semantics.
- `Items` length matching `Len`, contents matching after sorting, and empty/nil sets yielding an empty result.
- Two different pointers to equal structs being distinct elements (pointer identity, not pointee equality).
- 10,000 insertions with duplicates, verifying `Len` and membership.
