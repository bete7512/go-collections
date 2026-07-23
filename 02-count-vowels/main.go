package main

import (
	"fmt"
	"strings"
)

func main() {
	tests := []string{
		"",
		"hello",
		"Hello World",
		"aeiou",
		"AEIOU",
		"golang",
		"rhythm",
		"beautiful",
		"héllo",
		"こんにちは",
		"Go👋",
	}

	for _, value := range tests {
		fmt.Printf("value: %q vowels: %d\n", value, CountVowels(value))
	}
}

func CountVowels(s string) int {
	count := 0

	for _, r := range strings.ToLower(s) {
		switch r {
		case 'a', 'e', 'i', 'o', 'u':
			count++
		}
	}
	return count
}

// naive me creative one though
// func CountVowels(s string) int {
// 	s = strings.ToLower(s)
// 	vowelCount := 0
// 	vowels := map[rune]int{
// 		'a': 0,
// 		'e': 0,
// 		'i': 0,
// 		'o': 0,
// 		'u': 0,
// 	}

// 	for _, r := range s {
// 		if _, ok := vowels[r]; ok {
// 			vowels[r]++
// 		}
// 	}

// 	for _, v := range vowels {
// 		vowelCount += v
// 	}
// 	return vowelCount
// }
