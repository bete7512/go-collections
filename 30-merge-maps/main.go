package main

import "maps"

func main() {}
func Merge(a, b map[string]int) map[string]int {

	copied := make(map[string]int, len(b))

	maps.Copy(copied, b)
	for key, val := range a {
		if _, ok := b[key]; !ok {
			copied[key] = val
		}
	}
	return copied
}
