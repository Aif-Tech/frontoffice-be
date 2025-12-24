package mail

import (
	"github.com/rs/zerolog/log"
)

type MailWorker struct {
	queue   MailQueue
	service Service
}

func NewMailWorker(q MailQueue, s Service) *MailWorker {
	return &MailWorker{queue: q, service: s}
}

func (w *MailWorker) Start() {
	go func() {
		for {
			mail, err := w.queue.Dequeue(30)
			if err != nil {
				continue
			}

			if err := w.service.Send(mail); err != nil {
				log.Warn().
					Err(err).
					Msg("failed to send mail")

				// todo: retry
			}
		}
	}()
}
