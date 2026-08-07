package main

import (
	"errors"
	"fmt"
)

func main() {}

var (
	ErrNotFound   = errors.New("not found")
	ErrPermission = errors.New("permission denied")
)

func LoadUser(id int) error {
	if id == 0 {
		return nil
	}
	return fmt.Errorf("loading user %d: %w", id, ErrNotFound)
} // wraps ErrNotFound with %w   — id 0 succeeds (nil)
func LoadUserBadly(id int) error {
	if id == 0 {
		return nil
	}
	return fmt.Errorf("loading user %d: %v", id, ErrNotFound)
} // wraps ErrNotFound with %v   — the counter-example
func LoadProfile(id int) error {
	err := LoadUser(id)
	if err == nil {
		return nil
	}
	return fmt.Errorf("loading profile for %d: %w", id, err)
} // wraps LoadUser's error again — two layers
func Deny() error {
	return ErrPermission
} // wraps ErrPermission with %w
