package auth

import (
	"front-office/configs/application"
	"front-office/internal/core/activation"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/member"
	"front-office/internal/core/passwordreset"
	"front-office/internal/core/role"
	"front-office/internal/middleware"

	"front-office/pkg/httpclient"

	"github.com/gofiber/fiber/v2"
)

func SetupInit(authAPI fiber.Router, cfg *application.Config, client httpclient.HTTPClient) {
	repo := NewRepository(cfg, client, nil)
	memberRepo := member.NewRepository(cfg, client, nil)
	roleRepo := role.NewRepository(cfg, client)
	activationTokenRepo := activation.NewRepository(cfg, client, nil)
	passwordResetRepo := passwordreset.NewRepository(cfg, client, nil)
	logOperationRepo := operation.NewRepository(cfg, client, nil)

	service := NewService(cfg, repo, memberRepo, roleRepo, logOperationRepo, activationTokenRepo, passwordResetRepo)
	serviceUser := member.NewService(memberRepo, roleRepo, logOperationRepo)
	serviceActivationToken := activation.NewService(activationTokenRepo, cfg)
	servicePasswordResetToken := passwordreset.NewService(passwordResetRepo, cfg)
	serviceLogOperation := operation.NewService(logOperationRepo)

	controller := NewController(service, serviceUser, serviceActivationToken, servicePasswordResetToken, serviceLogOperation, cfg)

	authAPI.Post("/register-member", middleware.AdminAuth(), middleware.GetJWTPayloadFromCookie(), middleware.IsRequestValid(member.RegisterMemberRequest{}), controller.RegisterMember)
	authAPI.Post("/login", middleware.IsRequestValid(userLoginRequest{}), controller.Login)
	authAPI.Put("/verify/:token", middleware.SetHeaderAuth, middleware.IsRequestValid(PasswordResetRequest{}), controller.VerifyUser)
	authAPI.Post("/logout", middleware.Auth(), middleware.GetJWTPayloadFromCookie(), controller.Logout)
	authAPI.Post("/refresh-access", middleware.GetPayloadFromRefreshToken(), controller.RefreshAccessToken)
	authAPI.Put("/send-email-activation/:email", middleware.Auth(), middleware.AdminAuth(), middleware.GetJWTPayloadFromCookie(), controller.RequestActivation)
	authAPI.Post("/request-password-reset", middleware.IsRequestValid(RequestPasswordResetRequest{}), controller.RequestPasswordReset)
	authAPI.Put("/password-reset/:token", middleware.SetCookiePasswordResetToken, middleware.GetJWTPayloadPasswordResetFromCookie(), middleware.IsRequestValid(PasswordResetRequest{}), controller.PasswordReset)
	authAPI.Put("/change-password", middleware.GetJWTPayloadFromCookie(), middleware.IsRequestValid(ChangePasswordRequest{}), controller.ChangePassword)
}
