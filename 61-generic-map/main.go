package main

import (
	"log"
	"strconv"
)

func main() {
	log.Println(Map([]int{1, 2, 3}, strconv.Itoa))
}
func Map[T, U any](s []T, f func(T) U) []U {
	us := make([]U, len(s))
	for i, v := range s {
		us[i] = f(v)
	}

	return us
}
