package genretail

import (
	"front-office/configs/application"
	"front-office/internal/core/grade"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/member"
	"front-office/internal/core/product"
	"front-office/internal/datahub/job"
	"front-office/internal/middleware"
	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(apiGroup fiber.Router, cfg *application.Config, client httpclient.HTTPClient) {
	repo := NewRepository(cfg, client, nil)
	gradeRepo := grade.NewRepository(cfg, client, nil)
	transRepo := transaction.NewRepository(cfg, client, nil)
	productRepo := product.NewRepository(cfg, client)
	logRepo := operation.NewRepository(cfg, client, nil)
	jobRepo := job.NewRepository(cfg, client, nil)
	memberRepo := member.NewRepository(cfg, client, nil)

	service := NewService(repo, gradeRepo, transRepo, productRepo, logRepo, jobRepo, memberRepo)

	controller := NewController(service)

	apiGroup.Post("/dummy-request", middleware.GetJWTPayloadFromCookie(cfg), middleware.ValidateRequest(genRetailRequest{}), controller.DummyRequestScore)
	apiGroup.Post("/single-request", middleware.GetJWTPayloadFromCookie(cfg), middleware.ValidateRequest(genRetailRequest{}), controller.SingleRequest)
	apiGroup.Post("/bulk-request", middleware.ValidateCSVFile(), middleware.GetJWTPayloadFromCookie(cfg), controller.BulkRequest)
	apiGroup.Get("/logs", middleware.GetJWTPayloadFromCookie(cfg), controller.GetLogsScoreezy)
	apiGroup.Get("/logs/export", middleware.GetJWTPayloadFromCookie(cfg), controller.ExportJobDetails)
	apiGroup.Get("/logs/:trx_id", middleware.GetJWTPayloadFromCookie(cfg), controller.GetLogScoreezy)
	// genRetailAPI.Put("/upload-scoring-template", middleware.ValidateRequest(UploadScoringRequest{}), middleware.GetJWTPayloadFromCookie(cfg), middleware.DocUpload(), controller.UploadCSV)
}
