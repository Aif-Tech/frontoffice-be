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
	Data     []byte
	MimeType string
}

type MailModule struct {
	SendMail *SendMailService
	Worker   *MailWorker
}

const (
	MimeXlsx = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	MimePdf  = "application/pdf"
	MimeCsv  = "text/csv"
)
