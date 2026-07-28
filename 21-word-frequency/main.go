package main

import (
	"strings"
	"unicode"
)

func main() {}



func WordFreq(text string) map[string]int {
	text = strings.ToLower(text)

	// TODO: probably the appropriate way is using regex
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '-' && r != '\''
	})

	counts := make(map[string]int, len(words))

	for _, w := range words {
		w = strings.Trim(w, "-'")
		if w == "" {
			continue
		}
		counts[w]++
	}

	return counts
}
