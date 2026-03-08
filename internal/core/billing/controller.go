package billing

import (
	"front-office/pkg/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func NewController(service Service) Controller {
	return &controller{
		svc: service,
	}
}

type controller struct {
	svc Service
}

type Controller interface {
	SendMonthlyUsageReport(c *fiber.Ctx) error
}

func (ctrl *controller) SendMonthlyUsageReport(c *fiber.Ctx) error {
	if err := ctrl.svc.SendMonthlyUsageReport(); err != nil {
		log.Error().
			Err(err).
			Msg("failed to send monthly usage report")
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"succeed to send monthly usage report",
		nil,
	))
}
