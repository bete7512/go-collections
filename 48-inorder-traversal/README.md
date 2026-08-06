# 48 · In-order traversal

## Problem

Walk #47's tree in *in-order* sequence — left subtree, node, right subtree, recursively — and collect the values into a slice. The payoff is a small piece of magic: **on a BST, in-order traversal yields the values in sorted order, always**, regardless of insertion order or tree shape. That's not a coincidence to memorize; it's the invariant becoming visible (everything smaller lives left, so visiting left-self-right *is* ascending order). It also means the test practically writes itself as a property: insert any shuffle, traverse, expect sorted — and that property test is stronger than any hand-picked example.

Own package, so `BST`/`Insert` get rebuilt from memory — #47 was yesterday's drill; today it's the warm-up.

## Contract (what the tests enforce)

```go
type TreeNode struct {
    Val         int
    Left, Right *TreeNode
}

type BST struct {
    root *TreeNode
    size int
}

func (t *BST) Insert(v int)       // as #47: BST invariant, duplicates ignored
func (t *BST) Size() int          // as #47
func (t *BST) InOrder() []int     // left, node, right — the whole tree
```

- **`InOrder` returns every stored value, ascending, exactly once.** For a BST built from any insert order, `InOrder()` equals the sorted distinct inserted values.
- Empty tree → empty slice (nil or empty both accepted), no panic.
- **Read-only:** traversal must not modify the tree. The tests call `InOrder` repeatedly and interleave inserts, expecting consistent, correct answers throughout. `len(InOrder())` must equal `Size()` at all times.
- Recursion is the natural shape (visit left, append node, visit right). The accumulation strategy is your design decision — a helper taking `*[]int`, a closure appending to a captured slice, or a helper returning slices to concatenate. They differ meaningfully in allocation behavior; pick one deliberately and note why in a comment. An iterative version with an explicit stack (#43's structure!) is a worthy bonus, not required.
- Duplicates were ignored at insert (per #47), so the output never contains repeats.

## Worked examples

**Example 1 — the tree from #47:**

```
Insert 5, 3, 8, 1:      In-order walk:
        5               visit left of 5 → subtree(3)
       / \                visit left of 3 → subtree(1) → emit 1
      3   8               emit 3 (its left is done)
     /                    right of 3 is nil
    1                   emit 5
                        visit right of 5 → emit 8
InOrder() → [1 3 5 8]   — sorted, from a tree built 5-first
```

**Example 2 — insertion order doesn't matter:**

```
Insert 1, 3, 5, 8  (ascending — degenerate right chain)
Insert 8, 5, 3, 1  (descending — degenerate left chain)
Insert 3, 8, 1, 5  (mixed)

All three trees have different SHAPES.
All three return InOrder() → [1 3 5 8].
```

The shape is an implementation detail; the invariant makes the traversal's output identical.

**Example 3 — the property that is the real test:**

```go
// insert 0..999 in a random shuffle
tree.InOrder()   // → exactly [0 1 2 ... 999], every time, any shuffle
```

## Edge cases the tests cover

- Empty tree; single node.
- The 5,3,8,1 example → `[1 3 5 8]`.
- Left-skewed (descending inserts) and right-skewed (ascending inserts) trees.
- Duplicate inserts → no repeated values in the output.
- Negative values and zero mixed in.
- The sorted-output property over multiple fixed-seed shuffles of 1,000 sequential values (expected slice built by construction, not by sorting in the test).
- `InOrder` called twice in a row → identical results (read-only proof).
- Insert-traverse-insert-traverse interleaving, with `len(InOrder()) == Size()` checked at every step.
