package middleware

import (
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"reflect"

	"github.com/gofiber/fiber/v2"
	"github.com/usepzaka/validator"
)

func IsRequestValid(model interface{}) fiber.Handler {
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
