package main

import (
	"maps"
	"testing"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		name     string
		a        map[string]int
		b        map[string]int
		expected map[string]int
	}{
		{
			name:     "basic merge with conflict b wins",
			a:        map[string]int{"x": 1, "y": 2},
			b:        map[string]int{"y": 9, "z": 3},
			expected: map[string]int{"x": 1, "y": 9, "z": 3},
		},
		{
			name:     "disjoint keys union",
			a:        map[string]int{"a": 1, "b": 2},
			b:        map[string]int{"c": 3, "d": 4},
			expected: map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
		},
		{
			name:     "b wins with zero value",
			a:        map[string]int{"count": 5},
			b:        map[string]int{"count": 0},
			expected: map[string]int{"count": 0},
		},
		{
			name:     "conflict with equal values",
			a:        map[string]int{"k": 7},
			b:        map[string]int{"k": 7},
			expected: map[string]int{"k": 7},
		},
		{
			name:     "all keys conflict",
			a:        map[string]int{"a": 1, "b": 2, "c": 3},
			b:        map[string]int{"a": 10, "b": 20, "c": 30},
			expected: map[string]int{"a": 10, "b": 20, "c": 30},
		},
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: map[string]int{},
		},
		{
			name:     "a nil",
			a:        nil,
			b:        map[string]int{"a": 1},
			expected: map[string]int{"a": 1},
		},
		{
			name:     "b nil",
			a:        map[string]int{"a": 1},
			b:        nil,
			expected: map[string]int{"a": 1},
		},
		{
			name:     "both empty",
			a:        map[string]int{},
			b:        map[string]int{},
			expected: map[string]int{},
		},
		{
			name:     "negative values",
			a:        map[string]int{"debt": -100, "balance": 50},
			b:        map[string]int{"debt": -250},
			expected: map[string]int{"debt": -250, "balance": 50},
		},
		{
			name:     "empty string key",
			a:        map[string]int{"": 1, "x": 2},
			b:        map[string]int{"": 9},
			expected: map[string]int{"": 9, "x": 2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aSnapshot := maps.Clone(tc.a)
			bSnapshot := maps.Clone(tc.b)

			got := Merge(tc.a, tc.b)

			if got == nil {
				t.Fatalf("Merge(%v, %v) returned nil, want non-nil map", aSnapshot, bSnapshot)
			}
			if !maps.Equal(got, tc.expected) {
				t.Errorf("Merge(%v, %v) = %v, want %v", aSnapshot, bSnapshot, got, tc.expected)
			}
			if !maps.Equal(tc.a, aSnapshot) {
				t.Errorf("input a was modified: had %v, now %v", aSnapshot, tc.a)
			}
			if !maps.Equal(tc.b, bSnapshot) {
				t.Errorf("input b was modified: had %v, now %v", bSnapshot, tc.b)
			}
		})
	}
}

func TestMergeResultIsIndependent(t *testing.T) {
	a := map[string]int{"x": 1, "y": 2}
	b := map[string]int{"y": 9}

	got := Merge(a, b)

	got["x"] = 999
	got["new"] = 42
	delete(got, "y")

	if a["x"] != 1 || a["y"] != 2 || len(a) != 2 {
		t.Errorf("mutating the result changed input a: %v", a)
	}
	if b["y"] != 9 || len(b) != 1 {
		t.Errorf("mutating the result changed input b: %v", b)
	}
}

func TestMergeDisjointLength(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2, "c": 3}
	b := map[string]int{"d": 4, "e": 5}

	got := Merge(a, b)

	if len(got) != len(a)+len(b) {
		t.Errorf("disjoint merge length = %d, want %d", len(got), len(a)+len(b))
	}
}
