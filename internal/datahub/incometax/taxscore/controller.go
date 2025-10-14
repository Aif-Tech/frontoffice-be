package taxscore

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
	SingleSearch(c *fiber.Ctx) error
	BulkSearch(c *fiber.Ctx) error
}

func (ctrl *controller) SingleSearch(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*taxScoreRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	result, err := ctrl.svc.TaxScore(authCtx.APIKey, authCtx.UserIdStr(), authCtx.CompanyIdStr(), reqBody)
	if err != nil {
		return err
	}

	return c.Status(result.StatusCode).JSON(result)
}

func (ctrl *controller) BulkSearch(c *fiber.Ctx) error {
	file, ok := c.Locals(constant.ValidatedFile).(*multipart.FileHeader)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	if err := ctrl.svc.BulkTaxScore(authCtx.APIKey, authCtx.QuotaTypeStr(), authCtx.UserId, authCtx.CompanyId, file); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		constant.Success,
		nil,
	))
}
