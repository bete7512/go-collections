package main

import (
	"testing"
)

type Data struct {
	input    string
	expected int
}

func TestCountVowels(t *testing.T) {
	tests := []Data{
		{
			input:    "",
			expected: 0,
		},
		{
			input:    "hello",
			expected: 2,
		},
		{
			input:    "Hello World",
			expected: 3,
		},
		{
			input:    "aeiou",
			expected: 5,
		},
		{
			input:    "AEIOU",
			expected: 5,
		},
		{
			input:    "golang",
			expected: 2,
		},
		{
			input:    "rhythm",
			expected: 0,
		},
		{
			input:    "beautiful",
			expected: 5,
		},
		{
			input:    "héllo",
			expected: 1,
		},
		{
			input:    "こんにちは",
			expected: 0,
		},
		{
			input:    "Go👋",
			expected: 1,
		},
		{
			input:    "programming",
			expected: 3,
		},
		{
			input:    "The quick brown fox",
			expected: 5,
		},
		{
			input:    "UPPERCASE",
			expected: 4,
		},
		{
			input:    "12345!@#$%",
			expected: 0,
		},
	}

	for _, tc := range tests {
		got := CountVowels(tc.input)
		if got != tc.expected {
			t.Errorf("CountVowels(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}
