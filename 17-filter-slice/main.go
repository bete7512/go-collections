package main

func main() {}
func Filter(s []int, keep func(int) bool) []int {

	outputs := make([]int, 0, len(s))
	for _, val := range s {
		if keep(val) {
			outputs = append(outputs, val)
		}
	}

	return outputs
}
