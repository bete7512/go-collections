package main

import (
	"slices"
	"testing"
)

func deepClone(m Matrix) Matrix {
	if m == nil {
		return nil
	}
	c := make(Matrix, len(m))
	for i, row := range m {
		c[i] = slices.Clone(row)
	}
	return c
}

func deepEqual(a, b Matrix) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !slices.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

func TestTranspose(t *testing.T) {
	tests := []struct {
		name     string
		input    Matrix
		expected Matrix
	}{
		{
			name:     "empty matrix",
			input:    Matrix{},
			expected: Matrix{},
		},
		{
			name:     "1x1",
			input:    Matrix{{7}},
			expected: Matrix{{7}},
		},
		{
			name:     "2x3 becomes 3x2",
			input:    Matrix{{1, 2, 3}, {4, 5, 6}},
			expected: Matrix{{1, 4}, {2, 5}, {3, 6}},
		},
		{
			name:     "3x2 becomes 2x3",
			input:    Matrix{{1, 4}, {2, 5}, {3, 6}},
			expected: Matrix{{1, 2, 3}, {4, 5, 6}},
		},
		{
			name:     "square 3x3 mirrors across diagonal",
			input:    Matrix{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
			expected: Matrix{{1, 4, 7}, {2, 5, 8}, {3, 6, 9}},
		},
		{
			name:     "single row becomes column",
			input:    Matrix{{1, 2, 3}},
			expected: Matrix{{1}, {2}, {3}},
		},
		{
			name:     "single column becomes row",
			input:    Matrix{{1}, {2}, {3}},
			expected: Matrix{{1, 2, 3}},
		},
		{
			name:     "negatives and duplicates",
			input:    Matrix{{0, -1}, {-1, 0}},
			expected: Matrix{{0, -1}, {-1, 0}},
		},
		{
			name:     "rows of length zero transpose to empty",
			input:    Matrix{{}, {}},
			expected: Matrix{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := deepClone(tc.input)

			got := tc.input.Transpose()

			if !deepEqual(got, tc.expected) {
				t.Errorf("Transpose(%v) = %v, want %v", snapshot, got, tc.expected)
			}
			if !deepEqual(tc.input, snapshot) {
				t.Errorf("input was modified: had %v, now %v", snapshot, tc.input)
			}
		})
	}
}

func TestDims(t *testing.T) {
	tests := []struct {
		name       string
		input      Matrix
		rows, cols int
	}{
		{"empty", Matrix{}, 0, 0},
		{"1x1", Matrix{{5}}, 1, 1},
		{"2x3", Matrix{{1, 2, 3}, {4, 5, 6}}, 2, 3},
		{"3x1", Matrix{{1}, {2}, {3}}, 3, 1},
		{"2x0", Matrix{{}, {}}, 2, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, c := tc.input.Dims()
			if r != tc.rows || c != tc.cols {
				t.Errorf("Dims() = (%d, %d), want (%d, %d)", r, c, tc.rows, tc.cols)
			}
		})
	}
}

func TestTransposeDimsFlip(t *testing.T) {
	m := Matrix{{1, 2, 3, 4}, {5, 6, 7, 8}} // 2×4

	got := m.Transpose()

	r, c := got.Dims()
	if r != 4 || c != 2 {
		t.Errorf("transposed Dims = (%d, %d), want (4, 2)", r, c)
	}
}

func TestResultIsIndependent(t *testing.T) {
	m := Matrix{{1, 2}, {3, 4}}

	got := m.Transpose()
	got[0][0] = 999
	got[1][1] = 888

	if m[0][0] != 1 || m[1][1] != 4 {
		t.Errorf("mutating the result changed the input: %v — rows are aliased, allocate fresh ones", m)
	}
}

func TestDoubleTransposeIsIdentity(t *testing.T) {
	// 50×80 with a distinct value per cell: any index mix-up breaks equality.
	const rows, cols = 50, 80
	m := make(Matrix, rows)
	for i := range m {
		m[i] = make([]int, cols)
		for j := range m[i] {
			m[i][j] = i*cols + j
		}
	}
	snapshot := deepClone(m)

	got := m.Transpose().Transpose()

	if !deepEqual(got, snapshot) {
		t.Fatalf("Transpose twice != original")
	}
	// Spot-check the single transpose too: corner and interior cells.
	tr := m.Transpose()
	if tr[0][0] != m[0][0] || tr[cols-1][rows-1] != m[rows-1][cols-1] || tr[3][7] != m[7][3] {
		t.Errorf("single transpose misplaced cells: tr[3][7]=%d want m[7][3]=%d", tr[3][7], m[7][3])
	}
}
