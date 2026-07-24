package main

import "testing"

func TestTitleCase(t *testing.T) {
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
			name:   "single lowercase word",
			input:  "hello",
			output: "Hello",
		},
		{
			name:   "single uppercase word",
			input:  "HELLO",
			output: "Hello",
		},
		{
			name:   "mixed case word",
			input:  "hELLo",
			output: "Hello",
		},
		{
			name:   "multiple words",
			input:  "hello world",
			output: "Hello World",
		},
		{
			name:   "multiple spaces",
			input:  "hello   world",
			output: "Hello   World",
		},
		{
			name:   "leading and trailing spaces",
			input:  "  hello world  ",
			output: "  Hello World  ",
		},
		{
			name:   "already title case",
			input:  "Hello World",
			output: "Hello World",
		},
		{
			name:   "mixed casing",
			input:  "hElLo WoRLd",
			output: "Hello World",
		},
		{
			name:   "numbers",
			input:  "go 1.24 released",
			output: "Go 1.24 Released",
		},
		{
			name:   "punctuation",
			input:  "hello, world!",
			output: "Hello, World!",
		},
		{
			name:   "hyphenated word",
			input:  "well-known fact",
			output: "Well-Known Fact",
		},
		{
			name:   "apostrophe",
			input:  "john'S book",
			output: "John's Book",
		},
		{
			name:   "tab separated",
			input:  "hello\tworld",
			output: "Hello\tWorld",
		},
		{
			name:   "newline separated",
			input:  "hello\nworld",
			output: "Hello\nWorld",
		},
		{
			name:   "unicode",
			input:  "élève français",
			output: "Élève Français",
		},
		{
			name:   "single character",
			input:  "a",
			output: "A",
		},
		{
			name:   "only spaces",
			input:  "   ",
			output: "   ",
		},
		{
			name:   "symbols only",
			input:  "!@#$%^",
			output: "!@#$%^",
		},
		{
			name:   "word after punctuation",
			input:  "hello.world",
			output: "Hello.World",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := TitleCase(tc.input)
			if got != tc.output {
				t.Errorf("TitleCase(%q) = %q, want %q", tc.input, got, tc.output)
			}
		})
	}
}
