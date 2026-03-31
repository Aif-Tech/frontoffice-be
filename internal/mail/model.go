package mail

type Mail struct {
	To          string
	ToList      []string
	CC          []string
	Subject     string
	Body        string
	Retry       int
	MaxRetry    int
	Attachments []MailAttachment
}

type MailAttachment struct {
	FileName string
	Content  []byte
	MimeType string
}

type MailModule struct {
	SendMail *SendMailService
	Worker   *MailWorker
}
