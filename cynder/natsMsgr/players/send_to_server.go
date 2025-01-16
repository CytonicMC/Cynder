package players

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/CytonicMC/Cynder/cynder/messaging"
	"github.com/nats-io/nats.go"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type SendPlayerToServerContainer struct {
	Player   uuid.UUID `json:"player"`
	ServerID string    `json:"serverId"`
	Instance uuid.UUID `json:"instance"`
}

func HandlePlayerSend(subscriber messaging.NatsService, proxy *proxy.Proxy, ctx context.Context) {
	err := subscriber.Subscribe("players.send", func(msg *nats.Msg) {
		data := string(msg.Data)
		var container SendPlayerToServerContainer
		err := json.Unmarshal([]byte(data), &container)

		if err != nil {
			fmt.Println(err)
			return
		}

		player := proxy.Player(container.Player)

		if player == nil {
			// player not on this proxy
			return
		}

		server := proxy.Server(container.ServerID)
		//todo: convey the instance somehow. Perhaps spoofchat?? Or cookies
		player.IdentifiedKey()
		player.CreateConnectionRequest(server).ConnectWithIndication(ctx)
	})
	if err != nil {
		fmt.Printf("error subscribing to subject player.send: %s", err)
	}
}
