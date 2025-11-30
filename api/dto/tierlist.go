package dto

import tierlistrepo "goleague/api/repositories/tierlist"

// Result of a tierlist fetch.
type TierlistResult struct {
	BanCount     int     `json:"banCount"`
	BanRate      float64 `json:"banRate"`
	ChampionId   int     `json:"championId"`
	PickCount    int     `json:"pickCount"`
	PickRate     float64 `json:"pickRate"`
	TeamPosition string  `json:"teamPosition"`
	WinRate      float64 `json:"winRate"`
}

// FromRepositorySlice creates the DTO from the repository result (Same structure)
func (TierlistResult) FromRepositorySlice(repoResults []*tierlistrepo.TierlistResult) []*TierlistResult {
	dtoResults := make([]*TierlistResult, len(repoResults))

	for i, repo := range repoResults {
		dtoResults[i] = &TierlistResult{
			BanCount:     repo.BanCount,
			BanRate:      repo.BanRate,
			ChampionId:   repo.ChampionId,
			PickCount:    repo.PickCount,
			PickRate:     repo.PickRate,
			TeamPosition: repo.TeamPosition,
			WinRate:      repo.WinRate,
		}
	}

	return dtoResults
}
