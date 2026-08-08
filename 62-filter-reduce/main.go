package main

func main() {}
func Filter[T any](s []T, keep func(T) bool) []T {

	filtered := make([]T, 0, len(s))

	for _, val := range s {
		if keep(val) {
			filtered = append(filtered, val)
		}
	}

	return filtered

}
func Reduce[T, U any](s []T, init U, f func(acc U, v T) U) U {
	u := init

	for _, val := range s {
		u = f(u, val)
	}

	return u
}

func MapViaReduce[T, U any](s []T, f func(T) U) []U {
	return Reduce(s, make([]U, 0, len(s)), func(acc []U, v T) []U { return append(acc, f(v)) })
}
