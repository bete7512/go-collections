package main

import (
	"bytes"
	"io"
)

func main() {}

type UpperWriter struct {
	W io.Writer
}

func (u *UpperWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	upper := bytes.ToUpper(p)
	n, err := u.W.Write(upper)
	if err != nil {
		return n, err
	}
	if n < len(upper) {
		return n, io.ErrShortWrite
	}
	return len(p), nil
}
