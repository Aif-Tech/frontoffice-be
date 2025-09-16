package genretail

import (
	"front-office/configs/application"
	"front-office/internal/core/grade"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/product"
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

	service := NewService(repo, gradeRepo, transRepo, productRepo, logRepo)

	controller := NewController(service)

	genRetailGroup := apiGroup.Group("gen-retail")
	genRetailGroup.Post("/dummy-request", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), middleware.IsRequestValid(genRetailRequest{}), controller.DummyRequestScore)
	genRetailGroup.Post("/single-request", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), middleware.IsRequestValid(genRetailRequest{}), controller.SingleRequest)
	genRetailGroup.Post("/bulk-request", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.BulkRequest)
	genRetailGroup.Get("/logs", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetLogsScoreezy)
	genRetailGroup.Get("/logs/export", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.ExportJobDetails)
	genRetailGroup.Get("/logs/:trx_id", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetLogScoreezy)
	// genRetailAPI.Put("/upload-scoring-template", middleware.Auth(), middleware.IsRequestValid(UploadScoringRequest{}), middleware.GetJWTPayloadFromCookie(), middleware.DocUpload(), controller.UploadCSV)
}
