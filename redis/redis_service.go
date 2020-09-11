package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type DataStore interface {
	GenerateUserId() uint64
}

type service struct {
	*redis.Client
}

var ctx = context.Background()
var RedisService DataStore
var UserIdKey = "user:id"

func (s *service) GenerateUserId() uint64 {
	incr := s.Client.Incr(ctx, UserIdKey)
	return uint64(incr.Val())
}

func init() {
	var (
		//TODO: set from env
		host     = "localhost"
		port     = "6379"
		password = ""
	)

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})
	_ = client.SetNX(ctx, UserIdKey, 0, 0)
	RedisService = &service{client}
}
