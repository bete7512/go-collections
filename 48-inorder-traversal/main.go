package main

func main() {}

type TreeNode struct {
	Val         int
	Left, Right *TreeNode
}

type BST struct {
	root *TreeNode
	size int
}

func (t *BST) Insert(v int) {
	if t.root == nil {
		t.root = &TreeNode{Val: v}
		t.size++
	}

	cur := t.root

	for {
		if v == cur.Val {
			return
		}
		if v < cur.Val {
			if cur.Left != nil {
				cur = cur.Left
				continue
			}
			cur.Left = &TreeNode{Val: v}
			break
		} else {
			if cur.Right != nil {
				cur = cur.Right
				continue
			}
			cur.Right = &TreeNode{Val: v}
			break
		}
	}
	t.size++

}
func (t *BST) Size() int {
	return t.size
}
func (t *BST) InOrder() []int {
	values := make([]int, 0, t.size)
	inOrder(t.root, &values)
	return values
} // left, node, right — the whole tree
func inOrder(node *TreeNode, out *[]int) {
	if node == nil {
		return
	}
	inOrder(node.Left, out)
	*out = append(*out, node.Val)
	inOrder(node.Right, out)
}
