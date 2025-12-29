package mail

type Mail struct {
	To       string
	Subject  string
	Body     string
	Retry    int
	MaxRetry int
}

type MailModule struct {
	SendMail *SendMailService
	Worker   *MailWorker
}
