package application

type (
	Config struct {
		App *Environment
		Db  Db
	}

	App struct {
		Port string
	}

	Db struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SSLMode  string
		TimeZone string
	}
)

func GetConfig() Config {
	env := LoadEnvironment()

	return Config{
		App: env,
		Db: Db{
			Host:     env.DbHost,
			Port:     env.DbPort,
			User:     env.DbUser,
			Password: env.DbPassword,
			Name:     env.DbName,
			SSLMode:  "disable",
			TimeZone: "Asia/Jakarta",
		},
	}
}
