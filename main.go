package main

import (
	"front-office/configs/application"
	"front-office/configs/server"
	"front-office/internal/mail"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	loc := time.FixedZone("Asia/Jakarta", 25200)
	time.Local = loc

	cfg := application.GetConfig()
	initLogger(cfg.App.AppEnv)

	mailModule := mail.Init(&cfg)

	srv := server.NewServer(&cfg, mailModule)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("server stopped")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutdown signal received")

	mailModule.Worker.Stop()

	if err := srv.Shutdown(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown http server")
	}

	log.Info().Msg("graceful shutdown completed")
}

func initLogger(env string) {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "local" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}
}
