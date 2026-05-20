package application

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Environment struct {
	AppEnv                         string
	CloudProvider                  string
	FrontendBaseUrl                string
	Port                           string
	MailHost                       string
	MailPass                       string
	MailPort                       string
	MailUser                       string
	MailMaxRetry                   string
	MailjetPublicKey               string
	MailjetSecretKey               string
	JwtSecretKey                   string
	JwtActivationSecretKey         string
	JwtResetPasswordSecretKey      string
	JwtRefreshSecretKey            string
	JwtExpiresMinutes              string
	JwtRefreshTokenExpiresMinutes  string
	JwtVerificationExpiresMinutes  string
	JwtActivationExpiresMinutes    string
	JwtResetPasswordExpiresMinutes string
	CoreModuleKey                  string
	PartnerServiceHost             string
	ProductCatalogHost             string
	AifcoreHost                    string
	ScoreezyHost                   string
	GenretailV3                    string
	AllowingDomains                string
	RedisAddr                      string
}

func GetEnvironment(key string) string {
	return os.Getenv(key)
}

func LoadEnvironment() *Environment {
	env := os.Getenv("FO_APP_ENV")
	if env == "" {
		log.Warn().
			Str("env_var", "FO_APP_ENV").
			Msg("environment variable not set, fallback to local")
		env = "local"
	}

	if env == "local" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("error loading .env file")
		}
	}

	return &Environment{
		AppEnv:                         GetEnvironment("FO_APP_ENV"),
		FrontendBaseUrl:                GetEnvironment("FO_FRONTEND_BASE_URL"),
		CloudProvider:                  GetEnvironment("CLOUD_PROVIDER"),
		Port:                           GetEnvironment("FO_APP_PORT"),
		MailHost:                       GetEnvironment("FO_MAIL_HOST"),
		MailMaxRetry:                   GetEnvironment("FO_MAIL_MAX_RETRY"),
		MailPort:                       GetEnvironment("FO_MAIL_PORT"),
		MailPass:                       GetEnvironment("FO_MAIL_PASSWORD"),
		MailUser:                       GetEnvironment("FO_MAIL_USER"),
		JwtSecretKey:                   GetEnvironment("FO_JWT_SECRET_KEY"),
		JwtActivationSecretKey:         GetEnvironment("FO_JWT_ACTIVATION_SECRET_KEY"),
		JwtResetPasswordSecretKey:      GetEnvironment("FO_JWT_RESET_PASSWORD_SECRET_KEY"),
		JwtRefreshSecretKey:            GetEnvironment("FO_JWT_REFRESH_SECRET_KEY"),
		JwtExpiresMinutes:              GetEnvironment("FO_JWT_EXPIRES_MINUTES"),
		JwtRefreshTokenExpiresMinutes:  GetEnvironment("FO_JWT_REFRESH_EXPIRES_MINUTES"),
		JwtVerificationExpiresMinutes:  GetEnvironment("FO_JWT_VERIFICATION_EXPIRES_MINUTES"),
		JwtActivationExpiresMinutes:    GetEnvironment("FO_JWT_ACTIVATION_EXPIRES_MINUTES"),
		JwtResetPasswordExpiresMinutes: GetEnvironment("FO_JWT_RESET_PASSWORD_EXPIRES_MINUTES"),
		ProductCatalogHost:             GetEnvironment("FO_DATAHUB_HOST"),
		AifcoreHost:                    GetEnvironment("FO_CORE_HOST"),
		ScoreezyHost:                   GetEnvironment("FO_SCOREEZY_HOST"),
		CoreModuleKey:                  GetEnvironment("FO_CORE_KEY"),
		RedisAddr:                      GetEnvironment("FO_REDIS_URL"),
	}
}

func (e *Environment) Validate() error {
	var missing []string

	required := map[string]string{
		"FO_APP_ENV":                            e.AppEnv,
		"FO_FRONTEND_BASE_URL":                  e.FrontendBaseUrl,
		"FO_APP_PORT":                           e.Port,
		"FO_MAIL_HOST":                          e.MailHost,
		"FO_MAIL_MAX_RETRY":                     e.MailMaxRetry,
		"FO_MAIL_PORT":                          e.MailPort,
		"FO_MAIL_PASSWORD":                      e.MailPass,
		"FO_MAIL_USER":                          e.MailUser,
		"FO_JWT_SECRET_KEY":                     e.JwtSecretKey,
		"FO_JWT_ACTIVATION_SECRET_KEY":          e.JwtActivationSecretKey,
		"FO_JWT_RESET_PASSWORD_SECRET_KEY":      e.JwtResetPasswordSecretKey,
		"FO_JWT_REFRESH_SECRET_KEY":             e.JwtRefreshSecretKey,
		"FO_JWT_EXPIRES_MINUTES":                e.JwtExpiresMinutes,
		"FO_JWT_REFRESH_EXPIRES_MINUTES":        e.JwtRefreshTokenExpiresMinutes,
		"FO_JWT_VERIFICATION_EXPIRES_MINUTES":   e.JwtVerificationExpiresMinutes,
		"FO_JWT_ACTIVATION_EXPIRES_MINUTES":     e.JwtActivationExpiresMinutes,
		"FO_JWT_RESET_PASSWORD_EXPIRES_MINUTES": e.JwtResetPasswordExpiresMinutes,
		"FO_DATAHUB_HOST":                       e.ProductCatalogHost,
		"FO_CORE_HOST":                          e.AifcoreHost,
		"FO_SCOREEZY_HOST":                      e.ScoreezyHost,
		"FO_CORE_KEY":                           e.CoreModuleKey,
		"FO_REDIS_URL":                          e.RedisAddr,
	}

	for key, val := range required {
		if strings.TrimSpace(val) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}
