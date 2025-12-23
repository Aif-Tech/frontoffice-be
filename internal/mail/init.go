package mail

import (
	"front-office/configs/application"

	"github.com/rs/zerolog/log"
)

func Init(cfg *application.Config) *Module {
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

	queue := NewInMemoryMailQueue(100)
	worker := NewMailWorker(queue, smtpService)

	worker.Start()

	sendMailSvc := NewMailService(smtpService, renderer, queue)

	return &Module{
		SendMail: sendMailSvc,
	}
}
