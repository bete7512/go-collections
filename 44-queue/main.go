package main

func main() {}

type Queue struct {
	items []int
}

func (q *Queue) Enqueue(v int) {
	q.items = append(q.items, v)
}
func (q *Queue) Dequeue() (int, bool) {
	if q.Len() <= 0 {
		return 0, false
	}

	first := q.items[0]

	q.items = q.items[1:]
	return first, true
}
func (q *Queue) Len() int {
	return len(q.items)
}
func (q *Queue) Empty() bool {
	if len(q.items) <= 0 {
		return true
	}
	return false
}
