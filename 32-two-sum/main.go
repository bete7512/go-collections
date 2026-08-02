package main

func main() {}
func TwoSum(nums []int, target int) (int, int, bool) {
	seen := map[int]int{}

	for i, val := range nums {
		compliment := target - val
		if _, ok := seen[val]; ok {
			return seen[val], i, true
		}
		seen[compliment] = i
	}

	return 0, 0, false
}
