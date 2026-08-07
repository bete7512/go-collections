package main

import (
	"fmt"
	"math"
)

func main() {}

type Shape interface {
	Area() float64
}

type Circle struct {
	R float64
}
type Square struct {
	Side float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.R * c.R
} // π·R²
func (s Square) Area() float64 {
	return s.Side * s.Side
} // Side²

/*
| input | returns |
|---|---|
| `Circle{R: 1}` | `circle with radius 1` |
| `Circle{R: 2.5}` | `circle with radius 2.5` |
| `Square{Side: 2}` | `square with side 2` |
| `nil` | `no shape` |
| any other `Shape` | `unknown shape with area <A>` |
*/
func Describe(s Shape) string {
	shapeType := ""

	switch v := s.(type) {
	case Circle:
		shapeType = fmt.Sprintf("circle with radius %g", v.R)
	case Square:
		shapeType = fmt.Sprintf("square with side %g", v.Side)
	case nil:
		shapeType = "no shape"
	default:
		shapeType = fmt.Sprintf("unknown shape with area %g", v.Area())
	}

	return shapeType
}
