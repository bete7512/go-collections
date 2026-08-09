# 65 · Generic Result[T]

## Problem

Rust has `Result<T, E>`, Haskell has `Either`, Scala has `Try`. Go has `(T, error)`. This drill builds the Result type in Go — properly, with all the ergonomics — so that the closing task can be to write down, from experience rather than hearsay, **why Go doesn't use it**. An opinion you've earned by building the alternative is worth more than one you inherited.

You'll also meet a genuine wall of Go's type inference: `Err[T]` cannot infer `T`, because an `error` argument says nothing about the success type. Every error construction needs an explicit type parameter. That friction is not a flaw in your design — it's structural, and it's a large part of the answer to the closing question.

## Contract (what the tests enforce)

```go
type Result[T any] struct {
    val T
    err error
}

func Ok[T any](v T) Result[T]
func Err[T any](e error) Result[T]

func (r Result[T]) IsOk() bool
func (r Result[T]) IsErr() bool
func (r Result[T]) Unwrap() (T, error)
func (r Result[T]) OrElse(fallback T) T
func (r Result[T]) Must() T                       // panics on Err
func MapResult[T, U any](r Result[T], f func(T) U) Result[U]
```

- **Unexported fields.** The invariant — exactly one of `val`/`err` is meaningful — must not be violable from outside the package. Constructors are the only way in.
- **`Ok(v)`**: `IsOk` true, `IsErr` false, `Unwrap` returns `(v, nil)`, `OrElse` returns `v`, `Must` returns `v`.
- **`Err[T](e)`**: `IsOk` false, `IsErr` true, `Unwrap` returns `(zero, e)` where zero comes from `var zero T`, `OrElse` returns the fallback, `Must` **panics**.
- **`Err[T]` requires an explicit type parameter** at every call site — `Err[int](someErr)`. Inference cannot derive `T` from an `error`. The tests are written that way; notice the asymmetry with `Ok(42)`, which infers fine.
- **The zero value is pinned as an Ok of the zero value.** `Result[int]{}` has a nil `err`, so `IsOk()` is true and `Unwrap()` returns `(0, nil)`. Document this: it's a real weakness — a variable that was never assigned reads as a successful zero, and no constructor was involved. Plain `(T, error)` never has to answer this question because there's no container to leave uninitialized.
- **`Must` panics on Err** and must include the error in the panic value. Document it as test/init-only — the `regexp.MustCompile` convention.
- **`MapResult`** applies `f` to an Ok's value producing `Result[U]`; an Err passes through unchanged with its error preserved. It's a **free function, not a method** — Go methods cannot introduce new type parameters, so `func (r Result[T]) Map[U any](...)` does not compile. Try it once and read the error: "method must have no type parameters". That restriction is another structural piece of the answer.
- **Value receivers throughout** — a Result is a small immutable value; no method mutates.

## Worked examples

**Example 1 — the happy shapes:**

```go
Ok(42).OrElse(0)                          // 42
Err[int](errors.New("boom")).OrElse(-1)   // -1     ← explicit [int] required
Ok("x").Unwrap()                          // ("x", nil)
Err[string](e).Unwrap()                   // ("", e)  ← zero via `var zero T`
```

**Example 2 — the inference wall:**

```go
Ok(42)                    // fine: T inferred from the argument
Err(errors.New("boom"))   // DOES NOT COMPILE: cannot infer T
Err[int](errors.New("boom"))   // required
```

Every error path in every function needs the type spelled out. In a codebase with hundreds of error returns, that's hundreds of annotations that `return 0, err` never needed.

**Example 3 — chaining, and where it stops:**

```go
r := MapResult(Ok(21), func(x int) int { return x * 2 })   // Ok(42)
r = MapResult(Err[int](e), func(x int) int { return x * 2 }) // Err — f never called
```

Chaining is the whole appeal of Result. But note it must be a free function, and that each step needs its own generic call — no method chaining, which is exactly what makes Result pleasant in other languages.

## Edge cases the tests cover

- Ok and Err paths for every method, across `int`, `string`, a struct, a pointer, and a slice as `T`.
- `Unwrap` on Err returning the correct zero value for each of those types.
- `OrElse` in both directions, including a fallback that differs from the zero value.
- `Must` returning the value on Ok, and panicking on Err — captured with `recover`, with the panic value required to mention the error.
- The zero value `Result[int]{}` reading as Ok — the pinned weakness.
- `Ok` of a nil pointer and of a nil slice: still Ok (the value being nil is not an error).
- `errors.Is` working on the error retrieved from an Err Result — wrapping survives storage.
- `MapResult`: transforms on Ok, passes the error through untouched on Err (verified with `errors.Is`), never calls `f` on an Err, and changes the type (`Result[int]` → `Result[string]`).
- Chained `MapResult` calls, and a chain where the first step is an Err (short-circuits).

## The closing deliverable

Once everything passes, write a comment of 3–5 lines in `main.go` answering: **why does idiomatic Go still prefer `(T, error)`?** Draw on what you just built — the `Err[T]` annotation burden, the meaningless zero value, methods that cannot add type parameters (so no fluent chaining), and the fact that `if err != nil` composes with `errors.Is`/`errors.As`, multiple return values, and every existing API without a wrapper. That comment is the actual output of this challenge; the code is how you earned the right to write it.
