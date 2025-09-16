package middleware

import (
	"fmt"
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
	resp := helper.ResponseFailed(err.Error())
	return c.Status(fiber.StatusUnauthorized).JSON(resp)
}

func SetHeaderAuth(c *fiber.Ctx) error {
	token := c.Params("token")
	c.Request().Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	return c.Next()
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

func GetJWTPayloadFromCookie() fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := os.Getenv("FO_JWT_SECRET_KEY")
		token := c.Cookies("aif_token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": constant.MissingAccessToken,
			})
		}

		claims, err := helper.ExtractClaimsFromJWT(token, secret)
		if err != nil {
			resp := helper.ResponseFailed(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		userId, err := helper.ExtractUserIdFromClaims(claims)
		if err != nil {
			statusCode, resp := helper.GetError(err.Error())
			return c.Status(statusCode).JSON(resp)
		}

		companyId, err := helper.ExtractCompanyIdFromClaims(claims)
		if err != nil {
			resp := helper.ResponseFailed(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		roleId, err := helper.ExtractRoleIdFromClaims(claims)
		if err != nil {
			resp := helper.ResponseFailed(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		apiKey, err := helper.ExtractApiKeyFromClaims(claims)
		if err != nil {
			resp := helper.ResponseFailed(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		c.Locals(constant.UserId, userId)
		c.Locals(constant.CompanyId, companyId)
		c.Locals(constant.RoleId, roleId)
		c.Locals(constant.APIKey, apiKey)

		return c.Next()
	}
}

func GetJWTPayloadPasswordResetFromCookie() fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := os.Getenv("FO_JWT_SECRET_KEY")
		token := c.Cookies("password_reset_cookie")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": constant.MissingAccessToken,
			})
		}

		claims, err := helper.ExtractClaimsFromJWT(token, secret)
		if err != nil {
			resp := helper.ResponseFailed(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		userId, err := helper.ExtractUserIdFromClaims(claims)
		if err != nil {
			statusCode, resp := helper.GetError(err.Error())
			return c.Status(statusCode).JSON(resp)
		}

		companyId, err := helper.ExtractUserIdFromClaims(claims)
		if err != nil {
			resp := helper.ResponseFailed(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		tierLevel, err := helper.ExtractUserIdFromClaims(claims)
		if err != nil {
			resp := helper.ResponseFailed(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		c.Locals(constant.UserId, userId)
		c.Locals(constant.CompanyId, companyId)
		c.Locals("tierLevel", tierLevel)

		return c.Next()
	}
}

func GetPayloadFromRefreshToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := os.Getenv("FO_JWT_SECRET_KEY")
		token := c.Cookies("aif_refresh_token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "no refresh token provided",
			})
		}

		claims, err := helper.ExtractClaimsFromJWT(token, secret)
		if err != nil {
			resp := helper.ResponseFailed(err.Error())
			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		userId, err := helper.ExtractUserIdFromClaims(claims)
		if err != nil {
			statusCode, resp := helper.GetError(err.Error())
			return c.Status(statusCode).JSON(resp)
		}

		companyId, err := helper.ExtractCompanyIdFromClaims(claims)
		if err != nil {
			resp := helper.ResponseFailed(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		roleId, err := helper.ExtractRoleIdFromClaims(claims)
		if err != nil {
			resp := helper.ResponseFailed(err.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		c.Locals(constant.UserId, userId)
		c.Locals(constant.CompanyId, companyId)
		c.Locals(constant.RoleId, roleId)

		return c.Next()
	}
}

func AdminAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := os.Getenv("FO_JWT_SECRET_KEY")
		token := c.Cookies("aif_token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": constant.MissingAccessToken,
			})
		}

		claims, err := helper.ExtractClaimsFromJWT(token, secret)
		if err != nil {
			statusCode, resp := helper.GetError(err.Error())
			return c.Status(statusCode).JSON(resp)
		}

		roleId, err := helper.ExtractRoleIdFromClaims(claims)
		if err != nil {
			resp := helper.ResponseFailed(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(resp)
		}
		if roleId == 2 {
			resp := helper.ResponseFailed(constant.RequestProhibited)
			return c.Status(fiber.StatusUnauthorized).JSON(resp)
		}

		return c.Next()
	}
}
