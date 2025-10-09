package grade

import (
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"

	"github.com/gofiber/fiber/v2"
)

func NewController(service Service) Controller {
	return &controller{Svc: service}
}

type controller struct {
	Svc Service
}

type Controller interface {
	SaveGrading(c *fiber.Ctx) error
	GetGrades(c *fiber.Ctx) error
}

func (ctrl *controller) SaveGrading(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*createGradeRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	if err := ctrl.Svc.SaveGrading(&createGradePayload{
		CompanyId:   authCtx.CompanyIdStr(),
		ProductSlug: constant.SlugGenRetailV3,
		Request: createGradeRequest{
			reqBody.Grades,
		},
	}); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(helper.SuccessResponse[any](
		constant.Success,
		nil,
	))
}

func (ctrl *controller) GetGrades(c *fiber.Ctx) error {
	authCtx, err := helper.GetAuthContext(c)
	if err != nil {
		return apperror.Unauthorized(err.Error())
	}

	grades, err := ctrl.Svc.GetGrades(constant.SlugGenRetailV3, authCtx.CompanyIdStr())
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		constant.Success,
		grades,
	))
}
