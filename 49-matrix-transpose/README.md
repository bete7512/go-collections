# 49 · Matrix transpose

## Problem

Model a matrix and flip it across its diagonal: rows become columns, `result[j][i] = m[i][j]`. The index swap is one line — the actual drill is allocating a 2-D structure correctly in Go (there is no true 2-D array behind `[][]int`, only a slice of independently allocated row slices) and surviving the **non-square** case, which is where every swapped-index bug surfaces: a 2×3 matrix must come back 3×2, so code that mixes up which dimension drives which loop dies immediately.

## Contract (what the tests enforce)

```go
type Matrix [][]int

func (m Matrix) Transpose() Matrix
func (m Matrix) Dims() (rows, cols int)
```

- **`Transpose` returns a NEW matrix** with `result[j][i] == m[i][j]` for every element. An r×c input yields a c×r output.
- **The input must be untouched** — the tests snapshot it deep (every row) and verify after the call. Watch the allocation: `make(Matrix, cols)` gives you a slice of **nil rows**; each row needs its own `make([]int, rows)` before you can index into it. Indexing a nil row panics — this is the 2-D allocation lesson.
- **`Dims`** returns (number of rows, length of the first row). An empty matrix is (0, 0).
- **Rectangularity is a documented precondition.** All rows are assumed the same length; behavior on ragged input is undefined and untested. State the precondition in a doc comment — deciding *and writing down* what you don't handle is as much API design as handling it. (The alternative — flat `[]int` + width, indexed as `row*cols + col` — makes ragged input unrepresentable and is how numerical code actually stores matrices, for cache locality. Worth a comment; the harder variant is a great optional rewrite.)
- Empty matrix transposes to empty. A matrix with zero-length rows (shape r×0) is legal input; its transpose is 0×r — i.e. empty. This asymmetry (r×0 → empty, not 0×r reconstructed) is pinned: you cannot represent zero rows *of known length* in `[][]int`, which is itself a small lesson about the representation.
- Values are irrelevant to the operation — negatives, zeros, duplicates all just move.

## Worked examples

**Example 1 — non-square, the case that matters:**

```
input (2×3):        output (3×2):
  1 2 3               1 4
  4 5 6               2 5
                      3 6
m[0][2] == 3  →  result[2][0] == 3
```

**Example 2 — square:**

```
input (3×3):        output:
  1 2 3               1 4 7
  4 5 6               2 5 8
  7 8 9               3 6 9
Diagonal (1, 5, 9) stays put; everything else mirrors across it.
```

**Example 3 — single row and single column are each other's transpose:**

```
[[1 2 3]]  (1×3)  ⇄  [[1] [2] [3]]  (3×1)
```

**Property:** `m.Transpose().Transpose()` is deeply equal to `m` — the involution test, and the cheapest strong check you have.

## Edge cases the tests cover

- 2×3 and 3×2 (non-square both ways, exact expected matrices).
- Square 3×3; 1×1.
- Single row → column; single column → row.
- Empty matrix; matrix of empty rows (2×0 → empty).
- Negative values and duplicates.
- Input-unmodified deep snapshot on every case.
- Result independence: writing into the result must not alter the input (fresh row allocations, no aliasing).
- Double-transpose equals original (involution), including on a 50×80 matrix with distinct values per cell.
- `Dims` consistency: r×c in, c×r out, checked on every case.
