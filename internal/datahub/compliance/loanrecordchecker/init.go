package loanrecordchecker

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

	loanRecordCheckerGroup := apiGroup.Group("loan-record-checker")
	loanRecordCheckerGroup.Post("/single-request", middleware.Auth(), middleware.IsRequestValid(loanRecordCheckerRequest{}), middleware.GetJWTPayloadFromCookie(), controller.SingleSearch)
	loanRecordCheckerGroup.Post("/bulk-request", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.BulkSearch)
}
