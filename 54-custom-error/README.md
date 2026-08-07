# 54 · Custom error type

## Problem

`error` is just an interface with one method — `Error() string` — so any type with that method is an error. The reason to define your own instead of reaching for `errors.New` is that a custom type carries **structured data**: the caller can recover *which key* was missing, *which field* failed validation, *which status code* came back, and branch on it programmatically instead of parsing English out of a message.

This drill also stages Go's most notorious gotcha — the **typed nil in an interface** — as an executable test rather than a warning you'll forget.

## Contract (what the tests enforce)

```go
type NotFoundError struct {
    Key string
}

func (e *NotFoundError) Error() string   // `key "foo" not found`

func Lookup(m map[string]string, key string) (string, error)

func BadReturnType(fail bool) error      // demonstrates the typed-nil trap
```

- **`Error()` returns exactly** `key "foo" not found` — the key rendered with `%q` (so it's quoted, and an empty key shows as `""`). Pinned so tests can be literal.
- **Pointer receiver on `Error()`**, and always construct it as `&NotFoundError{...}`. Consistency matters: with a pointer receiver only `*NotFoundError` satisfies `error`, and `errors.As` in #56 must be told the same form. Mixing value and pointer receivers across a codebase is how `errors.As` starts mysteriously returning false.
- **`Lookup`** returns the value and nil when the key is present (including when the stored value is `""` — a key mapped to the empty string is a *hit*, distinguished with the `v, ok := m[k]` form). On a miss it returns `("", &NotFoundError{Key: key})`. A nil map is a legal empty map — miss, not panic.
- **`BadReturnType(fail bool) error` is the trap, and it must behave "wrongly" on purpose:**
  ```go
  func BadReturnType(fail bool) error {
      var e *NotFoundError          // a nil *NotFoundError
      if fail {
          e = &NotFoundError{Key: "x"}
      }
      return e                      // returning the CONCRETE type as error
  }
  ```
  With `fail == false` this returns a non-nil `error` interface holding a nil `*NotFoundError` — so `BadReturnType(false) != nil` is **true**, and the caller's `if err != nil` fires on a success path. The tests assert exactly that, because encoding the bug is how you stop writing it. An interface value is nil only when *both* its type and value are nil; here the type word holds `*NotFoundError`.
  Then note in a comment the rule that prevents it: **declare the return type as `error`, and return a literal `nil`** — never a typed nil variable. `Lookup` follows the rule; `BadReturnType` exists to show the alternative.
- Compile-time assertion required in your file: `var _ error = (*NotFoundError)(nil)`.
- The error message convention: lowercase, no trailing punctuation — it will be embedded in longer chains in #55.

## Worked examples

**Example 1 — the data survives:**

```go
_, err := Lookup(map[string]string{"a": "1"}, "missing")
err.Error()      // → `key "missing" not found`

var nfe *NotFoundError
if errors.As(err, &nfe) {
    nfe.Key      // → "missing"   ← the whole reason for a custom type
}
```

**Example 2 — empty value is a hit, not a miss:**

```go
m := map[string]string{"empty": ""}
v, err := Lookup(m, "empty")     // → ("", nil)   — found, value happens to be ""
v, err = Lookup(m, "absent")     // → ("", &NotFoundError{"absent"})
```

Both return `""`. Only the error distinguishes them, which is why `v, ok := m[k]` exists.

**Example 3 — the trap, in full:**

```go
err := BadReturnType(false)
err == nil        // → FALSE, even though nothing failed
fmt.Println(err)  // → <nil>   ← prints like nil, isn't nil
```

## Edge cases the tests cover

- `Error()` message for ordinary keys, keys with spaces, and the empty key (`key "" not found`).
- Compile-time `var _ error = (*NotFoundError)(nil)`.
- `Lookup` hit, miss, hit-with-empty-value, empty-string key, and nil map.
- The returned error's concrete type recovered via `errors.As`, with `.Key` matching the requested key.
- `errors.Is(err, err)` — a custom error is its own sentinel by pointer identity.
- Two `&NotFoundError{Key: "same"}` values being *different* errors (pointer comparison), which is why `errors.As` (type) beats `==` (identity) for typed errors.
- `BadReturnType(false) != nil` — the typed-nil trap, asserted.
- `BadReturnType(true)` producing a usable error whose `.Key` is readable.
- `fmt.Sprintf("%v", err)` and `%s` both rendering `Error()`.
