package main

import (
	"strings"
)

func main() {}
func DedupChars(s string) string {
	occurrences := make(map[rune]int)
	runes := []rune(s)
	var collector strings.Builder
	for _, val := range runes {
		if _, ok := occurrences[val]; !ok {
			occurrences[val]++
			collector.WriteString(string(val))
		}
	}

	return collector.String()

}
