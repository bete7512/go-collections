package main

import (
	"slices"
	"strconv"
	"strings"
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
		t.Fatalf("%s did not finish within %v — a closed channel is always ready; "+
			"nil out its variable so the select case stops firing", what, d)
	}
}

// feed returns a channel carrying vals, closed when done.
func feed(vals ...int) <-chan int {
	ch := make(chan int, len(vals))
	for _, v := range vals {
		ch <- v
	}
	close(ch)
	return ch
}

func TestMerge2CollectsEverything(t *testing.T) {
	tests := []struct {
		name string
		a, b []int
	}{
		{"both have values", []int{1, 2, 3}, []int{10, 20}},
		{"a empty", nil, []int{10, 20}},
		{"b empty", []int{1, 2, 3}, nil},
		{"both empty", nil, nil},
		{"single each", []int{1}, []int{10}},
		{"very unequal", []int{1}, func() []int {
			v := make([]int, 1000)
			for i := range v {
				v[i] = 100 + i
			}
			return v
		}()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []int
			withTimeout(t, 3*time.Second, "Merge2", func() {
				got = Merge2(feed(tc.a...), feed(tc.b...))
			})

			if got == nil {
				t.Fatalf("Merge2 returned nil, want an empty non-nil slice")
			}

			want := append(slices.Clone(tc.a), tc.b...)
			gotSorted := slices.Clone(got)
			slices.Sort(gotSorted)
			slices.Sort(want)

			if !slices.Equal(gotSorted, want) {
				t.Errorf("Merge2 collected %d values, want %d (sorted got %v, want %v)",
					len(got), len(want), gotSorted, want)
			}
		})
	}
}

func TestMerge2WithNilChannels(t *testing.T) {
	// A nil channel never fires — treat it as already finished.
	var nilCh chan int

	t.Run("a nil", func(t *testing.T) {
		var got []int
		withTimeout(t, 3*time.Second, "Merge2 with nil a", func() {
			got = Merge2(nilCh, feed(1, 2))
		})
		slices.Sort(got)
		if !slices.Equal(got, []int{1, 2}) {
			t.Errorf("got %v, want [1 2]", got)
		}
	})

	t.Run("b nil", func(t *testing.T) {
		var got []int
		withTimeout(t, 3*time.Second, "Merge2 with nil b", func() {
			got = Merge2(feed(1, 2), nilCh)
		})
		slices.Sort(got)
		if !slices.Equal(got, []int{1, 2}) {
			t.Errorf("got %v, want [1 2]", got)
		}
	})

	t.Run("both nil", func(t *testing.T) {
		var got []int
		withTimeout(t, 3*time.Second, "Merge2 with both nil", func() {
			got = Merge2(nilCh, nilCh)
		})
		if len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})
}

func TestMerge2OneClosesLongBeforeTheOther(t *testing.T) {
	// `a` finishes immediately; `b` keeps producing. Without nilling `a`,
	// the loop spins on a's closed case and starves or never terminates.
	a := feed(1)

	b := make(chan int)
	go func() {
		defer close(b)
		for i := 0; i < 200; i++ {
			b <- 100 + i
		}
	}()

	var got []int
	withTimeout(t, 5*time.Second, "Merge2 with early close", func() {
		got = Merge2(a, b)
	})

	if len(got) != 201 {
		t.Errorf("collected %d values, want 201 — an early close must not drop or duplicate values", len(got))
	}
}

func TestMerge2RepeatedRuns(t *testing.T) {
	for run := 0; run < 50; run++ {
		var got []int
		withTimeout(t, 3*time.Second, "repeated Merge2", func() {
			got = Merge2(feed(1, 2, 3), feed(10, 20, 30))
		})

		slices.Sort(got)
		if !slices.Equal(got, []int{1, 2, 3, 10, 20, 30}) {
			t.Fatalf("run %d: got %v, want all six values exactly once", run, got)
		}
	}
}

func TestTaggedPreservesPerSourceOrder(t *testing.T) {
	var got []string
	withTimeout(t, 3*time.Second, "Tagged", func() {
		got = Tagged(feed(1, 2, 3, 4), feed(10, 20, 30))
	})

	if len(got) != 7 {
		t.Fatalf("got %d tagged values, want 7: %v", len(got), got)
	}

	// Extract each source's subsequence; interleaving is unspecified, but
	// within one source the order must hold.
	var aSeq, bSeq []int
	for _, tag := range got {
		src, num, found := strings.Cut(tag, ":")
		if !found {
			t.Fatalf("tag %q is not in the form \"a:1\"", tag)
		}
		v, err := strconv.Atoi(num)
		if err != nil {
			t.Fatalf("tag %q has a non-numeric value", tag)
		}
		switch src {
		case "a":
			aSeq = append(aSeq, v)
		case "b":
			bSeq = append(bSeq, v)
		default:
			t.Fatalf("tag %q has an unknown source", tag)
		}
	}

	if !slices.Equal(aSeq, []int{1, 2, 3, 4}) {
		t.Errorf("a's subsequence = %v, want [1 2 3 4] — per-source order must be preserved", aSeq)
	}
	if !slices.Equal(bSeq, []int{10, 20, 30}) {
		t.Errorf("b's subsequence = %v, want [10 20 30]", bSeq)
	}
}

func TestFirstReady(t *testing.T) {
	t.Run("only a ready", func(t *testing.T) {
		var v int
		var ok bool
		withTimeout(t, 3*time.Second, "FirstReady", func() {
			v, ok = FirstReady(feed(7), make(chan int))
		})
		if !ok || v != 7 {
			t.Errorf("FirstReady = (%d, %v), want (7, true)", v, ok)
		}
	})

	t.Run("only b ready", func(t *testing.T) {
		var v int
		var ok bool
		withTimeout(t, 3*time.Second, "FirstReady", func() {
			v, ok = FirstReady(make(chan int), feed(9))
		})
		if !ok || v != 9 {
			t.Errorf("FirstReady = (%d, %v), want (9, true)", v, ok)
		}
	})

	t.Run("both closed", func(t *testing.T) {
		var v int
		var ok bool
		withTimeout(t, 3*time.Second, "FirstReady on closed channels", func() {
			v, ok = FirstReady(feed(), feed())
		})
		if ok {
			t.Errorf("FirstReady on two closed channels = (%d, true), want (_, false)", v)
		}
	})
}

func TestSelectPicksRandomlyWhenBothReady(t *testing.T) {
	// With both cases always ready, select must choose uniformly at random.
	// An implementation that always prefers the first case fails here.
	fromA, fromB := 0, 0

	withTimeout(t, 10*time.Second, "randomness sampling", func() {
		for i := 0; i < 1000; i++ {
			a := feed(1)
			b := feed(2)
			v, ok := FirstReady(a, b)
			if !ok {
				t.Errorf("FirstReady reported not-ok with both channels ready")
				return
			}
			switch v {
			case 1:
				fromA++
			case 2:
				fromB++
			}
		}
	})

	if fromA == 0 || fromB == 0 {
		t.Errorf("over 1000 selections got a=%d b=%d — select must choose uniformly at random "+
			"when several cases are ready", fromA, fromB)
	}
}
