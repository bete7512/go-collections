package main

import "math"

func main() {}

type Point struct{ X, Y float64 }

func (p Point) Distance(q Point) float64 {

	//TODO: it needs me to read maths again to do it without Hypot built in function.
	// x2 := (q.X - p.X) * (q.X - p.X)
	// y2 := (q.Y - p.Y) * (q.Y - p.Y)
	// return math.Sqrt(x2 + y2)
	return math.Hypot(q.X-p.X, q.Y-p.Y)

}
