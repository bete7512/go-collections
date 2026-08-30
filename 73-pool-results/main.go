package main

import (
	"slices"
	"sync"
)

func main() {}
func PoolCounted(jobs []int, workers int, process func(int) int) []int {
	if workers <= 0 {
		workers = 1
	}
	type task struct{ idx, val int }
	results := make(chan task, len(jobs))
	jobsCh := make(chan task)
	var wg sync.WaitGroup
	var mu sync.RWMutex
	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for job := range jobsCh {
				mu.Lock()
				results <- task{job.idx, process(job.val)}
				mu.Unlock()
			}
		}(w)
	}
	for i, job := range jobs {
		jobsCh <- task{i, job}
	}
	close(jobsCh)
	wg.Wait()
	close(results)

	resInts := make([]int, len(results))
	for task := range results {
		resInts[task.idx] = task.val
	}
	slices.Sort(resInts)
	return resInts
}
func PoolStream(jobs []int, workers int, process func(int) int) []int {
	if workers <= 0 {
		workers = 1
	}
	type task struct{ idx, val int }
	results := make(chan int, len(jobs))
	jobsCh := make(chan int)
	var wg sync.WaitGroup

	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for job := range jobsCh {
				results <- process(job)
			}
		}(w)
	}
	for _, job := range jobs {
		jobsCh <- job
	}
	close(jobsCh)
	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]int, 0, len(jobs))
	for v := range results {
		out = append(out, v)
	}
	return out
}
func PoolResultsWithErrors(jobs []int, workers int, process func(int) (int, error)) ([]int, []error) {
	if workers <= 0 {
		workers = 1
	}

	// One message per job, carrying EITHER outcome.
	type result struct {
		value int
		err   error
	}

	jobsCh := make(chan int)
	results := make(chan result, len(jobs)) // buffered so workers never block
	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobsCh {
				v, err := process(j)
				results <- result{value: v, err: err} // exactly one send, success or failure
			}
		}()
	}

	for _, j := range jobs {
		jobsCh <- j
	}
	close(jobsCh)

	go func() {
		wg.Wait()
		close(results)
	}()

	// Non-nil even when empty — grow by append, so each job lands in exactly ONE slice.
	values := make([]int, 0, len(jobs))
	errs := make([]error, 0, len(jobs))

	for r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
		} else {
			values = append(values, r.value)
		}
	}

	return values, errs
}
