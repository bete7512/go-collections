package main

import (
	"math/rand"
	"testing"
)

func TestEmptyTree(t *testing.T) {
	var tree BST

	if tree.Contains(5) {
		t.Errorf("Contains on empty tree = true, want false")
	}
	if got := tree.Size(); got != 0 {
		t.Errorf("Size of empty tree = %d, want 0", got)
	}
}

func TestSingleNode(t *testing.T) {
	var tree BST
	tree.Insert(42)

	if !tree.Contains(42) {
		t.Errorf("Contains(42) = false, want true")
	}
	if tree.Contains(41) || tree.Contains(43) {
		t.Errorf("Contains reports values that were never inserted")
	}
	if got := tree.Size(); got != 1 {
		t.Errorf("Size = %d, want 1", got)
	}
}

func TestExampleTree(t *testing.T) {
	var tree BST
	for _, v := range []int{5, 3, 8, 1} {
		tree.Insert(v)
	}

	for _, v := range []int{5, 3, 8, 1} {
		if !tree.Contains(v) {
			t.Errorf("Contains(%d) = false, want true", v)
		}
	}
	// Probe BETWEEN, BELOW, and ABOVE the stored values: these walks cross
	// both subtrees and expose invariant violations one-sided probes miss.
	for _, v := range []int{0, 2, 4, 6, 7, 9} {
		if tree.Contains(v) {
			t.Errorf("Contains(%d) = true, want false", v)
		}
	}
	if got := tree.Size(); got != 4 {
		t.Errorf("Size = %d, want 4", got)
	}
}

func TestDuplicatesIgnored(t *testing.T) {
	var tree BST
	tree.Insert(5)
	tree.Insert(5)
	tree.Insert(5)

	if got := tree.Size(); got != 1 {
		t.Errorf("Size after triple insert of 5 = %d, want 1 (duplicates must be ignored)", got)
	}
	if !tree.Contains(5) {
		t.Errorf("Contains(5) = false, want true")
	}

	tree.Insert(3)
	tree.Insert(8)
	tree.Insert(3) // duplicate of a non-root node
	tree.Insert(8)

	if got := tree.Size(); got != 3 {
		t.Errorf("Size = %d, want 3 — non-root duplicates must be ignored too", got)
	}
}

func TestDegenerateAscending(t *testing.T) {
	var tree BST
	for v := 1; v <= 100; v++ {
		tree.Insert(v)
	}

	for v := 1; v <= 100; v++ {
		if !tree.Contains(v) {
			t.Fatalf("ascending-insert tree: Contains(%d) = false, want true", v)
		}
	}
	if tree.Contains(0) || tree.Contains(101) {
		t.Errorf("degenerate tree reports absent boundary values")
	}
	if got := tree.Size(); got != 100 {
		t.Errorf("Size = %d, want 100", got)
	}
}

func TestDegenerateDescending(t *testing.T) {
	var tree BST
	for v := 100; v >= 1; v-- {
		tree.Insert(v)
	}

	for v := 1; v <= 100; v++ {
		if !tree.Contains(v) {
			t.Fatalf("descending-insert tree: Contains(%d) = false, want true", v)
		}
	}
	if got := tree.Size(); got != 100 {
		t.Errorf("Size = %d, want 100", got)
	}
}

func TestNegativeAndZero(t *testing.T) {
	var tree BST
	for _, v := range []int{0, -5, 5, -10, 10} {
		tree.Insert(v)
	}

	for _, v := range []int{0, -5, 5, -10, 10} {
		if !tree.Contains(v) {
			t.Errorf("Contains(%d) = false, want true", v)
		}
	}
	for _, v := range []int{-11, -1, 1, 11} {
		if tree.Contains(v) {
			t.Errorf("Contains(%d) = true, want false", v)
		}
	}
}

func TestInterleavedInsertContains(t *testing.T) {
	var tree BST

	tree.Insert(50)
	if !tree.Contains(50) || tree.Contains(25) {
		t.Fatalf("after Insert(50): Contains(50)=%v Contains(25)=%v, want true false",
			tree.Contains(50), tree.Contains(25))
	}

	tree.Insert(25)
	tree.Insert(75)
	if !tree.Contains(25) || !tree.Contains(75) {
		t.Fatalf("newly inserted 25 and 75 not found")
	}
	if tree.Contains(60) {
		t.Fatalf("Contains(60) = true before inserting it")
	}

	tree.Insert(60)
	if !tree.Contains(60) {
		t.Fatalf("Contains(60) = false after inserting it")
	}
	if got := tree.Size(); got != 4 {
		t.Errorf("Size = %d, want 4", got)
	}
}

func TestLargeShuffled(t *testing.T) {
	const n = 1000
	rng := rand.New(rand.NewSource(47))

	// Insert the EVEN numbers 0..2n-2 shuffled; every odd number is a
	// guaranteed-absent probe that lands between stored values.
	values := make([]int, n)
	for i := range values {
		values[i] = i * 2
	}
	rng.Shuffle(len(values), func(i, j int) { values[i], values[j] = values[j], values[i] })

	var tree BST
	for _, v := range values {
		tree.Insert(v)
	}

	if got := tree.Size(); got != n {
		t.Fatalf("Size = %d, want %d", got, n)
	}
	for _, v := range values {
		if !tree.Contains(v) {
			t.Fatalf("Contains(%d) = false, want true", v)
		}
	}
	for i := 0; i < n; i++ {
		odd := i*2 + 1
		if tree.Contains(odd) {
			t.Fatalf("Contains(%d) = true, want false — never inserted", odd)
		}
	}
}
