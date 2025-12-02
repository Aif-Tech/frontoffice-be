package multipleloan

import (
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
)

func NewController(
	svc Service,
) Controller {
	return &controller{svc}
}

type controller struct {
	svc Service
}

type Controller interface {
	MultipleLoan(c *fiber.Ctx) error
	BulkMultipleLoan(c *fiber.Ctx) error
}

func (ctrl *controller) MultipleLoan(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*multipleLoanRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	slug := c.Params("product_slug")

	multipleLoanRes, err := ctrl.svc.MultipleLoan(authCtx, slug, reqBody)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(multipleLoanRes)
}

func (ctrl *controller) BulkMultipleLoan(c *fiber.Ctx) error {
	file, ok := c.Locals(constant.ValidatedFile).(*multipart.FileHeader)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	slug := c.Params("product_slug")

	if err := ctrl.svc.BulkMultipleLoan(authCtx, slug, file); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		constant.Success,
		nil,
	))
}
