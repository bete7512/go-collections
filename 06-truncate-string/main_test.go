package main

import "testing"

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input struct {
			str    string
			length int
		}
		output string
	}{
		{
			name: "empty string",
			input: struct {
				str    string
				length int
			}{
				str:    "",
				length: 5,
			},
			output: "",
		},
		{
			name: "string shorter than limit",
			input: struct {
				str    string
				length int
			}{
				str:    "hello",
				length: 10,
			},
			output: "hello",
		},
		{
			name: "string equal to limit",
			input: struct {
				str    string
				length int
			}{
				str:    "hello",
				length: 5,
			},
			output: "hello",
		},
		{
			name: "string longer than limit",
			input: struct {
				str    string
				length int
			}{
				str:    "hello world",
				length: 5,
			},
			output: "hello...",
		},
		{
			name: "truncate one character",
			input: struct {
				str    string
				length int
			}{
				str:    "hello",
				length: 1,
			},
			output: "h...",
		},
		{
			name: "length zero",
			input: struct {
				str    string
				length int
			}{
				str:    "hello",
				length: 0,
			},
			output: "...",
		},
		{
			name: "unicode characters",
			input: struct {
				str    string
				length int
			}{
				str:    "こんにちは世界",
				length: 5,
			},
			output: "こんにちは...",
		},
		{
			name: "spaces",
			input: struct {
				str    string
				length int
			}{
				str:    "hello world",
				length: 6,
			},
			output: "hello ...",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Truncate(tc.input.str, tc.input.length)
			if got != tc.output {
				t.Errorf("Truncate(%s)=%s, wanted = %s", tc.input.str, got, tc.output)
			}
		})
	}
}
