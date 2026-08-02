package main

import (
	"math/rand"
	"slices"
	"sort"
	"testing"
)

// Compile-time proof that ByAge satisfies sort.Interface.
var _ sort.Interface = ByAge(nil)

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

func TestLen(t *testing.T) {
	if got := ByAge(nil).Len(); got != 0 {
		t.Errorf("ByAge(nil).Len() = %d, want 0", got)
	}
	if got := (ByAge{}).Len(); got != 0 {
		t.Errorf("empty Len() = %d, want 0", got)
	}
	if got := (ByAge{{"a", 1}, {"b", 2}, {"c", 3}}).Len(); got != 3 {
		t.Errorf("Len() = %d, want 3", got)
	}
}

func TestLessIsStrict(t *testing.T) {
	a := ByAge{{"young", 10}, {"old", 20}, {"twin", 10}}

	if !a.Less(0, 1) {
		t.Errorf("Less(0,1) with ages 10,20 = false, want true")
	}
	if a.Less(1, 0) {
		t.Errorf("Less(1,0) with ages 20,10 = true, want false")
	}
	// Equal ages: strictly-less means BOTH directions are false.
	if a.Less(0, 2) {
		t.Errorf("Less(0,2) with equal ages = true, want false (use <, not <=)")
	}
	if a.Less(2, 0) {
		t.Errorf("Less(2,0) with equal ages = true, want false")
	}
}

func TestSwapMutatesSharedBackingArray(t *testing.T) {
	people := []Person{{"first", 1}, {"second", 2}, {"third", 3}}

	// The conversion shares the backing array — Swap through ByAge must be
	// visible in the original []Person variable.
	ByAge(people).Swap(0, 2)

	want := []Person{{"third", 3}, {"second", 2}, {"first", 1}}
	if !slices.Equal(people, want) {
		t.Fatalf("after Swap(0,2) through ByAge: %v, want %v (value receiver still mutates shared array)", people, want)
	}

	// Swapping back restores the original.
	ByAge(people).Swap(2, 0)
	want = []Person{{"first", 1}, {"second", 2}, {"third", 3}}
	if !slices.Equal(people, want) {
		t.Errorf("double swap did not restore: %v, want %v", people, want)
	}
}

func TestSortSortDistinctAges(t *testing.T) {
	tests := []struct {
		name     string
		input    []Person
		expected []Person
	}{
		{
			name:     "empty",
			input:    []Person{},
			expected: []Person{},
		},
		{
			name:     "single",
			input:    []Person{{"solo", 42}},
			expected: []Person{{"solo", 42}},
		},
		{
			name:     "already sorted",
			input:    []Person{{"a", 1}, {"b", 2}, {"c", 3}},
			expected: []Person{{"a", 1}, {"b", 2}, {"c", 3}},
		},
		{
			name:     "reverse sorted",
			input:    []Person{{"c", 3}, {"b", 2}, {"a", 1}},
			expected: []Person{{"a", 1}, {"b", 2}, {"c", 3}},
		},
		{
			name:     "shuffled",
			input:    []Person{{"bob", 30}, {"alice", 25}, {"carol", 20}},
			expected: []Person{{"carol", 20}, {"alice", 25}, {"bob", 30}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(ByAge(tc.input))
			// Assert on the ORIGINAL slice: the conversion shares memory.
			if !slices.Equal(tc.input, tc.expected) {
				t.Errorf("after sort.Sort(ByAge(...)): %v, want %v", tc.input, tc.expected)
			}
		})
	}
}

func TestSortSortWithTies(t *testing.T) {
	input := []Person{{"e", 2}, {"a", 1}, {"f", 2}, {"b", 1}, {"g", 3}}
	original := slices.Clone(input)

	sort.Sort(ByAge(input))

	wantAges := []int{1, 1, 2, 2, 3}
	gotAges := make([]int, len(input))
	for i, p := range input {
		gotAges[i] = p.Age
	}
	if !slices.Equal(gotAges, wantAges) {
		t.Errorf("age sequence = %v, want %v", gotAges, wantAges)
	}
	requirePermutation(t, original, input)
}

func TestSortSortLarge(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	people := make([]Person, 10_000)
	for i := range people {
		people[i] = Person{Name: string(rune('a' + i%26)), Age: rng.Intn(120)}
	}
	original := slices.Clone(people)

	sort.Sort(ByAge(people))

	for i := 1; i < len(people); i++ {
		if people[i-1].Age > people[i].Age {
			t.Fatalf("ages not ascending at index %d: %d before %d", i, people[i-1].Age, people[i].Age)
		}
	}
	requirePermutation(t, original, people)
}
