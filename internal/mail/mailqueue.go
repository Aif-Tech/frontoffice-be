package mail

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type MailQueue interface {
	Enqueue(mail Mail) error
	Dequeue(timeoutSeconds int) (Mail, error)
}

type RedisMailQueue struct {
	client *redis.Client
	key    string
}

func NewRedisMailQueue(client *redis.Client) *RedisMailQueue {
	return &RedisMailQueue{
		client: client,
		key:    "mail:queue",
	}
}

func (q *RedisMailQueue) Enqueue(mail Mail) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	payload, err := json.Marshal(mail)
	if err != nil {
		return err
	}

	return q.client.LPush(ctx, q.key, payload).Err()
}

func (q *RedisMailQueue) Dequeue(timeoutSeconds int) (Mail, error) {
	ctx := context.Background()

	result, err := q.client.BRPop(
		ctx,
		time.Duration(timeoutSeconds)*time.Second,
		q.key,
	).Result()

	if err == redis.Nil {
		return Mail{}, errors.New("queue empty")
	}

	if err != nil {
		return Mail{}, err
	}

	var mail Mail
	if err := json.Unmarshal([]byte(result[1]), &mail); err != nil {
		return Mail{}, err
	}

	return mail, nil
}
