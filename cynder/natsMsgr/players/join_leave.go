package players

import (
	"encoding/json"
	"github.com/CytonicMC/Cynder/cynder/messaging"
	"github.com/go-logr/logr"
	"go.minekube.com/gate/pkg/util/uuid"
)

type PlayerStatus struct {
	UUID     uuid.UUID `json:"uuid"`
	Username string    `json:"username"`
}

func BroadcastPlayerLeave(nc messaging.NatsService, username string, id uuid.UUID, logger logr.Logger) {
	const subject = "players.disconnect"
	msg, err := json.Marshal(PlayerStatus{
		Username: username,
		UUID:     id,
	})
	if err != nil {
		logger.Error(err, "Failed to marshal player status!")
		return
	}

	err1 := nc.Publish(subject, msg)
	if err1 != nil {
		logger.Error(err1, "Failed to send player disconnect!")
		return
	}
}

func BroadcastPlayerJoin(nc messaging.NatsService, username string, id uuid.UUID, logger logr.Logger) {
	const subject = "players.connect"
	msg, err := json.Marshal(PlayerStatus{
		Username: username,
		UUID:     id,
	})
	if err != nil {
		logger.Error(err, "Failed to marshal player status!")
		return
	}

	err1 := nc.Publish(subject, msg)
	if err1 != nil {
		logger.Error(err1, "Failed to send player connect!")
		return
	}
}
