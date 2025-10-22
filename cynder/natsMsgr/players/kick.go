package players

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CytonicMC/Cynder/cynder/natsMsgr/servers"
	"github.com/CytonicMC/Cynder/cynder/util"
	"github.com/CytonicMC/Cynder/cynder/util/mini"
	"github.com/nats-io/nats.go"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/componentutil"
	"go.minekube.com/gate/pkg/util/uuid"
)

func HandlePlayerKick(services *util.Services, proxyInstance *proxy.Proxy, ctx context.Context) {
	const subject = "players.kick"
	err := services.Nats.Subscribe(subject, func(msg *nats.Msg) {
		data := string(msg.Data)
		packet, err1 := Deserialize(data)
		if err1 != nil {
			fmt.Println(err1)
			return
		}
		parsedUuid, _ := uuid.Parse(packet.UUID)
		player := proxyInstance.Player(parsedUuid)

		if player == nil {
			// player not on this proxyInstance
			return
		}

		parsedComponent, err := componentutil.ParseTextComponent(player.Protocol(), packet.KickMessage)

		if err != nil {
			fmt.Println(err)
		}

		if packet.Reason.Rescuable {
			// rescue the player
			conn := player.CurrentServer()
			if conn == nil {
				//player.SendMessage(parsedComponent)
				//player.Disconnect(parsedComponent)
				connect, err := player.CreateConnectionRequest(servers.GetLeastLoadedServer("cytosis", "lobby")).Connect(ctx)
				if err != nil {
					player.Disconnect(parsedComponent)
				}
				if connect.Status().ConnectionInProgress() {
					player.Disconnect(mini.Parse("<color:red><bold>WHOOPS!</bold></color:red><color:gray> Failed to rescue your connection! <color:red>ERR_ALREADY_CONNECTING"))
				}
				return
			}
			newServer, success := servers.GetFallbackFromServer(player.CurrentServer().Server())
			if !success {
				player.Disconnect(parsedComponent)
				return
			}
			player.CreateConnectionRequest(newServer).ConnectWithIndication(ctx)
		} else {
			player.Disconnect(parsedComponent)
		}
	})
	if err != nil {
		fmt.Printf("error subscribing to subject %s: %s", subject, err)
		return
	}
}

// KickReason represents a kick reason with a rescuable flag.
type KickReason struct {
	Name      string
	Rescuable bool
}

// Predefined kick reasons
var (
	BANNED         = KickReason{"BANNED", false}
	INTERNAL_ERROR = KickReason{"INTERNAL_ERROR", true}
	INVALID_WORLD  = KickReason{"INVALID_WORLD", true}
	COMMAND        = KickReason{"COMMAND", false}
)

// ReasonMap maps string names to KickReason instances.
var ReasonMap = map[string]KickReason{
	"BANNED":         BANNED,
	"INTERNAL_ERROR": INTERNAL_ERROR,
	"INVALID_WORLD":  INVALID_WORLD,
	"COMMAND":        COMMAND,
}

// PlayerKickContainer represents the data sent over the network.
type PlayerKickContainer struct {
	UUID        string     `json:"uuid"`
	Reason      KickReason `json:"reason"`
	KickMessage string     `json:"kickMessage"`
}

// Serialize converts the PlayerKickContainer to JSON.
func (p PlayerKickContainer) Serialize() (string, error) {
	bytes, err := json.Marshal(p)
	return string(bytes), err
}

// Deserialize converts a JSON string to a PlayerKickContainer.
func Deserialize(data string) (*PlayerKickContainer, error) {
	var raw struct {
		UUID        string `json:"uuid"`
		ReasonName  string `json:"reason"`
		KickMessage string `json:"kickMessage"`
	}
	err := json.Unmarshal([]byte(data), &raw)
	if err != nil {
		return nil, err
	}

	// Map reason name back to KickReason
	reason, exists := ReasonMap[raw.ReasonName]
	if !exists {
		return nil, fmt.Errorf("unknown KickReason: %s", raw.ReasonName)
	}

	return &PlayerKickContainer{
		UUID:        raw.UUID,
		Reason:      reason,
		KickMessage: raw.KickMessage,
	}, nil
}
