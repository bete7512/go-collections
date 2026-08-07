# 56 · Extract a typed error with errors.As

## Problem

`errors.Is` answers *"is this that specific error value?"* — good for sentinels, which carry no data. `errors.As` answers a different question: *"is there an error of this type anywhere in the chain, and if so, give it to me."* You need it whenever the error carries fields worth reading — the missing key, the invalid field name, the HTTP status, the retry-after duration.

The mechanics have one sharp edge that catches everyone once: the target argument must be a **pointer to** the type you're searching for. Since you're looking for a `*ValidationError`, you pass a `**ValidationError` — `errors.As(err, &ve)` where `ve` is already a pointer. Get it wrong and it panics at runtime rather than failing to compile, so this drill makes you trigger it deliberately, inside a `recover`.

## Contract (what the tests enforce)

```go
type ValidationError struct {
    Field string
    Rule  string
}
func (e *ValidationError) Error() string   // `field "email" failed rule "required"`

type RateLimitError struct {
    RetryAfter int
}
func (e *RateLimitError) Error() string    // `rate limited, retry after 30s`

var ErrDatabase = errors.New("database unavailable")

func Validate(field, rule string) error    // wraps *ValidationError once with %w
func SaveUser(field, rule string) error    // wraps Validate's error again — two layers
func Throttle(seconds int) error           // wraps *RateLimitError with %w
func Backend() error                       // wraps the ErrDatabase sentinel with %w
func MustPanic(err error)                  // calls errors.As with a NON-pointer target
```

- **Messages** (pinned, `%q` on the strings): `field "email" failed rule "required"` and `rate limited, retry after 30s`.
- **`Validate(field, rule)`** returns `fmt.Errorf("validating input: %w", &ValidationError{...})`.
  Message: `validating input: field "email" failed rule "required"`.
- **`SaveUser`** wraps that again: `fmt.Errorf("saving user: %w", err)` →
  `saving user: validating input: field "email" failed rule "required"`.
  `errors.As` must still extract the `*ValidationError` through **two** layers, with `Field` and `Rule` intact — that data surviving arbitrary wrapping is the entire value proposition.
- **Pointer receivers throughout**, matching #54's discipline: the target searched for is `*ValidationError`, never `ValidationError`. If `Error()` had a value receiver, searching for `*ValidationError` would fail — a real source of "errors.As mysteriously returns false".
- **`Throttle` and `Backend` exist so the tests can prove discrimination:** searching a `Throttle` error for `*ValidationError` must return false, and `Backend`'s chain contains a plain sentinel with no custom type in it at all — `errors.As` for either custom type returns false, while `errors.Is(err, ErrDatabase)` is true. That contrast (`Is` for sentinels, `As` for typed) is the takeaway.
- **`MustPanic(err error)`** calls `errors.As(err, ve)` — passing the `*ValidationError` **without** the address-of. That's the mistake, and `errors.As` panics with "errors: target must be a non-nil pointer to either a type that implements error, or to any interface type". The test calls it inside a `defer recover()` and asserts a panic occurred. Write it as the compiler allows (the parameter is `any`, so it compiles fine — which is exactly why this bites at runtime).
- `errors.As(nil, &target)` returns false, no panic.

## Worked examples

**Example 1 — extraction through two wraps:**

```go
err := SaveUser("email", "required")
err.Error()   // "saving user: validating input: field \"email\" failed rule \"required\""

var ve *ValidationError
if errors.As(err, &ve) {
    ve.Field   // "email"    ← recovered from two layers down
    ve.Rule    // "required"
}
```

**Example 2 — Is vs As, side by side:**

```go
err := Backend()                          // wraps the ErrDatabase sentinel
errors.Is(err, ErrDatabase)               // true  — "is it that value?"
errors.As(err, &ve)                       // false — no *ValidationError in the chain

err = Validate("age", "min")              // wraps a typed error
errors.Is(err, ErrDatabase)               // false
errors.As(err, &ve)                       // true, and ve.Field == "age"
```

**Example 3 — the panic:**

```go
var ve *ValidationError
errors.As(err, ve)    // ← missing &. Compiles (target is `any`). Panics at runtime.
errors.As(err, &ve)   // ← correct: a **ValidationError
```

## Edge cases the tests cover

- Exact `Error()` messages for both custom types, including zero values.
- `errors.As` through one layer and two layers, with `Field`/`Rule` (and `RetryAfter`) verified after extraction.
- `errors.As` against an unwrapped error (no wrapping at all).
- Wrong-type search returning false and leaving the target untouched.
- A chain containing `*RateLimitError` searched for `*ValidationError` → false, and vice versa.
- `Backend()`: `errors.Is(err, ErrDatabase)` true while both `errors.As` searches are false.
- `errors.As(nil, &target)` → false.
- The missing-`&` panic, captured with `recover`.
- `errors.Is` on a typed-error chain against an unrelated sentinel → false.
