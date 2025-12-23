package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

type SMTPService struct {
	host string
	port string
	user string
	pass string
}

func NewSMTPService(host, port, user, pass string) *SMTPService {
	return &SMTPService{host, port, user, pass}
}

func (s *SMTPService) Send(ctx context.Context, mail Email) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("AIForesee <%s>", s.user)
	e.To = []string{mail.To}
	e.Subject = mail.Subject
	e.HTML = []byte(mail.Body)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	tlsConfig := &tls.Config{
		ServerName: s.host,
		MinVersion: tls.VersionTLS12,
	}

	if s.port == "587" {
		return e.SendWithStartTLS(addr, auth, tlsConfig)
	}

	if s.port == "465" {
		return e.SendWithTLS(addr, auth, tlsConfig)
	}

	return fmt.Errorf("unsupported SMTP port: %s", s.port)
}
