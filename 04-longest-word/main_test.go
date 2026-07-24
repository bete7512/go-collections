package main

import "testing"

func TestLongestWord(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "empty string",
			input:  "",
			output: "",
		},
		{
			name:   "single word",
			input:  "hello",
			output: "hello",
		},
		{
			name:   "multiple words",
			input:  "Go is an amazing language",
			output: "language",
		},
		{
			name:   "longest word at beginning",
			input:  "elephant is big",
			output: "elephant",
		},
		{
			name:   "longest word in middle",
			input:  "I love programming every day",
			output: "programming",
		},
		{
			name:   "longest word at end",
			input:  "I like supercalifragilisticexpialidocious",
			output: "supercalifragilisticexpialidocious",
		},
		{
			name:   "multiple spaces",
			input:  "Go    is     awesome",
			output: "awesome",
		},
		{
			name:   "leading and trailing spaces",
			input:  "   hello world   ",
			output: "hello",
		},
		{
			name:   "all words same length",
			input:  "cat dog pig",
			output: "cat", // first longest
		},
		{
			name:   "tie for longest",
			input:  "apple zebra house",
			output: "apple", // first 5-letter word
		},
		{
			name:   "one character words",
			input:  "a b c d",
			output: "a",
		},
		{
			name:   "numbers as words",
			input:  "1234 123456 12",
			output: "123456",
		},
		{
			name:   "mixed letters and numbers",
			input:  "go golang123 abc",
			output: "golang123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := LongestWord(tc.input)
			if got != tc.output {
				t.Errorf("LongestWord(%s)=%s, expected=%s", tc.input, got, tc.output)
			}
		})
	}
}
