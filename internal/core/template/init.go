package template

import (
	"github.com/gofiber/fiber/v2"
)

func SetupInit(apiGroup fiber.Router) {
	repository := NewRepository()
	service := NewService(repository)
	controller := NewController(service)

	apiGroup.Get("/", controller.ListTemplates)
	apiGroup.Get("/download", controller.DownloadTemplate)
}
