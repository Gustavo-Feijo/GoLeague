package repositories

import (
	"goleague/pkg/database/models"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func seedPlayerTestData(t *testing.T, db *gorm.DB) map[string]*models.PlayerInfo {
	// Clean up existing data
	db.Exec("TRUNCATE TABLE player_infos")

	seededData := map[string]*models.PlayerInfo{
		"brtt":        {ID: 1, Puuid: "br-puuid-brtt", ProfileIcon: 29, RiotIdGameName: "Brtt", RiotIdTagline: "BR1", SummonerLevel: 450, Region: "BR1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		"brtt2":       {ID: 2, Puuid: "br-puuid-brtt2", ProfileIcon: 4901, RiotIdGameName: "BrTT", RiotIdTagline: "GG", SummonerLevel: 380, Region: "BR1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		"minerva":     {ID: 3, Puuid: "br-puuid-minerva", ProfileIcon: 5012, RiotIdGameName: "Minerva", RiotIdTagline: "BR1", SummonerLevel: 320, Region: "BR1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		"doublelift":  {ID: 4, Puuid: "na-puuid-doublelift", ProfileIcon: 588, RiotIdGameName: "Doublelift", RiotIdTagline: "NA1", SummonerLevel: 500, Region: "NA1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		"doublelift2": {ID: 5, Puuid: "na-puuid-doublelift2", ProfileIcon: 4568, RiotIdGameName: "DoubleLift", RiotIdTagline: "TSM", SummonerLevel: 412, Region: "NA1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		"sneaky":      {ID: 6, Puuid: "na-puuid-sneaky", ProfileIcon: 4221, RiotIdGameName: "Sneaky", RiotIdTagline: "C9", SummonerLevel: 398, Region: "NA1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		"faker_euw":   {ID: 7, Puuid: "euw-puuid-faker", ProfileIcon: 3150, RiotIdGameName: "Faker", RiotIdTagline: "EUW", SummonerLevel: 425, Region: "EUW1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		"rekkles":     {ID: 8, Puuid: "euw-puuid-rekkles", ProfileIcon: 4320, RiotIdGameName: "Rekkles", RiotIdTagline: "FNC", SummonerLevel: 441, Region: "EUW1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		"faker":       {ID: 9, Puuid: "kr-puuid-faker", ProfileIcon: 4895, RiotIdGameName: "Faker", RiotIdTagline: "T1", SummonerLevel: 523, Region: "KR1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		"showmaker":   {ID: 10, Puuid: "kr-puuid-showmaker", ProfileIcon: 4770, RiotIdGameName: "ShowMaker", RiotIdTagline: "DK", SummonerLevel: 489, Region: "KR1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		"chovy":       {ID: 11, Puuid: "kr-puuid-chovy", ProfileIcon: 4655, RiotIdGameName: "Chovy", RiotIdTagline: "GEN", SummonerLevel: 467, Region: "KR1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		"fafafa":      {ID: 12, Puuid: "kr-puuid-fafafa", ProfileIcon: 123, RiotIdGameName: "Fafafa", RiotIdTagline: "T1", SummonerLevel: 500, Region: "KR1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
	}

	for _, pi := range seededData {
		err := db.Create(pi).Error
		require.NoError(t, err)
	}

	return seededData
}
