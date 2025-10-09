package taxcompliancestatus

import (
	"front-office/configs/application"
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

	jobService := job.NewService(jobRepo, transactionRepo)
	service := NewService(repo, memberRepo, jobRepo, transactionRepo, jobService)

	controller := NewController(service)

	taxComplianceGroup := apiGroup.Group("tax-compliance-status")
	taxComplianceGroup.Post("/single-request", middleware.Auth(), middleware.ValidateRequest(taxComplianceStatusRequest{}), middleware.GetJWTPayloadFromCookie(), controller.SingleSearch)
	taxComplianceGroup.Post("/bulk-request", middleware.Auth(), middleware.ValidateCSVFile(), middleware.GetJWTPayloadFromCookie(), controller.BulkSearch)
}
