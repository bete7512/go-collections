package main

import (
	"math/rand"
	"testing"
)

func mustRing(t *testing.T, capacity int) *Ring {
	t.Helper()
	r, err := NewRing(capacity)
	if err != nil {
		t.Fatalf("NewRing(%d) returned error: %v", capacity, err)
	}
	if r == nil {
		t.Fatalf("NewRing(%d) returned nil ring with nil error", capacity)
	}
	return r
}

func TestNewRingRejectsBadCapacity(t *testing.T) {
	for _, capacity := range []int{0, -1, -100} {
		if _, err := NewRing(capacity); err == nil {
			t.Errorf("NewRing(%d) = nil error, want an error", capacity)
		}
	}
}

func TestCapAndInitialState(t *testing.T) {
	r := mustRing(t, 3)

	if got := r.Cap(); got != 3 {
		t.Errorf("Cap() = %d, want 3", got)
	}
	if got := r.Len(); got != 0 {
		t.Errorf("Len() of fresh ring = %d, want 0", got)
	}
	if v, ok := r.Read(); ok || v != 0 {
		t.Errorf("Read on fresh ring = (%d, %v), want (0, false)", v, ok)
	}
}

func TestScriptedSequence(t *testing.T) {
	// The pinned sequence from the README, step by step.
	r := mustRing(t, 3)

	if !r.Write(1) || !r.Write(2) || !r.Write(3) {
		t.Fatalf("writes into a non-full ring returned false")
	}
	if got := r.Len(); got != 3 {
		t.Fatalf("Len after filling = %d, want 3", got)
	}

	if r.Write(4) {
		t.Fatalf("Write(4) on a full ring = true, want false")
	}
	if got := r.Len(); got != 3 {
		t.Fatalf("Len after rejected write = %d, want 3 (nothing stored)", got)
	}

	if v, ok := r.Read(); !ok || v != 1 {
		t.Fatalf("Read = (%d, %v), want (1, true)", v, ok)
	}

	if !r.Write(4) {
		t.Fatalf("Write(4) after freeing a slot = false, want true")
	}

	for _, want := range []int{2, 3, 4} {
		v, ok := r.Read()
		if !ok || v != want {
			t.Fatalf("Read = (%d, %v), want (%d, true) — FIFO through the wrap", v, ok, want)
		}
	}

	if v, ok := r.Read(); ok || v != 0 {
		t.Errorf("Read on drained ring = (%d, %v), want (0, false)", v, ok)
	}
}

func TestRejectedWriteLosesNothing(t *testing.T) {
	r := mustRing(t, 2)
	r.Write(10)
	r.Write(20)

	for i := 0; i < 5; i++ {
		if r.Write(99) {
			t.Fatalf("Write on full ring #%d = true, want false", i)
		}
	}

	if v, _ := r.Read(); v != 10 {
		t.Errorf("first Read = %d, want 10 — rejected writes must not disturb contents", v)
	}
	if v, _ := r.Read(); v != 20 {
		t.Errorf("second Read = %d, want 20", v)
	}
}

func TestCapacityOne(t *testing.T) {
	r := mustRing(t, 1)

	if !r.Write(7) {
		t.Fatalf("Write into empty cap-1 ring = false")
	}
	if r.Write(8) {
		t.Fatalf("Write into full cap-1 ring = true")
	}
	if v, ok := r.Read(); !ok || v != 7 {
		t.Fatalf("Read = (%d, %v), want (7, true)", v, ok)
	}
	if !r.Write(8) {
		t.Fatalf("Write after drain = false — slot must be reusable")
	}
	if v, ok := r.Read(); !ok || v != 8 {
		t.Fatalf("Read = (%d, %v), want (8, true)", v, ok)
	}
}

func TestFillDrainCycles(t *testing.T) {
	r := mustRing(t, 3)

	for cycle := 0; cycle < 10; cycle++ {
		base := cycle * 100
		for i := 0; i < 3; i++ {
			if !r.Write(base + i) {
				t.Fatalf("cycle %d: Write(%d) = false", cycle, base+i)
			}
		}
		if got := r.Len(); got != 3 {
			t.Fatalf("cycle %d: Len = %d, want 3", cycle, got)
		}
		for i := 0; i < 3; i++ {
			v, ok := r.Read()
			if !ok || v != base+i {
				t.Fatalf("cycle %d: Read = (%d, %v), want (%d, true)", cycle, v, ok, base+i)
			}
		}
		if got := r.Len(); got != 0 {
			t.Fatalf("cycle %d: Len after drain = %d, want 0", cycle, got)
		}
	}
}

func TestZeroAndNegativeValues(t *testing.T) {
	r := mustRing(t, 3)
	r.Write(0)
	r.Write(-1)

	if v, ok := r.Read(); !ok || v != 0 {
		t.Errorf("Read = (%d, %v), want (0, true) — a real stored zero", v, ok)
	}
	if v, ok := r.Read(); !ok || v != -1 {
		t.Errorf("Read = (%d, %v), want (-1, true)", v, ok)
	}
}

func TestWrapStress(t *testing.T) {
	// 1000 lockstep write/read rounds on capacity 4: the indices lap the
	// backing array 250 times. Any modulo slip fails within a few laps.
	r := mustRing(t, 4)

	for i := 0; i < 1000; i++ {
		if !r.Write(i) {
			t.Fatalf("round %d: Write = false on non-full ring", i)
		}
		v, ok := r.Read()
		if !ok || v != i {
			t.Fatalf("round %d: Read = (%d, %v), want (%d, true)", i, v, ok, i)
		}
	}
	if got := r.Len(); got != 0 {
		t.Errorf("Len after lockstep rounds = %d, want 0", got)
	}
}

func TestOracleRandomOps(t *testing.T) {
	// 5000 random ops mirrored against a trivially correct model queue.
	// Every accept/reject and every value must match the model exactly.
	const capacity = 4
	r := mustRing(t, capacity)
	var model []int

	rng := rand.New(rand.NewSource(50))
	for op := 0; op < 5000; op++ {
		if rng.Intn(2) == 0 {
			v := rng.Intn(1000) - 500
			gotOK := r.Write(v)
			wantOK := len(model) < capacity
			if gotOK != wantOK {
				t.Fatalf("op %d: Write(%d) = %v, want %v (model holds %d/%d)",
					op, v, gotOK, wantOK, len(model), capacity)
			}
			if wantOK {
				model = append(model, v)
			}
		} else {
			gotV, gotOK := r.Read()
			if len(model) == 0 {
				if gotOK {
					t.Fatalf("op %d: Read on empty = (%d, true), want (_, false)", op, gotV)
				}
			} else {
				wantV := model[0]
				model = model[1:]
				if !gotOK || gotV != wantV {
					t.Fatalf("op %d: Read = (%d, %v), want (%d, true)", op, gotV, gotOK, wantV)
				}
			}
		}
		if r.Len() != len(model) {
			t.Fatalf("op %d: Len = %d, model holds %d", op, r.Len(), len(model))
		}
	}
}
