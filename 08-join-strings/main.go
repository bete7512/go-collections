package main

import "strings"

func main() {}
func Join(parts []string, sep string) string {
	var joined strings.Builder
	for i, part := range parts {
		joined.WriteString(part)
		if i != len(parts)-1 {
			joined.WriteString(sep)
		}
	}

	return joined.String()
}
