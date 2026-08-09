package main

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
)

type point struct{ X, Y int }

var errBoom = errors.New("boom")

func TestOkBasics(t *testing.T) {
	r := Ok(42) // T inferred from the argument

	if !r.IsOk() {
		t.Errorf("Ok(42).IsOk() = false, want true")
	}
	if r.IsErr() {
		t.Errorf("Ok(42).IsErr() = true, want false")
	}

	v, err := r.Unwrap()
	if err != nil || v != 42 {
		t.Errorf("Unwrap() = (%d, %v), want (42, nil)", v, err)
	}
	if got := r.OrElse(0); got != 42 {
		t.Errorf("OrElse(0) = %d, want 42", got)
	}
	if got := r.Must(); got != 42 {
		t.Errorf("Must() = %d, want 42", got)
	}
}

func TestErrBasics(t *testing.T) {
	// Note the explicit [int]: inference cannot derive T from an error.
	r := Err[int](errBoom)

	if r.IsOk() {
		t.Errorf("Err.IsOk() = true, want false")
	}
	if !r.IsErr() {
		t.Errorf("Err.IsErr() = false, want true")
	}

	v, err := r.Unwrap()
	if !errors.Is(err, errBoom) {
		t.Errorf("Unwrap() error = %v, want errBoom", err)
	}
	if v != 0 {
		t.Errorf("Unwrap() value = %d, want the zero value 0", v)
	}
	if got := r.OrElse(-1); got != -1 {
		t.Errorf("OrElse(-1) = %d, want -1", got)
	}
}

func TestZeroValuesForVariousTypes(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		v, err := Err[string](errBoom).Unwrap()
		if v != "" || !errors.Is(err, errBoom) {
			t.Errorf("Unwrap = (%q, %v), want (\"\", errBoom)", v, err)
		}
	})

	t.Run("struct", func(t *testing.T) {
		v, err := Err[point](errBoom).Unwrap()
		if v != (point{}) || !errors.Is(err, errBoom) {
			t.Errorf("Unwrap = (%+v, %v), want ({0 0}, errBoom)", v, err)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		v, err := Err[*point](errBoom).Unwrap()
		if v != nil || !errors.Is(err, errBoom) {
			t.Errorf("Unwrap = (%v, %v), want (nil, errBoom)", v, err)
		}
	})

	t.Run("slice", func(t *testing.T) {
		v, err := Err[[]int](errBoom).Unwrap()
		if v != nil || !errors.Is(err, errBoom) {
			t.Errorf("Unwrap = (%v, %v), want (nil, errBoom)", v, err)
		}
	})
}

func TestOkWithNilValues(t *testing.T) {
	// A nil value is not an error: Ok(nil pointer) is still Ok.
	var p *point
	r := Ok(p)
	if !r.IsOk() {
		t.Errorf("Ok(nil pointer).IsOk() = false, want true")
	}
	if v, err := r.Unwrap(); err != nil || v != nil {
		t.Errorf("Unwrap = (%v, %v), want (nil, nil)", v, err)
	}

	var s []int
	rs := Ok(s)
	if !rs.IsOk() {
		t.Errorf("Ok(nil slice).IsOk() = false, want true")
	}
}

func TestOrElseWithNonZeroFallback(t *testing.T) {
	if got := Err[string](errBoom).OrElse("fallback"); got != "fallback" {
		t.Errorf("OrElse = %q, want %q", got, "fallback")
	}
	// On Ok, the fallback must be ignored even when it differs from the zero value.
	if got := Ok("real").OrElse("fallback"); got != "real" {
		t.Errorf("OrElse on Ok = %q, want %q", got, "real")
	}
}

func TestMustPanicsOnErr(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("Must() on an Err did not panic")
		}
		if !strings.Contains(fmt.Sprint(r), "boom") {
			t.Errorf("panic value %v does not mention the underlying error", r)
		}
	}()

	Err[int](errBoom).Must()
}

func TestZeroValueResultReadsAsOk(t *testing.T) {
	// PINNED WEAKNESS: a Result that no constructor ever touched has a nil
	// error, so it reads as a successful zero value.
	var r Result[int]

	if !r.IsOk() {
		t.Errorf("zero-value Result.IsOk() = false; the pinned semantics are Ok-of-zero")
	}
	v, err := r.Unwrap()
	if err != nil || v != 0 {
		t.Errorf("zero-value Unwrap = (%d, %v), want (0, nil)", v, err)
	}
}

func TestWrappedErrorSurvivesStorage(t *testing.T) {
	wrapped := fmt.Errorf("loading config: %w", errBoom)
	r := Err[int](wrapped)

	_, err := r.Unwrap()
	if !errors.Is(err, errBoom) {
		t.Errorf("errors.Is through a stored wrapped error = false — the chain must be preserved intact")
	}
}

func TestMapResultOnOk(t *testing.T) {
	// Result[int] -> Result[string]: the type changes, which is why MapResult
	// must be a free function (methods cannot add type parameters).
	r := MapResult(Ok(42), strconv.Itoa)

	if !r.IsOk() {
		t.Fatalf("MapResult(Ok).IsOk() = false, want true")
	}
	v, err := r.Unwrap()
	if err != nil || v != "42" {
		t.Errorf("Unwrap = (%q, %v), want (\"42\", nil)", v, err)
	}
}

func TestMapResultOnErrShortCircuits(t *testing.T) {
	called := false
	f := func(x int) string {
		called = true
		return strconv.Itoa(x)
	}

	r := MapResult(Err[int](errBoom), f)

	if called {
		t.Errorf("the transform ran on an Err result; it must be skipped")
	}
	if !r.IsErr() {
		t.Errorf("MapResult(Err).IsErr() = false, want true")
	}
	_, err := r.Unwrap()
	if !errors.Is(err, errBoom) {
		t.Errorf("the original error was not preserved: %v", err)
	}
}

func TestMapResultChaining(t *testing.T) {
	// int -> string -> int
	r := MapResult(MapResult(Ok(7), strconv.Itoa), func(s string) int { return len(s) })

	v, err := r.Unwrap()
	if err != nil || v != 1 {
		t.Errorf("chained Unwrap = (%d, %v), want (1, nil)", v, err)
	}

	// A chain starting from an Err short-circuits all the way through.
	bad := MapResult(MapResult(Err[int](errBoom), strconv.Itoa), func(s string) int { return len(s) })
	if !bad.IsErr() {
		t.Errorf("a chain starting from Err must stay Err")
	}
	if _, err := bad.Unwrap(); !errors.Is(err, errBoom) {
		t.Errorf("error lost through the chain: %v", err)
	}
}

func TestMapResultToCompositeTypes(t *testing.T) {
	r := MapResult(Ok("a,b,c"), func(s string) []string { return strings.Split(s, ",") })

	v, err := r.Unwrap()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(v, []string{"a", "b", "c"}) {
		t.Errorf("value = %v, want [a b c]", v)
	}

	rp := MapResult(Ok(point{1, 2}), func(p point) int { return p.X + p.Y })
	if got := rp.OrElse(-1); got != 3 {
		t.Errorf("struct transform = %d, want 3", got)
	}
}

func TestResultsAreIndependentValues(t *testing.T) {
	a := Ok(1)
	b := a // value copy

	if v, _ := b.Unwrap(); v != 1 {
		t.Errorf("copy Unwrap = %d, want 1", v)
	}
	// Both remain usable and unchanged; no method mutates.
	if v, _ := a.Unwrap(); v != 1 {
		t.Errorf("original Unwrap = %d, want 1", v)
	}
	if !a.IsOk() || !b.IsOk() {
		t.Errorf("IsOk changed after copying")
	}
}
