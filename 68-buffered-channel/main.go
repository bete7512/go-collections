package main

func main() {}
func TrySend(ch chan int, v int) bool {
	select {
	case ch <- v:
		return true
	default:
		return false
	}
} // non-blocking send: false if it would block
func TryReceive(ch chan int) (int, bool) {
	val := 0
	select {
	case val, ok := <-ch:
		if !ok {
			return 0, true
		}
		return val, true
	default:
		return val, false
	}
} // non-blocking receive: false if it would block
func FillBuffer(ch chan int, vals []int) int {
	count := 0
	for _, val := range vals {
		ok := TrySend(ch, val)
		if ok {
			count++
		}
	}
	return count
} // sends until full; returns how many landed
func Stats(ch chan int) (length, capacity int) {
	return len(ch), cap(ch)
}
