package main

import "log"

func main() {
	strs := []string{"racecar", "neveroddoreven"}
	for _, str := range strs {
		log.Printf("str: %s is palindrome %t", str, IsPalindrome(str))
	}
}
func IsPalindrome(s string) bool {
	runes := []rune(s)
	i := 0
	j := len(runes)-1
	for i <= j {
		if runes[i] != runes[j]{
			return false
		}
		i++
		j--
	}
	return true
}
