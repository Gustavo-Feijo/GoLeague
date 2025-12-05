package converters

import (
	"errors"
	"fmt"
	"goleague/api/dto"
	matchrepo "goleague/api/repositories/match"
	"goleague/pkg/database/models"
	tiervalues "goleague/pkg/riotvalues/tier"
)

const (
	ErrFailedToConvert = "failed to convert match %s: %w"
	ErrMismatchedIds   = "mismatched match IDs: expected %s, got %s"
	ErrNoPreviews      = "no previews provided"
)

// ConvertSingleMatch parses all previews for a given match and return it as the DTO.
func ConvertSingleMatch(matchPreviews []matchrepo.RawMatchPreview) (*dto.MatchPreview, error) {
	if len(matchPreviews) == 0 {
		return nil, errors.New(ErrNoPreviews)
	}

	// Validate all previews belong to the same match
	firstMatchID := matchPreviews[0].MatchID
	for _, preview := range matchPreviews {
		if preview.MatchID != firstMatchID {
			return nil, fmt.Errorf(ErrMismatchedIds, firstMatchID, preview.MatchID)
		}
	}

	firstPreview := matchPreviews[0]

	result := &dto.MatchPreview{
		Metadata: NewMatchPreviewMetadata(firstPreview),
		Data:     make([]*dto.MatchPreviewData, 0, len(matchPreviews)),
	}

	// Convert each participant's data
	for _, rawPreview := range matchPreviews {
		previewData := NewMatchPreviewData(rawPreview)
		result.Data = append(result.Data, previewData)
	}

	return result, nil
}

// ConvertMultipleMatches creates a preview list from multiple raw match previews.
func ConvertMultipleMatches(rawPreviews []matchrepo.RawMatchPreview) (dto.MatchPreviewList, error) {
	if len(rawPreviews) == 0 {
		return dto.NewMatchPreviewList(), nil
	}

	// Group first
	grouped := GroupRawMatchPreviewsByMatchId(rawPreviews)

	result := dto.NewMatchPreviewList()

	// Convert each group
	for matchID, matchPreviews := range grouped {
		matchPreview, err := ConvertSingleMatch(matchPreviews)
		if err != nil {
			return nil, fmt.Errorf(ErrFailedToConvert, matchID, err)
		}

		result.AddMatch(matchID, matchPreview)
	}

	return result, nil
}

// GroupRawMatchPreviewsByMatchId creates a map with the match id as key and the match previews array as value.
func GroupRawMatchPreviewsByMatchId(rawPreviews []matchrepo.RawMatchPreview) map[string][]matchrepo.RawMatchPreview {
	grouped := make(map[string][]matchrepo.RawMatchPreview)

	for _, preview := range rawPreviews {
		grouped[preview.MatchID] = append(grouped[preview.MatchID], preview)
	}

	return grouped
}

// NewMatchPreviewData create a formatted match preview DTO.
func NewMatchPreviewData(matchPreview matchrepo.RawMatchPreview) *dto.MatchPreviewData {
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
func NewMatchPreviewMetadata(matchPreview matchrepo.RawMatchPreview) *dto.MatchPreviewMetadata {
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
func GroupParticipantFramesByParticipantId(participantFrames []matchrepo.RawMatchParticipantFrame) dto.ParticipantFrameList {
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
func ConvertEvents(rawEvents []models.AllEvents) []dto.MatchEvents {
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
