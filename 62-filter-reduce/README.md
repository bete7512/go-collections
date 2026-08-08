# 62 · Generic Filter and Reduce

## Problem

`Filter` is the easy half — #17 with a type parameter. `Reduce` is where this drill earns its place: it collapses a slice into a **single value of a different type**, and that accumulator type is what makes it the most general of the three. Map and Filter are both special cases of Reduce; the last test in this suite makes you prove it by implementing `Map` in one line on top of your `Reduce`.

Worth knowing where Go actually stands: `slices` ships `IndexFunc`, `DeleteFunc`, `ContainsFunc`, `SortFunc` — but deliberately **no** `Map`/`Filter`/`Reduce`. The proposals exist and keep being declined, largely because a `for` loop is clearer than a pipeline for most Go code and generics arrived late enough that the idiom was settled. Building these teaches the shape; using them in production Go is a style choice you should make consciously, not by reflex.

## Contract (what the tests enforce)

```go
func Filter[T any](s []T, keep func(T) bool) []T
func Reduce[T, U any](s []T, init U, f func(acc U, v T) U) U
```

**Filter:**
- Keeps elements where `keep` returns true, preserving relative order.
- Calls `keep` **exactly once per element, in index order**.
- Contains no domain logic — all decisions live in the passed function.
- Input untouched; result is fresh memory (mutating the result must not affect the input).
- Empty, nil, and match-nothing inputs return an **empty non-nil** slice (explicitly checked, since `slices.Equal` can't distinguish nil from empty).

**Reduce:**
- **Folds left:** starts from `init`, applies `f(acc, s[i])` for each index in order, returns the final accumulator.
- **Argument order is `(accumulator, element)`** — pinned, because swapping it is the classic self-inflicted bug and the tests use non-commutative operations (string concatenation, slice append) that catch it immediately.
- **Empty or nil input returns `init` unchanged**, and `f` is never called.
- `U` may be any type: a number, a string, a map, a slice, a struct. The tests exercise all of these — a Reduce that only works when `T == U` fails here.
- `f` is called exactly once per element, in index order.

## Worked examples

**Example 1 — Reduce with T ≠ U, the interesting case:**

```go
words := []string{"go", "rust", "zig"}
Reduce(words, 0, func(acc int, s string) int { return acc + len(s) })
// → 9      T = string, U = int
```

The accumulator is a different type from the elements. That's the capability `Map` and `Filter` don't have.

**Example 2 — argument order matters:**

```go
Reduce([]string{"a", "b", "c"}, "", func(acc, v string) string { return acc + v })
// → "abc"     correct: accumulator on the left

// If your implementation calls f(element, accumulator) instead:
// → "cba"     — same signature, silently reversed result
```

**Example 3 — U as a composite:**

```go
// Build a set from a slice.
Reduce([]string{"a", "b", "a"}, map[string]bool{}, func(acc map[string]bool, s string) map[string]bool {
    acc[s] = true
    return acc
})
// → map[a:true b:true]
```

**Example 4 — the capstone: Map via Reduce.**

```go
func MapViaReduce[T, U any](s []T, f func(T) U) []U {
    return Reduce(s, make([]U, 0, len(s)), func(acc []U, v T) []U { return append(acc, f(v)) })
}
```

One line, and it satisfies #61's contract. The suite runs it against Map-style assertions — including the empty-input non-nil rule, which is why `init` is `make([]U, 0, len(s))` rather than `nil`.

## Edge cases the tests cover

**Filter:** even/odd predicates, match-all, match-nothing, empty, nil, single element, struct elements, a predicate with captured state, call-order verification, input-unmodified, result independence, 10,000 elements.

**Reduce:** sum, product, max; string concatenation (order-sensitive); `T ≠ U` sum-of-lengths; `U` as `map[string]bool`, as `[]int`, and as a struct accumulator; non-zero `init`; empty and nil returning `init` with `f` never called; call-order verification; 10,000 elements.

**Composition:** `Filter` then `Reduce` in a pipeline, and `MapViaReduce` matching direct construction on several type pairs including the empty-input case.
