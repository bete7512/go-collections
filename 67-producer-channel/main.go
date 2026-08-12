package main

import "sync"

func main() {}
func Produce(n int) <-chan int {
	result := make(chan int)
	if n <= 0 {
		close(result)
		return result
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range n {
			result <- i + 1
		}
	}()
	go func() {
		wg.Wait()
		close(result)
	}()
	return result
} // sends 1..n, then closes
func Collect(ch <-chan int) []int {
	collected := []int{}
	for res := range ch {
		collected = append(collected, res)
	}
	return collected
} // ranges until the channel closes
func ProduceSquares(nums []int) <-chan int {
	var wg sync.WaitGroup
	result := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, val := range nums {
			result <- val * val
		}
		close(result)
	}()
	go func() {
		wg.Wait()
	}()
	return result
} // sends each input squared, then closes
