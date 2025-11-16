package converters

import (
	"errors"
	"fmt"
	"goleague/api/dto"
	"goleague/api/repositories"
	"goleague/pkg/database/models"
	tiervalues "goleague/pkg/riotvalues/tier"
)

// MatchConverter is used to parse data to more usable/formatted structures.
type MatchConverter struct{}

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

// GroupRawMatchPreviewsByMatchId creates a map with the match id as key and the match previews array as value.
func (c MatchConverter) GroupRawMatchPreviewsByMatchId(rawPreviews []repositories.RawMatchPreview) map[string][]repositories.RawMatchPreview {
	grouped := make(map[string][]repositories.RawMatchPreview)

	for _, preview := range rawPreviews {
		grouped[preview.MatchID] = append(grouped[preview.MatchID], preview)
	}

	return grouped
}

// NewMatchPreviewData create a formatted match preview DTO.
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
		ParticipantId: matchPreview.ParticipantId,
		PlayerId:      matchPreview.PlayerId,
		QueueID:       matchPreview.QueueID,
		Region:        matchPreview.Region,
		Tag:           matchPreview.RiotIDTagline,
		TeamId:        matchPreview.Team,
		TotalCs:       matchPreview.TotalMinionsKilled + matchPreview.NeutralMinionsKilled,
		Win:           matchPreview.Win,
	}
}

// NewMatchPreviewMetadata generates the match metadata.
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

// GroupParticipantFramesByParticipantId creates a formatted dto of the participant frames.
func (c MatchConverter) GroupParticipantFramesByParticipantId(participantFrames []repositories.RawMatchParticipantFrame) dto.ParticipantFrameList {
	if len(participantFrames) == 0 {
		return dto.NewParticipantFrameList()
	}

	participantList := dto.NewParticipantFrameList()

	for _, frame := range participantFrames {
		formattedFrame := dto.ParticipantFrame{
			CurrentGold:                   frame.CurrentGold,
			FrameIndex:                    frame.FrameIndex,
			JungleMinionsKilled:           frame.JungleMinionsKilled,
			Level:                         frame.Level,
			MagicDamageDone:               frame.MagicDamageDone,
			MagicDamageDoneToChampions:    frame.MagicDamageDoneToChampions,
			MagicDamageTaken:              frame.MagicDamageTaken,
			MatchStatID:                   frame.MatchStatID,
			MinionsKilled:                 frame.MinionsKilled,
			ParticipantID:                 frame.ParticipantID,
			PhysicalDamageDone:            frame.PhysicalDamageDone,
			PhysicalDamageDoneToChampions: frame.PhysicalDamageDoneToChampions,
			PhysicalDamageTaken:           frame.PhysicalDamageTaken,
			TotalDamageDone:               frame.TotalDamageDone,
			TotalDamageDoneToChampions:    frame.TotalDamageDoneToChampions,
			TotalDamageTaken:              frame.TotalDamageTaken,
			TotalGold:                     frame.TotalGold,
			TrueDamageDone:                frame.TrueDamageDone,
			TrueDamageDoneToChampions:     frame.TrueDamageDoneToChampions,
			TrueDamageTaken:               frame.TrueDamageTaken,
			XP:                            frame.XP,
		}

		participantList.AddFrame(formattedFrame)
	}

	return participantList
}

// ConvertEvents converts raw events to formatted dto events.
func (c MatchConverter) ConvertEvents(rawEvents []models.AllEvents) []dto.MatchEvents {
	converted := make([]dto.MatchEvents, len(rawEvents))

	for i, event := range rawEvents {
		converted[i] = dto.MatchEvents{
			Timestamp:     event.Timestamp,
			EventType:     event.EventType,
			ParticipantId: event.ParticipantId,
			Data:          event.Data,
		}
	}

	return converted
}
