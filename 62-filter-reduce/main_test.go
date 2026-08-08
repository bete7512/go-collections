package main

import (
	"maps"
	"slices"
	"strconv"
	"testing"
)

type item struct {
	Name  string
	Price int
}

// ---------- Filter ----------

func TestFilterBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		keep     func(int) bool
		expected []int
	}{
		{
			name:     "evens",
			input:    []int{1, 2, 3, 4, 5, 6},
			keep:     func(x int) bool { return x%2 == 0 },
			expected: []int{2, 4, 6},
		},
		{
			name:     "odds",
			input:    []int{1, 2, 3, 4, 5, 6},
			keep:     func(x int) bool { return x%2 == 1 },
			expected: []int{1, 3, 5},
		},
		{
			name:     "match all",
			input:    []int{1, 2, 3},
			keep:     func(int) bool { return true },
			expected: []int{1, 2, 3},
		},
		{
			name:     "match nothing",
			input:    []int{1, 2, 3},
			keep:     func(int) bool { return false },
			expected: []int{},
		},
		{
			name:     "single element kept",
			input:    []int{7},
			keep:     func(x int) bool { return x > 0 },
			expected: []int{7},
		},
		{
			name:     "negatives",
			input:    []int{-3, 0, 3, -1},
			keep:     func(x int) bool { return x < 0 },
			expected: []int{-3, -1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Filter(tc.input, tc.keep)
			if !slices.Equal(got, tc.expected) {
				t.Errorf("Filter = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestFilterEmptyAndNilReturnNonNil(t *testing.T) {
	calls := 0
	keep := func(int) bool { calls++; return true }

	for _, name := range []string{"empty", "nil"} {
		t.Run(name, func(t *testing.T) {
			var in []int
			if name == "empty" {
				in = []int{}
			}

			got := Filter(in, keep)
			if got == nil {
				t.Errorf("Filter(%s) returned nil, want an empty non-nil slice", name)
			}
			if len(got) != 0 {
				t.Errorf("Filter(%s) = %v, want empty", name, got)
			}
		})
	}

	if calls != 0 {
		t.Errorf("keep was called %d times on empty input, want 0", calls)
	}
}

func TestFilterMatchNothingIsNonNil(t *testing.T) {
	got := Filter([]int{1, 2, 3}, func(int) bool { return false })
	if got == nil {
		t.Errorf("Filter matching nothing returned nil, want an empty non-nil slice")
	}
}

func TestFilterStructs(t *testing.T) {
	items := []item{{"pen", 2}, {"desk", 300}, {"mug", 12}}

	got := Filter(items, func(i item) bool { return i.Price > 10 })

	if !slices.Equal(got, []item{{"desk", 300}, {"mug", 12}}) {
		t.Errorf("Filter = %v, want [{desk 300} {mug 12}]", got)
	}
}

func TestFilterCallsOncePerElementInOrder(t *testing.T) {
	var seen []int
	in := []int{5, 3, 9, 1}

	Filter(in, func(x int) bool {
		seen = append(seen, x)
		return true
	})

	if !slices.Equal(seen, in) {
		t.Errorf("keep received %v, want %v — once each, in index order", seen, in)
	}
}

func TestFilterDoesNotModifyInputAndIsIndependent(t *testing.T) {
	in := []int{1, 2, 3, 4}
	snapshot := slices.Clone(in)

	got := Filter(in, func(x int) bool { return x%2 == 0 })

	if !slices.Equal(in, snapshot) {
		t.Errorf("input modified: %v, want %v", in, snapshot)
	}

	if len(got) > 0 {
		got[0] = 999
		if !slices.Equal(in, snapshot) {
			t.Errorf("mutating the result changed the input: %v — result must be fresh memory", in)
		}
	}
}

func TestFilterLarge(t *testing.T) {
	const n = 10_000
	in := make([]int, n)
	for i := range in {
		in[i] = i
	}

	got := Filter(in, func(x int) bool { return x%3 == 0 })

	want := (n + 2) / 3
	if len(got) != want {
		t.Fatalf("len = %d, want %d", len(got), want)
	}
	for i, v := range got {
		if v != i*3 {
			t.Fatalf("got[%d] = %d, want %d", i, v, i*3)
		}
	}
}

// ---------- Reduce ----------

func TestReduceNumeric(t *testing.T) {
	nums := []int{1, 2, 3, 4}

	if got := Reduce(nums, 0, func(acc, v int) int { return acc + v }); got != 10 {
		t.Errorf("sum = %d, want 10", got)
	}
	if got := Reduce(nums, 1, func(acc, v int) int { return acc * v }); got != 24 {
		t.Errorf("product = %d, want 24", got)
	}
	if got := Reduce(nums, 100, func(acc, v int) int { return acc + v }); got != 110 {
		t.Errorf("sum with init 100 = %d, want 110", got)
	}
}

func TestReduceArgumentOrder(t *testing.T) {
	// Concatenation is not commutative: a swapped (element, accumulator)
	// implementation produces "cba" instead of "abc".
	got := Reduce([]string{"a", "b", "c"}, "", func(acc, v string) string { return acc + v })

	if got != "abc" {
		t.Errorf("Reduce concat = %q, want %q — f must be called as f(accumulator, element)", got, "abc")
	}
}

func TestReduceDifferentAccumulatorType(t *testing.T) {
	words := []string{"go", "rust", "zig"}

	got := Reduce(words, 0, func(acc int, s string) int { return acc + len(s) })
	if got != 9 {
		t.Errorf("sum of lengths = %d, want 9", got)
	}

	// int elements, string accumulator.
	joined := Reduce([]int{1, 2, 3}, "", func(acc string, v int) string { return acc + strconv.Itoa(v) })
	if joined != "123" {
		t.Errorf("joined = %q, want %q", joined, "123")
	}
}

func TestReduceCompositeAccumulators(t *testing.T) {
	t.Run("map accumulator", func(t *testing.T) {
		got := Reduce([]string{"a", "b", "a", "c"}, map[string]int{}, func(acc map[string]int, s string) map[string]int {
			acc[s]++
			return acc
		})
		want := map[string]int{"a": 2, "b": 1, "c": 1}
		if !maps.Equal(got, want) {
			t.Errorf("counts = %v, want %v", got, want)
		}
	})

	t.Run("slice accumulator reverses", func(t *testing.T) {
		got := Reduce([]int{1, 2, 3}, []int{}, func(acc []int, v int) []int {
			return append([]int{v}, acc...)
		})
		if !slices.Equal(got, []int{3, 2, 1}) {
			t.Errorf("reversed = %v, want [3 2 1]", got)
		}
	})

	t.Run("struct accumulator", func(t *testing.T) {
		type stats struct{ Count, Total int }
		got := Reduce([]item{{"a", 5}, {"b", 10}}, stats{}, func(acc stats, i item) stats {
			return stats{Count: acc.Count + 1, Total: acc.Total + i.Price}
		})
		if got != (stats{Count: 2, Total: 15}) {
			t.Errorf("stats = %+v, want {Count:2 Total:15}", got)
		}
	})
}

func TestReduceEmptyReturnsInit(t *testing.T) {
	calls := 0
	f := func(acc int, v int) int { calls++; return acc + v }

	if got := Reduce([]int{}, 42, f); got != 42 {
		t.Errorf("Reduce(empty) = %d, want the init value 42", got)
	}
	if got := Reduce(nil, 42, f); got != 42 {
		t.Errorf("Reduce(nil) = %d, want the init value 42", got)
	}
	if calls != 0 {
		t.Errorf("f was called %d times on empty input, want 0", calls)
	}
}

func TestReduceCallsOncePerElementInOrder(t *testing.T) {
	var seen []int
	in := []int{5, 3, 9, 1}

	Reduce(in, 0, func(acc, v int) int {
		seen = append(seen, v)
		return acc
	})

	if !slices.Equal(seen, in) {
		t.Errorf("f received %v, want %v — once each, in index order", seen, in)
	}
}

func TestReduceLarge(t *testing.T) {
	const n = 10_000
	in := make([]int, n)
	for i := range in {
		in[i] = 1
	}

	if got := Reduce(in, 0, func(acc, v int) int { return acc + v }); got != n {
		t.Errorf("sum = %d, want %d", got, n)
	}
}

// ---------- Composition ----------

func TestFilterThenReduce(t *testing.T) {
	items := []item{{"pen", 2}, {"desk", 300}, {"mug", 12}, {"clip", 1}}

	expensive := Filter(items, func(i item) bool { return i.Price >= 10 })
	total := Reduce(expensive, 0, func(acc int, i item) int { return acc + i.Price })

	if total != 312 {
		t.Errorf("total of expensive items = %d, want 312", total)
	}
}

func TestMapViaReduce(t *testing.T) {
	// Map is a special case of Reduce; this must satisfy #61's contract.
	got := MapViaReduce([]int{1, 2, 3}, strconv.Itoa)
	if !slices.Equal(got, []string{"1", "2", "3"}) {
		t.Errorf("MapViaReduce = %v, want [1 2 3] as strings", got)
	}

	lengths := MapViaReduce([]string{"", "ab", "cde"}, func(s string) int { return len(s) })
	if !slices.Equal(lengths, []int{0, 2, 3}) {
		t.Errorf("MapViaReduce lengths = %v, want [0 2 3]", lengths)
	}

	empty := MapViaReduce([]int{}, strconv.Itoa)
	if empty == nil {
		t.Errorf("MapViaReduce(empty) returned nil, want an empty non-nil slice")
	}
	if len(empty) != 0 {
		t.Errorf("MapViaReduce(empty) = %v, want empty", empty)
	}
}
