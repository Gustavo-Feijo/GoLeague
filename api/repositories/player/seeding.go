package repositories

import (
	"goleague/pkg/database/models"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func seedPlayerTestData(t *testing.T, db *gorm.DB) {
	// Clean up existing data
	db.Exec("TRUNCATE TABLE player_infos")

	playerInfo := []*models.PlayerInfo{
		{ID: 1, Puuid: "br-puuid-brtt", ProfileIcon: 29, RiotIdGameName: "Brtt", RiotIdTagline: "BR1", SummonerLevel: 450, Region: "br1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		{ID: 2, Puuid: "br-puuid-brtt2", ProfileIcon: 4901, RiotIdGameName: "BrTT", RiotIdTagline: "GG", SummonerLevel: 380, Region: "br1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		{ID: 3, Puuid: "br-puuid-minerva", ProfileIcon: 5012, RiotIdGameName: "Minerva", RiotIdTagline: "BR1", SummonerLevel: 320, Region: "br1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		{ID: 4, Puuid: "na-puuid-doublelift", ProfileIcon: 588, RiotIdGameName: "Doublelift", RiotIdTagline: "NA1", SummonerLevel: 500, Region: "na1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		{ID: 5, Puuid: "na-puuid-doublelift2", ProfileIcon: 4568, RiotIdGameName: "DoubleLift", RiotIdTagline: "TSM", SummonerLevel: 412, Region: "na1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		{ID: 6, Puuid: "na-puuid-sneaky", ProfileIcon: 4221, RiotIdGameName: "Sneaky", RiotIdTagline: "C9", SummonerLevel: 398, Region: "na1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		{ID: 7, Puuid: "euw-puuid-faker", ProfileIcon: 3150, RiotIdGameName: "Faker", RiotIdTagline: "EUW", SummonerLevel: 425, Region: "euw1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		{ID: 8, Puuid: "euw-puuid-rekkles", ProfileIcon: 4320, RiotIdGameName: "Rekkles", RiotIdTagline: "FNC", SummonerLevel: 441, Region: "euw1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		{ID: 9, Puuid: "kr-puuid-faker", ProfileIcon: 4895, RiotIdGameName: "Faker", RiotIdTagline: "T1", SummonerLevel: 523, Region: "kr1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		{ID: 10, Puuid: "kr-puuid-showmaker", ProfileIcon: 4770, RiotIdGameName: "ShowMaker", RiotIdTagline: "DK", SummonerLevel: 489, Region: "kr1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		{ID: 11, Puuid: "kr-puuid-chovy", ProfileIcon: 4655, RiotIdGameName: "Chovy", RiotIdTagline: "GEN", SummonerLevel: 467, Region: "kr1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
		{ID: 12, Puuid: "kr-puuid-fafafa", ProfileIcon: 123, RiotIdGameName: "Fafafa", RiotIdTagline: "T1", SummonerLevel: 500, Region: "kr1", UnfetchedMatch: false, UpdatedAt: fixedDate, LastMatchFetch: fixedDate},
	}

	for _, pi := range playerInfo {
		err := db.Create(pi).Error
		require.NoError(t, err)
	}
}
