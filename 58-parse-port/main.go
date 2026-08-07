package main

import (
	"errors"
	"fmt"
	"strconv"
)

func main() {}

var ErrPortRange = errors.New("port out of range")

func ParsePort(s string) (int, error) {
	port, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port %d: %w", port, ErrPortRange)
	}
	return port, nil
}
