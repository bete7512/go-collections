package main

import "sync"

func main() {}
func CountUntilDone(done <-chan struct{}) int {
	counter := int(0)
	for {
		select {
		case <-done:
			return counter
		default:
			counter++

		}
	}
}
func WorkersUntilDone(done <-chan struct{}, workers int) []int {
	if workers <= 0 {
		workers = 1
	}
	results := make([]int, workers)
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			select {
			case <-done:
				return
			default:
				results[w]++
			}
		}(w)
	}
	wg.Wait()
	return results
}

func ProcessUntilDone(done <-chan struct{}, jobs <-chan int, results chan<- int) int {
	counter := int(0)
	for {
		select {
		case <-done:
			return counter
		case v, ok := <-jobs:
			if !ok {
				return counter
			}
			select {
			case results <- v:
				counter++
			case <-done:
				return counter
			}
		}
	}
}

type Stopper struct { /* done channel, waitgroup, counts, sync.Once */
	done   chan struct{}
	wg     sync.WaitGroup
	counts []int // len == workers; counts[i] belongs to worker i
	once   sync.Once
}

func NewStopper(workers int) *Stopper {
	if workers <= 0 {
		workers = 1
	}
	s := &Stopper{
		done:   make(chan struct{}),
		counts: make([]int, workers),
	}
	s.wg.Add(workers) // add BEFORE launching, never inside the goroutine
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer s.wg.Done()
			for {
				select {
				case <-s.done:
					return // channel closed → exit promptly
				default:
					s.counts[id]++ // one unit of "work"
				}
			}
		}(i)
	}
	return s
}
func (s *Stopper) Stop() {
	s.once.Do(func() { close(s.done) })
}
func (s *Stopper) Wait() []int {
	s.wg.Wait()     // every worker has called Done() → all have exited
	return s.counts // safe to read now: no goroutine is writing anymore
}
