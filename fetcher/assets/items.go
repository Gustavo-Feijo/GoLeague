package assets

import (
	"encoding/json"
	"fmt"
	"goleague/fetcher/requests"
	"goleague/pkg/models/item"
	"goleague/pkg/redis"
	"log"
)

// Revalidate the full item cache.
// Returns a specific item if a id was passed.
func RevalidateItemCache(language string, itemId string) (*item.Item, error) {
	// Get the latest version.
	// Usually only GetLatestVersion should be used to get the current running latest.
	// But we are using GetNewVersion to also revalidate the versions.
	latestVersion := ""
	versions, err := GetNewVersion()
	if err != nil {
		latestVersion, err = GetLatestVersion()
		if err != nil {
			log.Fatalf("Can't get the latest version: %v", err)
		}
	} else {
		latestVersion = versions[0]
	}

	// Format the champion api url.
	url := fmt.Sprintf("%scdn/%s/data/%s/item.json", ddragon, latestVersion, language)
	resp, err := requests.Request(url, "GET")
	if err != nil {
		return nil, fmt.Errorf("couldn't get the current version: %v", err)
	}

	defer resp.Body.Close()

	// Read the champion json.
	var itemData fullItem
	if err := json.NewDecoder(resp.Body).Decode(&itemData); err != nil {
		return nil, fmt.Errorf("couldn't convert the body to json: %v", err)
	}

	// Initialize the item to be returned if found.
	var returnItem *item.Item

	// Loop through each item.
	for itemKey, itemData := range itemData.Data {
		// Create the new item.
		newItem := &item.Item{
			ID:          itemKey,
			Name:        getStringOrDefault(itemData, "name"),
			Description: getStringOrDefault(itemData, "description"),
			Plaintext:   getStringOrDefault(itemData, "plaintext"),
		}

		// Get the image data.
		if imgData, ok := itemData["image"].(map[string]any); ok {
			newItem.Image = mapToImage(imgData)
		}

		// Get the gold data.
		if goldData, ok := itemData["gold"].(map[string]any); ok {
			newItem.Gold = mapToGold(goldData)
		}

		// Verify if it's the searched item.
		if itemKey == itemId {
			returnItem = newItem
		}

		// Convert the item to json.
		itemJson, err := json.Marshal(newItem)
		if err != nil {
			return nil, fmt.Errorf("can't convert the item back to json: %v", err)
		}

		keyWithId := fmt.Sprint(itemPrefix, itemKey)

		// Get the client and set the champion.
		client := redis.GetClient()
		if err := client.Set(ctx, keyWithId, itemJson, 0); err != nil {
			return nil, fmt.Errorf("can't set the item json on redis: %v", err)
		}
	}

	// Return the champion and no error.
	// Will be returned nil if not found.
	return returnItem, nil
}
