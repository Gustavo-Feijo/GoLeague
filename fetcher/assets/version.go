package assets

import (
	"encoding/json"
	"errors"
	"fmt"
	"goleague/fetcher/requests"
	"goleague/pkg/redis"
)

// Get the latest version of the data from the ddragon.
func GetLatestVersion() (string, error) {
	// Try to find the latest version in the redis cache.
	client := redis.GetClient()
	result, err := client.LIndex(ctx, versionKey, 0).Result()
	if err == nil {
		return result, nil
	}

	// The version was not found, fetch from ddragon.
	newVersions, err := GetNewVersion()
	if err != nil {
		// In that case, can't proceed with the fetching.
		panic("Can't get the latest version.")
	}
	return newVersions[0], nil
}

// Get all the versions from the ddragon.
// Set the latest three on the Redis cache and return.
func GetNewVersion() ([]string, error) {
	// Format the versions api url.
	url := fmt.Sprint(ddragon, "api/versions.json")
	resp, err := requests.Request(url, "GET")
	if err != nil {
		return nil, fmt.Errorf("couldn't get the current version: %v", err)
	}

	defer resp.Body.Close()

	// Read the version json/array into the version.
	var version []string
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return nil, fmt.Errorf("couldn't convert the body to json: %v", err)
	}

	if len(version) == 0 {
		return nil, errors.New("no versions available")
	}

	client := redis.GetClient()

	// Delete the version key.
	err = client.Del(ctx, versionKey).Err()
	if err != nil {
		return nil, fmt.Errorf("couldn't delete the Redis key: %v", err)
	}

	// Push the version to the redis.
	client.RPush(ctx, versionKey, []interface{}{version[0], version[1], version[2]}).Result()
	return version[:3], nil
}

// Get the versions from the cache and return.
func GetVersion() ([]string, error) {
	client := redis.GetClient()
	result, err := client.LRange(ctx, versionKey, 0, 3).Result()
	if err != nil {
		return nil, err
	}
	return result[:3], nil
}
