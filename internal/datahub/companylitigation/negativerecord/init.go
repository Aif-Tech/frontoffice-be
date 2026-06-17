package negativerecord

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
	jobRepo := job.NewRepository(cfg, client, nil)
	transactionRepo := transaction.NewRepository(cfg, client, nil)
	operationRepo := operation.NewRepository(cfg, client, nil)
	memberRepo := member.NewRepository(cfg, client, nil)

	jobService := job.NewService(jobRepo, transactionRepo, operationRepo)
	service := NewService(repo, memberRepo, jobRepo, operationRepo, transactionRepo, jobService)

	controller := NewController(service)

	negativeRecordGroup := apiGroup.Group("negative-record")
	negativeRecordGroup.Post("/single-request", middleware.GetJWTPayloadFromCookie(cfg), middleware.ValidateRequest(negativeRecordRequest{}), controller.SingleRequest)
	negativeRecordGroup.Post("/bulk-request", middleware.GetJWTPayloadFromCookie(cfg), middleware.ValidateCSVFile(), controller.BulkSearch)
}
