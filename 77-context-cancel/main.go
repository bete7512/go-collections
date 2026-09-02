package main

import (
	"context"
	"sync"
)

func main() {}
func CountUntilCancel(ctx context.Context) int {
	counter := int(0)
	for {
		select {
		case <-ctx.Done():
			return counter
		default:
			counter++
		}
	}
}

func RunUntilCancel(ctx context.Context, workers int) int {
	if workers <= 0 {
		workers = 1
	}
	var wg sync.WaitGroup
	counts := make([]int, workers)
	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					counts[w]++
				}
			}
		}(w)
	}
	wg.Wait()
	sum := int(0)
	for _, v := range counts {
		sum += v
	}

	return sum

}

func ProcessUntilCancel(ctx context.Context, jobs <-chan int, results chan<- int) (int, error) {
	counter := int(0)
	for {
		select {
		case <-ctx.Done():
			return counter, ctx.Err()
		case v, ok := <-jobs:
			if !ok {
				return counter, nil
			}
			select {
			case results <- v * 2:
				counter++
			case <-ctx.Done():
				return counter, ctx.Err()
			}
		}
	}
}

type CancelRunner struct {
	cancel func()
	wg     sync.WaitGroup
	counts []int
}

func NewCancelRunner(workers int) *CancelRunner {
	if workers <= 0 {
		workers = 1
	}
	c := &CancelRunner{
		wg:     sync.WaitGroup{},
		counts: make([]int, workers),
	}
	c.wg.Add(workers)
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer c.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return // channel closed → exit promptly
				default:
					c.counts[id]++ // one unit of "work"
				}
			}
		}(i)
	}
	return c
}
func (r *CancelRunner) Stop() {
	r.cancel()
} // idempotent: safe to call many times, concurrently
func (r *CancelRunner) Wait() []int {
	r.wg.Wait()
	return r.counts
} // blocks until all workers exited; per-worker counts
