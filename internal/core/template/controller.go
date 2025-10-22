package template

import (
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"

	"github.com/gofiber/fiber/v2"
)

type Controller interface {
	ListTemplates(c *fiber.Ctx) error
	DownloadTemplate(c *fiber.Ctx) error
}

type controller struct {
	svc Service
}

func NewController(service Service) Controller {
	return &controller{svc: service}
}
func (ctrl *controller) ListTemplates(c *fiber.Ctx) error {
	templates, err := ctrl.svc.ListTemplates()
	if err != nil {
		return apperror.Internal("failed to fetch template list", err)
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		constant.Success,
		templates,
	))
}

// Download specific template
func (ctrl *controller) DownloadTemplate(c *fiber.Ctx) error {
	var req DownloadRequest
	if err := c.QueryParser(&req); err != nil {
		return apperror.BadRequest(err.Error())
	}

	path, err := ctrl.svc.DownloadTemplate(req)
	if err != nil {
		return err
	}

	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	return c.Download(path)
}
