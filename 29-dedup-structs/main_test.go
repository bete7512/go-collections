package main

import (
	"fmt"
	"slices"
	"testing"
)

func TestDedupByID(t *testing.T) {
	tests := []struct {
		name     string
		input    []User
		expected []User
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []User{},
		},
		{
			name:     "empty input",
			input:    []User{},
			expected: []User{},
		},
		{
			name:     "single element",
			input:    []User{{1, "a"}},
			expected: []User{{1, "a"}},
		},
		{
			name:     "no duplicates",
			input:    []User{{1, "a"}, {2, "b"}, {3, "c"}},
			expected: []User{{1, "a"}, {2, "b"}, {3, "c"}},
		},
		{
			name:     "adjacent duplicate first wins",
			input:    []User{{1, "alice"}, {1, "carol"}, {2, "bob"}},
			expected: []User{{1, "alice"}, {2, "bob"}},
		},
		{
			name:     "non-adjacent duplicate first wins",
			input:    []User{{1, "alice"}, {2, "bob"}, {1, "carol"}},
			expected: []User{{1, "alice"}, {2, "bob"}},
		},
		{
			name:     "interleaved duplicates keep order",
			input:    []User{{3, "x"}, {1, "y"}, {3, "z"}, {2, "w"}, {1, "q"}},
			expected: []User{{3, "x"}, {1, "y"}, {2, "w"}},
		},
		{
			name:     "all same id",
			input:    []User{{7, "a"}, {7, "b"}, {7, "c"}, {7, "d"}},
			expected: []User{{7, "a"}},
		},
		{
			name:     "zero id is a real id",
			input:    []User{{0, "zero-a"}, {1, "one"}, {0, "zero-b"}},
			expected: []User{{0, "zero-a"}, {1, "one"}},
		},
		{
			name:     "negative ids",
			input:    []User{{-1, "neg"}, {1, "pos"}, {-1, "neg-dup"}, {0, "zero"}},
			expected: []User{{-1, "neg"}, {1, "pos"}, {0, "zero"}},
		},
		{
			name:     "same name different ids are not duplicates",
			input:    []User{{1, "same"}, {2, "same"}, {3, "same"}},
			expected: []User{{1, "same"}, {2, "same"}, {3, "same"}},
		},
		{
			name:     "three plus occurrences only first survives",
			input:    []User{{5, "first"}, {5, "second"}, {6, "x"}, {5, "third"}, {5, "fourth"}},
			expected: []User{{5, "first"}, {6, "x"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := slices.Clone(tc.input)

			got := DedupByID(tc.input)

			if !slices.Equal(got, tc.expected) {
				t.Errorf("DedupByID(%v) = %v, want %v", snapshot, got, tc.expected)
			}
			if !slices.Equal(tc.input, snapshot) {
				t.Errorf("input was modified: had %v, now %v", snapshot, tc.input)
			}
		})
	}
}

func TestDedupByIDResultIsIndependent(t *testing.T) {
	input := []User{{1, "a"}, {2, "b"}, {1, "c"}}
	got := DedupByID(input)

	if len(got) == 0 {
		t.Fatalf("DedupByID(%v) returned empty result", input)
	}
	got[0] = User{99, "mutated"}

	if input[0] != (User{1, "a"}) {
		t.Errorf("mutating the result changed the input: input[0] = %v, want {1 a}", input[0])
	}
}

func TestDedupByIDLargeInput(t *testing.T) {
	const total = 10000
	const distinct = 100

	input := make([]User, 0, total)
	for i := 0; i < total; i++ {
		id := i % distinct
		input = append(input, User{ID: id, Name: fmt.Sprintf("user-%d-occurrence-%d", id, i/distinct)})
	}

	got := DedupByID(input)

	if len(got) != distinct {
		t.Fatalf("expected %d distinct users, got %d", distinct, len(got))
	}
	for i, u := range got {
		if u.ID != i {
			t.Errorf("survivor order broken at index %d: got ID %d, want %d", i, u.ID, i)
		}
		want := fmt.Sprintf("user-%d-occurrence-0", u.ID)
		if u.Name != want {
			t.Errorf("first occurrence did not win for ID %d: got name %q, want %q", u.ID, u.Name, want)
		}
	}
}
