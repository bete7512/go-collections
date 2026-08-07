package main

import (
	"math"
	"testing"
)

// Compile-time proof that both concrete types satisfy Shape as VALUES.
// If Area used a pointer receiver, these lines would not compile.
var (
	_ Shape = Circle{}
	_ Shape = Square{}
	_ Shape = &Circle{}
	_ Shape = &Square{}
)

func approxEqual(got, want float64) bool {
	if got == want {
		return true
	}
	scale := math.Max(1, math.Max(math.Abs(got), math.Abs(want)))
	return math.Abs(got-want) <= 1e-12*scale
}

func TestCircleArea(t *testing.T) {
	tests := []struct {
		name     string
		circle   Circle
		expected float64
	}{
		{"unit circle", Circle{R: 1}, math.Pi},
		{"radius 2", Circle{R: 2}, 4 * math.Pi},
		{"zero radius", Circle{R: 0}, 0},
		{"zero value", Circle{}, 0},
		{"fractional radius", Circle{R: 0.5}, 0.25 * math.Pi},
		{"negative radius squares positive", Circle{R: -2}, 4 * math.Pi},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.circle.Area(); !approxEqual(got, tc.expected) {
				t.Errorf("Circle{%v}.Area() = %v, want %v", tc.circle.R, got, tc.expected)
			}
		})
	}
}

func TestSquareArea(t *testing.T) {
	tests := []struct {
		name     string
		square   Square
		expected float64
	}{
		{"unit square", Square{Side: 1}, 1},
		{"side 2", Square{Side: 2}, 4},
		{"zero side", Square{Side: 0}, 0},
		{"zero value", Square{}, 0},
		{"fractional side", Square{Side: 2.5}, 6.25},
		{"negative side squares positive", Square{Side: -3}, 9},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.square.Area(); !approxEqual(got, tc.expected) {
				t.Errorf("Square{%v}.Area() = %v, want %v", tc.square.Side, got, tc.expected)
			}
		})
	}
}

func TestOneVariableManyTypes(t *testing.T) {
	var s Shape

	s = Circle{R: 1}
	if got := s.Area(); !approxEqual(got, math.Pi) {
		t.Errorf("through Shape holding Circle: Area() = %v, want %v", got, math.Pi)
	}

	s = Square{Side: 2}
	if got := s.Area(); !approxEqual(got, 4) {
		t.Errorf("through Shape holding Square: Area() = %v, want 4", got)
	}
}

func TestPointersSatisfyShape(t *testing.T) {
	// A value receiver puts Area() in *Circle's method set too.
	c := Circle{R: 2}
	var s Shape = &c
	if got := s.Area(); !approxEqual(got, 4*math.Pi) {
		t.Errorf("through Shape holding *Circle: Area() = %v, want %v", got, 4*math.Pi)
	}

	sq := Square{Side: 3}
	s = &sq
	if got := s.Area(); !approxEqual(got, 9) {
		t.Errorf("through Shape holding *Square: Area() = %v, want 9", got)
	}
}

func TestHeterogeneousSlice(t *testing.T) {
	shapes := []Shape{
		Circle{R: 1},
		Square{Side: 2},
		Circle{R: 0},
		Square{Side: 0.5},
	}
	expected := []float64{math.Pi, 4, 0, 0.25}

	for i, s := range shapes {
		if got := s.Area(); !approxEqual(got, expected[i]) {
			t.Errorf("shapes[%d].Area() = %v, want %v", i, got, expected[i])
		}
	}
}

func TestNilShapeIsNil(t *testing.T) {
	var s Shape
	if s != nil {
		t.Errorf("zero-value Shape = %v, want nil", s)
	}
}

// triangle is defined HERE, in the test file, and never mentioned in main.go.
// It satisfies Shape purely by having the method — the open-set property of
// implicit interfaces.
type triangle struct {
	base, height float64
}

func (tr triangle) Area() float64 { return tr.base * tr.height / 2 }

func TestForeignTypeSatisfiesShape(t *testing.T) {
	var s Shape = triangle{base: 4, height: 3}
	if got := s.Area(); !approxEqual(got, 6) {
		t.Errorf("triangle.Area() through Shape = %v, want 6", got)
	}

	mixed := []Shape{Circle{R: 1}, triangle{base: 2, height: 2}, Square{Side: 1}}
	total := 0.0
	for _, sh := range mixed {
		total += sh.Area()
	}
	if !approxEqual(total, math.Pi+2+1) {
		t.Errorf("mixed slice total = %v, want %v", total, math.Pi+2+1)
	}
}
