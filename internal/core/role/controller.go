package role

import (
	"front-office/pkg/apperror"
	"front-office/pkg/helper"

	"github.com/gofiber/fiber/v2"
)

func NewController(service Service) Controller {
	return &controller{svc: service}
}

type controller struct {
	svc Service
}

type Controller interface {
	GetRoleById(c *fiber.Ctx) error
	GetRoles(c *fiber.Ctx) error
}

func (ctrl *controller) GetRoleById(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apperror.BadRequest("missing role id")
	}

	role, err := ctrl.svc.GetRoleById(id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		"succeed to get a role by Id",
		role,
	))
}

func (ctrl *controller) GetRoles(c *fiber.Ctx) error {
	name := c.Query("name", "")

	filter := RoleFilter{
		Name: name,
	}

	roles, err := ctrl.svc.GetRoles(filter)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		"succeed to get list of roles",
		roles,
	))
}
