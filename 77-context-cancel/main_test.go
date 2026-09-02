package main

import (
	"context"
	"errors"
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
		t.Fatalf("%s did not finish within %v — every blocking operation must race "+
			"against <-ctx.Done() inside a select", what, d)
	}
}

func canceledCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// ---------- CountUntilCancel ----------

func TestCountUntilCancelStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan int, 1)
	go func() { result <- CountUntilCancel(ctx) }()

	time.Sleep(30 * time.Millisecond) // let it spin

	if err := ctx.Err(); err != nil {
		t.Fatalf("ctx.Err() = %v before cancel, want nil", err)
	}
	cancel()
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("ctx.Err() = %v after cancel, want context.Canceled", ctx.Err())
	}

	select {
	case n := <-result:
		if n <= 0 {
			t.Errorf("count = %d, want a positive number of iterations", n)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("CountUntilCancel did not return after cancel()")
	}
}

func TestCountUntilCancelAlreadyCanceled(t *testing.T) {
	withTimeout(t, 3*time.Second, "CountUntilCancel with a pre-canceled context", func() {
		CountUntilCancel(canceledCtx())
	})
}

func TestCountUntilCancelDoubleCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // third call by the end of this test — all harmless

	result := make(chan int, 1)
	go func() { result <- CountUntilCancel(ctx) }()

	cancel()
	cancel() // idempotent: no panic, unlike close(done) twice

	select {
	case <-result:
	case <-time.After(3 * time.Second):
		t.Fatalf("CountUntilCancel did not return after cancel()")
	}
}

func TestCountUntilCancelRepeatedCycles(t *testing.T) {
	for run := 0; run < 50; run++ {
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan int, 1)

		go func() { result <- CountUntilCancel(ctx) }()
		cancel()

		select {
		case <-result:
		case <-time.After(3 * time.Second):
			t.Fatalf("run %d: CountUntilCancel did not return", run)
		}
	}
}

// ---------- RunUntilCancel ----------

func TestRunUntilCancelReleasesAllWorkers(t *testing.T) {
	for _, workers := range []int{1, 5, 50} {
		ctx, cancel := context.WithCancel(context.Background())

		result := make(chan int, 1)
		go func() { result <- RunUntilCancel(ctx, workers) }()

		time.Sleep(30 * time.Millisecond)
		cancel() // ONE cancel releases every worker

		select {
		case total := <-result:
			if total <= 0 {
				t.Errorf("workers=%d: total = %d, want a positive amount of work", workers, total)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("workers=%d: RunUntilCancel did not return after one cancel()", workers)
		}
		cancel()
	}
}

func TestRunUntilCancelEdgeCases(t *testing.T) {
	t.Run("zero workers treated as one", func(t *testing.T) {
		var total int
		withTimeout(t, 3*time.Second, "RunUntilCancel(ctx, 0)", func() {
			total = RunUntilCancel(canceledCtx(), 0)
		})
		if total < 0 {
			t.Errorf("total = %d, want a non-negative number", total)
		}
	})

	t.Run("negative workers treated as one", func(t *testing.T) {
		var total int
		withTimeout(t, 3*time.Second, "RunUntilCancel(ctx, -3)", func() {
			total = RunUntilCancel(canceledCtx(), -3)
		})
		if total < 0 {
			t.Errorf("total = %d, want a non-negative number", total)
		}
	})

	t.Run("already canceled returns promptly", func(t *testing.T) {
		withTimeout(t, 3*time.Second, "RunUntilCancel pre-canceled", func() {
			RunUntilCancel(canceledCtx(), 4)
		})
	})
}

func TestRunUntilCancelParentCancelsChild(t *testing.T) {
	// The cancellation tree: workers watch a CHILD context; the test
	// cancels only the PARENT. Cancellation must propagate down.
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	child, cancelChild := context.WithCancel(parent)
	defer cancelChild()

	result := make(chan int, 1)
	go func() { result <- RunUntilCancel(child, 5) }()

	time.Sleep(30 * time.Millisecond)
	cancelParent() // never touch cancelChild before the assertion

	select {
	case <-result:
	case <-time.After(5 * time.Second):
		t.Fatalf("RunUntilCancel did not return — cancelling the parent must cancel the derived child context")
	}

	if !errors.Is(child.Err(), context.Canceled) {
		t.Errorf("child.Err() = %v after parent cancel, want context.Canceled", child.Err())
	}
}

func TestRunUntilCancelNoLeak(t *testing.T) {
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	before := runtime.NumGoroutine()

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan int, 1)
		go func() { result <- RunUntilCancel(ctx, 10) }()
		cancel()

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
		t.Errorf("goroutines went from %d to %d — workers are leaking after cancel", before, after)
	}
}

// ---------- ProcessUntilCancel ----------

func TestProcessUntilCancelStopsWithNoResultReader(t *testing.T) {
	// THE leak test: results is unbuffered and nobody receives from it.
	// A send placed outside the select blocks forever and ctx.Done()
	// cannot rescue it.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan int)
	results := make(chan int) // no reader, ever

	go func() {
		for i := 0; ; i++ {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	type outcome struct {
		n   int
		err error
	}
	finished := make(chan outcome, 1)
	go func() {
		n, err := ProcessUntilCancel(ctx, jobs, results)
		finished <- outcome{n, err}
	}()

	time.Sleep(50 * time.Millisecond) // let it block on the results send
	cancel()

	select {
	case out := <-finished:
		if !errors.Is(out.err, context.Canceled) {
			t.Errorf("err = %v, want context.Canceled — the function must report why it stopped", out.err)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("ProcessUntilCancel did not return — the send into results must be a select case " +
			"racing against <-ctx.Done(), otherwise the goroutine is uncancellable")
	}
}

func TestProcessUntilCancelStopsWhenJobsClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan int, 4)
	for i := 1; i <= 4; i++ {
		jobs <- i
	}
	close(jobs)

	results := make(chan int, 4)

	var n int
	var err error
	withTimeout(t, 3*time.Second, "ProcessUntilCancel with closed jobs", func() {
		n, err = ProcessUntilCancel(ctx, jobs, results)
	})

	if err != nil {
		t.Errorf("err = %v, want nil — a normal jobs-channel drain is not an error", err)
	}
	if n != 4 {
		t.Errorf("processed %d jobs, want 4", n)
	}
}

func TestProcessUntilCancelProcessesJobs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan int, 3)
	jobs <- 1
	jobs <- 2
	jobs <- 3
	close(jobs)

	results := make(chan int, 3)

	var n int
	var err error
	withTimeout(t, 3*time.Second, "ProcessUntilCancel", func() {
		n, err = ProcessUntilCancel(ctx, jobs, results)
	})

	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if n != 3 {
		t.Fatalf("processed %d jobs, want 3", n)
	}

	close(results)
	got := map[int]bool{}
	for v := range results {
		got[v] = true
	}
	for _, want := range []int{2, 4, 6} {
		if !got[want] {
			t.Errorf("results missing %d — every job must be doubled and delivered", want)
		}
	}
	if len(got) != 3 {
		t.Errorf("got %d results, want 3", len(got))
	}
}

func TestProcessUntilCancelAlreadyCanceled(t *testing.T) {
	jobs := make(chan int) // open but empty: only Done() can fire
	results := make(chan int, 1)

	var n int
	var err error
	withTimeout(t, 3*time.Second, "ProcessUntilCancel pre-canceled", func() {
		n, err = ProcessUntilCancel(canceledCtx(), jobs, results)
	})

	if n != 0 {
		t.Errorf("count = %d, want 0 — no jobs were available", n)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

// ---------- CancelRunner ----------

func TestCancelRunnerLifecycle(t *testing.T) {
	r := NewCancelRunner(4)

	time.Sleep(30 * time.Millisecond)
	r.Stop()

	var counts []int
	withTimeout(t, 5*time.Second, "CancelRunner.Wait", func() {
		counts = r.Wait()
	})

	if len(counts) != 4 {
		t.Errorf("got %d counts, want 4", len(counts))
	}
}

func TestCancelRunnerStopIsIdempotent(t *testing.T) {
	r := NewCancelRunner(3)

	// cancel() is idempotent by contract — unlike #76's close(done),
	// no sync.Once is needed for this to be safe.
	for i := 0; i < 5; i++ {
		r.Stop()
	}

	withTimeout(t, 5*time.Second, "CancelRunner.Wait after repeated Stop", func() {
		r.Wait()
	})
}

func TestCancelRunnerConcurrentStop(t *testing.T) {
	r := NewCancelRunner(3)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Stop()
		}()
	}
	wg.Wait()

	withTimeout(t, 5*time.Second, "CancelRunner.Wait after concurrent Stop", func() {
		r.Wait()
	})
}

func TestCancelRunnerNoLeak(t *testing.T) {
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	before := runtime.NumGoroutine()

	for i := 0; i < 10; i++ {
		r := NewCancelRunner(8)
		r.Stop()
		withTimeout(t, 5*time.Second, "CancelRunner.Wait", func() {
			r.Wait()
		})
	}

	runtime.GC()
	time.Sleep(300 * time.Millisecond)
	after := runtime.NumGoroutine()

	if after > before+5 {
		t.Errorf("goroutines went from %d to %d — CancelRunner workers are leaking", before, after)
	}
}

func TestCancelRunnerZeroWorkers(t *testing.T) {
	r := NewCancelRunner(0)
	r.Stop()

	var counts []int
	withTimeout(t, 5*time.Second, "CancelRunner with zero workers", func() {
		counts = r.Wait()
	})

	if len(counts) != 1 {
		t.Errorf("got %d counts, want 1 (workers<=0 is treated as 1)", len(counts))
	}
}
