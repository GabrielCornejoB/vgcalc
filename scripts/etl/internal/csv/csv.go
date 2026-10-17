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
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ReadCSVFile(): failed to read records of csv file at path: %v. %w", filePath, err)
	}
	if len(records) <= 1 {
		return nil, fmt.Errorf("ReadCSVFile(): file doesn't has enough lines. File path: %v", filePath)
	}

	records = records[1:]

	return records, nil
}
