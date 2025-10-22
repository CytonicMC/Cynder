package util

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/CytonicMC/Cynder/cynder/messaging"
	"github.com/CytonicMC/Cynder/cynder/redis"
	"github.com/CytonicMC/Cynder/cynder/util/mini"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/component/codec/legacy"
	"go.minekube.com/gate/pkg/util/uuid"
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

func CanJoinRestrictedServer(uuid uuid.UUID, redis redis.Service) bool {
	rank := redis.ReadHashGlobal("player_ranks", uuid.String())
	if rank == "OWNER" || rank == "ADMIN" || rank == "MODERATOR" || rank == "HELPER" {
		return true
	}

	// check whitelist
	return redis.SetContainsGlobal("player_whitelist", uuid.String())
}

func IsBanned(uuid uuid.UUID, redis redis.Service) (bool, *c.Text) {
	raw := redis.ReadHashGlobal("banned_players", uuid.String())
	if raw == "" {
		return false, nil
	}
	var data struct {
		Reason   *string    `json:"reason,omitempty"`
		Expiry   *time.Time `json:"expiry,omitempty"`
		IsBanned bool       `json:"isBanned"`
	}
	err := json.Unmarshal([]byte(raw), &data)
	if err != nil {
		return false, nil
	}

	if data.Expiry.Before(time.Now()) {
		redis.RemHashGlobal("banned_players", uuid.String())
		return false, nil
	}

	return data.IsBanned, FormatBanMessage(data)
}

func FormatBanMessage(b BanData) *c.Text {
	if !b.IsBanned {
		return &c.Text{
			Content: "",
		}
	}

	var builder strings.Builder

	if b.Expiry == nil {
		builder.WriteString("<color:red>You are permanently banned from the Cytonic Network!</color><newline><newline>")
	} else {
		builder.WriteString("<color:red>You are currently banned from the Cytonic Network!</color><newline><newline>")
	}

	reason := "No reason provided"
	if b.Reason != nil {
		reason = *b.Reason
	}
	builder.WriteString(fmt.Sprintf("<color:gray>Reason:</color> <color:white>%s</color><newline>", reason))

	if b.Expiry != nil {
		expiry := Unparse(b.Expiry, " ")
		builder.WriteString(fmt.Sprintf("<color:gray>Expires in: </color><color:aqua>%s</color><newline><newline>", expiry))
	} else {
		builder.WriteString("<newline>")
	}

	builder.WriteString("<color:gray>Appeal at:</color> <color:aqua><underlined>https://cytonic.net/appeal</underlined></color><newline>")

	return mini.Parse(builder.String())
}

func Unparse(instant *time.Time, spacing string) string {
	if instant == nil {
		return ""
	}

	now := time.Now()
	duration := instant.Sub(now)
	if duration < 0 {
		duration = -duration // handle inverted durations
	}

	days := int64(duration.Hours() / 24)
	years := days / 365
	days = days % 365
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	var b strings.Builder

	if years > 0 {
		fmt.Fprintf(&b, "%dy%s", years, spacing)
	}
	if days > 0 {
		fmt.Fprintf(&b, "%dd%s", days, spacing)
	}
	if hours > 0 {
		fmt.Fprintf(&b, "%dh%s", hours, spacing)
	}
	if minutes > 0 {
		fmt.Fprintf(&b, "%dm%s", minutes, spacing)
	}
	if seconds > 0 {
		fmt.Fprintf(&b, "%ds%s", seconds, spacing)
	} else {
		fmt.Fprintf(&b, "<1s%s", spacing)
	}

	return b.String()
}

type Services struct {
	Nats  messaging.NatsService
	Redis redis.Service
}
