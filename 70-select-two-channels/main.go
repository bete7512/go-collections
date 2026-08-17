package main

import "fmt"

func main() {}
func Merge2(a, b <-chan int) []int {
	values := []int{}
	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				a = nil
				continue
			}
			values = append(values, v)
		case v, ok := <-b:
			if !ok {
				b = nil
				continue
			}
			values = append(values, v)

		}
	}
	return values
} // drains both until both are closed
func Tagged(a, b <-chan int) []string {
	values := []string{}
	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				a = nil
				continue
			}
			values = append(values, fmt.Sprintf("a:%d", v))
		case v, ok := <-b:
			if !ok {
				b = nil
				continue
			}
			values = append(values, fmt.Sprintf("b:%d", v))

		}

	}
	return values
} // same, but records which source each value came from
func FirstReady(a, b <-chan int) (int, bool) {
	val := 0
	recieved := false
	select {
	case v, ok := <-a:
		if ok {
			val = v
		}
		recieved = ok
	case v, ok := <-b:
		if ok {
			val = v
		}
		recieved = ok

	}
	return val, recieved
} // one value from whichever is ready first
