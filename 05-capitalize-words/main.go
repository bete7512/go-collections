package main

import (
	"unicode"
)

func main() {}
func TitleCase(s string) string {
	runes := []rune(s)

	for i, char := range runes {
		if unicode.IsLetter(char) {
			if i == 0 {
				runes[i] = unicode.ToUpper(char)
				continue
			} else {
				if !unicode.IsLetter(runes[i-1]) && runes[i-1] != '\'' {
					runes[i] = unicode.ToUpper(char)
					continue
				}
			}
			if unicode.IsLetter(char) {
				if runes[i-1] == '\'' && char == 's' {
					continue
				}
				runes[i] = unicode.ToLower(char)
			}

		}
	}

	return string(runes)
}
