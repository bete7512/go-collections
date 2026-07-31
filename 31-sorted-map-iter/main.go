package main

import (
	"slices"
	"strconv"
	"strings"
)

func main() {}

type keyValue struct {
	key   string
	value int
}

func SortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
func FormatSorted(m map[string]int) string {
	var b strings.Builder
	for _, k := range SortedKeys(m) {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(strconv.Itoa(m[k]))
		b.WriteString("\n")
	}
	return b.String()
}
