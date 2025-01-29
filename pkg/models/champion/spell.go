package champion

import "goleague/pkg/models/image"

// Struct for holding a champion spell.
type Spell struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cooldown    string `json:"cooldown"`
	Cost        string `json:"cost"`

	ChampionID string      `json:"championId"`
	Image      image.Image `json:"image"`
}
