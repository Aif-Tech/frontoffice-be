package scoreezy

import (
	"front-office/configs/application"
	"front-office/internal/datahub/job"
	"front-office/internal/scoreezy/genretail"
	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(routeAPI fiber.Router, cfg *application.Config, client httpclient.HTTPClient) {
	scoreezyGroup := routeAPI.Group("scoreezy")
	job.SetupInit(scoreezyGroup, cfg, client)

	genRetailGroupAPI := scoreezyGroup.Group("gen-retail")
	genretail.SetupInit(genRetailGroupAPI, cfg, client)
}
