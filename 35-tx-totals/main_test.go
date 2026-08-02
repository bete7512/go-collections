package main

import (
	"fmt"
	"math"
	"slices"
	"testing"
)

const tolerance = 1e-9

// requireTotals asserts got and want hold the same user set and that every
// total matches within the tolerance. Exact float equality is deliberately
// not used — see the README.
func requireTotals(t *testing.T, got, want map[string]float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d users %v, want %d users %v", len(got), got, len(want), want)
	}
	for user, wantTotal := range want {
		gotTotal, ok := got[user]
		if !ok {
			t.Fatalf("user %q missing from result %v", user, got)
		}
		if math.Abs(gotTotal-wantTotal) > tolerance {
			t.Errorf("user %q total = %v, want %v (±%v)", user, gotTotal, wantTotal, tolerance)
		}
	}
}

func TestTotals(t *testing.T) {
	tests := []struct {
		name     string
		input    []Tx
		expected map[string]float64
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: map[string]float64{},
		},
		{
			name:     "empty input",
			input:    []Tx{},
			expected: map[string]float64{},
		},
		{
			name:     "single transaction",
			input:    []Tx{{"u1", 42.5}},
			expected: map[string]float64{"u1": 42.5},
		},
		{
			name:     "basic aggregation",
			input:    []Tx{{"u1", 10}, {"u2", 5}, {"u1", 2.5}},
			expected: map[string]float64{"u1": 12.5, "u2": 5},
		},
		{
			name:     "single user many transactions",
			input:    []Tx{{"u1", 1}, {"u1", 2}, {"u1", 3}, {"u1", 4}},
			expected: map[string]float64{"u1": 10},
		},
		{
			name:     "refunds subtract",
			input:    []Tx{{"u1", 100}, {"u1", -30}, {"u1", -20.5}},
			expected: map[string]float64{"u1": 49.5},
		},
		{
			name:     "net zero user still present",
			input:    []Tx{{"u1", 100}, {"u1", -30}, {"u2", 20}, {"u2", -20}},
			expected: map[string]float64{"u1": 70, "u2": 0},
		},
		{
			name:     "zero amount transaction creates the user",
			input:    []Tx{{"u1", 0}},
			expected: map[string]float64{"u1": 0},
		},
		{
			name:     "all negative",
			input:    []Tx{{"u1", -5}, {"u1", -7.25}},
			expected: map[string]float64{"u1": -12.25},
		},
		{
			name:     "empty string user id is a real key",
			input:    []Tx{{"", 5}, {"u1", 1}, {"", 2.5}},
			expected: map[string]float64{"": 7.5, "u1": 1},
		},
		{
			name: "interleaved users",
			input: []Tx{
				{"a", 1}, {"b", 10}, {"c", 100},
				{"a", 2}, {"b", 20}, {"c", 200},
				{"a", 3}, {"b", 30}, {"c", 300},
			},
			expected: map[string]float64{"a": 6, "b": 60, "c": 600},
		},
		{
			name: "float accumulation within tolerance",
			input: []Tx{
				{"u1", 0.1}, {"u1", 0.1}, {"u1", 0.1}, {"u1", 0.1}, {"u1", 0.1},
				{"u1", 0.1}, {"u1", 0.1}, {"u1", 0.1}, {"u1", 0.1}, {"u1", 0.1},
			},
			expected: map[string]float64{"u1": 1.0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := slices.Clone(tc.input)

			got := Totals(tc.input)

			requireTotals(t, got, tc.expected)

			if !slices.Equal(tc.input, snapshot) {
				t.Errorf("input was modified: had %v, now %v", snapshot, tc.input)
			}
		})
	}
}

func TestTotalsLargeInput(t *testing.T) {
	const users = 100
	const perUser = 100

	var txs []Tx
	for round := 0; round < perUser; round++ {
		for u := 0; u < users; u++ {
			txs = append(txs, Tx{
				UserID: fmt.Sprintf("user-%02d", u),
				Amount: float64(u) + 0.25,
			})
		}
	}

	got := Totals(txs)

	if len(got) != users {
		t.Fatalf("got %d users, want %d", len(got), users)
	}
	for u := 0; u < users; u++ {
		id := fmt.Sprintf("user-%02d", u)
		want := (float64(u) + 0.25) * perUser
		if math.Abs(got[id]-want) > tolerance {
			t.Errorf("user %s total = %v, want %v", id, got[id], want)
		}
	}
}
