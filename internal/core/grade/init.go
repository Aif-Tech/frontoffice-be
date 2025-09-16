package grade

import (
	"front-office/configs/application"
	"front-office/internal/middleware"
	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(gradingAPI fiber.Router, cfg *application.Config, client httpclient.HTTPClient) {
	repo := NewRepository(cfg, client, nil)
	service := NewService(repo)
	controller := NewController(service)

	gradingAPI.Put("/", middleware.AdminAuth(), middleware.GetJWTPayloadFromCookie(), middleware.IsRequestValid(createGradeRequest{}), controller.SaveGrading)
	gradingAPI.Get("/", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetGrades)
}
