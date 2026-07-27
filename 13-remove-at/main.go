package main

import "errors"

func main() {}
func RemoveAt(s []int, i int) ([]int, error) {
	if i < 0 || i >= len(s) {
		return nil, errors.New("invalid index")
	}
	return append(s[:i], s[i+1:]...), nil
}
