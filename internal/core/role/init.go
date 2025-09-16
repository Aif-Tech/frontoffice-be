package role

import (
	"front-office/configs/application"
	"front-office/internal/middleware"
	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(roleAPI fiber.Router, cfg *application.Config, client httpclient.HTTPClient) {
	repo := NewRepository(cfg, client)
	service := NewService(repo)
	controller := NewController(service)

	roleAPI.Get("/", middleware.Auth(), controller.GetRoles)
	roleAPI.Get("/:id", middleware.Auth(), controller.GetRoleById)
}
