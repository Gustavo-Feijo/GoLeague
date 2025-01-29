package assets

import (
	"encoding/json"
	"fmt"
	"goleague/fetcher/requests"
	"goleague/pkg/models/champion"
	"goleague/pkg/redis"
	"log"
	"sync"
)

// Get the champion from the datadragon based on it's key.
// If a champion key is passed, also return the given champion.
func RevalidateChampionCache(language string, championId string) (*champion.Champion, error) {
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
	url := fmt.Sprintf("%scdn/%s/data/%s/champion.json", ddragon, latestVersion, language)
	fmt.Println(url)
	resp, err := requests.Request(url, "GET")
	if err != nil {
		return nil, fmt.Errorf("couldn't get the current version: %v", err)
	}

	defer resp.Body.Close()

	// Read the champion json.
	var championsData fullChampion
	if err := json.NewDecoder(resp.Body).Decode(&championsData); err != nil {
		return nil, fmt.Errorf("couldn't convert the body to json: %v", err)
	}

	var wg sync.WaitGroup

	// Channel for the champion keys. (Champion names on the DDragon)
	championKeys := make(chan string, len(championsData.Data))

	// Start workers
	for i := 0; i < workerCount; i++ {
		go func() {
			for championKey := range championKeys {
				RevalidateSingleChampionByKey(language, championKey)
				wg.Done()
			}
		}()
	}

	// Enqueue tasks
	for championKey := range championsData.Data {
		wg.Add(1)
		championKeys <- championKey
	}

	// Close the channel and wait for all workers to finish
	close(championKeys)
	wg.Wait()

	// Handle the case where a champion is being explicitly fetched.
	if championId != "" {

		// Get the redis client to read the champion.
		client := redis.GetClient()
		championReturn, err := client.Get(ctx, championPrefix+championId)
		if err != nil {
			return nil, fmt.Errorf("can't get the new fetched champion from the redis client: %v", err)
		}

		// Get the champion from the json and return it.
		champ := &champion.Champion{}
		if err := json.Unmarshal([]byte(championReturn), &champ); err != nil {
			return nil, fmt.Errorf("can't parse the stored champion json: %v", err)
		}
		return champ, nil
	}

	// No error occurred and return in not requested.
	return nil, nil
}

func RevalidateSingleChampionByKey(language string, championKey string) (*champion.Champion, error) {
	latestVersion, err := GetLatestVersion()
	if err != nil {
		log.Fatalf("Can't get the latest version: %v", err)
	}

	// Format the champion api url.
	url := fmt.Sprintf("%scdn/%s/data/%s/champion/%s.json", ddragon, latestVersion, language, championKey)
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
	championData, ok := championsData.Data[championKey].(map[string]interface{})
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
	if imgData, ok := championData["image"].(map[string]interface{}); ok {
		champ.Image = mapToImage(imgData)
	}

	// Map the spells.
	if spellsData, ok := championData["spells"].([]interface{}); ok {
		champ.Spells = make([]champion.Spell, len(spellsData))
		for i, spellData := range spellsData {
			spellMap := spellData.(map[string]interface{})
			champ.Spells[i] = mapToSpell(spellMap, champ.ID)
		}
	}

	// Handle passive
	if passiveData, ok := championData["passive"].(map[string]interface{}); ok {
		champ.Passive = mapToSpell(passiveData, champ.ID)
	}

	// Convert the champion to json.
	champJson, err := json.Marshal(champ)
	if err != nil {
		return nil, fmt.Errorf("can't convert the champion back to json: %v", err)
	}

	// The champion key on the redis cache.
	keyWithId := fmt.Sprint(championPrefix, champ.ID)

	// Get the client and set the champion.
	client := redis.GetClient()
	if err := client.Set(ctx, keyWithId, champJson, 0); err != nil {
		return nil, fmt.Errorf("can't set the champion json on redis: %v", err)
	}

	return champ, nil
}
