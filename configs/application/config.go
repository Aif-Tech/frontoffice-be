package application

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
	}
)

func GetConfig() Config {
	env := LoadEnvironment()

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
