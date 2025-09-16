package main

import (
	"front-office/configs/application"
	"front-office/configs/server"
	"time"
)

func main() {
	loc := time.FixedZone("Asia/Jakarta", 25200)
	time.Local = loc

	cfg := application.GetConfig()

	// migrate.PostgreDB(db)

	server.NewServer(&cfg).Start()
}
