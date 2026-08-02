package main

import (
	"math"
	"testing"
)

// approxEqual compares with a RELATIVE tolerance so it works at any magnitude
// (an absolute epsilon is meaningless next to values like 1e200).
func approxEqual(got, want float64) bool {
	if got == want {
		return true
	}
	diff := math.Abs(got - want)
	scale := math.Max(1, math.Max(math.Abs(got), math.Abs(want)))
	return diff <= 1e-12*scale
}

func TestDistance(t *testing.T) {
	tests := []struct {
		name     string
		p, q     Point
		expected float64
	}{
		{
			name:     "3-4-5 triangle from origin",
			p:        Point{0, 0},
			q:        Point{3, 4},
			expected: 5,
		},
		{
			name:     "same point is zero",
			p:        Point{2.5, -7},
			q:        Point{2.5, -7},
			expected: 0,
		},
		{
			name:     "origin to origin",
			p:        Point{0, 0},
			q:        Point{0, 0},
			expected: 0,
		},
		{
			name:     "negative coordinates shifted 3-4-5",
			p:        Point{-1, -1},
			q:        Point{2, 3},
			expected: 5,
		},
		{
			name:     "horizontal only",
			p:        Point{2, 3},
			q:        Point{7, 3},
			expected: 5,
		},
		{
			name:     "vertical only",
			p:        Point{1, 1},
			q:        Point{1, 5},
			expected: 4,
		},
		{
			name:     "irrational result sqrt 2",
			p:        Point{0, 0},
			q:        Point{1, 1},
			expected: math.Sqrt2,
		},
		{
			name:     "both points negative quadrant",
			p:        Point{-3, -4},
			q:        Point{-6, -8},
			expected: 5,
		},
		{
			name:     "fractional coordinates",
			p:        Point{0.5, 0.5},
			q:        Point{2, 2.5},
			expected: 2.5,
		},
		{
			name:     "huge coordinates must not overflow",
			p:        Point{1e200, 0},
			q:        Point{0, 1e200},
			expected: 1.4142135623730951e200, // sqrt(2) * 1e200
		},
		{
			name:     "tiny coordinates must not underflow",
			p:        Point{0, 0},
			q:        Point{3e-200, 4e-200},
			expected: 5e-200,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.p.Distance(tc.q)

			if math.IsInf(got, 0) || math.IsNaN(got) {
				t.Fatalf("Distance(%v, %v) = %v — overflow/underflow in the formula (math.Hypot avoids this)",
					tc.p, tc.q, got)
			}
			if !approxEqual(got, tc.expected) {
				t.Errorf("Distance(%v, %v) = %v, want %v", tc.p, tc.q, got, tc.expected)
			}
		})
	}
}

func TestDistanceSymmetry(t *testing.T) {
	pairs := []struct{ p, q Point }{
		{Point{0, 0}, Point{3, 4}},
		{Point{-5, 2}, Point{7, -9}},
		{Point{0.1, 0.2}, Point{0.3, 0.4}},
		{Point{1e200, -1e200}, Point{-1e200, 1e200}},
		{Point{0, 0}, Point{0, 0}},
	}

	for _, pair := range pairs {
		ab := pair.p.Distance(pair.q)
		ba := pair.q.Distance(pair.p)
		if ab != ba {
			t.Errorf("symmetry broken: %v.Distance(%v) = %v but %v.Distance(%v) = %v",
				pair.p, pair.q, ab, pair.q, pair.p, ba)
		}
		if ab < 0 {
			t.Errorf("distance must be non-negative, got %v", ab)
		}
	}
}
