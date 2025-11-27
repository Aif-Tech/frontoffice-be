package multipleloan

import (
	"front-office/configs/application"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/member"
	"front-office/internal/datahub/job"
	"front-office/internal/middleware"
	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(apiGroup fiber.Router, cfg *application.Config, client httpclient.HTTPClient) {
	repo := NewRepository(cfg, client, nil)
	memberRepo := member.NewRepository(cfg, client, nil)
	jobRepo := job.NewRepository(cfg, client, nil)
	transactionRepo := transaction.NewRepository(cfg, client, nil)
	operationRepo := operation.NewRepository(cfg, client, nil)

	jobService := job.NewService(jobRepo, transactionRepo, operationRepo)
	service := NewService(repo, memberRepo, jobRepo, transactionRepo, operationRepo, jobService)

	controller := NewController(service)

	apiGroup.Post("/:product_slug/single-request", middleware.Auth(), middleware.ValidateRequest(multipleLoanRequest{}), middleware.GetJWTPayloadFromCookie(), controller.MultipleLoan)
	apiGroup.Post("/:product_slug/bulk-request", middleware.Auth(), middleware.ValidateCSVFile(), middleware.GetJWTPayloadFromCookie(), controller.BulkMultipleLoan)
}
