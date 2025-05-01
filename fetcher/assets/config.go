package assets

import "context"

// Consts used across the package.
const (
	championPrefix = "ddragon:champion:"
	ddragon        = "https://ddragon.leagueoflegends.com/"
	versionKey     = "ddragon:versions"
	itemPrefix     = "ddragon:item:"
	workerCount    = 10
)

// Default context.
var ctx = context.Background()

// Definition for extracting the champion data.
type fullChampion struct {
	Data map[string]any `json:"data"`
}

// Definition for extracting the item data.
type fullItem struct {
	Data map[string]map[string]any `json:"data"`
}
