package main

import (
	"slices"
	"testing"
)

func chainValues(head *Node) []int {
	vals := []int{}
	for n := head; n != nil; n = n.Next {
		vals = append(vals, n.Val)
	}
	return vals
}

func TestReverseChainNil(t *testing.T) {
	if got := ReverseChain(nil); got != nil {
		t.Errorf("ReverseChain(nil) = %v, want nil", got)
	}
}

func TestReverseChainSingleNode(t *testing.T) {
	n := &Node{Val: 7}

	got := ReverseChain(n)

	if got != n {
		t.Fatalf("ReverseChain(single) returned a different node — must return the same node, not a copy")
	}
	if n.Next != nil {
		t.Errorf("single node's Next = %v, want nil", n.Next)
	}
}

func TestReverseChainPointerIdentity(t *testing.T) {
	// Build a → b → c by hand and keep the pointers. After reversal the SAME
	// nodes must be rewired: new head is c, c.Next is b, b.Next is a, a.Next nil.
	c := &Node{Val: 3}
	b := &Node{Val: 2, Next: c}
	a := &Node{Val: 1, Next: b}

	got := ReverseChain(a)

	if got != c {
		t.Fatalf("new head is not the old tail node (allocated new nodes instead of rewiring?)")
	}
	if c.Next != b {
		t.Errorf("c.Next is not the original b node")
	}
	if b.Next != a {
		t.Errorf("b.Next is not the original a node")
	}
	if a.Next != nil {
		t.Errorf("old head's Next = %v, want nil — the termination bug", a.Next)
	}
	if !slices.Equal(chainValues(got), []int{3, 2, 1}) {
		t.Errorf("values after reversal = %v, want [3 2 1]", chainValues(got))
	}
}

func TestReverseChainTwoNodes(t *testing.T) {
	b := &Node{Val: 2}
	a := &Node{Val: 1, Next: b}

	got := ReverseChain(a)

	if got != b {
		t.Fatalf("new head is not the old second node")
	}
	if b.Next != a || a.Next != nil {
		t.Errorf("two-node rewiring wrong: b.Next==a is %v, a.Next==nil is %v", b.Next == a, a.Next == nil)
	}
}

func TestListReverseEmpty(t *testing.T) {
	var l List
	l.Reverse() // must not panic
	if got := l.Values(); len(got) != 0 {
		t.Errorf("Values after reversing empty list = %v, want []", got)
	}
	if got := l.Length(); got != 0 {
		t.Errorf("Length = %d, want 0", got)
	}
}

func TestListReverseSingle(t *testing.T) {
	var l List
	l.Add(42)
	l.Reverse()
	if got := l.Values(); !slices.Equal(got, []int{42}) {
		t.Errorf("Values = %v, want [42]", got)
	}
}

func TestListReverseOrder(t *testing.T) {
	var l List
	for _, v := range []int{1, 2, 3, 4, 5} {
		l.Add(v)
	}

	l.Reverse()

	if got := l.Values(); !slices.Equal(got, []int{5, 4, 3, 2, 1}) {
		t.Errorf("Values after Reverse = %v, want [5 4 3 2 1]", got)
	}
	if got := l.Length(); got != 5 {
		t.Errorf("Length after Reverse = %d, want 5 — reversal must not change the count", got)
	}
}

func TestListReverseTwiceRestores(t *testing.T) {
	var l List
	want := []int{3, 1, 4, 1, 5, 9, 2, 6}
	for _, v := range want {
		l.Add(v)
	}

	l.Reverse()
	l.Reverse()

	if got := l.Values(); !slices.Equal(got, want) {
		t.Errorf("Values after double Reverse = %v, want %v", got, want)
	}
}

func TestAddAfterReverse(t *testing.T) {
	var l List
	l.Add(1)
	l.Add(2)
	l.Add(3)

	l.Reverse() // now [3 2 1]
	l.Add(0)    // must append at the NEW tail

	if got := l.Values(); !slices.Equal(got, []int{3, 2, 1, 0}) {
		t.Errorf("Values after Reverse+Add = %v, want [3 2 1 0] (stale tail handling?)", got)
	}
	if got := l.Length(); got != 4 {
		t.Errorf("Length = %d, want 4", got)
	}
}

func TestReverseWithDuplicatesAndNegatives(t *testing.T) {
	var l List
	input := []int{-1, 0, -1, 7, 0}
	for _, v := range input {
		l.Add(v)
	}

	l.Reverse()

	if got := l.Values(); !slices.Equal(got, []int{0, 7, -1, 0, -1}) {
		t.Errorf("Values = %v, want [0 7 -1 0 -1]", got)
	}
}

func TestReverseLarge(t *testing.T) {
	const n = 10_000
	var l List
	want := make([]int, n)
	for i := 0; i < n; i++ {
		l.Add(i)
		want[n-1-i] = i
	}

	l.Reverse()

	if got := l.Values(); !slices.Equal(got, want) {
		t.Fatalf("large reverse wrong (first few: got %v..., want %v...)", got[:5], want[:5])
	}
	if got := l.Length(); got != n {
		t.Errorf("Length = %d, want %d", got, n)
	}

	l.Reverse()
	vals := l.Values()
	for i := 0; i < n; i++ {
		if vals[i] != i {
			t.Fatalf("double reverse broke order at index %d: got %d, want %d", i, vals[i], i)
		}
	}
}
