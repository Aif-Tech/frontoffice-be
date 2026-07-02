package billing

import (
	"front-office/configs/application"
	"front-office/internal/core/internalteam"
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
	internalTeamRepo := internalteam.NewRepository(cfg, client, nil)
	service := NewService(cfg, repo, transactionRepo, internalTeamRepo, mailSvc)
	controller := NewController(service)

	billingAPI.Get("/usage", middleware.GetJWTPayloadFromCookie(cfg), middleware.AdminAuth(), controller.GetUsageReport)
	billingAPI.Get("/usage/export", middleware.GetJWTPayloadFromCookie(cfg), middleware.AdminAuth(), controller.ExportUsage)
	billingAPI.Post("/send-monthly-report", controller.SendMonthlyUsageReport)

	setupCron(service)
}

func targetRunDay(now time.Time) int {
	day3 := time.Date(now.Year(), now.Month(), 3, 0, 0, 0, 0, now.Location())
	switch day3.Weekday() {
	case time.Saturday:
		return 5
	case time.Sunday:
		return 4
	default:
		return 3
	}
}

func setupCron(service Service) {
	jakartaTime, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load Asia/Jakarta timezone")
	}

	scd := gocron.NewScheduler(jakartaTime)

	_, err = scd.Every(1).Day().At("09:00").Do(func() {
		now := time.Now().In(jakartaTime)
		target := targetRunDay(now)

		if now.Day() != target {
			return
		}

		log.Info().
			Int("day", now.Day()).
			Str("weekday", now.Weekday().String()).
			Msg("running monthly usage report")

		if err := service.SendMonthlyUsageReport(); err != nil {
			log.Error().Err(err).Msg("failed to send monthly usage report")
		}
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register monthly usage report cron")
	}

	scd.StartAsync()
}
