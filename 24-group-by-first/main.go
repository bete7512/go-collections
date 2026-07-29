package main

func main() {}
func GroupByFirst(words []string) map[rune][]string {
	mappedstrs := make(map[rune][]string)
	for _, word := range words {
		if word == "" {
			continue
		}
		var firstChar rune

		for _, r := range word {
			firstChar = r
			break
		}
		mappedstrs[firstChar] = append(mappedstrs[firstChar], word)
	}
	return mappedstrs
}
