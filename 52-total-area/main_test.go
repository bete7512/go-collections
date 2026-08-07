package main

import (
	"math"
	"slices"
	"testing"
)

var (
	_ Shape = Circle{}
	_ Shape = Square{}
)

func approxEqual(got, want float64) bool {
	if got == want {
		return true
	}
	scale := math.Max(1, math.Max(math.Abs(got), math.Abs(want)))
	return math.Abs(got-want) <= 1e-12*scale
}

func TestTotalArea(t *testing.T) {
	tests := []struct {
		name     string
		shapes   []Shape
		expected float64
	}{
		{
			name:     "nil slice",
			shapes:   nil,
			expected: 0,
		},
		{
			name:     "empty slice",
			shapes:   []Shape{},
			expected: 0,
		},
		{
			name:     "single circle",
			shapes:   []Shape{Circle{R: 1}},
			expected: math.Pi,
		},
		{
			name:     "single square",
			shapes:   []Shape{Square{Side: 3}},
			expected: 9,
		},
		{
			name:     "mixed types",
			shapes:   []Shape{Circle{R: 1}, Square{Side: 2}},
			expected: math.Pi + 4,
		},
		{
			name:     "many mixed",
			shapes:   []Shape{Circle{R: 1}, Square{Side: 2}, Circle{R: 2}, Square{Side: 3}},
			expected: math.Pi + 4 + 4*math.Pi + 9,
		},
		{
			name:     "all zero area",
			shapes:   []Shape{Circle{}, Square{}, Circle{R: 0}},
			expected: 0,
		},
		{
			name:     "negative dimensions still positive areas",
			shapes:   []Shape{Circle{R: -2}, Square{Side: -3}},
			expected: 4*math.Pi + 9,
		},
		{
			name:     "fractional dimensions",
			shapes:   []Shape{Square{Side: 0.5}, Square{Side: 1.5}},
			expected: 0.25 + 2.25,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := slices.Clone(tc.shapes)

			got := TotalArea(tc.shapes)

			if !approxEqual(got, tc.expected) {
				t.Errorf("TotalArea(%v) = %v, want %v", tc.shapes, got, tc.expected)
			}
			if !slices.Equal(tc.shapes, snapshot) {
				t.Errorf("input slice was modified")
			}
		})
	}
}

func TestNilElementsAreSkipped(t *testing.T) {
	shapes := []Shape{Circle{R: 1}, nil, Square{Side: 2}, nil}

	got := TotalArea(shapes)

	want := math.Pi + 4
	if !approxEqual(got, want) {
		t.Errorf("TotalArea with nil elements = %v, want %v (nil entries must be skipped, not panic)", got, want)
	}
}

func TestAllNilElements(t *testing.T) {
	shapes := []Shape{nil, nil, nil}
	if got := TotalArea(shapes); !approxEqual(got, 0) {
		t.Errorf("TotalArea(all nils) = %v, want 0", got)
	}
}

func TestPointersAndValuesMixed(t *testing.T) {
	shapes := []Shape{Circle{R: 1}, &Circle{R: 2}, Square{Side: 2}, &Square{Side: 3}}

	got := TotalArea(shapes)

	want := math.Pi + 4*math.Pi + 4 + 9
	if !approxEqual(got, want) {
		t.Errorf("TotalArea(mixed values and pointers) = %v, want %v", got, want)
	}
}

// Two types TotalArea has never seen, declared only in this test file.
type hexagon struct{ side float64 }

func (h hexagon) Area() float64 { return 3 * math.Sqrt(3) / 2 * h.side * h.side }

type unitShape struct{}

func (unitShape) Area() float64 { return 1 }

func TestForeignTypesSumCorrectly(t *testing.T) {
	hexArea := 3 * math.Sqrt(3) / 2 * 4 // side 2

	shapes := []Shape{
		Circle{R: 1},
		hexagon{side: 2},
		unitShape{},
		Square{Side: 2},
	}

	got := TotalArea(shapes)

	want := math.Pi + hexArea + 1 + 4
	if !approxEqual(got, want) {
		t.Errorf("TotalArea with foreign types = %v, want %v — TotalArea must not know concrete types", got, want)
	}
}

func TestLargeSlice(t *testing.T) {
	// 500 unit circles + 500 unit squares.
	shapes := make([]Shape, 0, 1000)
	for i := 0; i < 500; i++ {
		shapes = append(shapes, Circle{R: 1}, Square{Side: 1})
	}

	got := TotalArea(shapes)

	want := 500*math.Pi + 500
	if !approxEqual(got, want) {
		t.Errorf("TotalArea(1000 shapes) = %v, want %v", got, want)
	}
}
