package billing

import (
	"front-office/configs/application"
	"front-office/internal/core/log/transaction"
	"front-office/internal/mail"
	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(billingAPI fiber.Router, cfg *application.Config, client httpclient.HTTPClient, mailSvc *mail.SendMailService) {
	repo := NewRepository(cfg, client, nil)
	transactionRepo := transaction.NewRepository(cfg, client, nil)
	service := NewService(cfg, repo, transactionRepo, mailSvc)
	controller := NewController(service)

	billingAPI.Post("/send-monthly-report", controller.SendMonthlyUsageReport)
}
