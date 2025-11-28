package member

import (
	"front-office/configs/application"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/role"
	"front-office/internal/middleware"
	"front-office/pkg/httpclient"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func SetupInit(userAPI fiber.Router, cfg *application.Config, client httpclient.HTTPClient) {
	repo := NewRepository(cfg, client, nil)
	roleRepo := role.NewRepository(cfg, client)
	logOperationRepo := operation.NewRepository(cfg, client, nil)

	serviceRole := role.NewService(roleRepo)
	service := NewService(repo, roleRepo, logOperationRepo)
	serviceLogOperation := operation.NewService(logOperationRepo)

	controller := NewController(service, serviceRole, serviceLogOperation)

	userAPI.Get("/", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetList)
	userAPI.Put("/profile", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), middleware.ValidateRequest(updateProfileRequest{}), controller.UpdateProfile)
	userAPI.Put("/upload-profile-image", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), middleware.FileUpload(), controller.UploadProfileImage)
	userAPI.Get("/by", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetBy)
	userAPI.Get("/:id", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.GetById)
	userAPI.Put("/:id", middleware.AdminAuth(), middleware.ValidateRequest(updateUserRequest{}), middleware.GetJWTPayloadFromCookie(), controller.UpdateMemberById)
	userAPI.Delete("/:id", middleware.AdminAuth(), middleware.GetJWTPayloadFromCookie(), controller.DeleteById)

	// Cron Update Expired Mail Status
	jakartaTime, _ := time.LoadLocation("Asia/Jakarta")
	scd := gocron.NewScheduler(jakartaTime)
	_, err := scd.Every(30).Minute().Do(controller.UpdateExpiredMailStatus)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register UpdateExpiredMailStatus cron")
	}

	scd.StartAsync()
}
