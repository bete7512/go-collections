package main

import "errors"

func main() {}
func RemoveAtFast(s []int, i int) ([]int, error) {
	if i < 0 || i >= len(s) {
		return nil, errors.New("invalid index")
	}

	s[i] = s[len(s)-1]

	return s[:len(s)-1], nil

}
