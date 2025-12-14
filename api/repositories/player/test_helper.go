package repositories

import (
	"goleague/pkg/database/models"
	"time"
)

var fixedDate = time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

func normalizePlayerTimes(players []*models.PlayerInfo) {
	for _, p := range players {
		p.LastMatchFetch = p.LastMatchFetch.UTC()
		p.UpdatedAt = p.UpdatedAt.UTC()
	}
}

func buildPartialPlayer(full *models.PlayerInfo) *models.PlayerInfo {
	return &models.PlayerInfo{
		ID:             full.ID,
		ProfileIcon:    full.ProfileIcon,
		Puuid:          full.Puuid,
		RiotIdGameName: full.RiotIdGameName,
		RiotIdTagline:  full.RiotIdTagline,
		SummonerLevel:  full.SummonerLevel,
		Region:         full.Region,
	}
}
