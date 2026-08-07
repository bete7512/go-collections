package main

import "errors"

func main() {}

type Ring struct {
	buf   []int
	head  int // next slot to read
	tail  int // next slot to write
	count int // number of stored elements — the ambiguity resolver
}

func NewRing(capacity int) (*Ring, error) {
	if capacity <= 0 {
		return nil, errors.New("invalid capacity")
	}

	return &Ring{buf: make([]int, capacity), head: 0, tail: 0}, nil
} // error when capacity <= 0
func (r *Ring) Write(v int) bool {
	if r.Len() == r.Cap() {
		return false
	}
	r.buf[r.tail] = v
	r.tail++
	if r.tail >= r.Cap() {
		r.tail = r.tail - r.Cap()
	}
	r.count++
	return true
} // false when full — value NOT stored
func (r *Ring) Read() (int, bool) {
	if r.Len() == 0 {
		return 0, false
	}
	val := r.buf[r.head]
	r.head++
	if r.head >= r.Cap() {
		r.head = r.head - r.Cap()
	}
	r.count--
	return val, true

} // false when empty
func (r *Ring) Len() int {
	return r.count
} // elements currently stored
func (r *Ring) Cap() int {
	return len(r.buf)
} // fixed capacity
