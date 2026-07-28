package main

import (
	"slices"
	"testing"
)

func TestFilter(t *testing.T) {
	var (
		even     = func(n int) bool { return n%2 == 0 }
		odd      = func(n int) bool { return n%2 != 0 }
		gtTen    = func(n int) bool { return n > 10 }
		always   = func(int) bool { return true }
		never    = func(int) bool { return false }
		negative = func(n int) bool { return n < 0 }
	)

	tests := []struct {
		name     string
		input    []int
		keep     func(int) bool
		expected []int
	}{
		{"nil input", nil, even, []int{}},
		{"empty input", []int{}, even, []int{}},
		{"even", []int{1, 2, 3, 4, 5, 6}, even, []int{2, 4, 6}},
		{"odd, same input", []int{1, 2, 3, 4, 5, 6}, odd, []int{1, 3, 5}},
		{"nothing matches", []int{1, 2, 3, 4, 5, 6}, gtTen, []int{}},
		{"everything matches", []int{1, 2, 3, 4, 5, 6}, always, []int{1, 2, 3, 4, 5, 6}},
		{"never matches", []int{1, 2, 3}, never, []int{}},
		{"first only", []int{2, 1, 3}, even, []int{2}},
		{"last only", []int{1, 3, 2}, even, []int{2}},
		{"preserves order, not sorted", []int{9, 4, 7, 2}, even, []int{4, 2}},
		{"preserves duplicates", []int{2, 2, 3, 2}, even, []int{2, 2, 2}},
		{"negatives and zero", []int{-3, -2, 0, 1}, negative, []int{-3, -2}},
		{"zero is even", []int{0, 1}, even, []int{0}},
		{"single element kept", []int{4}, even, []int{4}},
		{"single element dropped", []int{5}, even, []int{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Filter(tc.input, tc.keep)
			if !slices.Equal(got, tc.expected) {
				t.Errorf("Filter(%v) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}
