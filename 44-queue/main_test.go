package main

import "testing"

func TestZeroValueQueue(t *testing.T) {
	var q Queue

	if !q.Empty() {
		t.Errorf("zero-value queue: Empty() = false, want true")
	}
	if got := q.Len(); got != 0 {
		t.Errorf("zero-value queue: Len() = %d, want 0", got)
	}
	if v, ok := q.Dequeue(); ok || v != 0 {
		t.Errorf("Dequeue on empty = (%d, %v), want (0, false)", v, ok)
	}

	q.Enqueue(1)
	if got := q.Len(); got != 1 {
		t.Errorf("after Enqueue on zero value: Len() = %d, want 1", got)
	}
}

func TestFIFOOrder(t *testing.T) {
	var q Queue
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	for _, want := range []int{1, 2, 3} {
		v, ok := q.Dequeue()
		if !ok {
			t.Fatalf("Dequeue = (_, false), want (%d, true)", want)
		}
		if v != want {
			t.Fatalf("Dequeue = %d, want %d (FIFO order)", v, want)
		}
	}
	if v, ok := q.Dequeue(); ok || v != 0 {
		t.Errorf("Dequeue after drain = (%d, %v), want (0, false)", v, ok)
	}
}

func TestZeroAndNegativeValues(t *testing.T) {
	var q Queue
	q.Enqueue(0)
	q.Enqueue(-1)

	if v, ok := q.Dequeue(); !ok || v != 0 {
		t.Errorf("Dequeue = (%d, %v), want (0, true) — a real enqueued zero", v, ok)
	}
	if v, ok := q.Dequeue(); !ok || v != -1 {
		t.Errorf("Dequeue = (%d, %v), want (-1, true)", v, ok)
	}
	if _, ok := q.Dequeue(); ok {
		t.Errorf("Dequeue on drained queue reports ok = true, want false")
	}
}

func TestDuplicates(t *testing.T) {
	var q Queue
	q.Enqueue(5)
	q.Enqueue(5)
	q.Enqueue(5)

	if got := q.Len(); got != 3 {
		t.Fatalf("Len = %d, want 3", got)
	}
	for i := 0; i < 3; i++ {
		if v, ok := q.Dequeue(); !ok || v != 5 {
			t.Fatalf("Dequeue #%d = (%d, %v), want (5, true)", i, v, ok)
		}
	}
}

func TestInterleaved(t *testing.T) {
	var q Queue

	q.Enqueue(1)
	q.Enqueue(2)
	if v, _ := q.Dequeue(); v != 1 {
		t.Fatalf("Dequeue = %d, want 1", v)
	}
	q.Enqueue(3)
	if v, _ := q.Dequeue(); v != 2 {
		t.Fatalf("Dequeue = %d, want 2 — 2 was ahead of 3", v)
	}
	q.Enqueue(4)
	q.Enqueue(5)
	for _, want := range []int{3, 4, 5} {
		if v, _ := q.Dequeue(); v != want {
			t.Fatalf("Dequeue = %d, want %d", v, want)
		}
	}
	if !q.Empty() {
		t.Errorf("Empty() = false after draining")
	}
}

func TestReusableAfterDrain(t *testing.T) {
	var q Queue
	q.Enqueue(1)
	q.Dequeue()

	q.Enqueue(42)
	if v, ok := q.Dequeue(); !ok || v != 42 {
		t.Errorf("Dequeue after drain-and-reuse = (%d, %v), want (42, true)", v, ok)
	}
}

func TestLenAndEmptyTrackEveryOperation(t *testing.T) {
	var q Queue
	steps := []struct {
		op      func()
		wantLen int
	}{
		{func() { q.Enqueue(1) }, 1},
		{func() { q.Enqueue(2) }, 2},
		{func() { q.Enqueue(3) }, 3},
		{func() { q.Dequeue() }, 2},
		{func() { q.Enqueue(4) }, 3},
		{func() { q.Dequeue() }, 2},
		{func() { q.Dequeue() }, 1},
		{func() { q.Dequeue() }, 0},
		{func() { q.Dequeue() }, 0}, // dequeue on empty stays 0
	}

	for i, s := range steps {
		s.op()
		if got := q.Len(); got != s.wantLen {
			t.Fatalf("step %d: Len() = %d, want %d", i, got, s.wantLen)
		}
		if q.Empty() != (s.wantLen == 0) {
			t.Fatalf("step %d: Empty() = %v inconsistent with Len %d", i, q.Empty(), s.wantLen)
		}
	}
}

func TestAlternatingAtLengthOne(t *testing.T) {
	// Enqueue/dequeue in lockstep: length never exceeds 1, but the slice
	// front keeps advancing — the pattern behind the [1:] memory caveat.
	// Correctness must hold regardless.
	var q Queue
	for i := 0; i < 1000; i++ {
		q.Enqueue(i)
		v, ok := q.Dequeue()
		if !ok || v != i {
			t.Fatalf("round %d: Dequeue = (%d, %v), want (%d, true)", i, v, ok, i)
		}
		if !q.Empty() {
			t.Fatalf("round %d: queue not empty after lockstep dequeue", i)
		}
	}
}

func TestLargeQueue(t *testing.T) {
	const n = 10_000
	var q Queue

	for i := 0; i < n; i++ {
		q.Enqueue(i)
	}
	if got := q.Len(); got != n {
		t.Fatalf("Len = %d, want %d", got, n)
	}

	for want := 0; want < n; want++ {
		v, ok := q.Dequeue()
		if !ok || v != want {
			t.Fatalf("Dequeue = (%d, %v), want (%d, true)", v, ok, want)
		}
	}
	if !q.Empty() {
		t.Errorf("Empty() = false after draining %d elements", n)
	}
}
