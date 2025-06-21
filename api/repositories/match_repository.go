package repositories

import (
	"fmt"
	"goleague/api/dto"
	"goleague/pkg/database"
	tiervalues "goleague/pkg/riotvalues/tier"
	"time"

	"gorm.io/gorm"
)

// MatchRepository is the public interface for accessing the player repository.
type MatchRepository interface {
	GetMatchPreviews(matchIDs []uint) (dto.MatchPreviewList, error)
}

// matchRepository repository structure.
type matchRepository struct {
	db *gorm.DB
}

// NewMatchRepository creates a matchrepository.
func NewMatchRepository() (MatchRepository, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &matchRepository{db: db}, nil
}

// RawMatchPreview is the raw data when getting a match data preview.
type RawMatchPreview struct {
	RiotIDGameName       string    `gorm:"column:riot_id_game_name"`
	RiotIDTagline        string    `gorm:"column:riot_id_tagline"`
	Region               string    `gorm:"column:region"`
	MatchID              string    `gorm:"column:match_id"`
	InternalId           uint      `gorm:"column:id"`
	Assists              int       `gorm:"column:assists"`
	Kills                int       `gorm:"column:kills"`
	Deaths               int       `gorm:"column:deaths"`
	ChampionLevel        int       `gorm:"column:champion_level"`
	ChampionID           int       `gorm:"column:champion_id"`
	Item0                *int      `gorm:"column:item0"`
	Item1                *int      `gorm:"column:item1"`
	Item2                *int      `gorm:"column:item2"`
	Item3                *int      `gorm:"column:item3"`
	Item4                *int      `gorm:"column:item4"`
	Item5                *int      `gorm:"column:item5"`
	NeutralMinionsKilled int       `gorm:"column:neutral_minions_killed"`
	TotalMinionsKilled   int       `gorm:"column:total_minions_killed"`
	Duration             int       `gorm:"column:match_duration"`
	Date                 time.Time `gorm:"column:match_start"`
	Win                  bool      `gorm:"column:win"`
	AverageRating        float64   `gorm:"column:average_rating"`
	QueueID              int       `gorm:"column:queue_id"`
}

// GetMatchPreviews gets the formatted preview for a list of matches.
func (ms *matchRepository) GetMatchPreviews(matchIDs []uint) (dto.MatchPreviewList, error) {
	var rawResults []RawMatchPreview

	query := `
		SELECT 
			pi.riot_id_game_name,
			pi.riot_id_tagline,
			pi.region,
			mi.match_id,
			ms.assists,
			ms.kills,
			ms.deaths,
			ms.champion_level,
			ms.champion_id,
			ms.item0,
			ms.item1,
			ms.item2,
			ms.item3,
			ms.item4,
			ms.item5,
			ms.neutral_minions_killed,
			ms.total_minions_killed,
			ms.win,
			mi.id,
			mi.match_duration,
			mi.match_start,
			mi.average_rating,
			mi.queue_id
		FROM match_stats ms
		JOIN match_infos mi ON ms.match_id = mi.id
		JOIN player_infos pi ON ms.player_id = pi.id
		WHERE ms.match_id IN ?
	`

	if err := ms.db.Raw(query, matchIDs).Scan(&rawResults).Error; err != nil {
		return nil, err
	}

	return formatPreviews(rawResults), nil
}

// formatPreviews return the formatted dto
func formatPreviews(rawPreviews []RawMatchPreview) dto.MatchPreviewList {
	fullPreview := make(dto.MatchPreviewList)

	// Range through each raw preview and format it.
	for _, r := range rawPreviews {

		// Initialize the full preview.
		if _, ok := fullPreview[r.MatchID]; !ok {
			fullPreview[r.MatchID] = &dto.MatchPreview{
				Metadata: &dto.MatchPreviewMetadata{
					AverageElo: tiervalues.CalculateInverseRank(int(r.AverageRating)),
					Date:       r.Date,
					Duration:   r.Duration,
					InternalId: r.InternalId,
					MatchId:    r.MatchID,
					QueueId:    r.QueueID,
				},
				Data: make([]*dto.MatchPreviewData, 0),
			}
		}

		items := make([]int, 0, 6)

		// Add non-null items to the array
		if r.Item0 != nil && *r.Item0 != 0 {
			items = append(items, *r.Item0)
		}
		if r.Item1 != nil && *r.Item1 != 0 {
			items = append(items, *r.Item1)
		}
		if r.Item2 != nil && *r.Item2 != 0 {
			items = append(items, *r.Item2)
		}
		if r.Item3 != nil && *r.Item3 != 0 {
			items = append(items, *r.Item3)
		}
		if r.Item4 != nil && *r.Item4 != 0 {
			items = append(items, *r.Item4)
		}
		if r.Item5 != nil && *r.Item5 != 0 {
			items = append(items, *r.Item5)
		}

		preview := &dto.MatchPreviewData{
			GameName:      r.RiotIDGameName,
			Tag:           r.RiotIDTagline,
			Region:        r.Region,
			Assists:       r.Assists,
			Kills:         r.Kills,
			Deaths:        r.Deaths,
			ChampionLevel: r.ChampionLevel,
			ChampionID:    r.ChampionID,
			TotalCs:       r.TotalMinionsKilled + r.NeutralMinionsKilled,
			Items:         items,
			Win:           r.Win,
			QueueID:       r.QueueID,
		}

		fullPreview[r.MatchID].Data = append(fullPreview[r.MatchID].Data, preview)
	}

	return fullPreview
}
