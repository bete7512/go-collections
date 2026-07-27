package main

import (
	"slices"
	"testing"
)

func TestRotateLeft(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		k        int
		expected []int
	}{
		{
			name:     "rotate by 0",
			input:    []int{1, 2, 3, 4, 5},
			k:        0,
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "rotate by 1",
			input:    []int{1, 2, 3, 4, 5},
			k:        1,
			expected: []int{2, 3, 4, 5, 1},
		},
		{
			name:     "rotate by 2",
			input:    []int{1, 2, 3, 4, 5},
			k:        2,
			expected: []int{3, 4, 5, 1, 2},
		},
		{
			name:     "rotate by length",
			input:    []int{1, 2, 3, 4, 5},
			k:        5,
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "rotate by greater than length",
			input:    []int{1, 2, 3, 4, 5},
			k:        7, // 7 % 5 = 2
			expected: []int{3, 4, 5, 1, 2},
		},
		{
			name:     "single element",
			input:    []int{42},
			k:        3,
			expected: []int{42},
		},
		{
			name:     "two elements",
			input:    []int{1, 2},
			k:        1,
			expected: []int{2, 1},
		},
		{
			name:     "empty slice",
			input:    []int{},
			k:        5,
			expected: []int{},
		},
		{
			name:     "duplicate values",
			input:    []int{1, 2, 2, 3, 3},
			k:        3,
			expected: []int{3, 3, 1, 2, 2},
		},
		{
			name:     "negative numbers",
			input:    []int{-3, -2, -1, 0, 1},
			k:        2,
			expected: []int{-1, 0, 1, -3, -2},
		},
		{
			name:     "all same values",
			input:    []int{7, 7, 7, 7},
			k:        2,
			expected: []int{7, 7, 7, 7},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			RotateLeft(tc.input, tc.k)

			if !slices.Equal(tc.expected, tc.input) {
				t.Errorf("expected=%v, k = %d, got=%v", tc.expected,tc.k, tc.input)
			}
		})
	}
}
