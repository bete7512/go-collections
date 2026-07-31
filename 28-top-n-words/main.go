package main

import (
	"cmp"
	"slices"
	"strings"
)

func main() {}
func TopN(text string, n int) []string {
	text = strings.ToLower(text)
	counts := map[string]int{}
	words := strings.Fields(text)

	for _, word := range words {
		counts[word]++
	}

	type wordCount struct {
		word  string
		count int
	}

	countMaps := []wordCount{}
	for key, val := range counts {
		countMaps = append(countMaps, wordCount{word: key, count: val})
	}

	slices.SortFunc(countMaps, func(a, b wordCount) int {
		return cmp.Or(
			cmp.Compare(b.count, a.count),
			cmp.Compare(a.word, b.word),
		)
	})

	topNs := []string{}

	for i := 0; i < n && i < len(countMaps); i = i + 1 {
		topNs = append(topNs, countMaps[i].word)
	}

	return topNs
}
