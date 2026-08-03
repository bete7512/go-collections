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
		l.head = &Node{Val: v}
		l.n++
		return
	}

	cur := l.head
	for cur.Next != nil {
		cur = cur.Next
	}
	cur.Next = &Node{Val: v}
	l.n++
}
func (l *List) Remove(v int) bool {
	if l.head == nil {
		return false
	}

	// head is a special case: no prev exists
	if l.head.Val == v {
		l.head = l.head.Next
		l.n--
		return true
	}

	prev := l.head
	for cur := l.head.Next; cur != nil; cur = cur.Next {
		if cur.Val == v {
			prev.Next = cur.Next
			l.n--
			return true
		}
		prev = cur
	}
	return false
}

func (l *List) Length() int {
	return l.n
}
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
}
