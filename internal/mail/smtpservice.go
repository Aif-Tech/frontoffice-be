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
	e.To = []string{mail.To}
	e.Cc = mail.CC

	e.Subject = mail.Subject
	e.HTML = []byte(mail.Body)

	// attachment
	for _, att := range mail.Attachments {
		mimeType := att.MimeType
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		_, err := e.Attach(
			bytes.NewReader(att.Content),
			att.FileName,
			mimeType,
		)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", att.FileName, err)
		}
	}

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	fmt.Println("!!!!!", addr)

	switch s.port {
	case "587":
		auth := smtp.PlainAuth("", s.user, s.pass, s.host)
		tlsConfig := &tls.Config{
			ServerName: s.host,
			MinVersion: tls.VersionTLS12,
		}
		return e.SendWithStartTLS(addr, auth, tlsConfig)
	case "465":
		auth := smtp.PlainAuth("", s.user, s.pass, s.host)
		tlsConfig := &tls.Config{
			ServerName: s.host,
			MinVersion: tls.VersionTLS12,
		}
		return e.SendWithTLS(addr, auth, tlsConfig)
	default:
		// plain SMTP tanpa TLS, cocok untuk local dev (MailHog dll)
		var auth smtp.Auth
		if s.user != "" {
			auth = smtp.PlainAuth("", s.user, s.pass, s.host)
		}
		return e.Send(addr, auth)
	}
}
