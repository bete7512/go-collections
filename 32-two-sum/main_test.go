package main

import (
	"slices"
	"testing"
)

// requireValidPair asserts the returned pair is a correct answer without
// pinning WHICH valid pair, per the contract.
func requireValidPair(t *testing.T, nums []int, target, i, j int) {
	t.Helper()
	if i >= j {
		t.Fatalf("indices not ascending: i=%d, j=%d", i, j)
	}
	if i < 0 || j >= len(nums) {
		t.Fatalf("indices out of range: i=%d, j=%d, len=%d", i, j, len(nums))
	}
	if nums[i]+nums[j] != target {
		t.Fatalf("nums[%d]+nums[%d] = %d+%d = %d, want %d",
			i, j, nums[i], nums[j], nums[i]+nums[j], target)
	}
}

func TestTwoSumFound(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
	}{
		{
			name:   "basic first two",
			nums:   []int{2, 7, 11, 15},
			target: 9,
		},
		{
			name:   "answer in middle",
			nums:   []int{3, 2, 4},
			target: 6,
		},
		{
			name:   "answer is last two elements",
			nums:   []int{1, 5, 9, 13},
			target: 22,
		},
		{
			name:   "answer is first and last",
			nums:   []int{8, 100, 200, 300, -3},
			target: 5,
		},
		{
			name:   "same value at two indices",
			nums:   []int{3, 3},
			target: 6,
		},
		{
			name:   "half target appears twice among others",
			nums:   []int{1, 5, 7, 5, 2},
			target: 10,
		},
		{
			name:   "negatives",
			nums:   []int{-1, -2, -3, -4, -5},
			target: -8,
		},
		{
			name:   "negative and positive crossing zero",
			nums:   []int{-3, 4, 3, 90},
			target: 0,
		},
		{
			name:   "negative target",
			nums:   []int{-10, 20, 4, -6},
			target: -16,
		},
		{
			name:   "zeros summing to zero target",
			nums:   []int{7, 0, 3, 0},
			target: 0,
		},
		{
			name:   "multiple valid pairs any accepted",
			nums:   []int{1, 9, 2, 8, 3, 7},
			target: 10,
		},
		{
			name:   "two elements exactly",
			nums:   []int{4, 6},
			target: 10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := slices.Clone(tc.nums)

			i, j, ok := TwoSum(tc.nums, tc.target)

			if !ok {
				t.Fatalf("TwoSum(%v, %d) = not found, want a valid pair", snapshot, tc.target)
			}
			requireValidPair(t, snapshot, tc.target, i, j)

			if !slices.Equal(tc.nums, snapshot) {
				t.Errorf("input was modified: had %v, now %v", snapshot, tc.nums)
			}
		})
	}
}

func TestTwoSumNotFound(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
	}{
		{
			name:   "empty slice",
			nums:   []int{},
			target: 5,
		},
		{
			name:   "nil slice",
			nums:   nil,
			target: 0,
		},
		{
			name:   "single element",
			nums:   []int{5},
			target: 5,
		},
		{
			name:   "single element cannot pair with itself",
			nums:   []int{5},
			target: 10,
		},
		{
			name:   "half target appears only once",
			nums:   []int{5, 1, 2},
			target: 10,
		},
		{
			name:   "no pair among many",
			nums:   []int{1, 2, 3, 4, 5},
			target: 100,
		},
		{
			name:   "would need same negative twice",
			nums:   []int{-4, 1, 2},
			target: -8,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, ok := TwoSum(tc.nums, tc.target)
			if ok {
				t.Errorf("TwoSum(%v, %d) reported a pair, want not found", tc.nums, tc.target)
			}
		})
	}
}

func TestTwoSumLargeInput(t *testing.T) {
	const n = 100_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i * 3 // no two distinct elements sum to the target below except the planted pair
	}
	// Plant the only valid pair at the extreme ends.
	nums[0] = -1_000_000
	nums[n-1] = 1_000_003
	target := 3 // only -1_000_000 + 1_000_003 works: all other sums are multiples of 3 >= 9

	i, j, ok := TwoSum(nums, target)
	if !ok {
		t.Fatalf("TwoSum on large input: pair not found, want (0, %d)", n-1)
	}
	requireValidPair(t, nums, target, i, j)
	if i != 0 || j != n-1 {
		t.Errorf("expected the planted pair (0, %d), got (%d, %d)", n-1, i, j)
	}
}
