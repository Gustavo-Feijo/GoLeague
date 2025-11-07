package repositories

import (
	"fmt"
	"goleague/pkg/database/models"
	"time"

	"gorm.io/gorm"
)

// MatchRepository is the public interface for accessing the player repository.
type MatchRepository interface {
	GetMatchByMatchId(matchID string) (*models.MatchInfo, error)
	GetMatchPreviewsByInternalId(matchID uint) ([]RawMatchPreview, error)
	GetMatchPreviewsByInternalIds(matchIDs []uint) ([]RawMatchPreview, error)
	GetParticipantFramesByInternalId(matchID uint) ([]RawMatchParticipantFrame, error)
}

// matchRepository repository structure.
type matchRepository struct {
	db *gorm.DB
}

// NewMatchRepository creates a matchrepository.
func NewMatchRepository(db *gorm.DB) MatchRepository {
	return &matchRepository{db: db}
}

// RawMatchPreview is the raw data when getting a match data preview.
type RawMatchPreview struct {
	Assists              int       `gorm:"column:assists"`
	AverageRating        float64   `gorm:"column:average_rating"`
	ChampionID           int       `gorm:"column:champion_id"`
	ChampionLevel        int       `gorm:"column:champion_level"`
	Date                 time.Time `gorm:"column:match_start"`
	Deaths               int       `gorm:"column:deaths"`
	Duration             int       `gorm:"column:match_duration"`
	InternalId           uint      `gorm:"column:id"`
	Item0                *int      `gorm:"column:item0"`
	Item1                *int      `gorm:"column:item1"`
	Item2                *int      `gorm:"column:item2"`
	Item3                *int      `gorm:"column:item3"`
	Item4                *int      `gorm:"column:item4"`
	Item5                *int      `gorm:"column:item5"`
	Kills                int       `gorm:"column:kills"`
	MatchID              string    `gorm:"column:match_id"`
	NeutralMinionsKilled int       `gorm:"column:neutral_minions_killed"`
	ParticipantId        int       `gorm:"column:participant_id"`
	PlayerId             uint      `gorm:"column:player_id"`
	QueueID              int       `gorm:"column:queue_id"`
	Region               string    `gorm:"column:region"`
	RiotIDGameName       string    `gorm:"column:riot_id_game_name"`
	RiotIDTagline        string    `gorm:"column:riot_id_tagline"`
	Team                 int       `gorm:"column:team_id"`
	TotalMinionsKilled   int       `gorm:"column:total_minions_killed"`
	Win                  bool      `gorm:"column:win"`
	WinnerTeamId         int       `gorm:"column:winner_team_id"`
}

type RawMatchParticipantFrame struct {
	CurrentGold                   int `gorm:"column:current_gold"`
	FrameIndex                    int `gorm:"column:frame_index"`
	JungleMinionsKilled           int `gorm:"column:jungle_minions_killed"`
	Level                         int `gorm:"column:level"`
	MagicDamageDone               int `gorm:"column:magic_damage_done"`
	MagicDamageDoneToChampions    int `gorm:"column:magic_damage_done_to_champions"`
	MagicDamageTaken              int `gorm:"column:magic_damage_taken"`
	MatchStatID                   int `gorm:"column:match_stat_id"`
	MinionsKilled                 int `gorm:"column:minions_killed"`
	ParticipantID                 int `gorm:"column:participant_id"`
	PhysicalDamageDone            int `gorm:"column:physical_damage_done"`
	PhysicalDamageDoneToChampions int `gorm:"column:physical_damage_done_to_champions"`
	PhysicalDamageTaken           int `gorm:"column:physical_damage_taken"`
	TotalDamageDone               int `gorm:"column:total_damage_done"`
	TotalDamageDoneToChampions    int `gorm:"column:total_damage_done_to_champions"`
	TotalDamageTaken              int `gorm:"column:total_damage_taken"`
	TotalGold                     int `gorm:"column:total_gold"`
	TrueDamageDone                int `gorm:"column:true_damage_done"`
	TrueDamageDoneToChampions     int `gorm:"column:true_damage_done_to_champions"`
	TrueDamageTaken               int `gorm:"column:true_damage_taken"`
	XP                            int `gorm:"column:xp"`
}

// GetMatchPreviews gets the formatted preview for a list of matches.
func (ms *matchRepository) GetMatchPreviewsByInternalIds(matchIDs []uint) ([]RawMatchPreview, error) {
	var rawResults []RawMatchPreview

	query := `
		SELECT 
			mi.average_rating,
			mi.id,
			mi.match_duration,
			mi.match_id,
			mi.match_start,
			mi.match_winner as winner_team_id,
			mi.queue_id,
			ms.assists,
			ms.champion_id,
			ms.champion_level,
			ms.deaths,
			ms.item0,
			ms.item1,
			ms.item2,
			ms.item3,
			ms.item4,
			ms.item5,
			ms.kills,
			ms.neutral_minions_killed,
			ms.participant_id,
			ms.team_id,
			ms.total_minions_killed,
			ms.win,
			pi.id as player_id,
			pi.region,
			pi.riot_id_game_name,
			pi.riot_id_tagline
		FROM match_stats ms
		JOIN match_infos mi ON ms.match_id = mi.id
		JOIN player_infos pi ON ms.player_id = pi.id
		WHERE ms.match_id IN ?
	`

	if err := ms.db.Raw(query, matchIDs).Scan(&rawResults).Error; err != nil {
		return nil, err
	}

	return rawResults, nil
}

// GetMatchPreviews gets the formatted preview for a given match.
func (ms *matchRepository) GetMatchPreviewsByInternalId(matchID uint) ([]RawMatchPreview, error) {
	var rawResults []RawMatchPreview

	query := `
		SELECT 
			mi.average_rating,
			mi.id,
			mi.match_duration,
			mi.match_id,
			mi.match_start,
			mi.match_winner as winner_team_id,
			mi.queue_id,
			ms.assists,
			ms.champion_id,
			ms.champion_level,
			ms.deaths,
			ms.item0,
			ms.item1,
			ms.item2,
			ms.item3,
			ms.item4,
			ms.item5,
			ms.kills,
			ms.neutral_minions_killed,
			ms.participant_id,
			ms.team_id,
			ms.total_minions_killed,
			ms.win,
			pi.id as player_id,
			pi.region,
			pi.riot_id_game_name,
			pi.riot_id_tagline
		FROM match_stats ms
		JOIN match_infos mi ON ms.match_id = mi.id
		JOIN player_infos pi ON ms.player_id = pi.id
		WHERE ms.match_id = ?
	`

	if err := ms.db.Raw(query, matchID).Scan(&rawResults).Error; err != nil {
		return nil, err
	}

	return rawResults, nil
}

// GetMatchInfo returns all the matches information.
func (ms *matchRepository) GetMatchByMatchId(matchID string) (*models.MatchInfo, error) {
	var match models.MatchInfo
	if err := ms.db.Where(&models.MatchInfo{MatchId: matchID}).First(&match).Error; err != nil {
		return nil, fmt.Errorf("couldn't get the match by the match ID: %v", err)
	}

	return &match, nil
}

// GetParticipantFramesByInternalId retrieves all participant frames by the auto increment internal ID.
func (ms *matchRepository) GetParticipantFramesByInternalId(matchID uint) ([]RawMatchParticipantFrame, error) {
	var results []RawMatchParticipantFrame

	err := ms.db.
		Table("participant_frames").
		Select("participant_frames.*").
		Joins("JOIN match_stats ON participant_frames.match_stat_id = match_stats.id").
		Where("match_stats.match_id = ?", matchID).
		Find(&results).Error

	if err != nil {
		return nil, err
	}
	return results, nil
}
