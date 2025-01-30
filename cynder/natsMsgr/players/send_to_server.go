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
		connect, err := player.CreateConnectionRequest(server).Connect(ctx)
		if err != nil {
			fmt.Println(err)
			respond(msg, messaging.ServerSendResponse{
				Success: false,
				Message: "ERR_CONNECT_REQUEST_FAILED",
			})
		} else if connect.Status().ServerDisconnected() {
			respond(msg, messaging.ServerSendResponse{
				Success: false,
				Message: "ERR_SERVER_DISCONNECT",
			})
		} else if connect.Status().Canceled() {
			respond(msg, messaging.ServerSendResponse{
				Success: false,
				Message: "ERR_CONNECTION_CANCELED",
			})
		} else if connect.Status().AlreadyConnected() {
			respond(msg, messaging.ServerSendResponse{
				Success: false,
				Message: "ERR_ALREADY_CONNECTED",
			})
		} else if connect.Status().ConnectionInProgress() {
			respond(msg, messaging.ServerSendResponse{
				Success: false,
				Message: "ERR_DIFFERENT_CONNECTION_IN_PROGRESS",
			})
		} else if connect.Status().Successful() {
			respond(msg, messaging.ServerSendResponse{
				Success: true,
				Message: "",
			})
		}
	})
	if err != nil {
		fmt.Printf("error subscribing to subject player.send: %s", err)
	}
}

func respond(msg *nats.Msg, response messaging.ServerSendResponse) {
	reponse, err1 := json.Marshal(&response)
	if err1 != nil {
		fmt.Println(err1) // aka ya done did gonna be kicked
		return
	}
	err := msg.Respond(reponse)
	if err != nil {
		fmt.Println(err)
	}
}
