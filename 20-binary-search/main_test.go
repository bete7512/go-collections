package main

import (
	"sort"
	"testing"
)

// func BinarySearch(s []int, target int) (int, bool)
// Precondition: s is sorted ascending.
func TestBinarySearch(t *testing.T) {
	tests := []struct {
		name      string
		s         []int
		target    int
		wantIdx   int
		wantFound bool
	}{
		{"nil slice", nil, 5, -1, false},
		{"empty slice", []int{}, 5, -1, false},
		{"single element hit", []int{5}, 5, 0, true},
		{"single element miss low", []int{5}, 4, -1, false},
		{"single element miss high", []int{5}, 6, -1, false},
		{"two elements, first", []int{1, 2}, 1, 0, true},
		{"two elements, second", []int{1, 2}, 2, 1, true},
		{"two elements, miss between", []int{1, 3}, 2, -1, false},
		{"odd length, middle", []int{1, 3, 5, 7, 9}, 5, 2, true},
		{"odd length, first", []int{1, 3, 5, 7, 9}, 1, 0, true},
		{"odd length, last", []int{1, 3, 5, 7, 9}, 9, 4, true},
		{"odd length, second", []int{1, 3, 5, 7, 9}, 3, 1, true},
		{"odd length, fourth", []int{1, 3, 5, 7, 9}, 7, 3, true},
		{"miss between elements", []int{1, 3, 5, 7, 9}, 4, -1, false},
		{"miss below first", []int{1, 3, 5, 7, 9}, 0, -1, false},
		{"miss above last", []int{1, 3, 5, 7, 9}, 10, -1, false},
		{"even length, lower middle", []int{2, 4, 6, 8}, 4, 1, true},
		{"even length, upper middle", []int{2, 4, 6, 8}, 6, 2, true},
		{"even length, first", []int{2, 4, 6, 8}, 2, 0, true},
		{"even length, last", []int{2, 4, 6, 8}, 8, 3, true},
		{"negatives, hit", []int{-9, -5, -1, 0, 3}, -5, 1, true},
		{"negatives, miss", []int{-9, -5, -1, 0, 3}, -7, -1, false},
		{"zero as target", []int{-2, 0, 2}, 0, 1, true},
		{"all same, hit", []int{4, 4, 4, 4}, 4, 1, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotIdx, gotFound := BinarySearch(tc.s, tc.target)

			if gotFound != tc.wantFound {
				t.Fatalf("BinarySearch(%v, %d) found = %v, want %v", tc.s, tc.target, gotFound, tc.wantFound)
			}
			if !gotFound {
				if gotIdx != tc.wantIdx {
					t.Errorf("BinarySearch(%v, %d) index = %d, want %d on miss", tc.s, tc.target, gotIdx, tc.wantIdx)
				}
				return
			}
			// On a hit, any index holding the target is acceptable.
			if gotIdx < 0 || gotIdx >= len(tc.s) || tc.s[gotIdx] != tc.target {
				t.Errorf("BinarySearch(%v, %d) index = %d, which does not hold the target", tc.s, tc.target, gotIdx)
			}
		})
	}

	t.Run("oracle over a value range", func(t *testing.T) {
		s := []int{1, 3, 5, 7, 9, 11}

		for target := -1; target <= 12; target++ {
			gotIdx, gotFound := BinarySearch(s, target)

			i := sort.SearchInts(s, target)
			wantFound := i < len(s) && s[i] == target

			if gotFound != wantFound {
				t.Errorf("target %d: found = %v, want %v", target, gotFound, wantFound)
				continue
			}
			if gotFound && s[gotIdx] != target {
				t.Errorf("target %d: index %d holds %d, want %d", target, gotIdx, s[gotIdx], target)
			}
			if !gotFound && gotIdx != -1 {
				t.Errorf("target %d: miss index = %d, want -1", target, gotIdx)
			}
		}
	})

	t.Run("oracle with duplicates and gaps", func(t *testing.T) {
		s := []int{2, 2, 2, 5, 5, 9, 9, 9, 9, 14}

		for target := 0; target <= 16; target++ {
			gotIdx, gotFound := BinarySearch(s, target)

			i := sort.SearchInts(s, target)
			wantFound := i < len(s) && s[i] == target

			if gotFound != wantFound {
				t.Errorf("target %d: found = %v, want %v", target, gotFound, wantFound)
				continue
			}
			if gotFound && s[gotIdx] != target {
				t.Errorf("target %d: index %d holds %d, want %d", target, gotIdx, s[gotIdx], target)
			}
		}
	})

	t.Run("does not modify input", func(t *testing.T) {
		s := []int{1, 3, 5, 7, 9}
		before := append([]int(nil), s...)

		BinarySearch(s, 5)
		BinarySearch(s, 6)

		for i := range s {
			if s[i] != before[i] {
				t.Fatalf("input mutated: %v, was %v", s, before)
			}
		}
	})

	t.Run("terminates on large input", func(t *testing.T) {
		s := make([]int, 100000)
		for i := range s {
			s[i] = i * 2
		}

		if idx, found := BinarySearch(s, 199998); !found || s[idx] != 199998 {
			t.Errorf("last element: idx = %d, found = %v", idx, found)
		}
		if _, found := BinarySearch(s, 199999); found {
			t.Error("odd value should not be found")
		}
	})
}