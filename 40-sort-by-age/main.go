package main

import (
	"cmp"
	"slices"
)

func main() {}

type Person struct {
	Name string
	Age  int
}

func SortByAge(peoples []Person) {
	slices.SortFunc(peoples, func(a, b Person) int {
		return cmp.Compare(a.Age, b.Age)
	})
}
