package main

import (
	"math/rand"
	"slices"
	"testing"
)

// requirePermutation asserts result contains exactly the same Person values
// as original (same multiset) — nobody added, dropped, or duplicated.
func requirePermutation(t *testing.T, original, result []Person) {
	t.Helper()
	if len(original) != len(result) {
		t.Fatalf("length changed: had %d, now %d", len(original), len(result))
	}
	counts := make(map[Person]int, len(original))
	for _, p := range original {
		counts[p]++
	}
	for _, p := range result {
		counts[p]--
		if counts[p] < 0 {
			t.Fatalf("person %v appears more often than in the input", p)
		}
	}
}

// requireAgesAscending asserts the age sequence never decreases.
func requireAgesAscending(t *testing.T, people []Person) {
	t.Helper()
	for i := 1; i < len(people); i++ {
		if people[i-1].Age > people[i].Age {
			t.Fatalf("ages not ascending at index %d: %d before %d\nfull: %v",
				i, people[i-1].Age, people[i].Age, people)
		}
	}
}

func TestSortByAgeDistinctAges(t *testing.T) {
	tests := []struct {
		name     string
		input    []Person
		expected []Person
	}{
		{
			name:     "nil slice",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty slice",
			input:    []Person{},
			expected: []Person{},
		},
		{
			name:     "single element",
			input:    []Person{{"alice", 30}},
			expected: []Person{{"alice", 30}},
		},
		{
			name:     "already sorted",
			input:    []Person{{"carol", 20}, {"alice", 25}, {"bob", 30}},
			expected: []Person{{"carol", 20}, {"alice", 25}, {"bob", 30}},
		},
		{
			name:     "reverse sorted",
			input:    []Person{{"bob", 30}, {"alice", 25}, {"carol", 20}},
			expected: []Person{{"carol", 20}, {"alice", 25}, {"bob", 30}},
		},
		{
			name:     "shuffled",
			input:    []Person{{"d", 40}, {"a", 10}, {"c", 30}, {"b", 20}},
			expected: []Person{{"a", 10}, {"b", 20}, {"c", 30}, {"d", 40}},
		},
		{
			name:     "age zero sorts first",
			input:    []Person{{"old", 99}, {"newborn", 0}, {"mid", 50}},
			expected: []Person{{"newborn", 0}, {"mid", 50}, {"old", 99}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Sort in place: assert on the very slice we pass in.
			SortByAge(tc.input)
			if !slices.Equal(tc.input, tc.expected) {
				t.Errorf("after SortByAge: %v, want %v", tc.input, tc.expected)
			}
		})
	}
}

func TestSortByAgeWithTies(t *testing.T) {
	// Equal-age order is unspecified (sort.Slice is not stable), so these
	// tests assert the age sequence and the permutation property — never
	// which of two same-age people comes first.
	tests := []struct {
		name  string
		input []Person
		ages  []int
	}{
		{
			name:  "one tie",
			input: []Person{{"bob", 25}, {"alice", 25}, {"carol", 20}},
			ages:  []int{20, 25, 25},
		},
		{
			name:  "all equal ages",
			input: []Person{{"a", 30}, {"b", 30}, {"c", 30}, {"d", 30}},
			ages:  []int{30, 30, 30, 30},
		},
		{
			name:  "several tie groups",
			input: []Person{{"e", 2}, {"a", 1}, {"f", 2}, {"b", 1}, {"g", 3}, {"c", 1}},
			ages:  []int{1, 1, 1, 2, 2, 3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := slices.Clone(tc.input)

			SortByAge(tc.input)

			gotAges := make([]int, len(tc.input))
			for i, p := range tc.input {
				gotAges[i] = p.Age
			}
			if !slices.Equal(gotAges, tc.ages) {
				t.Errorf("age sequence = %v, want %v", gotAges, tc.ages)
			}
			requirePermutation(t, original, tc.input)
		})
	}
}

func TestSortByAgeLarge(t *testing.T) {
	rng := rand.New(rand.NewSource(40)) // fixed seed: deterministic test data
	people := make([]Person, 10_000)
	for i := range people {
		people[i] = Person{
			Name: string(rune('a' + i%26)),
			Age:  rng.Intn(120),
		}
	}
	original := slices.Clone(people)

	SortByAge(people)

	requireAgesAscending(t, people)
	requirePermutation(t, original, people)
}
