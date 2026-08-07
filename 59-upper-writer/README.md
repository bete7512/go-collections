# 59 · An io.Writer that uppercases

## Problem

`io.Writer` is one method — `Write([]byte) (int, error)` — and satisfying it plugs your type into a huge amount of existing machinery: `fmt.Fprintf`, `io.Copy`, `json.NewEncoder`, `log.New`, `bufio.NewWriter`, `http.ResponseWriter` pipelines, compression writers. That's the real lesson: a tiny interface is a large amount of interoperability, and this is why Go's stdlib composes so well.

The catch is that `Write` has a **contract**, not just a signature, and the docs spell it out. Getting the returned `n` wrong doesn't fail loudly — it fails later, inside `io.Copy`, as a mysterious `short write` error. This drill makes you honor the contract precisely.

## Contract (what the tests enforce)

```go
type UpperWriter struct {
    W io.Writer
}

func (u *UpperWriter) Write(p []byte) (int, error)
```

- **Uppercases and forwards.** Bytes written through the wrapper reach `u.W` uppercased (`bytes.ToUpper` handles ASCII and multi-byte UTF-8 correctly).
- **Return `n == len(p)` on success — the length of what you *consumed*, not what you wrote downstream.** These can differ: uppercasing can change byte length in principle, and any transformation could. Callers like `io.Copy` compare your `n` against `len(p)` and report `io.ErrShortWrite` when it's smaller. Returning the downstream writer's count is the classic bug this test suite hunts.
- **Never modify `p`.** The caller owns that slice and may reuse it across calls (`io.Copy` reuses one 32KB buffer for the whole stream). `bytes.ToUpper` returns a new slice, which is the safe path — but the test verifies it by passing a buffer and checking it afterward.
- **Don't retain `p`** beyond the call.
- **Empty write:** `Write(nil)` and `Write([]byte{})` return `(0, nil)` and write nothing downstream.
- **Propagate downstream errors.** If `u.W.Write` fails, return its error. When the downstream reports a short write (`n < len(p)` with nil error), surface that as an error too — returning `(shortN, nil)` would silently lose data.
- **Pointer receiver**, matching the stdlib convention for writers (they're stateful in general and must not be copied).
- Compile-time assertion in your file: `var _ io.Writer = (*UpperWriter)(nil)`.
- The zero value `UpperWriter{}` has a nil `W`; the tests don't exercise it — but note in a comment what would happen (nil-pointer panic on the forward), because "the zero value must be useful" is a Go principle you're consciously *not* satisfying here, and knowing when you've broken a convention is the point.

## Worked examples

**Example 1 — through fmt:**

```go
var buf bytes.Buffer
uw := &UpperWriter{W: &buf}

fmt.Fprint(uw, "hello")
buf.String()   // → "HELLO"
```

`fmt` has no idea it's talking to your type. That's the payoff of a one-method interface.

**Example 2 — the n contract:**

```go
n, err := uw.Write([]byte("abc"))
n     // → 3   (len of the INPUT)
err   // → nil
```

If uppercasing had produced 4 bytes downstream, `n` would still be 3. `n` answers "how much of my input did you take?", not "how many bytes did you emit?".

**Example 3 — composition with io.Copy:**

```go
var buf bytes.Buffer
io.Copy(&UpperWriter{W: &buf}, strings.NewReader("stream me"))
buf.String()   // → "STREAM ME"
```

`io.Copy` only succeeds if your `n` is honest; a wrong count fails here with `short write`.

## Edge cases the tests cover

- Simple ASCII, mixed case, already-uppercase, and non-letter bytes (digits/punctuation pass through).
- Accented UTF-8 (`"héllo"` → `"HÉLLO"`) and a multi-byte-only string.
- Bytes with no uppercase form (CJK, emoji) passing through unchanged.
- `n == len(p)` asserted on **every** write, including multi-byte cases.
- Empty write via nil slice and empty slice.
- Multiple sequential writes accumulating in order.
- **Caller's buffer unmodified** after `Write` returns.
- A failing downstream writer: the error propagates.
- A short-writing downstream writer: an error is returned rather than a silent partial success.
- `fmt.Fprintf` and `io.Copy` composition, end to end.
- Compile-time `io.Writer` satisfaction.
