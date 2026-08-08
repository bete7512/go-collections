package main

import (
	"cmp"
	"errors"
)

func main() {}

var ErrEmpty = errors.New("max of empty slice")

func Max[T cmp.Ordered](s []T) (T, error) {
	var max T
	if len(s) == 0 {
		return max, ErrEmpty
	}
	max = s[0]

	for _, val := range s {
		if val > max {
			max = val
		}
	}

	return max, nil
}
