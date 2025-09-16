package apperror

import (
	"errors"
	"front-office/pkg/common/constant"
	"strings"
)

func MapExternalAPIError(err *ExternalAPIError) error {
	switch err.StatusCode {
	case 400:
		return BadRequest(err.Message)
	case 401:
		return Unauthorized(err.Message)
	case 403:
		return Forbidden(err.Message)
	case 404:
		return NotFound(err.Message)
	case 409:
		return Conflict(err.Message)
	case 422:
		return UnprocessableEntity(err.Message)
	case 429:
		return TooManyRequests("too many requests")
	case 500:
		return Internal(err.Message, err)
	case 502:
		return BadGateway("bad gateway from external service")
	case 503:
		return ServiceUnavailable("external service unavailable")
	case 504:
		return GatewayTimeout("external service timeout")
	case 512:
		return Unknown(err.Message)
	default:
		return BadGateway("unexpected external service error")
	}
}

func MapRepoError(err error, context string) error {
	var apiErr *ExternalAPIError
	if errors.As(err, &apiErr) {
		return MapExternalAPIError(apiErr)
	}

	return Internal(context, err)
}

func MapAuthError(err *ExternalAPIError) error {
	if err.StatusCode == 401 {
		if strings.Contains(err.Message, "not Active") {
			return Unauthorized("your account is not active")
		}

		return Unauthorized(constant.InvalidEmailOrPassword)
	}

	return MapExternalAPIError(err)
}

func MapLoanError(err *ExternalAPIError) error {
	if err.StatusCode == 512 {
		return Unknown("The data partner service is currently unavailable.")
	}

	return MapExternalAPIError(err)
}

func MapChangePasswordError(err *ExternalAPIError) error {
	if strings.Contains(err.Message, "not the hash") {
		return BadRequest("current password is wrong")
	}

	return MapExternalAPIError(err)
}
