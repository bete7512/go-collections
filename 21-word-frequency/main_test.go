package main

import (
	"maps"
	"testing"
)

// func WordFreq(text string) map[string]int
//
// Normalization: lowercased; split on whitespace; leading/trailing
// punctuation trimmed, internal punctuation kept.
func TestWordFreq(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]int
	}{
		{"empty string", "", map[string]int{}},
		{"only spaces", "     ", map[string]int{}},
		{"only tabs and newlines", "\t\n  \n\t", map[string]int{}},
		{"single word", "cat", map[string]int{"cat": 1}},
		{"single word repeated", "cat cat cat", map[string]int{"cat": 3}},
		{"canonical example", "the cat and the hat", map[string]int{"the": 2, "cat": 1, "and": 1, "hat": 1}},
		{"case normalized", "The the THE", map[string]int{"the": 3}},
		{"mixed case across words", "Cat DOG cat Dog", map[string]int{"cat": 2, "dog": 2}},
		{"leading and trailing whitespace", "  cat dog  ", map[string]int{"cat": 1, "dog": 1}},
		{"runs of spaces", "cat     dog", map[string]int{"cat": 1, "dog": 1}},
		{"tabs and newlines as separators", "cat\tdog\nbird", map[string]int{"cat": 1, "dog": 1, "bird": 1}},
		{"mixed whitespace", "cat \t\n  dog", map[string]int{"cat": 1, "dog": 1}},
		{"all distinct", "a b c", map[string]int{"a": 1, "b": 1, "c": 1}},
		{"repeats not adjacent", "a b a c a", map[string]int{"a": 3, "b": 1, "c": 1}},

		// --- punctuation rule: strip leading/trailing, keep internal ---
		{"trailing comma", "cat, cat", map[string]int{"cat": 2}},
		{"trailing period", "the end.", map[string]int{"the": 1, "end": 1}},
		{"leading quote", `"cat" cat`, map[string]int{"cat": 2}},
		{"surrounding parens", "(cat) cat", map[string]int{"cat": 2}},
		{"question and exclamation", "why? why!", map[string]int{"why": 2}},
		{"internal apostrophe kept", "don't don't", map[string]int{"don't": 2}},
		{"internal hyphen kept", "well-known well-known", map[string]int{"well-known": 2}},
		{"apostrophe vs bare differ", "cant can't", map[string]int{"cant": 1, "can't": 1}},
		{"punctuation-only token drops out", "cat -- dog", map[string]int{"cat": 1, "dog": 1}},
		{"sentence", "The cat sat. The cat, again!", map[string]int{"the": 2, "cat": 2, "sat": 1, "again": 1}},

		// --- things that must NOT be normalized away ---
		{"digits are words", "1 2 2", map[string]int{"1": 1, "2": 2}},
		{"non-ascii lowercased", "Café CAFÉ", map[string]int{"café": 2}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := WordFreq(tc.input)
			if !maps.Equal(got, tc.expected) {
				t.Errorf("WordFreq(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}

	t.Run("empty input returns non-nil map", func(t *testing.T) {
		for _, in := range []string{"", "   ", "\n\t"} {
			got := WordFreq(in)
			if got == nil {
				t.Errorf("WordFreq(%q) returned nil, want empty non-nil map", in)
			}
			if len(got) != 0 {
				t.Errorf("WordFreq(%q) = %v, want empty", in, got)
			}
		}
	})

	t.Run("counts sum to word count", func(t *testing.T) {
		got := WordFreq("a b a c a b")

		total := 0
		for _, n := range got {
			total += n
		}
		if total != 6 {
			t.Errorf("counts sum to %d, want 6 (%v)", total, got)
		}
	})

	t.Run("returned map is independently owned", func(t *testing.T) {
		got := WordFreq("cat dog")
		got["cat"] = 99

		again := WordFreq("cat dog")
		if again["cat"] != 1 {
			t.Errorf("second call returned %d for cat; results share state", again["cat"])
		}
	})
}
