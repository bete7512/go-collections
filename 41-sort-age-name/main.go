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

func SortByAgeThenName(people []Person) {
	slices.SortFunc(people, func(i, j Person) int {
		return cmp.Or(
			cmp.Compare(i.Age, j.Age),
			cmp.Compare(i.Name, j.Name),
		)
	})
}
