package main

func main() {}
func BinarySearch(s []int, target int) (int, bool) {
	result := -1

	i := 0
	j := len(s) - 1

	for i <= j {
		mid := (i + j) / 2

		if s[mid] == target {
			result = mid
			return result, true
		}

		if target < s[mid] {
			j = mid - 1
		}

		if target > s[mid] {
			i = mid + 1
		}
	}

	return -1, false
}
