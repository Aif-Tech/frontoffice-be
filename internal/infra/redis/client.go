package redis

import (
	"crypto/tls"

	"github.com/redis/go-redis/v9"
)

func NewUpstashClient(url, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:      url,
		Password:  password,
		TLSConfig: &tls.Config{},
	},
	)
}
