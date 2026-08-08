package main

import "io"

func main() {}

type SliceReader struct {
	data []byte
	pos  int
}

func NewSliceReader(b []byte) *SliceReader {
	return &SliceReader{
		data: b,
	}
}
func (r *SliceReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
func (r *SliceReader) Len() int {
	return len(r.data) - r.pos
} // bytes remaining
