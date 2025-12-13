package repositories

import (
	"goleague/pkg/database/models"
	"testing"
	"time"
)

var fixedDate = time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

func normalizePlayerTimes(players []*models.PlayerInfo) {
	for _, p := range players {
		p.LastMatchFetch = p.LastMatchFetch.UTC()
		p.UpdatedAt = p.UpdatedAt.UTC()
	}
}

func getPlayerSearchExpectedResult(t *testing.T, testName string) []*models.PlayerInfo {
	t.Helper()

	switch testName {
	case "allfilters":
		return getPlayerSearchAllFilters()
	}

	return nil
}

func getPlayerSearchAllFilters() []*models.PlayerInfo {
	return []*models.PlayerInfo{
		{
			ID:             12,
			ProfileIcon:    123,
			Puuid:          "kr-puuid-fafafa",
			RiotIdGameName: "Fafafa",
			RiotIdTagline:  "T1",
			SummonerLevel:  500,
			Region:         "kr1",
			UnfetchedMatch: false,
		}, {
			ID:             9,
			ProfileIcon:    4895,
			Puuid:          "kr-puuid-faker",
			RiotIdGameName: "Faker",
			RiotIdTagline:  "T1",
			SummonerLevel:  523,
			Region:         "kr1",
			UnfetchedMatch: false,
		},
	}
}

func getPlayerByIdExpectedResult(t *testing.T, testName string) *models.PlayerInfo {
	t.Helper()

	switch testName {
	case "existentplayer":
		return getExistentPlayer()
	}

	return nil
}

func getExistentPlayer() *models.PlayerInfo {
	return &models.PlayerInfo{
		ID:             12,
		ProfileIcon:    123,
		Puuid:          "kr-puuid-fafafa",
		RiotIdGameName: "Fafafa",
		RiotIdTagline:  "T1",
		SummonerLevel:  500,
		Region:         "kr1",
		UnfetchedMatch: false,
		LastMatchFetch: fixedDate,
		UpdatedAt:      fixedDate,
	}
}
