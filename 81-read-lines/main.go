package main

import (
	"bufio"
	"os"
)

func main() {}
func ReadLines(path string) ([]string, error) {
	fs, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}
	defer fs.Close()
	scanner := bufio.NewScanner(fs)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())

	}
	if err := scanner.Err(); err != nil {
		return []string{}, err
	}
	return lines, nil
}
