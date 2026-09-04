package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readBack(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back %s: %v", path, err)
	}
	return string(data)
}

func TestWriteLinesFormat(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{
			name:  "three lines each newline terminated",
			lines: []string{"alpha", "beta", "gamma"},
			want:  "alpha\nbeta\ngamma\n",
		},
		{
			name:  "single line gets trailing newline",
			lines: []string{"only"},
			want:  "only\n",
		},
		{
			name:  "blank entries become bare newlines",
			lines: []string{"", "x", ""},
			want:  "\nx\n\n",
		},
		{
			name:  "embedded newline written verbatim",
			lines: []string{"a\nb", "c"},
			want:  "a\nb\nc\n",
		},
		{
			name:  "unicode byte identical",
			lines: []string{"héllo wörld", "日本語の行", "émoji 🎉"},
			want:  "héllo wörld\n日本語の行\némoji 🎉\n",
		},
		{
			name:  "whitespace preserved",
			lines: []string{"  indented", "trailing   ", "\ttab"},
			want:  "  indented\ntrailing   \n\ttab\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "out.txt")
			if err := WriteLines(path, tt.lines); err != nil {
				t.Fatalf("WriteLines() error = %v, want nil", err)
			}
			if got := readBack(t, path); got != tt.want {
				t.Errorf("file = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteLinesEmptySlice(t *testing.T) {
	for _, tt := range []struct {
		name  string
		lines []string
	}{
		{"empty slice", []string{}},
		{"nil slice", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "empty.txt")
			if err := WriteLines(path, tt.lines); err != nil {
				t.Fatalf("WriteLines() error = %v, want nil — an empty slice is not an error", err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("the file must still be created: %v", err)
			}
			if info.Size() != 0 {
				t.Errorf("file size = %d bytes, want 0", info.Size())
			}
		})
	}
}

func TestWriteLinesOverwritesEntirely(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")

	big := strings.Repeat("AAAAAAAA\n", 1300) // >10KB of prior content
	if err := os.WriteFile(path, []byte(big), 0o644); err != nil {
		t.Fatalf("seeding existing file: %v", err)
	}

	if err := WriteLines(path, []string{"tiny"}); err != nil {
		t.Fatalf("WriteLines() error = %v", err)
	}

	if got := readBack(t, path); got != "tiny\n" {
		t.Errorf("file = %q, want %q — overwriting must leave no leftover bytes from the old, longer file", got, "tiny\n")
	}
}

func TestWriteLinesRoundTrip(t *testing.T) {
	// Write, read the bytes back, split — entries without embedded
	// newlines must survive the trip exactly.
	cases := [][]string{
		{"a", "b", "c"},
		{"solo"},
		{"", "", ""},
		{"mixed", "", "  spaced  ", "日本語"},
	}

	for i, lines := range cases {
		path := filepath.Join(t.TempDir(), "rt.txt")
		if err := WriteLines(path, lines); err != nil {
			t.Fatalf("case %d: WriteLines() error = %v", i, err)
		}

		content := readBack(t, path)
		if !strings.HasSuffix(content, "\n") {
			t.Fatalf("case %d: file %q does not end in a newline", i, content)
		}
		got := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
		if len(got) != len(lines) {
			t.Fatalf("case %d: round-trip produced %d lines, want %d", i, len(got), len(lines))
		}
		for j := range lines {
			if got[j] != lines[j] {
				t.Errorf("case %d line %d: round-trip = %q, want %q", i, j, got[j], lines[j])
			}
		}
	}
}

func TestWriteLinesMissingParentDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no", "such", "dir", "out.txt")
	err := WriteLines(path, []string{"x"})

	if err == nil {
		t.Fatalf("WriteLines(missing parent dir) error = nil, want an error — never a panic")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("errors.Is(err, os.ErrNotExist) = false for %v — wrap with %%w, don't swallow", err)
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error %q does not mention the path", err)
	}
}

func TestWriteLinesPathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := WriteLines(dir, []string{"x"}); err == nil {
		t.Fatalf("WriteLines(directory) error = nil, want an error")
	}
}

func TestWriteLinesLarge(t *testing.T) {
	// 50k lines. The tail of a large write is exactly what an unflushed
	// bufio.Writer loses: the buffer auto-flushes while full, then the
	// final partial buffer silently disappears without Flush().
	const n = 50000
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("line-%d", i)
	}

	path := filepath.Join(t.TempDir(), "large.txt")
	if err := WriteLines(path, lines); err != nil {
		t.Fatalf("WriteLines() error = %v", err)
	}

	content := readBack(t, path)
	got := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	if len(got) != n {
		t.Fatalf("round-trip produced %d lines, want %d — a missing Flush() drops the tail", len(got), n)
	}
	for _, i := range []int{0, 1, n / 2, n - 2, n - 1} {
		if want := fmt.Sprintf("line-%d", i); got[i] != want {
			t.Errorf("line %d = %q, want %q", i, got[i], want)
		}
	}
}
