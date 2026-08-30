package main

import "sync"

func main() {}
func RunPool(jobs []int, workers int) int {
	if workers < 1 {
		workers = 1
	}
	var wg sync.WaitGroup
	var counter int
	jobsCh := make(chan int)
	var mu sync.RWMutex
	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for _ = range jobsCh {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}(w)
	}

	for job := range jobs {
		jobsCh <- job
	}
	close(jobsCh)
	wg.Wait()
	return counter
} // processes every job exactly once; returns how many were processed

func PoolWithWorkerIDs(jobs []int, workers int) map[int]int {
	if workers < 1 {
		workers = 1
	}
	var wg sync.WaitGroup
	jobWithWorker := make(map[int]int)
	jobsCh := make(chan int)
	var mu sync.RWMutex
	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for job := range jobsCh {
				mu.Lock()
				jobWithWorker[job] = w
				mu.Unlock()
			}
		}(w)
	}

	for _, job := range jobs {
		jobsCh <- job
	}
	close(jobsCh)
	wg.Wait()

	return jobWithWorker
} // job value -> ID of the worker that handled it

func ProcessAll(jobs []int, workers int, process func(int) int) []int {
	if workers < 1 {
		workers = 1
	}
	results := make([]int, len(jobs))
	var wg sync.WaitGroup
	type task struct{ idx, val int }
	jobsCh := make(chan task)
	var mu sync.RWMutex
	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for job := range jobsCh {
				mu.Lock()
				results[job.idx] = process(job.val)
				mu.Unlock()

			}
		}(w)
	}

	for i, job := range jobs {
		jobsCh <- task{i, job}
	}
	close(jobsCh)
	wg.Wait()
	return results
} // applies process to each job; results in INPUT order
