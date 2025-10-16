package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"front-office/pkg/apperror"
	"front-office/pkg/common/model"
	"io"
	"net/http"
)

func SuccessResponse[T any](message string, data T, meta ...*model.Meta) *model.APIResponse[T] {
	var m *model.Meta
	if len(meta) > 0 {
		m = meta[0]
	}

	return &model.APIResponse[T]{
		Success: true,
		Message: message,
		Data:    &data,
		Meta:    m,
	}
}

func ErrorResponse(message string) *model.APIResponse[any] {
	return &model.APIResponse[any]{
		Success: false,
		Message: message,
	}
}

func ParseAifcoreAPIResponse[T any](response *http.Response) (*model.AifcoreAPIResponse[T], error) {
	if response == nil {
		return nil, errors.New("nil http response")
	}

	dataBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp model.AifcoreAPIResponse[T]
	if err := json.Unmarshal(dataBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w; raw: %s", err, string(dataBytes))
	}

	apiResp.StatusCode = response.StatusCode

	if apiResp.StatusCode >= 400 || !apiResp.Success {
		return nil, &apperror.ExternalAPIError{
			StatusCode: apiResp.StatusCode,
			Message:    apiResp.Message,
		}
	}

	return &apiResp, nil
}

func ParseProCatAPIResponse[T any](response *http.Response) (*model.ProCatAPIResponse[T], error) {
	if response == nil {
		return nil, errors.New("nil http response")
	}

	dataBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp model.ProCatAPIResponse[T]
	if err := json.Unmarshal(dataBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w; raw: %s", err, string(dataBytes))
	}

	apiResp.StatusCode = response.StatusCode

	if apiResp.StatusCode >= 400 || !apiResp.Success {
		return &apiResp, &apperror.ExternalAPIError{
			StatusCode: apiResp.StatusCode,
			Message:    apiResp.Message,
		}
	}

	return &apiResp, nil
}

func ParseScoreezyAPIResponse[T any](response *http.Response) (*model.ScoreezyAPIResponse[T], error) {
	if response == nil {
		return nil, errors.New("nil http response")
	}

	dataBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp model.ScoreezyAPIResponse[T]
	if err := json.Unmarshal(dataBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w; raw: %s", err, string(dataBytes))
	}

	apiResp.StatusCode = response.StatusCode

	return &apiResp, nil
}
