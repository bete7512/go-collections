package main

import (
	"slices"
	"testing"
)

func TestTopN(t *testing.T) {
	tests := []struct {
		name string
		text string
		n    int
		want []string
	}{
		{
			name: "basic frequency ordering",
			text: "b b a a a c",
			n:    2,
			want: []string{"a", "b"},
		},
		{
			name: "longer sentence with distinct counts",
			text: "the day is sunny the the the sunny is is",
			n:    4,
			want: []string{"the", "is", "sunny", "day"},
		},
		{
			name: "tie broken alphabetically",
			text: "b a a b",
			n:    2,
			want: []string{"a", "b"},
		},
		{
			name: "three way tie excludes lower count",
			text: "z z q m q m b",
			n:    3,
			want: []string{"m", "q", "z"}, // m:2 q:2 z:2 all tied; b:1 loses
		},
		{
			name: "all same frequency is pure alphabetical",
			text: "d c b a",
			n:    3,
			want: []string{"a", "b", "c"},
		},
		{
			name: "n equals vocabulary size",
			text: "x y x z y x",
			n:    3,
			want: []string{"x", "y", "z"}, // x:3, y:2, z:1
		},
		{
			name: "n larger than vocabulary returns everything ranked",
			text: "cat dog cat",
			n:    10,
			want: []string{"cat", "dog"},
		},
		{
			name: "case folding merges words",
			text: "Go go GO gopher",
			n:    1,
			want: []string{"go"},
		},
		{
			name: "case folded output is lowercase",
			text: "HELLO Hello",
			n:    1,
			want: []string{"hello"},
		},
		{
			name: "punctuation is not stripped",
			text: "go go, go,",
			n:    2,
			want: []string{"go,", "go"}, // "go,":2 beats "go":1
		},
		{
			name: "mixed whitespace separators",
			text: "a\tb\na a\t b",
			n:    2,
			want: []string{"a", "b"}, // a:3, b:2
		},
		{
			name: "single word repeated",
			text: "x x x x",
			n:    5,
			want: []string{"x"},
		},
		{
			name: "single word n one",
			text: "solo",
			n:    1,
			want: []string{"solo"},
		},
		{
			name: "unicode words",
			text: "héllo héllo wörld",
			n:    1,
			want: []string{"héllo"},
		},
		{
			name: "numeric tokens are words",
			text: "42 7 42 42 7 1",
			n:    2,
			want: []string{"42", "7"},
		},
		{
			name: "count beats alphabet",
			text: "zebra zebra apple",
			n:    2,
			want: []string{"zebra", "apple"}, // zebra:2 outranks apple:1 despite alphabet
		},
		{
			name: "n zero returns empty",
			text: "a b c",
			n:    0,
			want: []string{},
		},
		{
			name: "negative n returns empty",
			text: "a b c",
			n:    -3,
			want: []string{},
		},
		{
			name: "empty text returns empty",
			text: "",
			n:    5,
			want: []string{},
		},
		{
			name: "whitespace only text returns empty",
			text: " \t\n  ",
			n:    5,
			want: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := TopN(tc.text, tc.n)
			if len(tc.want) == 0 {
				if len(got) != 0 {
					t.Fatalf("TopN(%q, %d) = %v, want empty", tc.text, tc.n, got)
				}
				return
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("TopN(%q, %d) = %v, want %v", tc.text, tc.n, got, tc.want)
			}
		})
	}
}

// A tie-heavy input must produce the identical answer on every call.
// If map iteration order leaks into the result, this flakes.
func TestTopNDeterministic(t *testing.T) {
	text := "kiwi mango kiwi apple banana mango apple banana cherry cherry"
	// every word has count 2 → ordering is decided entirely by the tiebreak
	want := TopN(text, 3)
	if len(want) != 3 {
		t.Fatalf("expected 3 results, got %v", want)
	}
	for i := 0; i < 20; i++ {
		got := TopN(text, 3)
		if !slices.Equal(got, want) {
			t.Fatalf("run %d: got %v, previous runs got %v — output depends on map order", i, got, want)
		}
	}
	// and the tiebreak itself must be alphabetical
	if !slices.Equal(want, []string{"apple", "banana", "cherry"}) {
		t.Fatalf("all-tied input: got %v, want [apple banana cherry]", want)
	}
}

// Result length is always min(n, distinct words).
func TestTopNResultLength(t *testing.T) {
	text := "a b c d e a b c a b" // 5 distinct words
	for n := 0; n <= 8; n++ {
		got := TopN(text, n)
		wantLen := min(n, 5)
		if len(got) != wantLen {
			t.Fatalf("n=%d: got %d results (%v), want %d", n, len(got), got, wantLen)
		}
	}
}
