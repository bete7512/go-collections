package main

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestParsePortValid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"lowest valid port", "1", 1},
		{"well-known port", "80", 80},
		{"common dev port", "8080", 8080},
		{"highest valid port", "65535", 65535},
		{"explicit plus sign", "+80", 80},
		{"leading zeros are decimal", "08080", 8080},
		{"all zeros prefix", "00001", 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePort(tc.input)
			if err != nil {
				t.Fatalf("ParsePort(%q) returned error: %v", tc.input, err)
			}
			if got != tc.expected {
				t.Errorf("ParsePort(%q) = %d, want %d", tc.input, got, tc.expected)
			}
		})
	}
}

func TestParsePortSyntaxErrors(t *testing.T) {
	inputs := []struct {
		name  string
		input string
	}{
		{"letters", "abc"},
		{"empty string", ""},
		{"leading space not trimmed", " 8080"},
		{"trailing space not trimmed", "8080 "},
		{"decimal point", "80.5"},
		{"letters mixed in", "8o8o"},
		{"hex notation", "0x1F"},
		{"full-width digits", "１２３"},
		{"just a sign", "+"},
		{"double sign", "++80"},
	}

	for _, tc := range inputs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePort(tc.input)
			if err == nil {
				t.Fatalf("ParsePort(%q) = (%d, nil), want a syntax error", tc.input, got)
			}
			if got != 0 {
				t.Errorf("ParsePort(%q) returned %d with an error, want 0", tc.input, got)
			}
			if !errors.Is(err, strconv.ErrSyntax) {
				t.Errorf("ParsePort(%q): errors.Is(err, strconv.ErrSyntax) = false — wrap the strconv error with %%w, don't replace it",
					tc.input)
			}
			if errors.Is(err, ErrPortRange) {
				t.Errorf("ParsePort(%q) matched ErrPortRange — a syntax error is not a range error", tc.input)
			}
		})
	}
}

func TestParsePortRangeErrors(t *testing.T) {
	inputs := []struct {
		name  string
		input string
	}{
		{"zero is reserved", "0"},
		{"negative one", "-1"},
		{"negative port", "-8080"},
		{"one past the maximum", "65536"},
		{"far above maximum", "999999"},
		{"negative zero", "-0"},
	}

	for _, tc := range inputs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePort(tc.input)
			if err == nil {
				t.Fatalf("ParsePort(%q) = (%d, nil), want a range error", tc.input, got)
			}
			if got != 0 {
				t.Errorf("ParsePort(%q) returned %d with an error, want 0", tc.input, got)
			}
			if !errors.Is(err, ErrPortRange) {
				t.Errorf("ParsePort(%q): errors.Is(err, ErrPortRange) = false", tc.input)
			}
			if errors.Is(err, strconv.ErrSyntax) {
				t.Errorf("ParsePort(%q) matched strconv.ErrSyntax — it parsed fine, it is just out of range", tc.input)
			}
		})
	}
}

func TestBoundaries(t *testing.T) {
	// The four values around the valid interval, in one place.
	if _, err := ParsePort("1"); err != nil {
		t.Errorf("port 1 must be valid, got %v", err)
	}
	if _, err := ParsePort("65535"); err != nil {
		t.Errorf("port 65535 must be valid, got %v", err)
	}
	if _, err := ParsePort("0"); !errors.Is(err, ErrPortRange) {
		t.Errorf("port 0 must be out of range")
	}
	if _, err := ParsePort("65536"); !errors.Is(err, ErrPortRange) {
		t.Errorf("port 65536 must be out of range")
	}
}

func TestStrconvRangeIsNotPortRange(t *testing.T) {
	// A number too large for int: strconv reports ErrRange. That is a
	// DIFFERENT failure from a syntactically fine port outside 1-65535.
	huge := "99999999999999999999"

	got, err := ParsePort(huge)
	if err == nil {
		t.Fatalf("ParsePort(%q) = (%d, nil), want an error", huge, got)
	}
	if got != 0 {
		t.Errorf("returned %d with an error, want 0", got)
	}
	if !errors.Is(err, strconv.ErrRange) {
		t.Errorf("errors.Is(err, strconv.ErrRange) = false — strconv's range error must survive wrapping")
	}
	if errors.Is(err, ErrPortRange) {
		t.Errorf("an int-overflow value also matched ErrPortRange — the two range failures must stay distinct")
	}

	// And the converse: an in-int but out-of-port value is NOT strconv.ErrRange.
	if _, err := ParsePort("70000"); errors.Is(err, strconv.ErrRange) {
		t.Errorf("\"70000\" matched strconv.ErrRange — it fits in an int, so only ErrPortRange applies")
	}
}

func TestWrappedMessagePreservesStrconvDetail(t *testing.T) {
	// The only test that looks at message text, and only to prove the wrap
	// kept strconv's detail rather than discarding it.
	_, err := ParsePort("abc")
	if err == nil {
		t.Fatalf("ParsePort(\"abc\") returned nil error")
	}

	msg := err.Error()
	if !strings.Contains(msg, "abc") {
		t.Errorf("message %q does not mention the offending input", msg)
	}
	if !strings.Contains(msg, "invalid syntax") {
		t.Errorf("message %q lost strconv's detail — %%w should keep the original message in the chain", msg)
	}
}

func TestUnwrapReachesStrconv(t *testing.T) {
	_, err := ParsePort("nope")

	inner := errors.Unwrap(err)
	if inner == nil {
		t.Fatalf("errors.Unwrap returned nil — the error does not wrap anything")
	}
	if !errors.Is(inner, strconv.ErrSyntax) {
		t.Errorf("unwrapped error is not strconv's: %v", inner)
	}
}
