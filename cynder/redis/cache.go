package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/CytonicMC/Cynder/cynder/env"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Service interface {
	SetValuePrefixed(key string, value string)
	SetExpiringValuePrefixed(key string, value string, expiry int64)
	SetHashPrefixed(key string, field string, value string)
	RemHashPrefixed(key string, field string)
	ReadValuePrefixed(key string) string
}

type ServiceImpl struct {
	Client *redis.Client
}

func ConnectToRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func (n *ServiceImpl) SetValuePrefixed(key string, value string) {
	// this can be fully async
	go func() {
		prefixed := env.EnsurePrefixed(key)
		n.Client.Set(ctx, prefixed, value, 0)
	}()
}

func (n *ServiceImpl) SetExpiringValuePrefixed(key string, value string, expiry int64) {
	// this can be fully async
	go func() {
		prefixed := env.EnsurePrefixed(key)
		n.Client.Set(ctx, prefixed, value, time.Duration(expiry)*time.Second)
	}()
}

func (n *ServiceImpl) SetHashPrefixed(key string, field string, value string) {
	go func() {
		prefixed := env.EnsurePrefixed(key)
		n.Client.HSet(ctx, prefixed, field, value)
	}()
}

func (n *ServiceImpl) RemHashPrefixed(key string, field string) {
	go func() {
		prefixed := env.EnsurePrefixed(key)
		n.Client.HDel(ctx, prefixed, field)
	}()
}

func (n *ServiceImpl) ReadValuePrefix(key string) string {
	prefixed := env.EnsurePrefixed(key)
	val, err := n.Client.Get(ctx, prefixed).Result()
	if err != nil {
		return "" // no key found
	}
	return val
}
