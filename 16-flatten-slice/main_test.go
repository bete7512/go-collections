package main

import (
	"slices"
	"testing"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]int
		expected []int
	}{
		{"nil outer", nil, []int{}},
		{"empty outer", [][]int{}, []int{}},
		{"single empty inner", [][]int{{}}, []int{}},
		{"all inners empty", [][]int{{}, {}, {}}, []int{}},
		{"all inners nil", [][]int{nil, nil}, []int{}},
		{"single inner", [][]int{{1, 2, 3}}, []int{1, 2, 3}},
		{"one element each", [][]int{{1}, {2}, {3}}, []int{1, 2, 3}},
		{"uneven lengths", [][]int{{1}, {2, 3, 4}, {5, 6}}, []int{1, 2, 3, 4, 5, 6}},
		{"empties interleaved", [][]int{{}, {1, 2}, {}, {3}, {}}, []int{1, 2, 3}},
		{"nil inner mixed in", [][]int{nil, {1}}, []int{1}},
		{"nil inner in middle", [][]int{{1}, nil, {2}}, []int{1, 2}},
		{"preserves duplicates", [][]int{{1, 1}, {1}}, []int{1, 1, 1}},
		{"preserves order", [][]int{{5, 3}, {9, 1}}, []int{5, 3, 9, 1}},
		{"negatives and zero", [][]int{{-1, 0}, {-5}}, []int{-1, 0, -5}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Flatten(tc.input)
			if !slices.Equal(got, tc.expected) {
				t.Errorf("Flatten(%v) = %v, want %v", tc.input, got, tc.expected)

			}
		})
	}
}
