package champion

import "goleague/pkg/models/image"

// Struct for holding a champion data.
type Champion struct {
	ID      string      `json:"id"`
	NameKey string      `json:"key"`
	Name    string      `json:"name"`
	Title   string      `json:"title"`
	Image   image.Image `json:"image"`
	Spells  []Spell     `json:"spells"`
	Passive Spell       `json:"passive"`
}
