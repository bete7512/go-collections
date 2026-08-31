package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
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
		t.Fatalf("%s did not finish within %v — every channel a goroutine ranges over "+
			"must eventually be closed (the dispatcher owns the per-worker channels)", what, d)
	}
}

// feedInts returns a closed channel carrying vals.
func feedInts(vals ...int) <-chan int {
	ch := make(chan int, len(vals))
	for _, v := range vals {
		ch <- v
	}
	close(ch)
	return ch
}

func feedStrings(vals ...string) <-chan string {
	ch := make(chan string, len(vals))
	for _, v := range vals {
		ch <- v
	}
	close(ch)
	return ch
}

type fanFunc func(in <-chan int, workers int) []Result

var implementations = []struct {
	name string
	fn   fanFunc
}{
	{"FanOut", FanOut},
	{"FanOutDispatch", FanOutDispatch},
}

func TestEveryValueHandledExactlyOnce(t *testing.T) {
	tests := []struct {
		name    string
		vals    []int
		workers int
	}{
		{"six values three workers", []int{1, 2, 3, 4, 5, 6}, 3},
		{"single value", []int{42}, 3},
		{"one worker", []int{1, 2, 3, 4}, 1},
		{"more workers than values", []int{1, 2}, 20},
		{"zero workers treated as one", []int{1, 2, 3}, 0},
		{"negative workers treated as one", []int{1, 2, 3}, -3},
		{"duplicate values", []int{7, 7, 7}, 2},
		{"negatives and zero", []int{-1, 0, 1}, 2},
		{"empty stream", nil, 3},
	}

	for _, impl := range implementations {
		for _, tc := range tests {
			t.Run(impl.name+"/"+tc.name, func(t *testing.T) {
				var got []Result
				withTimeout(t, 5*time.Second, impl.name, func() {
					got = impl.fn(feedInts(tc.vals...), tc.workers)
				})

				if got == nil {
					t.Fatalf("returned nil, want an empty non-nil slice")
				}
				if len(got) != len(tc.vals) {
					t.Fatalf("got %d results, want %d — every value must produce exactly one result",
						len(got), len(tc.vals))
				}

				gotVals := make([]int, len(got))
				for i, r := range got {
					gotVals[i] = r.Value
				}
				slices.Sort(gotVals)
				want := slices.Clone(tc.vals)
				slices.Sort(want)
				if !slices.Equal(gotVals, want) {
					t.Errorf("sorted values = %v, want %v", gotVals, want)
				}

				// Worker IDs must be in range.
				workers := tc.workers
				if workers <= 0 {
					workers = 1
				}
				for _, r := range got {
					if r.WorkerID < 0 || r.WorkerID >= workers {
						t.Errorf("value %d handled by worker %d, out of range [0,%d)",
							r.Value, r.WorkerID, workers)
					}
				}
			})
		}
	}
}

func TestLargeStream(t *testing.T) {
	const n = 1000
	vals := make([]int, n)
	for i := range vals {
		vals[i] = i
	}

	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			var got []Result
			withTimeout(t, 15*time.Second, impl.name+" large", func() {
				got = impl.fn(feedInts(vals...), 8)
			})

			if len(got) != n {
				t.Fatalf("got %d results, want %d", len(got), n)
			}
			seen := map[int]int{}
			for _, r := range got {
				seen[r.Value]++
			}
			for _, v := range vals {
				if seen[v] != 1 {
					t.Fatalf("value %d appeared %d times, want exactly 1", v, seen[v])
				}
			}
		})
	}
}

func TestFanOutSpreadsWork(t *testing.T) {
	// Distribution is not fair and is never asserted precisely — but with
	// 1000 values and 4 workers, more than one worker must participate.
	const n = 1000
	vals := make([]int, n)
	for i := range vals {
		vals[i] = i
	}

	var got []Result
	withTimeout(t, 15*time.Second, "FanOut", func() {
		got = FanOut(feedInts(vals...), 4)
	})

	participants := map[int]bool{}
	for _, r := range got {
		participants[r.WorkerID] = true
	}
	if len(participants) < 2 {
		t.Errorf("only %d worker(s) participated — multiple receivers on one channel should spread work",
			len(participants))
	}
}

func TestFanOutRepeatedRuns(t *testing.T) {
	vals := []int{1, 2, 3, 4, 5, 6, 7, 8}

	for run := 0; run < 20; run++ {
		var got []Result
		withTimeout(t, 5*time.Second, "repeated FanOut", func() {
			got = FanOut(feedInts(vals...), 3)
		})

		gotVals := make([]int, len(got))
		for i, r := range got {
			gotVals[i] = r.Value
		}
		slices.Sort(gotVals)
		if !slices.Equal(gotVals, vals) {
			t.Fatalf("run %d: sorted values = %v, want %v", run, gotVals, vals)
		}
	}
}

func TestFanOutDispatchIsRoundRobin(t *testing.T) {
	// Unlike FanOut, the explicit dispatcher assigns deterministically:
	// the value at input index i goes to worker i%workers.
	vals := []int{10, 20, 30, 40, 50, 60, 70}
	const workers = 3

	for run := 0; run < 10; run++ {
		var got []Result
		withTimeout(t, 5*time.Second, "FanOutDispatch", func() {
			got = FanOutDispatch(feedInts(vals...), workers)
		})

		assigned := map[int]int{}
		for _, r := range got {
			assigned[r.Value] = r.WorkerID
		}
		for i, v := range vals {
			want := i % workers
			if assigned[v] != want {
				t.Fatalf("run %d: value %d (index %d) went to worker %d, want %d — "+
					"round-robin dispatch must be deterministic", run, v, i, assigned[v], want)
			}
		}
	}
}

// ---------- PartitionBy ----------

// userKey extracts the numeric part of "userN:payload".
func userKey(s string) int {
	name, _, _ := strings.Cut(s, ":")
	n, err := strconv.Atoi(strings.TrimPrefix(name, "user"))
	if err != nil {
		return 0
	}
	return n
}

func TestPartitionBySameKeySameWorker(t *testing.T) {
	vals := []string{
		"user1:a", "user2:x", "user1:b", "user3:p",
		"user1:c", "user2:y", "user4:q", "user3:r",
	}

	for run := 0; run < 10; run++ {
		var got []StringResult
		withTimeout(t, 5*time.Second, "PartitionBy", func() {
			got = PartitionBy(feedStrings(vals...), 3, userKey)
		})

		if len(got) != len(vals) {
			t.Fatalf("run %d: got %d results, want %d", run, len(got), len(vals))
		}

		// Group worker IDs by key; each key must map to exactly one worker.
		byKey := map[int]map[int]bool{}
		for _, r := range got {
			k := userKey(r.Value)
			if byKey[k] == nil {
				byKey[k] = map[int]bool{}
			}
			byKey[k][r.WorkerID] = true
		}
		for k, workers := range byKey {
			if len(workers) != 1 {
				t.Fatalf("run %d: key %d was handled by %d different workers — "+
					"all values with the same key must go to the same worker", run, k, len(workers))
			}
		}
	}
}

func TestPartitionByAllOneKey(t *testing.T) {
	vals := []string{"user7:a", "user7:b", "user7:c", "user7:d"}

	var got []StringResult
	withTimeout(t, 5*time.Second, "PartitionBy single key", func() {
		got = PartitionBy(feedStrings(vals...), 4, userKey)
	})

	if len(got) != len(vals) {
		t.Fatalf("got %d results, want %d", len(got), len(vals))
	}
	workers := map[int]bool{}
	for _, r := range got {
		workers[r.WorkerID] = true
	}
	if len(workers) != 1 {
		t.Errorf("values sharing one key were spread across %d workers, want 1", len(workers))
	}
}

func TestPartitionByNegativeKeys(t *testing.T) {
	// A key function may return negatives; key%workers would then be
	// negative and index out of range.
	negKey := func(s string) int { return -len(s) }

	vals := []string{"a", "bb", "ccc", "dddd", "eeeee"}

	var got []StringResult
	withTimeout(t, 5*time.Second, "PartitionBy negative keys", func() {
		got = PartitionBy(feedStrings(vals...), 3, negKey)
	})

	if len(got) != len(vals) {
		t.Fatalf("got %d results, want %d", len(got), len(vals))
	}
	for _, r := range got {
		if r.WorkerID < 0 || r.WorkerID >= 3 {
			t.Errorf("value %q got worker ID %d, out of range [0,3) — handle negative keys",
				r.Value, r.WorkerID)
		}
	}
}

func TestPartitionByEdgeCases(t *testing.T) {
	t.Run("empty stream", func(t *testing.T) {
		var got []StringResult
		withTimeout(t, 5*time.Second, "PartitionBy empty", func() {
			got = PartitionBy(feedStrings(), 3, userKey)
		})
		if got == nil {
			t.Errorf("returned nil, want an empty non-nil slice")
		}
		if len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})

	t.Run("zero workers treated as one", func(t *testing.T) {
		vals := []string{"user1:a", "user2:b"}
		var got []StringResult
		withTimeout(t, 5*time.Second, "PartitionBy zero workers", func() {
			got = PartitionBy(feedStrings(vals...), 0, userKey)
		})
		if len(got) != len(vals) {
			t.Fatalf("got %d results, want %d", len(got), len(vals))
		}
		for _, r := range got {
			if r.WorkerID != 0 {
				t.Errorf("worker ID %d with workers<=0, want 0", r.WorkerID)
			}
		}
	})

	t.Run("more keys than workers", func(t *testing.T) {
		// Collisions are expected: distinct keys may share a worker.
		var vals []string
		for i := 0; i < 20; i++ {
			vals = append(vals, fmt.Sprintf("user%d:v", i))
		}

		var got []StringResult
		withTimeout(t, 5*time.Second, "PartitionBy many keys", func() {
			got = PartitionBy(feedStrings(vals...), 3, userKey)
		})

		if len(got) != len(vals) {
			t.Fatalf("got %d results, want %d", len(got), len(vals))
		}
		byKey := map[int]int{}
		for _, r := range got {
			k := userKey(r.Value)
			if prev, seen := byKey[k]; seen && prev != r.WorkerID {
				t.Fatalf("key %d landed on workers %d and %d", k, prev, r.WorkerID)
			}
			byKey[k] = r.WorkerID
		}
	})
}
