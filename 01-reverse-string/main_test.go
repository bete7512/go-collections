package main

import "testing"

type Data struct {
	input    string
	expected string
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name string
		data []Data
	}{
		{
			name: "reverse string tests",
			data: []Data{
				// Basic
				{"", ""},
				{"a", "a"},
				{"ab", "ba"},
				{"abc", "cba"},
				{"hello", "olleh"},
				{"Hello, World!", "!dlroW ,olleH"},

				// Whitespace
				{" ", " "},
				{
					input:    "  hello  ",
					expected: "  olleh  ",
				},
				{
					input:    " hello   ",
					expected: "   olleh ",
				},
				{"hello\nworld", "dlrow\nolleh"},

				// Numbers & symbols
				{"12345", "54321"},
				{"!@#$%^&*()", ")(*&^%$#@!"},
				{"a-b_c+d", "d+c_b-a"},

				// Accented Latin
				{"héllo", "olléh"},
				{"café", "éfac"},
				{"naïve", "evïan"},
				{"façade", "edaçaf"},
				{"über", "rebü"},
				{"résumé", "émusér"},

				// Non-Latin scripts
				{"こんにちは", "はちにんこ"},
				{"안녕하세요", "요세하녕안"},
				{"你好世界", "界世好你"},
				{"Привет", "тевирП"},
				{"مرحبا", "ابحرم"},
				{"שלום", "םולש"},
				{"ሰላም", "ምላሰ"},

				// Emoji
				{"😀", "😀"},
				{"😀😃😄😁", "😁😄😃😀"},
				{"go👋", "👋og"},

				// Mixed
				{"Go语言", "言语oG"},
				{"Hello 世界", "界世 olleH"},
				{"Go👋世界🌍", "🌍界世👋oG"},
				{"abc😀def", "fed😀cba"},

				// Palindromes
				{"racecar", "racecar"},
				{"madam", "madam"},
				{"あいいあ", "あいいあ"},

				// Combining characters (reversed by runes)
				{"e\u0301", "\u0301e"},
				{"noe\u0308l", "l\u0308eon"},

				// Long
				{"abcdefghijklmnopqrstuvwxyz", "zyxwvutsrqponmlkjihgfedcba"},
				{"aaaaaaaaaaaaaaaaaaaaaaaaaaaa", "aaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, data := range tc.data {
				got := Reverse(data.input)
				if data.expected != got {
					t.Errorf("input=%s, expected=%s, got=%s", data.input, data.expected, got)
				}
			}
		})
	}
}
