package main

func main() {}

type Matrix [][]int

func (m Matrix) Transpose() Matrix {
	rows, cols := m.Dims()
	transposed := make(Matrix, cols)
	for k := range transposed {
		transposed[k] = make([]int, rows)
	}
	for i := 0; i < rows; i = i + 1 {
		for j := 0; j < cols; j = j + 1 {
			transposed[j][i] = m[i][j]
		}
	}
	return transposed
}
func (m Matrix) Dims() (rows, cols int) {
	if len(m) == 0 {
		return 0, 0
	}

	return len(m), len(m[0])
}
