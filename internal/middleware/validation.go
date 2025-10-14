package middleware

import (
	"fmt"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/usepzaka/validator"
)

func ValidateRequest(model interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		request := reflect.New(reflect.TypeOf(model)).Interface()

		if err := c.BodyParser(request); err != nil {
			resp := helper.ErrorResponse(constant.InvalidRequestFormat)

			return c.Status(fiber.StatusBadRequest).JSON(resp)
		}

		if errValid := validator.ValidateStruct(request); errValid != nil {
			resp := helper.ErrorResponse(errValid.Error())

			return c.Status(fiber.StatusBadRequest).JSON(resp)
		}

		c.Locals(constant.Request, request)

		return c.Next()
	}
}

func ValidateCSVFile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return apperror.BadRequest("file is required")
		}

		const maxSize = 30 * 1024 * 1024 // 30 MB
		if file.Size > maxSize {
			return apperror.BadRequest(fmt.Sprintf("file too large (max %dmb)", maxSize/1024/1024))
		}

		if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
			return apperror.BadRequest("invalid file type")
		}

		c.Locals(constant.ValidatedFile, file)
		return c.Next()
	}
}
