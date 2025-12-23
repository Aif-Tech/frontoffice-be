package mail

type Email struct {
	To      string
	Subject string
	Body    string
}

type Module struct {
	SendMail *SendMailService
}
