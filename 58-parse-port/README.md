# 58 · (T, error) with every branch handled

## Problem

A realistic parse-and-validate function: turn a string into a TCP port, or explain precisely why you can't. Several *different* things can go wrong — it isn't a number, it's a number too large for `int`, it's a number outside 1–65535 — and each deserves its own identifiable failure. This is where #54's typed errors, #55's `%w` wrapping, and #57's sentinels stop being separate exercises and become one habit.

The specific discipline being drilled: **wrap the stdlib's error, don't replace it.** `strconv.Atoi` already produced a precise, classifiable error; flattening it into `errors.New("bad port")` throws away information your caller might need (`strconv.ErrSyntax` vs `strconv.ErrRange` are genuinely different problems). Wrapping costs one verb and preserves everything.

## Contract (what the tests enforce)

```go
var ErrPortRange = errors.New("port out of range")

func ParsePort(s string) (int, error)
```

- **Success:** a decimal integer in **1–65535 inclusive** → `(port, nil)`.
- **Not a number** → an error wrapping `strconv`'s: `fmt.Errorf("parsing port %q: %w", s, err)`.
  - `errors.Is(err, strconv.ErrSyntax)` must be **true** — that's the proof you wrapped rather than replaced.
  - The empty string goes down this path too (`Atoi("")` is a syntax error).
- **Out of range** → an error wrapping your sentinel: `fmt.Errorf("port %d: %w", n, ErrPortRange)`, with `errors.Is(err, ErrPortRange)` true. Applies to `0`, negatives, and anything above 65535.
- **Numeric but too large for `int`** (e.g. `"99999999999999999999"`) → `strconv.ErrRange` from Atoi, wrapped. `errors.Is(err, strconv.ErrRange)` is true. Note this is a *different* failure from your `ErrPortRange`, and the tests check they don't collide: a value-too-big-for-int64 must **not** match `ErrPortRange`, and `"70000"` must **not** match `strconv.ErrRange`.
- **On any error the returned int is 0.** Never a half-valid value.
- **Whitespace is rejected**, not trimmed: `" 8080"` and `"8080 "` are syntax errors. Pinned deliberately — silent trimming hides malformed config, and `Atoi` already gives you the strict behavior for free. Write the decision in a comment.
- **`"+80"` is accepted** (Atoi accepts a leading `+`) and yields 80; `"08080"` is accepted as decimal 8080 (no octal interpretation). Both pinned so you don't "fix" what isn't broken.
- Boundaries: `"1"` and `"65535"` succeed; `"0"` and `"65536"` fail with `ErrPortRange`.

## Worked examples

**Example 1 — the three failure kinds, distinguished:**

```go
ParsePort("8080")     // (8080, nil)
ParsePort("abc")      // (0, err) — errors.Is(err, strconv.ErrSyntax)  == true
ParsePort("70000")    // (0, err) — errors.Is(err, ErrPortRange)       == true
ParsePort("9"×20)     // (0, err) — errors.Is(err, strconv.ErrRange)   == true
```

Three different questions a caller might ask, three different answers — all from `errors.Is`, none from string matching.

**Example 2 — why wrapping beats replacing:**

```go
_, err := ParsePort("abc")
err.Error()
// `parsing port "abc": strconv.Atoi: parsing "abc": invalid syntax`
//  ↑ your context        ↑ everything strconv knew, still there

errors.Is(err, strconv.ErrSyntax)   // true — a caller can still classify it
```

If you had written `errors.New("invalid port")`, the message would be tidier and the caller would be blind.

**Example 3 — boundary behavior:**

```go
ParsePort("1")       // (1, nil)         — lowest valid
ParsePort("65535")   // (65535, nil)     — highest valid
ParsePort("0")       // ErrPortRange     — 0 is reserved, not a usable port
ParsePort("-1")      // ErrPortRange     — parses fine, out of range
ParsePort("65536")   // ErrPortRange
```

## Edge cases the tests cover

- Valid: `"1"`, `"80"`, `"8080"`, `"65535"`, `"+80"`, `"08080"`.
- Syntax: `"abc"`, `""`, `" 8080"`, `"8080 "`, `"80.5"`, `"8o8o"`, `"0x1F"`, `"１２３"` (full-width digits).
- Range (yours): `"0"`, `"-1"`, `"-8080"`, `"65536"`, `"999999"`.
- Range (strconv's): a 20-digit number — `strconv.ErrRange` true, `ErrPortRange` false.
- Cross-checks that the two range errors never both match the same input.
- Every error path returning `0` as the int.
- A table asserting `errors.Is` classification for every input, with no `err.Error()` string comparisons for classification (one dedicated test does check that the wrapped message *contains* strconv's text, to prove the wrap preserved it).
