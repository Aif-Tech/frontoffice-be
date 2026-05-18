package middleware

import (
	"front-office/configs/application"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
)

func Auth() func(c *fiber.Ctx) error {
	config := jwtware.Config{
		SigningKey:   []byte(os.Getenv("FO_JWT_SECRET_KEY")),
		ErrorHandler: jwtError,
		TokenLookup:  "cookie:aif_token",
	}

	return jwtware.New(config)
}

func jwtError(c *fiber.Ctx, err error) error {
	resp := helper.ErrorResponse(err.Error())
	return c.Status(fiber.StatusUnauthorized).JSON(resp)
}

func SetCookiePasswordResetToken(c *fiber.Ctx) error {
	token := c.Params("token")
	minutesToExpired, err := strconv.Atoi(os.Getenv("FO_JWT_RESET_PASSWORD_EXPIRES_MINUTES"))
	if err != nil {
		return apperror.Internal("invalid password reset expiry config", err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "password_reset_cookie",
		Value:    token,
		Expires:  time.Now().Add(time.Duration(minutesToExpired) * time.Minute),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})

	return c.Next()
}

func GetJWTPayloadFromCookie(cfg *application.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := cfg.App.JwtSecretKey
		token := c.Cookies("aif_token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": constant.MissingAccessToken,
			})
		}

		claims, err := helper.ExtractClaimsFromJWT(token, secret)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		userId, err := helper.ExtractUserIdFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		companyId, err := helper.ExtractCompanyIdFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		roleId, err := helper.ExtractRoleIdFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		quotaType, err := helper.ExtractQuotaTypeFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		apiKey, err := helper.ExtractApiKeyFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		c.Locals(constant.UserId, userId)
		c.Locals(constant.CompanyId, companyId)
		c.Locals(constant.RoleId, roleId)
		c.Locals(constant.QuotaType, quotaType)
		c.Locals(constant.APIKey, apiKey)

		return c.Next()
	}
}

func GetPayloadFromRefreshToken(cfg *application.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := cfg.App.JwtRefreshSecretKey
		token := c.Cookies("aif_refresh_token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "no refresh token provided",
			})
		}

		claims, err := helper.ExtractClaimsFromJWT(token, secret)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		tokenType, err := helper.ExtractTokenTypeFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}
		if tokenType != constant.TokenTypeRefresh {
			return c.Status(fiber.StatusUnauthorized).JSON(helper.ErrorResponse("invalid token type"))
		}

		userId, err := helper.ExtractUserIdFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		companyId, err := helper.ExtractCompanyIdFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		roleId, err := helper.ExtractRoleIdFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		apiKey, err := helper.ExtractApiKeyFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		quotaType, err := helper.ExtractQuotaTypeFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		c.Locals(constant.UserId, userId)
		c.Locals(constant.CompanyId, companyId)
		c.Locals(constant.RoleId, roleId)
		c.Locals(constant.APIKey, apiKey)
		c.Locals(constant.QuotaType, quotaType)

		return c.Next()
	}
}

func AdminAuth(cfg *application.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := cfg.App.JwtSecretKey
		token := c.Cookies("aif_token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": constant.MissingAccessToken,
			})
		}

		claims, err := helper.ExtractClaimsFromJWT(token, secret)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		roleId, err := helper.ExtractRoleIdFromClaims(claims)
		if err != nil {
			resp := helper.ErrorResponse(err.Error())

			return c.Status(fiber.StatusBadRequest).JSON(resp)
		}
		if roleId == 2 {
			resp := helper.ErrorResponse(constant.RequestProhibited)

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		return c.Next()
	}
}
