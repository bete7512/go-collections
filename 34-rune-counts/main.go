package main

import (
	"slices"
	"strconv"
)

func main() {}
func RuneCounts(s string) []string {
	counts := map[rune]int{}

	for _, val := range s {
		counts[val]++
	}

	strings := make([]string, 0, len(counts))
	for key, val := range counts {
		str := string(key) + ":" + strconv.Itoa(val)
		strings = append(strings, str)

	}

	slices.Sort(strings)

	return strings
}
