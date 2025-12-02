package job

import (
	"front-office/configs/application"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/log/transaction"
	"front-office/internal/middleware"
	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(apiGroup fiber.Router, cfg *application.Config, client httpclient.HTTPClient) {
	repository := NewRepository(cfg, client, nil)
	transactionRepo := transaction.NewRepository(cfg, client, nil)
	operationRepo := operation.NewRepository(cfg, client, nil)

	service := NewService(repository, transactionRepo, operationRepo)
	controller := NewController(service)

	apiGroup.Get("/gen-retail/jobs", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetGenRetailJobs)
	apiGroup.Get("/:product_slug/jobs", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetJobs)
	apiGroup.Get("/:product_slug/jobs/:job_id", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetJobDetails)
	apiGroup.Get("/:product_slug/jobs/:job_id/export", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.ExportJobDetails)
	apiGroup.Get("/:product_slug/jobs-summary", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetJobDetailsByDateRange)
	apiGroup.Get("/:product_slug/jobs-summary/export", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.ExportJobDetailsByDateRange)
}
