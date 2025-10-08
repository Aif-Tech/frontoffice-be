package auth

import (
	"fmt"
	"front-office/configs/application"
	"front-office/internal/core/activation"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/member"
	"front-office/internal/core/passwordreset"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"

	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func NewController(
	service Service,
	svcUser member.Service,
	svcActivationToken activation.Service,
	svcPasswordResetToken passwordreset.Service,
	svcLogOperation operation.Service,
	cfg *application.Config,
) Controller {
	return &controller{
		svc:                   service,
		svcUser:               svcUser,
		svcActivationToken:    svcActivationToken,
		svcPasswordResetToken: svcPasswordResetToken,
		svcLogOperation:       svcLogOperation,
		cfg:                   cfg,
	}
}

type controller struct {
	svc                   Service
	svcUser               member.Service
	svcActivationToken    activation.Service
	svcPasswordResetToken passwordreset.Service
	svcLogOperation       operation.Service
	cfg                   *application.Config
}

type Controller interface {
	RegisterMember(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	VerifyUser(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	RequestActivation(c *fiber.Ctx) error
	RefreshAccessToken(c *fiber.Ctx) error
	RequestPasswordReset(c *fiber.Ctx) error
	PasswordReset(c *fiber.Ctx) error
	ChangePassword(c *fiber.Ctx) error
}

func (ctrl *controller) RegisterMember(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*member.RegisterMemberRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	currentUserId, err := helper.InterfaceToUint(c.Locals(constant.UserId))
	if err != nil {
		return apperror.Unauthorized(constant.InvalidUserSession)
	}

	companyId, err := helper.InterfaceToUint(c.Locals(constant.CompanyId))
	if err != nil {
		return apperror.Unauthorized(constant.InvalidCompanySession)
	}

	reqBody.CompanyId = companyId
	reqBody.RoleId = uint(memberRoleId)

	if err := ctrl.svc.AddMember(currentUserId, reqBody); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(helper.SuccessResponse[any](
		fmt.Sprintf("we've sent an email to %s with a link to activate the account", reqBody.Email),
		nil,
	))
}

func (ctrl *controller) VerifyUser(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*PasswordResetRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	token := c.Params("token")
	if token == "" {
		return apperror.BadRequest("missing activation token")
	}

	if err := ctrl.svc.VerifyMember(token, reqBody); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"your account has been verified",
		nil,
	))
}

func (ctrl *controller) Logout(c *fiber.Ctx) error {
	memberId, err := helper.InterfaceToUint(c.Locals(constant.UserId))
	if err != nil {
		return apperror.Unauthorized(constant.InvalidUserSession)
	}

	companyId, err := helper.InterfaceToUint(c.Locals(constant.CompanyId))
	if err != nil {
		return apperror.Unauthorized(constant.InvalidCompanySession)
	}

	// Clear access & refresh token cookies
	clearAuthCookie(c, "aif_token")
	clearAuthCookie(c, "aif_refresh_token")

	err = ctrl.svc.Logout(memberId, companyId)
	if err != nil {
		log.Warn().Err(err).Msg("failed to log logout event")
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"succeed to logout",
		nil,
	))
}

func (ctrl *controller) RequestActivation(c *fiber.Ctx) error {
	email := c.Params("email")
	if email == "" {
		return apperror.BadRequest("missing email")
	}

	if err := ctrl.svc.RequestActivation(email); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		fmt.Sprintf("we've sent an email to %s with a link to activate the account", email),
		nil,
	))
}

func (ctrl *controller) ChangePassword(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*ChangePasswordRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	userId := fmt.Sprintf("%v", c.Locals(constant.UserId))

	if err := ctrl.svc.ChangePassword(userId, reqBody); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"succeed to change password",
		nil,
	))
}

func (ctrl *controller) RefreshAccessToken(c *fiber.Ctx) error {
	memberId, err := helper.InterfaceToUint(c.Locals(constant.UserId))
	if err != nil {
		return apperror.Unauthorized(constant.InvalidUserSession)
	}

	companyId, err := helper.InterfaceToUint(c.Locals(constant.CompanyId))
	if err != nil {
		return apperror.Unauthorized(constant.InvalidCompanySession)
	}

	roleId, err := helper.InterfaceToUint(c.Locals("tierLevel"))
	if err != nil {
		return apperror.Unauthorized("invalid tier level session")
	}

	apiKey := fmt.Sprintf("%v", c.Locals(constant.APIKey))

	accessToken, err := ctrl.svc.RefreshAccessToken(memberId, companyId, roleId, apiKey)
	if err != nil {
		return err
	}

	if err := setTokenCookie(c, "aif_token", accessToken, ctrl.cfg.Env.JwtExpiresMinutes); err != nil {
		return apperror.Internal("failed to set access token cookie", err)
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"access token refreshed",
		nil,
	))
}

func (ctrl *controller) Login(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*userLoginRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	accessToken, refreshToken, loginResp, err := ctrl.svc.LoginMember(reqBody)
	if err != nil {
		return err
	}

	const accessCookieName = "aif_token"
	const refreshCookieName = "aif_refresh_token"

	// Set access token cookie
	if err := setTokenCookie(c, accessCookieName, accessToken, ctrl.cfg.Env.JwtExpiresMinutes); err != nil {
		return apperror.Internal("failed to set access token cookie", err)
	}

	// Set refresh token cookie
	if err := setTokenCookie(c, refreshCookieName, refreshToken, ctrl.cfg.Env.JwtRefreshTokenExpiresMinutes); err != nil {
		return apperror.Internal("failed to set refresh token cookie", err)
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse(
		"succeed to login", loginResp,
	))
}

func (ctrl *controller) RequestPasswordReset(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*RequestPasswordResetRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	if err := ctrl.svc.RequestPasswordReset(reqBody.Email); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		fmt.Sprintf("we've sent an email to %s with a link to reset your password", reqBody.Email),
		nil,
	))
}

func (ctrl *controller) PasswordReset(c *fiber.Ctx) error {
	reqBody, ok := c.Locals(constant.Request).(*PasswordResetRequest)
	if !ok {
		return apperror.BadRequest(constant.InvalidRequestFormat)
	}

	token := c.Params("token")
	if token == "" {
		return apperror.BadRequest("missing password reset token")
	}

	if err := ctrl.svc.PasswordReset(token, reqBody); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(helper.SuccessResponse[any](
		"succeed to reset password",
		nil,
	))
}

func setTokenCookie(c *fiber.Ctx, name, value, durationStr string) error {
	minutes, err := strconv.Atoi(durationStr)
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(time.Duration(minutes) * time.Minute),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})

	return nil
}

func clearAuthCookie(c *fiber.Ctx, name string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax", // Atau "Strict" jika lebih aman
	})
}
