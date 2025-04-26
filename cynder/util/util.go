package util

import (
	"github.com/CytonicMC/Cynder/cynder/messaging"
	"github.com/CytonicMC/Cynder/cynder/redis"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/component/codec/legacy"
)

func Join(components ...c.Component) *c.Text {
	return &c.Text{Extra: components}
}

// Text converts a styled chat message like "&cHello &lWorld!" to a component.
func Text(content string) c.Component {
	legacyCodec := &legacy.Legacy{Char: legacy.AmpersandChar}
	text, _ := legacyCodec.Unmarshal([]byte(content))
	return text
}

type Services struct {
	Nats  messaging.NatsService
	Redis redis.Service
}
