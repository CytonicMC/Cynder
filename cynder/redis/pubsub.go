package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
)

var ctx = context.Background()
var client *redis.Client

func PubsubClient() {
	// todo: make this use env vars or something.
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func SetValue(key string, value string) {
	// this can be fully async
	go func() {
		client.Set(ctx, key, value, 0)
	}()
}

func ReadValue(key string) string {
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return "" // no key found
	}
	return val
}

func SendMessage(channel string, message string) {
	client.Publish(ctx, channel, message)
}

func SetListener(channel string, handler func(message *redis.Message)) {
	sub := client.Subscribe(ctx, channel)
	ch := sub.Channel()

	// run in a goroutine to prevent blocking issues :)
	go func() {
		for msg := range ch {
			handler(msg)
		}
	}()
}
