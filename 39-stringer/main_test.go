package main

import (
	"fmt"
	"testing"
)

// Compile-time proof that Temp satisfies fmt.Stringer implicitly.
var _ fmt.Stringer = Temp{}

func TestStringFormat(t *testing.T) {
	tests := []struct {
		name     string
		temp     Temp
		expected string
	}{
		{
			name:     "fractional",
			temp:     Temp{21.5},
			expected: "21.5°C",
		},
		{
			name:     "whole number has no trailing decimal",
			temp:     Temp{21},
			expected: "21°C",
		},
		{
			name:     "zero",
			temp:     Temp{0},
			expected: "0°C",
		},
		{
			name:     "negative",
			temp:     Temp{-40},
			expected: "-40°C",
		},
		{
			name:     "absolute zero",
			temp:     Temp{-273.15},
			expected: "-273.15°C",
		},
		{
			name:     "body temperature",
			temp:     Temp{36.6},
			expected: "36.6°C",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.temp.String(); got != tc.expected {
				t.Errorf("Temp{%v}.String() = %q, want %q", tc.temp.C, got, tc.expected)
			}
		})
	}
}

func TestFmtUsesStringerAutomatically(t *testing.T) {
	temp := Temp{21.5}

	if got := fmt.Sprint(temp); got != "21.5°C" {
		t.Errorf("fmt.Sprint = %q, want %q — fmt should detect Stringer", got, "21.5°C")
	}
	if got := fmt.Sprintf("%v", temp); got != "21.5°C" {
		t.Errorf("%%v = %q, want %q", got, "21.5°C")
	}
	if got := fmt.Sprintf("%s", temp); got != "21.5°C" {
		t.Errorf("%%s = %q, want %q", got, "21.5°C")
	}
	if got := fmt.Sprintf("it is %v today", temp); got != "it is 21.5°C today" {
		t.Errorf("embedded %%v = %q, want %q", got, "it is 21.5°C today")
	}
}

func TestPointerAlsoFormats(t *testing.T) {
	temp := Temp{9.5}

	// With a VALUE receiver, *Temp's method set includes String() too.
	if got := fmt.Sprint(&temp); got != "9.5°C" {
		t.Errorf("fmt.Sprint(&temp) = %q, want %q — value-receiver String should cover pointers (got the default &{...} rendering?)",
			got, "9.5°C")
	}
}

func TestCompositesFormatElements(t *testing.T) {
	temps := []Temp{{1}, {2.5}, {-3}}

	if got := fmt.Sprint(temps); got != "[1°C 2.5°C -3°C]" {
		t.Errorf("fmt.Sprint(slice) = %q, want %q", got, "[1°C 2.5°C -3°C]")
	}

	m := map[string]Temp{"berlin": {18.5}}
	if got := fmt.Sprint(m); got != "map[berlin:18.5°C]" {
		t.Errorf("fmt.Sprint(map) = %q, want %q", got, "map[berlin:18.5°C]")
	}
}
