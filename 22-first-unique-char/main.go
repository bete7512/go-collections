package main

func main() {}
func FirstUnique(s string) (rune, bool) {
	counts := map[rune]int{}

	for _, val := range s {
		counts[val]++
	}

	for _, val := range s {
		if counts[val] == 1 {
			return val, true
		}
	}
	return 0, false
}
