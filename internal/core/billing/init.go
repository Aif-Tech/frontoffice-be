package billing

import (
	"front-office/configs/application"
	"front-office/internal/core/log/transaction"
	"front-office/internal/mail"
	"front-office/internal/middleware"
	"front-office/pkg/httpclient"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func SetupInit(billingAPI fiber.Router, cfg *application.Config, client httpclient.HTTPClient, mailSvc *mail.SendMailService) {
	repo := NewRepository(cfg, client, nil)
	transactionRepo := transaction.NewRepository(cfg, client, nil)
	service := NewService(cfg, repo, transactionRepo, mailSvc)
	controller := NewController(service)

	billingAPI.Get("/usage/export", middleware.AdminAuth(), middleware.GetJWTPayloadFromCookie(), controller.ExportUsage)
	billingAPI.Post("/send-monthly-report", controller.SendMonthlyUsageReport)

	// Cron SendMonthlyUsageReport
	jakartaTime, _ := time.LoadLocation("Asia/Jakarta")
	scd := gocron.NewScheduler(jakartaTime)
	_, err := scd.Every(1).Month(1).At("00:00").Do(func() {
		if err := service.SendMonthlyUsageReport(); err != nil {
			log.Error().Err(err).Msg("failed to send monthly usage report")
		}
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register SendMonthlyUsageReport cron")
	}

	scd.StartAsync()
}
