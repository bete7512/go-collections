package main

import (
	"slices"
	"testing"
)

func TestRuneCounts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single rune",
			input:    "x",
			expected: []string{"x:1"},
		},
		{
			name:     "single rune repeated",
			input:    "aaaaaa",
			expected: []string{"a:6"},
		},
		{
			name:     "basic word",
			input:    "banana",
			expected: []string{"a:3", "b:1", "n:2"},
		},
		{
			name:     "case sensitivity uppercase sorts first",
			input:    "aA",
			expected: []string{"A:1", "a:1"},
		},
		{
			name:     "space and punctuation counted",
			input:    "Go 1!",
			expected: []string{" :1", "!:1", "1:1", "G:1", "o:1"},
		},
		{
			name:     "repeated words with space",
			input:    "Go Go!",
			expected: []string{" :1", "!:1", "G:2", "o:2"},
		},
		{
			name:     "digits before uppercase before lowercase",
			input:    "aA bB1",
			expected: []string{" :1", "1:1", "A:1", "B:1", "a:1", "b:1"},
		},
		{
			name:     "accented and cjk sort after ascii",
			input:    "zéz世é",
			expected: []string{"z:2", "é:2", "世:1"},
		},
		{
			name:     "cjk repeated",
			input:    "世界世界世",
			expected: []string{"世:3", "界:2"},
		},
		{
			name:     "emoji counted as one rune",
			input:    "a😀a😀",
			expected: []string{"a:2", "😀:2"},
		},
		{
			name:     "mixed everything",
			input:    "ab BA!!é1é",
			expected: []string{" :1", "!:2", "1:1", "A:1", "B:1", "a:1", "b:1", "é:2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RuneCounts(tc.input)
			if !slices.Equal(got, tc.expected) {
				t.Errorf("RuneCounts(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestRuneCountsDeterminism(t *testing.T) {
	const input = "the quick brown fox jumps over the lazy dog — twice! 世界 😀"

	first := RuneCounts(input)

	for i := 0; i < 50; i++ {
		if got := RuneCounts(input); !slices.Equal(got, first) {
			t.Fatalf("run %d differed:\nfirst: %v\n  now: %v", i, first, got)
		}
	}
}

func TestRuneCountsTotalPreserved(t *testing.T) {
	// The counts must add up to the number of runes in the input.
	input := "abc abc 世世 😀!!"
	runeTotal := len([]rune(input))

	got := RuneCounts(input)

	sum := 0
	for _, entry := range got {
		// Entry format is "<rune>:<count>"; the rune itself may be multi-byte,
		// so parse from the LAST colon.
		last := -1
		for i, r := range entry {
			if r == ':' {
				last = i
			}
		}
		if last < 0 {
			t.Fatalf("entry %q has no colon separator", entry)
		}
		count := 0
		for _, d := range entry[last+1:] {
			if d < '0' || d > '9' {
				t.Fatalf("entry %q has a non-numeric count", entry)
			}
			count = count*10 + int(d-'0')
		}
		if count == 0 {
			t.Errorf("entry %q has count 0 — absent runes must not appear", entry)
		}
		sum += count
	}

	if sum != runeTotal {
		t.Errorf("counts sum to %d, input has %d runes", sum, runeTotal)
	}
}
