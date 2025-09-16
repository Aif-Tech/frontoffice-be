package apperror

import "fmt"

// ExternalAPIError is used to wrap error response from external APIs.
type ExternalAPIError struct {
	StatusCode int
	Message    string
}

func (e *ExternalAPIError) Error() string {
	return fmt.Sprintf("external api error [%d]: %s", e.StatusCode, e.Message)
}
