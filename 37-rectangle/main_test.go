package main

import "testing"

func TestRectAreaAndPerimeter(t *testing.T) {
	tests := []struct {
		name      string
		rect      Rect
		area      float64
		perimeter float64
	}{
		{
			name:      "basic 3x4",
			rect:      Rect{3, 4},
			area:      12,
			perimeter: 14,
		},
		{
			name:      "zero value",
			rect:      Rect{},
			area:      0,
			perimeter: 0,
		},
		{
			name:      "zero width only",
			rect:      Rect{0, 7},
			area:      0,
			perimeter: 14,
		},
		{
			name:      "zero height only",
			rect:      Rect{7, 0},
			area:      0,
			perimeter: 14,
		},
		{
			name:      "square",
			rect:      Rect{5, 5},
			area:      25,
			perimeter: 20,
		},
		{
			name:      "fractional halves are exact",
			rect:      Rect{2.5, 4},
			area:      10,
			perimeter: 13,
		},
		{
			name:      "unit square",
			rect:      Rect{1, 1},
			area:      1,
			perimeter: 4,
		},
		{
			name:      "negative width raw formula",
			rect:      Rect{-3, 4},
			area:      -12,
			perimeter: 2,
		},
		{
			name:      "both negative raw formula",
			rect:      Rect{-3, -4},
			area:      12,
			perimeter: -14,
		},
		{
			name:      "large dimensions",
			rect:      Rect{1e6, 2e6},
			area:      2e12,
			perimeter: 6e6,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rect.Area(); got != tc.area {
				t.Errorf("Rect{%v, %v}.Area() = %v, want %v", tc.rect.W, tc.rect.H, got, tc.area)
			}
			if got := tc.rect.Perimeter(); got != tc.perimeter {
				t.Errorf("Rect{%v, %v}.Perimeter() = %v, want %v", tc.rect.W, tc.rect.H, got, tc.perimeter)
			}
		})
	}
}

func TestRectMethodsDoNotMutate(t *testing.T) {
	r := Rect{3, 4}

	for i := 0; i < 5; i++ {
		if got := r.Area(); got != 12 {
			t.Fatalf("call %d: Area() = %v, want 12 — receiver mutated?", i, got)
		}
		if got := r.Perimeter(); got != 14 {
			t.Fatalf("call %d: Perimeter() = %v, want 14 — receiver mutated?", i, got)
		}
	}

	if r.W != 3 || r.H != 4 {
		t.Errorf("fields changed after method calls: %+v, want {3 4}", r)
	}
}

func TestRectValueSemantics(t *testing.T) {
	r := Rect{6, 2}
	copied := r

	if r.Area() != copied.Area() || r.Perimeter() != copied.Perimeter() {
		t.Errorf("copy computes differently: original {%v %v}, copy {%v %v}",
			r.Area(), r.Perimeter(), copied.Area(), copied.Perimeter())
	}

	copied.W = 100
	if r.W != 6 {
		t.Errorf("mutating the copy changed the original: %+v", r)
	}
	if r.Area() != 12 {
		t.Errorf("original Area() after copy mutation = %v, want 12", r.Area())
	}

	// Value types are comparable — usable with == and as map keys.
	if r != (Rect{6, 2}) {
		t.Errorf("Rect{6,2} != Rect{6,2} — comparability broken")
	}
	seen := map[Rect]bool{r: true}
	if !seen[Rect{6, 2}] {
		t.Errorf("Rect not usable as a map key with value semantics")
	}
}
