package application

import "github.com/rs/zerolog/log"

type (
	Config struct {
		App  *Environment
		Mail *Mail
	}

	Mail struct {
		Host     string
		Port     string
		User     string
		Password string
		MaxRetry string
		CCMails  string
	}
)

func GetConfig() Config {
	env := LoadEnvironment()
	if err := env.Validate(); err != nil {
		log.Fatal().Err(err).Msg("invalid environment configuration")
	}

	return Config{
		App: env,
		Mail: &Mail{
			Host:     env.MailHost,
			Port:     env.MailPort,
			User:     env.MailUser,
			Password: env.MailPass,
			MaxRetry: env.MailMaxRetry,
		},
	}
}
