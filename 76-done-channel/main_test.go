package main

import (
	"runtime"
	"sync"
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
		t.Fatalf("%s did not finish within %v — every blocking operation in the goroutine "+
			"must race against <-done inside a select", what, d)
	}
}

func closedDone() chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

// ---------- CountUntilDone ----------

func TestCountUntilDoneStopsOnClose(t *testing.T) {
	done := make(chan struct{})
	result := make(chan int, 1)

	go func() { result <- CountUntilDone(done) }()

	time.Sleep(30 * time.Millisecond) // let it spin
	close(done)

	select {
	case n := <-result:
		if n <= 0 {
			t.Errorf("count = %d, want a positive number of iterations", n)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("CountUntilDone did not return after close(done)")
	}
}

func TestCountUntilDoneAlreadyClosed(t *testing.T) {
	withTimeout(t, 3*time.Second, "CountUntilDone with a pre-closed channel", func() {
		CountUntilDone(closedDone())
	})
}

func TestCountUntilDoneRepeatedCycles(t *testing.T) {
	for run := 0; run < 50; run++ {
		done := make(chan struct{})
		result := make(chan int, 1)

		go func() { result <- CountUntilDone(done) }()
		close(done)

		select {
		case <-result:
		case <-time.After(3 * time.Second):
			t.Fatalf("run %d: CountUntilDone did not return", run)
		}
	}
}

// ---------- WorkersUntilDone ----------

func TestWorkersUntilDoneReleasesAll(t *testing.T) {
	for _, workers := range []int{1, 5, 50} {
		done := make(chan struct{})
		result := make(chan []int, 1)

		go func() { result <- WorkersUntilDone(done, workers) }()

		time.Sleep(30 * time.Millisecond)
		close(done) // ONE close releases every worker

		select {
		case counts := <-result:
			if len(counts) != workers {
				t.Errorf("workers=%d: got %d counts, want %d", workers, len(counts), workers)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("workers=%d: WorkersUntilDone did not return after one close(done)", workers)
		}
	}
}

func TestWorkersUntilDoneEdgeCases(t *testing.T) {
	t.Run("zero workers treated as one", func(t *testing.T) {
		done := closedDone()
		var counts []int
		withTimeout(t, 3*time.Second, "WorkersUntilDone(0)", func() {
			counts = WorkersUntilDone(done, 0)
		})
		if len(counts) != 1 {
			t.Errorf("got %d counts, want 1", len(counts))
		}
	})

	t.Run("negative workers treated as one", func(t *testing.T) {
		done := closedDone()
		var counts []int
		withTimeout(t, 3*time.Second, "WorkersUntilDone(-3)", func() {
			counts = WorkersUntilDone(done, -3)
		})
		if len(counts) != 1 {
			t.Errorf("got %d counts, want 1", len(counts))
		}
	})

	t.Run("already closed returns immediately", func(t *testing.T) {
		var counts []int
		withTimeout(t, 3*time.Second, "WorkersUntilDone pre-closed", func() {
			counts = WorkersUntilDone(closedDone(), 4)
		})
		if len(counts) != 4 {
			t.Errorf("got %d counts, want 4", len(counts))
		}
	})
}

func TestWorkersUntilDoneNoLeak(t *testing.T) {
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	before := runtime.NumGoroutine()

	for i := 0; i < 10; i++ {
		done := make(chan struct{})
		result := make(chan []int, 1)
		go func() { result <- WorkersUntilDone(done, 10) }()
		close(done)

		select {
		case <-result:
		case <-time.After(5 * time.Second):
			t.Fatalf("iteration %d did not return", i)
		}
	}

	runtime.GC()
	time.Sleep(300 * time.Millisecond)
	after := runtime.NumGoroutine()

	if after > before+5 {
		t.Errorf("goroutines went from %d to %d — workers are leaking after done closed", before, after)
	}
}

// ---------- ProcessUntilDone ----------

func TestProcessUntilDoneStopsWithNoResultReader(t *testing.T) {
	// THE leak test: results is unbuffered and nobody receives from it.
	// A send placed outside the select blocks forever and done cannot
	// rescue it.
	done := make(chan struct{})
	jobs := make(chan int)
	results := make(chan int) // no reader, ever
	finished := make(chan int, 1)

	go func() {
		for i := 0; ; i++ {
			select {
			case jobs <- i:
			case <-done:
				return
			}
		}
	}()

	go func() { finished <- ProcessUntilDone(done, jobs, results) }()

	time.Sleep(50 * time.Millisecond) // let it block on the results send
	close(done)

	select {
	case <-finished:
	case <-time.After(3 * time.Second):
		t.Fatalf("ProcessUntilDone did not return — the send into results must be a select case " +
			"racing against <-done, otherwise the goroutine is uncancellable")
	}
}

func TestProcessUntilDoneStopsWhenJobsClose(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	jobs := make(chan int, 4)
	for i := 1; i <= 4; i++ {
		jobs <- i
	}
	close(jobs)

	results := make(chan int, 4)
	finished := make(chan int, 1)

	go func() { finished <- ProcessUntilDone(done, jobs, results) }()

	select {
	case n := <-finished:
		if n != 4 {
			t.Errorf("processed %d jobs, want 4", n)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("ProcessUntilDone did not return when the jobs channel closed")
	}
}

func TestProcessUntilDoneProcessesJobs(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	jobs := make(chan int, 3)
	jobs <- 1
	jobs <- 2
	jobs <- 3
	close(jobs)

	results := make(chan int, 3)
	finished := make(chan int, 1)

	go func() { finished <- ProcessUntilDone(done, jobs, results) }()

	select {
	case n := <-finished:
		if n != 3 {
			t.Fatalf("processed %d jobs, want 3", n)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("ProcessUntilDone did not return")
	}

	close(results)
	count := 0
	for range results {
		count++
	}
	if count != 3 {
		t.Errorf("got %d results, want 3 — every processed job must be delivered", count)
	}
}

func TestProcessUntilDoneAlreadyClosed(t *testing.T) {
	jobs := make(chan int, 1)
	jobs <- 1
	results := make(chan int, 1)

	var n int
	withTimeout(t, 3*time.Second, "ProcessUntilDone pre-closed", func() {
		n = ProcessUntilDone(closedDone(), jobs, results)
	})

	if n < 0 {
		t.Errorf("count = %d, want a non-negative number", n)
	}
}

// ---------- Stopper ----------

func TestStopperLifecycle(t *testing.T) {
	s := NewStopper(4)

	time.Sleep(30 * time.Millisecond)
	s.Stop()

	var counts []int
	withTimeout(t, 5*time.Second, "Stopper.Wait", func() {
		counts = s.Wait()
	})

	if len(counts) != 4 {
		t.Errorf("got %d counts, want 4", len(counts))
	}
}

func TestStopperStopIsIdempotent(t *testing.T) {
	s := NewStopper(3)

	// Closing a channel twice panics; Stop must guard with sync.Once.
	for i := 0; i < 5; i++ {
		s.Stop()
	}

	withTimeout(t, 5*time.Second, "Stopper.Wait after repeated Stop", func() {
		s.Wait()
	})
}

func TestStopperConcurrentStop(t *testing.T) {
	s := NewStopper(3)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Stop()
		}()
	}
	wg.Wait()

	withTimeout(t, 5*time.Second, "Stopper.Wait after concurrent Stop", func() {
		s.Wait()
	})
}

func TestStopperNoLeak(t *testing.T) {
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	before := runtime.NumGoroutine()

	for i := 0; i < 10; i++ {
		s := NewStopper(8)
		s.Stop()
		withTimeout(t, 5*time.Second, "Stopper.Wait", func() {
			s.Wait()
		})
	}

	runtime.GC()
	time.Sleep(300 * time.Millisecond)
	after := runtime.NumGoroutine()

	if after > before+5 {
		t.Errorf("goroutines went from %d to %d — Stopper workers are leaking", before, after)
	}
}

func TestStopperZeroWorkers(t *testing.T) {
	s := NewStopper(0)
	s.Stop()

	var counts []int
	withTimeout(t, 5*time.Second, "Stopper with zero workers", func() {
		counts = s.Wait()
	})

	if len(counts) != 1 {
		t.Errorf("got %d counts, want 1 (workers<=0 is treated as 1)", len(counts))
	}
}
