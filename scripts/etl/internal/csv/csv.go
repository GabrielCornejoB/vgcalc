package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/jszwec/csvutil"
)

// unmarshal maps CSV bytes to a struct slice via header tags.
// TrimLeadingSpace stays on so quoted fields preceded by a space (`", "`) still parse.
func unmarshal(data []byte, v any) error {
	r := csv.NewReader(bytes.NewReader(data))
	r.TrimLeadingSpace = true
	dec, err := csvutil.NewDecoder(r)
	if err != nil {
		return err
	}
	return dec.Decode(v)
}

// LoadFile reads a CSV file from path and maps it to a struct slice via header tags.
// I/O + decode live here so strategy builders stay pure (in-memory slices, no os/csvutil).
func LoadFile[T any](path string) ([]T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("csv.LoadFile(%s): %w", path, err)
	}
	var v []T
	if err := unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("csv.LoadFile(%s): %w", path, err)
	}
	return v, nil
}
