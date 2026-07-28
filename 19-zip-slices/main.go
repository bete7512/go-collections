package main

func main() {}

type Pair struct{ A, B int }

func Zip(a, b []int) []Pair {

	min := len(a)
	if len(a) > len(b) {
		min = len(b)
	}

	pairs := []Pair{}
	for i := 0; i < min; i = i + 1 {
		pairs = append(pairs, Pair{a[i], b[i]})
	}
	return pairs
}
