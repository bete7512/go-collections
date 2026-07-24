package main

import (
	"log"
	"strings"
)

func main() {
	strs := []string{"In Go, the way you measure string length depends entirely on whether you want to count"}
	for _, str := range strs {
		log.Println("......", LongestWord(str))
	}
}
func LongestWord(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	longest := words[0]
	for _, word := range words {
		if len(word) > len(longest) {
			longest = word
		}
	}

	return longest
}
