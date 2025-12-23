package mail

import (
	"front-office/pkg/apperror"
)

type Service interface {
	Send(mail Mail) error
}

type SendMailService struct {
	service  Service
	renderer *TemplateRenderer
	queue    MailQueue
}

func NewMailService(s Service, r *TemplateRenderer, q MailQueue) *SendMailService {
	return &SendMailService{service: s, renderer: r, queue: q}
}

func (svc *SendMailService) Execute(mail Mail) error {
	if mail.To == "" {
		return apperror.BadRequest("recipient is required")
	}

	// return svc.service.Send(mail)
	return svc.queue.Enqueue(mail)
}

func (svc *SendMailService) SendWithTemplate(
	to string,
	subject string,
	templateName string,
	data any,
) error {
	body, err := svc.renderer.Render(templateName, data)
	if err != nil {
		return apperror.Internal("failed to render template", err)
	}

	// return svc.service.Send(Mail{
	// 	To:      to,
	// 	Subject: subject,
	// 	Body:    body,
	// })

	return svc.Execute(Mail{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}
