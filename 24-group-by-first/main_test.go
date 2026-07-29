package main

import (
	"reflect"
	"testing"
)

// func GroupByFirst(words []string) map[rune][]string

func TestGroupByFirst(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[rune][]string
	}{
		{
			name:  "empty input",
			input: []string{},
			expected: map[rune][]string{},
		},
		{
			name:  "single word",
			input: []string{"zebra"},
			expected: map[rune][]string{
				'z': {"zebra"},
			},
		},
		{
			name:  "example",
			input: []string{"apple", "avocado", "banana"},
			expected: map[rune][]string{
				'a': {"apple", "avocado"},
				'b': {"banana"},
			},
		},
		{
			name:  "preserves order within groups",
			input: []string{"apple", "banana", "apricot", "blueberry", "avocado"},
			expected: map[rune][]string{
				'a': {"apple", "apricot", "avocado"},
				'b': {"banana", "blueberry"},
			},
		},
		{
			name:  "single rune words",
			input: []string{"a", "b", "a"},
			expected: map[rune][]string{
				'a': {"a", "a"},
				'b': {"b"},
			},
		},
		{
			name:  "unicode first letter",
			input: []string{"élan", "éclair", "apple"},
			expected: map[rune][]string{
				'é': {"élan", "éclair"},
				'a': {"apple"},
			},
		},
		{
			name:  "case sensitive",
			input: []string{"Apple", "apple", "Avocado"},
			expected: map[rune][]string{
				'A': {"Apple", "Avocado"},
				'a': {"apple"},
			},
		},
		{
			name:  "skip empty strings",
			input: []string{"", "apple", "", "banana"},
			expected: map[rune][]string{
				'a': {"apple"},
				'b': {"banana"},
			},
		},
		{
			name:  "only empty strings",
			input: []string{"", "", ""},
			expected: map[rune][]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GroupByFirst(tc.input)

			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("expected %#v, got %#v", tc.expected, got)
			}
		})
	}
}

func TestGroupByFirstPreservesOrder(t *testing.T) {
	input := []string{
		"ant",
		"boat",
		"apple",
		"banana",
		"anchor",
	}

	got := GroupByFirst(input)

	expected := []string{"ant", "apple", "anchor"}

	if !reflect.DeepEqual(got['a'], expected) {
		t.Fatalf("expected order %v, got %v", expected, got['a'])
	}
}

func TestGroupByFirstEmptyStringRule(t *testing.T) {
	got := GroupByFirst([]string{"", "apple"})

	if _, ok := got[0]; ok {
		t.Fatalf("expected empty strings to be skipped, found bucket for zero rune")
	}

	if len(got) != 1 {
		t.Fatalf("expected only one bucket, got %d", len(got))
	}
}