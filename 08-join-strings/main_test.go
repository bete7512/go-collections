package main

import "testing"

func TestJoin(t *testing.T) {
	tests := []struct {
		name   string
		strs   []string
		sep    string
		output string
	}{
		{
			name:   "join words with space",
			strs:   []string{"hello", "world"},
			sep:    " ",
			output: "hello world",
		},
		{
			name:   "join words with comma",
			strs:   []string{"apple", "banana", "orange"},
			sep:    ",",
			output: "apple,banana,orange",
		},
		{
			name:   "single string",
			strs:   []string{"hello"},
			sep:    "-",
			output: "hello",
		},
		{
			name:   "empty slice",
			strs:   []string{},
			sep:    ",",
			output: "",
		},
		{
			name:   "empty separator",
			strs:   []string{"hello", "world"},
			sep:    "",
			output: "helloworld",
		},
		{
			name:   "empty strings",
			strs:   []string{"", ""},
			sep:    ",",
			output: ",",
		},
		{
			name:   "separator at multiple positions",
			strs:   []string{"a", "b", "c", "d"},
			sep:    "::",
			output: "a::b::c::d",
		},
		{
			name:   "unicode strings",
			strs:   []string{"こんにちは", "世界"},
			sep:    " ",
			output: "こんにちは 世界",
		},
		{
			name:   "numbers",
			strs:   []string{"1", "2", "3"},
			sep:    "-",
			output: "1-2-3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Join(tc.strs, tc.sep)
			if got != tc.output {
				t.Errorf("Join(%v,%s)=%s, expected=%s", tc.strs, tc.sep, got, tc.output)
			}
		})
	}
}
