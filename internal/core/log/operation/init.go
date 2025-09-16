package operation

import (
	"front-office/configs/application"
	"front-office/internal/middleware"
	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(logAPI fiber.Router, cfg *application.Config, client httpclient.HTTPClient) {
	repository := NewRepository(cfg, client, nil)
	service := NewService(repository)
	controller := NewController(service)

	logOperationAPI := logAPI.Group("operation")
	logOperationAPI.Get("/", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetList)
	logOperationAPI.Get("/range", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetListByRange)
}
