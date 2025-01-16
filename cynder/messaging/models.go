package messaging

import "go.minekube.com/gate/pkg/util/uuid"

type PlayerChangeServerContainer struct {
	Player    uuid.UUID `json:"player"`
	OldServer string    `json:"oldServer"`
	NewServer string    `json:"newServer"`
}
