package main

import "testing"

func TestMemoizeReturnsCorrectValues(t *testing.T) {
	double := func(x int) int { return x * 2 }
	m := Memoize(double)

	tests := []struct {
		arg      int
		expected int
	}{
		{0, 0},
		{1, 2},
		{5, 10},
		{-3, -6},
		{1000, 2000},
	}

	for _, tc := range tests {
		if got := m(tc.arg); got != tc.expected {
			t.Errorf("memoized(%d) = %d, want %d", tc.arg, got, tc.expected)
		}
		// Second call must return the same value from cache.
		if got := m(tc.arg); got != tc.expected {
			t.Errorf("memoized(%d) second call = %d, want %d", tc.arg, got, tc.expected)
		}
	}
}

func TestMemoizeCallsUnderlyingOncePerArg(t *testing.T) {
	calls := 0
	slow := func(x int) int {
		calls++
		return x * x
	}
	m := Memoize(slow)

	m(7)
	m(7)
	m(7)
	if calls != 1 {
		t.Fatalf("after three calls with the same arg, underlying ran %d times, want 1", calls)
	}

	m(8)
	if calls != 2 {
		t.Fatalf("after a new arg, underlying ran %d times, want 2", calls)
	}

	m(7)
	m(8)
	if calls != 2 {
		t.Fatalf("after repeats of cached args, underlying ran %d times, want 2", calls)
	}
}

func TestMemoizeCachesZeroResults(t *testing.T) {
	calls := 0
	alwaysZero := func(x int) int {
		calls++
		return 0
	}
	m := Memoize(alwaysZero)

	if got := m(42); got != 0 {
		t.Fatalf("memoized(42) = %d, want 0", got)
	}
	if got := m(42); got != 0 {
		t.Fatalf("memoized(42) second call = %d, want 0", got)
	}
	if calls != 1 {
		t.Fatalf("zero result was not cached: underlying ran %d times, want 1 (comma-ok lookup missing?)", calls)
	}
}

func TestMemoizeZeroAndNegativeKeys(t *testing.T) {
	calls := 0
	inc := func(x int) int {
		calls++
		return x + 1
	}
	m := Memoize(inc)

	if got := m(0); got != 1 {
		t.Errorf("memoized(0) = %d, want 1", got)
	}
	if got := m(-5); got != -4 {
		t.Errorf("memoized(-5) = %d, want -4", got)
	}
	m(0)
	m(-5)
	if calls != 2 {
		t.Errorf("underlying ran %d times for keys {0, -5}, want 2", calls)
	}
}

func TestMemoizeIndependentCachesSameFunc(t *testing.T) {
	calls := 0
	f := func(x int) int {
		calls++
		return x * 10
	}

	a := Memoize(f)
	b := Memoize(f)

	a(1)
	if calls != 1 {
		t.Fatalf("after a(1): underlying ran %d times, want 1", calls)
	}
	b(1)
	if calls != 2 {
		t.Fatalf("b shares a's cache: after b(1) underlying ran %d times, want 2", calls)
	}
	a(1)
	b(1)
	if calls != 2 {
		t.Fatalf("caches not retained: underlying ran %d times, want 2", calls)
	}
}

func TestMemoizeNoCrossTalkBetweenFuncs(t *testing.T) {
	double := Memoize(func(x int) int { return x * 2 })
	triple := Memoize(func(x int) int { return x * 3 })

	if got := double(4); got != 8 {
		t.Errorf("double(4) = %d, want 8", got)
	}
	if got := triple(4); got != 12 {
		t.Errorf("triple(4) = %d, want 12 (cross-talk with double's cache?)", got)
	}
	if got := double(4); got != 8 {
		t.Errorf("double(4) after triple(4) = %d, want 8", got)
	}
}

func TestMemoizeInterleavedArgs(t *testing.T) {
	calls := 0
	m := Memoize(func(x int) int {
		calls++
		return -x
	})

	sequence := []int{1, 2, 1, 3, 2, 1, 3, 3, 2}
	for _, arg := range sequence {
		if got := m(arg); got != -arg {
			t.Fatalf("memoized(%d) = %d, want %d", arg, got, -arg)
		}
	}
	if calls != 3 {
		t.Errorf("3 distinct args in interleaved sequence, underlying ran %d times, want 3", calls)
	}
}

func TestMemoizeManyArgs(t *testing.T) {
	calls := 0
	m := Memoize(func(x int) int {
		calls++
		return x*x - x
	})

	for round := 0; round < 3; round++ {
		for arg := 0; arg < 100; arg++ {
			want := arg*arg - arg
			if got := m(arg); got != want {
				t.Fatalf("round %d: memoized(%d) = %d, want %d", round, arg, got, want)
			}
		}
	}
	if calls != 100 {
		t.Errorf("100 distinct args over 3 rounds: underlying ran %d times, want 100", calls)
	}
}
