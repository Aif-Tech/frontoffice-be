package mail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type MailQueue interface {
	Enqueue(mail Mail) error
	Dequeue(timeoutSeconds int) (Mail, error)
	EnqueueRetry(mail Mail, delay time.Duration) error
	EnqueueDLQ(mail Mail) error
	MoveReadyRetries() error
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

func (q *RedisMailQueue) EnqueueRetry(mail Mail, delay time.Duration) error {
	ctx := context.Background()

	payload, _ := json.Marshal(mail)
	score := time.Now().Add(delay).Unix()

	return q.client.ZAdd(ctx, "mail:queue:retry", redis.Z{
		Score:  float64(score),
		Member: payload,
	}).Err()
}

func (q *RedisMailQueue) EnqueueDLQ(mail Mail) error {
	ctx := context.Background()

	payload, _ := json.Marshal(mail)

	return q.client.LPush(ctx, "mail:queue:dlq", payload).Err()
}

func (q *RedisMailQueue) MoveReadyRetries() error {
	ctx := context.Background()
	now := time.Now().Unix()

	items, err := q.client.ZRangeByScore(
		ctx,
		"mail:queue:retry",
		&redis.ZRangeBy{
			Min: "-inf",
			Max: fmt.Sprint(now),
		},
	).Result()

	if err != nil || len(items) == 0 {
		return nil
	}

	for _, item := range items {
		q.client.ZRem(ctx, "mail:queue:retry", item)
		q.client.LPush(ctx, "mail:queue", item)
	}

	return nil
}
