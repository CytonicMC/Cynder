package messaging

import (
	"time"

	"github.com/CytonicMC/Cynder/cynder/env"
	"github.com/nats-io/nats.go"
)

type NatsServiceImpl struct {
	Conn *nats.Conn
}

func (n *NatsServiceImpl) Subscribe(subject string, handler nats.MsgHandler) error {
	prefixed := env.EnsurePrefixed(subject)
	_, err := n.Conn.Subscribe(prefixed, handler)
	return err
}

func (n *NatsServiceImpl) Publish(subject string, msg []byte) error {
	prefixed := env.EnsurePrefixed(subject)
	return n.Conn.Publish(prefixed, msg)
}

func (n *NatsServiceImpl) Request(subject string, msg []byte, duration time.Duration) (*nats.Msg, error) {
	prefixed := env.EnsurePrefixed(subject)
	return n.Conn.Request(prefixed, msg, duration)
}
