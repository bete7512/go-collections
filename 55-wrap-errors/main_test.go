package main

import (
	"errors"
	"testing"
)

func TestSentinelsAreDistinct(t *testing.T) {
	if errors.Is(ErrNotFound, ErrPermission) {
		t.Errorf("ErrNotFound and ErrPermission compare equal — sentinels must be distinct values")
	}
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Errorf("ErrNotFound is not itself")
	}
	if ErrNotFound.Error() != "not found" {
		t.Errorf("ErrNotFound.Error() = %q, want %q", ErrNotFound.Error(), "not found")
	}
	if ErrPermission.Error() != "permission denied" {
		t.Errorf("ErrPermission.Error() = %q, want %q", ErrPermission.Error(), "permission denied")
	}
}

func TestLoadUserSuccess(t *testing.T) {
	if err := LoadUser(0); err != nil {
		t.Errorf("LoadUser(0) = %v, want nil", err)
	}
	if errors.Is(LoadUser(0), ErrNotFound) {
		t.Errorf("errors.Is(nil, ErrNotFound) = true, want false")
	}
}

func TestLoadUserWrapsWithW(t *testing.T) {
	err := LoadUser(42)
	if err == nil {
		t.Fatalf("LoadUser(42) = nil, want an error")
	}

	if got, want := err.Error(), "loading user 42: not found"; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false for %v — did you use %%v instead of %%w?", err)
	}
	if errors.Is(err, ErrPermission) {
		t.Errorf("errors.Is(err, ErrPermission) = true, want false")
	}
}

func TestUnwrapPeelsOneLayer(t *testing.T) {
	err := LoadUser(42)

	inner := errors.Unwrap(err)
	if inner == nil {
		t.Fatalf("errors.Unwrap(LoadUser(42)) = nil — the error does not wrap anything")
	}
	if inner != ErrNotFound {
		t.Errorf("unwrapped = %v (%T), want the ErrNotFound sentinel itself", inner, inner)
	}
	if errors.Unwrap(ErrNotFound) != nil {
		t.Errorf("errors.Unwrap(sentinel) != nil — a plain error wraps nothing")
	}
}

func TestPercentVLosesIdentity(t *testing.T) {
	good := LoadUser(42)
	bad := LoadUserBadly(42)

	if bad == nil {
		t.Fatalf("LoadUserBadly(42) = nil, want an error")
	}

	// The messages are indistinguishable...
	if good.Error() != bad.Error() {
		t.Errorf("messages differ: %%w version %q vs %%v version %q — they should be identical text",
			good.Error(), bad.Error())
	}
	// ...but only the %w version keeps the identity.
	if !errors.Is(good, ErrNotFound) {
		t.Errorf("errors.Is on the %%w version = false, want true")
	}
	if errors.Is(bad, ErrNotFound) {
		t.Errorf("errors.Is on the %%v version = true, want false — %%v discards the wrapped error")
	}
	if errors.Unwrap(bad) != nil {
		t.Errorf("errors.Unwrap on the %%v version returned non-nil, want nil")
	}
}

func TestTwoLayerWrap(t *testing.T) {
	err := LoadProfile(42)
	if err == nil {
		t.Fatalf("LoadProfile(42) = nil, want an error")
	}

	if got, want := err.Error(), "loading profile for 42: loading user 42: not found"; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is through two layers = false, want true — chain depth must not matter")
	}
	if errors.Is(err, ErrPermission) {
		t.Errorf("errors.Is(err, ErrPermission) = true, want false")
	}
}

func TestUnwrapChainStepByStep(t *testing.T) {
	err := LoadProfile(42)

	one := errors.Unwrap(err)
	if one == nil {
		t.Fatalf("first Unwrap = nil")
	}
	if got, want := one.Error(), "loading user 42: not found"; got != want {
		t.Errorf("after one Unwrap: %q, want %q", got, want)
	}

	two := errors.Unwrap(one)
	if two != ErrNotFound {
		t.Fatalf("after two Unwraps: %v, want the ErrNotFound sentinel", two)
	}

	if three := errors.Unwrap(two); three != nil {
		t.Errorf("unwrapping the sentinel = %v, want nil", three)
	}
}

func TestLoadProfileSuccess(t *testing.T) {
	if err := LoadProfile(0); err != nil {
		t.Errorf("LoadProfile(0) = %v, want nil", err)
	}
}

func TestDenyUsesItsOwnSentinel(t *testing.T) {
	err := Deny()
	if err == nil {
		t.Fatalf("Deny() = nil, want an error")
	}
	if !errors.Is(err, ErrPermission) {
		t.Errorf("errors.Is(Deny(), ErrPermission) = false, want true")
	}
	if errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(Deny(), ErrNotFound) = true, want false — Is must discriminate")
	}
}

func TestDifferentIdsSameIdentity(t *testing.T) {
	a := LoadUser(1)
	b := LoadUser(999)

	if a.Error() == b.Error() {
		t.Errorf("different ids produced identical messages: %q", a.Error())
	}
	if !errors.Is(a, ErrNotFound) || !errors.Is(b, ErrNotFound) {
		t.Errorf("both errors must still be ErrNotFound regardless of id")
	}
}

func TestIsWithNil(t *testing.T) {
	if errors.Is(LoadUser(42), nil) {
		t.Errorf("errors.Is(realError, nil) = true, want false")
	}
	if !errors.Is(nil, nil) {
		t.Errorf("errors.Is(nil, nil) = false, want true")
	}
}
