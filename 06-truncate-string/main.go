package main

func main() {}
func Truncate(s string, n int) string {
	if n <= 0 {
		return "..."
	}
	runes := []rune(s)
	if len(runes) == 0 {
		return ""
	}
	if n >= len(runes) {
		return s
	}

	return string(runes[:n]) + "..."
}
