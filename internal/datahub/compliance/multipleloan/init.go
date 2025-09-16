package multipleloan

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
	repo := NewRepository(cfg, client, nil)
	productRepo := product.NewRepository(cfg, client)
	jobRepo := job.NewRepository(cfg, client, nil)
	transactionRepo := transaction.NewRepository(cfg, client, nil)

	jobService := job.NewService(jobRepo, transactionRepo)
	service := NewService(repo, productRepo, jobRepo, transactionRepo, jobService)

	controller := NewController(service)

	apiGroup.Post("/:product_slug/single-request", middleware.Auth(), middleware.IsRequestValid(multipleLoanRequest{}), middleware.GetJWTPayloadFromCookie(), controller.MultipleLoan)
	apiGroup.Post("/:product_slug/bulk-request", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.BulkMultipleLoan)
}
