package csv

import (
	"encoding/csv"
	"fmt"
	"os"
)

func ReadCSVFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ReadCSVFile(): failed to open csv file: %w", err)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ReadCSVFile(): failed to read records of csv file at path: %w", err)
	}
	if len(records) <= 1 {
		return nil, fmt.Errorf("ReadCSVFile(): file doesn't has enough lines. File path: %v", filePath)
	}

	records = records[1:]

	return records, nil
}
