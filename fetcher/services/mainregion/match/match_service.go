package matchservice

import (
	"errors"
	"fmt"
	"goleague/fetcher/data"
	matchfetcher "goleague/fetcher/data/match"
	"goleague/fetcher/regions"
	"goleague/fetcher/repositories"
	playerservice "goleague/fetcher/services/mainregion/player"
	"goleague/pkg/database/models"

	"time"
)

// MatchService handles functionality related to matches.
type MatchService struct {
	fetcher            data.MainFetcher
	MatchRepository    repositories.MatchRepository
	PlayerRepository   repositories.PlayerRepository
	RatingRepository   repositories.RatingRepository
	TimelineRepository repositories.TimelineRepository
	maxRetries         int
}

// NewMatchService creates a new match service.
func NewMatchService(
	fetcher data.MainFetcher,
	matchRepo repositories.MatchRepository,
	playerRepo repositories.PlayerRepository,
	ratingRepo repositories.RatingRepository,
	timelineRepo repositories.TimelineRepository,
	maxRetries int,
) *MatchService {
	return &MatchService{
		fetcher:            fetcher,
		MatchRepository:    matchRepo,
		PlayerRepository:   playerRepo,
		RatingRepository:   ratingRepo,
		TimelineRepository: timelineRepo,
		maxRetries:         maxRetries,
	}
}

// GetMatchData gets the data of the match from the Riot API.
func (m *MatchService) GetMatchData(matchId string) (*matchfetcher.MatchData, error) {
	var matchData *matchfetcher.MatchData
	var err error

	for attempt := 1; attempt < m.maxRetries; attempt++ {
		// Get the match data.
		matchData, err = m.fetcher.Match.GetMatchData(matchId, false)

		// Everything went right, just continue normally.
		if err == nil {
			break
		}

		// Wait 5 seconds in case anything is wrong with the Riot API and try again.
		time.Sleep(5 * time.Second)
	}

	// Couldn't get even after multiple attempts.
	if err != nil {
		return nil, fmt.Errorf("couldn't get the match data: %v", err)
	}

	return matchData, nil
}

// ProcessMatchInfo retrieves the match info and inserts it into the database.
func (m *MatchService) ProcessMatchInfo(
	match *matchfetcher.MatchData,
	matchId string,
) (*models.MatchInfo, error) {
	// Create a match to be inserted.
	matchInfo := &models.MatchInfo{
		GameVersion:    match.Info.GameVersion,
		MatchId:        matchId,
		MatchStart:     match.Info.GameCreation.Time(),
		MatchDuration:  match.Info.GameDuration,
		MatchSurrender: match.Info.Participants[0].GameEndedInSurrender,
		MatchRemake:    match.Info.Participants[0].GameEndedInEarlySurrender,
		QueueId:        match.Info.QueueId,
	}

	// Create the match.
	// Return the match that we tried to insert and the error result of the insert (Nil or error).
	return matchInfo, m.MatchRepository.CreateMatchInfo(matchInfo)
}

// ProcessMatchBans retrieves the bans and creates them.
func (m *MatchService) ProcessMatchBans(
	matchTeams []matchfetcher.TeamInfo,
	matchInfo *models.MatchInfo,
) ([]*models.MatchBans, error) {
	var bans []*models.MatchBans

	// Get all the bans.
	for _, team := range matchTeams {
		for _, ban := range team.Bans {
			bans = append(bans, &models.MatchBans{
				MatchId:    matchInfo.ID,
				PickTurn:   ban.PickTurn,
				ChampionId: ban.ChampionId,
			})
		}
	}

	// Only need to insert if there is any bans.
	// Some modes don't have bans.
	if len(bans) != 0 {
		// Create the bans.
		if err := m.MatchRepository.CreateMatchBans(bans); err != nil {
			return nil, err
		}
	}

	return bans, nil
}

// ProcessMatchData processes the match data and inserts it into the database.
func (m *MatchService) ProcessMatchData(
	match *matchfetcher.MatchData,
	matchId string,
	region regions.SubRegion,
) (*models.MatchInfo, []*models.MatchBans, []*models.MatchStats, error) {
	// Process the match infos.
	matchInfo, err := m.ProcessMatchInfo(match, matchId)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("couldn't create the match info for the match %s: %v", matchId, err)
	}

	// Process the bans.
	bans, err := m.ProcessMatchBans(match.Info.Teams, matchInfo)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("couldn't create the bans for the match %s: %v", matchInfo.MatchId, err)
	}

	// Process each player.
	playerService := playerservice.NewPlayerService(m.MatchRepository, m.PlayerRepository, m.RatingRepository)
	playersToUpsert, participantByPuuid, err := playerService.ProcessPlayersFromMatch(match.Info.Participants, matchInfo, region)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("couldn't create the players for the match %s: %v", matchInfo.MatchId, err)
	}

	// Process the match stats.
	stats, err := m.ProcessMatchStats(playersToUpsert, participantByPuuid, matchInfo)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("couldn't create the stats for the match %s: %v", matchInfo.MatchId, err)
	}

	return matchInfo, bans, stats, nil
}

// ProcessMatchStats procesesses and inserts match stats for each player.
func (m *MatchService) ProcessMatchStats(
	playersToUpsert []*models.PlayerInfo,
	participants map[string]matchfetcher.MatchPlayer,
	matchInfo *models.MatchInfo,
) ([]*models.MatchStats, error) {
	var statsToUpsert []*models.MatchStats
	for _, player := range playersToUpsert {
		participant, exists := participants[player.Puuid]
		if !exists {
			// Should never occur.
			return nil, errors.New("the participant is not present in the map")
		}

		// Create the match stats.
		newStat := &models.MatchStats{
			MatchId:    matchInfo.ID,
			PlayerId:   player.ID,
			PlayerData: participant,
		}

		statsToUpsert = append(statsToUpsert, newStat)
	}

	// Create/update the players.
	if err := m.MatchRepository.CreateMatchStats(statsToUpsert); err != nil {
		return nil, err
	}

	return statsToUpsert, nil
}
