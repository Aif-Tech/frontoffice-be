package helper

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
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
	if len(header) == 0 {
		return nil, errors.New("missing csv header row")
	}

	headerMap := make(map[string]bool)
	for _, h := range header {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = true
	}

	var missingHeaders []string
	for _, expected := range expectedHeaders {
		if !headerMap[strings.ToLower(strings.TrimSpace(expected))] {
			missingHeaders = append(missingHeaders, expected)
		}
	}

	if len(missingHeaders) > 0 {
		return nil, fmt.Errorf("csv file missing required headers: %v", strings.Join(missingHeaders, ", "))
	}

	return csvData, nil
}
