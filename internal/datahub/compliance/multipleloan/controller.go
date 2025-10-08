package multipleloan

import (
	"errors"
	"fmt"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"

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
	req := c.Locals(constant.Request).(*multipleLoanRequest)
	apiKey := fmt.Sprintf("%v", c.Locals(constant.APIKey))
	memberIdStr := fmt.Sprintf("%v", c.Locals(constant.UserId))
	companyIdStr := fmt.Sprintf("%v", c.Locals(constant.CompanyId))
	slug := c.Params("product_slug")

	multipleLoanRes, err := ctrl.svc.MultipleLoan(apiKey, slug, memberIdStr, companyIdStr, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(multipleLoanRes)
}

func (ctrl *controller) BulkMultipleLoan(c *fiber.Ctx) error {
	apiKey := fmt.Sprintf("%v", c.Locals(constant.APIKey))
	quotaType := fmt.Sprintf("%v", c.Locals(constant.QuotaType))
	slug := c.Params("product_slug")

	memberId, err := helper.InterfaceToUint(c.Locals(constant.UserId))
	if err != nil {
		return apperror.Unauthorized(constant.InvalidUserSession)
	}

	companyId, err := helper.InterfaceToUint(c.Locals(constant.CompanyId))
	if err != nil {
		return apperror.Unauthorized(constant.InvalidCompanySession)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	err = ctrl.svc.BulkMultipleLoan(apiKey, quotaType, slug, memberId, companyId, file)
	if err != nil {
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
