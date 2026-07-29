package main

import (
	"reflect"
	"testing"
)

// func Invert(m map[string]string) map[string]string
func TestInvert(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected map[string]string
	}{
		{
			name: "simple invert",
			input: map[string]string{
				"a": "1",
				"b": "2",
			},
			expected: map[string]string{
				"1": "a",
				"2": "b",
			},
		},
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: map[string]string{},
		},
		{
			name:     "nil map",
			input:    nil,
			expected: map[string]string{},
		},
		{
			name: "empty string key",
			input: map[string]string{
				"": "value",
			},
			expected: map[string]string{
				"value": "",
			},
		},
		{
			name: "empty string value",
			input: map[string]string{
				"key": "",
			},
			expected: map[string]string{
				"": "key",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Invert(tc.input)

			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("expected %#v, got %#v", tc.expected, got)
			}
		})
	}
}

func TestInvertDoesNotModifyInput(t *testing.T) {
	input := map[string]string{
		"a": "1",
		"b": "2",
	}

	original := map[string]string{
		"a": "1",
		"b": "2",
	}

	Invert(input)

	if !reflect.DeepEqual(input, original) {
		t.Fatalf("input map was modified: got %#v", input)
	}
}

func TestInvertCollision(t *testing.T) {
	input := map[string]string{
		"a": "x",
		"b": "x",
	}

	for i := 0; i < 50; i++ {
		got := Invert(input)

		// Collision means only one entry can exist.
		if len(got) != 1 {
			t.Fatalf("expected one entry, got %#v", got)
		}

		// Since winner is unspecified, only validate allowed keys.
		value, ok := got["x"]
		if !ok {
			t.Fatalf("expected key 'x' in result, got %#v", got)
		}

		if value != "a" && value != "b" {
			t.Fatalf("unexpected collision winner %q", value)
		}
	}
}
