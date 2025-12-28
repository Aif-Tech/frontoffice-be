package main

import (
	"front-office/configs/application"
	"front-office/configs/server"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	loc := time.FixedZone("Asia/Jakarta", 25200)
	time.Local = loc

	cfg := application.GetConfig()
	initLogger(cfg.App.AppEnv)

	server.NewServer(&cfg).Start()
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
