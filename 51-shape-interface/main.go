package main

import "math"

func main() {}

type Shape interface {
	Area() float64
}

type Circle struct{ R float64 }
type Square struct{ Side float64 }

func (c Circle) Area() float64 {
	return math.Pi * c.R * c.R
} // π·R²
func (s Square) Area() float64 {
	return s.Side * s.Side
} // Side²
