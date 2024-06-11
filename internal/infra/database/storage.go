package database

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type StorageStrategy interface {
	Create(key string, value interface{}, ttl time.Duration) error
	Exists(key string) (bool, error)
}

type RedisClient struct {
	*redis.Client
}

func LoadStorage(options *redis.Options) *RedisClient {
	return &RedisClient{redis.NewClient(options)}
}

func (r *RedisClient) Create(key string, value interface{}, ttl time.Duration) error {
	return r.Set(context.Background(), key, value, ttl).Err()
}

func (r *RedisClient) Exists(ip string, token string) (bool, error) {
	ctx := context.Background()
	ipExists, err := r.Client.Exists(ctx, ip).Result()
	if err != nil {
		return false, err
	}

	tokenExists, err := r.Client.Exists(ctx, token).Result()
	if err != nil {
		return false, err
	}

	return ipExists > 0 || tokenExists > 0, nil
}
