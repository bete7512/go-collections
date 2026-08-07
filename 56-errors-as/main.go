package main

import (
	"errors"
	"fmt"
)

func main() {
	// MustPanic(errors.New("...."))
	// MustPanic(Validate("email", "required"))
}

type ValidationError struct {
	Field string
	Rule  string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("field %q failed rule %q", e.Field, e.Rule)
} // `field "email" failed rule "required"`

type RateLimitError struct {
	RetryAfter int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after %ds", e.RetryAfter)
} // `rate limited, retry after 30s`

var ErrDatabase = errors.New("database unavailable")

func Validate(field, rule string) error {
	return fmt.Errorf("validating input: %w", &ValidationError{Field: field, Rule: rule})
} // wraps *ValidationError once with %w
func SaveUser(field, rule string) error {
	err := Validate(field, rule)
	return fmt.Errorf("saving user: %w", err)
} // wraps Validate's error again — two layers
func Throttle(seconds int) error {
	return fmt.Errorf("throttling request: %w", &RateLimitError{RetryAfter: seconds})

} // wraps *RateLimitError with %w
func Backend() error {
	return fmt.Errorf("backend call failed: %w", ErrDatabase)
} // wraps the ErrDatabase sentinel with %w
func MustPanic(err error) {
	ve := &ValidationError{}
	_ = errors.As(err, any(ve))
} // calls errors.As with a NON-pointer target
