package main

func main() {}
func GroupByFirst(words []string) map[rune][]string {
	mappedstrs := make(map[rune][]string)
	for _, word := range words {
		if word == "" {
			continue
		}
		firstChar := []rune(word)[0]
		mappedstrs[firstChar] = append(mappedstrs[firstChar], word)
	}
	return mappedstrs
}
