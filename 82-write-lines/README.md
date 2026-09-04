# 82 · Write a slice of strings to a file

## The idea in one paragraph

The mirror image of #81 — and the trap is inverted too. Reading punished you for skipping `scanner.Err()`; writing punishes you for skipping **`Flush()`**. A `bufio.Writer` batches your writes in memory and hands them to the file only when its buffer fills or you flush explicitly. `f.Close()` does **not** flush it — the buffered writer is a separate object that knows nothing about the file's Close — so the classic failure is a program that "worked", closed cleanly, and shipped an empty or truncated file. The second, subtler discipline: **close errors matter for writes**. A `defer f.Close()` that discards the error can hide the moment the OS finally tried to put your bytes on disk and failed. Flush, check; close, check.

## Vocabulary

| term | meaning |
|---|---|
| `os.Create(path)` | create-or-**truncate** for writing — an existing file is emptied the moment this returns |
| `bufio.NewWriter(f)` | buffered writer (4KB default); writes go to memory until the buffer fills or you `Flush()` |
| `w.Flush()` | pushes the buffered remainder to the file and reports the write error — the step everyone forgets |
| deferred-close error | `defer func() { err = errors.Join(err, f.Close()) }()` on a **named return** — the idiomatic way a close failure reaches the caller |
| `os.OpenFile(path, os.O_APPEND\|os.O_CREATE\|os.O_WRONLY, 0o644)` | the appending variant, when truncation is not what you want — know it exists |

**Do it wrong once (not graded, but do it):** comment out your `Flush()`, write a few short lines, and `cat` the file — empty. Then write ~10KB of lines and look again — *partially* written, because the buffer filled and auto-flushed once, then dropped the tail. That partial case is the one that gets past code review.

---

## The function

```go
func WriteLines(path string, lines []string) error
```

Write each line, in order, followed by `\n`, buffered, flushed, closed — and return any error from any of those stages.

## Pinned contract (everything the tests assert)

**Output format:**

- `["a", "b", "c"]` → the file contains exactly `"a\nb\nc\n"` — every line newline-terminated, including the last; no BOM, no extra bytes.
- An empty or nil slice → the file is **created** and is exactly 0 bytes; nil error. (Creating the empty file is the correct behavior, not an edge to reject.)
- Line content is written **verbatim** — no escaping, no trimming. A string that itself contains `\n` (e.g. `"a\nb"`) therefore produces more lines on disk than entries in the slice: `["a\nb", "c"]` → `"a\nb\nc\n"`. That asymmetry with #81's `ReadLines` is deliberate and documented: `WriteLines` writes what it is given; round-tripping is only exact for entries with no embedded newlines. The tests assert the verbatim behavior.
- Unicode and significant whitespace pass through untouched.

**Overwrite semantics:**

- Writing to an existing file **replaces it entirely** — after writing 2 short lines over a 10KB file, the file holds exactly those 2 lines and nothing else. No leftover tail bytes.

**Error policy:**

- A path whose parent directory doesn't exist → non-nil error, `errors.Is(err, os.ErrNotExist)` must hold through any wrapping, and the message must mention the path. Never a panic.
- A path that names an existing **directory** → non-nil error.
- On any error, whatever partial file state exists is unspecified — but the error must be returned, not swallowed. This includes flush and close errors (the tests can't fault your disk, but the round-trip test fails loudly if an unflushed buffer loses the tail).

## Worked examples

```
WriteLines(p, ["alpha", "beta", "gamma"])   → p holds "alpha\nbeta\ngamma\n"   (nil error)
WriteLines(p, [])                            → p exists, 0 bytes                (nil error)
WriteLines(p, ["", "x", ""])                 → p holds "\nx\n\n"                (blank lines are real lines)
p previously "AAAA…(10KB)…", then
WriteLines(p, ["tiny"])                      → p holds exactly "tiny\n"         (truncation, no leftovers)
WriteLines("/no/such/dir/f.txt", ["x"])      → error wrapping os.ErrNotExist, mentions the path
```

## Edge cases the tests hit

- Empty slice and nil slice (0-byte file, nil error).
- Blank strings as entries (`""` → a bare newline in the file).
- Entries with embedded newlines written verbatim.
- Overwriting a much larger existing file — no leftover bytes.
- Parent directory missing; path is a directory.
- 50,000 lines written and verified (buffering is what makes this fast — and the tail of a large write is exactly what an unflushed buffer loses).
- Unicode content byte-identical on disk.

## What the tests cover

All paths live under `t.TempDir()`. The headline test is the **round-trip**: write a table of line-slices, read the raw bytes back, and require exact equality with `join(lines, "\n") + "\n"` — plus a `ReadLines`-style scan for the no-embedded-newline cases. Format, overwrite, error-policy, and large-slice cases as listed above; `errors.Is` through wraps on the bad-path cases; a 0-byte-file assertion for the empty slice.

**Done when:** the round-trip test passes, the Flush-before-Close ordering is deliberate (you did it wrong once and watched the bytes vanish), and close errors aren't silently swallowed.
