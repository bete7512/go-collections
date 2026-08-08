# 60 · An io.Reader over a byte slice

## Problem

The other half of Go's I/O universe. `io.Reader` is also one method — `Read([]byte) (int, error)` — but its contract is meaningfully subtler than `io.Writer`'s, because the reader **fills a buffer the caller owns** and has to report both progress and termination through the same two return values. You're reimplementing the core of `bytes.Reader`, and the EOF rules are where people get it wrong.

The one rule that surprises everyone: `Read` is allowed to return fewer bytes than the buffer holds, *and* it may return `n > 0` together with `io.EOF`. Callers must handle bytes-and-error together — which is why the canonical loop processes `n` bytes **before** checking `err`.

## Contract (what the tests enforce)

```go
type SliceReader struct {
    data []byte
    pos  int
}

func NewSliceReader(b []byte) *SliceReader
func (r *SliceReader) Read(p []byte) (int, error)
func (r *SliceReader) Len() int   // bytes remaining
```

- **Copy up to `len(p)` bytes** from the unread remainder into `p`, advance `pos`, return the count.
- **EOF policy (pinned):** return `(n, nil)` while data is being delivered, and `(0, io.EOF)` **only** once nothing remains. So a reader over 4 bytes read with a 4-byte buffer returns `(4, nil)` and then `(0, io.EOF)` on the next call.
  - The alternative — returning `(4, io.EOF)` together — is equally legal per the docs, and `io.ReadAll` handles both. This suite pins the first because it's what `bytes.Reader` does and testing it exactly requires picking one. Note the other form in a comment.
- **Never return `(0, nil)` when data remains and `len(p) > 0`.** The docs call this discouraged: a caller looping on it spins forever. (`(0, nil)` *is* correct when `len(p) == 0` — that's the one exception, and it's tested.)
- **Read after EOF keeps returning `(0, io.EOF)`** — idempotent, never a panic, never a reset.
- **Empty data** → the very first `Read` returns `(0, io.EOF)`.
- **Zero-length buffer** → `(0, nil)` regardless of remaining data, and `pos` must not move.
- **Never modify or retain `data`;** never assume `p` is zeroed or empty — you fill it, you don't append to it.
- `Len()` reports unread bytes remaining: `len(data) - pos`, reaching 0 at EOF.
- Compile-time assertion in your file: `var _ io.Reader = (*SliceReader)(nil)`.
- Pointer receiver — `Read` advances state.

## Worked examples

**Example 1 — chunked reading:**

```go
r := NewSliceReader([]byte("hello world"))   // 11 bytes
p := make([]byte, 4)

r.Read(p)   // (4, nil)  p = "hell"
r.Read(p)   // (4, nil)  p = "o wo"
r.Read(p)   // (3, nil)  p = "rld" + one stale byte from the previous read
r.Read(p)   // (0, io.EOF)
```

Note the third call: it returns 3, and `p[3]` still holds `o` from before. **The caller must use `p[:n]`, never all of `p`** — that's why `n` exists, and it's a real bug source when reading into a reused buffer.

**Example 2 — the payoff:**

```go
r := NewSliceReader([]byte("round trip"))
got, err := io.ReadAll(r)   // ("round trip", nil)
```

`io.ReadAll` drives your `Read` in a loop with a growing buffer. If it round-trips, your contract is right.

**Example 3 — buffer larger than the data:**

```go
r := NewSliceReader([]byte("hi"))
p := make([]byte, 100)
r.Read(p)   // (2, nil)  — a partial fill is normal, not an error
r.Read(p)   // (0, io.EOF)
```

## Edge cases the tests cover

- Buffer smaller than the data (multiple reads, exact `(n, err)` sequence asserted).
- Buffer larger than the data (partial fill).
- Buffer exactly the data's size, then EOF on the following call.
- Empty data → immediate EOF.
- Zero-length buffer with data remaining → `(0, nil)`, position unchanged.
- Repeated reads after EOF (called five times).
- `io.ReadAll` round-trip on empty, small, and 100KB inputs.
- `io.Copy` into a `bytes.Buffer` round-trip.
- `bufio.Scanner` reading lines through it — a real consumer with its own buffering.
- The caller's buffer beyond `n` being untouched (stale-data-is-normal, use `p[:n]`).
- `Len()` decreasing correctly across reads and hitting 0 at EOF.
- The reader not modifying or aliasing the original data slice.
