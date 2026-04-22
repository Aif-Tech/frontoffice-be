package middleware

import (
	"fmt"
	"front-office/pkg/apperror"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var appErr *apperror.AppError
		method := c.Method()
		path := c.OriginalURL()

		if ok := apperror.AsAppError(err, &appErr); ok {
			log.Error().
				Err(err).
				Int("status_code", appErr.StatusCode).
				Str("message", appErr.Message).
				Str("method", method).
				Str("path", path).
				Msg("app error")

			return c.Status(appErr.StatusCode).JSON(fiber.Map{
				"message": appErr.Message,
			})
		}

		// Jika error biasa → fallback ke 500
		log.Error().
			Err(err).
			Str("method", method).
			Str("path", path).
			Str("error_type", fmt.Sprintf("%T", err)).
			Msg("internal server error")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": fmt.Sprintf("something went wrong: %v", err),
		})
	}
}
