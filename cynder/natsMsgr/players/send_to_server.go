package players

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CytonicMC/Cynder/cynder/messaging"
	"github.com/CytonicMC/Cynder/cynder/natsMsgr/servers"
	"github.com/CytonicMC/Cynder/cynder/util"
	"github.com/nats-io/nats.go"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type SendPlayerToServerContainer struct {
	Player   uuid.UUID  `json:"player"`
	ServerID string     `json:"serverId"`
	Instance *uuid.UUID `json:"instance"`
}

type SendPlayerToGenericServerContainer struct {
	Player uuid.UUID `json:"player"`
	Group  string    `json:"group"`
	Type   string    `json:"type"`
}

func HandlePlayerSend(services *util.Services, proxy *proxy.Proxy, ctx context.Context) {
	err := services.Nats.Subscribe("players.send", func(msg *nats.Msg) {
		data := string(msg.Data)
		var container SendPlayerToServerContainer
		err := json.Unmarshal([]byte(data), &container)

		if err != nil {
			fmt.Printf("failed to unmarshal SendPlayerToServerContainer: %s", err)
			respond(msg, messaging.ServerSendResponse{
				Success: false,
				Message: "ERR_REQUEST_PROCESSING_FAILURE",
			})
			return
		}

		player := proxy.Player(container.Player)

		if player == nil {
			// player not on this proxy
			return
		}

		server := proxy.Server(container.ServerID)

		if container.Instance != nil {
			services.Redis.SetExpiringValueGlobal(
				fmt.Sprintf("%s#target_instance", container.Player.String()),
				container.Instance.String(),
				3,
			)
		}

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

func HandleGenericSend(services *util.Services, proxy *proxy.Proxy, ctx context.Context) {
	err := services.Nats.Subscribe("players.send.generic", func(msg *nats.Msg) {
		data := string(msg.Data)
		var container SendPlayerToGenericServerContainer
		err := json.Unmarshal([]byte(data), &container)

		fmt.Printf("Generic?? %s\n", data)

		if err != nil {
			fmt.Println(err)
			return
		}

		player := proxy.Player(container.Player)

		if player == nil {
			// player not on this proxy
			return
		}

		server := servers.GetLeastLoadedServer(container.Group, container.Type)
		if server == nil {
			respond(msg, messaging.ServerSendResponse{
				Success: false,
				Message: "ERR_SERVER_NOT_FOUND",
			})
			return
		}

		newCtx, cancel := context.WithTimeout(ctx, 7*time.Second)
		defer cancel()

		connect, err1 := player.CreateConnectionRequest(server).Connect(newCtx)

		if err1 != nil {
			fmt.Printf("Error: %v\n", err)

			respond(msg, messaging.ServerSendResponse{
				Success: false,
				Message: "ERR_CONNECT_REQUEST_FAILED",
			})
			return
		}
		if connect.Status().ServerDisconnected() {
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
		} else {
			fmt.Printf("Generic?? %+v\n", connect)
			respond(msg, messaging.ServerSendResponse{
				Success: false,
				Message: "ERR_UNKNOWN",
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
	fmt.Println(string(reponse))
	err := msg.Respond(reponse)
	if err != nil {
		fmt.Println(err)
	}
}
