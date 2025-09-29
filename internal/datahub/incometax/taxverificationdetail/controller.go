package taxverificationdetail

import (
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
	SingleSearch(c *fiber.Ctx) error
	BulkSearch(c *fiber.Ctx) error
}

func (ctrl *controller) SingleSearch(c *fiber.Ctx) error {
	reqBody := c.Locals(constant.Request).(*taxVerificationRequest)
	apiKey, _ := c.Locals(constant.APIKey).(string)
	memberId := fmt.Sprintf("%v", c.Locals(constant.UserId))
	companyId := fmt.Sprintf("%v", c.Locals(constant.CompanyId))

	result, err := ctrl.svc.CallTaxVerification(apiKey, memberId, companyId, reqBody)
	if err != nil {
		statusCode, resp := helper.GetError(err.Error())

		return c.Status(statusCode).JSON(resp)
	}

	return c.Status(result.StatusCode).JSON(result)
}

func (ctrl *controller) BulkSearch(c *fiber.Ctx) error {
	apiKey := fmt.Sprintf("%v", c.Locals(constant.APIKey))
	quotaType := fmt.Sprintf("%v", c.Locals(constant.QuotaType))

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

	err = ctrl.svc.BulkTaxVerification(apiKey, quotaType, memberId, companyId, file)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.ResponseSuccess(
		"success",
		nil,
	))
}
