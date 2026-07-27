package main

func main() {}
func RotateLeft(s []int, k int) {
	if k == 0 || len(s) == 0 {
		return
	}

	k = ((k % len(s)) + len(s)) % len(s)

	for i, j := 0, len(s[:k])-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	for i, j := k, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// naive me
// func RotateLeft(s []int, k int) {
// 	if k == 0 || len(s) == 0 {
// 		return
// 	}

// 	k = k % len(s)

// 	right := s[:k]
// 	left := s[k:]

// 	rotated := append(left, right...)

// 	copy(s, rotated)
// }
