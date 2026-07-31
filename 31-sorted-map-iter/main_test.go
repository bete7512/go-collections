package main

import (
	"maps"
	"slices"
	"testing"
)

func TestSortedKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected []string
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: []string{},
		},
		{
			name:     "single entry",
			input:    map[string]int{"only": 1},
			expected: []string{"only"},
		},
		{
			name:     "already alphabetical",
			input:    map[string]int{"a": 1, "b": 2, "c": 3},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "reverse insertion order",
			input:    map[string]int{"z": 26, "m": 13, "a": 1},
			expected: []string{"a", "m", "z"},
		},
		{
			name:     "uppercase sorts before lowercase",
			input:    map[string]int{"apple": 1, "Banana": 2, "cherry": 3},
			expected: []string{"Banana", "apple", "cherry"},
		},
		{
			name:     "numeric-looking keys sort as strings",
			input:    map[string]int{"9": 9, "10": 10, "1": 1},
			expected: []string{"1", "10", "9"},
		},
		{
			name:     "empty string key sorts first",
			input:    map[string]int{"a": 1, "": 0, "b": 2},
			expected: []string{"", "a", "b"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := maps.Clone(tc.input)

			got := SortedKeys(tc.input)

			if !slices.Equal(got, tc.expected) {
				t.Errorf("SortedKeys(%v) = %v, want %v", snapshot, got, tc.expected)
			}
			if !maps.Equal(tc.input, snapshot) {
				t.Errorf("input map was modified: had %v, now %v", snapshot, tc.input)
			}
		})
	}
}

func TestFormatSorted(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected string
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: "",
		},
		{
			name:     "single entry ends with newline",
			input:    map[string]int{"a": 1},
			expected: "a=1\n",
		},
		{
			name:     "multiple entries sorted",
			input:    map[string]int{"b": 2, "a": 1, "c": 3},
			expected: "a=1\nb=2\nc=3\n",
		},
		{
			name:     "negative and zero values",
			input:    map[string]int{"x": -5, "y": 0, "z": 100},
			expected: "x=-5\ny=0\nz=100\n",
		},
		{
			name:     "uppercase before lowercase",
			input:    map[string]int{"delta": 4, "Alpha": 1},
			expected: "Alpha=1\ndelta=4\n",
		},
		{
			name:     "empty string key",
			input:    map[string]int{"": 7, "a": 1},
			expected: "=7\na=1\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatSorted(tc.input)
			if got != tc.expected {
				t.Errorf("FormatSorted(%v) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestDeterminismAcrossRuns(t *testing.T) {
	m := map[string]int{
		"kappa": 10, "alpha": 1, "omega": 24, "beta": 2, "gamma": 3,
		"delta": 4, "zeta": 6, "eta": 7, "theta": 8, "iota": 9,
	}

	firstKeys := SortedKeys(m)
	firstFormat := FormatSorted(m)

	if !slices.IsSorted(firstKeys) {
		t.Fatalf("SortedKeys result is not sorted: %v", firstKeys)
	}
	if len(firstKeys) != len(m) {
		t.Fatalf("SortedKeys returned %d keys, map has %d", len(firstKeys), len(m))
	}

	for i := 0; i < 50; i++ {
		if keys := SortedKeys(m); !slices.Equal(keys, firstKeys) {
			t.Fatalf("run %d: SortedKeys differed:\nfirst: %v\n now: %v", i, firstKeys, keys)
		}
		if out := FormatSorted(m); out != firstFormat {
			t.Fatalf("run %d: FormatSorted differed:\nfirst: %q\n now: %q", i, firstFormat, out)
		}
	}
}
