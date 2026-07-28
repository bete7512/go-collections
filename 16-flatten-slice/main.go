package main

func main() {}
func Flatten(s [][]int) []int {
	var output []int

	for _, val := range s {
		if len(val) > 0 {
			output = append(output, val...)
		}
	}

	return output
}
