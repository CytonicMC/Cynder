package messaging

import (
	"github.com/nats-io/nats.go"
	"time"
)

type NatsServiceImpl struct {
	Conn *nats.Conn
}

func (n *NatsServiceImpl) Subscribe(subject string, handler nats.MsgHandler) error {
	_, err := n.Conn.Subscribe(subject, handler)
	return err
}

func (n *NatsServiceImpl) Publish(subject string, msg []byte) error {
	return n.Conn.Publish(subject, msg)
}

func (n *NatsServiceImpl) Request(subject string, msg []byte, duration time.Duration) (*nats.Msg, error) {
	return n.Conn.Request(subject, msg, duration)
}
