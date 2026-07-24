package main

import (
	"slices"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		sep     string
		outputs []string
	}{
		{
			name:    "simple split by space",
			str:     "hello world",
			sep:     " ",
			outputs: []string{"hello", "world"},
		},
		{
			name:    "split by comma",
			str:     "apple,banana,orange",
			sep:     ",",
			outputs: []string{"apple", "banana", "orange"},
		},
		{
			name:    "separator not found",
			str:     "hello",
			sep:     ",",
			outputs: []string{"hello"},
		},
		{
			name:    "empty string",
			str:     "",
			sep:     ",",
			outputs: []string{""},
		},
		{
			name:    "empty separator",
			str:     "hello",
			sep:     "",
			outputs: []string{"h", "e", "l", "l", "o"},
		},
		{
			name:    "multiple separators",
			str:     "hello,,world",
			sep:     ",",
			outputs: []string{"hello", "", "world"},
		},
		{
			name:    "separator at beginning",
			str:     ",hello",
			sep:     ",",
			outputs: []string{"", "hello"},
		},
		{
			name:    "separator at end",
			str:     "hello,",
			sep:     ",",
			outputs: []string{"hello", ""},
		},
		{
			name:    "split by multiple characters",
			str:     "one::two::three",
			sep:     "::",
			outputs: []string{"one", "two", "three"},
		},
		{
			name:    "unicode string",
			str:     "hello世界你好",
			sep:     "世界",
			outputs: []string{"hello", "你好"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Split(tc.str, tc.sep)
			if !slices.Equal(got, tc.outputs) {
				t.Errorf("Split(%s,%s)=%v, expected=%v", tc.str, tc.sep, got, tc.outputs)
			}
		})
	}
}
