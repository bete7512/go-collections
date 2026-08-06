package main

import "fmt"

func main() {
	t := &BST{}

	for _, v := range []int{50, 30, 70, 20, 40} {
		t.Insert(v)
	}

	fmt.Println("size:", t.size)

	var out []int
	inorder(t.root, &out)
	fmt.Println("in-order:", out) // sorted if Insert is correct
}

func inorder(n *TreeNode, out *[]int) {
	if n == nil {
		return
	}
	inorder(n.Left, out)
	*out = append(*out, n.Val)
	inorder(n.Right, out)
}

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
		return
	}
	cur := t.root
	for {
		if cur.Val == v {
			return
		}
		newNode := &TreeNode{Val: v}
		if v < cur.Val {
			if cur.Left != nil {
				cur = cur.Left
				continue
			}
			cur.Left = newNode
			break
		} else {
			if cur.Right != nil {
				cur = cur.Right
				continue
			}
			cur.Right = newNode
			break
		}
	}
	t.size++
}
func (t *BST) Contains(v int) bool {
	if t.root == nil {
		return false
	}
	cur := t.root
	for {
		if cur.Val == v {
			return true
		}
		if v < cur.Val {
			if cur.Left != nil {
				cur = cur.Left
				continue
			}
			return false
		} else {
			if cur.Right != nil {
				cur = cur.Right
				continue
			}
			return false
		}
	}
}
func (t *BST) Size() int {
	return t.size
}
