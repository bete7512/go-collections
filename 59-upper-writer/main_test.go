package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

// Compile-time proof that *UpperWriter satisfies io.Writer.
var _ io.Writer = (*UpperWriter)(nil)

func TestWriteUppercases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase ascii", "hello", "HELLO"},
		{"mixed case", "HeLLo WoRld", "HELLO WORLD"},
		{"already uppercase", "SHOUT", "SHOUT"},
		{"digits and punctuation", "go1.24! @#", "GO1.24! @#"},
		{"accented utf8", "héllo", "HÉLLO"},
		{"accented only", "éàü", "ÉÀÜ"},
		{"cjk has no upper form", "世界", "世界"},
		{"emoji passes through", "go 🚀", "GO 🚀"},
		{"mixed script", "abc世界déf", "ABC世界DÉF"},
		{"single char", "a", "A"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			uw := &UpperWriter{W: &buf}

			p := []byte(tc.input)
			n, err := uw.Write(p)

			if err != nil {
				t.Fatalf("Write(%q) returned error: %v", tc.input, err)
			}
			if n != len(p) {
				t.Errorf("Write(%q) = %d, want %d (n must be the length of the INPUT consumed)",
					tc.input, n, len(p))
			}
			if got := buf.String(); got != tc.expected {
				t.Errorf("downstream got %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestEmptyWrites(t *testing.T) {
	var buf bytes.Buffer
	uw := &UpperWriter{W: &buf}

	n, err := uw.Write(nil)
	if err != nil || n != 0 {
		t.Errorf("Write(nil) = (%d, %v), want (0, nil)", n, err)
	}

	n, err = uw.Write([]byte{})
	if err != nil || n != 0 {
		t.Errorf("Write(empty) = (%d, %v), want (0, nil)", n, err)
	}

	if buf.Len() != 0 {
		t.Errorf("empty writes produced %q downstream, want nothing", buf.String())
	}
}

func TestSequentialWrites(t *testing.T) {
	var buf bytes.Buffer
	uw := &UpperWriter{W: &buf}

	for _, part := range []string{"one ", "two ", "three"} {
		p := []byte(part)
		n, err := uw.Write(p)
		if err != nil {
			t.Fatalf("Write(%q) error: %v", part, err)
		}
		if n != len(p) {
			t.Fatalf("Write(%q) = %d, want %d", part, n, len(p))
		}
	}

	if got, want := buf.String(), "ONE TWO THREE"; got != want {
		t.Errorf("accumulated = %q, want %q", got, want)
	}
}

func TestDoesNotModifyCallersBuffer(t *testing.T) {
	var buf bytes.Buffer
	uw := &UpperWriter{W: &buf}

	p := []byte("hello world")
	original := string(p)

	if _, err := uw.Write(p); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	if string(p) != original {
		t.Errorf("caller's slice was modified: %q, want %q — the caller owns p and may reuse it",
			string(p), original)
	}
}

// failingWriter always fails.
type failingWriter struct{ err error }

func (f failingWriter) Write(p []byte) (int, error) { return 0, f.err }

func TestPropagatesDownstreamError(t *testing.T) {
	sentinel := errors.New("disk on fire")
	uw := &UpperWriter{W: failingWriter{err: sentinel}}

	_, err := uw.Write([]byte("hello"))
	if err == nil {
		t.Fatalf("Write returned nil error when the downstream writer failed")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("errors.Is(err, downstreamErr) = false; got %v", err)
	}
}

// shortWriter accepts only the first byte and reports no error — the exact
// situation the io.Writer contract says must be surfaced as an error.
type shortWriter struct{ buf bytes.Buffer }

func (s *shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	s.buf.Write(p[:1])
	return 1, nil
}

func TestShortDownstreamWriteIsAnError(t *testing.T) {
	sw := &shortWriter{}
	uw := &UpperWriter{W: sw}

	n, err := uw.Write([]byte("hello"))
	if err == nil {
		t.Fatalf("Write = (%d, nil) with a short-writing downstream; want an error — "+
			"returning a partial success silently loses data", n)
	}
}

func TestComposesWithFmt(t *testing.T) {
	var buf bytes.Buffer
	uw := &UpperWriter{W: &buf}

	if _, err := fmt.Fprintf(uw, "user %s has %d items", "bete", 7); err != nil {
		t.Fatalf("Fprintf error: %v", err)
	}

	if got, want := buf.String(), "USER BETE HAS 7 ITEMS"; got != want {
		t.Errorf("Fprintf produced %q, want %q", got, want)
	}
}

func TestComposesWithIoCopy(t *testing.T) {
	// io.Copy checks n against len(p) on every chunk: a dishonest n fails
	// here with io.ErrShortWrite.
	var buf bytes.Buffer
	uw := &UpperWriter{W: &buf}

	src := strings.NewReader("stream this text through the writer")
	n, err := io.Copy(uw, src)
	if err != nil {
		t.Fatalf("io.Copy error: %v (a wrong n reports short write)", err)
	}
	if want := int64(len("stream this text through the writer")); n != want {
		t.Errorf("io.Copy copied %d bytes, want %d", n, want)
	}

	if got, want := buf.String(), "STREAM THIS TEXT THROUGH THE WRITER"; got != want {
		t.Errorf("copied content = %q, want %q", got, want)
	}
}

func TestLargeCopy(t *testing.T) {
	// Larger than io.Copy's internal 32KB buffer: forces multiple Write calls
	// with a REUSED buffer, which is where modifying p would corrupt data.
	var buf bytes.Buffer
	uw := &UpperWriter{W: &buf}

	chunk := "abcdefghij"
	src := strings.NewReader(strings.Repeat(chunk, 10_000)) // 100KB

	if _, err := io.Copy(uw, src); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}

	want := strings.Repeat(strings.ToUpper(chunk), 10_000)
	if buf.String() != want {
		t.Errorf("large copy produced %d bytes, want %d (content mismatch)", buf.Len(), len(want))
	}
}
