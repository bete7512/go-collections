package main

import (
	"sync"
)

func main() {}

type Result struct {
	Value    int // the input value
	WorkerID int // which worker handled it
}

func FanOut(in <-chan int, workers int) []Result {
	if workers <= 0 {
		workers = 1
	}
	results := []Result{}
	var wg sync.WaitGroup
	var mu sync.RWMutex
	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for v := range in {
				mu.Lock()
				results = append(results, Result{Value: v, WorkerID: w})
				mu.Unlock()
			}
		}(w)
	}
	wg.Wait()
	return results
}
func FanOutDispatch(in <-chan int, workers int) []Result {
	if workers <= 0 {
		workers = 1
	}
	results := []Result{}
	var wg sync.WaitGroup
	var mu sync.RWMutex

	chans := make([]chan int, workers)
	for i := range chans {
		chans[i] = make(chan int)
	}
	// dispatcher
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		for v := range in {
			chans[i%workers] <- v
			i++
		}
		for i := range chans {
			close(chans[i])
		}
	}()

	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for v := range chans[w] {
				mu.Lock()
				results = append(results, Result{Value: v, WorkerID: w})
				mu.Unlock()
			}
		}(w)
	}
	wg.Wait()

	return results
}

type StringResult struct {
	Value    string
	WorkerID int
}

func PartitionBy(in <-chan string, workers int, key func(string) int) []StringResult {
	if workers <= 0 {
		workers = 1
	}
	results := []StringResult{}
	var wg sync.WaitGroup
	var mu sync.RWMutex

	chans := make([]chan string, workers)
	for i := range chans {
		chans[i] = make(chan string)
	}
	// dispatcher
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		for v := range in {
			keyVal := key(v)
			if keyVal < 0 {
				keyVal = -keyVal
			}
			chans[keyVal%workers] <- v
			i++
		}
		for i := range chans {
			close(chans[i])
		}
	}()

	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for v := range chans[w] {
				mu.Lock()
				results = append(results, StringResult{Value: v, WorkerID: w})
				mu.Unlock()
			}
		}(w)
	}
	wg.Wait()

	return results
}
