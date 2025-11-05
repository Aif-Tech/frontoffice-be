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
		return nil, fmt.Errorf("failed to open file: %w", err)
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
		return nil, fmt.Errorf("failed to read csv: %w", err)
	}
	if len(csvData) == 0 {
		return nil, errors.New("empty CSV file — please use the provided template")
	}

	header := csvData[0]
	if len(header) == 0 {
		return nil, errors.New("missing CSV header row — please include the first row from the template")
	}

	// --- Normalize helper
	normalize := func(headers []string) []string {
		out := make([]string, len(headers))
		for i, h := range headers {
			out[i] = strings.ToLower(strings.TrimSpace(h))
		}
		return out
	}

	normalizedHeader := normalize(header)
	normalizedExpected := normalize(expectedHeaders)

	// --- Step 1: Check missing or extra headers
	headerSet := make(map[string]bool)
	for _, h := range normalizedHeader {
		headerSet[h] = true
	}

	var missingHeaders []string
	for _, expected := range normalizedExpected {
		if !headerSet[expected] {
			missingHeaders = append(missingHeaders, expected)
		}
	}

	if len(missingHeaders) > 0 {
		return nil, fmt.Errorf(
			"invalid CSV header — missing required columns: %s. Please use the provided template",
			strings.Join(missingHeaders, ", "),
		)
	}

	if len(normalizedHeader) != len(normalizedExpected) {
		return nil, fmt.Errorf(
			"invalid CSV header — expected %d columns, got %d. Please use the provided template",
			len(normalizedExpected), len(normalizedHeader),
		)
	}

	// --- Step 2: Validate order
	for i := range normalizedExpected {
		if normalizedHeader[i] != normalizedExpected[i] {
			return nil, fmt.Errorf(
				"invalid CSV header order — expected [%s], but got [%s]. Please do not modify column order in the template",
				strings.Join(expectedHeaders, ", "),
				strings.Join(header, ", "),
			)
		}
	}

	return csvData, nil
}
