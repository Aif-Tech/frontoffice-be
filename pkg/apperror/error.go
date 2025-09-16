package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

func newAppError(status int, message string, err error) *AppError {
	// log.Error().Err(err).Msg(message)

	return &AppError{
		StatusCode: status,
		Message:    message,
		Err:        err,
	}
}

// Common helpers
func BadRequest(msg string) *AppError {
	return newAppError(http.StatusBadRequest, msg, nil)
}

func Unauthorized(msg string) *AppError {
	return newAppError(http.StatusUnauthorized, msg, nil)
}

func NotFound(msg string) *AppError {
	return newAppError(http.StatusNotFound, msg, nil)
}

func Internal(msg string, err error) *AppError {
	return newAppError(http.StatusInternalServerError, msg, err)
}

func BadGateway(msg string) *AppError {
	return newAppError(http.StatusBadGateway, msg, nil)
}

func UnprocessableEntity(msg string) *AppError {
	return newAppError(http.StatusUnprocessableEntity, msg, nil)
}

func Forbidden(msg string) *AppError {
	return newAppError(http.StatusForbidden, msg, nil)
}

func Conflict(msg string) *AppError {
	return newAppError(http.StatusConflict, msg, nil)
}

func TooManyRequests(msg string) *AppError {
	return newAppError(http.StatusTooManyRequests, msg, nil)
}

func ServiceUnavailable(msg string) *AppError {
	return newAppError(http.StatusServiceUnavailable, msg, nil)
}

func GatewayTimeout(msg string) *AppError {
	return newAppError(http.StatusGatewayTimeout, msg, nil)
}

func Unknown(msg string) *AppError {
	return newAppError(512, msg, nil)
}

func AsAppError(err error, target **AppError) bool {
	return errors.As(err, target)
}
