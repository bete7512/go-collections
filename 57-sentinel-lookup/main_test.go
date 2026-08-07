package main

import (
	"errors"
	"slices"
	"testing"
)

// NOTE: no test in this file inspects err.Error(). Identity is the API;
// message text is not. That constraint is the point of the drill.

func TestSentinelsAreDistinct(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrNotFound", ErrNotFound},
		{"ErrEmptyKey", ErrEmptyKey},
		{"ErrNilStore", ErrNilStore},
	}

	for i, a := range sentinels {
		if !errors.Is(a.err, a.err) {
			t.Errorf("%s does not match itself", a.name)
		}
		for j, b := range sentinels {
			if i == j {
				continue
			}
			if errors.Is(a.err, b.err) {
				t.Errorf("%s matches %s — sentinels must be distinct values", a.name, b.name)
			}
		}
	}
}

func TestErrorsNewIdentityNotText(t *testing.T) {
	// Two errors.New calls with identical text are different errors.
	a := errors.New("same text")
	b := errors.New("same text")
	if errors.Is(a, b) {
		t.Errorf("two errors.New values with equal text matched — identity, not text, is what Is compares")
	}
}

func TestLookupHit(t *testing.T) {
	m := map[string]string{"a": "1", "b": "2"}

	v, err := Lookup(m, "a")
	if err != nil {
		t.Fatalf("hit returned error")
	}
	if v != "1" {
		t.Errorf("value = %q, want %q", v, "1")
	}
}

func TestLookupEmptyValueIsAHit(t *testing.T) {
	m := map[string]string{"blank": ""}

	v, err := Lookup(m, "blank")
	if err != nil {
		t.Fatalf("key present with empty value must be a hit, got an error")
	}
	if v != "" {
		t.Errorf("value = %q, want empty string", v)
	}
}

func TestLookupMiss(t *testing.T) {
	m := map[string]string{"a": "1"}

	v, err := Lookup(m, "zzz")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("miss did not match ErrNotFound")
	}
	if v != "" {
		t.Errorf("value on miss = %q, want empty", v)
	}
	// A miss must not be confused with the other failure kinds.
	if errors.Is(err, ErrEmptyKey) || errors.Is(err, ErrNilStore) {
		t.Errorf("miss also matched another sentinel")
	}
}

func TestLookupEmptyKey(t *testing.T) {
	m := map[string]string{"a": "1"}

	_, err := Lookup(m, "")
	if !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("empty key did not match ErrEmptyKey")
	}
	if errors.Is(err, ErrNotFound) {
		t.Errorf("empty key also matched ErrNotFound")
	}
}

func TestEmptyKeyBeatsNilStore(t *testing.T) {
	// Validation order is pinned: empty key is checked first.
	_, err := Lookup(nil, "")
	if !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Lookup(nil, \"\") did not match ErrEmptyKey — empty key must be checked before the nil map")
	}
}

func TestEmptyKeyEvenWhenStored(t *testing.T) {
	// "" is a legal map key, but the empty-key check comes first regardless.
	m := map[string]string{"": "stored-under-empty"}

	_, err := Lookup(m, "")
	if !errors.Is(err, ErrEmptyKey) {
		t.Errorf("empty key must be rejected even when \"\" is present in the map")
	}
}

func TestLookupNilStore(t *testing.T) {
	_, err := Lookup(nil, "a")
	if !errors.Is(err, ErrNilStore) {
		t.Fatalf("nil map did not match ErrNilStore")
	}
	if errors.Is(err, ErrNotFound) {
		t.Errorf("nil map also matched ErrNotFound — these are different failures")
	}
}

func TestEmptyMapIsNotNilStore(t *testing.T) {
	_, err := Lookup(map[string]string{}, "a")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("empty (non-nil) map did not match ErrNotFound")
	}
	if errors.Is(err, ErrNilStore) {
		t.Errorf("empty map matched ErrNilStore — an empty map is not a nil map")
	}
}

func TestUnusualKeys(t *testing.T) {
	m := map[string]string{
		"two words":  "spaces",
		"世界":         "unicode",
		"a-very-long-key-that-keeps-going-and-going": "long",
	}

	for key, want := range m {
		v, err := Lookup(m, key)
		if err != nil {
			t.Errorf("key %q returned an error", key)
			continue
		}
		if v != want {
			t.Errorf("key %q: value = %q, want %q", key, v, want)
		}
	}

	if _, err := Lookup(m, "世"); !errors.Is(err, ErrNotFound) {
		t.Errorf("partial unicode key did not match ErrNotFound")
	}
}

func TestLookupAllSuccess(t *testing.T) {
	m := map[string]string{"a": "1", "b": "2", "c": "3"}

	got, err := LookupAll(m, []string{"c", "a", "b"})
	if err != nil {
		t.Fatalf("LookupAll returned an error for all-present keys")
	}
	if !slices.Equal(got, []string{"3", "1", "2"}) {
		t.Errorf("values = %v, want [3 1 2] in requested order", got)
	}
}

func TestLookupAllEmptyKeyList(t *testing.T) {
	m := map[string]string{"a": "1"}

	got, err := LookupAll(m, nil)
	if err != nil {
		t.Fatalf("LookupAll with no keys returned an error")
	}
	if len(got) != 0 {
		t.Errorf("values = %v, want empty", got)
	}
}

func TestLookupAllPropagatesSentinels(t *testing.T) {
	m := map[string]string{"a": "1"}

	tests := []struct {
		name string
		keys []string
		want error
	}{
		{"missing key", []string{"a", "zzz"}, ErrNotFound},
		{"empty key", []string{"a", ""}, ErrEmptyKey},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := LookupAll(m, tc.keys)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error did not match the expected sentinel through LookupAll's wrapping")
			}
			if got != nil {
				t.Errorf("values on failure = %v, want nil", got)
			}
		})
	}

	if _, err := LookupAll(nil, []string{"a"}); !errors.Is(err, ErrNilStore) {
		t.Errorf("LookupAll(nil, ...) did not match ErrNilStore")
	}
}
