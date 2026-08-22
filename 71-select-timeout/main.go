package main

import "time"

func main() {}
func ReadWithTimeout(ch <-chan int, d time.Duration) (int, bool) {
	select {
	case v, ok := <-ch:
		if ok {
			return v, true
		}
		return 0, false
	case <-time.After(d):
		return 0, false
	}
}
func ReadWithTimer(ch <-chan int, d time.Duration) (int, bool) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case v, ok := <-ch:
		return v, ok // ok=false means channel closed
	case <-t.C:
		return 0, false // timed out
	}
} // reusable timer, no per-call allocation churn
func DrainWithIdleTimeout(ch <-chan int, idle time.Duration) []int {
	values := []int{}
	for {
		select {
		case <-time.After(idle):
			return values
		case v, ok := <-ch:
			if !ok {
				return values
			}
			values = append(values, v)
		}
	}
}
