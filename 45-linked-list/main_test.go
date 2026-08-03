package main

import (
	"slices"
	"testing"
)

// check asserts values and the Length/Values consistency invariant in one go.
func check(t *testing.T, l *List, want []int) {
	t.Helper()
	got := l.Values()
	if !slices.Equal(got, want) {
		t.Fatalf("Values() = %v, want %v", got, want)
	}
	if l.Length() != len(got) {
		t.Fatalf("Length() = %d but Values() has %d elements — counter drifted", l.Length(), len(got))
	}
}

func TestZeroValueList(t *testing.T) {
	var l List
	check(t, &l, []int{})

	if l.Remove(5) {
		t.Errorf("Remove on empty list = true, want false")
	}
	check(t, &l, []int{})

	l.Add(1)
	check(t, &l, []int{1})
}

func TestAddAppendsAtTail(t *testing.T) {
	var l List
	l.Add(1)
	check(t, &l, []int{1})
	l.Add(2)
	check(t, &l, []int{1, 2})
	l.Add(3)
	check(t, &l, []int{1, 2, 3})
}

func TestRemoveMiddle(t *testing.T) {
	var l List
	l.Add(1)
	l.Add(2)
	l.Add(3)

	if !l.Remove(2) {
		t.Fatalf("Remove(2) = false, want true")
	}
	check(t, &l, []int{1, 3})
}

func TestRemoveHead(t *testing.T) {
	var l List
	l.Add(1)
	l.Add(2)
	l.Add(3)

	if !l.Remove(1) {
		t.Fatalf("Remove(head) = false, want true")
	}
	check(t, &l, []int{2, 3})

	// Remove the new head too — the head path must work repeatedly.
	if !l.Remove(2) {
		t.Fatalf("Remove(new head) = false, want true")
	}
	check(t, &l, []int{3})
}

func TestRemoveTail(t *testing.T) {
	var l List
	l.Add(1)
	l.Add(2)
	l.Add(3)

	if !l.Remove(3) {
		t.Fatalf("Remove(tail) = false, want true")
	}
	check(t, &l, []int{1, 2})

	// Add after removing the tail: the list must still append correctly
	// (a stale tail pointer shows up exactly here).
	l.Add(9)
	check(t, &l, []int{1, 2, 9})
}

func TestRemoveOnlyNode(t *testing.T) {
	var l List
	l.Add(42)

	if !l.Remove(42) {
		t.Fatalf("Remove(only node) = false, want true")
	}
	check(t, &l, []int{})

	// Reusable after becoming empty.
	l.Add(7)
	check(t, &l, []int{7})
}

func TestRemoveAbsent(t *testing.T) {
	var l List
	l.Add(1)
	l.Add(2)

	if l.Remove(99) {
		t.Errorf("Remove(absent) = true, want false")
	}
	check(t, &l, []int{1, 2})
}

func TestRemoveFirstMatchOnly(t *testing.T) {
	var l List
	l.Add(7)
	l.Add(8)
	l.Add(7)

	if !l.Remove(7) {
		t.Fatalf("Remove(7) = false, want true")
	}
	check(t, &l, []int{8, 7}) // the SECOND 7 survives

	if !l.Remove(7) {
		t.Fatalf("second Remove(7) = false, want true")
	}
	check(t, &l, []int{8})
}

func TestZeroAndNegativeValues(t *testing.T) {
	var l List
	l.Add(0)
	l.Add(-5)
	l.Add(0)

	check(t, &l, []int{0, -5, 0})

	if !l.Remove(0) {
		t.Fatalf("Remove(0) = false, want true — zero is a real value")
	}
	check(t, &l, []int{-5, 0})

	if !l.Remove(-5) {
		t.Fatalf("Remove(-5) = false, want true")
	}
	check(t, &l, []int{0})
}

func TestScriptedSequence(t *testing.T) {
	var l List
	steps := []struct {
		op   func() bool
		want []int
	}{
		{func() bool { l.Add(3); return true }, []int{3}},
		{func() bool { l.Add(1); return true }, []int{3, 1}},
		{func() bool { l.Add(4); return true }, []int{3, 1, 4}},
		{func() bool { return l.Remove(1) }, []int{3, 4}},
		{func() bool { l.Add(1); return true }, []int{3, 4, 1}},
		{func() bool { return l.Remove(3) }, []int{4, 1}},
		{func() bool { return l.Remove(1) }, []int{4}},
		{func() bool { return !l.Remove(9) }, []int{4}},
		{func() bool { return l.Remove(4) }, []int{}},
		{func() bool { l.Add(5); return true }, []int{5}},
	}

	for i, s := range steps {
		if !s.op() {
			t.Fatalf("step %d: operation reported unexpected result", i)
		}
		check(t, &l, s.want)
	}
}

func TestLargeList(t *testing.T) {
	const n = 1000
	var l List

	want := make([]int, 0, n)
	for i := 0; i < n; i++ {
		l.Add(i)
		want = append(want, i)
	}
	check(t, &l, want)

	// Remove front, middle, back.
	for _, v := range []int{0, n / 2, n - 1} {
		if !l.Remove(v) {
			t.Fatalf("Remove(%d) = false, want true", v)
		}
		idx := slices.Index(want, v)
		want = slices.Delete(want, idx, idx+1)
	}
	check(t, &l, want)
}
