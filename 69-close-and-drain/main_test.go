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
		t.Fatalf("%s did not finish within %v", what, d)
	}
}

func TestDrainAllDeliversBufferedValues(t *testing.T) {
	tests := []struct {
		name string
		vals []int
	}{
		{"three values", []int{1, 2, 3}},
		{"single value", []int{42}},
		{"no values", []int{}},
		{"includes a zero", []int{0, 5, 0}},
		{"negatives", []int{-1, -2, -3}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := make(chan int, len(tc.vals)+1)
			for _, v := range tc.vals {
				ch <- v
			}
			close(ch) // values are still queued behind the close

			var got []int
			withTimeout(t, 2*time.Second, "DrainAll", func() {
				got = DrainAll(ch)
			})

			if got == nil {
				t.Fatalf("DrainAll returned nil, want an empty non-nil slice")
			}
			if !slices.Equal(got, tc.vals) {
				t.Errorf("DrainAll = %v, want %v — close must not discard buffered values", got, tc.vals)
			}
		})
	}
}

func TestDrainAllOnPreClosedEmptyChannel(t *testing.T) {
	ch := make(chan int)
	close(ch)

	var got []int
	withTimeout(t, 2*time.Second, "DrainAll on closed empty channel", func() {
		got = DrainAll(ch)
	})

	if got == nil || len(got) != 0 {
		t.Errorf("DrainAll = %v, want an empty non-nil slice", got)
	}
}

func TestDrainWithOKReceiveCount(t *testing.T) {
	for _, n := range []int{0, 1, 5, 20} {
		ch := make(chan int, n+1)
		for i := 0; i < n; i++ {
			ch <- i
		}
		close(ch)

		var vals []int
		var receives int
		withTimeout(t, 2*time.Second, "DrainWithOK", func() {
			vals, receives = DrainWithOK(ch)
		})

		if len(vals) != n {
			t.Errorf("n=%d: got %d values, want %d", n, len(vals), n)
		}
		if receives != n+1 {
			t.Errorf("n=%d: receives = %d, want %d — the final receive is the one reporting ok==false",
				n, receives, n+1)
		}
	}
}

func TestRealZeroVersusClosedZero(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 0 // a genuine zero
	close(ch)

	withTimeout(t, 2*time.Second, "zero disambiguation", func() {
		v, ok := <-ch
		if v != 0 || !ok {
			t.Errorf("first receive = (%d, %v), want (0, true) — a real zero was sent", v, ok)
		}

		v, ok = <-ch
		if v != 0 || ok {
			t.Errorf("second receive = (%d, %v), want (0, false) — now it is the closed-channel zero", v, ok)
		}
	})
}

func TestCountAfterClose(t *testing.T) {
	for _, n := range []int{1, 5, 100} {
		ch := make(chan int, 1)
		ch <- 7
		close(ch)
		<-ch // drain the single value

		var flags []bool
		withTimeout(t, 2*time.Second, "CountAfterClose", func() {
			flags = CountAfterClose(ch, n)
		})

		if len(flags) != n {
			t.Fatalf("n=%d: got %d flags, want %d", n, len(flags), n)
		}
		for i, ok := range flags {
			if ok {
				t.Fatalf("n=%d: receive %d reported ok=true, want false for every post-close receive", n, i)
			}
		}
	}
}

func TestDoubleClosePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("closing an already-closed channel did not panic")
		}
	}()

	ch := make(chan int)
	close(ch)
	close(ch)
}

func TestSendOnClosedPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("sending on a closed channel did not panic")
		}
	}()

	ch := make(chan int, 1)
	close(ch)
	ch <- 1
}

func TestCloseNilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("closing a nil channel did not panic")
		}
	}()

	var ch chan int
	close(ch)
}

func TestSafeClose(t *testing.T) {
	ch := make(chan int)

	if !SafeClose(ch) {
		t.Fatalf("SafeClose on an open channel = false, want true")
	}
	if SafeClose(ch) {
		t.Errorf("SafeClose on an already-closed channel = true, want false (the panic must be recovered)")
	}

	// The channel really is closed.
	withTimeout(t, 2*time.Second, "receive after SafeClose", func() {
		if _, ok := <-ch; ok {
			t.Errorf("receive after SafeClose reported ok=true")
		}
	})
}

func TestSendAndClose(t *testing.T) {
	tests := []struct {
		name string
		vals []int
	}{
		{"several values", []int{1, 2, 3, 4}},
		{"single value", []int{9}},
		{"empty", []int{}},
		{"nil", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := make(chan int, len(tc.vals)+1)

			var got []int
			withTimeout(t, 2*time.Second, "SendAndClose then DrainAll", func() {
				SendAndClose(ch, tc.vals)
				got = DrainAll(ch)
			})

			want := tc.vals
			if want == nil {
				want = []int{}
			}
			if !slices.Equal(got, want) {
				t.Errorf("drained %v, want %v", got, want)
			}
		})
	}
}

func TestProducerCloseDrainRoundTrip(t *testing.T) {
	const n = 10_000
	ch := make(chan int)

	go func() {
		for i := 0; i < n; i++ {
			ch <- i
		}
		close(ch) // the sender closes
	}()

	var got []int
	withTimeout(t, 10*time.Second, "large round trip", func() {
		got = DrainAll(ch)
	})

	if len(got) != n {
		t.Fatalf("drained %d values, want %d", len(got), n)
	}
	for i, v := range got {
		if v != i {
			t.Fatalf("got[%d] = %d, want %d", i, v, i)
		}
	}
}
