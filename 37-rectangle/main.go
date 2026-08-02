package main

func main() {}

type Rect struct{ W, H float64 }

func (r Rect) Area() float64 {
	return r.H * r.W
} // W × H
func (r Rect) Perimeter() float64 {
	return 2 * (r.H + r.W)
} // 2 × (W + H)
