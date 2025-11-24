package multipleloan

import (
	"errors"
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

	if err := ctrl.svc.BulkMultipleLoan(authCtx.APIKey, authCtx.QuotaTypeStr(), slug, authCtx.UserId, authCtx.CompanyId, file); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		constant.Success,
		nil,
	))
}

var productSlugMap = map[string]string{
	"7d-multiple-loan":  constant.SlugMultipleLoan7Days,
	"30d-multiple-loan": constant.SlugMultipleLoan30Days,
	"90d-multiple-loan": constant.SlugMultipleLoan90Days,
}

func mapProductSlug(slug string) (string, error) {
	if val, ok := productSlugMap[slug]; ok {
		return val, nil
	}

	return "", errors.New("unsupported product slug")
}
