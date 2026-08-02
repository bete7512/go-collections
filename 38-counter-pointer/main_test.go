package main

import "testing"

func TestZeroValueUsable(t *testing.T) {
	var c Counter
	if got := c.Value(); got != 0 {
		t.Fatalf("zero-value Counter.Value() = %d, want 0", got)
	}
	c.Inc()
	if got := c.Value(); got != 1 {
		t.Fatalf("after one Inc on zero value: Value() = %d, want 1", got)
	}
}

func TestIncPersists(t *testing.T) {
	var c Counter
	c.Inc()
	c.Inc()
	c.Inc()
	if got := c.Value(); got != 3 {
		t.Errorf("after three Inc: Value() = %d, want 3", got)
	}
}

func TestIncManyTimes(t *testing.T) {
	var c Counter
	for i := 0; i < 1000; i++ {
		c.Inc()
	}
	if got := c.Value(); got != 1000 {
		t.Errorf("after 1000 Inc: Value() = %d, want 1000", got)
	}
}

func TestIncBrokenStaysZero(t *testing.T) {
	// IncBroken increments a COPY of the receiver. This test pins the bug:
	// the counter must remain untouched no matter how often it's called.
	var c Counter
	c.IncBroken()
	c.IncBroken()
	c.IncBroken()
	if got := c.Value(); got != 0 {
		t.Errorf("after three IncBroken: Value() = %d, want 0 — a value receiver must not mutate the original", got)
	}
}

func TestInterleavedIncAndIncBroken(t *testing.T) {
	var c Counter
	c.Inc()       // 1
	c.IncBroken() // no effect
	c.Inc()       // 2
	c.IncBroken() // no effect
	if got := c.Value(); got != 2 {
		t.Errorf("interleaved: Value() = %d, want 2 (only Inc calls count)", got)
	}
}

func TestCountersAreIndependent(t *testing.T) {
	var a, b Counter
	a.Inc()
	a.Inc()
	b.Inc()
	if a.Value() != 2 {
		t.Errorf("a.Value() = %d, want 2", a.Value())
	}
	if b.Value() != 1 {
		t.Errorf("b.Value() = %d, want 1", b.Value())
	}
}

func TestSliceElementIsAddressable(t *testing.T) {
	s := make([]Counter, 2)
	s[0].Inc()
	s[0].Inc()
	s[1].Inc()
	if got := s[0].Value(); got != 2 {
		t.Errorf("s[0].Value() = %d, want 2 — slice elements are addressable, Inc must mutate in place", got)
	}
	if got := s[1].Value(); got != 1 {
		t.Errorf("s[1].Value() = %d, want 1", got)
	}
}

func TestPointerSharesValueCopyDiverges(t *testing.T) {
	a := Counter{}
	p := &a
	p.Inc()
	if got := a.Value(); got != 1 {
		t.Fatalf("after p.Inc(): a.Value() = %d, want 1 — pointer and original are the same counter", got)
	}

	b := a // copy taken at value 1
	b.Inc()
	if got := a.Value(); got != 1 {
		t.Errorf("after b.Inc(): a.Value() = %d, want 1 — the copy's increments are its own", got)
	}
	if got := b.Value(); got != 2 {
		t.Errorf("b.Value() = %d, want 2", got)
	}
}
