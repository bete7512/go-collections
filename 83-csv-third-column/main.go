package main

import (
	"encoding/csv"
	"errors"
	"io"
)

func main() {}

func ThirdColumn(r io.Reader) ([]string, error) {
	csvReader := csv.NewReader(r)
	csvReader.FieldsPerRecord = -1
	var result []string

	if _, err := csvReader.Read(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, err
	}

	for {
		record, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			return result, nil
		}
		if err != nil {
			return []string{}, err
		}
		if len(record) > 2 {
			result = append(result, record[2])
		}

	}
}
