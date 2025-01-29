package assets

import (
	"goleague/pkg/models/champion"
	"goleague/pkg/models/image"
	"goleague/pkg/models/item"
)

// Convert the default map from the DDragon to a image type.
func mapToImage(imgData map[string]interface{}) image.Image {
	return image.Image{
		Full:   imgData["full"].(string),
		Sprite: imgData["sprite"].(string),
		X:      uint16(imgData["x"].(float64)),
		Y:      uint16(imgData["y"].(float64)),
		W:      uint16(imgData["w"].(float64)),
		H:      uint16(imgData["h"].(float64)),
	}
}

// Convert the default map from the DDragon to a spell type.
func mapToSpell(spellData map[string]interface{}, championID string) champion.Spell {
	spell := champion.Spell{
		Name:        spellData["name"].(string),
		Description: spellData["description"].(string),
		Cooldown:    getStringOrDefault(spellData, "cooldown"),
		Cost:        getStringOrDefault(spellData, "cost"),
		ChampionID:  championID,
	}

	// Get the spell image.
	if imgData, ok := spellData["image"].(map[string]interface{}); ok {
		spell.Image = mapToImage(imgData)
	}

	return spell
}

// Convert the gold of the item to a gold type.
func mapToGold(goldData map[string]interface{}) item.Gold {
	gold := item.Gold{
		Base:        uint16(goldData["base"].(float64)),
		Total:       uint16(goldData["total"].(float64)),
		Sell:        uint16(goldData["sell"].(float64)),
		Purchasable: goldData["purchasable"].(bool),
	}

	return gold
}

// Return the string if it's available, else returns a empty string.
func getStringOrDefault(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}
