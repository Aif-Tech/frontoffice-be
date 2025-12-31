package mail

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type MailWorker struct {
	queue   MailQueue
	service Service
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewMailWorker(q MailQueue, s Service) *MailWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &MailWorker{
		queue:   q,
		service: s,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (w *MailWorker) Start() {
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

		log.Info().Msg("[mail-worker] started")

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {

			// shutdown signal
			case <-w.ctx.Done():
				return

			// move retry mails
			case <-ticker.C:
				_ = w.queue.MoveReadyRetries()

			// main dequeue loop
			default:
				mail, err := w.queue.Dequeue(120)
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

func (w *MailWorker) Stop() {
	w.cancel()
	w.wg.Wait()
	log.Info().Msg("[mail worker] stopped")
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
