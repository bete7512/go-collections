package main

import "errors"

func main() {}
func Chunk(s []int, n int) ([][]int, error) {
	if n <= 0 {
		return nil, errors.New("invalid argument")
	}

	chunked := [][]int{}

	if len(s) == 0 {
		return chunked, nil
	}
	if n >= len(s) {
		return append(chunked, s), nil
	}

	divider := len(s) - (len(s) % n)
	for i := 0; i < divider; i = i + n {
		chunked = append(chunked, s[i:i+n])
	}

	if len(s[divider:]) > 0 {
		chunked = append(chunked, s[divider:])
	}

	return chunked, nil
}
