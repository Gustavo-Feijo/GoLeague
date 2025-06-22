package repositories

import (
	"goleague/pkg/database/models"

	"gorm.io/gorm"
)

// TimelineRepository is the public interface for handling timeline data.
type TimelineRepository interface {
	CreateBatchParticipantFrame(frames []*models.ParticipantFrame) error
}

// timelineRepository is the repository instance.
type timelineRepository struct {
	db *gorm.DB
}

// NewTimelineRepository creates a new timeline repository and returns it.
func NewTimelineRepository(db *gorm.DB) (TimelineRepository, error) {
	return &timelineRepository{db: db}, nil
}

// CreateBatchParticipantFrame creates the participant frames in batches of 1000.
func (ts *timelineRepository) CreateBatchParticipantFrame(frames []*models.ParticipantFrame) error {
	if len(frames) == 0 {
		return nil
	}
	return ts.db.CreateInBatches(&frames, 1000).Error
}

// CreateEventBatch is a generic function for creating the events in batches.
func CreateEventBatch[T any](db *gorm.DB, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}
	return db.CreateInBatches(&entities, 1000).Error
}
