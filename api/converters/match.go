package converters

import (
	"errors"
	"fmt"
	"goleague/api/dto"
	"goleague/api/repositories"
	tiervalues "goleague/pkg/riotvalues/tier"
)

type MatchConverter struct{}

// GroupRawMatchPreviewsByMatchId creates a map with the match id as key and the match previews array as value.
func (c MatchConverter) GroupRawMatchPreviewsByMatchId(rawPreviews []repositories.RawMatchPreview) map[string][]repositories.RawMatchPreview {
	grouped := make(map[string][]repositories.RawMatchPreview)

	for _, preview := range rawPreviews {
		grouped[preview.MatchID] = append(grouped[preview.MatchID], preview)
	}

	return grouped
}

// ConvertSingleMatch parses all previews for a given match and return it as the DTO.
func (c *MatchConverter) ConvertSingleMatch(matchPreviews []repositories.RawMatchPreview) (*dto.MatchPreview, error) {
	if len(matchPreviews) == 0 {
		return nil, errors.New("no previews provided")
	}

	// Validate all previews belong to the same match
	firstMatchID := matchPreviews[0].MatchID
	for _, preview := range matchPreviews {
		if preview.MatchID != firstMatchID {
			return nil, fmt.Errorf("mismatched match IDs: expected %s, got %s", firstMatchID, preview.MatchID)
		}
	}

	firstPreview := matchPreviews[0]

	result := &dto.MatchPreview{
		Metadata: c.NewMatchPreviewMetadata(firstPreview),
		Data:     make([]*dto.MatchPreviewData, 0, len(matchPreviews)),
	}

	// Convert each participant's data
	for _, rawPreview := range matchPreviews {
		previewData := c.NewMatchPreviewData(rawPreview)
		result.Data = append(result.Data, previewData)
	}

	return result, nil
}

// ConvertMultipleMatches creates a preview list from multiple raw match previews.
func (c *MatchConverter) ConvertMultipleMatches(rawPreviews []repositories.RawMatchPreview) (dto.MatchPreviewList, error) {
	if len(rawPreviews) == 0 {
		return dto.NewMatchPreviewList(), nil
	}

	// Group first
	grouped := c.GroupRawMatchPreviewsByMatchId(rawPreviews)

	result := dto.NewMatchPreviewList()

	// Convert each group
	for matchID, matchPreviews := range grouped {
		matchPreview, err := c.ConvertSingleMatch(matchPreviews)
		if err != nil {
			return nil, fmt.Errorf("failed to convert match %s: %w", matchID, err)
		}

		result.AddMatch(matchID, matchPreview)
	}

	return result, nil
}

func (c MatchConverter) NewMatchPreviewData(matchPreview repositories.RawMatchPreview) *dto.MatchPreviewData {
	rawItems := []*int{matchPreview.Item0, matchPreview.Item1, matchPreview.Item2, matchPreview.Item3, matchPreview.Item4, matchPreview.Item5}
	items := make([]int, 0, 6)
	for _, it := range rawItems {
		if it != nil && *it != 0 {
			items = append(items, *it)
		}
	}

	return &dto.MatchPreviewData{
		Assists:       matchPreview.Assists,
		ChampionID:    matchPreview.ChampionID,
		ChampionLevel: matchPreview.ChampionLevel,
		Deaths:        matchPreview.Deaths,
		GameName:      matchPreview.RiotIDGameName,
		Items:         items,
		Kills:         matchPreview.Kills,
		QueueID:       matchPreview.QueueID,
		Region:        matchPreview.Region,
		Tag:           matchPreview.RiotIDTagline,
		TeamId:        matchPreview.Team,
		TotalCs:       matchPreview.TotalMinionsKilled + matchPreview.NeutralMinionsKilled,
		Win:           matchPreview.Win,
	}
}

func (c MatchConverter) NewMatchPreviewMetadata(matchPreview repositories.RawMatchPreview) *dto.MatchPreviewMetadata {
	return &dto.MatchPreviewMetadata{
		AverageElo:   tiervalues.CalculateInverseRank(int(matchPreview.AverageRating)),
		Date:         matchPreview.Date,
		Duration:     matchPreview.Duration,
		InternalId:   matchPreview.InternalId,
		MatchId:      matchPreview.MatchID,
		QueueId:      matchPreview.QueueID,
		WinnerTeamId: matchPreview.WinnerTeamId,
	}
}
