package main

import "testing"

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Empty / simple
		{"", true},
		{"a", true},
		{"aa", true},
		{"ab", false},

		// Words
		{"racecar", true},
		{"hello", false},
		{"level", true},

		// Phrases (exact character comparison)
		{"neveroddoreven", true},
		{"never odd or even", false}, // spaces included
		{"madam im adam", false},
		{"step on no pets", true},

		// Sentences
		{"was it a car or a cat i saw", false},
		{"hello world", false},
		{"go doog", false},

		// If punctuation is considered part of the string
		{"!!", true},
		{"!hello!", false},
		{"hello!", false},
		{"a,b,a", true},

		// Numbers
		{"12321", true},
		{"12345", false},
		{"1221", true},

		// Mixed
		{"abc123cba", false},
		{"abc123abc", false},
		{"a1b2b1a", true},

		// Unicode
		{"あいいあ", true},
		{"こんにちは", false},
		{"你好你", true},

		// Emoji
		{"😀😃😀", true},
		{"😀😃😄", false},
		{"hello😀olleh", true},
	}

	for _, tc := range tests {

		t.Run("palindrome testing", func(t *testing.T) {
			got := IsPalindrome(tc.input)
			if tc.expected != got {
				t.Errorf("IsPalindrome(%s)=%t, expected = %t", tc.input, got, tc.expected)
			}
		})
	}
}
