package main

import (
	"errors"
	"fmt"
	"testing"
)

// Compile-time proof that *NotFoundError satisfies error.
var _ error = (*NotFoundError)(nil)

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"ordinary key", "foo", `key "foo" not found`},
		{"key with spaces", "two words", `key "two words" not found`},
		{"empty key", "", `key "" not found`},
		{"numeric key", "42", `key "42" not found`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := &NotFoundError{Key: tc.key}
			if got := err.Error(); got != tc.expected {
				t.Errorf("Error() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestErrorFormatsWithFmt(t *testing.T) {
	err := &NotFoundError{Key: "foo"}
	want := `key "foo" not found`

	if got := fmt.Sprintf("%v", err); got != want {
		t.Errorf("%%v = %q, want %q", got, want)
	}
	if got := fmt.Sprintf("%s", err); got != want {
		t.Errorf("%%s = %q, want %q", got, want)
	}
}

func TestLookupHit(t *testing.T) {
	m := map[string]string{"a": "1", "b": "2"}

	v, err := Lookup(m, "a")
	if err != nil {
		t.Fatalf("Lookup(existing key) returned error: %v", err)
	}
	if v != "1" {
		t.Errorf("Lookup = %q, want %q", v, "1")
	}
}

func TestLookupEmptyValueIsAHit(t *testing.T) {
	m := map[string]string{"empty": ""}

	v, err := Lookup(m, "empty")
	if err != nil {
		t.Fatalf("key present with empty value must be a HIT, got error: %v", err)
	}
	if v != "" {
		t.Errorf("Lookup = %q, want %q", v, "")
	}
}

func TestLookupMiss(t *testing.T) {
	m := map[string]string{"a": "1"}

	v, err := Lookup(m, "missing")
	if err == nil {
		t.Fatalf("Lookup(absent key) returned nil error")
	}
	if v != "" {
		t.Errorf("value on miss = %q, want empty string", v)
	}

	var nfe *NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("error is not a *NotFoundError: %T", err)
	}
	if nfe.Key != "missing" {
		t.Errorf("recovered Key = %q, want %q — the data must survive", nfe.Key, "missing")
	}
}

func TestLookupNilMap(t *testing.T) {
	v, err := Lookup(nil, "anything")
	if err == nil {
		t.Fatalf("Lookup on nil map returned nil error, want a not-found error")
	}
	if v != "" {
		t.Errorf("value = %q, want empty", v)
	}

	var nfe *NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("error is not a *NotFoundError: %T", err)
	}
}

func TestLookupEmptyStringKey(t *testing.T) {
	m := map[string]string{"": "empty-key-value"}

	v, err := Lookup(m, "")
	if err != nil {
		t.Fatalf("empty string is a legal key; got error: %v", err)
	}
	if v != "empty-key-value" {
		t.Errorf("Lookup(\"\") = %q, want %q", v, "empty-key-value")
	}
}

func TestErrorIdentityVsType(t *testing.T) {
	a := &NotFoundError{Key: "same"}
	b := &NotFoundError{Key: "same"}

	// Same data, different pointers: errors.Is (identity) says no...
	if errors.Is(a, b) {
		t.Errorf("errors.Is(a, b) = true for two distinct pointers with equal data")
	}
	// ...but each IS itself.
	if !errors.Is(a, a) {
		t.Errorf("errors.Is(a, a) = false")
	}
	// ...and errors.As (type) finds either.
	var nfe *NotFoundError
	if !errors.As(error(b), &nfe) || nfe.Key != "same" {
		t.Errorf("errors.As failed to recover b")
	}
}

func TestTypedNilTrap(t *testing.T) {
	// BadReturnType declares a *NotFoundError variable and returns it as an
	// error. On the success path that variable is nil — but the interface
	// holding it is NOT, because its type word is set.
	err := BadReturnType(false)

	if err == nil {
		t.Fatalf("BadReturnType(false) == nil — the trap is not reproduced; " +
			"the function must declare a *NotFoundError variable and return it as error")
	}

	// It prints like nil while not being nil. That is the whole trap.
	if got := fmt.Sprintf("%v", err); got != "<nil>" {
		t.Logf("note: typed-nil error formats as %q", got)
	}

	// The failure path still produces a usable error.
	err = BadReturnType(true)
	if err == nil {
		t.Fatalf("BadReturnType(true) == nil, want an error")
	}
	var nfe *NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("BadReturnType(true) is not a *NotFoundError: %T", err)
	}
	if nfe.Key == "" {
		t.Errorf("recovered Key is empty, want the key the function set")
	}
}

func TestLookupFollowsTheRule(t *testing.T) {
	// Unlike BadReturnType, Lookup must return a literal nil on success:
	// no typed-nil leakage on the happy path.
	m := map[string]string{"k": "v"}
	for i := 0; i < 100; i++ {
		if _, err := Lookup(m, "k"); err != nil {
			t.Fatalf("Lookup on a hit returned non-nil error %v (%T) — typed nil leaking?", err, err)
		}
	}
}
