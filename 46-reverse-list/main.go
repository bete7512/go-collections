package main

func main() {}

type Node struct {
	Val  int
	Next *Node
}

type List struct {
	head *Node
	n    int
}

func (l *List) Add(v int) {
	if l.head == nil {
		l.head = &Node{Val: v, Next: nil}
		l.n++
		return
	}

	cur := l.head

	for cur.Next != nil {
		cur = cur.Next
	}

	cur.Next = &Node{Val: v}
	l.n++
} // append at tail (from #45)
func (l *List) Values() []int {
	if l.head == nil {
		return []int{}
	}
	values := []int{}
	current := l.head
	for {
		values = append(values, current.Val)
		if current.Next == nil {
			break
		}
		current = current.Next
	}
	return values
} // head→tail traversal (from #45)
func (l *List) Length() int {
	return l.n
} // O(1) counter (from #45)
func (l *List) Reverse() {
	if l.head == nil || l.head.Next == nil {
		return
	}

	var prev *Node
	cur := l.head

	for cur != nil {
		next := cur.Next
		cur.Next = prev
		prev = cur
		cur = next
	}

	l.head = prev
} // reverse in place

func ReverseChain(head *Node) *Node {
	var prev *Node
	cur := head

	for cur != nil {
		next := cur.Next
		cur.Next = prev
		prev = cur
		cur = next
	}

	return prev
}
