package transaction

import (
	"github.com/gofiber/fiber/v2"
)

func NewController(service Service) Controller {
	return &controller{svc: service}
}

type controller struct {
	svc Service
}

type Controller interface {
	// scoreezy
	GetLogScoreezy(c *fiber.Ctx) error
	GetLogScoreezyByDate(c *fiber.Ctx) error
	GetLogScoreezyByDateRange(c *fiber.Ctx) error
	GetLogScoreezyByMonth(c *fiber.Ctx) error

	// product catalog
}
