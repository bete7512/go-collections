package main

import (
	"reflect"
	"testing"
)

func TestChunk(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		n        int
		expected [][]int
		wantErr  bool
	}{
		{name: "even split", input: []int{1, 2, 3, 4, 5, 6}, n: 3, expected: [][]int{{1, 2, 3}, {4, 5, 6}}, wantErr: false},
		{"ragged last chunk", []int{1, 2, 3, 4, 5, 6, 7}, 3, [][]int{{1, 2, 3}, {4, 5, 6}, {7}}, false},
		{"n larger than len", []int{1, 2}, 5, [][]int{{1, 2}}, false},
		{"n equals len", []int{1, 2, 3}, 3, [][]int{{1, 2, 3}}, false},
		{"n is one", []int{1, 2, 3}, 1, [][]int{{1}, {2}, {3}}, false},
		{"single element", []int{9}, 3, [][]int{{9}}, false},
		{"empty input", []int{}, 3, [][]int{}, false},
		{"nil input", nil, 3, [][]int{}, false},
		{"n zero", []int{1, 2, 3}, 0, nil, true},
		{"n negative", []int{1, 2, 3}, -2, nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Chunk(tc.input, tc.n)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Chunk(%v, %d): want error, got nil", tc.input, tc.n)
				}
				return
			}
			if err != nil {
				t.Fatalf("Chunk(%v, %d): unexpected error: %v", tc.input, tc.n, err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("Chunk(%v, %d) = %v, want %v", tc.input, tc.n, got, tc.expected)
			}
		})
	}
}
