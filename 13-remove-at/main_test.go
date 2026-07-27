package main

import (
	"slices"
	"testing"
)

func TestRemoveAt(t *testing.T) {
	tests := []struct {
		name      string
		inputs    []int
		index     int
		expected  []int
		shouldErr bool
	}{
		{
			name:     "remove from middle",
			inputs:   []int{1, 2, 3, 4},
			index:    1,
			expected: []int{1, 3, 4},
		},
		{
			name:     "remove first element",
			inputs:   []int{1, 2, 3, 4},
			index:    0,
			expected: []int{2, 3, 4},
		},
		{
			name:     "remove last element",
			inputs:   []int{1, 2, 3, 4},
			index:    3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "remove only element",
			inputs:   []int{42},
			index:    0,
			expected: []int{},
		},
		{
			name:      "negative index",
			inputs:    []int{1, 2, 3},
			index:     -1,
			shouldErr: true,
		},
		{
			name:      "index equals length",
			inputs:    []int{1, 2, 3},
			index:     3,
			shouldErr: true,
		},
		{
			name:      "index greater than length",
			inputs:    []int{1, 2, 3},
			index:     10,
			shouldErr: true,
		},
		{
			name:      "empty slice",
			inputs:    []int{},
			index:     0,
			shouldErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RemoveAt(tc.inputs, tc.index)

			if tc.shouldErr {
				if err == nil {
					t.Fatalf("expected an error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !slices.Equal(got, tc.expected) {
				t.Fatalf("RemoveAt(%v, %d) = %v, want %v",
					tc.inputs, tc.index, got, tc.expected)
			}
		})
	}
}
