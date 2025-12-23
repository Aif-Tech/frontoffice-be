package mail

type MailQueue interface {
	Enqueue(mail Mail) error
	Dequeue() (Mail, error)
}

type InMemoryMailQueue struct {
	ch chan Mail
}

func NewInMemoryMailQueue(buffer int) *InMemoryMailQueue {
	return &InMemoryMailQueue{
		ch: make(chan Mail, buffer),
	}
}

func (q *InMemoryMailQueue) Enqueue(mail Mail) error {
	q.ch <- mail

	return nil
}

func (q *InMemoryMailQueue) Dequeue() (Mail, error) {
	return <-q.ch, nil
}
