package utils

import (
	"os"
	"encoding/csv"
	"fmt"
)

func ReadCsvFile(filePath string) ([][]string, error){
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ReadCsvFile(): Failed to open csv file: %w", err)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ReadCsvFile(): Failed to read records of csv file at path: %w", err) 
	}

	return records, nil
}

