package mail

import (
	"front-office/pkg/apperror"
	"strconv"
)

type Service interface {
	Send(mail Mail) error
}

type SendMailService struct {
	service  Service
	renderer *TemplateRenderer
	queue    MailQueue
	maxRetry string
}

func NewMailService(s Service, r *TemplateRenderer, q MailQueue, maxRetry string) *SendMailService {
	return &SendMailService{service: s, renderer: r, queue: q, maxRetry: maxRetry}
}

func (svc *SendMailService) Execute(mail Mail) error {
	mail.Retry = 0
	maxRetry, err := strconv.Atoi(svc.maxRetry)
	if err != nil {
		return err
	}

	if mail.MaxRetry == 0 {
		mail.MaxRetry = maxRetry
	}

	if mail.To == "" && len(mail.ToList) == 0 {
		return apperror.BadRequest("recipient is required")
	}

	if mail.To != "" && len(mail.ToList) == 0 {
		mail.ToList = []string{mail.To}
	}

	// return svc.service.Send(mail)
	return svc.queue.Enqueue(mail)
}

func (svc *SendMailService) SendWithTemplate(
	toList []string,
	cc []string,
	subject string,
	templateName string,
	data any,
	attachment []MailAttachment,
) error {
	body, err := svc.renderer.Render(templateName, data)
	if err != nil {
		return apperror.Internal("failed to render template", err)
	}

	return svc.Execute(Mail{
		ToList:      toList,
		CC:          cc,
		Subject:     subject,
		Body:        body,
		Attachments: attachment,
	})
}
