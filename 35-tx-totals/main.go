package main

func main() {}

type Tx struct {
	UserID string
	Amount float64
}

func Totals(txs []Tx) map[string]float64 {
	totals := make(map[string]float64)

	for _, tx := range txs {
		totals[tx.UserID] += tx.Amount
	}

	return totals
}
