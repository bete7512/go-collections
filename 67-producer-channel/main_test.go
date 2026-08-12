package main

import (
	"slices"
	"testing"
	"time"
)

// withTimeout runs fn and fails the test if it does not finish in time.
// Without this, a producer that forgets to close() hangs the suite forever
// instead of reporting a failure.
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
		t.Fatalf("%s did not finish within %v — is the channel ever closed?", what, d)
	}
}

func TestProduceOrder(t *testing.T) {
	for _, n := range []int{1, 5, 100} {
		var got []int
		withTimeout(t, 2*time.Second, "Collect(Produce(n))", func() {
			got = Collect(Produce(n))
		})

		want := make([]int, n)
		for i := range want {
			want[i] = i + 1
		}
		if !slices.Equal(got, want) {
			t.Errorf("Collect(Produce(%d)) = %v, want %v", n, got, want)
		}
	}
}

func TestProduceZeroAndNegative(t *testing.T) {
	for _, n := range []int{0, -1, -50} {
		var got []int
		withTimeout(t, 2*time.Second, "Collect on empty producer", func() {
			got = Collect(Produce(n))
		})

		if got == nil {
			t.Errorf("Collect(Produce(%d)) returned nil, want an empty non-nil slice", n)
		}
		if len(got) != 0 {
			t.Errorf("Collect(Produce(%d)) = %v, want empty", n, got)
		}
	}
}

func TestProduceReturnsImmediately(t *testing.T) {
	// Produce must launch a goroutine and return the channel without waiting
	// for the sends to be consumed.
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = Produce(1000) // large enough that sending it all would take time
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("Produce blocked instead of returning the channel immediately")
	}
}

func TestChannelClosesAfterDraining(t *testing.T) {
	withTimeout(t, 2*time.Second, "manual drain", func() {
		ch := Produce(2)

		if v, ok := <-ch; !ok || v != 1 {
			t.Errorf("first receive = (%d, %v), want (1, true)", v, ok)
		}
		if v, ok := <-ch; !ok || v != 2 {
			t.Errorf("second receive = (%d, %v), want (2, true)", v, ok)
		}

		v, ok := <-ch
		if ok {
			t.Errorf("receive after draining = (%d, true), want (0, false) — the sender must close", v)
		}
		if v != 0 {
			t.Errorf("value from a closed channel = %d, want the zero value 0", v)
		}

		// Receiving from a closed channel is repeatable and never blocks.
		for i := 0; i < 3; i++ {
			if _, ok := <-ch; ok {
				t.Fatalf("receive %d after close reported ok = true", i)
			}
		}
	})
}

func TestClosedChannelNeverBlocks(t *testing.T) {
	withTimeout(t, 2*time.Second, "select on closed channel", func() {
		ch := Produce(0) // closed with nothing sent

		// A closed channel is always ready: the default branch must not win.
		select {
		case _, ok := <-ch:
			if ok {
				t.Errorf("received a value from a channel that should be closed and empty")
			}
		default:
			t.Errorf("receiving from a closed channel took the default branch — it should always be ready")
		}
	})
}

func TestCollectOnClosedEmptyChannel(t *testing.T) {
	withTimeout(t, 2*time.Second, "Collect on a pre-closed channel", func() {
		ch := make(chan int)
		close(ch)

		got := Collect(ch)
		if got == nil {
			t.Errorf("Collect returned nil, want an empty non-nil slice")
		}
		if len(got) != 0 {
			t.Errorf("Collect = %v, want empty", got)
		}
	})
}

func TestProduceSquares(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"positives", []int{1, 2, 3, 4}, []int{1, 4, 9, 16}},
		{"negatives square positive", []int{-1, -2, -3}, []int{1, 4, 9}},
		{"zeros", []int{0, 0}, []int{0, 0}},
		{"mixed", []int{-2, 0, 3}, []int{4, 0, 9}},
		{"single", []int{7}, []int{49}},
		{"empty", []int{}, []int{}},
		{"nil", nil, []int{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []int
			withTimeout(t, 2*time.Second, "Collect(ProduceSquares(...))", func() {
				got = Collect(ProduceSquares(tc.input))
			})

			if !slices.Equal(got, tc.want) {
				t.Errorf("Collect(ProduceSquares(%v)) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestLargeProduction(t *testing.T) {
	const n = 10_000

	var got []int
	withTimeout(t, 10*time.Second, "large production", func() {
		got = Collect(Produce(n))
	})

	if len(got) != n {
		t.Fatalf("collected %d values, want %d", len(got), n)
	}
	for i, v := range got {
		if v != i+1 {
			t.Fatalf("got[%d] = %d, want %d — order must be preserved with one sender", i, v, i+1)
		}
	}
}

func TestRepeatedRuns(t *testing.T) {
	// Ordering with a single sender and single receiver must be stable.
	want := []int{1, 2, 3, 4, 5}

	for run := 0; run < 100; run++ {
		var got []int
		withTimeout(t, 2*time.Second, "repeated run", func() {
			got = Collect(Produce(5))
		})
		if !slices.Equal(got, want) {
			t.Fatalf("run %d: got %v, want %v", run, got, want)
		}
	}
}
