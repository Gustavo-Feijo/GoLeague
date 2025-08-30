package assets

import (
	"encoding/json"
	"fmt"
	"goleague/fetcher/repositories"
	"goleague/fetcher/requests"
	"goleague/pkg/models/champion"
	"goleague/pkg/redis"
	"log"
	"sync"

	"gorm.io/gorm"
)

// Get the champion from the datadragon based on it's key.
// If a champion key is passed, also return the given champion.
func RevalidateChampionCache(redis *redis.RedisClient, db *gorm.DB, language string) error {
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
	url := fmt.Sprintf("%scdn/%s/data/%s/champion.json", ddragon, *latestVersion, language)
	fmt.Println(url)
	resp, err := requests.Request(url, "GET")
	if err != nil {
		return fmt.Errorf("couldn't get the current version: %v", err)
	}

	defer resp.Body.Close()

	// Read the champion json.
	var championsData fullChampion
	if err := json.NewDecoder(resp.Body).Decode(&championsData); err != nil {
		return fmt.Errorf("couldn't convert the body to json: %v", err)
	}

	var wg sync.WaitGroup

	// Channel for the champion keys. (Champion names on the DDragon).
	championKeys := make(chan string, len(championsData.Data))

	// Start workers.
	for range workerCount {
		go func() {
			for championKey := range championKeys {
				RevalidateSingleChampionByKey(redis, language, championKey, repo)
				wg.Done()
			}
		}()
	}

	// Enqueue tasks.
	for championKey := range championsData.Data {
		wg.Add(1)
		championKeys <- championKey
	}

	// Close the channel and wait for all workers to finish.
	close(championKeys)
	wg.Wait()

	return nil
}

func RevalidateSingleChampionByKey(redis *redis.RedisClient, language string, championKey string, repo repositories.CacheRepository) (*champion.Champion, error) {
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
	url := fmt.Sprintf("%scdn/%s/data/%s/champion/%s.json", ddragon, *latestVersion, language, championKey)
	resp, err := requests.Request(url, "GET")
	if err != nil {
		return nil, fmt.Errorf("couldn't get the champion: %v", err)
	}

	defer resp.Body.Close()

	// Read the champion json into the version.
	var championsData fullChampion
	if err := json.NewDecoder(resp.Body).Decode(&championsData); err != nil {
		return nil, fmt.Errorf("couldn't convert the body to json: %v", err)
	}
	championData, ok := championsData.Data[championKey].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid data format for champion: %s", championKey)
	}

	// Create the champion.
	champ := &champion.Champion{
		ID:      getStringOrDefault(championData, "key"),
		NameKey: getStringOrDefault(championData, "id"),
		Name:    getStringOrDefault(championData, "name"),
		Title:   getStringOrDefault(championData, "title"),
	}

	// Set the champion image.
	if imgData, ok := championData["image"].(map[string]any); ok {
		champ.Image = mapToImage(imgData)
	}

	// Map the spells.
	if spellsData, ok := championData["spells"].([]any); ok {
		champ.Spells = make([]champion.Spell, len(spellsData))
		for i, spellData := range spellsData {
			spellMap := spellData.(map[string]any)
			champ.Spells[i] = mapToSpell(spellMap, champ.ID)
		}
	}

	// Handle passive.
	if passiveData, ok := championData["passive"].(map[string]any); ok {
		champ.Passive = mapToSpell(passiveData, champ.ID)
	}

	// Convert the champion to json.
	champJson, err := json.Marshal(champ)
	if err != nil {
		return nil, fmt.Errorf("can't convert the champion back to json: %v", err)
	}

	// The champion key on the redis cache.
	keyWithId := fmt.Sprint(championPrefix, champ.ID)

	if repo != nil {
		repo.Setkey(keyWithId, string(champJson))
	}

	// Get the client and set the champion.
	if err := redis.Set(ctx, keyWithId, champJson, 0); err != nil {
		return nil, fmt.Errorf("can't set the champion json on redis: %v", err)
	}

	return champ, nil
}
