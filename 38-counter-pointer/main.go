package main

func main() {}

type Counter struct{ n int }

func (c *Counter) Inc() {
	c.n++
} // pointer receiver — the mutation persists
func (c Counter) IncBroken() {
	c.n++
} // value receiver — mutates a copy, deliberately kept
func (c Counter) Value() int {
	return c.n
}
