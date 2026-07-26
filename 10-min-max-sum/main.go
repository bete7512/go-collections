package main

import "errors"

func main() {}
func MinMaxSum(nums []int) (min, max, sum int, err error) {
	if len(nums) == 0 {
		return 0, 0, 0, errors.New("empty")
	}
	if len(nums) == 1 {
		return nums[0], nums[0], nums[0], nil
	}
	min = nums[0]
	max = nums[0]
	for _, val := range nums {
		if val > max {
			max = val
		}

		if val < min {
			min = val
		}
		sum += val
	}

	return min, max, sum, nil
}
