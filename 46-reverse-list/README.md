# 46 · Reverse a linked list

## Problem

Reverse the list in place: every `Next` pointer flips to point backward, the old tail becomes the head, and **no new nodes are allocated** — the same nodes get rewired. This is the most-asked pointer question in interviews, and deservedly: the three-pointer loop is four lines, yet writing it cold, correctly, first try, is a real test of whether pointer manipulation has become reflex. That's the bar here — drill it until you can write it a day later without thinking.

Each challenge folder is its own package, so `Node` and `List` are (re)defined here — rebuilding #45's `Add`/`Values`/`Length` from memory is deliberate re-drill, per the every-ten-redo rule.

## Contract (what the tests enforce)

```go
type Node struct {
    Val  int
    Next *Node
}

type List struct {
    head *Node
    n    int
}

func (l *List) Add(v int)          // append at tail (from #45)
func (l *List) Values() []int      // head→tail traversal (from #45)
func (l *List) Length() int        // O(1) counter (from #45)
func (l *List) Reverse()           // reverse in place

func ReverseChain(head *Node) *Node // the same operation as a free function
```

- **`ReverseChain` takes a chain's head, returns the new head** (the old tail). It must **reuse the existing nodes**: the tests build chains by hand, keep pointers to each node, and assert *pointer identity* after reversal — the returned head must be the very node that was the tail, with every `Next` flipped. Allocating fresh nodes produces equal values but different pointers, and fails.
- `ReverseChain(nil)` → `nil`; a single node returns itself with `Next == nil`.
- **`List.Reverse()`** applies the same rewiring to the list's own chain and updates `l.head`. `Length` is unchanged; `Values` comes back reversed. Reversing twice restores the original order. An empty list is a no-op.
- **The required technique: iterative, three pointers (`prev`/`curr`/`next`), single pass, O(1) extra space.** No new nodes, no slice buffer, no recursion. The loop-body *order* is the entire trick — you must save the onward link before you flip it, or the rest of the list is unreachable; work out the exact sequence yourself, that's the drill. (The recursive version is legal Go but costs O(n) stack — know why, don't submit it.)
- `Add` must still work after a `Reverse` (the tests append post-reversal and check order) — if you cached a tail pointer in #45, it needs updating here; the walk-to-end version is immune.

## Worked examples

**Example 1 — the rewiring, node by node:**

```
before: a(1) → b(2) → c(3) → nil        head = a
after:  c(3) → b(2) → a(1) → nil        head = c
```

Same three nodes. `c.Next` now points at `b`, `b.Next` at `a`, `a.Next` is nil. The tests hold pointers to a, b, c and check exactly these four facts.

**Example 2 — through the List API:**

```go
var l List
l.Add(1); l.Add(2); l.Add(3)
l.Reverse()
l.Values()   // → [3 2 1]
l.Length()   // → 3
l.Reverse()
l.Values()   // → [1 2 3]  — an involution: twice = identity
```

**Example 3 — the degenerate cases:**

```go
ReverseChain(nil)          // → nil
n := &Node{Val: 7}
ReverseChain(n)            // → n itself, n.Next still nil
```

## Edge cases the tests cover

- nil chain; single node (returned by identity).
- Two nodes — the smallest case where the pointer dance can go wrong.
- Pointer-identity assertions on a hand-built 3-node chain (no allocation allowed).
- Old head's `Next` is nil after reversal (the forgotten-termination bug).
- Empty list `Reverse()`; single-element list.
- `Values` reversed, `Length` unchanged, reverse-twice-restores.
- `Add` after `Reverse` appends at the *new* tail.
- Duplicates and negative values (rewiring is value-blind).
- 10,000-node reverse: correct order, correct length, double-reverse restores.
