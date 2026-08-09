package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	// port := ParsePort(os.Getenv("PORT")).OrElse(8080)
	// port
	res := ParsePort(os.Getenv("PORT"))
	if res.IsOk() {
		// ...
	}
	_, err := res.Unwrap()
	if err != nil {
		//
	}

}

func ParsePort(s string) Result[int] {
	n, err := strconv.Atoi(s)
	if err != nil {
		return Err[int](fmt.Errorf("parse port %q: %w", s, err))
	}
	if n < 1 || n > 65535 {
		return Err[int](fmt.Errorf("port %d out of range", n))
	}
	return Ok(n)
}

type Result[T any] struct {
	val T
	err error
}

func Ok[T any](v T) Result[T] {
	return Result[T]{val: v}

}
func Err[T any](e error) Result[T] {
	return Result[T]{err: e}
}

func (r Result[T]) IsOk() bool {
	if r.err != nil {
		return false
	}
	return true
}
func (r Result[T]) IsErr() bool {
	if r.err != nil {
		return true
	}
	return false
}
func (r Result[T]) Unwrap() (T, error) {
	return r.val, r.err
}
func (r Result[T]) OrElse(fallback T) T {
	if r.err != nil {
		return fallback
	}
	return r.val
}
func (r Result[T]) Must() T {
	if r.err != nil {
		panic(r.err)
	}
	return r.val
} // panics on Err

func MapResult[T, U any](r Result[T], f func(T) U) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Ok(f(r.val))
}
