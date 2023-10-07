package database

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisOptions struct {
	Addr     string
	Password string
	DB       int
}

var Ctx = context.Background()

func NewRedisClient(options RedisOptions) *redis.Client {
	opt := redis.Options{
		Addr:     options.Addr,
		Password: options.Password,
		DB:       options.DB,
	}

	return redis.NewClient(&opt)
}
