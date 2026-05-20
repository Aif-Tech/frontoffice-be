package member

import (
	"front-office/configs/application"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/role"
	"front-office/internal/mail"
	"front-office/internal/middleware"
	"front-office/pkg/httpclient"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func SetupInit(userAPI fiber.Router, cfg *application.Config, client httpclient.HTTPClient, mailSvc *mail.SendMailService) {
	repo := NewRepository(cfg, client, nil)
	roleRepo := role.NewRepository(cfg, client)
	logOperationRepo := operation.NewRepository(cfg, client, nil)

	serviceRole := role.NewService(roleRepo)
	service := NewService(repo, roleRepo, logOperationRepo, mailSvc)
	serviceLogOperation := operation.NewService(logOperationRepo)

	controller := NewController(service, serviceRole, serviceLogOperation)

	userAPI.Get("/", middleware.GetJWTPayloadFromCookie(cfg), controller.GetList)
	userAPI.Put("/profile", middleware.GetJWTPayloadFromCookie(cfg), middleware.ValidateRequest(updateProfileRequest{}), controller.UpdateProfile)
	userAPI.Put("/upload-profile-image", middleware.GetJWTPayloadFromCookie(cfg), middleware.FileUpload(), controller.UploadProfileImage)
	userAPI.Get("/by", middleware.GetJWTPayloadFromCookie(cfg), controller.GetBy)
	userAPI.Get("/:id", middleware.GetJWTPayloadFromCookie(cfg), controller.GetById)
	userAPI.Put("/:id", middleware.GetJWTPayloadFromCookie(cfg), middleware.AdminAuth(), middleware.ValidateRequest(updateUserRequest{}), controller.UpdateMemberById)
	userAPI.Delete("/:id", middleware.GetJWTPayloadFromCookie(cfg), middleware.AdminAuth(), controller.DeleteById)

	// Cron Update Expired Mail Status
	jakartaTime, _ := time.LoadLocation("Asia/Jakarta")
	scd := gocron.NewScheduler(jakartaTime)
	_, err := scd.Every(30).Minute().Do(controller.UpdateExpiredMailStatus)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register UpdateExpiredMailStatus cron")
	}

	scd.StartAsync()
}
