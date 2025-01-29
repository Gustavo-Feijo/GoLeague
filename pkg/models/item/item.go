package item

import "goleague/pkg/models/image"

// Struct for holding the item gold.
type Gold struct {
	Base        uint16 `json:"base"`
	Total       uint16 `json:"total"`
	Sell        uint16 `json:"sell"`
	Purchasable bool   `json:"purchasable"`
}

// Struct for holding a champion data.
type Item struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"Description"`
	Plaintext   string      `json:"plaintext"`
	Image       image.Image `json:"image"`
	Gold        Gold        `json:"gold"`
}
