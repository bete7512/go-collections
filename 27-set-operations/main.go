package main

import "maps"

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
	if s.m == nil {
		return 0
	}
	return len(s.m)
}

// everything in either
func (s *Set) Union(other *Set) *Set {
	newSet := Set{m: maps.Clone(s.m)}
	for key := range other.m {
		newSet.Add(key)
	}
	return &newSet
}

// only what's in both
func (s *Set) Intersect(other *Set) *Set {
	newSet := NewSet()

	shortestSet := other.m
	if s.Len() < other.Len() {
		shortestSet = s.m
	}
	for key := range shortestSet {
		if _, ok := s.m[key]; ok {
			newSet.Add(key)
		}
	}
	return newSet
}

// in s, not in other  (s − other)
func (s *Set) Diff(other *Set) *Set {
	result := &Set{
		m: make(map[string]struct{}),
	}

	for key := range s.m {
		if _, ok := other.m[key]; !ok {
			result.Add(key)
		}
	}
	return result
}
