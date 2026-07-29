package main

func main() {}
func IsAnagram(a, b string) bool {
	counts := map[rune]int{}

	for _, val := range a {
		counts[val]++
	}
	for _, val := range b {
		counts[val]--
	}

	for _, val := range counts {
		if val != 0 {
			return false
		}
	}

	return true

}

// naive me
// func IsAnagram(a, b string) bool {
// 	arunes := map[rune]int{}
// 	brunes := map[rune]int{}

// 	for _, val := range a {
// 		arunes[val]++
// 	}
// 	for _, val := range b {
// 		brunes[val]++
// 	}

// 	if maps.Equal(arunes, brunes) {
// 		return true
// 	}

// 	return false
// }
