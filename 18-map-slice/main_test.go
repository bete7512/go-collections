package main

import (
	"slices"
	"testing"
)

// func MapSlice(s []int, f func(int) int) []int

func TestMapSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		f        func(int) int
		expected []int
	}{
		{"nil input", nil, func(n int) int { return n * 2 }, []int{}},
		{"empty input", []int{}, func(n int) int { return n * 2 }, []int{}},
		{"double", []int{1, 2, 3}, func(n int) int { return n * 2 }, []int{2, 4, 6}},
		{"same input, square", []int{1, 2, 3}, func(n int) int { return n * n }, []int{1, 4, 9}},
		{"identity", []int{4, 7, 2}, func(n int) int { return n }, []int{4, 7, 2}},
		{"constant", []int{4, 7, 2}, func(int) int { return 0 }, []int{0, 0, 0}},
		{"negate", []int{-1, 0, 1}, func(n int) int { return -n }, []int{1, 0, -1}},
		{"preserves order", []int{9, 1, 5}, func(n int) int { return n + 1 }, []int{10, 2, 6}},
		{"preserves duplicates", []int{2, 2, 3}, func(n int) int { return n * 10 }, []int{20, 20, 30}},
		{"collapses distinct to same", []int{1, 2, 3}, func(int) int { return 7 }, []int{7, 7, 7}},
		{"single element", []int{5}, func(n int) int { return n * 3 }, []int{15}},
		{"zero preserved by identity", []int{0}, func(n int) int { return n }, []int{0}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MapSlice(tc.input, tc.f)
			if !slices.Equal(got, tc.expected) {
				t.Errorf("MapSlice(%v) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}

	t.Run("length always matches input", func(t *testing.T) {
		inputs := [][]int{
			{},
			{1},
			{1, 2, 3, 4, 5},
			{0, 0, 0},
		}

		for _, in := range inputs {
			got := MapSlice(in, func(int) int { return 0 })
			if len(got) != len(in) {
				t.Errorf("len(MapSlice(%v)) = %d, want %d", in, len(got), len(in))
			}
		}
	})

	t.Run("calls f once per element in order", func(t *testing.T) {
		input := []int{10, 20, 30}
		var seen []int

		MapSlice(input, func(n int) int {
			seen = append(seen, n)
			return n
		})

		if !slices.Equal(seen, input) {
			t.Errorf("f saw %v, want %v (once each, in order)", seen, input)
		}
	})

	t.Run("does not modify input", func(t *testing.T) {
		input := []int{1, 2, 3}
		before := slices.Clone(input)

		MapSlice(input, func(n int) int { return n * 100 })

		if !slices.Equal(input, before) {
			t.Errorf("input mutated: %v, was %v", input, before)
		}
	})

	t.Run("identity returns fresh backing array", func(t *testing.T) {
		input := []int{1, 2, 3}
		got := MapSlice(input, func(n int) int { return n })

		if !slices.Equal(got, input) {
			t.Fatalf("identity transform: got %v, want %v", got, input)
		}

		got[0] = 99
		if input[0] != 1 {
			t.Errorf("result aliases input: input = %v, want [1 2 3]", input)
		}
	})
}
