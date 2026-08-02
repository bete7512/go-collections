package main

import "testing"

func TestZeroValueStack(t *testing.T) {
	var s Stack

	if !s.Empty() {
		t.Errorf("zero-value stack: Empty() = false, want true")
	}
	if got := s.Len(); got != 0 {
		t.Errorf("zero-value stack: Len() = %d, want 0", got)
	}
	if v, ok := s.Pop(); ok || v != 0 {
		t.Errorf("Pop on empty = (%d, %v), want (0, false)", v, ok)
	}
	if v, ok := s.Peek(); ok || v != 0 {
		t.Errorf("Peek on empty = (%d, %v), want (0, false)", v, ok)
	}

	// Still usable after the empty-ops.
	s.Push(1)
	if got := s.Len(); got != 1 {
		t.Errorf("after Push on zero value: Len() = %d, want 1", got)
	}
}

func TestLIFOOrder(t *testing.T) {
	var s Stack
	s.Push(1)
	s.Push(2)
	s.Push(3)

	for _, want := range []int{3, 2, 1} {
		v, ok := s.Pop()
		if !ok {
			t.Fatalf("Pop = (_, false) with %d elements expected remaining", want)
		}
		if v != want {
			t.Fatalf("Pop = %d, want %d (LIFO order)", v, want)
		}
	}
	if v, ok := s.Pop(); ok || v != 0 {
		t.Errorf("Pop after drain = (%d, %v), want (0, false)", v, ok)
	}
}

func TestPeekDoesNotConsume(t *testing.T) {
	var s Stack
	s.Push(7)

	for i := 0; i < 3; i++ {
		v, ok := s.Peek()
		if !ok || v != 7 {
			t.Fatalf("Peek #%d = (%d, %v), want (7, true)", i, v, ok)
		}
		if got := s.Len(); got != 1 {
			t.Fatalf("Len after Peek #%d = %d, want 1 — Peek must not remove", i, got)
		}
	}

	v, ok := s.Pop()
	if !ok || v != 7 {
		t.Errorf("Pop after Peek = (%d, %v), want (7, true) — same value Peek reported", v, ok)
	}
}

func TestNegativeAndZeroValues(t *testing.T) {
	// A sentinel-based API (returning -1 or 0 for "empty") cannot pass this.
	var s Stack
	s.Push(-1)
	s.Push(0)

	if v, ok := s.Pop(); !ok || v != 0 {
		t.Errorf("Pop = (%d, %v), want (0, true) — a real pushed zero", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != -1 {
		t.Errorf("Pop = (%d, %v), want (-1, true) — a real pushed -1", v, ok)
	}
	if _, ok := s.Pop(); ok {
		t.Errorf("Pop on drained stack reports ok = true, want false")
	}
}

func TestDuplicates(t *testing.T) {
	var s Stack
	s.Push(5)
	s.Push(5)
	s.Push(5)

	if got := s.Len(); got != 3 {
		t.Fatalf("Len = %d, want 3 — duplicates are distinct elements", got)
	}
	for i := 0; i < 3; i++ {
		if v, ok := s.Pop(); !ok || v != 5 {
			t.Fatalf("Pop #%d = (%d, %v), want (5, true)", i, v, ok)
		}
	}
}

func TestInterleavedPushPop(t *testing.T) {
	var s Stack

	s.Push(1)
	s.Push(2)
	if v, _ := s.Pop(); v != 2 {
		t.Fatalf("Pop = %d, want 2", v)
	}
	s.Push(3)
	s.Push(4)
	if v, _ := s.Pop(); v != 4 {
		t.Fatalf("Pop = %d, want 4", v)
	}
	if v, _ := s.Pop(); v != 3 {
		t.Fatalf("Pop = %d, want 3", v)
	}
	if v, _ := s.Pop(); v != 1 {
		t.Fatalf("Pop = %d, want 1", v)
	}
	if !s.Empty() {
		t.Errorf("Empty() = false after popping everything")
	}
}

func TestReusableAfterDrain(t *testing.T) {
	var s Stack
	s.Push(1)
	s.Pop()

	s.Push(42)
	if v, ok := s.Peek(); !ok || v != 42 {
		t.Errorf("Peek after drain-and-reuse = (%d, %v), want (42, true)", v, ok)
	}
	if got := s.Len(); got != 1 {
		t.Errorf("Len = %d, want 1", got)
	}
}

func TestLenTracksEveryOperation(t *testing.T) {
	var s Stack
	checks := []struct {
		op      func()
		wantLen int
	}{
		{func() { s.Push(1) }, 1},
		{func() { s.Push(2) }, 2},
		{func() { s.Peek() }, 2},
		{func() { s.Pop() }, 1},
		{func() { s.Push(3) }, 2},
		{func() { s.Pop() }, 1},
		{func() { s.Pop() }, 0},
		{func() { s.Pop() }, 0}, // popping empty stays 0
	}

	for i, c := range checks {
		c.op()
		if got := s.Len(); got != c.wantLen {
			t.Fatalf("step %d: Len() = %d, want %d", i, got, c.wantLen)
		}
		if s.Empty() != (c.wantLen == 0) {
			t.Fatalf("step %d: Empty() = %v inconsistent with Len %d", i, s.Empty(), c.wantLen)
		}
	}
}

func TestLargeStack(t *testing.T) {
	const n = 10_000
	var s Stack

	for i := 0; i < n; i++ {
		s.Push(i)
	}
	if got := s.Len(); got != n {
		t.Fatalf("Len = %d, want %d", got, n)
	}

	for want := n - 1; want >= 0; want-- {
		v, ok := s.Pop()
		if !ok || v != want {
			t.Fatalf("Pop = (%d, %v), want (%d, true)", v, ok, want)
		}
	}
	if !s.Empty() {
		t.Errorf("Empty() = false after draining %d elements", n)
	}
}
