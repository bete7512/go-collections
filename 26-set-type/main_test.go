package main

import "testing"

// type Set struct{ m map[string]struct{} }

// func NewSet(items ...string) *Set
// func (s *Set) Add(v string)
// func (s *Set) Has(v string) bool
// func (s *Set) Remove(v string)
// func (s *Set) Len() int

func TestSetLifecycle(t *testing.T) {
	s := NewSet()

	if s.Len() != 0 {
		t.Fatalf("expected empty set, got len=%d", s.Len())
	}

	s.Add("apple")

	if s.Len() != 1 {
		t.Fatalf("after add: expected len=1, got %d", s.Len())
	}

	if !s.Has("apple") {
		t.Fatal("expected apple to exist")
	}

	if s.Has("banana") {
		t.Fatal("did not expect banana to exist")
	}

	s.Remove("apple")

	if s.Len() != 0 {
		t.Fatalf("after remove: expected len=0, got %d", s.Len())
	}

	if s.Has("apple") {
		t.Fatal("expected apple to be removed")
	}
}

func TestSetConstructor(t *testing.T) {
	s := NewSet("a", "a", "b")

	if s.Len() != 2 {
		t.Fatalf("expected duplicate values ignored, got len=%d", s.Len())
	}

	if !s.Has("a") {
		t.Fatal("expected a to exist")
	}

	if !s.Has("b") {
		t.Fatal("expected b to exist")
	}
}

func TestSetDoubleAdd(t *testing.T) {
	s := NewSet()

	s.Add("x")
	s.Add("x")

	if s.Len() != 1 {
		t.Fatalf("expected len=1 after duplicate add, got %d", s.Len())
	}
}

func TestSetRemoveAbsent(t *testing.T) {
	s := NewSet("a")

	s.Remove("missing")

	if s.Len() != 1 {
		t.Fatalf("remove absent item changed length, got %d", s.Len())
	}

	if !s.Has("a") {
		t.Fatal("expected existing item to remain")
	}
}

func TestSetRemoveUntilEmpty(t *testing.T) {
	s := NewSet("a", "b", "c")

	s.Remove("a")

	if s.Len() != 2 {
		t.Fatalf("expected len=2, got %d", s.Len())
	}

	s.Remove("b")
	s.Remove("c")

	if s.Len() != 0 {
		t.Fatalf("expected empty set, got len=%d", s.Len())
	}
}

func TestSetEmptyHas(t *testing.T) {
	s := NewSet()

	if s.Has("anything") {
		t.Fatal("empty set should not contain values")
	}
}

func TestSetZeroValue(t *testing.T) {
	var s Set

	// Assumes zero value is supported through lazy initialization.
	s.Add("hello")

	if !s.Has("hello") {
		t.Fatal("expected zero-value set to support Add")
	}

	if s.Len() != 1 {
		t.Fatalf("expected len=1, got %d", s.Len())
	}
}
