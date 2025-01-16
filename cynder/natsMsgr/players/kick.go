package players

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/CytonicMC/Cynder/cynder/messaging"
	"github.com/CytonicMC/Cynder/cynder/natsMsgr/servers"
	"github.com/nats-io/nats.go"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/componentutil"
	"go.minekube.com/gate/pkg/util/uuid"
)

func HandlePlayerKick(nc messaging.NatsService, proxy *proxy.Proxy, ctx context.Context) {
	const subject = "players.kick"
	err := nc.Subscribe(subject, func(msg *nats.Msg) {
		data := string(msg.Data)
		packet, err1 := Deserialize(data)
		if err1 != nil {
			fmt.Println(err1)
			return
		}
		parsedUuid, _ := uuid.Parse(packet.UUID)
		player := proxy.Player(parsedUuid)

		if player == nil {
			// player not on this proxy
			return
		}

		parsedComponent, err := componentutil.ParseTextComponent(player.Protocol(), packet.KickMessage)

		if err != nil {
			fmt.Println(err)
		}

		if packet.Reason.Rescuable {
			// rescue the player
			//todo: chage this to automatically go down a layer
			newServer := servers.GetLeastLoadedServer("lobby", player.CurrentServer().Server().ServerInfo().Name())
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
