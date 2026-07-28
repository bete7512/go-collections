package main

func main() {}
func MapSlice(s []int, f func(int) int) []int {

	out := make([]int, len(s))

	for i, val := range s {
		out[i] = f(val)
	}
	return out
}
