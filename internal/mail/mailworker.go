package mail

import (
	"time"
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
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = w.queue.MoveReadyRetries()

			default:
				mail, err := w.queue.Dequeue(30)
				if err != nil {
					continue
				}

				if err := w.service.Send(mail); err != nil {
					w.handleFailure(mail)
				}
			}
		}
	}()
}

func (w *MailWorker) handleFailure(mail Mail) {
	mail.Retry++

	if mail.Retry > mail.MaxRetry {
		_ = w.queue.EnqueueDLQ(mail)
		return
	}

	delay := backoffDuration(mail.Retry)
	_ = w.queue.EnqueueRetry(mail, delay)
}

func backoffDuration(retry int) time.Duration {
	base := 10 * time.Second

	return time.Duration(1<<retry) * base
}
