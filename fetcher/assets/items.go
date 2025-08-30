package assets

import (
	"encoding/json"
	"fmt"
	"goleague/fetcher/repositories"
	"goleague/fetcher/requests"
	"goleague/pkg/models/item"
	"goleague/pkg/redis"
	"log"

	"gorm.io/gorm"
)

// Revalidate the full item cache.
func RevalidateItemCache(redis *redis.RedisClient, db *gorm.DB, language string) error {
	repo, _ := repositories.NewCacheRepository(db)

	// Get the latest version.
	// Usually only GetLatestVersion should be used to get the current running latest.
	// But we are using GetNewVersion to also revalidate the versions.
	var latestVersion *string
	versions, err := GetNewVersion(redis)
	if err != nil {
		latestVersion = GetLatestVersion(redis)
	} else {
		latestVersion = &versions[0]
	}

	if latestVersion == nil {
		log.Fatalf("couldn't get the latest version")
	}

	// Format the champion api url.
	url := fmt.Sprintf("%scdn/%s/data/%s/item.json", ddragon, *latestVersion, language)
	resp, err := requests.Request(url, "GET")
	if err != nil {
		return fmt.Errorf("couldn't get the current version: %v", err)
	}

	defer resp.Body.Close()

	// Read the champion json.
	var itemData fullItem
	if err := json.NewDecoder(resp.Body).Decode(&itemData); err != nil {
		return fmt.Errorf("couldn't convert the body to json: %v", err)
	}

	// Loop through each item.
	for itemKey, itemData := range itemData.Data {
		if err := handleItemEntry(redis, repo, itemKey, itemData); err != nil {
			return err
		}
	}

	return nil
}

// handleItemEntry handles a single item cache entry.
func handleItemEntry(redis *redis.RedisClient, repo repositories.CacheRepository, itemKey string, itemData map[string]any) error {
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

	// Convert the item to json.
	itemJson, err := json.Marshal(newItem)
	if err != nil {
		return fmt.Errorf("can't convert the item back to json: %v", err)
	}

	keyWithId := fmt.Sprint(itemPrefix, itemKey)

	// Set the key on the database. Fallback.
	if repo != nil {
		repo.Setkey(keyWithId, string(itemJson))
	}

	// Get the client and set the champion.
	if err := redis.Set(ctx, keyWithId, itemJson, 0); err != nil {
		return fmt.Errorf("can't set the item json on redis: %v", err)
	}

	return nil
}
