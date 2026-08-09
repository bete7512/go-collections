package main

func main() {}

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](items ...T) Set[T] {
	set := Set[T]{}

	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}
func (s Set[T]) Add(v T) {
	s[v] = struct{}{}
}
func (s Set[T]) Has(v T) bool {
	if _, ok := s[v]; ok {
		return true
	}
	return false
}
func (s Set[T]) Remove(v T) {
	delete(s, v)
}
func (s Set[T]) Len() int {
	return len(s)
}
func (s Set[T]) Items() []T {
	items := []T{}
	for key := range s {
		items = append(items, key)
	}
	return items
} // unordered; for tests and iteration
