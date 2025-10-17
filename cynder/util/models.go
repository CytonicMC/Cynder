package util

import (
	"encoding/json"
	"time"
)

type BanData struct {
	Reason   *string    `json:"reason,omitempty"`
	Expiry   *time.Time `json:"expiry,omitempty"`
	IsBanned bool       `json:"isBanned"`
}

// Optional helper for JSON unmarshalling
func (b *BanData) UnmarshalJSON(data []byte) error {
	type Alias BanData
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	return json.Unmarshal(data, &aux)
}
