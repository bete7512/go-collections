package main

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"
)

func TestSortByAgeThenName(t *testing.T) {
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
			name:     "tie resolved alphabetically",
			input:    []Person{{"bob", 25}, {"alice", 25}, {"carol", 20}},
			expected: []Person{{"carol", 20}, {"alice", 25}, {"bob", 25}},
		},
		{
			name:     "tiebreaker must not fire on distinct ages",
			input:    []Person{{"zoe", 18}, {"adam", 30}},
			expected: []Person{{"zoe", 18}, {"adam", 30}},
		},
		{
			name: "names anti-correlated with ages",
			// Alphabetical order is exactly the reverse of age order; a
			// comparator that mixes the keys (the || bug) reorders these wrongly.
			input:    []Person{{"a", 40}, {"b", 30}, {"c", 20}, {"d", 10}},
			expected: []Person{{"d", 10}, {"c", 20}, {"b", 30}, {"a", 40}},
		},
		{
			name:     "all ages equal pure name sort",
			input:    []Person{{"dave", 30}, {"alice", 30}, {"carol", 30}, {"bob", 30}},
			expected: []Person{{"alice", 30}, {"bob", 30}, {"carol", 30}, {"dave", 30}},
		},
		{
			name:     "all names equal pure age sort",
			input:    []Person{{"sam", 50}, {"sam", 10}, {"sam", 30}},
			expected: []Person{{"sam", 10}, {"sam", 30}, {"sam", 50}},
		},
		{
			name:     "mixed case byte order",
			input:    []Person{{"bob", 25}, {"Alice", 25}, {"ann", 25}},
			expected: []Person{{"Alice", 25}, {"ann", 25}, {"bob", 25}},
		},
		{
			name:     "multiple tie groups",
			input:    []Person{{"b", 2}, {"d", 1}, {"a", 2}, {"c", 1}, {"e", 3}},
			expected: []Person{{"c", 1}, {"d", 1}, {"a", 2}, {"b", 2}, {"e", 3}},
		},
		{
			name:     "fully identical people",
			input:    []Person{{"twin", 9}, {"twin", 9}, {"ada", 9}},
			expected: []Person{{"ada", 9}, {"twin", 9}, {"twin", 9}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			SortByAgeThenName(tc.input)
			if !slices.Equal(tc.input, tc.expected) {
				t.Errorf("after sort: %v\nwant:       %v", tc.input, tc.expected)
			}
		})
	}
}

// buildExpected constructs people already in age-then-name order BY
// CONSTRUCTION (nested loops), so the test never has to implement the
// comparator it is judging.
func buildExpected(ages, namesPerAge int) []Person {
	expected := make([]Person, 0, ages*namesPerAge)
	for age := 0; age < ages; age++ {
		for n := 0; n < namesPerAge; n++ {
			expected = append(expected, Person{
				Name: fmt.Sprintf("n%03d", n), // fixed width: lexicographic == numeric
				Age:  age,
			})
		}
	}
	return expected
}

func TestSortByAgeThenNameDeterminism(t *testing.T) {
	expected := buildExpected(5, 4)
	rng := rand.New(rand.NewSource(41))

	for run := 0; run < 20; run++ {
		shuffled := slices.Clone(expected)
		rng.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		SortByAgeThenName(shuffled)

		if !slices.Equal(shuffled, expected) {
			t.Fatalf("run %d: sort of a reshuffle differed from the fully determined order\ngot:  %v\nwant: %v",
				run, shuffled, expected)
		}
	}
}

func TestSortByAgeThenNameLarge(t *testing.T) {
	expected := buildExpected(50, 100) // 5000 people
	shuffled := slices.Clone(expected)
	rng := rand.New(rand.NewSource(4141))
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	SortByAgeThenName(shuffled)

	if !slices.Equal(shuffled, expected) {
		for i := range expected {
			if shuffled[i] != expected[i] {
				t.Fatalf("first mismatch at index %d: got %v, want %v", i, shuffled[i], expected[i])
			}
		}
	}
}
