package main

import (
	"math/rand"
	"slices"
	"testing"
)

func TestInOrderEmpty(t *testing.T) {
	var tree BST
	if got := tree.InOrder(); len(got) != 0 {
		t.Errorf("InOrder on empty tree = %v, want empty", got)
	}
}

func TestInOrderSingle(t *testing.T) {
	var tree BST
	tree.Insert(42)
	if got := tree.InOrder(); !slices.Equal(got, []int{42}) {
		t.Errorf("InOrder = %v, want [42]", got)
	}
}

func TestInOrderExampleTree(t *testing.T) {
	var tree BST
	for _, v := range []int{5, 3, 8, 1} {
		tree.Insert(v)
	}
	if got := tree.InOrder(); !slices.Equal(got, []int{1, 3, 5, 8}) {
		t.Errorf("InOrder = %v, want [1 3 5 8]", got)
	}
}

func TestInOrderShapeIndependent(t *testing.T) {
	orders := [][]int{
		{1, 3, 5, 8}, // ascending — right-skewed tree
		{8, 5, 3, 1}, // descending — left-skewed tree
		{3, 8, 1, 5}, // mixed
		{5, 8, 3, 1}, // another mix
	}
	want := []int{1, 3, 5, 8}

	for _, order := range orders {
		var tree BST
		for _, v := range order {
			tree.Insert(v)
		}
		if got := tree.InOrder(); !slices.Equal(got, want) {
			t.Errorf("insert order %v: InOrder = %v, want %v (shape must not matter)", order, got, want)
		}
	}
}

func TestInOrderSkewedLarge(t *testing.T) {
	// Both degenerate shapes, 100 nodes each.
	var asc, desc BST
	want := make([]int, 100)
	for i := 0; i < 100; i++ {
		asc.Insert(i)
		desc.Insert(99 - i)
		want[i] = i
	}

	if got := asc.InOrder(); !slices.Equal(got, want) {
		t.Errorf("right-skewed tree: InOrder wrong (first 5: %v, want %v)", got[:5], want[:5])
	}
	if got := desc.InOrder(); !slices.Equal(got, want) {
		t.Errorf("left-skewed tree: InOrder wrong (first 5: %v, want %v)", got[:5], want[:5])
	}
}

func TestInOrderNoDuplicates(t *testing.T) {
	var tree BST
	for _, v := range []int{5, 3, 5, 8, 3, 5, 1, 1} {
		tree.Insert(v)
	}
	if got := tree.InOrder(); !slices.Equal(got, []int{1, 3, 5, 8}) {
		t.Errorf("InOrder = %v, want [1 3 5 8] — duplicates ignored at insert", got)
	}
}

func TestInOrderNegatives(t *testing.T) {
	var tree BST
	for _, v := range []int{0, -5, 5, -10, 10, -1, 1} {
		tree.Insert(v)
	}
	want := []int{-10, -5, -1, 0, 1, 5, 10}
	if got := tree.InOrder(); !slices.Equal(got, want) {
		t.Errorf("InOrder = %v, want %v", got, want)
	}
}

func TestInOrderIsReadOnly(t *testing.T) {
	var tree BST
	for _, v := range []int{2, 1, 3} {
		tree.Insert(v)
	}

	first := tree.InOrder()
	second := tree.InOrder()

	if !slices.Equal(first, second) {
		t.Errorf("two consecutive InOrder calls differ: %v then %v — traversal mutated the tree?", first, second)
	}
	if got := tree.Size(); got != 3 {
		t.Errorf("Size after traversals = %d, want 3", got)
	}
}

func TestInOrderInterleavedWithInserts(t *testing.T) {
	var tree BST
	steps := []struct {
		insert int
		want   []int
	}{
		{5, []int{5}},
		{2, []int{2, 5}},
		{8, []int{2, 5, 8}},
		{1, []int{1, 2, 5, 8}},
		{3, []int{1, 2, 3, 5, 8}},
		{9, []int{1, 2, 3, 5, 8, 9}},
	}

	for i, s := range steps {
		tree.Insert(s.insert)
		got := tree.InOrder()
		if !slices.Equal(got, s.want) {
			t.Fatalf("step %d (insert %d): InOrder = %v, want %v", i, s.insert, got, s.want)
		}
		if len(got) != tree.Size() {
			t.Fatalf("step %d: len(InOrder()) = %d but Size() = %d", i, len(got), tree.Size())
		}
	}
}

func TestInOrderSortedProperty(t *testing.T) {
	// Insert shuffles of 0..999; the output must be exactly 0..999 in order.
	// The expected slice is built by construction — no sorting in this test.
	const n = 1000
	want := make([]int, n)
	for i := range want {
		want[i] = i
	}

	rng := rand.New(rand.NewSource(48))
	for round := 0; round < 5; round++ {
		values := slices.Clone(want)
		rng.Shuffle(len(values), func(i, j int) { values[i], values[j] = values[j], values[i] })

		var tree BST
		for _, v := range values {
			tree.Insert(v)
		}

		got := tree.InOrder()
		if !slices.Equal(got, want) {
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("round %d: first mismatch at index %d: got %d, want %d", round, i, got[i], want[i])
				}
			}
			t.Fatalf("round %d: length mismatch: got %d, want %d", round, len(got), n)
		}
	}
}
