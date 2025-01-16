package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var client *redis.Client

func PubsubClient() {
	// todo: make this use env vars or something.
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
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
