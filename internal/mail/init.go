package mail

import (
	"front-office/configs/application"

	redisinfra "front-office/internal/infra/redis"

	"github.com/rs/zerolog/log"
)

func Init(cfg *application.Config) *MailModule {
	redisClient, err := redisinfra.NewRedisClient(
		cfg.App.AppEnv,
		cfg.App.RedisAddr,
	)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to create redis connection")
	}

	renderer, err := NewTemplateRenderer(
		"internal/mail/template",
	)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to execute mail service")
	}

	smtpService := NewSMTPService(
		cfg.Mail.Host,
		cfg.Mail.Port,
		cfg.App.MailUser,
		cfg.Mail.Password,
	)

	queue := NewRedisMailQueue(redisClient)
	worker := NewMailWorker(queue, smtpService)

	worker.Start()

	sendMailSvc := NewMailService(smtpService, renderer, queue, cfg.Mail.MaxRetry)

	return &MailModule{
		SendMail: sendMailSvc,
		Worker:   worker,
	}
}
