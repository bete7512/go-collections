package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeFixture(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	return path
}

func TestReadLinesContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "three lines with trailing newline",
			content: "alpha\nbeta\ngamma\n",
			want:    []string{"alpha", "beta", "gamma"},
		},
		{
			name:    "no trailing newline on last line",
			content: "alpha\nbeta\ngamma",
			want:    []string{"alpha", "beta", "gamma"},
		},
		{
			name:    "single line no newline",
			content: "only",
			want:    []string{"only"},
		},
		{
			name:    "crlf endings stripped",
			content: "alpha\r\nbeta\r\ngamma\r\n",
			want:    []string{"alpha", "beta", "gamma"},
		},
		{
			name:    "blank line in the middle preserved",
			content: "alpha\n\nbeta\n",
			want:    []string{"alpha", "", "beta"},
		},
		{
			name:    "consecutive blank lines preserved",
			content: "a\n\n\n\nb\n",
			want:    []string{"a", "", "", "", "b"},
		},
		{
			name:    "file of exactly one newline",
			content: "\n",
			want:    []string{""},
		},
		{
			name:    "unicode passes through",
			content: "héllo wörld\n日本語の行\némoji 🎉 line\n",
			want:    []string{"héllo wörld", "日本語の行", "émoji 🎉 line"},
		},
		{
			name:    "whitespace is content",
			content: "  indented\ntrailing spaces   \n\ttab\n",
			want:    []string{"  indented", "trailing spaces   ", "\ttab"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFixture(t, "fixture.txt", tt.content)
			got, err := ReadLines(path)
			if err != nil {
				t.Fatalf("ReadLines() error = %v, want nil", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("ReadLines() = %q, want %q", got, tt.want)
			}
			for i, line := range got {
				if strings.ContainsAny(line, "\r\n") {
					t.Errorf("line %d %q still contains a line ending", i, line)
				}
			}
		})
	}
}

func TestReadLinesEmptyFile(t *testing.T) {
	path := writeFixture(t, "empty.txt", "")
	got, err := ReadLines(path)
	if err != nil {
		t.Fatalf("ReadLines(empty) error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("ReadLines(empty) = %q, want no lines", got)
	}
}

func TestReadLinesMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does", "not", "exist.txt")
	got, err := ReadLines(path)

	if err == nil {
		t.Fatalf("ReadLines(missing) error = nil, want an error — never a panic, never a silent empty result")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("errors.Is(err, os.ErrNotExist) = false for %v — wrap with %%w, don't swallow", err)
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error %q does not mention the path — wrap with context (e.g. \"reading %%s: %%w\")", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d lines alongside an error, want none", len(got))
	}
}

func TestReadLinesDirectoryPath(t *testing.T) {
	// os.Open on a directory succeeds; the first Read fails. Only the
	// scanner.Err() check after the loop can surface this — without it,
	// a directory silently reads as an empty file.
	dir := t.TempDir()
	got, err := ReadLines(dir)

	if err == nil {
		t.Fatalf("ReadLines(directory) error = nil — scanner.Err() after the loop is not optional")
	}
	if len(got) != 0 {
		t.Errorf("got %d lines alongside an error, want none", len(got))
	}
}

func TestReadLinesLineOver64KB(t *testing.T) {
	// Scanner's default max token size is 64KB. This line is ~70KB, so
	// with default limits Scan() stops and scanner.Err() reports
	// bufio.ErrTooLong — which must survive your wrapping.
	long := strings.Repeat("x", 70*1024)
	path := writeFixture(t, "long.txt", "short\n"+long+"\nafter\n")

	got, err := ReadLines(path)
	if err == nil {
		t.Fatalf("ReadLines(70KB line) error = nil, want bufio.ErrTooLong under default Scanner limits")
	}
	if !errors.Is(err, bufio.ErrTooLong) {
		t.Errorf("errors.Is(err, bufio.ErrTooLong) = false for %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d lines alongside an error, want none", len(got))
	}
}

func TestReadLinesManyLines(t *testing.T) {
	const n = 10000
	var sb strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&sb, "line-%d\n", i)
	}
	path := writeFixture(t, "many.txt", sb.String())

	got, err := ReadLines(path)
	if err != nil {
		t.Fatalf("ReadLines() error = %v", err)
	}
	if len(got) != n {
		t.Fatalf("got %d lines, want %d", len(got), n)
	}
	for _, i := range []int{0, 1, 4999, 9998, 9999} {
		if want := fmt.Sprintf("line-%d", i); got[i] != want {
			t.Errorf("line %d = %q, want %q", i, got[i], want)
		}
	}
}
