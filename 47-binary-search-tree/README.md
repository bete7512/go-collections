# 47 · Binary search tree

## Problem

A binary search tree is #45's linked list with two `Next` pointers and one rule. The rule — the **BST invariant** — is what makes it searchable: *every* value in a node's left subtree is less than the node's value, and *every* value in its right subtree is greater. Not just the immediate children: the entire subtrees. Insert and lookup both become "compare, go left or right, repeat," discarding half the (balanced) tree per step. This drill is also your first recursion over a recursive data structure — the tree's shape *is* the call tree.

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

func (t *BST) Insert(v int)
func (t *BST) Contains(v int) bool
func (t *BST) Size() int
```

- **`Insert` places v where the invariant demands**: walk from the root, left when v is smaller, right when larger, attach a new node at the nil you fall off. Inserting into an empty tree sets the root.
- **Duplicates are ignored** (set semantics — pinned so the tests can be exact): inserting a value already present changes nothing, and `Size` does not grow.
- **`Contains`** walks the same comparison path; `false` on an empty tree, no panic.
- **`Size`** is a maintained counter of distinct stored values — O(1), incremented only on inserts that actually add a node. The tests use it to catch silent duplicate-insertion.
- **Zero value usable:** `var t BST` works immediately.
- **Correctness must not depend on balance.** Inserting 1,2,3,4,5 in order produces a degenerate tree — every node's Left is nil, effectively a linked list. Lookups degrade to O(n) but must still be *correct*; the tests include exactly this case. Write a comment naming the problem and the fix you're not building (self-balancing trees: AVL, red-black — which is what `sort.Search` over a sorted slice, or a database B-tree, sidestep by construction).
- Implementation shape is your choice — iterative walk or a recursive helper (the recursive insert that *returns* the possibly-new subtree root, `insert(n *TreeNode, v int) *TreeNode`, handles the empty-tree case with zero special-casing; the `**TreeNode` walking pointer from #45's toolbox works too).

## Worked examples

**Example 1 — building a tree:**

```
Insert order: 5, 3, 8, 1

        5           5 is the root
       / \          3 < 5 → left of 5
      3   8         8 > 5 → right of 5
     /              1 < 5 → left; 1 < 3 → left of 3
    1
```

`Contains(3)` → true (5→left→found). `Contains(7)` → false (5→right to 8→left is nil → not there). `Size()` → 4.

**Example 2 — duplicates don't grow the tree:**

```go
var t BST
t.Insert(5)
t.Insert(5)
t.Insert(5)
t.Size()       // → 1
t.Contains(5)  // → true
```

**Example 3 — the degenerate case:**

```
Insert order: 1, 2, 3, 4

    1
     \
      2         every value went right — a linked list wearing
       \        a tree costume. Contains(4) still true; it just
        3       walked all four nodes instead of log₂(4) = 2.
         \
          4
```

## Edge cases the tests cover

- Empty tree: `Contains` false, `Size` 0, zero value usable.
- Single node; the example tree from inserts 5,3,8,1.
- Contains for every inserted value (true) and for values between, below, and above them (false) — the between-values probe catches invariant violations that same-side-only tests miss.
- Duplicate inserts: `Size` unchanged, `Contains` still true.
- Sorted-ascending and sorted-descending insert orders (both degenerate directions) — correctness at O(n) shape.
- Negative values and zero.
- 1,000 shuffled distinct values (fixed seed): every one `Contains` true, 1,000 absent values false, `Size` exact.
- Interleaved insert/contains sequences.
