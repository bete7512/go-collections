package main

import "testing"

// func (s *Set) Union(other *Set) *Set
// func (s *Set) Intersect(other *Set) *Set
// func (s *Set) Diff(other *Set) *Set

func assertSetEqual(t *testing.T, got *Set, expected []string) {
	t.Helper()

	if got.Len() != len(expected) {
		t.Fatalf("expected len %d, got %d", len(expected), got.Len())
	}

	for _, item := range expected {
		if !got.Has(item) {
			t.Fatalf("expected %q to exist", item)
		}
	}
}

func TestSetOperations(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		operation string
		expected []string
	}{
		{
			name:      "union",
			a:         []string{"1", "2", "3"},
			b:         []string{"2", "3", "4"},
			operation: "union",
			expected:  []string{"1", "2", "3", "4"},
		},
		{
			name:      "intersection",
			a:         []string{"1", "2", "3"},
			b:         []string{"2", "3", "4"},
			operation: "intersection",
			expected:  []string{"2", "3"},
		},
		{
			name:      "difference A-B",
			a:         []string{"1", "2", "3"},
			b:         []string{"2", "3", "4"},
			operation: "diff",
			expected:  []string{"1"},
		},
		{
			name:      "difference B-A",
			a:         []string{"2", "3", "4"},
			b:         []string{"1", "2", "3"},
			operation: "diff",
			expected:  []string{"4"},
		},
		{
			name:      "disjoint intersection",
			a:         []string{"1", "2"},
			b:         []string{"3", "4"},
			operation: "intersection",
			expected:  []string{},
		},
		{
			name:      "identical difference",
			a:         []string{"1", "2", "3"},
			b:         []string{"1", "2", "3"},
			operation: "diff",
			expected:  []string{},
		},
		{
			name:      "empty receiver",
			a:         []string{},
			b:         []string{"1", "2"},
			operation: "union",
			expected:  []string{"1", "2"},
		},
		{
			name:      "empty argument",
			a:         []string{"1", "2"},
			b:         []string{},
			operation: "diff",
			expected:  []string{"1", "2"},
		},
		{
			name:      "self union",
			a:         []string{"a", "b"},
			b:         []string{"a", "b"},
			operation: "union",
			expected:  []string{"a", "b"},
		},
		{
			name:      "self intersection",
			a:         []string{"a", "b"},
			b:         []string{"a", "b"},
			operation: "intersection",
			expected:  []string{"a", "b"},
		},
		{
			name:      "self difference",
			a:         []string{"a", "b"},
			b:         []string{"a", "b"},
			operation: "diff",
			expected:  []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := NewSet(tc.a...)
			b := NewSet(tc.b...)

			gotBeforeA := NewSet(tc.a...)
			gotBeforeB := NewSet(tc.b...)

			var result *Set

			switch tc.operation {
			case "union":
				result = a.Union(b)

			case "intersection":
				result = a.Intersect(b)

			case "diff":
				result = a.Diff(b)
			}

			assertSetEqual(t, result, tc.expected)

			// Ensure original sets were not modified
			assertSetEqual(t, a, tc.a)
			assertSetEqual(t, b, tc.b)

			// Ensure result is independent from originals
			if result == a || result == b {
				t.Fatal("operation returned original set instead of new set")
			}

			// Extra mutation check
			result.Add("mutation-check")

			assertSetEqual(t, a, tc.a)
			assertSetEqual(t, b, tc.b)

			// avoid unused variables if you remove mutation check later
			_ = gotBeforeA
			_ = gotBeforeB
		})
	}
}