# 45 · Singly linked list

## Problem

Build a linked list by hand: nodes pointing at nodes, a head pointer, and the pointer surgery slices have been sparing you. This is the foundational pointer-manipulation drill — #46 (reversal), #47 (trees, which are just nodes with two nexts), and #96 (the LRU's doubly-linked list) all build on the moves you learn here. The boss fight is `Remove`: **removing the head is where every naive implementation breaks**, because the head is reached by a different pointer (the list's `head` field) than every other node (the previous node's `Next`). There's an elegant way to make that asymmetry vanish — finding it is the drill.

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

func (l *List) Add(v int)           // append at the tail
func (l *List) Remove(v int) bool   // unlink the FIRST node with this value
func (l *List) Length() int         // O(1) — a maintained counter, not a walk
func (l *List) Values() []int       // traversal into a slice, for inspection
```

- **`Add` appends at the tail**: `Add(1); Add(2); Add(3)` → values `[1 2 3]`. (Walk-to-the-end is fine; a tail pointer making Add O(1) is a documented-in-a-comment upgrade — your choice, tests accept both.)
- **`Remove` unlinks the first node whose value matches** and reports whether it found one. `false` for absent values, `false` on an empty list — never a panic. Duplicates: only the first match goes.
- **`Length` is O(1)** — maintain `n` on every Add and successful Remove. The tests cross-check it against `len(Values())` after every operation, so a drifting counter fails fast.
- **`Values` walks head→tail** into a fresh slice; empty list → empty slice. (It exists so the tests can see inside; it's also your debugging tool.)
- **Zero value usable:** `var l List` works immediately.
- **The head-removal insight (required reading):** removing a middle node is `prev.Next = curr.Next`, but the head has no `prev` — its incoming pointer is `l.head` itself. Three ways out:
  1. Special-case branch (`if l.head.Val == v {...}`) — works, but two code paths that must both be right;
  2. **A `**Node` walking pointer**: start with `p := &l.head` and advance with `p = &(*p).Next`; then *every* node's incoming pointer — including the head's — is just `*p`, and removal is uniformly `*p = (*p).Next`. One code path, no special case;
  3. A dummy head node making the real head an ordinary middle node.
  The tests can't see which you chose — but try version 2 at least once; the moment `*p = (*p).Next` clicks, pointers-to-pointers stop being scary forever.

## Worked examples

**Example 1 — build and inspect:**

```go
var l List
l.Add(1); l.Add(2); l.Add(3)
l.Values()   // → [1 2 3]
l.Length()   // → 3
```

**Example 2 — the three removal positions:**

```go
// list: 1 → 2 → 3
l.Remove(2)   // middle: true, values [1 3]   (1.Next now points at 3)
l.Remove(1)   // HEAD:   true, values [3]     (l.head itself had to move)
l.Remove(3)   // last:   true, values []      (list now empty, head nil)
l.Remove(3)   // absent: false, values []
```

**Example 3 — duplicates, first match only:**

```go
// list: 7 → 8 → 7
l.Remove(7)   // → true, values [8 7] — the SECOND 7 survives
```

## Edge cases the tests cover

- Zero-value list: Length 0, Values empty, Remove safe.
- Remove the head; remove the tail; remove a middle node.
- Remove the only node (head becomes nil, list ends empty — then reusable).
- Remove from an empty list; remove an absent value.
- Duplicates: first match only, verified by what survives.
- Add after emptying the list completely.
- Length cross-checked against `len(Values())` after **every** operation in a scripted sequence.
- Zero and negative values stored (values are data, not signals).
- 1,000-element build, removals from front/middle/back, order intact throughout.
