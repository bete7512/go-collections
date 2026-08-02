package main

import (
	"fmt"
	"strings"
)

func main() {}

type Temp struct {
	C float64
}

func (t Temp) String() string {
	s := fmt.Sprintf("%.2f", t.C)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s + "°C"
}
