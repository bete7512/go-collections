package main

import (
	"slices"
	"strconv"
	"strings"
	"testing"
)

type user struct {
	ID   int
	Name string
}

type wrapped struct{ V int }

func TestMapIntToString(t *testing.T) {
	// Note: no explicit type arguments anywhere in this file. If the
	// signature breaks inference, this file will not compile.
	got := Map([]int{1, 2, 3}, strconv.Itoa)

	if !slices.Equal(got, []string{"1", "2", "3"}) {
		t.Errorf("Map(ints, strconv.Itoa) = %v, want [1 2 3] as strings", got)
	}
}

func TestMapStringToInt(t *testing.T) {
	got := Map([]string{"", "a", "abc"}, func(s string) int { return len(s) })

	if !slices.Equal(got, []int{0, 1, 3}) {
		t.Errorf("Map(strings, len) = %v, want [0 1 3]", got)
	}
}

func TestMapStructToField(t *testing.T) {
	users := []user{{1, "ada"}, {2, "grace"}, {3, "alan"}}

	names := Map(users, func(u user) string { return u.Name })
	if !slices.Equal(names, []string{"ada", "grace", "alan"}) {
		t.Errorf("names = %v, want [ada grace alan]", names)
	}

	ids := Map(users, func(u user) int { return u.ID })
	if !slices.Equal(ids, []int{1, 2, 3}) {
		t.Errorf("ids = %v, want [1 2 3]", ids)
	}
}

func TestMapToStructAndBool(t *testing.T) {
	structs := Map([]int{1, 2}, func(x int) wrapped { return wrapped{V: x} })
	if !slices.Equal(structs, []wrapped{{1}, {2}}) {
		t.Errorf("structs = %v, want [{1} {2}]", structs)
	}

	bools := Map([]int{1, 2, 3, 4}, func(x int) bool { return x%2 == 0 })
	if !slices.Equal(bools, []bool{false, true, false, true}) {
		t.Errorf("bools = %v, want [false true false true]", bools)
	}
}

func TestMapWithPointers(t *testing.T) {
	// U is a pointer type.
	ptrs := Map([]int{1, 2}, func(x int) *wrapped { return &wrapped{V: x} })
	if len(ptrs) != 2 || ptrs[0] == nil || ptrs[1] == nil {
		t.Fatalf("Map to pointers produced %v", ptrs)
	}
	if ptrs[0].V != 1 || ptrs[1].V != 2 {
		t.Errorf("pointer values = %d, %d; want 1, 2", ptrs[0].V, ptrs[1].V)
	}

	// T is a pointer type.
	back := Map(ptrs, func(w *wrapped) int { return w.V })
	if !slices.Equal(back, []int{1, 2}) {
		t.Errorf("Map from pointers = %v, want [1 2]", back)
	}
}

func TestMapSameTypeReturnsFreshSlice(t *testing.T) {
	in := []int{1, 2, 3}
	out := Map(in, func(x int) int { return x * 2 })

	if !slices.Equal(out, []int{2, 4, 6}) {
		t.Fatalf("out = %v, want [2 4 6]", out)
	}

	out[0] = 999
	if in[0] != 1 {
		t.Errorf("mutating the result changed the input: in = %v — result must be fresh memory", in)
	}
}

func TestMapEmptyAndNil(t *testing.T) {
	calls := 0
	f := func(x int) string {
		calls++
		return strconv.Itoa(x)
	}

	for _, name := range []string{"empty", "nil"} {
		t.Run(name, func(t *testing.T) {
			var in []int
			if name == "empty" {
				in = []int{}
			}

			got := Map(in, f)

			if got == nil {
				t.Errorf("Map(%s) returned a nil slice, want an empty non-nil slice", name)
			}
			if len(got) != 0 {
				t.Errorf("Map(%s) = %v, want empty", name, got)
			}
		})
	}

	if calls != 0 {
		t.Errorf("f was called %d times for empty inputs, want 0", calls)
	}
}

func TestMapSingleElement(t *testing.T) {
	got := Map([]string{"solo"}, strings.ToUpper)
	if !slices.Equal(got, []string{"SOLO"}) {
		t.Errorf("Map = %v, want [SOLO]", got)
	}
}

func TestMapCallsOncePerElementInOrder(t *testing.T) {
	var seen []int
	f := func(x int) int {
		seen = append(seen, x)
		return x
	}

	in := []int{5, 3, 9, 1}
	Map(in, f)

	if !slices.Equal(seen, in) {
		t.Errorf("f received %v, want %v — exactly once each, in index order", seen, in)
	}
}

func TestMapWithStatefulTransform(t *testing.T) {
	// A closure whose output depends on call order: proves left-to-right.
	i := 0
	got := Map([]string{"a", "b", "c"}, func(s string) string {
		i++
		return s + strconv.Itoa(i)
	})

	if !slices.Equal(got, []string{"a1", "b2", "c3"}) {
		t.Errorf("got %v, want [a1 b2 c3] — evaluation must be left to right", got)
	}
}

func TestMapDoesNotModifyInput(t *testing.T) {
	in := []user{{1, "ada"}, {2, "grace"}}
	snapshot := slices.Clone(in)

	Map(in, func(u user) string { return u.Name })

	if !slices.Equal(in, snapshot) {
		t.Errorf("input was modified: %v, want %v", in, snapshot)
	}
}

func TestMapComposition(t *testing.T) {
	in := []int{1, 2, 3, 4}

	// Map then Map: int → string → int
	twice := Map(Map(in, strconv.Itoa), func(s string) int { return len(s) * 10 })

	// The same thing in one pass.
	once := Map(in, func(x int) int { return len(strconv.Itoa(x)) * 10 })

	if !slices.Equal(twice, once) {
		t.Errorf("composed = %v, single pass = %v; want identical", twice, once)
	}
}

func TestMapLarge(t *testing.T) {
	const n = 10_000
	in := make([]int, n)
	for i := range in {
		in[i] = i
	}

	got := Map(in, func(x int) int { return x + 1 })

	if len(got) != n {
		t.Fatalf("len = %d, want %d", len(got), n)
	}
	for i := range got {
		if got[i] != i+1 {
			t.Fatalf("got[%d] = %d, want %d", i, got[i], i+1)
		}
	}
}
