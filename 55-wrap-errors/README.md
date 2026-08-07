# 55 · Wrap with %w, match with errors.Is

## Problem

An error crossing five layers should gain context at each one — *what* was being attempted — without losing the identity of what actually went wrong. `%w` is how Go does that: `fmt.Errorf("loading user %d: %w", id, err)` produces a new error whose message includes both, and which still "is" the original as far as `errors.Is` is concerned. The wrapped error is retained behind an `Unwrap() error` method, forming a chain that `errors.Is` walks.

The single most valuable thing in this drill is the contrast: build the same function with `%v` instead of `%w`, watch the message look *identical*, and watch `errors.Is` return false. One character, and the caller's ability to handle the error correctly is gone.

## Contract (what the tests enforce)

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrPermission   = errors.New("permission denied")
)

func LoadUser(id int) error       // wraps ErrNotFound with %w   — id 0 succeeds (nil)
func LoadUserBadly(id int) error  // wraps ErrNotFound with %v   — the counter-example
func LoadProfile(id int) error    // wraps LoadUser's error again — two layers
func Deny() error                 // wraps ErrPermission with %w
```

- **Sentinels** are package-level `errors.New` values named `ErrXxx`. They're compared by identity, which is exactly why they must be package-level variables and never recreated per call.
- **`LoadUser(id)`** returns `nil` for `id == 0` (the success path), otherwise `fmt.Errorf("loading user %d: %w", id, ErrNotFound)`.
  - Message for id 42: `loading user 42: not found`
  - `errors.Is(err, ErrNotFound)` → true
- **`LoadUserBadly(id)`** is identical but uses `%v`. Message for id 42: `loading user 42: not found` — **byte-for-byte the same** — but `errors.Is(err, ErrNotFound)` → **false**. This function exists only to prove the difference; the tests assert both facts.
- **`LoadProfile(id)`** wraps `LoadUser`'s error: `fmt.Errorf("loading profile for %d: %w", id, err)`.
  - Message for id 42: `loading profile for 42: loading user 42: not found`
  - `errors.Is` still finds `ErrNotFound` through **two** layers — chain depth doesn't matter.
  - Returns nil when `LoadUser` returns nil.
- **`Deny()`** wraps a *different* sentinel, so the tests can verify `errors.Is` discriminates: `errors.Is(Deny(), ErrNotFound)` must be false.
- **`errors.Unwrap`** peels exactly one layer: `errors.Unwrap(LoadProfile(42))` returns an error whose message is `loading user 42: not found`. Unwrapping a non-wrapping error returns nil.
- Message convention: lowercase, no trailing punctuation, context reading left-to-right outermost-first — so a chain reads like a stack trace in prose.

## Worked examples

**Example 1 — the wrap:**

```go
err := LoadUser(42)
err.Error()                       // "loading user 42: not found"
errors.Is(err, ErrNotFound)       // true — identity survived
errors.Unwrap(err) == ErrNotFound // true — one layer down is the sentinel itself
```

**Example 2 — the counter-example that makes the lesson stick:**

```go
good := LoadUser(42)
bad  := LoadUserBadly(42)

good.Error() == bad.Error()        // TRUE — identical text
errors.Is(good, ErrNotFound)       // true
errors.Is(bad,  ErrNotFound)       // FALSE
```

`%v` formats the error into a string and throws the error away. `%w` formats it *and* keeps it. Same output, different capability.

**Example 3 — depth is irrelevant:**

```go
err := LoadProfile(42)
err.Error()                    // "loading profile for 42: loading user 42: not found"
errors.Is(err, ErrNotFound)    // true — errors.Is walks the whole chain
errors.Is(err, ErrPermission)  // false — it walks it correctly
```

## Edge cases the tests cover

- Exact messages for one-layer and two-layer wraps.
- `errors.Is` true through one layer, two layers, and against the sentinel directly.
- `errors.Is` false for a *different* sentinel (discrimination, not blanket true).
- The `%v` version: identical message, `errors.Is` false.
- `errors.Unwrap` peeling one layer at a time down to the sentinel, then returning nil below it.
- Success paths returning nil (`LoadUser(0)`, `LoadProfile(0)`), with `errors.Is(nil, ErrNotFound)` false.
- `errors.Is(err, nil)` false for a real error; `errors.Is(nil, nil)` true.
- Wrapping with different ids producing different messages but the same identity.
- Sentinels not equal to each other, and each equal to itself.
