package repositories

import (
	"fmt"
	"goleague/pkg/database"
	"goleague/pkg/database/models"

	"gorm.io/gorm"
)

// Public Interface.
type TimelineRepository interface {
	CreateBatchParticipantFrame(frames []*models.ParticipantFrame) error
	GetDb() *gorm.DB
}

// Timeline repository structure.
type timelineRepository struct {
	db *gorm.DB
}

// Create a timeline repository.
func NewTimelineRepository() (TimelineRepository, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &timelineRepository{db: db}, nil
}

// Create the participant frames in batches.
func (ts *timelineRepository) CreateBatchParticipantFrame(frames []*models.ParticipantFrame) error {
	if len(frames) == 0 {
		return nil
	}
	return ts.db.CreateInBatches(&frames, 1000).Error
}

// Create a way to make the timeline service available.
// Need to access the function above, since Go doesn't support it as method.
func (ts *timelineRepository) GetDb() *gorm.DB {
	return ts.db
}

// Generic function for creating the events in batches.
func CreateEventBatch[T any](db *gorm.DB, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}
	return db.CreateInBatches(&entities, 1000).Error
}
