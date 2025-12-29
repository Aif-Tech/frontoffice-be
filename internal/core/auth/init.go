package auth

import (
	"front-office/configs/application"
	"front-office/internal/core/activation"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/member"
	"front-office/internal/core/passwordreset"
	"front-office/internal/core/role"
	"front-office/internal/mail"
	"front-office/internal/middleware"

	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(authAPI fiber.Router, cfg *application.Config, client httpclient.HTTPClient, mailSvc *mail.SendMailService) {
	repo := NewRepository(cfg, client, nil)
	memberRepo := member.NewRepository(cfg, client, nil)
	roleRepo := role.NewRepository(cfg, client)
	activationTokenRepo := activation.NewRepository(cfg, client, nil)
	passwordResetRepo := passwordreset.NewRepository(cfg, client, nil)
	logOperationRepo := operation.NewRepository(cfg, client, nil)

	serviceUser := member.NewService(memberRepo, roleRepo, logOperationRepo)
	serviceActivationToken := activation.NewService(activationTokenRepo, cfg)
	servicePasswordResetToken := passwordreset.NewService(passwordResetRepo, cfg)
	serviceLogOperation := operation.NewService(logOperationRepo)
	service := NewService(cfg, repo, memberRepo, roleRepo, logOperationRepo, activationTokenRepo, passwordResetRepo, mailSvc)

	controller := NewController(service, serviceUser, serviceActivationToken, servicePasswordResetToken, serviceLogOperation, cfg)

	authAPI.Post("/register-member", middleware.AdminAuth(), middleware.GetJWTPayloadFromCookie(), middleware.ValidateRequest(member.RegisterMemberRequest{}), controller.RegisterMember)
	authAPI.Post("/login", middleware.ValidateRequest(userLoginRequest{}), controller.Login)
	authAPI.Put("/verify/:token", middleware.SetHeaderAuth, middleware.ValidateRequest(passwordResetRequest{}), controller.VerifyUser)
	authAPI.Post("/logout", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.Logout)
	authAPI.Post("/refresh-access", middleware.GetPayloadFromRefreshToken(), controller.RefreshAccessToken)
	authAPI.Put("/send-email-activation/:email", middleware.Auth(), middleware.AdminAuth(), middleware.GetJWTPayloadFromCookie(), controller.RequestActivation)
	authAPI.Post("/request-password-reset", middleware.ValidateRequest(requestPasswordResetRequest{}), controller.RequestPasswordReset)
	authAPI.Put("/password-reset/:token", middleware.SetCookiePasswordResetToken, middleware.GetJWTPayloadPasswordResetFromCookie(), middleware.ValidateRequest(passwordResetRequest{}), controller.PasswordReset)
	authAPI.Put("/change-password", middleware.GetJWTPayloadFromCookie(), middleware.ValidateRequest(changePasswordRequest{}), controller.ChangePassword)
}
