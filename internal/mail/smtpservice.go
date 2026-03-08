package mail

import (
	"bytes"
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

func (s *SMTPService) Send(mail Mail) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("AIForesee <%s>", s.user)

	if len(mail.ToList) > 0 {
		e.To = mail.ToList
	} else if mail.To != "" {
		e.To = []string{mail.To}
	}

	if len(mail.CC) > 0 {
		e.Cc = mail.CC
	}

	e.Subject = mail.Subject
	e.HTML = []byte(mail.Body)

	for _, att := range mail.Attachments {
		if _, err := e.Attach(
			bytes.NewReader(att.Data),
			att.FileName,
			att.MimeType,
		); err != nil {
			return fmt.Errorf("failed to attach file %s: %w", att.FileName, err)
		}
	}

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
