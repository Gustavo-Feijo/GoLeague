package assets

import (
	"encoding/json"
	"errors"
	"fmt"
	"goleague/fetcher/requests"
	"goleague/pkg/redis"
)

// Get the latest version of the data from the ddragon.
func GetLatestVersion(redis *redis.RedisClient) *string {
	// Try to find the latest version in the redis cache.
	result, err := redis.LIndex(ctx, versionKey, 0).Result()
	if err == nil {
		return &result
	}

	// The version was not found, fetch from ddragon.
	newVersions, err := GetNewVersion(redis)
	if err != nil {
		return nil
	}

	return &newVersions[0]
}

// Get all the versions from the ddragon.
// Set the latest three on the Redis cache and return.
func GetNewVersion(redis *redis.RedisClient) ([]string, error) {
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

	// Delete the version key.
	err = redis.Del(ctx, versionKey).Err()
	if err != nil {
		return nil, fmt.Errorf("couldn't delete the Redis key: %v", err)
	}

	// Push the version to the redis.
	redis.RPush(ctx, versionKey, []any{version[0], version[1], version[2]}).Result()
	return version[:3], nil
}

// Get the versions from the cache and return.
func GetVersion(redis *redis.RedisClient) ([]string, error) {
	result, err := redis.LRange(ctx, versionKey, 0, 3).Result()
	if err != nil {
		return nil, err
	}
	return result[:3], nil
}
