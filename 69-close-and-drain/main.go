package main

func main() {
	ch := make(chan int)
	SafeClose(ch)
}
func DrainAll(ch <-chan int) []int {
	values := []int{}
	for v := range ch {
		values = append(values, v)
	}
	return values
} // range until closed
func DrainWithOK(ch <-chan int) ([]int, int) {
	values := []int{}
	count := 0
	for {
		count++
		v, ok := <-ch
		if !ok {
			break
		}
		values = append(values, v)
	}
	return values, count
} // manual receive; returns values + how many receives happened
func SendAndClose(ch chan<- int, vals []int) {
	for _, val := range vals {
		ch <- val
	}
	close(ch)
} // sends all, then closes
func CountAfterClose(ch <-chan int, n int) []bool {
	flags := []bool{}
	for i := 0; i < n; i++ {
		_, ok := <-ch
		flags = append(flags, ok)
	}
	return flags
} // n receives past drain; each element is the ok flag
func SafeClose(ch chan int) (closed bool) {
	defer func() {
		if r := recover(); r != nil {

		}
	}()
	close(ch)
	close(ch)
	return true
} // closes, recovering from a double-close panic
