package mail

import (
	"context"
	"encoding/json"
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
	client   *redis.Client
	key      string
	retryKey string
	dlqKey   string
}

func NewRedisMailQueue(client *redis.Client) *RedisMailQueue {
	return &RedisMailQueue{
		client:   client,
		key:      "mail:queue",
		retryKey: "mail:queue:retry",
		dlqKey:   "mail:queue:dlq",
	}
}

func (q *RedisMailQueue) Enqueue(mail Mail) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	payload, err := json.Marshal(mail)
	if err != nil {
		return fmt.Errorf("failed to marshal mail: %w", err)
	}

	return q.client.RPush(ctx, q.key, payload).Err()
}

func (q *RedisMailQueue) Dequeue(timeoutSeconds int) (Mail, error) {
	result, err := q.client.BRPop(
		context.Background(),
		time.Duration(timeoutSeconds)*time.Second,
		q.key,
	).Result()
	if err != nil {
		return Mail{}, err
	}

	if len(result) < 2 {
		return Mail{}, fmt.Errorf("unexpected BRPop result")
	}

	var mail Mail
	if err := json.Unmarshal([]byte(result[1]), &mail); err != nil {
		return Mail{}, fmt.Errorf("failed to unmarshal mail: %w", err)
	}

	return mail, nil
}

func (q *RedisMailQueue) EnqueueRetry(mail Mail, delay time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	payload, err := json.Marshal(mail)
	if err != nil {
		return fmt.Errorf("failed to marshal mail for retry: %w", err)
	}

	score := float64(time.Now().Add(delay).Unix())

	return q.client.ZAdd(ctx, q.retryKey, redis.Z{Score: score, Member: payload}).Err()
}

func (q *RedisMailQueue) EnqueueDLQ(mail Mail) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	payload, err := json.Marshal(mail)
	if err != nil {
		return fmt.Errorf("failed to marshal mail for DLQ: %w", err)
	}

	return q.client.RPush(ctx, q.dlqKey, payload).Err()
}

func (q *RedisMailQueue) MoveReadyRetries() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	members, err := q.client.ZRangeByScore(ctx, q.retryKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", time.Now().Unix()),
	}).Result()
	if err != nil || len(members) == 0 {
		return err
	}

	pipe := q.client.Pipeline()
	for _, m := range members {
		pipe.RPush(ctx, q.key, m)
		pipe.ZRem(ctx, q.retryKey, m)
	}

	_, err = pipe.Exec(ctx)

	return err
}
