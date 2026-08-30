package main

import (
	"slices"
	"testing"
	"time"
)

func withTimeout(t *testing.T, d time.Duration, what string, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("%s did not finish within %v — did you close(jobs)? "+
			"worker range loops only end when the channel is closed", what, d)
	}
}

func TestRunPoolProcessesEveryJob(t *testing.T) {
	tests := []struct {
		name    string
		jobs    []int
		workers int
	}{
		{"nine jobs three workers", []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, 3},
		{"single job", []int{42}, 3},
		{"more workers than jobs", []int{1, 2, 3}, 100},
		{"one worker", []int{1, 2, 3, 4, 5}, 1},
		{"zero workers treated as one", []int{1, 2, 3}, 0},
		{"negative workers treated as one", []int{1, 2, 3}, -5},
		{"duplicate values", []int{7, 7, 7, 7}, 2},
		{"empty jobs", []int{}, 3},
		{"nil jobs", nil, 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got int
			withTimeout(t, 5*time.Second, "RunPool", func() {
				got = RunPool(tc.jobs, tc.workers)
			})

			if got != len(tc.jobs) {
				t.Errorf("RunPool processed %d jobs, want %d", got, len(tc.jobs))
			}
		})
	}
}

func TestRunPoolRepeatedRuns(t *testing.T) {
	// A dropped or duplicated job shows up as a varying count.
	jobs := make([]int, 200)
	for i := range jobs {
		jobs[i] = i
	}

	for run := 0; run < 50; run++ {
		var got int
		withTimeout(t, 5*time.Second, "repeated RunPool", func() {
			got = RunPool(jobs, 8)
		})
		if got != len(jobs) {
			t.Fatalf("run %d: processed %d jobs, want %d — every job must be handled exactly once",
				run, got, len(jobs))
		}
	}
}

func TestEveryJobHandledExactlyOnce(t *testing.T) {
	jobs := []int{10, 20, 30, 40, 50, 60, 70, 80, 90}

	var handled map[int]int
	withTimeout(t, 5*time.Second, "PoolWithWorkerIDs", func() {
		handled = PoolWithWorkerIDs(jobs, 3)
	})

	if len(handled) != len(jobs) {
		t.Fatalf("map has %d entries, want %d — each job must appear exactly once", len(handled), len(jobs))
	}
	for _, j := range jobs {
		if _, ok := handled[j]; !ok {
			t.Errorf("job %d was never handled", j)
		}
	}
}

func TestWorkParticipation(t *testing.T) {
	// With many jobs and several workers, more than one worker must do work.
	// Distribution is not fair and is never asserted precisely.
	jobs := make([]int, 1000)
	for i := range jobs {
		jobs[i] = i
	}

	var handled map[int]int
	withTimeout(t, 10*time.Second, "PoolWithWorkerIDs", func() {
		handled = PoolWithWorkerIDs(jobs, 4)
	})

	participants := map[int]bool{}
	for _, workerID := range handled {
		participants[workerID] = true
	}

	if len(participants) < 2 {
		t.Errorf("only %d worker(s) participated across 1000 jobs with 4 workers — "+
			"multiple receivers on one channel should spread the work", len(participants))
	}
}

func TestProcessAllPreservesInputOrder(t *testing.T) {
	tests := []struct {
		name    string
		jobs    []int
		workers int
		process func(int) int
		want    []int
	}{
		{
			name:    "times ten",
			jobs:    []int{1, 2, 3, 4},
			workers: 3,
			process: func(x int) int { return x * 10 },
			want:    []int{10, 20, 30, 40},
		},
		{
			name:    "identity",
			jobs:    []int{5, 3, 9, 1},
			workers: 2,
			process: func(x int) int { return x },
			want:    []int{5, 3, 9, 1},
		},
		{
			name:    "negate",
			jobs:    []int{1, -2, 3},
			workers: 4,
			process: func(x int) int { return -x },
			want:    []int{-1, 2, -3},
		},
		{
			name:    "single worker",
			jobs:    []int{1, 2, 3},
			workers: 1,
			process: func(x int) int { return x + 1 },
			want:    []int{2, 3, 4},
		},
		{
			name:    "empty",
			jobs:    []int{},
			workers: 3,
			process: func(x int) int { return x },
			want:    []int{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []int
			withTimeout(t, 5*time.Second, "ProcessAll", func() {
				got = ProcessAll(tc.jobs, tc.workers, tc.process)
			})

			if got == nil {
				t.Fatalf("ProcessAll returned nil, want an empty non-nil slice")
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("ProcessAll = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestProcessAllOrderWithVaryingWork(t *testing.T) {
	// Later jobs finish first; the result order must still follow the input.
	jobs := []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}

	process := func(x int) int {
		// Larger values take longer, so the input order is the reverse of
		// the completion order.
		time.Sleep(time.Duration(x) * time.Microsecond * 50)
		return x * 2
	}

	var got []int
	withTimeout(t, 10*time.Second, "ProcessAll with varying work", func() {
		got = ProcessAll(jobs, 4, process)
	})

	want := make([]int, len(jobs))
	for i, j := range jobs {
		want[i] = j * 2
	}
	if !slices.Equal(got, want) {
		t.Errorf("ProcessAll = %v, want %v — results must be in INPUT order, not completion order", got, want)
	}
}

func TestProcessAllLarge(t *testing.T) {
	const n = 1000
	jobs := make([]int, n)
	for i := range jobs {
		jobs[i] = i
	}

	var got []int
	withTimeout(t, 10*time.Second, "large ProcessAll", func() {
		got = ProcessAll(jobs, 8, func(x int) int { return x * x })
	})

	if len(got) != n {
		t.Fatalf("got %d results, want %d", len(got), n)
	}
	for i := range got {
		if got[i] != i*i {
			t.Fatalf("got[%d] = %d, want %d", i, got[i], i*i)
		}
	}
}

func TestProcessAllCallsProcessOncePerJob(t *testing.T) {
	jobs := []int{1, 2, 3, 4, 5, 6, 7, 8}

	calls := make(chan int, len(jobs)*4) // room for duplicates if buggy

	withTimeout(t, 5*time.Second, "ProcessAll call counting", func() {
		ProcessAll(jobs, 3, func(x int) int {
			calls <- x
			return x
		})
	})
	close(calls)

	seen := map[int]int{}
	total := 0
	for v := range calls {
		seen[v]++
		total++
	}

	if total != len(jobs) {
		t.Errorf("process was called %d times, want %d", total, len(jobs))
	}
	for _, j := range jobs {
		if seen[j] != 1 {
			t.Errorf("job %d was processed %d times, want exactly 1", j, seen[j])
		}
	}
}
