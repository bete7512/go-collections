package main

import (
	"sync"
)

func main() {}
func RunWorkers(n int) []int {
	if n <= 0 {
		return []int{}
	}
	var wg sync.WaitGroup
	workers := []int{}
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
		}(i)
		workers = append(workers, i)
	}
	wg.Wait()
	return workers
} // returns each worker's ID, any order
func SumParallel(nums []int, workers int) int {
	if len(nums) == 0 {
		return 0
	}
	if workers <= 0 {
		workers = 1
	}
	if workers > len(nums) {
		workers = len(nums)
	}

	results := make(chan int, workers)
	chunk := (len(nums) + workers - 1) / workers
	var wg sync.WaitGroup
	for w := range workers {
		start := w * chunk
		end := min(start+chunk, len(nums))
		wg.Add(1)
		go func(w, start, end int) {
			defer wg.Done()
			s := 0
			for _, v := range nums[start:end] {
				s += v
			}
			results <- s
		}(w, start, end)
	}
	// wg.Wait()
	go func(){
		wg.Wait()
		close(results)
	}()
	total := 0
	for  res := range results {
		total += res
	}
	return total
} // splits the work, sums the parts
