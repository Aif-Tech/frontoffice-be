package template

import (
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"strings"

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

	return c.Status(fiber.StatusOK).JSON(helper.ResponseSuccess(
		"success",
		templates,
	))
}

// Download specific template
func (ctrl *controller) DownloadTemplate(c *fiber.Ctx) error {
	var req DownloadRequest
	if err := c.QueryParser(&req); err != nil {
		return apperror.BadRequest(err.Error())
	}

	if req.Product == "" {
		return apperror.BadRequest("product parameter is required")
	}

	if req.Filename == "" {
		req.Filename = "template.csv"
	} else if !strings.HasSuffix(req.Filename, ".csv") {
		req.Filename += ".csv"
	}

	path, err := ctrl.svc.DownloadTemplate(req)
	if err != nil {
		statusCode, resp := helper.GetError(constant.TemplateNotFound)

		return c.Status(statusCode).JSON(resp)
	}

	return c.Download(path)
}
