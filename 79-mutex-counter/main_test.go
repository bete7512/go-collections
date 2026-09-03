package main

import (
	"os"
	"sync"
	"testing"
)

const (
	hammerGoroutines = 100
	hammerIncrements = 100
	hammerTotal      = hammerGoroutines * hammerIncrements
)

// ---------- UnsafeCounter ----------

func TestUnsafeCounterSequential(t *testing.T) {
	var c UnsafeCounter
	if got := c.Value(); got != 0 {
		t.Fatalf("zero value: Value() = %d, want 0", got)
	}
	for i := 0; i < 5; i++ {
		c.Inc()
	}
	if got := c.Value(); got != 5 {
		t.Errorf("after 5 Incs: Value() = %d, want 5 — sequential use has no race", got)
	}
}

func TestUnsafeCounterRaceDemonstration(t *testing.T) {
	// Deliberate data race, skipped by default so the graded suite stays
	// green under -race. Opt in once, on purpose, and read the report:
	//
	//   RACE_DEMO=1 go test -race -run UnsafeCounterRaceDemonstration ./79-mutex-counter/
	//
	// Look for: the WARNING header, the two conflicting accesses with both
	// stack traces, and the "created at" goroutine creation sites.
	if os.Getenv("RACE_DEMO") == "" {
		t.Skip("set RACE_DEMO=1 (with -race) to run the deliberate data race and read the report")
	}

	var c UnsafeCounter
	var wg sync.WaitGroup
	for g := 0; g < hammerGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < hammerIncrements; i++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()

	t.Logf("UnsafeCounter after %d×%d increments: %d (10000 expected; anything less is lost updates)",
		hammerGoroutines, hammerIncrements, c.Value())
}

// ---------- SafeCounter ----------

func TestSafeCounterZeroValueReady(t *testing.T) {
	var c SafeCounter // no constructor — the zero value must work
	if got := c.Value(); got != 0 {
		t.Fatalf("zero value: Value() = %d, want 0", got)
	}
	c.Inc()
	if got := c.Value(); got != 1 {
		t.Errorf("after one Inc: Value() = %d, want 1", got)
	}
}

func TestSafeCounterSequential(t *testing.T) {
	c := new(SafeCounter)
	for i := 1; i <= 250; i++ {
		c.Inc()
		if got := c.Value(); got != i {
			t.Fatalf("after %d Incs: Value() = %d", i, got)
		}
	}
}

func TestSafeCounterConcurrentIncrements(t *testing.T) {
	// 100 goroutines × 100 increments must produce exactly 10000 —
	// every single run. Run the whole hammer three times.
	for run := 0; run < 3; run++ {
		var c SafeCounter
		var wg sync.WaitGroup
		for g := 0; g < hammerGoroutines; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < hammerIncrements; i++ {
					c.Inc()
				}
			}()
		}
		wg.Wait()

		if got := c.Value(); got != hammerTotal {
			t.Errorf("run %d: Value() = %d, want %d — increments were lost", run, got, hammerTotal)
		}
	}
}

func TestSafeCounterConcurrentReadWrite(t *testing.T) {
	// Readers hammering Value() while writers hammer Inc(). This is the
	// test that catches a Value() without the lock: the read is a data
	// race even though it modifies nothing, and -race will say so.
	var c SafeCounter
	var wg sync.WaitGroup

	for g := 0; g < 50; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < hammerIncrements; i++ {
				c.Inc()
			}
		}()
	}

	for g := 0; g < 50; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < hammerIncrements; i++ {
				v := c.Value()
				if v < 0 || v > 50*hammerIncrements {
					t.Errorf("Value() = %d, out of range [0, %d]", v, 50*hammerIncrements)
					return
				}
			}
		}()
	}

	wg.Wait()
	if got := c.Value(); got != 50*hammerIncrements {
		t.Errorf("final Value() = %d, want %d", got, 50*hammerIncrements)
	}
}

// ---------- AtomicCounter ----------

func TestAtomicCounterZeroValueReady(t *testing.T) {
	var c AtomicCounter
	if got := c.Value(); got != 0 {
		t.Fatalf("zero value: Value() = %d, want 0", got)
	}
	c.Inc()
	if got := c.Value(); got != 1 {
		t.Errorf("after one Inc: Value() = %d, want 1", got)
	}
}

func TestAtomicCounterConcurrentIncrements(t *testing.T) {
	var c AtomicCounter
	var wg sync.WaitGroup
	for g := 0; g < hammerGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < hammerIncrements; i++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()

	if got := c.Value(); got != int64(hammerTotal) {
		t.Errorf("Value() = %d, want %d", got, hammerTotal)
	}
}

// ---------- benchmarks: mutex vs atomic ----------

func BenchmarkSafeCounterInc(b *testing.B) {
	var c SafeCounter
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
}

func BenchmarkAtomicCounterInc(b *testing.B) {
	var c AtomicCounter
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
}
