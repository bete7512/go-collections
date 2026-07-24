package main

func main() {}
func Split(s, sep string) []string {
	srunes := []rune(s)
	seprunes := []rune(sep)

	var splitteds []string

	jumperIndex := 0
	for i := 0; i < len(srunes); i++ {
		if sep == "" {
			splitteds = append(splitteds, string(srunes[i]))
			continue
		}
		if seprunes[0] != srunes[i] {
			continue
		}
		if sep == string(srunes[i:i+len(seprunes)]) {
			splitteds = append(splitteds, string(srunes[jumperIndex:i]))
			jumperIndex = i + len(seprunes)
		}
	}
	if sep == "" {
		return splitteds
	}

	splitteds = append(splitteds, string(srunes[jumperIndex:]))
	return splitteds
}
