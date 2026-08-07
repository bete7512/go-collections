# 57 · Sentinel errors from a lookup

## Problem

A function that can fail in several distinct ways owes its caller a way to tell those ways apart — *without parsing English*. Sentinel errors are the simplest mechanism: package-level `errors.New` values, exported, compared with `errors.Is`. `io.EOF`, `sql.ErrNoRows`, and `os.ErrNotExist` are the stdlib's versions, and callers branch on them every day.

The constraint that makes this drill bite: **no test may look at `err.Error()`**. That's not an arbitrary rule — it's the reason sentinels exist. Message text is not API. The moment a caller writes `if strings.Contains(err.Error(), "not found")`, adding context with `%w` upstream breaks them. Sentinels let the message change freely while identity stays stable.

## Contract (what the tests enforce)

```go
var (
    ErrNotFound   = errors.New("key not found")
    ErrEmptyKey   = errors.New("key must not be empty")
    ErrNilStore   = errors.New("store is nil")
)

func Lookup(m map[string]string, key string) (string, error)
func LookupAll(m map[string]string, keys []string) ([]string, error)
```

- **Validation order is pinned** (the tests depend on it): empty key is checked **first**, then nil map, then presence.
  - `key == ""` → `("", ErrEmptyKey)` — even if the map is nil, even if `""` is a stored key.
  - `m == nil` → `("", ErrNilStore)` — a nil map is a caller mistake worth distinguishing from an ordinary miss.
  - key absent → `("", err)` where `errors.Is(err, ErrNotFound)` is true. Wrap it with the key for context: `fmt.Errorf("looking up %q: %w", key, ErrNotFound)` — the message gains detail, `errors.Is` still matches, and no caller has to care.
  - key present → `(value, nil)`.
- **A key mapped to `""` is a HIT.** `Lookup(map[string]string{"k": ""}, "k")` returns `("", nil)`. Only `v, ok := m[k]` can tell that apart from a miss — this is #54's lesson made load-bearing again, and there's a dedicated test.
- **`LookupAll`** returns the values for all keys in order, or fails on the first problem — returning `nil` values and an error that wraps the same sentinels, so `errors.Is` works on it too. An empty `keys` slice succeeds with an empty result. This exists to prove your sentinels survive a second layer of wrapping.
- **Sentinels must be distinct values.** Three separate `errors.New` calls; the tests verify no two match each other. (`errors.New` returns a pointer, so even two calls with identical text are different errors — worth knowing why.)
- Non-nil but **empty** map is not a nil map: `Lookup(map[string]string{}, "a")` → `ErrNotFound`, not `ErrNilStore`.

## Worked examples

**Example 1 — the four outcomes:**

```go
m := map[string]string{"a": "1", "blank": ""}

Lookup(m, "a")       // ("1", nil)
Lookup(m, "blank")   // ("",  nil)          ← hit; the value just happens to be empty
Lookup(m, "zzz")     // ("",  wrapped ErrNotFound)
Lookup(m, "")        // ("",  ErrEmptyKey)
Lookup(nil, "a")     // ("",  ErrNilStore)
```

**Example 2 — how a caller branches:**

```go
v, err := Lookup(m, key)
switch {
case errors.Is(err, ErrEmptyKey):
    // 400 — client sent nonsense
case errors.Is(err, ErrNotFound):
    // 404 — legitimate miss
case err != nil:
    // 500 — something unexpected
default:
    use(v)
}
```

No string matching anywhere. Add context upstream tomorrow and this switch keeps working.

**Example 3 — context without identity loss:**

```go
_, err := Lookup(m, "zzz")
// err.Error() is `looking up "zzz": key not found`  — useful in a log
errors.Is(err, ErrNotFound)   // still true         — useful in code
```

## Edge cases the tests cover

- All five outcomes above, checked exclusively with `errors.Is`.
- Empty key taking precedence over a nil map (order pinned).
- Empty key even when `""` is genuinely a key in the map.
- Empty (non-nil) map → `ErrNotFound`, not `ErrNilStore`.
- Key present with empty value → hit, nil error.
- Keys with spaces, unicode keys, and long keys.
- The three sentinels being mutually distinct and each matching itself.
- Two `errors.New("same text")` values not matching — identity, not text.
- `LookupAll` success with several keys, in order; empty key list; first-failure propagation for each sentinel kind.
- `errors.Is` on the `LookupAll` error finding the sentinel through both wrapping layers.
