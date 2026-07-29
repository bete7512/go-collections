package main

import "testing"

// func IsAnagram(a, b string) bool

func TestIsAnagram(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: true,
		},
		{
			name:     "empty and non empty",
			a:        "",
			b:        "a",
			expected: false,
		},
		{
			name:     "different rune counts",
			a:        "ab",
			b:        "abc",
			expected: false,
		},
		{
			name:     "example listen silent",
			a:        "listen",
			b:        "silent",
			expected: true,
		},
		{
			name:     "not an anagram",
			a:        "hello",
			b:        "world",
			expected: false,
		},
		{
			name:     "same letters different counts",
			a:        "aab",
			b:        "abb",
			expected: false,
		},
		{
			name:     "same string",
			a:        "golang",
			b:        "golang",
			expected: true,
		},
		{
			name:     "single character equal",
			a:        "x",
			b:        "x",
			expected: true,
		},
		{
			name:     "single character different",
			a:        "x",
			b:        "y",
			expected: false,
		},
		{
			name:     "unicode anagram",
			a:        "héllo",
			b:        "olléh",
			expected: true,
		},
		{
			name:     "unicode different counts",
			a:        "ééa",
			b:        "éaa",
			expected: false,
		},
		{
			name:     "emoji anagram",
			a:        "😀😃😀",
			b:        "😀😀😃",
			expected: true,
		},
		{
			name:     "emoji not anagram",
			a:        "😀😃😀",
			b:        "😀😃😃",
			expected: false,
		},
		{
			name:     "case sensitive",
			a:        "Listen",
			b:        "silent",
			expected: false, // assumes case-sensitive comparison
		},
		{
			name:     "whitespace matters",
			a:        "a b",
			b:        "ab ",
			expected: true,
		},
		{
			name:     "punctuation matters",
			a:        "a,b",
			b:        "ba,",
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsAnagram(tc.a, tc.b)

			if got != tc.expected {
				t.Fatalf("IsAnagram(%q, %q) = %v, want %v",
					tc.a, tc.b, got, tc.expected)
			}
		})
	}
}