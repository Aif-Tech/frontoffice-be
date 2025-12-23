package mail

type Mail struct {
	To      string
	Subject string
	Body    string
}

type Module struct {
	SendMail *SendMailService
}
