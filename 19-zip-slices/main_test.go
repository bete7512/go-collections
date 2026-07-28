package main

import (
	"slices"
	"testing"
)

func TestZip(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []int
		expected []Pair
	}{
		{"equal lengths", []int{1, 2}, []int{10, 20}, []Pair{{1, 10}, {2, 20}}},
		{"first shorter", []int{1}, []int{10, 20, 30}, []Pair{{1, 10}}},
		{"second shorter", []int{1, 2, 3}, []int{10, 20}, []Pair{{1, 10}, {2, 20}}},
		{"first empty", []int{}, []int{1, 2}, []Pair{}},
		{"second empty", []int{1, 2}, []int{}, []Pair{}},
		{"both empty", []int{}, []int{}, []Pair{}},
		{"first nil", nil, []int{1, 2}, []Pair{}},
		{"second nil", []int{1, 2}, nil, []Pair{}},
		{"both nil", nil, nil, []Pair{}},
		{"single pair", []int{7}, []int{8}, []Pair{{7, 8}}},
		{"order preserved, not sorted", []int{3, 1, 2}, []int{30, 10, 20}, []Pair{{3, 30}, {1, 10}, {2, 20}}},
		{"duplicates preserved", []int{1, 1}, []int{5, 5}, []Pair{{1, 5}, {1, 5}}},
		{"negatives and zero", []int{-1, 0}, []int{0, -2}, []Pair{{-1, 0}, {0, -2}}},
		{"a and b not swapped", []int{1, 2}, []int{9, 8}, []Pair{{1, 9}, {2, 8}}},
		{"much longer second", []int{1}, []int{1, 2, 3, 4, 5, 6, 7}, []Pair{{1, 1}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Zip(tc.a, tc.b)
			if !slices.Equal(got, tc.expected) {
				t.Errorf("Zip(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.expected)
			}
		})
	}

	t.Run("length is min of inputs", func(t *testing.T) {
		cases := [][2][]int{
			{nil, nil},
			{{}, {1, 2, 3}},
			{{1, 2, 3}, {}},
			{{1}, {1, 2}},
			{{1, 2, 3, 4}, {1, 2}},
			{{1, 2, 3}, {1, 2, 3}},
		}

		for _, c := range cases {
			a, b := c[0], c[1]
			want := min(len(a), len(b))
			if got := len(Zip(a, b)); got != want {
				t.Errorf("len(Zip(%v, %v)) = %d, want %d", a, b, got, want)
			}
		}
	})

	t.Run("does not modify inputs", func(t *testing.T) {
		a := []int{1, 2, 3}
		b := []int{10, 20, 30}
		aBefore := slices.Clone(a)
		bBefore := slices.Clone(b)

		got := Zip(a, b)
		got[0] = Pair{99, 99}

		if !slices.Equal(a, aBefore) {
			t.Errorf("a mutated: %v, was %v", a, aBefore)
		}
		if !slices.Equal(b, bBefore) {
			t.Errorf("b mutated: %v, was %v", b, bBefore)
		}
	})
}