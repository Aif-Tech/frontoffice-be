package helper

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/rs/zerolog/log"
)

func ParseCSVFile(file *multipart.FileHeader, expectedHeaders []string) ([][]string, error) {
	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Error().
				Err(cerr).
				Msg("failed to close file")
		}
	}()

	// Read sample for delimiter detection
	buf := make([]byte, 2048)
	n, _ := f.Read(buf)
	delimiter := detectDelimiter(buf[:n])

	// Reset reader after peek
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	reader := csv.NewReader(f)
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

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

	// Normalize helper
	normalize := func(headers []string) []string {
		out := make([]string, len(headers))
		for i, h := range headers {
			out[i] = strings.ToLower(strings.TrimSpace(h))
		}
		return out
	}

	normalizedHeader := normalize(header)
	normalizedExpected := normalize(expectedHeaders)

	//  Check missing or extra headers
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

	// Validate order
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

func detectDelimiter(sample []byte) rune {
	comma := bytes.Count(sample, []byte(","))
	semicolon := bytes.Count(sample, []byte(";"))

	if semicolon > comma {
		return ';'
	}

	return ','
}
