package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
)

var ctx = context.Background()

type Service interface {
	SetValue(key string, value string)
	SetHash(key string, field string, value string)
	RemHash(key string, field string)
	ReadValue(key string) string
	SendMessage(channel string, message string)
	SetListener(channel string, handler func(message *redis.Message))
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

func (n *ServiceImpl) SetValue(key string, value string) {
	// this can be fully async
	go func() {
		n.Client.Set(ctx, key, value, 0)
	}()
}

func (n *ServiceImpl) SetHash(key string, field string, value string) {
	go func() {
		n.Client.HSet(ctx, key, field, value)
	}()
}

func (n *ServiceImpl) RemHash(key string, field string) {
	go func() {
		n.Client.HDel(ctx, key, field)
	}()
}

func (n *ServiceImpl) ReadValue(key string) string {
	val, err := n.Client.Get(ctx, key).Result()
	if err != nil {
		return "" // no key found
	}
	return val
}

func (n *ServiceImpl) SendMessage(channel string, message string) {
	n.Client.Publish(ctx, channel, message)
}

func (n *ServiceImpl) SetListener(channel string, handler func(message *redis.Message)) {
	sub := n.Client.Subscribe(ctx, channel)
	ch := sub.Channel()

	// run in a goroutine to prevent blocking issues :)
	go func() {
		for msg := range ch {
			handler(msg)
		}
	}()
}
