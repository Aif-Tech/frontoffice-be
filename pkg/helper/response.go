package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
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

// func SuccessResponse(
// 	message string,
// 	data interface{},
// ) BaseResponseSuccess {
// 	return BaseResponseSuccess{
// 		Message: message,
// 		Success: true,
// 		Data:    data,
// 	}
// }

// func ErrorResponse(message string) BaseResponseFailed {
// 	return BaseResponseFailed{
// 		Message: message,
// 	}
// }

func GetError(errorMessage string) (int, interface{}) {
	var statusCode int

	switch errorMessage {
	case constant.UserNotFoundForgotEmail:
		statusCode = fiber.StatusOK
	case constant.AlreadyVerified,
		constant.ConfirmNewPasswordMismatch,
		constant.ConfirmPasswordMismatch,
		constant.DuplicateGrading,
		constant.FieldGradingLabelEmpty,
		constant.FieldMinGradeEmpty,
		constant.FieldMaxGradeEmpty,
		constant.FileSizeIsTooLarge,
		constant.IncorrectPassword,
		constant.InvalidActivationLink,
		constant.InvalidStatusValue,
		constant.InvalidDateFormat,
		constant.InvalidEmailOrPassword,
		constant.InvalidImageFile,
		constant.InvalidPassword,
		constant.InvalidPasswordResetLink,
		constant.HeaderTemplateNotValid,
		constant.OnlyUploadCSVfile,
		constant.WrongCurrentPassword,
		constant.ParamSettingIsNotSet:
		statusCode = fiber.StatusBadRequest
	case constant.RequestProhibited,
		constant.TokenExpired,
		constant.UnverifiedUser:
		statusCode = fiber.StatusUnauthorized
	case constant.DataNotFound,
		constant.RecordNotFound:
		statusCode = fiber.StatusNotFound
		errorMessage = constant.DataNotFound
	case constant.TemplateNotFound:
		statusCode = fiber.StatusNotFound
	case constant.DataAlreadyExist,
		constant.EmailAlreadyExists:
		statusCode = fiber.StatusConflict
	case constant.UpstreamError:
		statusCode = fiber.StatusBadGateway
	default:
		statusCode = fiber.StatusInternalServerError
	}

	resp := ErrorResponse(errorMessage)
	return statusCode, resp
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
