package main

import "fmt"

func main() {}

type NotFoundError struct {
	Key string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("key \"%s\" not found", e.Key)
} // `key "foo" not found`

func Lookup(m map[string]string, key string) (string, error) {
	val, ok := m[key]
	if !ok {
		return val, &NotFoundError{Key: key}
	}
	return val, nil
}

func BadReturnType(fail bool) error {
	var e *NotFoundError // a nil *NotFoundError
	if fail {
		e = &NotFoundError{Key: "x"}
	}
	return e
}
