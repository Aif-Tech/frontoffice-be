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
			case <-w.ctx.Done():
				log.Info().Msg("[mail-worker] stopped")

				return
			case <-ticker.C:
				if err := w.queue.MoveReadyRetries(); err != nil {
					log.Warn().Err(err).Msg("[mail-worker] failed to move ready retries")
				}

			default:
				mail, err := w.queue.Dequeue(5)
				if err != nil {
					time.Sleep(200 * time.Millisecond)
					continue
				}

				if err := w.service.Send(mail); err != nil {
					log.Error().Err(err).Str("to", mail.To).Msg("[mail-worker] failed to send mail")
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

	if mail.Retry >= mail.MaxRetry {
		if err := w.queue.EnqueueDLQ(mail); err != nil {
			log.Error().Err(err).Str("to", mail.To).Msg("[mail-worker] failed to enqueue DLQ")
		}
		log.Warn().Str("to", mail.To).Msg("[mail-worker] mail moved to DLQ")

		return
	}
	// Exponential backoff: retry ke-1 = 30s, ke-2 = 120s, ke-3 = 270s, dst.
	delay := time.Duration(mail.Retry*mail.Retry) * 30 * time.Second
	if err := w.queue.EnqueueRetry(mail, delay); err != nil {
		log.Error().Err(err).Str("to", mail.To).Msg("[mail-worker] failed to enqueue retry")
	}

	log.Warn().Str("to", mail.To).Int("retry", mail.Retry).Dur("delay", delay).
		Msg("[mail-worker] mail scheduled for retry")
}
