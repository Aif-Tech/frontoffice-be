package helper

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
)

func ParseCSVFile(file *multipart.FileHeader, expectedHeaders []string) ([][]string, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Printf("failed to close file: %v", cerr)
		}
	}()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1

	csvData, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(csvData) == 0 {
		return nil, errors.New("empty csv file")
	}

	header := csvData[0]
	if len(header) < len(expectedHeaders) {
		return nil, fmt.Errorf("invalid csv header length: expected %d, got %d", len(expectedHeaders), len(header))
	}

	for i, expectedHeader := range expectedHeaders {
		if header[i] != expectedHeader {
			return nil, fmt.Errorf("invalid csv header at column %d: expected %q, got %q", i+1, expectedHeader, header[i])
		}
	}

	return csvData, nil
}
