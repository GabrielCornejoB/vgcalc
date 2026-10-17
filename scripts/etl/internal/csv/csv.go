package csv

import (
	"encoding/csv"
	"fmt"
	"os"
)

func ReadCSVFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ReadCsvFile(): failed to open csv file: %w", err)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ReadCsvFile(): failed to read records of csv file at path: %w", err)
	}

	return records, nil
}
