package main

import (
	"slices"
	"testing"
)

func TestReverseInPlace(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "empty slice",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "single element",
			input:    []int{1},
			expected: []int{1},
		},
		{
			name:     "two elements",
			input:    []int{1, 2},
			expected: []int{2, 1},
		},
		{
			name:     "odd number of elements",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{5, 4, 3, 2, 1},
		},
		{
			name:     "even number of elements",
			input:    []int{1, 2, 3, 4},
			expected: []int{4, 3, 2, 1},
		},
		{
			name:     "duplicate elements",
			input:    []int{1, 2, 2, 3},
			expected: []int{3, 2, 2, 1},
		},
		{
			name:     "all same elements",
			input:    []int{7, 7, 7, 7},
			expected: []int{7, 7, 7, 7},
		},
		{
			name:     "negative numbers",
			input:    []int{-1, -2, -3},
			expected: []int{-3, -2, -1},
		},
		{
			name:     "mixed positive and negative",
			input:    []int{-2, 0, 3, -4, 5},
			expected: []int{5, -4, 3, 0, -2},
		},
		{
			name:     "already palindrome",
			input:    []int{1, 2, 3, 2, 1},
			expected: []int{1, 2, 3, 2, 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ReverseInPlace(tc.input)

			if !slices.Equal(tc.expected, tc.input) {
				t.Errorf("expected=%v, got=%v", tc.expected, tc.input)
			}
		})
	}
}
