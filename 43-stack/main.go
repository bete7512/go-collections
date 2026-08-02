package main

func main() {}

type Stack struct {
	items []int
}

func (s *Stack) Push(v int) {
	s.items = append(s.items, v)
}
func (s *Stack) Pop() (int, bool) {
	if s.Len() <= 0 {
		return 0, false
	}

	last := s.items[s.Len()-1]
	s.items = s.items[:s.Len()-1]
	return last, true
}
func (s *Stack) Peek() (int, bool) {
	if s.Len() <= 0 {
		return 0, false
	}

	last := s.items[s.Len()-1]
	return last, true
}
func (s *Stack) Empty() bool {
	if s.Len() <= 0 {
		return true
	}
	return false
}
func (s *Stack) Len() int {
	return len(s.items)
}
