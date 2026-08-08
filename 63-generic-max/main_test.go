package main

import (
	"errors"
	"math"
	"testing"
)

type Celsius float64
type Priority int

func TestMaxInts(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected int
	}{
		{"max at end", []int{3, 1, 4}, 4},
		{"max at start", []int{9, 1, 4}, 9},
		{"max in middle", []int{1, 42, 4}, 42},
		{"single element", []int{7}, 7},
		{"all equal", []int{5, 5, 5}, 5},
		{"duplicate maxima", []int{2, 9, 3, 9}, 9},
		{"all negative", []int{-5, -2, -10}, -2},
		{"negatives and zero", []int{-5, 0, -1}, 0},
		{"ascending", []int{1, 2, 3, 4, 5}, 5},
		{"descending", []int{5, 4, 3, 2, 1}, 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Max(tc.input)
			if err != nil {
				t.Fatalf("Max(%v) returned error: %v", tc.input, err)
			}
			if got != tc.expected {
				t.Errorf("Max(%v) = %d, want %d", tc.input, got, tc.expected)
			}
		})
	}
}

func TestMaxAllNegativeIsNotZero(t *testing.T) {
	// An implementation seeded with the zero value returns 0 here.
	got, err := Max([]int{-5, -2, -10, -1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != -1 {
		t.Errorf("Max(all negative) = %d, want -1 — seed the running max with s[0], not the zero value", got)
	}
}

func TestMaxFloats(t *testing.T) {
	got, err := Max([]float64{-2.5, -9.0, -3.25})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != -2.5 {
		t.Errorf("Max = %v, want -2.5", got)
	}

	withInf, err := Max([]float64{1, math.Inf(1), 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !math.IsInf(withInf, 1) {
		t.Errorf("Max with +Inf = %v, want +Inf", withInf)
	}

	negInf, err := Max([]float64{math.Inf(-1), -1e308})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if negInf != -1e308 {
		t.Errorf("Max with -Inf = %v, want -1e308", negInf)
	}
}

func TestMaxStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"lexicographic", []string{"apple", "pear", "fig"}, "pear"},
		{"uppercase sorts before lowercase", []string{"apple", "Zebra"}, "apple"},
		{"empty string in slice", []string{"", "a", ""}, "a"},
		{"all empty", []string{"", ""}, ""},
		{"unicode", []string{"a", "é", "世"}, "世"},
		{"length does not decide", []string{"zzz", "aaaaaaaa"}, "zzz"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Max(tc.input)
			if err != nil {
				t.Fatalf("Max(%v) returned error: %v", tc.input, err)
			}
			if got != tc.expected {
				t.Errorf("Max(%v) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestMaxOtherNumericKinds(t *testing.T) {
	if got, err := Max([]uint8{0, 255, 128}); err != nil || got != 255 {
		t.Errorf("Max(uint8) = (%d, %v), want (255, nil)", got, err)
	}

	big := int64(math.MaxInt64)
	if got, err := Max([]int64{1, big, -big}); err != nil || got != big {
		t.Errorf("Max(int64) = (%d, %v), want (%d, nil)", got, err, big)
	}

	if got, err := Max([]float32{1.5, 2.5}); err != nil || got != 2.5 {
		t.Errorf("Max(float32) = (%v, %v), want (2.5, nil)", got, err)
	}
}

func TestMaxNamedTypes(t *testing.T) {
	// Named types satisfy cmp.Ordered through their underlying type, and the
	// return keeps the named type.
	temps := []Celsius{21.5, 30.0, 18.2}
	hottest, err := Max(temps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var _ Celsius = hottest // compile-time: the result is a Celsius
	if hottest != Celsius(30.0) {
		t.Errorf("Max(temps) = %v, want 30", hottest)
	}

	priorities := []Priority{3, 1, 7, 2}
	top, err := Max(priorities)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var _ Priority = top
	if top != Priority(7) {
		t.Errorf("Max(priorities) = %v, want 7", top)
	}
}

func TestMaxEmptyAndNil(t *testing.T) {
	t.Run("empty ints", func(t *testing.T) {
		got, err := Max([]int{})
		if !errors.Is(err, ErrEmpty) {
			t.Fatalf("err = %v, want ErrEmpty", err)
		}
		if got != 0 {
			t.Errorf("value = %d, want the zero value 0", got)
		}
	})

	t.Run("nil ints", func(t *testing.T) {
		got, err := Max[int](nil)
		if !errors.Is(err, ErrEmpty) {
			t.Fatalf("err = %v, want ErrEmpty", err)
		}
		if got != 0 {
			t.Errorf("value = %d, want 0", got)
		}
	})

	t.Run("empty strings zero value", func(t *testing.T) {
		got, err := Max([]string{})
		if !errors.Is(err, ErrEmpty) {
			t.Fatalf("err = %v, want ErrEmpty", err)
		}
		if got != "" {
			t.Errorf("value = %q, want the zero value \"\"", got)
		}
	})

	t.Run("empty floats zero value", func(t *testing.T) {
		got, err := Max([]float64{})
		if !errors.Is(err, ErrEmpty) {
			t.Fatalf("err = %v, want ErrEmpty", err)
		}
		if got != 0.0 {
			t.Errorf("value = %v, want 0", got)
		}
	})

	t.Run("empty named type", func(t *testing.T) {
		got, err := Max([]Celsius{})
		if !errors.Is(err, ErrEmpty) {
			t.Fatalf("err = %v, want ErrEmpty", err)
		}
		if got != Celsius(0) {
			t.Errorf("value = %v, want 0", got)
		}
	})
}

func TestMaxNaNBehaviorIsDocumented(t *testing.T) {
	// Every comparison with NaN is false, so the result depends on position.
	// These assertions pin the behavior of a straightforward implementation
	// (strict > when updating, seeded from s[0]); they are documentation,
	// not a request to special-case NaN.
	nan := math.NaN()

	first, err := Max([]float64{nan, 1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !math.IsNaN(first) {
		t.Errorf("NaN at index 0: got %v, want NaN — nothing compares greater than NaN", first)
	}

	later, err := Max([]float64{1, nan, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if later != 2 {
		t.Errorf("NaN in the middle: got %v, want 2 — NaN loses every comparison and is skipped", later)
	}
}

func TestMaxLarge(t *testing.T) {
	const n = 10_000
	in := make([]int, n)
	for i := range in {
		in[i] = i % 100 // never exceeds 99...
	}
	in[n-1] = 1_000_000 // ...except the planted maximum at the very end

	got, err := Max(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1_000_000 {
		t.Errorf("Max = %d, want 1000000", got)
	}
}
