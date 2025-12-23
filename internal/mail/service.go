package mail

import (
	"context"
	"front-office/pkg/apperror"
)

type Service interface {
	Send(ctx context.Context, email Email) error
}

type SendMailService struct {
	service  Service
	renderer *TemplateRenderer
}

func NewMailService(s Service, r *TemplateRenderer) *SendMailService {
	return &SendMailService{service: s, renderer: r}
}

func (svc *SendMailService) Execute(ctx context.Context, email Email) error {
	if email.To == "" {
		return apperror.BadRequest("recipient is required")
	}

	return svc.service.Send(ctx, email)
}

func (svc *SendMailService) SendWithTemplate(
	ctx context.Context,
	to string,
	subject string,
	templateName string,
	data any,
) error {
	body, err := svc.renderer.Render(templateName, data)
	if err != nil {
		return err
	}

	return svc.service.Send(ctx, Email{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}
