package matchservice

import (
	"fmt"
	"goleague/fetcher/data"
	matchfetcher "goleague/fetcher/data/match"
	"goleague/fetcher/repositories"
	batchservice "goleague/fetcher/services/mainregion/batch"
	eventservice "goleague/fetcher/services/mainregion/events"
	"goleague/pkg/database/models"

	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// TimelineService handles match timeline operations.
type TimelineService struct {
	db                 *gorm.DB
	fetcher            data.MainFetcher
	TimelineRepository repositories.TimelineRepository
	maxRetries         int
}

// NewTimelineService creates a new timeline service.
func NewTimelineService(
	db *gorm.DB,
	fetcher data.MainFetcher,
	timelineRepo repositories.TimelineRepository,
	maxRetries int,
) *TimelineService {
	return &TimelineService{
		db:                 db,
		fetcher:            fetcher,
		TimelineRepository: timelineRepo,
		maxRetries:         maxRetries,
	}
}

// GetMatchTimeline gets the timeline data for a match.
func (t *TimelineService) GetMatchTimeline(matchId string, onDemand bool) (*matchfetcher.MatchTimeline, error) {
	var matchData *matchfetcher.MatchTimeline
	var err error

	for attempt := 1; attempt < t.maxRetries; attempt++ {
		// Get the match timeline.
		matchData, err = t.fetcher.Match.GetMatchTimelineData(matchId, onDemand)

		// Everything went right, just continue normally..
		if err == nil {
			break
		}

		// Wait 5 seconds in case anything is wrong with the Riot API and try again.
		time.Sleep(5 * time.Second)
	}

	// Couldn't get even after multiple attempts.
	if err != nil {
		return nil, fmt.Errorf("couldn't get the match timeline data: %v", err)
	}

	return matchData, nil
}

// ProcessMatchTimeline processes the match timeline data and inserts it into the database.
func (t *TimelineService) ProcessMatchTimeline(
	matchTimeline *matchfetcher.MatchTimeline,
	statIdByPuuid map[string]uint64,
	matchInfo *models.MatchInfo,
	matchRepo repositories.MatchRepository,
) error {
	// Extract the stat ID for each participant entry.
	statIdByParticipantId := make(map[string]uint64)
	for _, participant := range matchTimeline.Info.Participants {
		pId := strconv.Itoa(participant.ParticipantId) // Converting to string to match the frames type.
		statIdByParticipantId[pId] = statIdByPuuid[participant.Puuid]
	}

	// Get the default frame interval.
	frameInterval := matchTimeline.Info.FrameInterval
	if err := matchRepo.SetFrameInterval(matchInfo.ID, frameInterval); err != nil {
		return fmt.Errorf("couldn't save the frame interval: %v", err)
	}

	// Create the frames slice and the event collector for handling batch insert.
	var framesToInsert []*models.ParticipantFrame
	eventCollector := batchservice.NewBatchCollector(t.db)

	// Loop through each available frame.
	for frameIndex, frame := range matchTimeline.Info.Frames {
		// Loop through the map of participants.
		for participantId, frameData := range frame.ParticipantFrames {
			// Get the stat id based on the participant id.
			matchStatId := statIdByParticipantId[participantId]

			// Append the participant frame to the list to batch insert.
			framesToInsert = append(framesToInsert, t.prepareParticipantsFrames(frameData, matchStatId, frameIndex))
		}

		// Process event handler service.
		eventService := eventservice.NewEventService(matchRepo)

		// Loop through each event frame available.
		for _, event := range frame.Event {
			if err := eventService.PrepareEvents(event, matchInfo, eventCollector); err != nil {
				// Don't need to add to the logger, usually associated with monsters not being killed before the next one spawns.
				log.Printf("Couldn't insert event %s on timestamp %d on match %s: %v", event.Type, event.Timestamp, matchInfo.MatchId, err)
			}
		}
	}

	// Insert the participant frames in a batch.
	if err := t.TimelineRepository.CreateBatchParticipantFrame(framesToInsert); err != nil {
		return fmt.Errorf("couldn't insert the participant frames on match %s: %v", matchInfo.MatchId, err)
	}

	// Process the events.
	err := eventCollector.ProcessBatches()

	return err
}

// prepareParticipantsFrames prepares and returns a participant frame to be inserted.
func (t *TimelineService) prepareParticipantsFrames(
	frame matchfetcher.ParticipantFrame,
	matchStatId uint64,
	frameId int,
) *models.ParticipantFrame {
	// Create the participant to be inserted in the database.
	participant := &models.ParticipantFrame{
		MatchStatId: matchStatId,
		FrameIndex:  frameId,

		CurrentGold:                   frame.CurrentGold,
		MagicDamageDone:               frame.DamageStats.MagicDamageDone,
		MagicDamageDoneToChampions:    frame.DamageStats.MagicDamageDoneToChampions,
		MagicDamageTaken:              frame.DamageStats.MagicDamageTaken,
		PhysicalDamageDone:            frame.DamageStats.PhysicalDamageDone,
		PhysicalDamageDoneToChampions: frame.DamageStats.PhysicalDamageDoneToChampions,
		PhysicalDamageTaken:           frame.DamageStats.PhysicalDamageTaken,
		TotalDamageDone:               frame.DamageStats.TotalDamageDone,
		TotalDamageDoneToChampions:    frame.DamageStats.TotalDamageDoneToChampions,
		TotalDamageTaken:              frame.DamageStats.TotalDamageTaken,
		TrueDamageDone:                frame.DamageStats.TrueDamageDone,
		TrueDamageDoneToChampions:     frame.DamageStats.TrueDamageDoneToChampions,
		TrueDamageTaken:               frame.DamageStats.TrueDamageTaken,
		JungleMinionsKilled:           frame.JungleMinionsKilled,
		Level:                         frame.Level,
		MinionsKilled:                 frame.MinionsKilled,
		ParticipantId:                 frame.ParticipantId,
		TotalGold:                     frame.TotalGold,
		XP:                            frame.XP,
	}

	return participant
}
