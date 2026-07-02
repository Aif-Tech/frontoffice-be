package billing

import (
	"front-office/configs/application"
	"front-office/internal/core/internalteam"
	"front-office/internal/core/log/transaction"
	"front-office/internal/mail"
	"front-office/internal/middleware"
	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
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

	// setupCron(service)
}

// func setupCron(service Service) {
// 	jakartaTime, _ := time.LoadLocation("Asia/Jakarta")
// 	scd := gocron.NewScheduler(jakartaTime)

// 	isWeekend := func(t time.Time) bool {
// 		return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
// 	}

// 	runReport := func() {
// 		if err := service.SendMonthlyUsageReport(); err != nil {
// 			log.Error().Err(err).Msg("failed to send monthly usage report")
// 		}
// 	}

// 	_, err := scd.Every(1).Month(3).At("09:00").Do(func() {
// 		now := time.Now().In(jakartaTime)
// 		if isWeekend(now) {
// 			log.Info().Msg("tanggal 3 adalah weekend, skip")
// 			return
// 		}
// 		runReport()
// 	})
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("failed to register cron tanggal 3")
// 	}

// 	_, err = scd.Every(1).Month(4).At("09:00").Do(func() {
// 		now := time.Now().In(jakartaTime)
// 		if !isWeekend(now.AddDate(0, 0, -1)) {
// 			return
// 		}
// 		if isWeekend(now) {
// 			log.Info().Msg("tanggal 4 adalah weekend, skip")
// 			return
// 		}
// 		runReport()
// 	})
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("failed to register cron tanggal 4")
// 	}

// 	_, err = scd.Every(1).Month(5).At("09:00").Do(func() {
// 		now := time.Now().In(jakartaTime)
// 		day3 := now.AddDate(0, 0, -2).Weekday()
// 		if day3 != time.Saturday {
// 			return
// 		}
// 		runReport()
// 	})
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("failed to register cron tanggal 5")
// 	}

// 	scd.StartAsync()
// }
