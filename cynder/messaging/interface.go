package messaging

import (
	"github.com/nats-io/nats.go"
	"time"
)

type NatsService interface {
	Subscribe(subject string, handler nats.MsgHandler) error
	Publish(subject string, msg []byte) error
	Request(subject string, msg []byte, duration time.Duration) (*nats.Msg, error)
}
