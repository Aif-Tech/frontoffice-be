package server

import (
	"front-office/configs/application"
	"front-office/internal/core"
	"front-office/internal/middleware"

	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type fiberServer struct {
	App *fiber.App
	Cfg *application.Config
}

func NewServer(cfg *application.Config) Server {
	return &fiberServer{
		App: fiber.New(
			fiber.Config{
				ErrorHandler: middleware.ErrorHandler(),
			},
		),
		Cfg: cfg,
	}
}

func (s *fiberServer) Start() {
	s.App.Use(recover.New())
	// Healthcheck system
	// /live => Liveness
	// /ready => Readyness
	s.App.Use(healthcheck.New())
	s.App.Static("/", "./storage/uploads")
	s.App.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Authorization",
		AllowOrigins:     s.Cfg.Env.FrontendBaseUrl,
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		ExposeHeaders:    "Set-Cookie",
	}))

	api := s.App.Group("/api/fo")
	core.SetupInit(api, s.Cfg)

	log.Fatal(s.App.Listen(":" + s.Cfg.Env.Port))
}
