package main

import "log"

func main() {
	strings := []string{
		// Basic
		"",
		"a",
		"ab",
		"abc",
		"hello",
		"Hello, World!",

		// Whitespace
		" ",
		"  hello  ",
		"\thello\t",
		"hello\nworld",

		// Numbers & symbols
		"12345",
		"!@#$%^&*()",
		"a-b_c+d",

		// Accented Latin
		"héllo",
		"café",
		"naïve",
		"façade",
		"über",
		"résumé",

		// Non-Latin scripts
		"こんにちは",  // Japanese
		"안녕하세요",  // Korean
		"你好世界",   // Chinese
		"Привет", // Russian
		"مرحبا",  // Arabic
		"שלום",   // Hebrew
		"ሰላም",    // Amharic

		// Emoji
		"😀",
		"😀😃😄😁",
		"go👋",
		"👨‍💻",  // ZWJ sequence
		"👩🏽‍🚀", // skin tone + ZWJ
		"🏳️‍🌈", // rainbow flag
		"🇪🇹",   // flag (regional indicators)
		"❤️",   // heart with variation selector
		"👍🏽",

		// Mixed
		"Go语言",
		"Hello 世界",
		"Go👋世界🌍",
		"abc😀def",

		// Palindromes
		"racecar",
		"madam",
		"あいいあ",

		// Combining characters
		"e\u0301",    // e + combining acute
		"noe\u0308l", // noël

		// Long
		"abcdefghijklmnopqrstuvwxyz",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	for _, val := range strings {
		log.Printf("value:=> %s Reversed :=> %s", val, Reverse(val))
	}
}

func Reverse(s string) string {

	runes := []rune(s)
	i := 0
	j := len(runes) - 1
	for i < j {
		runes[i], runes[j] = runes[j], runes[i]
		i++
		j--
	}
	return string(runes)
}
