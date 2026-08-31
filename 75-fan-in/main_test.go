package main

import (
	"fmt"
	"runtime"
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
		t.Fatalf("%s did not finish within %v — the merged channel must be closed by a "+
			"dedicated goroutine once every forwarder has finished", what, d)
	}
}

// feed returns a closed channel carrying vals.
func feed(vals ...int) <-chan int {
	ch := make(chan int, len(vals))
	for _, v := range vals {
		ch <- v
	}
	close(ch)
	return ch
}

func TestMergeCollectsEverything(t *testing.T) {
	tests := []struct {
		name    string
		sources [][]int
	}{
		{"three sources", [][]int{{1, 2}, {10, 20, 30}, {100}}},
		{"single source", [][]int{{1, 2, 3}}},
		{"no sources", nil},
		{"one empty source", [][]int{{}}},
		{"all sources empty", [][]int{{}, {}, {}}},
		{"mixed empty and full", [][]int{{}, {1, 2, 3}, {}, {4}}},
		{"duplicate values across sources", [][]int{{7, 7}, {7}, {7, 7, 7}}},
		{"negatives and zero", [][]int{{-1, 0}, {1, -2}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chans := make([]<-chan int, 0, len(tc.sources))
			var want []int
			for _, src := range tc.sources {
				chans = append(chans, feed(src...))
				want = append(want, src...)
			}

			var got []int
			withTimeout(t, 5*time.Second, "MergeCollect", func() {
				got = MergeCollect(chans...)
			})

			if got == nil {
				t.Fatalf("MergeCollect returned nil, want an empty non-nil slice")
			}

			gotSorted := slices.Clone(got)
			slices.Sort(gotSorted)
			slices.Sort(want)
			if want == nil {
				want = []int{}
			}
			if !slices.Equal(gotSorted, want) {
				t.Errorf("sorted merged values = %v, want %v", gotSorted, want)
			}
		})
	}
}

func TestMergeZeroInputsReturnsClosedChannel(t *testing.T) {
	// With no inputs the WaitGroup is already at zero, so the closer fires
	// immediately. A caller ranging the result must not block.
	withTimeout(t, 5*time.Second, "Merge with no inputs", func() {
		out := Merge()

		count := 0
		for range out {
			count++
		}
		if count != 0 {
			t.Errorf("received %d values from an empty merge, want 0", count)
		}
	})
}

func TestMergeReturnsImmediately(t *testing.T) {
	// Merge must return the channel without waiting for forwarding to finish.
	slow := make(chan int)
	go func() {
		defer close(slow)
		for i := 0; i < 100; i++ {
			slow <- i
		}
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = Merge(slow, feed(1, 2, 3))
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("Merge blocked instead of returning the merged channel immediately")
	}

	// Drain so the goroutines can finish.
	withTimeout(t, 5*time.Second, "drain", func() {
		for range Merge(feed()) {
		}
	})
}

func TestMergeSkipsNilChannels(t *testing.T) {
	// A nil channel never becomes ready; its forwarder would block forever.
	var nilCh chan int

	var got []int
	withTimeout(t, 5*time.Second, "MergeCollect with a nil channel", func() {
		got = MergeCollect(feed(1, 2), nilCh, feed(3))
	})

	slices.Sort(got)
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("got %v, want [1 2 3] — nil inputs must be skipped, not awaited", got)
	}
}

func TestMergePreservesPerSourceOrder(t *testing.T) {
	// Cross-source interleaving is unspecified, but each source's own values
	// must stay in order: one forwarder receives and sends sequentially.
	a := feed(1, 2, 3, 4, 5)          // units
	b := feed(10, 20, 30, 40, 50)     // tens
	c := feed(100, 200, 300)          // hundreds

	var got []int
	withTimeout(t, 5*time.Second, "MergeCollect", func() {
		got = MergeCollect(a, b, c)
	})

	var aSeq, bSeq, cSeq []int
	for _, v := range got {
		switch {
		case v < 10:
			aSeq = append(aSeq, v)
		case v < 100:
			bSeq = append(bSeq, v)
		default:
			cSeq = append(cSeq, v)
		}
	}

	if !slices.Equal(aSeq, []int{1, 2, 3, 4, 5}) {
		t.Errorf("source a subsequence = %v, want [1 2 3 4 5]", aSeq)
	}
	if !slices.Equal(bSeq, []int{10, 20, 30, 40, 50}) {
		t.Errorf("source b subsequence = %v, want [10 20 30 40 50]", bSeq)
	}
	if !slices.Equal(cSeq, []int{100, 200, 300}) {
		t.Errorf("source c subsequence = %v, want [100 200 300]", cSeq)
	}
}

func TestMergeUnequalLengths(t *testing.T) {
	long := make([]int, 1000)
	for i := range long {
		long[i] = i
	}

	var got []int
	withTimeout(t, 10*time.Second, "MergeCollect unequal", func() {
		got = MergeCollect(feed(-1), feed(long...), feed(-2, -3))
	})

	if len(got) != 1003 {
		t.Fatalf("got %d values, want 1003 — a short source finishing early must not stall the merge",
			len(got))
	}
}

func TestMergeManySources(t *testing.T) {
	const sources, perSource = 8, 500

	chans := make([]<-chan int, 0, sources)
	for s := 0; s < sources; s++ {
		vals := make([]int, perSource)
		for i := range vals {
			vals[i] = s*perSource + i
		}
		chans = append(chans, feed(vals...))
	}

	var got []int
	withTimeout(t, 15*time.Second, "MergeCollect many sources", func() {
		got = MergeCollect(chans...)
	})

	if len(got) != sources*perSource {
		t.Fatalf("got %d values, want %d", len(got), sources*perSource)
	}
	seen := map[int]int{}
	for _, v := range got {
		seen[v]++
	}
	for i := 0; i < sources*perSource; i++ {
		if seen[i] != 1 {
			t.Fatalf("value %d appeared %d times, want exactly 1", i, seen[i])
		}
	}
}

func TestMergeRepeatedRuns(t *testing.T) {
	want := []int{1, 2, 3, 10, 20, 100}

	for run := 0; run < 50; run++ {
		var got []int
		withTimeout(t, 5*time.Second, "repeated MergeCollect", func() {
			got = MergeCollect(feed(1, 2, 3), feed(10, 20), feed(100))
		})

		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Fatalf("run %d: got %v, want %v", run, got, want)
		}
	}
}

func TestMergeDoesNotLeakGoroutines(t *testing.T) {
	// After the merged channel closes, every forwarder and the closer must
	// have returned.
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	before := runtime.NumGoroutine()

	for i := 0; i < 20; i++ {
		withTimeout(t, 5*time.Second, "MergeCollect", func() {
			MergeCollect(feed(1, 2, 3), feed(4, 5), feed(6))
		})
	}

	runtime.GC()
	time.Sleep(300 * time.Millisecond)
	after := runtime.NumGoroutine()

	if after > before+5 {
		t.Errorf("goroutine count went from %d to %d — forwarders or the closer are leaking",
			before, after)
	}
}

// ---------- labeled merge ----------

func TestMergeWithLabels(t *testing.T) {
	sources := map[string]<-chan int{
		"api":   feed(1, 2, 3),
		"cache": feed(10, 20),
		"db":    feed(100),
	}

	var got []Labeled
	withTimeout(t, 5*time.Second, "MergeWithLabels", func() {
		got = MergeWithLabels(sources)
	})

	if len(got) != 6 {
		t.Fatalf("got %d labeled values, want 6", len(got))
	}

	bySource := map[string][]int{}
	for _, l := range got {
		bySource[l.Source] = append(bySource[l.Source], l.Value)
	}

	// Correct label on every value, and per-source order preserved.
	want := map[string][]int{
		"api":   {1, 2, 3},
		"cache": {10, 20},
		"db":    {100},
	}
	for name, wantVals := range want {
		gotVals, ok := bySource[name]
		if !ok {
			t.Errorf("source %q missing from the merged output", name)
			continue
		}
		if !slices.Equal(gotVals, wantVals) {
			t.Errorf("source %q values = %v, want %v (per-source order must be preserved)",
				name, gotVals, wantVals)
		}
	}
	if len(bySource) != len(want) {
		t.Errorf("got %d distinct sources, want %d", len(bySource), len(want))
	}
}

func TestMergeWithLabelsEdgeCases(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		var got []Labeled
		withTimeout(t, 5*time.Second, "MergeWithLabels empty", func() {
			got = MergeWithLabels(map[string]<-chan int{})
		})
		if got == nil {
			t.Errorf("returned nil, want an empty non-nil slice")
		}
		if len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})

	t.Run("nil map", func(t *testing.T) {
		var got []Labeled
		withTimeout(t, 5*time.Second, "MergeWithLabels nil map", func() {
			got = MergeWithLabels(nil)
		})
		if len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})

	t.Run("single source", func(t *testing.T) {
		var got []Labeled
		withTimeout(t, 5*time.Second, "MergeWithLabels single", func() {
			got = MergeWithLabels(map[string]<-chan int{"only": feed(1, 2)})
		})
		if len(got) != 2 {
			t.Fatalf("got %d values, want 2", len(got))
		}
		for _, l := range got {
			if l.Source != "only" {
				t.Errorf("value %d labeled %q, want \"only\"", l.Value, l.Source)
			}
		}
	})

	t.Run("source with no values", func(t *testing.T) {
		var got []Labeled
		withTimeout(t, 5*time.Second, "MergeWithLabels empty source", func() {
			got = MergeWithLabels(map[string]<-chan int{
				"empty": feed(),
				"full":  feed(1, 2),
			})
		})
		if len(got) != 2 {
			t.Fatalf("got %d values, want 2", len(got))
		}
		for _, l := range got {
			if l.Source != "full" {
				t.Errorf("unexpected label %q", l.Source)
			}
		}
	})
}

func TestMergeWithLabelsMany(t *testing.T) {
	const sources, perSource = 5, 20

	m := map[string]<-chan int{}
	for s := 0; s < sources; s++ {
		vals := make([]int, perSource)
		for i := range vals {
			vals[i] = s*100 + i
		}
		m[fmt.Sprintf("src%d", s)] = feed(vals...)
	}

	var got []Labeled
	withTimeout(t, 10*time.Second, "MergeWithLabels many", func() {
		got = MergeWithLabels(m)
	})

	if len(got) != sources*perSource {
		t.Fatalf("got %d values, want %d", len(got), sources*perSource)
	}

	bySource := map[string][]int{}
	for _, l := range got {
		bySource[l.Source] = append(bySource[l.Source], l.Value)
	}
	for s := 0; s < sources; s++ {
		name := fmt.Sprintf("src%d", s)
		vals := bySource[name]
		if len(vals) != perSource {
			t.Fatalf("source %s has %d values, want %d", name, len(vals), perSource)
		}
		for i, v := range vals {
			if v != s*100+i {
				t.Fatalf("source %s value %d = %d, want %d (order lost)", name, i, v, s*100+i)
			}
		}
	}
}
