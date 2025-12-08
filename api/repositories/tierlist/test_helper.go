package repositories

import "testing"

func getTierlistExpectedResult(t *testing.T, testName string) []*TierlistResult {
	t.Helper()

	switch testName {
	case "nofilters":
		return getTierlistNoFilters()
	case "allgold":
		return getTierlistAllGold()
	case "allabovegold":
		return getTierlistAllAboveGold()
	}

	return nil
}

func getTierlistNoFilters() []*TierlistResult {
	return []*TierlistResult{
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   3,
			PickCount:    1,
			PickRate:     12.5,
			TeamPosition: "MIDDLE",
			WinRate:      100,
		},
		{
			BanCount:     4,
			BanRate:      50,
			ChampionId:   1,
			PickCount:    8,
			PickRate:     100,
			TeamPosition: "BOTTOM",
			WinRate:      75,
		},
		{
			BanCount:     2,
			BanRate:      25,
			ChampionId:   5,
			PickCount:    3,
			PickRate:     42.86,
			TeamPosition: "JUNGLE",
			WinRate:      66.67,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   4,
			PickCount:    8,
			PickRate:     100,
			TeamPosition: "TOP",
			WinRate:      62.5,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   2,
			PickCount:    8,
			PickRate:     100,
			TeamPosition: "UTILITY",
			WinRate:      62.5,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   6,
			PickCount:    7,
			PickRate:     87.5,
			TeamPosition: "MIDDLE",
			WinRate:      57.14,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   7,
			PickCount:    3,
			PickRate:     42.86,
			TeamPosition: "JUNGLE",
			WinRate:      0,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   8,
			PickCount:    1,
			PickRate:     14.29,
			TeamPosition: "JUNGLE",
			WinRate:      0,
		},
	}
}

func getTierlistAllGold() []*TierlistResult {
	return []*TierlistResult{
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   2,
			PickCount:    2,
			PickRate:     100,
			TeamPosition: "UTILITY",
			WinRate:      100,
		},
		{
			BanCount:     1,
			BanRate:      50,
			ChampionId:   1,
			PickCount:    2,
			PickRate:     100,
			TeamPosition: "BOTTOM",
			WinRate:      50,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   4,
			PickCount:    2,
			PickRate:     100,
			TeamPosition: "TOP",
			WinRate:      50,
		},
		{
			BanCount:     1,
			BanRate:      50,
			ChampionId:   5,
			PickCount:    2,
			PickRate:     100,
			TeamPosition: "JUNGLE",
			WinRate:      50,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   6,
			PickCount:    2,
			PickRate:     100,
			TeamPosition: "MIDDLE",
			WinRate:      50,
		},
	}
}

func getTierlistAllAboveGold() []*TierlistResult {
	return []*TierlistResult{
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   3,
			PickCount:    1,
			PickRate:     14.29,
			TeamPosition: "MIDDLE",
			WinRate:      100,
		},
		{
			BanCount:     4,
			BanRate:      57.14,
			ChampionId:   1,
			PickCount:    7,
			PickRate:     100,
			TeamPosition: "BOTTOM",
			WinRate:      85.71,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   6,
			PickCount:    6,
			PickRate:     85.71,
			TeamPosition: "MIDDLE",
			WinRate:      66.67,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   2,
			PickCount:    7,
			PickRate:     100,
			TeamPosition: "UTILITY",
			WinRate:      57.14,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   4,
			PickCount:    7,
			PickRate:     100,
			TeamPosition: "TOP",
			WinRate:      57.14,
		},
		{
			BanCount:     1,
			BanRate:      14.29,
			ChampionId:   5,
			PickCount:    2,
			PickRate:     33.33,
			TeamPosition: "JUNGLE",
			WinRate:      50,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   7,
			PickCount:    3,
			PickRate:     50,
			TeamPosition: "JUNGLE",
			WinRate:      0,
		},
		{
			BanCount:     0,
			BanRate:      0,
			ChampionId:   8,
			PickCount:    1,
			PickRate:     16.67,
			TeamPosition: "JUNGLE",
			WinRate:      0,
		},
	}
}
