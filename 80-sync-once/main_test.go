package main

import (
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------- package-level Get / InitCalls ----------

func TestGetConcurrentBurst(t *testing.T) {
	const callers = 10

	ptrs := make([]*Config, callers)
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ptrs[i] = Get()
		}(i)
	}
	wg.Wait()

	if ptrs[0] == nil {
		t.Fatalf("Get() returned nil")
	}
	for i, p := range ptrs {
		if p != ptrs[0] {
			t.Errorf("goroutine %d got a different pointer (%p vs %p) — init ran more than once", i, p, ptrs[0])
		}
	}
	if got := InitCalls(); got != 1 {
		t.Errorf("InitCalls() = %d after concurrent burst, want exactly 1", got)
	}
}

func TestGetSequentialAfterBurst(t *testing.T) {
	first := Get()
	if first == nil {
		t.Fatalf("Get() returned nil")
	}
	for i := 0; i < 100; i++ {
		if p := Get(); p != first {
			t.Fatalf("call %d returned a different pointer", i)
		}
	}
	if got := InitCalls(); got != 1 {
		t.Errorf("InitCalls() = %d after repeated sequential calls, want exactly 1", got)
	}
}

// ---------- LazyConfig ----------

func TestLazyConfigLoadsExactlyOnce(t *testing.T) {
	for run := 0; run < 20; run++ {
		var calls atomic.Int32
		lc := NewLazyConfig(func() *Config {
			calls.Add(1)
			return &Config{}
		})

		const callers = 10
		ptrs := make([]*Config, callers)
		var wg sync.WaitGroup
		for i := 0; i < callers; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				ptrs[i] = lc.Get()
			}(i)
		}
		wg.Wait()

		if got := calls.Load(); got != 1 {
			t.Fatalf("run %d: loader ran %d times, want exactly 1", run, got)
		}
		if ptrs[0] == nil {
			t.Fatalf("run %d: Get() returned nil", run)
		}
		for i, p := range ptrs {
			if p != ptrs[0] {
				t.Fatalf("run %d: goroutine %d got a different pointer", run, i)
			}
		}
	}
}

func TestLazyConfigBlocksUntilInitFinishes(t *testing.T) {
	// Do is not fire-and-forget: while the first caller is inside the
	// loader, every other caller must WAIT — nobody may see nil.
	loaded := &Config{}
	lc := NewLazyConfig(func() *Config {
		time.Sleep(50 * time.Millisecond)
		return loaded
	})

	const callers = 20
	var sawNil atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if lc.Get() == nil {
				sawNil.Add(1)
			}
		}()
	}
	wg.Wait()

	if n := sawNil.Load(); n != 0 {
		t.Errorf("%d callers observed nil — Do must block them until the loader returns", n)
	}
	if lc.Get() != loaded {
		t.Errorf("Get() did not return the loader's value")
	}
}

func TestLazyConfigPanicMarksDone(t *testing.T) {
	var calls atomic.Int32
	lc := NewLazyConfig(func() *Config {
		calls.Add(1)
		panic("transient failure inside init")
	})

	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("first Get() did not panic — the loader's panic must propagate")
			}
		}()
		lc.Get()
	}()

	// The Once is spent: no retry, no second panic, just the half-state.
	var second *Config
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("second Get() panicked (%v) — a panic marks the Once done, it must not re-run", r)
			}
		}()
		second = lc.Get()
	}()

	if second != nil {
		t.Errorf("second Get() = %p, want nil — the loader never completed and must not retry", second)
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("loader ran %d times, want exactly 1 — initialization never retries after a panic", got)
	}
}

// ---------- SameOnceTwice ----------

func TestSameOnceTwice(t *testing.T) {
	var firstRan, secondRan bool
	SameOnceTwice(
		func() { firstRan = true },
		func() { secondRan = true },
	)

	if !firstRan {
		t.Errorf("first func did not run")
	}
	if secondRan {
		t.Errorf("second func ran — once means once per Once VALUE, not per function")
	}
}

// ---------- RacyGet ----------

func TestRacyGetSequential(t *testing.T) {
	first := RacyGet()
	if first == nil {
		t.Fatalf("RacyGet() returned nil")
	}
	if RacyGet() != first {
		t.Errorf("sequential RacyGet() returned a different pointer")
	}
}

func TestRacyGetRaceDemonstration(t *testing.T) {
	// The naive `if cfg == nil { cfg = load() }` idiom is a data race.
	// Skipped by default so the graded suite stays green; opt in with:
	//
	//   RACE_DEMO=1 go test -race -run RacyGetRaceDemonstration ./80-sync-once/
	//
	// and watch the detector point at the exact line sync.Once replaces.
	if os.Getenv("RACE_DEMO") == "" {
		t.Skip("set RACE_DEMO=1 (with -race) to run the deliberate data race and read the report")
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = RacyGet()
		}()
	}
	wg.Wait()
}
