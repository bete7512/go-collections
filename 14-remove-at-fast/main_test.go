package main

import (
	"slices"
	"testing"
)

func TestRemoveAtFast(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []int
		index    int
		expected []int
	}{
		{
			name:     "remove middle",
			inputs:   []int{1, 2, 3, 4},
			index:    1,
			expected: []int{1, 4, 3},
		},
		{
			name:     "remove first",
			inputs:   []int{1, 2, 3, 4},
			index:    0,
			expected: []int{4, 2, 3},
		},
		{
			name:     "remove last",
			inputs:   []int{1, 2, 3, 4},
			index:    3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "remove only element",
			inputs:   []int{9},
			index:    0,
			expected: []int{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RemoveAtFast(tc.inputs, tc.index)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !slices.Equal(got, tc.expected) {
				t.Fatalf("RemoveAtFast(%v, %d) = %v, want %v",
					tc.inputs, tc.index, got, tc.expected)
			}
		})
	}
}

func TestRemoveAtFastNegativeIndex(t *testing.T) {
	_, err := RemoveAtFast([]int{1, 2, 3}, -1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoveAtFastIndexEqualLength(t *testing.T) {
	_, err := RemoveAtFast([]int{1, 2, 3}, 3)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoveAtFastIndexGreaterThanLength(t *testing.T) {
	_, err := RemoveAtFast([]int{1, 2, 3}, 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoveAtFastEmptySlice(t *testing.T) {
	_, err := RemoveAtFast([]int{}, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}
