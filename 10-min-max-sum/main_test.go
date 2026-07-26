package main

import "testing"

func TestMinMaxSum(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		min   int
		max   int
		sum   int
	}{
		{
			name:  "empty slice",
			input: []int{},
			min:   0,
			max:   0,
			sum:   0,
		},
		{
			name:  "single element",
			input: []int{5},
			min:   5,
			max:   5,
			sum:   5,
		},
		{
			name:  "positive numbers",
			input: []int{3, 1, 4, 2, 5},
			min:   1,
			max:   5,
			sum:   15,
		},
		{
			name:  "negative numbers",
			input: []int{-5, -2, -10, -1},
			min:   -10,
			max:   -1,
			sum:   -18,
		},
		{
			name:  "mixed numbers",
			input: []int{-3, 7, 0, -1, 4},
			min:   -3,
			max:   7,
			sum:   7,
		},
		{
			name:  "duplicates",
			input: []int{2, 2, 2, 2},
			min:   2,
			max:   2,
			sum:   8,
		},
		{
			name:  "already sorted",
			input: []int{1, 2, 3, 4, 5},
			min:   1,
			max:   5,
			sum:   15,
		},
		{
			name:  "reverse sorted",
			input: []int{5, 4, 3, 2, 1},
			min:   1,
			max:   5,
			sum:   15,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			min, max, sum, _ := MinMaxSum(tc.input)

			if min != tc.min {
				t.Errorf("min = %d, want %d", min, tc.min)
			}

			if max != tc.max {
				t.Errorf("max = %d, want %d", max, tc.max)
			}

			if sum != tc.sum {
				t.Errorf("sum = %d, want %d", sum, tc.sum)
			}
		})
	}
}
