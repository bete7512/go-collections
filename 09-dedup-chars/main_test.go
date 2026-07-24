package main

import "testing"

func TestDedupChars(t *testing.T) {
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
			name:   "no duplicates",
			input:  "hello",
			output: "helo",
		},
		{
			name:   "all characters unique",
			input:  "abcde",
			output: "abcde",
		},
		{
			name:   "all duplicates",
			input:  "aaaaa",
			output: "a",
		},
		{
			name:   "duplicates at beginning",
			input:  "aabbcc",
			output: "abc",
		},
		{
			name:   "duplicates in middle",
			input:  "abccba",
			output: "abc",
		},
		{
			name:   "duplicates at end",
			input:  "helloool",
			output: "helo",
		},
		{
			name:   "case sensitive",
			input:  "aAaA",
			output: "aA",
		},
		{
			name:   "spaces",
			input:  "hello world",
			output: "helo wrd",
		},
		{
			name:   "special characters",
			input:  "!!@@##",
			output: "!@#",
		},
		{
			name:   "numbers",
			input:  "112233",
			output: "123",
		},
		{
			name:   "unicode characters",
			input:  "こんにちはこんにちは",
			output: "こんにちは",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DedupChars(tc.input)
			if got != tc.output {
				t.Errorf("DedupChars(%s)=%s, expected=%s", tc.input, got, tc.output)
			}
		})
	}
}
