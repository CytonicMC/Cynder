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
	SetValueGlobal(key string, value string)
	SetExpiringValueGlobal(key string, value string, expiry int64)
	SetHashGlobal(key string, field string, value string)
	RemHashGlobal(key string, field string)
	ReadValueGlobal(key string) string
	ReadHashGlobal(key string, field string) string
	SetContainsGlobal(key string, val string) bool
	SetContainsPrefixed(key string, val string) bool
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

func (n *ServiceImpl) ReadValuePrefixed(key string) string {
	prefixed := env.EnsurePrefixed(key)
	val, err := n.Client.Get(ctx, prefixed).Result()
	if err != nil {
		return "" // no key found
	}
	return val
}

func (n *ServiceImpl) SetValueGlobal(key string, value string) {
	// this can be fully async
	go func() {
		n.Client.Set(ctx, key, value, 0)
	}()
}

func (n *ServiceImpl) SetExpiringValueGlobal(key string, value string, expiry int64) {
	// this can be fully async
	go func() {
		n.Client.Set(ctx, key, value, time.Duration(expiry)*time.Second)
	}()
}

func (n *ServiceImpl) SetHashGlobal(key string, field string, value string) {
	go func() {
		n.Client.HSet(ctx, key, field, value)
	}()
}

func (n *ServiceImpl) RemHashGlobal(key string, field string) {
	go func() {
		n.Client.HDel(ctx, key, field)
	}()
}

func (n *ServiceImpl) ReadValueGlobal(key string) string {
	val, err := n.Client.Get(ctx, key).Result()
	if err != nil {
		return "" // no key found
	}
	return val
}

func (n *ServiceImpl) ReadHashGlobal(key string, field string) string {
	val, err := n.Client.HGet(ctx, key, field).Result()
	if err != nil {
		return "" // no key found
	}
	return val
}

func (n *ServiceImpl) SetContainsPrefixed(key string, val string) bool {
	vals, err := n.Client.SMembers(ctx, env.EnsurePrefixed(key)).Result()
	if err != nil {
		return false // no key found
	}
	return contains(vals, val)
}

func (n *ServiceImpl) SetContainsGlobal(key string, val string) bool {
	vals, err := n.Client.SMembers(ctx, key).Result()
	if err != nil {
		return false // no key found
	}
	return contains(vals, val)
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
