package main

import "testing"

func TestFirstUnique(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected rune
		ok       bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			ok:       false,
		},
		{
			name:     "all repeating",
			input:    "aabb",
			expected: 0,
			ok:       false,
		},
		{
			name:     "example swiss",
			input:    "swiss",
			expected: 'w',
			ok:       true,
		},
		{
			name:     "all unique returns first",
			input:    "abc",
			expected: 'a',
			ok:       true,
		},
		{
			name:     "unicode runes",
			input:    "ééa",
			expected: 'a',
			ok:       true,
		},
		{
			name:     "single character",
			input:    "x",
			expected: 'x',
			ok:       true,
		},
		{
			name:     "unicode unique first",
			input:    "😀😃😀",
			expected: '😃',
			ok:       true,
		},
		{
			name:     "unicode all repeating",
			input:    "😀😃😀😃",
			expected: 0,
			ok:       false,
		},
		{
			name:     "case sensitive",
			input:    "aA",
			expected: 'a',
			ok:       true, // assumes 'a' != 'A'
		},
		{
			name:     "first unique in middle",
			input:    "aabbcdde",
			expected: 'c',
			ok:       true,
		},
		{
			name:     "last unique",
			input:    "aabbc",
			expected: 'c',
			ok:       true,
		},
		{
			name:     "all same character",
			input:    "zzzzzz",
			expected: 0,
			ok:       false,
		},
		{
			name:     "mixed unicode ascii",
			input:    "éabcébd",
			expected: 'a',
			ok:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := FirstUnique(tc.input)

			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}

			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestFirstUniqueStableOverRepeatedRuns(t *testing.T) {
	const (
		input      = "swiss"
		expected   = 'w'
		expectedOK = true
	)

	for i := 0; i < 100; i++ {
		got, ok := FirstUnique(input)

		if ok != expectedOK {
			t.Fatalf("run %d: expected ok=%v, got %v", i, expectedOK, ok)
		}

		if got != expected {
			t.Fatalf("run %d: expected %q, got %q", i, expected, got)
		}
	}
}
