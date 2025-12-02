package application

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Environment struct {
	Env                            string
	CloudProvider                  string
	FrontendBaseUrl                string
	Port                           string
	DbUser                         string
	DbPassword                     string
	DbName                         string
	DbPort                         string
	DbHost                         string
	MailjetEmail                   string
	MailtjetUsername               string
	MailjetPublicKey               string
	MailjetSecretKey               string
	JwtSecretKey                   string
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
		Env:                            GetEnvironment("FO_APP_ENV"),
		FrontendBaseUrl:                GetEnvironment("FO_FRONTEND_BASE_URL"),
		CloudProvider:                  GetEnvironment("CLOUD_PROVIDER"),
		Port:                           GetEnvironment("FO_APP_PORT"),
		DbUser:                         GetEnvironment("DB_USER"),
		DbPassword:                     GetEnvironment("DB_PASSWORD"),
		DbName:                         GetEnvironment("DB_NAME"),
		DbPort:                         GetEnvironment("DB_PORT"),
		DbHost:                         GetEnvironment("DB_HOST"),
		MailjetEmail:                   GetEnvironment("FO_MAILJET_EMAIL"),
		MailtjetUsername:               GetEnvironment("FO_MAILJET_USERNAME"),
		MailjetPublicKey:               GetEnvironment("FO_MAILJET_PUBLIC_KEY"),
		MailjetSecretKey:               GetEnvironment("FO_MAILJET_SECRET_KEY"),
		JwtSecretKey:                   GetEnvironment("FO_JWT_SECRET_KEY"),
		JwtExpiresMinutes:              GetEnvironment("FO_JWT_EXPIRES_MINUTES"),
		JwtRefreshTokenExpiresMinutes:  GetEnvironment("FO_JWT_REFRESH_EXPIRES_MINUTES"),
		JwtVerificationExpiresMinutes:  GetEnvironment("FO_JWT_VERIFICATION_EXPIRES_MINUTES"),
		JwtActivationExpiresMinutes:    GetEnvironment("FO_JWT_ACTIVATION_EXPIRES_MINUTES"),
		JwtResetPasswordExpiresMinutes: GetEnvironment("FO_JWT_RESET_PASSWORD_EXPIRES_MINUTES"),
		PartnerServiceHost:             GetEnvironment("PARTNER_SERVICE_HOST"),
		ProductCatalogHost:             GetEnvironment("FO_DATAHUB_HOST"),
		AifcoreHost:                    GetEnvironment("FO_CORE_HOST"),
		ScoreezyHost:                   GetEnvironment("FO_SCOREEZY_HOST"),
		AllowingDomains:                GetEnvironment("ALLOWING_DOMAINS"),
		CoreModuleKey:                  GetEnvironment("FO_CORE_KEY"),
	}
}
