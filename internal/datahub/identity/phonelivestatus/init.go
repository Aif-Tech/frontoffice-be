package phonelivestatus

import (
	"front-office/configs/application"
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/product"
	"front-office/internal/datahub/job"
	"front-office/internal/middleware"
	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(apiGroup fiber.Router, cfg *application.Config, client httpclient.HTTPClient) {
	repository := NewRepository(cfg, client, nil)
	productRepo := product.NewRepository(cfg, client)
	jobRepo := job.NewRepository(cfg, client, nil)
	transactionRepo := transaction.NewRepository(cfg, client, nil)
	jobService := job.NewService(jobRepo, transactionRepo)
	service := NewService(repository, productRepo, jobRepo, transactionRepo, jobService)
	controller := NewController(service)

	phoneLiveStatusGroup := apiGroup.Group("phone-live-status")
	phoneLiveStatusGroup.Post("/single-request", middleware.Auth(), middleware.IsRequestValid(phoneLiveStatusRequest{}), middleware.GetJWTPayloadFromCookie(), controller.SingleSearch)
	phoneLiveStatusGroup.Post("/bulk-request", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.BulkSearch)
	phoneLiveStatusGroup.Get("/jobs", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetJobs)
	phoneLiveStatusGroup.Get("/jobs/:id/details", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetJobDetails)
	phoneLiveStatusGroup.Get("/jobs/:id/details/export", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.ExportJobDetails)
	phoneLiveStatusGroup.Get("/jobs-summary", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetJobsSummary)
	phoneLiveStatusGroup.Get("/jobs-summary/export", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.ExportJobsSummary)
}
