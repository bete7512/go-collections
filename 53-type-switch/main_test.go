package main

import (
	"fmt"
	"testing"
)

var (
	_ Shape = Circle{}
	_ Shape = Square{}
)

func TestDescribeConcreteTypes(t *testing.T) {
	tests := []struct {
		name     string
		shape    Shape
		expected string
	}{
		{"unit circle", Circle{R: 1}, "circle with radius 1"},
		{"fractional circle", Circle{R: 2.5}, "circle with radius 2.5"},
		{"zero circle", Circle{}, "circle with radius 0"},
		{"negative circle", Circle{R: -2}, "circle with radius -2"},
		{"unit square", Square{Side: 1}, "square with side 1"},
		{"square side 2", Square{Side: 2}, "square with side 2"},
		{"fractional square", Square{Side: 0.5}, "square with side 0.5"},
		{"zero square", Square{}, "square with side 0"},
		{"negative square", Square{Side: -3}, "square with side -3"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Describe(tc.shape); got != tc.expected {
				t.Errorf("Describe(%#v) = %q, want %q", tc.shape, got, tc.expected)
			}
		})
	}
}

func TestDescribeNil(t *testing.T) {
	if got := Describe(nil); got != "no shape" {
		t.Errorf("Describe(nil) = %q, want %q", got, "no shape")
	}

	var s Shape // never assigned
	if got := Describe(s); got != "no shape" {
		t.Errorf("Describe(unassigned Shape) = %q, want %q", got, "no shape")
	}
}

func TestDescribePointersFallToDefault(t *testing.T) {
	// A type switch matches the DYNAMIC type exactly: *Circle is not Circle.
	got := Describe(&Circle{R: 1})
	want := fmt.Sprintf("unknown shape with area %v", (&Circle{R: 1}).Area())
	if got != want {
		t.Errorf("Describe(&Circle{1}) = %q, want %q (*Circle must not match case Circle)", got, want)
	}

	got = Describe(&Square{Side: 2})
	want = "unknown shape with area 4"
	if got != want {
		t.Errorf("Describe(&Square{2}) = %q, want %q", got, want)
	}
}

// Types Describe has never seen, declared only here.
type triangle struct{ b, h float64 }

func (tr triangle) Area() float64 { return tr.b * tr.h / 2 }

type unitShape struct{}

func (unitShape) Area() float64 { return 1 }

func TestDescribeForeignTypes(t *testing.T) {
	tests := []struct {
		name     string
		shape    Shape
		expected string
	}{
		{"triangle area 6", triangle{b: 4, h: 3}, "unknown shape with area 6"},
		{"triangle fractional", triangle{b: 3, h: 1}, "unknown shape with area 1.5"},
		{"unit shape", unitShape{}, "unknown shape with area 1"},
		{"zero-area triangle", triangle{}, "unknown shape with area 0"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Describe(tc.shape); got != tc.expected {
				t.Errorf("Describe(%#v) = %q, want %q", tc.shape, got, tc.expected)
			}
		})
	}
}

func TestDescribeOverMixedSlice(t *testing.T) {
	shapes := []Shape{
		Circle{R: 3},
		Square{Side: 4},
		nil,
		triangle{b: 2, h: 2},
	}
	expected := []string{
		"circle with radius 3",
		"square with side 4",
		"no shape",
		"unknown shape with area 2",
	}

	for i, s := range shapes {
		if got := Describe(s); got != expected[i] {
			t.Errorf("shapes[%d]: Describe = %q, want %q", i, got, expected[i])
		}
	}
}
