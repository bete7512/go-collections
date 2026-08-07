package main

import (
	"errors"
	"testing"
)

var (
	_ error = (*ValidationError)(nil)
	_ error = (*RateLimitError)(nil)
)

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "validation error",
			err:      &ValidationError{Field: "email", Rule: "required"},
			expected: `field "email" failed rule "required"`,
		},
		{
			name:     "validation error zero value",
			err:      &ValidationError{},
			expected: `field "" failed rule ""`,
		},
		{
			name:     "rate limit error",
			err:      &RateLimitError{RetryAfter: 30},
			expected: "rate limited, retry after 30s",
		},
		{
			name:     "rate limit zero",
			err:      &RateLimitError{},
			expected: "rate limited, retry after 0s",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.expected {
				t.Errorf("Error() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestExtractThroughOneLayer(t *testing.T) {
	err := Validate("email", "required")
	if err == nil {
		t.Fatalf("Validate returned nil")
	}

	want := `validating input: field "email" failed rule "required"`
	if got := err.Error(); got != want {
		t.Errorf("message = %q, want %q", got, want)
	}

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("errors.As failed to find *ValidationError in %v (%T)", err, err)
	}
	if ve.Field != "email" || ve.Rule != "required" {
		t.Errorf("extracted {Field:%q Rule:%q}, want {email required}", ve.Field, ve.Rule)
	}
}

func TestExtractThroughTwoLayers(t *testing.T) {
	err := SaveUser("age", "min")

	want := `saving user: validating input: field "age" failed rule "min"`
	if got := err.Error(); got != want {
		t.Errorf("message = %q, want %q", got, want)
	}

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("errors.As failed through two layers: %v", err)
	}
	if ve.Field != "age" || ve.Rule != "min" {
		t.Errorf("extracted {Field:%q Rule:%q}, want {age min} — data must survive arbitrary wrapping",
			ve.Field, ve.Rule)
	}
}

func TestExtractFromUnwrappedError(t *testing.T) {
	// errors.As works on a bare typed error too, not only on chains.
	var err error = &ValidationError{Field: "name", Rule: "maxlen"}

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("errors.As failed on an unwrapped typed error")
	}
	if ve.Field != "name" {
		t.Errorf("Field = %q, want %q", ve.Field, "name")
	}
}

func TestExtractRateLimit(t *testing.T) {
	err := Throttle(30)

	var rl *RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("errors.As failed to find *RateLimitError in %v", err)
	}
	if rl.RetryAfter != 30 {
		t.Errorf("RetryAfter = %d, want 30", rl.RetryAfter)
	}
}

func TestWrongTypeReturnsFalse(t *testing.T) {
	err := Throttle(5)

	var ve *ValidationError
	if errors.As(err, &ve) {
		t.Errorf("errors.As found a *ValidationError in a rate-limit chain")
	}
	if ve != nil {
		t.Errorf("target was written on a failed As: %v", ve)
	}

	// And the reverse direction.
	verr := Validate("email", "required")
	var rl *RateLimitError
	if errors.As(verr, &rl) {
		t.Errorf("errors.As found a *RateLimitError in a validation chain")
	}
}

func TestIsForSentinelsAsForTypes(t *testing.T) {
	err := Backend()

	// Is finds the sentinel...
	if !errors.Is(err, ErrDatabase) {
		t.Errorf("errors.Is(Backend(), ErrDatabase) = false, want true")
	}
	// ...but there is no custom type in this chain at all.
	var ve *ValidationError
	if errors.As(err, &ve) {
		t.Errorf("errors.As found *ValidationError in a sentinel-only chain")
	}
	var rl *RateLimitError
	if errors.As(err, &rl) {
		t.Errorf("errors.As found *RateLimitError in a sentinel-only chain")
	}

	// Conversely, a typed chain does not match an unrelated sentinel.
	verr := Validate("email", "required")
	if errors.Is(verr, ErrDatabase) {
		t.Errorf("errors.Is(Validate(...), ErrDatabase) = true, want false")
	}
}

func TestAsOnNilError(t *testing.T) {
	var ve *ValidationError
	if errors.As(nil, &ve) {
		t.Errorf("errors.As(nil, &target) = true, want false")
	}
}

func TestMissingAmpersandPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("MustPanic did not panic — it must call errors.As with a non-pointer target " +
				"(the missing & mistake), which panics at runtime")
		}
	}()

	MustPanic(Validate("email", "required"))
}
