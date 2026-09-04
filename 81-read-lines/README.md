# 81 · Read a file line by line

## The idea in one paragraph

Tier 6 starts at the filesystem. The standard Go idiom for reading a file line by line is a four-step chain — `os.Open` → `defer f.Close()` → `bufio.NewScanner(f)` → `for scanner.Scan() { scanner.Text() }` — and it hides **two mistakes almost everyone ships at least once**. First: `Scan()` returns `false` for both *end of file* and *a read error*, so skipping the `scanner.Err()` check after the loop makes a failing disk read look exactly like a short file — silent data loss. Second: Scanner has a default maximum token size of **64KB**; a longer line stops the scan with `bufio.ErrTooLong`, and the fix (`scanner.Buffer(make([]byte, 0, initial), maxSize)`) is something you must know exists. This challenge pins both mistakes into the tests.

## Vocabulary

| term | meaning |
|---|---|
| `os.Open(path)` | opens read-only, returns `*os.File`; the error already wraps `os.ErrNotExist` for missing paths |
| `defer f.Close()` | the resource discipline — closed on every return path |
| `bufio.Scanner` | streaming tokenizer; with the default split function, one token per line |
| `scanner.Text()` | current line **without** the trailing `\n` — and without the `\r` of CRLF endings |
| `scanner.Err()` | the error that stopped the scan, or nil at true EOF — **checking it is not optional** |
| `bufio.ErrTooLong` | what you get when a line exceeds the token limit (default 64KB) |
| error wrapping | `fmt.Errorf("reading %s: %w", path, err)` — context added, `errors.Is` still works through it |

**The alternative to know:** `os.ReadFile` + `strings.Split` is simpler and fine for small files you want wholly in memory anyway; Scanner **streams**, holding one line at a time — the right shape for logs and anything unbounded. State this trade-off in a comment.

---

## The function

```go
func ReadLines(path string) ([]string, error)
```

Read the file at `path` and return its lines, in order.

## Pinned contract (everything the tests assert)

**Line semantics:**

- Lines are returned **without** line endings: no `\n`, and no `\r` for CRLF files (both endings must work; `scanner.Text()` handles both).
- A trailing newline on the last line does **not** produce an extra empty element: `"a\nb\nc\n"` → `["a", "b", "c"]` (3 elements, not 4).
- A last line **without** a trailing newline is still returned: `"a\nb\nc"` → `["a", "b", "c"]`.
- Blank lines anywhere are preserved as `""` — including consecutive ones.
- A file containing exactly `"\n"` → `[""]` (one empty line).
- An empty file → a slice of length 0 and a nil error.
- Content is bytes-in, bytes-out per line — Unicode passes through untouched.

**Error policy:**

- Missing file → non-nil wrapped error that includes the path in its message; `errors.Is(err, os.ErrNotExist)` must hold **through the wrap**. Never a panic.
- `path` naming a directory → `os.Open` succeeds, the first read fails; only the `scanner.Err()` check catches this, and it must surface as a non-nil error.
- A line longer than 64KB → keep the Scanner defaults (note the `scanner.Buffer` escape hatch in a comment, don't use it) → non-nil error with `errors.Is(err, bufio.ErrTooLong)` intact through your wrapping.
- On **any** error, return zero lines (length 0) — no partial results alongside an error.

## Worked examples

```
file: "alpha\nbeta\ngamma\n"      → (["alpha" "beta" "gamma"], nil)
file: "alpha\r\nbeta\r\n"         → (["alpha" "beta"], nil)          CRLF stripped
file: "alpha\n\nbeta"             → (["alpha" "" "beta"], nil)       blank kept, last line kept
file: ""                          → ([], nil)
path: "no/such/file.txt"          → (len 0, error: wraps os.ErrNotExist, mentions the path)
```

## Edge cases the tests hit

- Missing file (`errors.Is(err, os.ErrNotExist)` + path in the message).
- Empty file; a file that is just `"\n"`.
- No trailing newline on the last line.
- CRLF endings; mixed content with Unicode.
- Blank line in the middle, and several in a row.
- A directory passed as the path.
- A ~70KB single line → `bufio.ErrTooLong`.
- A 10,000-line file read back exactly.

## What the tests cover

All fixtures are built with `t.TempDir()` + `os.WriteFile` — no real project files are touched. The suite asserts exact slice equality on every content case, `errors.Is` through your wraps on every error case, and zero returned lines whenever an error is returned.

**Done when:** `defer f.Close()` and the `scanner.Err()` check are both present without prompting, and the no-trailing-newline test passes.
