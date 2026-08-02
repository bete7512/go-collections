# 43 · Stack

## Problem

Build a LIFO (last-in-first-out) container on top of a slice. Go has no stack in the stdlib because a slice *is* one — `append` to push, truncate to pop — but wrapping it in a type with a safe API is the drill: this is where the `(value, ok bool)` convention for "might be empty" becomes muscle memory, the same convention as map lookups (`v, ok := m[k]`) and channel receives (`v, ok := <-ch`). Stacks are everywhere once you look: call stacks, undo histories, expression evaluators, DFS traversals, the parser you'll write in capstone #98.

## Contract (what the tests enforce)

```go
type Stack struct {
    items []int
}

func (s *Stack) Push(v int)
func (s *Stack) Pop() (int, bool)
func (s *Stack) Peek() (int, bool)
func (s *Stack) Empty() bool
func (s *Stack) Len() int
```

- **LIFO:** push 1, 2, 3 → pop yields 3, then 2, then 1.
- **`Pop` removes and returns the top**; on an empty stack it returns `(0, false)` — never a panic, never a magic sentinel like −1 (which would be indistinguishable from a pushed −1; the tests push negative values to make that point).
- **`Peek` returns the top without removing it:** `Len` unchanged after, and the following `Pop` returns the same value.
- **`Empty`** is true iff `Len() == 0`.
- **Pointer receivers on every method** — `Push` and `Pop` mutate, and a mixed method set (some value, some pointer) is a lint smell; keep them uniform. (`Empty`/`Len`/`Peek` don't mutate, but consistency wins.)
- **The zero value works:** `var s Stack` is immediately usable — `append` on a nil slice allocates. No constructor required. The tests start several cases from `var s Stack`.
- The field stays unexported; callers touch only the methods.
- Pop truncates with `s.items = s.items[:len(s.items)-1]`. For `int` elements that's the whole story; note in a comment that for pointer elements you'd nil the vacated slot first, or the backing array keeps the object alive (the #13/#14 lesson resurfacing).

## Worked examples

**Example 1 — LIFO order:**

```go
var s Stack
s.Push(1)
s.Push(2)
s.Push(3)
s.Pop()   // → (3, true)
s.Pop()   // → (2, true)
s.Pop()   // → (1, true)
s.Pop()   // → (0, false)  — empty, and the bool says so
```

**Example 2 — Peek doesn't consume:**

```go
var s Stack
s.Push(7)
s.Peek()  // → (7, true), Len() still 1
s.Peek()  // → (7, true), still 1 — peek any number of times
s.Pop()   // → (7, true), Len() now 0
```

**Example 3 — why `(0, false)` and not a sentinel:**

```go
var s Stack
s.Push(-1)
v, ok := s.Pop()   // → (-1, true): a real -1
v, ok = s.Pop()    // → (0, false): actually empty
```

With a sentinel-based API (`Pop() int` returning −1 when empty) these two situations would be indistinguishable. The two-value form is why Go APIs look the way they do.

## Edge cases the tests cover

- Zero-value stack usable immediately; `Empty`/`Len`/`Pop`/`Peek` all safe on it.
- Pop and Peek on empty → `(0, false)`.
- Push after draining to empty — the stack is reusable.
- Negative and zero values pushed (the sentinel trap).
- Duplicate values.
- Peek repeatedly: no consumption, Len stable, Peek-then-Pop agreement.
- Interleaved push/pop sequences maintaining LIFO throughout.
- Len tracked correctly after every operation.
- 10,000 pushes then 10,000 pops in exact reverse order, ending Empty.
