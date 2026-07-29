package main

import "maps"

func main() {}

func IsAnagram(a, b string) bool {
	arunes := map[rune]int{}
	brunes := map[rune]int{}

	for _, val := range a {
		arunes[val]++
	}
	for _, val := range b {
		brunes[val]++
	}

	if maps.Equal(arunes, brunes) {
		return true
	}

	return false
}
