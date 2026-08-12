package main

import (
	"slices"
	"testing"
)

func TestRunWorkersCollectsEveryID(t *testing.T) {
	tests := []int{1, 2, 5, 10, 100}

	for _, n := range tests {
		got := RunWorkers(n)

		if len(got) != n {
			t.Fatalf("RunWorkers(%d) returned %d ids, want %d", n, len(got), n)
		}

		// Order is unspecified — sort before comparing.
		slices.Sort(got)
		want := make([]int, n)
		for i := range want {
			want[i] = i
		}
		if !slices.Equal(got, want) {
			t.Errorf("RunWorkers(%d) sorted = %v, want %v", n, got, want)
		}
	}
}

func TestRunWorkersZeroAndNegative(t *testing.T) {
	for _, n := range []int{0, -1, -100} {
		got := RunWorkers(n)
		if got == nil {
			t.Errorf("RunWorkers(%d) returned nil, want an empty non-nil slice", n)
		}
		if len(got) != 0 {
			t.Errorf("RunWorkers(%d) = %v, want empty", n, got)
		}
	}
}

func TestRunWorkersSingle(t *testing.T) {
	got := RunWorkers(1)
	if !slices.Equal(got, []int{0}) {
		t.Errorf("RunWorkers(1) = %v, want [0]", got)
	}
}

func TestRunWorkersRepeatedRuns(t *testing.T) {
	// An Add() inside the goroutine races with Wait(): some runs return early
	// with an incomplete set. Repetition is what exposes it.
	want := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	for run := 0; run < 50; run++ {
		got := RunWorkers(10)
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Fatalf("run %d: RunWorkers(10) sorted = %v, want %v — "+
				"Add must be called before the go statement", run, got, want)
		}
	}
}

func TestSumParallelMatchesSequential(t *testing.T) {
	tests := []struct {
		name    string
		nums    []int
		workers int
	}{
		{"nil input", nil, 3},
		{"empty input", []int{}, 3},
		{"single element", []int{42}, 3},
		{"even split", []int{1, 2, 3, 4, 5, 6}, 3},
		{"ragged split", []int{1, 2, 3, 4, 5, 6, 7, 8}, 3},
		{"one worker", []int{1, 2, 3, 4, 5}, 1},
		{"workers equal length", []int{1, 2, 3, 4}, 4},
		{"more workers than elements", []int{1, 2, 3}, 100},
		{"zero workers treated as one", []int{1, 2, 3, 4}, 0},
		{"negative workers treated as one", []int{1, 2, 3, 4}, -5},
		{"negative numbers", []int{-5, 3, -2, 10, -1}, 3},
		{"all zeros", []int{0, 0, 0, 0}, 2},
		{"mixed signs cancel", []int{-10, 10, -20, 20}, 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want := 0
			for _, v := range tc.nums {
				want += v
			}

			got := SumParallel(tc.nums, tc.workers)
			if got != want {
				t.Errorf("SumParallel(%v, %d) = %d, want %d", tc.nums, tc.workers, got, want)
			}
		})
	}
}

func TestSumParallelIsDeterministic(t *testing.T) {
	// A lost update shows up as a total that varies between runs.
	nums := make([]int, 1000)
	for i := range nums {
		nums[i] = i + 1
	}
	want := 1000 * 1001 / 2

	for run := 0; run < 100; run++ {
		if got := SumParallel(nums, 8); got != want {
			t.Fatalf("run %d: SumParallel = %d, want %d — every element must be counted exactly once",
				run, got, want)
		}
	}
}

func TestSumParallelLarge(t *testing.T) {
	const n = 100_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = 1
	}

	for _, workers := range []int{1, 2, 7, 16, 64} {
		if got := SumParallel(nums, workers); got != n {
			t.Errorf("SumParallel(100k ones, %d workers) = %d, want %d", workers, got, n)
		}
	}
}

func TestAllWorkersFinishBeforeReturn(t *testing.T) {
	// Every id in the returned slice is proof that goroutine finished before
	// RunWorkers returned. A short slice means Wait unblocked too early.
	const n = 50

	for run := 0; run < 20; run++ {
		ids := RunWorkers(n)
		if len(ids) != n {
			t.Fatalf("run %d: got %d ids, want %d — Wait returned before every goroutine finished",
				run, len(ids), n)
		}
	}
}
