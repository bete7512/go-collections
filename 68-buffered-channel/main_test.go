package main

import (
	"testing"
	"time"
)

func TestBufferFillsAndRefuses(t *testing.T) {
	ch := make(chan int, 2)

	if !TrySend(ch, 1) {
		t.Fatalf("TrySend into an empty cap-2 buffer = false, want true")
	}
	if l, c := Stats(ch); l != 1 || c != 2 {
		t.Errorf("after one send: len=%d cap=%d, want 1 and 2", l, c)
	}

	if !TrySend(ch, 2) {
		t.Fatalf("TrySend filling the buffer = false, want true")
	}
	if l, _ := Stats(ch); l != 2 {
		t.Errorf("after two sends: len=%d, want 2", l)
	}

	if TrySend(ch, 3) {
		t.Fatalf("TrySend into a full buffer = true, want false")
	}
	if l, _ := Stats(ch); l != 2 {
		t.Errorf("a refused send changed len to %d, want 2 — nothing should have been stored", l)
	}

	// Free a slot, then the next send fits.
	if v, ok := TryReceive(ch); !ok || v != 1 {
		t.Fatalf("TryReceive = (%d, %v), want (1, true) — FIFO order", v, ok)
	}
	if !TrySend(ch, 3) {
		t.Errorf("TrySend after freeing a slot = false, want true")
	}
}

func TestCapacitiesOneAndFive(t *testing.T) {
	for _, capacity := range []int{1, 5} {
		ch := make(chan int, capacity)

		accepted := 0
		for i := 0; i < capacity+3; i++ {
			if TrySend(ch, i) {
				accepted++
			}
		}
		if accepted != capacity {
			t.Errorf("cap %d: accepted %d sends, want %d", capacity, accepted, capacity)
		}
		if l, c := Stats(ch); l != capacity || c != capacity {
			t.Errorf("cap %d: len=%d cap=%d, want both %d", capacity, l, c, capacity)
		}
	}
}

func TestDrainPreservesFIFO(t *testing.T) {
	ch := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		TrySend(ch, i)
	}

	for want := 1; want <= 5; want++ {
		v, ok := TryReceive(ch)
		if !ok || v != want {
			t.Fatalf("TryReceive = (%d, %v), want (%d, true)", v, ok, want)
		}
	}

	if v, ok := TryReceive(ch); ok {
		t.Errorf("TryReceive on a drained buffer = (%d, true), want (_, false)", v)
	}
	if l, _ := Stats(ch); l != 0 {
		t.Errorf("len after draining = %d, want 0", l)
	}
}

func TestUnbufferedNeverReadyAlone(t *testing.T) {
	ch := make(chan int) // cap 0

	if l, c := Stats(ch); l != 0 || c != 0 {
		t.Errorf("unbuffered: len=%d cap=%d, want 0 and 0", l, c)
	}
	if TrySend(ch, 1) {
		t.Errorf("TrySend on an unbuffered channel with no receiver = true, want false")
	}
	if _, ok := TryReceive(ch); ok {
		t.Errorf("TryReceive on an unbuffered channel with no sender = true, want false")
	}
}

func TestUnbufferedSucceedsWithWaitingReceiver(t *testing.T) {
	ch := make(chan int)
	ready := make(chan struct{})
	got := make(chan int, 1)

	go func() {
		close(ready) // announce that we are about to block on receive
		got <- <-ch
	}()

	<-ready
	// The receiver is parked (or about to be). Retry briefly without sleeping
	// on a fixed duration: this loop is bounded and deterministic in outcome.
	deadline := time.Now().Add(2 * time.Second)
	sent := false
	for time.Now().Before(deadline) {
		if TrySend(ch, 42) {
			sent = true
			break
		}
	}

	if !sent {
		t.Fatalf("TrySend never succeeded with a receiver waiting — an unbuffered send is ready when a receiver is")
	}
	if v := <-got; v != 42 {
		t.Errorf("receiver got %d, want 42", v)
	}
}

func TestNilChannelBlocksBothWays(t *testing.T) {
	// A nil channel blocks forever in both directions; the default branch
	// must win, with no panic. (#70 uses this to disable a select case.)
	var ch chan int

	if TrySend(ch, 1) {
		t.Errorf("TrySend on a nil channel = true, want false")
	}
	if v, ok := TryReceive(ch); ok {
		t.Errorf("TryReceive on a nil channel = (%d, true), want (_, false)", v)
	}
	if l, c := Stats(ch); l != 0 || c != 0 {
		t.Errorf("nil channel: len=%d cap=%d, want 0 and 0", l, c)
	}
}

func TestClosedChannelReceiveNeverBlocks(t *testing.T) {
	ch := make(chan int, 2)
	TrySend(ch, 7)
	close(ch)

	// Buffered values are still delivered after close.
	if v, ok := TryReceive(ch); !ok || v != 7 {
		t.Errorf("TryReceive = (%d, %v), want (7, true) — close does not discard buffered values", v, ok)
	}

	// Then it yields the zero value, and the receive is always ready.
	v, ok := TryReceive(ch)
	if !ok {
		t.Errorf("TryReceive on a drained closed channel = not ready; receives never block on a closed channel")
	}
	if v != 0 {
		t.Errorf("value from a closed channel = %d, want 0", v)
	}
}

func TestSendOnClosedChannelPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TrySend on a closed channel did not panic — select cannot protect a send on a closed channel")
		}
	}()

	ch := make(chan int, 1)
	close(ch)
	TrySend(ch, 1)
}

func TestFillBuffer(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		vals     []int
		want     int
	}{
		{"more values than capacity", 3, []int{1, 2, 3, 4, 5}, 3},
		{"fewer values than capacity", 5, []int{1, 2}, 2},
		{"exactly capacity", 4, []int{1, 2, 3, 4}, 4},
		{"empty values", 3, []int{}, 0},
		{"nil values", 3, nil, 0},
		{"unbuffered accepts nothing", 0, []int{1, 2}, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := make(chan int, tc.capacity)

			got := FillBuffer(ch, tc.vals)
			if got != tc.want {
				t.Errorf("FillBuffer = %d, want %d", got, tc.want)
			}
			if l, _ := Stats(ch); l != tc.want {
				t.Errorf("len after FillBuffer = %d, want %d", l, tc.want)
			}
		})
	}

	t.Run("nil channel", func(t *testing.T) {
		var ch chan int
		if got := FillBuffer(ch, []int{1, 2, 3}); got != 0 {
			t.Errorf("FillBuffer on a nil channel = %d, want 0", got)
		}
	})
}

func TestFillBufferPreservesOrder(t *testing.T) {
	ch := make(chan int, 3)

	FillBuffer(ch, []int{10, 20, 30, 40})

	for _, want := range []int{10, 20, 30} {
		if v, ok := TryReceive(ch); !ok || v != want {
			t.Fatalf("TryReceive = (%d, %v), want (%d, true)", v, ok, want)
		}
	}
	if _, ok := TryReceive(ch); ok {
		t.Errorf("a fourth value was stored; only the first cap values should land")
	}
}

func TestLargeBuffer(t *testing.T) {
	const n = 1000
	ch := make(chan int, n)

	vals := make([]int, n+50)
	for i := range vals {
		vals[i] = i
	}

	if got := FillBuffer(ch, vals); got != n {
		t.Fatalf("FillBuffer = %d, want %d", got, n)
	}
	if l, c := Stats(ch); l != n || c != n {
		t.Fatalf("len=%d cap=%d, want both %d", l, c, n)
	}

	for i := 0; i < n; i++ {
		v, ok := TryReceive(ch)
		if !ok || v != i {
			t.Fatalf("TryReceive #%d = (%d, %v), want (%d, true)", i, v, ok, i)
		}
	}
}
