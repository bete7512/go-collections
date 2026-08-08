package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
)

// Compile-time proof that *SliceReader satisfies io.Reader.
var _ io.Reader = (*SliceReader)(nil)

func TestChunkedReads(t *testing.T) {
	r := NewSliceReader([]byte("hello world")) // 11 bytes
	p := make([]byte, 4)

	steps := []struct {
		wantN    int
		wantEOF  bool
		wantData string
	}{
		{4, false, "hell"},
		{4, false, "o wo"},
		{3, false, "rld"},
		{0, true, ""},
	}

	for i, step := range steps {
		n, err := r.Read(p)

		if n != step.wantN {
			t.Fatalf("read %d: n = %d, want %d", i, n, step.wantN)
		}
		if step.wantEOF {
			if !errors.Is(err, io.EOF) {
				t.Fatalf("read %d: err = %v, want io.EOF", i, err)
			}
		} else {
			if err != nil {
				t.Fatalf("read %d: err = %v, want nil (data still being delivered)", i, err)
			}
			if got := string(p[:n]); got != step.wantData {
				t.Fatalf("read %d: p[:n] = %q, want %q", i, got, step.wantData)
			}
		}
	}
}

func TestBufferLargerThanData(t *testing.T) {
	r := NewSliceReader([]byte("hi"))
	p := make([]byte, 100)

	n, err := r.Read(p)
	if err != nil {
		t.Fatalf("partial fill returned error %v, want nil", err)
	}
	if n != 2 {
		t.Fatalf("n = %d, want 2", n)
	}
	if got := string(p[:n]); got != "hi" {
		t.Errorf("p[:n] = %q, want %q", got, "hi")
	}

	if n, err := r.Read(p); n != 0 || !errors.Is(err, io.EOF) {
		t.Errorf("second read = (%d, %v), want (0, io.EOF)", n, err)
	}
}

func TestExactSizeBufferThenEOF(t *testing.T) {
	data := []byte("exact")
	r := NewSliceReader(data)
	p := make([]byte, len(data))

	n, err := r.Read(p)
	if n != len(data) || err != nil {
		t.Fatalf("read = (%d, %v), want (%d, nil) — EOF comes on the NEXT call", n, err, len(data))
	}

	n, err = r.Read(p)
	if n != 0 || !errors.Is(err, io.EOF) {
		t.Errorf("read after exhausting = (%d, %v), want (0, io.EOF)", n, err)
	}
}

func TestEmptyData(t *testing.T) {
	for _, data := range [][]byte{nil, {}} {
		r := NewSliceReader(data)
		p := make([]byte, 8)

		n, err := r.Read(p)
		if n != 0 || !errors.Is(err, io.EOF) {
			t.Errorf("first read over empty data = (%d, %v), want (0, io.EOF)", n, err)
		}
	}
}

func TestZeroLengthBuffer(t *testing.T) {
	r := NewSliceReader([]byte("data remains"))

	n, err := r.Read([]byte{})
	if n != 0 || err != nil {
		t.Errorf("Read(zero-length buffer) = (%d, %v), want (0, nil)", n, err)
	}
	if got := r.Len(); got != len("data remains") {
		t.Errorf("Len() = %d after a zero-length read, want %d — position must not move",
			got, len("data remains"))
	}
}

func TestRepeatedReadsAfterEOF(t *testing.T) {
	r := NewSliceReader([]byte("x"))
	p := make([]byte, 4)

	r.Read(p) // consume it

	for i := 0; i < 5; i++ {
		n, err := r.Read(p)
		if n != 0 || !errors.Is(err, io.EOF) {
			t.Fatalf("post-EOF read %d = (%d, %v), want (0, io.EOF) — EOF must be idempotent", i, n, err)
		}
	}
}

func TestNeverReturnsZeroNilWithDataRemaining(t *testing.T) {
	// A (0, nil) return with data left makes callers spin forever.
	r := NewSliceReader([]byte("abcdefghij"))
	p := make([]byte, 3)

	for i := 0; i < 100; i++ {
		n, err := r.Read(p)
		if errors.Is(err, io.EOF) {
			return // drained cleanly
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n == 0 {
			t.Fatalf("read %d returned (0, nil) with %d bytes remaining — callers would spin", i, r.Len())
		}
	}
	t.Fatalf("reader never reached EOF")
}

func TestReadAllRoundTrip(t *testing.T) {
	inputs := []string{
		"",
		"a",
		"hello world",
		strings.Repeat("abcdefghij", 10_000), // 100KB, forces many reads
	}

	for _, input := range inputs {
		name := input
		if len(name) > 20 {
			name = name[:20] + "..."
		}
		t.Run(name, func(t *testing.T) {
			r := NewSliceReader([]byte(input))

			got, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("io.ReadAll error: %v", err)
			}
			if string(got) != input {
				t.Errorf("round trip produced %d bytes, want %d", len(got), len(input))
			}
		})
	}
}

func TestIoCopyRoundTrip(t *testing.T) {
	input := "copy this through the reader"
	r := NewSliceReader([]byte(input))

	var buf bytes.Buffer
	n, err := io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	if n != int64(len(input)) {
		t.Errorf("copied %d bytes, want %d", n, len(input))
	}
	if buf.String() != input {
		t.Errorf("copied %q, want %q", buf.String(), input)
	}
}

func TestBufioScannerConsumesIt(t *testing.T) {
	// A real consumer with its own buffering strategy.
	r := NewSliceReader([]byte("line one\nline two\nline three"))

	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	want := []string{"line one", "line two", "line three"}
	if !slices.Equal(lines, want) {
		t.Errorf("scanned %v, want %v", lines, want)
	}
}

func TestLenTracksRemaining(t *testing.T) {
	data := []byte("0123456789")
	r := NewSliceReader(data)
	p := make([]byte, 3)

	if got := r.Len(); got != 10 {
		t.Fatalf("initial Len() = %d, want 10", got)
	}

	wantAfter := []int{7, 4, 1, 0}
	for i, want := range wantAfter {
		r.Read(p)
		if got := r.Len(); got != want {
			t.Fatalf("after read %d: Len() = %d, want %d", i, got, want)
		}
	}
}

func TestDoesNotModifySourceData(t *testing.T) {
	data := []byte("original bytes")
	snapshot := slices.Clone(data)

	r := NewSliceReader(data)
	p := make([]byte, 4)
	for {
		if _, err := r.Read(p); err != nil {
			break
		}
		// Scribble on the caller's buffer between reads: must not affect data.
		for i := range p {
			p[i] = 'X'
		}
	}

	if !slices.Equal(data, snapshot) {
		t.Errorf("source data was modified: %q, want %q", data, snapshot)
	}
}
