package main

import (
	"slices"
	"testing"
	"time"
)

// readFunc lets the same table run against both implementations.
type readFunc func(ch <-chan int, d time.Duration) (int, bool)

var implementations = []struct {
	name string
	fn   readFunc
}{
	{"ReadWithTimeout", func(ch <-chan int, d time.Duration) (int, bool) { return ReadWithTimeout(ch, d) }},
	{"ReadWithTimer", func(ch <-chan int, d time.Duration) (int, bool) { return ReadWithTimer(ch, d) }},
}

func TestValueArrivesInTime(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			ch := make(chan int, 1)
			ch <- 42

			start := time.Now()
			v, ok := impl.fn(ch, 2*time.Second)
			elapsed := time.Since(start)

			if !ok || v != 42 {
				t.Fatalf("= (%d, %v), want (42, true)", v, ok)
			}
			if elapsed > 500*time.Millisecond {
				t.Errorf("took %v for a buffered value; it should return immediately", elapsed)
			}
		})
	}
}

func TestValueArrivesAfterDelay(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			ch := make(chan int)
			go func() {
				time.Sleep(20 * time.Millisecond)
				ch <- 7
			}()

			v, ok := impl.fn(ch, 2*time.Second)
			if !ok || v != 7 {
				t.Errorf("= (%d, %v), want (7, true) — the value arrived well before the deadline", v, ok)
			}
		})
	}
}

func TestTimeoutFires(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			ch := make(chan int) // nothing will ever arrive

			start := time.Now()
			v, ok := impl.fn(ch, 60*time.Millisecond)
			elapsed := time.Since(start)

			if ok {
				t.Fatalf("= (%d, true), want (0, false) after the deadline", v)
			}
			if v != 0 {
				t.Errorf("value on timeout = %d, want 0", v)
			}
			if elapsed < 40*time.Millisecond {
				t.Errorf("returned after %v, well before the 60ms deadline", elapsed)
			}
			if elapsed > 2*time.Second {
				t.Errorf("returned after %v, far beyond the 60ms deadline", elapsed)
			}
		})
	}
}

func TestClosedChannelReturnsFastNotOnTimeout(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			ch := make(chan int)
			close(ch)

			start := time.Now()
			v, ok := impl.fn(ch, 5*time.Second)
			elapsed := time.Since(start)

			if ok {
				t.Fatalf("= (%d, true), want (0, false) for a closed channel", v)
			}
			if elapsed > time.Second {
				t.Errorf("took %v — a closed channel is immediately ready and must not wait for the deadline",
					elapsed)
			}
		})
	}
}

func TestClosedChannelDeliversBufferedValueFirst(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			ch := make(chan int, 1)
			ch <- 5
			close(ch)

			v, ok := impl.fn(ch, time.Second)
			if !ok || v != 5 {
				t.Errorf("= (%d, %v), want (5, true) — buffered values survive close", v, ok)
			}
		})
	}
}

func TestNilChannelTimesOut(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			var ch chan int // nil: never ready

			start := time.Now()
			_, ok := impl.fn(ch, 50*time.Millisecond)
			elapsed := time.Since(start)

			if ok {
				t.Fatalf("a nil channel reported a value")
			}
			if elapsed > 2*time.Second {
				t.Errorf("took %v, want roughly the 50ms deadline", elapsed)
			}
		})
	}
}

func TestZeroDurationOnEmptyChannel(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			ch := make(chan int)

			start := time.Now()
			_, ok := impl.fn(ch, 0)
			elapsed := time.Since(start)

			if ok {
				t.Errorf("d=0 on an empty channel reported a value")
			}
			if elapsed > time.Second {
				t.Errorf("d=0 took %v, want a prompt return", elapsed)
			}
		})
	}
}

func TestReadWithTimerInLoop(t *testing.T) {
	// The loop is where time.After allocates a timer per iteration.
	for i := 0; i < 500; i++ {
		ch := make(chan int, 1)
		ch <- i

		v, ok := ReadWithTimer(ch, time.Second)
		if !ok || v != i {
			t.Fatalf("iteration %d: = (%d, %v), want (%d, true)", i, v, ok, i)
		}
	}
}

func TestDrainWithIdleTimeoutFastStream(t *testing.T) {
	// A steady stream must never trip the idle deadline: it resets per value.
	const n = 1000
	ch := make(chan int, n)
	for i := 0; i < n; i++ {
		ch <- i
	}
	close(ch)

	got := DrainWithIdleTimeout(ch, 50*time.Millisecond)

	if len(got) != n {
		t.Fatalf("collected %d values, want %d — the deadline must reset on every arrival", len(got), n)
	}
	for i, v := range got {
		if v != i {
			t.Fatalf("got[%d] = %d, want %d", i, v, i)
		}
	}
}

func TestDrainWithIdleTimeoutStall(t *testing.T) {
	ch := make(chan int)
	go func() {
		ch <- 1
		ch <- 2
		ch <- 3
		// then stalls forever without closing
		select {}
	}()

	start := time.Now()
	got := DrainWithIdleTimeout(ch, 80*time.Millisecond)
	elapsed := time.Since(start)

	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("got %v, want [1 2 3] — everything received before the stall", got)
	}
	if elapsed > 3*time.Second {
		t.Errorf("took %v; the idle timeout should fire ~80ms after the last value", elapsed)
	}
}

func TestDrainWithIdleTimeoutNormalClose(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	close(ch)

	start := time.Now()
	got := DrainWithIdleTimeout(ch, 5*time.Second)
	elapsed := time.Since(start)

	if !slices.Equal(got, []int{10, 20}) {
		t.Errorf("got %v, want [10 20]", got)
	}
	if elapsed > time.Second {
		t.Errorf("took %v — a close must end the drain immediately, not wait out the idle window", elapsed)
	}
}

func TestDrainWithIdleTimeoutEmptyCases(t *testing.T) {
	t.Run("closed immediately", func(t *testing.T) {
		ch := make(chan int)
		close(ch)

		got := DrainWithIdleTimeout(ch, time.Second)
		if got == nil {
			t.Errorf("returned nil, want an empty non-nil slice")
		}
		if len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})

	t.Run("never sends", func(t *testing.T) {
		ch := make(chan int)

		start := time.Now()
		got := DrainWithIdleTimeout(ch, 50*time.Millisecond)
		elapsed := time.Since(start)

		if len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
		if elapsed > 2*time.Second {
			t.Errorf("took %v, want roughly the 50ms idle window", elapsed)
		}
	})
}
