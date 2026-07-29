package main

func main() {}

type Set struct {
	m map[string]struct{}
}

func NewSet(items ...string) *Set {
	newSet := Set{
		m: make(map[string]struct{}),
	}
	for _, item := range items {
		newSet.m[item] = struct{}{}
	}
	return &newSet
}

func (s *Set) Add(v string) {
	if s.m == nil {
		s.m = make(map[string]struct{})
	}
	s.m[v] = struct{}{}

}

func (s *Set) Has(v string) bool {
	if _, ok := s.m[v]; !ok {
		return false
	}
	return true
}

func (s *Set) Remove(v string) {
	if _, ok := s.m[v]; !ok {
		return
	}
	delete(s.m, v)
}

func (s *Set) Len() int {
	return len(s.m)
}
