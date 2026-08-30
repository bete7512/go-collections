package main

import (
	"errors"
	"fmt"
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
		t.Fatalf("%s did not finish within %v — likely deadlock: "+
			"wg.Wait() must run in its own goroutine that closes results, "+
			"otherwise the caller waits for workers while workers wait for the caller", what, d)
	}
}

// poolFunc lets the same table run against both collection strategies.
type poolFunc func(jobs []int, workers int, process func(int) int) []int

var implementations = []struct {
	name string
	fn   poolFunc
}{
	{"PoolCounted", PoolCounted},
	{"PoolStream", PoolStream},
}

func TestPoolReturnsEveryResult(t *testing.T) {
	times10 := func(x int) int { return x * 10 }

	tests := []struct {
		name    string
		jobs    []int
		workers int
		want    []int
	}{
		{"four jobs two workers", []int{1, 2, 3, 4}, 2, []int{10, 20, 30, 40}},
		{"nine jobs three workers", []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, 3,
			[]int{10, 20, 30, 40, 50, 60, 70, 80, 90}},
		{"single job", []int{7}, 3, []int{70}},
		{"one worker", []int{1, 2, 3}, 1, []int{10, 20, 30}},
		{"more workers than jobs", []int{1, 2}, 50, []int{10, 20}},
		{"zero workers treated as one", []int{1, 2, 3}, 0, []int{10, 20, 30}},
		{"negative workers treated as one", []int{1, 2, 3}, -4, []int{10, 20, 30}},
		{"duplicate job values", []int{5, 5, 5}, 2, []int{50, 50, 50}},
		{"negative jobs", []int{-1, -2, 3}, 2, []int{-10, -20, 30}},
		{"empty jobs", []int{}, 3, []int{}},
		{"nil jobs", nil, 3, []int{}},
	}

	for _, impl := range implementations {
		for _, tc := range tests {
			t.Run(impl.name+"/"+tc.name, func(t *testing.T) {
				var got []int
				withTimeout(t, 5*time.Second, impl.name, func() {
					got = impl.fn(tc.jobs, tc.workers, times10)
				})

				if got == nil {
					t.Fatalf("returned nil, want an empty non-nil slice")
				}

				// Results arrive in unpredictable order.
				gotSorted := slices.Clone(got)
				slices.Sort(gotSorted)
				want := slices.Clone(tc.want)
				slices.Sort(want)

				if !slices.Equal(gotSorted, want) {
					t.Errorf("sorted results = %v, want %v", gotSorted, want)
				}
			})
		}
	}
}

func TestPoolLarge(t *testing.T) {
	const n = 1000
	jobs := make([]int, n)
	for i := range jobs {
		jobs[i] = i
	}

	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			var got []int
			withTimeout(t, 15*time.Second, impl.name+" large", func() {
				got = impl.fn(jobs, 8, func(x int) int { return x * x })
			})

			if len(got) != n {
				t.Fatalf("got %d results, want %d", len(got), n)
			}
			slices.Sort(got)
			for i := 0; i < n; i++ {
				if got[i] != i*i {
					t.Fatalf("sorted got[%d] = %d, want %d", i, got[i], i*i)
				}
			}
		})
	}
}

func TestPoolStreamRepeatedRuns(t *testing.T) {
	// The closer-goroutine pattern must work every time, not just usually.
	jobs := []int{1, 2, 3, 4, 5, 6, 7, 8}
	want := []int{2, 4, 6, 8, 10, 12, 14, 16}

	for run := 0; run < 50; run++ {
		var got []int
		withTimeout(t, 5*time.Second, "repeated PoolStream", func() {
			got = PoolStream(jobs, 4, func(x int) int { return x * 2 })
		})

		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Fatalf("run %d: sorted results = %v, want %v", run, got, want)
		}
	}
}

func TestProcessCalledOncePerJob(t *testing.T) {
	jobs := []int{1, 2, 3, 4, 5, 6, 7, 8}

	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			calls := make(chan int, len(jobs)*4) // room for duplicates if buggy

			withTimeout(t, 5*time.Second, impl.name+" call counting", func() {
				impl.fn(jobs, 3, func(x int) int {
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
				t.Errorf("process called %d times, want %d", total, len(jobs))
			}
			for _, j := range jobs {
				if seen[j] != 1 {
					t.Errorf("job %d processed %d times, want exactly 1", j, seen[j])
				}
			}
		})
	}
}

// ---------- results with errors ----------

var errOdd = errors.New("odd input")

// half succeeds on even numbers and fails on odd ones.
func half(x int) (int, error) {
	if x%2 != 0 {
		return 0, fmt.Errorf("halving %d: %w", x, errOdd)
	}
	return x / 2, nil
}

func TestPoolResultsWithErrorsAccounting(t *testing.T) {
	tests := []struct {
		name         string
		jobs         []int
		workers      int
		wantValues   []int
		wantErrCount int
	}{
		{"all succeed", []int{2, 4, 6}, 2, []int{1, 2, 3}, 0},
		{"all fail", []int{1, 3, 5}, 2, []int{}, 3},
		{"mixed", []int{1, 2, 3, 4, 5}, 2, []int{1, 2}, 3},
		{"single success", []int{8}, 3, []int{4}, 0},
		{"single failure", []int{7}, 3, []int{}, 1},
		{"one worker", []int{1, 2, 3, 4}, 1, []int{1, 2}, 2},
		{"more workers than jobs", []int{2, 4}, 30, []int{1, 2}, 0},
		{"empty", []int{}, 3, []int{}, 0},
		{"nil", nil, 3, []int{}, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var values []int
			var errs []error
			withTimeout(t, 5*time.Second, "PoolResultsWithErrors", func() {
				values, errs = PoolResultsWithErrors(tc.jobs, tc.workers, half)
			})

			if values == nil {
				t.Fatalf("values is nil, want an empty non-nil slice")
			}
			if errs == nil {
				t.Fatalf("errors is nil, want an empty non-nil slice")
			}

			// The accounting property: nothing dropped, nothing duplicated.
			if len(values)+len(errs) != len(tc.jobs) {
				t.Fatalf("%d values + %d errors = %d, want %d jobs accounted for",
					len(values), len(errs), len(values)+len(errs), len(tc.jobs))
			}

			gotValues := slices.Clone(values)
			slices.Sort(gotValues)
			wantValues := slices.Clone(tc.wantValues)
			slices.Sort(wantValues)
			if !slices.Equal(gotValues, wantValues) {
				t.Errorf("sorted values = %v, want %v", gotValues, wantValues)
			}
			if len(errs) != tc.wantErrCount {
				t.Errorf("got %d errors, want %d", len(errs), tc.wantErrCount)
			}
		})
	}
}

func TestPoolResultsWithErrorsPreservesErrors(t *testing.T) {
	// The actual errors must come back, not just a count.
	var values []int
	var errs []error
	withTimeout(t, 5*time.Second, "PoolResultsWithErrors", func() {
		values, errs = PoolResultsWithErrors([]int{1, 3}, 2, half)
	})

	if len(values) != 0 {
		t.Errorf("values = %v, want empty", values)
	}
	if len(errs) != 2 {
		t.Fatalf("got %d errors, want 2", len(errs))
	}
	for i, err := range errs {
		if !errors.Is(err, errOdd) {
			t.Errorf("errs[%d] = %v, want an error wrapping errOdd", i, err)
		}
	}
}

func TestPoolResultsWithErrorsLarge(t *testing.T) {
	const n = 1000
	jobs := make([]int, n)
	for i := range jobs {
		jobs[i] = i
	}

	var values []int
	var errs []error
	withTimeout(t, 15*time.Second, "large PoolResultsWithErrors", func() {
		values, errs = PoolResultsWithErrors(jobs, 8, half)
	})

	if len(values)+len(errs) != n {
		t.Fatalf("%d values + %d errors = %d, want %d", len(values), len(errs), len(values)+len(errs), n)
	}
	// 0..999 contains 500 even numbers (successes) and 500 odd ones (failures).
	if len(values) != 500 || len(errs) != 500 {
		t.Errorf("got %d values and %d errors, want 500 each", len(values), len(errs))
	}
}
