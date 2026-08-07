package main

import (
	"errors"
)

func main() {}

var (
	ErrNotFound = errors.New("key not found")
	ErrEmptyKey = errors.New("key must not be empty")
	ErrNilStore = errors.New("store is nil")
)

func Lookup(m map[string]string, key string) (string, error) {
	var err error
	if key == "" {
		err = ErrEmptyKey
		return "", err
	}
	if m == nil {
		err = ErrNilStore
		return "", err
	}
	if val, ok := m[key]; ok {
		return val, nil
	}
	err = errors.Join(err, ErrNotFound)
	return "", err
}
func LookupAll(m map[string]string, keys []string) ([]string, error) {
	var err error
	collected := []string{}
	if len(keys) == 0 || keys == nil {
		err = ErrEmptyKey
		return collected, nil
	}
	if m == nil {
		err = ErrNilStore
		return collected, err
	}
	for _, key := range keys {
		if key == "" {
			err = ErrEmptyKey
			return nil, err
		}
		if val, ok := m[key]; !ok {
			return nil, ErrNotFound
		} else {
			collected = append(collected, val)
		}
	}
	return collected, nil
}
