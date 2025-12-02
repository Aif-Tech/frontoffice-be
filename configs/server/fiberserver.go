package server

import (
	"front-office/configs/application"
	"front-office/internal/core"
	"front-office/internal/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"
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
		AllowOrigins:     s.Cfg.App.FrontendBaseUrl,
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		ExposeHeaders:    "Set-Cookie",
	}))

	s.App.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		log.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", time.Since(start)).
			Msg("http request")

		return err
	})

	api := s.App.Group("/api/fo")
	core.SetupInit(api, s.Cfg)

	log.Info().
		Str("port", s.Cfg.App.Port).
		Msg("starting fiber http server")

	addr := ":" + s.Cfg.App.Port
	if err := s.App.Listen(addr); err != nil {
		log.Fatal().
			Err(err).
			Str("addr", addr).
			Msg("failed to start http server")
	}
}
