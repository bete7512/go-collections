package main

func main() {}
func Invert(m map[string]string) map[string]string {
	reversed := make(map[string]string, len(m))

	for key, val := range m {
		reversed[val] = key
	}

	return reversed
}
