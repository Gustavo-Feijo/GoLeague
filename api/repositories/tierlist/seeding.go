package repositories

import (
	"goleague/pkg/database/models"
	tiervalues "goleague/pkg/riotvalues/tier"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func seedTestData(t *testing.T, db *gorm.DB) {
	// Clean up existing data
	db.Exec("TRUNCATE TABLE match_bans")
	db.Exec("TRUNCATE TABLE match_stats")
	db.Exec("TRUNCATE TABLE match_infos")

	// Seed match_infos
	// Queue 420 (Ranked Solo/Duo) with various tiers
	matchInfos := []*models.MatchInfo{
		{ID: 1, QueueId: 420, MatchId: "BR1", AverageRating: float64(tiervalues.CalculateRank("DIAMOND", "I", 0))}, // Diamond
		{ID: 2, QueueId: 420, MatchId: "BR2", AverageRating: float64(tiervalues.CalculateRank("DIAMOND", "I", 0))},
		{ID: 3, QueueId: 420, MatchId: "BR3", AverageRating: float64(tiervalues.CalculateRank("DIAMOND", "I", 0))},
		{ID: 4, QueueId: 420, MatchId: "BR4", AverageRating: float64(tiervalues.CalculateRank("DIAMOND", "I", 0))},
		{ID: 5, QueueId: 420, MatchId: "BR5", AverageRating: float64(tiervalues.CalculateRank("DIAMOND", "I", 0))},
		{ID: 6, QueueId: 420, MatchId: "BR6", AverageRating: float64(tiervalues.CalculateRank("EMERALD", "4", 0))}, // Gold
		{ID: 7, QueueId: 420, MatchId: "BR7", AverageRating: float64(tiervalues.CalculateRank("GOLD", "3", 0))},
		{ID: 8, QueueId: 420, MatchId: "BR8", AverageRating: float64(tiervalues.CalculateRank("GOLD", "I", 0))},
		{ID: 9, QueueId: 450, MatchId: "BR9", AverageRating: float64(tiervalues.CalculateRank("GOLD", "I", 0))}, // ARAM
		{ID: 10, QueueId: 450, MatchId: "BR10", AverageRating: float64(tiervalues.CalculateRank("DIAMOND", "I", 0))},
	}

	for _, mi := range matchInfos {
		err := db.Create(mi).Error
		require.NoError(t, err)
	}

	playerInfos := []*models.PlayerInfo{
		{ID: 1, Puuid: "P1"},
		{ID: 2, Puuid: "P2"},
		{ID: 3, Puuid: "P3"},
		{ID: 4, Puuid: "P4"},
		{ID: 5, Puuid: "P5"},
		{ID: 6, Puuid: "P6"},
		{ID: 7, Puuid: "P7"},
		{ID: 8, Puuid: "P8"},
		{ID: 9, Puuid: "P9"},
		{ID: 10, Puuid: "P10"},
	}

	for _, pi := range playerInfos {
		err := db.Create(pi).Error
		require.NoError(t, err)
	}

	// Seed match_stats
	// Champion 1 (ADC) - High win rate, high pick rate in Diamond
	// Champion 2 (Support) - Medium win rate
	// Champion 3 (MId) - Low pick rate (should be filtered out)
	// Champion 4 (Top) - Good stats
	// Champion 5 (Jungle) - In Gold only
	matchStats := []*models.MatchStats{
		// Match 1 - Diamond
		{ID: 1, MatchId: 1, PlayerId: 1, PlayerData: models.MatchPlayer{ChampionId: 1, TeamPosition: "BOTTOM", Win: true}},
		{ID: 2, MatchId: 1, PlayerId: 2, PlayerData: models.MatchPlayer{ChampionId: 2, TeamPosition: "UTILITY", Win: true}},
		{ID: 3, MatchId: 1, PlayerId: 4, PlayerData: models.MatchPlayer{ChampionId: 4, TeamPosition: "TOP", Win: true}},
		{ID: 4, MatchId: 1, PlayerId: 3, PlayerData: models.MatchPlayer{ChampionId: 6, TeamPosition: "MIDDLE", Win: false}},
		{ID: 5, MatchId: 1, PlayerId: 5, PlayerData: models.MatchPlayer{ChampionId: 7, TeamPosition: "JUNGLE", Win: false}},

		// Match 2 - Diamond
		{ID: 6, MatchId: 2, PlayerId: 6, PlayerData: models.MatchPlayer{ChampionId: 1, TeamPosition: "BOTTOM", Win: true}},
		{ID: 7, MatchId: 2, PlayerId: 1, PlayerData: models.MatchPlayer{ChampionId: 2, TeamPosition: "UTILITY", Win: false}},
		{ID: 8, MatchId: 2, PlayerId: 9, PlayerData: models.MatchPlayer{ChampionId: 4, TeamPosition: "TOP", Win: true}},
		{ID: 9, MatchId: 2, PlayerId: 7, PlayerData: models.MatchPlayer{ChampionId: 6, TeamPosition: "MIDDLE", Win: true}},
		{ID: 10, MatchId: 2, PlayerId: 8, PlayerData: models.MatchPlayer{ChampionId: 7, TeamPosition: "JUNGLE", Win: false}},

		// Match 3 - Diamond
		{ID: 11, MatchId: 3, PlayerId: 1, PlayerData: models.MatchPlayer{ChampionId: 1, TeamPosition: "BOTTOM", Win: true}},
		{ID: 12, MatchId: 3, PlayerId: 2, PlayerData: models.MatchPlayer{ChampionId: 2, TeamPosition: "UTILITY", Win: true}},
		{ID: 13, MatchId: 3, PlayerId: 3, PlayerData: models.MatchPlayer{ChampionId: 4, TeamPosition: "TOP", Win: false}},
		{ID: 14, MatchId: 3, PlayerId: 8, PlayerData: models.MatchPlayer{ChampionId: 6, TeamPosition: "MIDDLE", Win: true}},
		{ID: 15, MatchId: 3, PlayerId: 5, PlayerData: models.MatchPlayer{ChampionId: 7, TeamPosition: "", Win: true}},

		// Match 4 - Diamond
		{ID: 16, MatchId: 4, PlayerId: 9, PlayerData: models.MatchPlayer{ChampionId: 1, TeamPosition: "BOTTOM", Win: true}},
		{ID: 17, MatchId: 4, PlayerId: 6, PlayerData: models.MatchPlayer{ChampionId: 2, TeamPosition: "UTILITY", Win: false}},
		{ID: 18, MatchId: 4, PlayerId: 4, PlayerData: models.MatchPlayer{ChampionId: 4, TeamPosition: "TOP", Win: true}},
		{ID: 19, MatchId: 4, PlayerId: 5, PlayerData: models.MatchPlayer{ChampionId: 6, TeamPosition: "MIDDLE", Win: false}},
		{ID: 20, MatchId: 4, PlayerId: 2, PlayerData: models.MatchPlayer{ChampionId: 8, TeamPosition: "JUNGLE", Win: false}},

		// Match 5 - Diamond (Champion 3 appears once - low pick rate)
		{ID: 21, MatchId: 5, PlayerId: 3, PlayerData: models.MatchPlayer{ChampionId: 3, TeamPosition: "MIDDLE", Win: true}},
		{ID: 22, MatchId: 5, PlayerId: 6, PlayerData: models.MatchPlayer{ChampionId: 2, TeamPosition: "UTILITY", Win: true}},
		{ID: 23, MatchId: 5, PlayerId: 5, PlayerData: models.MatchPlayer{ChampionId: 4, TeamPosition: "TOP", Win: true}},
		{ID: 24, MatchId: 5, PlayerId: 4, PlayerData: models.MatchPlayer{ChampionId: 1, TeamPosition: "BOTTOM", Win: false}},
		{ID: 25, MatchId: 5, PlayerId: 2, PlayerData: models.MatchPlayer{ChampionId: 7, TeamPosition: "JUNGLE", Win: false}},

		// Match 6 - Gold
		{ID: 26, MatchId: 6, PlayerId: 1, PlayerData: models.MatchPlayer{ChampionId: 5, TeamPosition: "JUNGLE", Win: true}},
		{ID: 27, MatchId: 6, PlayerId: 2, PlayerData: models.MatchPlayer{ChampionId: 1, TeamPosition: "BOTTOM", Win: true}},
		{ID: 28, MatchId: 6, PlayerId: 7, PlayerData: models.MatchPlayer{ChampionId: 2, TeamPosition: "UTILITY", Win: false}},
		{ID: 29, MatchId: 6, PlayerId: 5, PlayerData: models.MatchPlayer{ChampionId: 4, TeamPosition: "TOP", Win: false}},
		{ID: 30, MatchId: 6, PlayerId: 4, PlayerData: models.MatchPlayer{ChampionId: 6, TeamPosition: "MIDDLE", Win: true}},

		// Match 7 - Gold
		{ID: 31, MatchId: 7, PlayerId: 6, PlayerData: models.MatchPlayer{ChampionId: 5, TeamPosition: "JUNGLE", Win: true}},
		{ID: 32, MatchId: 7, PlayerId: 2, PlayerData: models.MatchPlayer{ChampionId: 1, TeamPosition: "BOTTOM", Win: false}},
		{ID: 33, MatchId: 7, PlayerId: 5, PlayerData: models.MatchPlayer{ChampionId: 2, TeamPosition: "UTILITY", Win: true}},
		{ID: 34, MatchId: 7, PlayerId: 4, PlayerData: models.MatchPlayer{ChampionId: 4, TeamPosition: "TOP", Win: true}},
		{ID: 35, MatchId: 7, PlayerId: 1, PlayerData: models.MatchPlayer{ChampionId: 6, TeamPosition: "MIDDLE", Win: false}},

		// Match 8 - Gold
		{ID: 36, MatchId: 8, PlayerId: 8, PlayerData: models.MatchPlayer{ChampionId: 5, TeamPosition: "JUNGLE", Win: false}},
		{ID: 37, MatchId: 8, PlayerId: 7, PlayerData: models.MatchPlayer{ChampionId: 1, TeamPosition: "BOTTOM", Win: true}},
		{ID: 38, MatchId: 8, PlayerId: 6, PlayerData: models.MatchPlayer{ChampionId: 2, TeamPosition: "UTILITY", Win: true}},
		{ID: 39, MatchId: 8, PlayerId: 9, PlayerData: models.MatchPlayer{ChampionId: 4, TeamPosition: "TOP", Win: false}},
		{ID: 40, MatchId: 8, PlayerId: 5, PlayerData: models.MatchPlayer{ChampionId: 6, TeamPosition: "MIDDLE", Win: true}},

		// Match 9 - ARAM (no positions)
		{ID: 41, MatchId: 9, PlayerId: 4, PlayerData: models.MatchPlayer{ChampionId: 1, TeamPosition: "", Win: true}},
		{ID: 42, MatchId: 9, PlayerId: 2, PlayerData: models.MatchPlayer{ChampionId: 2, TeamPosition: "", Win: true}},
		{ID: 43, MatchId: 9, PlayerId: 5, PlayerData: models.MatchPlayer{ChampionId: 10, TeamPosition: "", Win: false}},
		{ID: 44, MatchId: 9, PlayerId: 8, PlayerData: models.MatchPlayer{ChampionId: 11, TeamPosition: "", Win: false}},
		{ID: 45, MatchId: 9, PlayerId: 1, PlayerData: models.MatchPlayer{ChampionId: 12, TeamPosition: "", Win: false}},

		// Match 10 - ARAM
		{ID: 46, MatchId: 10, PlayerId: 1, PlayerData: models.MatchPlayer{ChampionId: 1, TeamPosition: "", Win: false}},
		{ID: 47, MatchId: 10, PlayerId: 2, PlayerData: models.MatchPlayer{ChampionId: 2, TeamPosition: "", Win: false}},
		{ID: 48, MatchId: 10, PlayerId: 4, PlayerData: models.MatchPlayer{ChampionId: 10, TeamPosition: "", Win: true}},
		{ID: 49, MatchId: 10, PlayerId: 3, PlayerData: models.MatchPlayer{ChampionId: 11, TeamPosition: "", Win: true}},
		{ID: 50, MatchId: 10, PlayerId: 9, PlayerData: models.MatchPlayer{ChampionId: 12, TeamPosition: "", Win: true}},
	}

	for _, ms := range matchStats {
		err := db.Omit(clause.Associations).Create(ms).Error
		require.NoError(t, err)
	}

	// Seed match_bans
	matchBans := []*models.MatchBans{
		// Diamond matches - Champion 1 is frequently banned
		{MatchId: 1, ChampionId: 1},
		{MatchId: 2, ChampionId: 1},
		{MatchId: 3, ChampionId: 1},
		{MatchId: 4, ChampionId: 9}, // Different ban
		{MatchId: 5, ChampionId: 9},

		// Gold matches - Champion 5 banned
		{MatchId: 6, ChampionId: 5},
		{MatchId: 7, ChampionId: 5},
		{MatchId: 8, ChampionId: 1},

		// ARAM - some bans
		{MatchId: 9, ChampionId: 10},
		{MatchId: 10, ChampionId: 11},
	}

	for _, mb := range matchBans {
		err := db.Create(mb).Error
		require.NoError(t, err)
	}
}
