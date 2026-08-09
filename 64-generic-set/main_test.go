package main

import (
	"slices"
	"testing"
)

type Point struct{ X, Y int }

type UserID int

func TestLifecycleStrings(t *testing.T) {
	s := NewSet[string]()

	if s.Len() != 0 || s.Has("a") {
		t.Fatalf("fresh set: Len=%d Has(a)=%v, want 0 false", s.Len(), s.Has("a"))
	}

	s.Add("a")
	s.Add("b")
	if s.Len() != 2 || !s.Has("a") || !s.Has("b") {
		t.Fatalf("after two adds: Len=%d, want 2 with both present", s.Len())
	}

	s.Add("a") // duplicate
	if s.Len() != 2 {
		t.Errorf("Len after duplicate Add = %d, want 2", s.Len())
	}

	s.Remove("a")
	if s.Len() != 1 || s.Has("a") {
		t.Errorf("after Remove: Len=%d Has(a)=%v, want 1 false", s.Len(), s.Has("a"))
	}

	s.Remove("nope") // absent
	if s.Len() != 1 {
		t.Errorf("Len after removing an absent element = %d, want 1", s.Len())
	}

	s.Remove("b")
	if s.Len() != 0 {
		t.Errorf("Len after removing everything = %d, want 0", s.Len())
	}

	s.Add("c") // reusable after emptying
	if s.Len() != 1 || !s.Has("c") {
		t.Errorf("set not reusable after being emptied")
	}
}

func TestManyElementTypes(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		s := NewSet(1, 2, 3)
		if s.Len() != 3 || !s.Has(2) || s.Has(9) {
			t.Errorf("int set misbehaved: Len=%d", s.Len())
		}
	})

	t.Run("struct", func(t *testing.T) {
		s := NewSet(Point{1, 2}, Point{3, 4}, Point{1, 2})
		if s.Len() != 2 {
			t.Errorf("Len = %d, want 2 — equal structs are the same element", s.Len())
		}
		if !s.Has(Point{1, 2}) || s.Has(Point{9, 9}) {
			t.Errorf("struct membership wrong")
		}
	})

	t.Run("array", func(t *testing.T) {
		s := NewSet([2]int{1, 2}, [2]int{1, 2}, [2]int{3, 4})
		if s.Len() != 2 {
			t.Errorf("Len = %d, want 2 — arrays are comparable element-wise", s.Len())
		}
		if !s.Has([2]int{3, 4}) {
			t.Errorf("array membership wrong")
		}
	})

	t.Run("named type", func(t *testing.T) {
		s := NewSet(UserID(1), UserID(2), UserID(1))
		if s.Len() != 2 || !s.Has(UserID(2)) {
			t.Errorf("named-type set misbehaved: Len=%d", s.Len())
		}
	})

	t.Run("pointer identity", func(t *testing.T) {
		a := &Point{1, 2}
		b := &Point{1, 2} // equal contents, different address
		s := NewSet(a, b)
		if s.Len() != 2 {
			t.Errorf("Len = %d, want 2 — pointers compare by address, not by pointee", s.Len())
		}
		if !s.Has(a) || !s.Has(b) {
			t.Errorf("pointer membership wrong")
		}
	})
}

func TestZeroValuesAreRealElements(t *testing.T) {
	ints := NewSet(0)
	if !ints.Has(0) || ints.Len() != 1 {
		t.Errorf("0 must be a normal element, not a sentinel")
	}

	strs := NewSet("")
	if !strs.Has("") || strs.Len() != 1 {
		t.Errorf("empty string must be a normal element")
	}

	points := NewSet(Point{})
	if !points.Has(Point{}) || points.Len() != 1 {
		t.Errorf("zero-value struct must be a normal element")
	}
}

func TestNewSetVariadic(t *testing.T) {
	if got := NewSet[int]().Len(); got != 0 {
		t.Errorf("NewSet() Len = %d, want 0", got)
	}
	if got := NewSet(5).Len(); got != 1 {
		t.Errorf("NewSet(5) Len = %d, want 1", got)
	}
	if got := NewSet(1, 2, 1, 3, 2).Len(); got != 3 {
		t.Errorf("NewSet with duplicates Len = %d, want 3", got)
	}
}

func TestNilSetReadsAreSafe(t *testing.T) {
	var s Set[int] // nil map

	if s.Has(1) {
		t.Errorf("Has on a nil set = true, want false")
	}
	if got := s.Len(); got != 0 {
		t.Errorf("Len on a nil set = %d, want 0", got)
	}
	s.Remove(1) // must not panic
	if got := s.Items(); len(got) != 0 {
		t.Errorf("Items on a nil set = %v, want empty", got)
	}
}

func TestNilSetAddPanics(t *testing.T) {
	// Writing to a nil map panics. This asymmetry is why NewSet exists.
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Add on a nil set did not panic — assignment to a nil map must panic; " +
				"if you lazily initialize, document that deviation")
		}
	}()

	var s Set[int]
	s.Add(1)
}

func TestCopiesShareTheUnderlyingMap(t *testing.T) {
	a := NewSet(1, 2)
	b := a // copies the map header, not the contents

	b.Add(3)

	if !a.Has(3) {
		t.Errorf("a.Has(3) = false after b.Add(3) — a Set value is a map, copies share storage")
	}
	if a.Len() != 3 {
		t.Errorf("a.Len() = %d, want 3", a.Len())
	}

	b.Remove(1)
	if a.Has(1) {
		t.Errorf("removal through the copy was not visible in the original")
	}
}

func TestItems(t *testing.T) {
	s := NewSet(3, 1, 2)

	got := s.Items()
	if len(got) != s.Len() {
		t.Fatalf("len(Items()) = %d, Len() = %d — they must agree", len(got), s.Len())
	}

	// Map order is unspecified: sort before comparing.
	slices.Sort(got)
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("sorted Items = %v, want [1 2 3]", got)
	}

	empty := NewSet[string]()
	if got := empty.Items(); len(got) != 0 {
		t.Errorf("Items on an empty set = %v, want empty", got)
	}
}

func TestLargeSet(t *testing.T) {
	const n = 10_000
	s := NewSet[int]()

	// Insert every value twice.
	for i := 0; i < n; i++ {
		s.Add(i)
		s.Add(i)
	}

	if s.Len() != n {
		t.Fatalf("Len = %d, want %d — duplicates must not grow the set", s.Len(), n)
	}
	for i := 0; i < n; i += 997 { // sparse sample
		if !s.Has(i) {
			t.Fatalf("Has(%d) = false", i)
		}
	}
	if s.Has(n) {
		t.Errorf("Has(%d) = true, want false", n)
	}

	for i := 0; i < n/2; i++ {
		s.Remove(i)
	}
	if s.Len() != n/2 {
		t.Errorf("Len after removing half = %d, want %d", s.Len(), n/2)
	}
}
