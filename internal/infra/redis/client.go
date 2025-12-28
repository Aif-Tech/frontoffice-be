package redis

import (
	"crypto/tls"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(appEnv, url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	if appEnv != "local" {
		opt.TLSConfig = &tls.Config{}
	}

	return redis.NewClient(opt), nil
}
