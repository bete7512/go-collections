package main

func main() {}
func Memoize(f func(int) int) func(int) int{
	cache := map[int]int{}
	return func(i int) int {
		if _,ok := cache[i];ok{
			return cache[i]
		}
		cache[i] = f(i)
		return cache[i]
	}
}